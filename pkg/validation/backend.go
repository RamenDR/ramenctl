// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"

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

// GatherS3 collects S3 data for the application from all configured S3 profiles.
// It retrieves S3 profiles from the ramen hub config, determines the application's
// S3 prefix based on DRPC, and downloads objects from each profile's bucket.
func (b Backend) GatherS3(
	ctx Context,
	reader gathering.OutputReader,
	drpcName, drpcNamespace, outputDir string,
) error {
	configMapNamespace := ctx.Config().Namespaces.RamenHubNamespace
	s3Profiles, err := ramen.GetS3Profiles(reader, configMapNamespace)
	if err != nil {
		return fmt.Errorf("failed to get ramen S3 profiles: %w", err)
	}
	prefix, err := ramen.ApplicationS3Prefix(reader, drpcName, drpcNamespace)
	if err != nil {
		return err
	}
	for _, profile := range s3Profiles {
		if err := s3.Gather(ctx.Context(), profile, prefix, outputDir, ctx.Logger()); err != nil {
			return fmt.Errorf("failed to gather S3 data from profile %q: %w",
				profile.Name, err)
		}
	}
	return nil
}
