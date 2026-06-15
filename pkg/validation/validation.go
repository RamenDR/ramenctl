// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"context"

	"github.com/ramendr/ramen/e2e/types"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/s3"
)

// Context is validation context, decoupling the ramenctl command from the backend implementation.
type Context interface {
	Env() *types.Env
	Config() *config.Config
	Logger() *zap.SugaredLogger
	Context() context.Context
}

// Validation provides the validation operations.
type Validation interface {
	Validate(ctx Context) error
	ApplicationNamespaces(ctx Context, drpcName, drpcNamespace string) ([]string, error)
	Gather(
		ctx Context,
		clusters []*types.Cluster,
		options gathering.Options,
	) <-chan gathering.Result
	GetSecret(ctx Context, cluster *types.Cluster, name, namespace string) (*corev1.Secret, error)
	GatherS3(
		ctx Context,
		profiles []*s3.Profile,
		prefixes []string,
		outputDir string,
	) <-chan s3.Result
	CheckS3(ctx Context, profiles []*s3.Profile) <-chan s3.Result
}
