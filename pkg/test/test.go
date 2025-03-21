// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"

	"github.com/ramendr/ramen/e2e/deployers"
	"github.com/ramendr/ramen/e2e/dractions"
	"github.com/ramendr/ramen/e2e/types"
	"github.com/ramendr/ramen/e2e/workloads"

	"github.com/ramendr/ramenctl/pkg/console"
)

// Test perform DR opetaions for testing DR flow.
type Test struct {
	types.Context
	Config types.TestConfig
	Status Status
}

// newTest creates a test from test configuration and command context.
func newTest(tc types.TestConfig, cmd *Command) *Test {
	pvcSpec, ok := cmd.PVCSpecs[tc.PVCSpec]
	if !ok {
		panic(fmt.Sprintf("unknown pvcSpec %q", tc.PVCSpec))
	}

	workload, err := workloads.New(tc.Workload, cmd.Config.Repo.Branch, pvcSpec)
	if err != nil {
		panic(err)
	}

	deployer, err := deployers.New(tc.Deployer)
	if err != nil {
		panic(err)
	}

	return &Test{
		Context: newContext(workload, deployer, cmd),
		Config:  tc,
		Status:  Passed,
	}
}

func (t *Test) Deploy() bool {
	if err := t.Deployer().Deploy(t.Context); err != nil {
		err := fmt.Errorf("failed to deploy application %q: %w", t.Name(), err)
		t.Fail(err)
		return false
	}
	console.Completed("Application %q deployed", t.Name())
	return true
}

func (t *Test) Undeploy() bool {
	if err := t.Deployer().Undeploy(t.Context); err != nil {
		err := fmt.Errorf("failed to undeploy application %q: %w", t.Name(), err)
		t.Fail(err)
		return false
	}
	console.Completed("Application %q undeployed", t.Name())
	return true
}

func (t *Test) Protect() bool {
	if err := dractions.EnableProtection(t.Context); err != nil {
		err := fmt.Errorf("failed to protect application %q: %w", t.Name(), err)
		t.Fail(err)
		return false
	}
	console.Completed("Application %q protected", t.Name())
	return true
}

func (t *Test) Unprotect() bool {
	if err := dractions.DisableProtection(t.Context); err != nil {
		err := fmt.Errorf("failed to unprotect application %q: %w", t.Name(), err)
		t.Fail(err)
		return false
	}
	console.Completed("Application %q unprotected", t.Name())
	return true
}

func (t *Test) Failover() bool {
	if err := dractions.Failover(t.Context); err != nil {
		err := fmt.Errorf("failed to failover application %q: %w", t.Name(), err)
		t.Fail(err)
		return false
	}
	console.Completed("Application %q failed over", t.Name())
	return true
}

func (t *Test) Relocate() bool {
	if err := dractions.Relocate(t.Context); err != nil {
		err := fmt.Errorf("failed to relocate application %q: %w", t.Name(), err)
		t.Fail(err)
		return false
	}
	console.Completed("Application %q relocated", t.Name())
	return true
}

func (t *Test) Fail(err error) {
	console.Error(err)
	t.Logger().Error(err)
	t.Status = Failed
}
