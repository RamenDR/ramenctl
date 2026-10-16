// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package s3

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestCheckBucketCanceled(t *testing.T) {
	server := newSlowServer()
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	options := Options{Timeout: 30 * time.Second}
	err := checkBucket(ctx, newTestProfile(server.URL), options, newTestLogger())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestCheckBucketTimeout(t *testing.T) {
	server := newSlowServer()
	defer server.Close()

	options := Options{Timeout: 100 * time.Millisecond}
	err := checkBucket(context.Background(), newTestProfile(server.URL), options, newTestLogger())
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestGatherDataCanceled(t *testing.T) {
	var requests atomic.Int32
	server := newCountingSlowServer(&requests)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	options := Options{Timeout: 30 * time.Second}
	err := gatherData(
		ctx,
		newTestProfile(server.URL),
		[]string{"prefix1", "prefix2"},
		t.TempDir(),
		options,
		newTestLogger(),
	)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}

	// Verify that prefix2 was not processed after cancellation.
	if n := requests.Load(); n != 1 {
		t.Fatalf("expected 1 request, got %d", n)
	}
}

func TestGatherDataTimeout(t *testing.T) {
	server := newSlowServer()
	defer server.Close()

	options := Options{Timeout: 100 * time.Millisecond}
	err := gatherData(
		context.Background(),
		newTestProfile(server.URL),
		[]string{"prefix1"},
		t.TempDir(),
		options,
		newTestLogger(),
	)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}
}

// newSlowServer creates a test HTTP server that waits for 1 second before
// returning an error. If the code is working correctly, the request will be
// canceled or timed out before the server responds. If the code is broken,
// the test will delay by 1 second and fail with an unexpected error instead
// of hanging the entire build.
func newSlowServer() *httptest.Server {
	return newCountingSlowServer(nil)
}

// newCountingSlowServer creates a slow test server that counts requests.
func newCountingSlowServer(requests *atomic.Int32) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requests != nil {
			requests.Add(1)
		}

		select {
		case <-r.Context().Done():
		case <-time.After(1 * time.Second):
			http.Error(w, "slow server timeout", http.StatusServiceUnavailable)
		}
	}))
}

// newTestProfile creates a test S3 profile pointing to the given server URL.
func newTestProfile(url string) *Profile {
	return &Profile{
		Name:               "test-profile",
		Bucket:             "test-bucket",
		Region:             "us-east-1",
		Endpoint:           url,
		AWSAccessKeyID:     []byte("test-key"),
		AWSSecretAccessKey: []byte("test-secret"),
	}
}

// newTestLogger creates a no-op logger for tests.
func newTestLogger() *zap.SugaredLogger {
	return zap.NewNop().Sugar()
}
