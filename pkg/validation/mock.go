// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/s3"
)

type ContextFunc func(Context) error

// Mock implements the Validation interface. All operations succeed without accessing the clusters.
// To cause operations to fail or return non default values, set a function returning an error.
type Mock struct {
	ValidateFunc              ContextFunc
	ApplicationNamespacesFunc func(ctx Context, drpcName, drpcNamespace string) ([]string, error)
	GatherFunc                func(ctx Context, clsuters []*types.Cluster, options gathering.Options) <-chan gathering.Result
	GatherS3Func              func(ctx Context, profiles []*s3.Profile, prefixes []string, outputDir string) <-chan s3.Result
	CheckS3Func               func(ctx Context, profiles []*s3.Profile) <-chan s3.Result
}

var _ Validation = &Mock{}

func (m *Mock) Validate(ctx Context) error {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(ctx)
	}
	return nil
}

func (m *Mock) ApplicationNamespaces(
	ctx Context,
	drpcName, drpcNamespace string,
) ([]string, error) {
	if m.ApplicationNamespacesFunc != nil {
		return m.ApplicationNamespacesFunc(ctx, drpcName, drpcNamespace)
	}
	return nil, nil
}

func (m *Mock) Gather(
	ctx Context,
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

func (m *Mock) GatherS3(
	ctx Context,
	profiles []*s3.Profile,
	prefixes []string,
	outputDir string,
) <-chan s3.Result {
	if m.GatherS3Func != nil {
		return m.GatherS3Func(ctx, profiles, prefixes, outputDir)
	}
	results := make(chan s3.Result, len(profiles))
	for _, profile := range profiles {
		results <- s3.Result{ProfileName: profile.Name, Err: nil}
	}
	close(results)
	return results
}

func (m *Mock) CheckS3(ctx Context, profiles []*s3.Profile) <-chan s3.Result {
	if m.CheckS3Func != nil {
		return m.CheckS3Func(ctx, profiles)
	}
	results := make(chan s3.Result, len(profiles))
	for _, profile := range profiles {
		results <- s3.Result{ProfileName: profile.Name, Err: nil}
	}
	close(results)
	return results
}
