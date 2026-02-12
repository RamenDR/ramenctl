// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package testing

import (
	"errors"

	"github.com/ramendr/ramen/e2e/types"
	corev1 "k8s.io/api/core/v1"

	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/s3"
)

type ContextFunc func(types.Context) error
type TestContextFunc func(types.TestContext) error

// Mock implements the testing.Testing interface. All operations succeed without accessing
// the clusters. To cause operations to fail, set a function returning an error.
type Mock struct {
	// Operations on types.Context
	ValidateFunc ContextFunc
	SetupFunc    ContextFunc
	CleanupFunc  ContextFunc

	// Operations on types.TestContext
	DeployFunc    TestContextFunc
	UndeployFunc  TestContextFunc
	ProtectFunc   TestContextFunc
	UnprotectFunc TestContextFunc
	FailoverFunc  TestContextFunc
	RelocateFunc  TestContextFunc
	PurgeFunc     TestContextFunc

	// Handling failures.
	GatherFunc    func(ctx types.Context, clsuters []*types.Cluster, options gathering.Options) <-chan gathering.Result
	GetSecretFunc func(ctx types.Context, cluster *types.Cluster, name, namespace string) (*corev1.Secret, error)
	GatherS3Func  func(ctx types.Context, profiles []*s3.Profile, prefixes []string, outputDir string) <-chan s3.Result
}

var _ Testing = &Mock{}

func (m *Mock) Validate(ctx types.Context) error {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(ctx)
	}
	return nil
}

func (m *Mock) Setup(ctx types.Context) error {
	if m.SetupFunc != nil {
		return m.SetupFunc(ctx)
	}
	return nil
}

func (m *Mock) Cleanup(ctx types.Context) error {
	if m.CleanupFunc != nil {
		return m.CleanupFunc(ctx)
	}
	return nil
}

func (m *Mock) Deploy(ctx types.TestContext) error {
	if m.DeployFunc != nil {
		return m.DeployFunc(ctx)
	}
	return nil
}

func (m *Mock) Undeploy(ctx types.TestContext) error {
	if m.UndeployFunc != nil {
		return m.UndeployFunc(ctx)
	}
	return nil
}

func (m *Mock) Protect(ctx types.TestContext) error {
	if m.ProtectFunc != nil {
		return m.ProtectFunc(ctx)
	}
	return nil
}

func (m *Mock) Unprotect(ctx types.TestContext) error {
	if m.UnprotectFunc != nil {
		return m.UnprotectFunc(ctx)
	}
	return nil
}

func (m *Mock) Failover(ctx types.TestContext) error {
	if m.FailoverFunc != nil {
		return m.FailoverFunc(ctx)
	}
	return nil
}

func (m *Mock) Relocate(ctx types.TestContext) error {
	if m.RelocateFunc != nil {
		return m.RelocateFunc(ctx)
	}
	return nil
}

func (m *Mock) Purge(ctx types.TestContext) error {
	if m.PurgeFunc != nil {
		return m.PurgeFunc(ctx)
	}
	return nil
}

func (m *Mock) Gather(
	ctx types.Context,
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

func (m *Mock) GetSecret(
	ctx types.Context,
	cluster *types.Cluster,
	name, namespace string,
) (*corev1.Secret, error) {
	if m.GetSecretFunc != nil {
		return m.GetSecretFunc(ctx, cluster, name, namespace)
	}
	return &corev1.Secret{
		Data: map[string][]byte{
			"AWS_ACCESS_KEY_ID":     []byte(helpers.AccessKey),
			"AWS_SECRET_ACCESS_KEY": []byte(helpers.SecretKey),
		},
	}, nil
}

func (m *Mock) GatherS3(
	ctx types.Context,
	profiles []*s3.Profile,
	prefixes []string,
	outputDir string,
) <-chan s3.Result {
	if m.GatherS3Func != nil {
		return m.GatherS3Func(ctx, profiles, prefixes, outputDir)
	}
	results := make(chan s3.Result, len(profiles))
	for _, profile := range profiles {
		// Fail if s3 secret credentials don't match expected testdata values.
		if profile.AccessKey != helpers.AccessKey || profile.SecretKey != helpers.SecretKey {
			results <- s3.Result{ProfileName: profile.Name, Err: errors.New("invalid credentials")}
		} else {
			results <- s3.Result{ProfileName: profile.Name, Err: nil}
		}
	}
	close(results)
	return results
}
