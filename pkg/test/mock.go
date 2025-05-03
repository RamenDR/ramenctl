// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/e2e"
)

type ContextFunc func(types.Context) error
type TestContextFunc func(types.TestContext) error

// MockBackend implements the e2e.Testing interface. All operations succeed without accessing the
// clusters. To cause operations to fail, set a function returning an error.
type MockBackend struct {
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

var _ e2e.Testing = &MockBackend{}

func (m *MockBackend) Validate(ctx types.Context) error {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(ctx)
	}
	return nil
}

func (m *MockBackend) Setup(ctx types.Context) error {
	if m.SetupFunc != nil {
		return m.SetupFunc(ctx)
	}
	return nil
}

func (m *MockBackend) Cleanup(ctx types.Context) error {
	if m.CleanupFunc != nil {
		return m.CleanupFunc(ctx)
	}
	return nil
}

func (m *MockBackend) Deploy(ctx types.TestContext) error {
	if m.DeployFunc != nil {
		return m.DeployFunc(ctx)
	}
	return nil
}

func (m *MockBackend) Undeploy(ctx types.TestContext) error {
	if m.UndeployFunc != nil {
		return m.UndeployFunc(ctx)
	}
	return nil
}

func (m *MockBackend) Protect(ctx types.TestContext) error {
	if m.ProtectFunc != nil {
		return m.ProtectFunc(ctx)
	}
	return nil
}

func (m *MockBackend) Unprotect(ctx types.TestContext) error {
	if m.UnprotectFunc != nil {
		return m.UnprotectFunc(ctx)
	}
	return nil
}

func (m *MockBackend) Failover(ctx types.TestContext) error {
	if m.FailoverFunc != nil {
		return m.FailoverFunc(ctx)
	}
	return nil
}

func (m *MockBackend) Relocate(ctx types.TestContext) error {
	if m.RelocateFunc != nil {
		return m.RelocateFunc(ctx)
	}
	return nil
}
