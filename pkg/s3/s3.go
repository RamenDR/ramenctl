// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package s3

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/logging"
	"go.uber.org/zap"

	"github.com/ramendr/ramenctl/pkg/time"
)

const (
	// S3 output directory name
	dirName = "s3"

	// S3 output directory permission
	dirPerm = 0o750
)

// Profile contains S3 connection and authentication information.
type Profile struct {
	Name          string
	Bucket        string
	Region        string
	Endpoint      string
	CACertificate []byte
	AccessKey     string
	SecretKey     string
}

// objectStore wraps an S3 client with profile information and log.
type objectStore struct {
	client  *s3.Client
	profile *Profile
	log     *zap.SugaredLogger
}

// Gather creates client for the given profile and downloads objects from S3
// using the provided prefix.
func Gather(
	ctx context.Context,
	profile *Profile,
	prefix, outputDir string,
	log *zap.SugaredLogger,
) error {
	objectStore, err := new(ctx, profile, log)
	if err != nil {
		return fmt.Errorf("failed to create S3 client for profile %q: %w",
			profile.Name, err)
	}

	if err := objectStore.downloadObjects(ctx, prefix, outputDir); err != nil {
		return fmt.Errorf("failed to download objects from profile %q: %w",
			profile.Name, err)
	}

	return nil
}

// new creates an S3 client for the given profile.
func new(
	ctx context.Context,
	profile *Profile,
	log *zap.SugaredLogger,
) (*objectStore, error) {
	configOptions := []func(*config.LoadOptions) error{
		config.WithRegion(profile.Region),
		config.WithBaseEndpoint(profile.Endpoint),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			profile.AccessKey,
			profile.SecretKey,
			"",
		)),
		// Add zap logger to the config to redirect AWS SDK logs.
		config.WithLogger(awsSDKLogger(log)),
	}

	// Add CA certificate to the client config if provided in the profile.
	if len(profile.CACertificate) > 0 {
		log.Debugf("Using custom CA certificate for S3 profile %q", profile.Name)
		configOptions = append(
			configOptions,
			config.WithCustomCABundle(bytes.NewReader(profile.CACertificate)),
		)
	}

	config, err := config.LoadDefaultConfig(ctx, configOptions...)
	if err != nil {
		return nil, fmt.Errorf("failed to load s3 config for profile %q: %w",
			profile.Name, err)
	}

	s3Client := s3.NewFromConfig(config, func(o *s3.Options) {
		// Use path-style addressing(https://endpoint/bucket/key) instead of
		// virtual-hosted style(https://bucket.endpoint/key).
		o.UsePathStyle = true
	})

	return &objectStore{
		client:  s3Client,
		profile: profile,
		log:     log,
	}, nil
}

// downloadObjects downloads all objects with the given prefix to the output directory.
func (s *objectStore) downloadObjects(ctx context.Context, prefix, outputDir string) error {
	start := time.Now()

	keys, err := s.listObjects(ctx, prefix)
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		s.log.Warnf("No objects found in bucket %q with prefix %q",
			s.profile.Bucket, prefix)
		return nil
	}

	profileDir := filepath.Join(outputDir, dirName, s.profile.Name)
	if err := os.MkdirAll(profileDir, dirPerm); err != nil {
		return err
	}

	downloaded := 0
	for _, key := range keys {
		if err := s.downloadObject(ctx, key, profileDir); err != nil {
			s.log.Warnf("Failed to download object %q from bucket %q: %v",
				key, s.profile.Bucket, err)
			continue
		}
		downloaded++
	}

	s.log.Debugf("Downloaded %d/%d objects from bucket %q in %.3f seconds",
		downloaded, len(keys), s.profile.Bucket, time.Since(start).Seconds())

	return nil
}

// listObjects returns all object keys under the given prefix.
func (s *objectStore) listObjects(ctx context.Context, prefix string) ([]string, error) {
	start := time.Now()

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.profile.Bucket),
	}

	// Only set prefix if it's not empty.
	if prefix != "" {
		input.Prefix = aws.String(prefix)
	}

	paginator := s3.NewListObjectsV2Paginator(s.client, input)

	var keys []string
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects in bucket %q with prefix %q: %w",
				s.profile.Bucket, prefix, err)
		}
		for _, obj := range page.Contents {
			keys = append(keys, *obj.Key)
		}
	}

	s.log.Debugf("Listed %d objects in bucket %q with prefix %q in %.3f seconds",
		len(keys), s.profile.Bucket, prefix, time.Since(start).Seconds())

	return keys, nil
}

// downloadObject downloads and decompress an object from S3 store.
func (c *objectStore) downloadObject(ctx context.Context, key string, profileDir string) error {
	start := time.Now()

	result, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.profile.Bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to get object %q from bucket %q: %w",
			key, c.profile.Bucket, err)
	}
	defer result.Body.Close()

	// Log ContentEncoding for debugging.
	contentEncoding := ""
	if result.ContentEncoding != nil {
		contentEncoding = *result.ContentEncoding
	}
	c.log.Debugf("Object %q ContentEncoding: %q", key, contentEncoding)

	// Check Content-Encoding header first
	isGzippedByHeader := result.ContentEncoding != nil && *result.ContentEncoding == "gzip"

	// If not indicated by header, peek at first 2 bytes to check for gzip magic number.
	var reader io.Reader = result.Body
	var isGzipped bool

	if isGzippedByHeader {
		isGzipped = true
	} else {
		// Peek at first 2 bytes to detect gzip magic number (0x1f 0x8b).
		peekReader := bufio.NewReader(result.Body)
		magicBytes, err := peekReader.Peek(2)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to peek object %q: %w", key, err)
		}

		isGzipped = len(magicBytes) >= 2 && magicBytes[0] == 0x1f && magicBytes[1] == 0x8b
		reader = peekReader
	}

	// Remove .gz extension if present.
	outputKey := key
	if strings.HasSuffix(key, ".gz") {
		outputKey = strings.TrimSuffix(key, ".gz")
	}

	objectPath := filepath.Join(profileDir, outputKey)
	if err := os.MkdirAll(filepath.Dir(objectPath), dirPerm); err != nil {
		return err
	}

	file, err := os.Create(objectPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Decompress if gzipped.
	if isGzipped {
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader for object %q: %w", key, err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write object %q to file %q: %w",
			key, objectPath, err)
	}

	// Close explicitly after copy.
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file %q: %w", objectPath, err)
	}

	if isGzipped {
		c.log.Debugf("Downloaded and decompressed object %q in %.3f seconds",
			key, time.Since(start).Seconds())
	} else {
		c.log.Debugf("Downloaded object %q in %.3f seconds",
			key, time.Since(start).Seconds())
	}

	return nil
}

// awsSDKLogger creates an AWS SDK logger that redirects logs to zap logger.
func awsSDKLogger(log *zap.SugaredLogger) logging.LoggerFunc {
	s3Logger := log.Named("aws-sdk")

	return logging.LoggerFunc(
		func(classification logging.Classification, format string, v ...any) {
			s3Logger.Debugf(format, v...)
		},
	)
}
