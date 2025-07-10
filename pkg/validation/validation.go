// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"context"

	"github.com/ramendr/ramen/e2e/types"
	"go.uber.org/zap"

	"github.com/ramendr/ramenctl/pkg/config"
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
	Validate(Context) error
	ApplicationNamespaces(Context, string, string) ([]string, error)
}
