// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package testing

import (
	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/gathering"
)

// Testing interface for ramenctl commands.
type Testing interface {
	// Operations on types.Context.
	Validate(types.Context) error
	Setup(types.Context) error
	Cleanup(types.Context) error

	// Operations on types.TestContext.
	Deploy(types.TestContext) error
	Undeploy(types.TestContext) error
	Protect(types.TestContext) error
	Unprotect(types.TestContext) error
	Failover(types.TestContext) error
	Relocate(types.TestContext) error
	Purge(types.TestContext) error

	// Handling failures.
	Gather(
		ctx types.Context,
		clusters []*types.Cluster,
		options gathering.Options,
	) <-chan gathering.Result
}
