// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package helpers

import (
	"bytes"
	"errors"

	"github.com/ramendr/ramen/e2e/types"
	corev1 "k8s.io/api/core/v1"

	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/s3"
	"github.com/ramendr/ramenctl/pkg/validation"
)

// ValidationMock implements the validation.Validation interface. All operations succeed without
// accessing the clusters. To cause operations to fail or return non default values, set a function
// returning an error.
type ValidationMock struct {
	ValidateFunc              func(validation.Context) error
	ApplicationNamespacesFunc func(ctx validation.Context, drpcName, drpcNamespace string) ([]string, error)
	GatherFunc                func(ctx validation.Context, clsuters []*types.Cluster, options gathering.Options) <-chan gathering.Result
	GatherS3Func              func(ctx validation.Context, profiles []*s3.Profile, prefixes []string, outputDir string) <-chan s3.Result
	GetSecretFunc             func(ctx validation.Context, cluster *types.Cluster, name, namespace string) (*corev1.Secret, error)
	CheckS3Func               func(ctx validation.Context, profiles []*s3.Profile) <-chan s3.Result
}

var _ validation.Validation = &ValidationMock{}

func (m *ValidationMock) Validate(ctx validation.Context) error {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(ctx)
	}
	return nil
}

func (m *ValidationMock) ApplicationNamespaces(
	ctx validation.Context,
	drpcName, drpcNamespace string,
) ([]string, error) {
	if m.ApplicationNamespacesFunc != nil {
		return m.ApplicationNamespacesFunc(ctx, drpcName, drpcNamespace)
	}
	return nil, nil
}

func (m *ValidationMock) Gather(
	ctx validation.Context,
	clusters []*types.Cluster,
	options gathering.Options,
) <-chan gathering.Result {
	if m.GatherFunc != nil {
		return m.GatherFunc(ctx, clusters, options)
	}

	results := make(chan gathering.Result, len(clusters))
	for _, cluster := range clusters {
		results <- gathering.Result{Name: cluster.Name, Err: nil}
	}
	close(results)
	return results
}

func (m *ValidationMock) GetSecret(
	ctx validation.Context,
	cluster *types.Cluster,
	name, namespace string,
) (*corev1.Secret, error) {
	if m.GetSecretFunc != nil {
		return m.GetSecretFunc(ctx, cluster, name, namespace)
	}
	return &corev1.Secret{
		Data: map[string][]byte{
			"AWS_ACCESS_KEY_ID":     []byte(FakeAWSKeyID),
			"AWS_SECRET_ACCESS_KEY": []byte(FakeAWSKey),
		},
	}, nil
}

func (m *ValidationMock) GatherS3(
	ctx validation.Context,
	profiles []*s3.Profile,
	prefixes []string,
	outputDir string,
) <-chan s3.Result {
	if m.GatherS3Func != nil {
		return m.GatherS3Func(ctx, profiles, prefixes, outputDir)
	}
	results := make(chan s3.Result, len(profiles))
	for _, profile := range profiles {
		if !bytes.Equal(profile.AWSAccessKeyID, []byte(FakeAWSKeyID)) ||
			!bytes.Equal(profile.AWSSecretAccessKey, []byte(FakeAWSKey)) {
			results <- s3.Result{ProfileName: profile.Name, Err: errors.New("invalid credentials")}
		} else {
			results <- s3.Result{ProfileName: profile.Name, Err: nil}
		}
	}
	close(results)
	return results
}

func (m *ValidationMock) CheckS3(ctx validation.Context, profiles []*s3.Profile) <-chan s3.Result {
	if m.CheckS3Func != nil {
		return m.CheckS3Func(ctx, profiles)
	}
	results := make(chan s3.Result, len(profiles))
	for _, profile := range profiles {
		if !bytes.Equal(profile.AWSAccessKeyID, []byte(FakeAWSKeyID)) ||
			!bytes.Equal(profile.AWSSecretAccessKey, []byte(FakeAWSKey)) {
			results <- s3.Result{ProfileName: profile.Name, Err: errors.New("invalid credentials")}
		} else {
			results <- s3.Result{ProfileName: profile.Name, Err: nil}
		}
	}
	close(results)
	return results
}
