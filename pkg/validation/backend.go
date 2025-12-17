// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/ramen"
	"github.com/ramendr/ramenctl/pkg/s3"
)

// Backend performs validation with real clusters.
type Backend struct{}

var _ Validation = &Backend{}

// Validate the environment. Must be called once before calling other functions.
func (b Backend) Validate(ctx Context) error {
	if err := detectDistro(ctx); err != nil {
		return err
	}
	if err := validateClusterset(ctx); err != nil {
		return err
	}
	return nil
}

// ApplicationNamespaces inspects the application DRPC and returns the application namespaces on the
// hub and managed clusters.
func (b Backend) ApplicationNamespaces(
	ctx Context,
	drpcName, drpcNamespace string,
) ([]string, error) {
	drpc, err := ramen.GetDRPC(ctx, drpcName, drpcNamespace)
	if err != nil {
		return nil, err
	}
	return ramen.ApplicationNamespaces(drpc), nil
}

func (b Backend) Gather(
	ctx Context,
	clusters []*types.Cluster,
	options gathering.Options,
) <-chan gathering.Result {
	return gathering.Namespaces(ctx, clusters, options)
}

func (b Backend) GatherS3(
	ctx Context,
	profiles []*s3.Profile,
	prefixes []string,
	outputDir string,
) <-chan s3.Result {
	return s3.Gather(ctx.Context(), profiles, prefixes, outputDir, ctx.Logger())
}

func (b Backend) CheckS3(ctx Context, profiles []*s3.Profile) <-chan s3.Result {
	return s3.Check(ctx.Context(), profiles, ctx.Logger())
}
