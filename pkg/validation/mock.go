// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"github.com/ramendr/ramen/e2e/types"
	"github.com/ramendr/ramenctl/pkg/gather"
)

type ContextFunc func(Context) error

// Mock implements the Validation interface. All operations succeed without accessing the clusters.
// To cause operations to fail or return non default values, set a function returning an error.
type Mock struct {
	ValidateFunc              ContextFunc
	ApplicationNamespacesFunc func(ctx Context, drpcName, drpcNamespace string) ([]string, error)
	GatherFunc                func(ctx Context, clsuters []*types.Cluster, namespaces []string, outputDir string) <-chan gather.Result
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
	namespaces []string,
	outputDir string,
) <-chan gather.Result {
	if m.GatherFunc != nil {
		return m.GatherFunc(ctx, clusters, namespaces, outputDir)
	}

	results := make(chan gather.Result, len(clusters))
	for _, cluster := range clusters {
		results <- gather.Result{Name: cluster.Name, Err: nil}
	}
	close(results)
	return results
}
