// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package testing

import (
	"github.com/ramendr/ramen/e2e/dractions"
	"github.com/ramendr/ramen/e2e/types"
	"github.com/ramendr/ramen/e2e/util"
	"github.com/ramendr/ramen/e2e/validate"
)

type Backend struct{}

var _ Testing = &Backend{}

func (Backend) Validate(ctx types.Context) error {
	return validate.TestConfig(ctx)
}

func (Backend) Setup(ctx types.Context) error {
	return util.EnsureChannel(ctx)
}

func (Backend) Cleanup(ctx types.Context) error {
	return util.EnsureChannelDeleted(ctx)
}

func (Backend) Deploy(ctx types.TestContext) error {
	return ctx.Deployer().Deploy(ctx)
}

func (Backend) Undeploy(ctx types.TestContext) error {
	return ctx.Deployer().Undeploy(ctx)
}

func (Backend) Protect(ctx types.TestContext) error {
	return dractions.EnableProtection(ctx)
}

func (Backend) Unprotect(ctx types.TestContext) error {
	return dractions.DisableProtection(ctx)
}

func (Backend) Failover(ctx types.TestContext) error {
	return dractions.Failover(ctx)
}

func (Backend) Relocate(ctx types.TestContext) error {
	return dractions.Relocate(ctx)
}

func (Backend) Purge(ctx types.TestContext) error {
	return dractions.Purge(ctx)
}
