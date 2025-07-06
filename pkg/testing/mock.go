// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package testing

import (
	"github.com/ramendr/ramen/e2e/types"
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
