// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"context"

	"github.com/ramendr/ramen/e2e/types"
	"go.uber.org/zap"

	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/gathering"
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
		namespaces []string,
		outputDir string,
	) <-chan gathering.Result
}
