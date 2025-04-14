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
	Status Status
	Config *types.TestConfig
	Steps  []*Step
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
		Status:  Passed,
		Config:  &tc,
	}
}

func (t *Test) Deploy() bool {
	step := &Step{Name: "deploy"}
	if err := t.Deployer().Deploy(t.Context); err != nil {
		msg := fmt.Sprintf("failed to deploy application %q", t.Name())
		return t.fail(step, msg, err)
	}
	msg := fmt.Sprintf("Application %q deployed", t.Name())
	return t.pass(step, msg)
}

func (t *Test) Undeploy() bool {
	step := &Step{Name: "undeploy"}
	if err := t.Deployer().Undeploy(t.Context); err != nil {
		msg := fmt.Sprintf("failed to undeploy application %q", t.Name())
		return t.fail(step, msg, err)
	}
	msg := fmt.Sprintf("Application %q undeployed", t.Name())
	return t.pass(step, msg)
}

func (t *Test) Protect() bool {
	step := &Step{Name: "protect"}
	if err := dractions.EnableProtection(t.Context); err != nil {
		msg := fmt.Sprintf("failed to protect application %q", t.Name())
		return t.fail(step, msg, err)
	}
	msg := fmt.Sprintf("Application %q protected", t.Name())
	return t.pass(step, msg)
}

func (t *Test) Unprotect() bool {
	step := &Step{Name: "unprotect"}
	if err := dractions.DisableProtection(t.Context); err != nil {
		msg := fmt.Sprintf("failed to unprotect application %q", t.Name())
		return t.fail(step, msg, err)
	}
	msg := fmt.Sprintf("Application %q unprotected", t.Name())
	return t.pass(step, msg)
}

func (t *Test) Failover() bool {
	step := &Step{Name: "failover"}
	if err := dractions.Failover(t.Context); err != nil {
		msg := fmt.Sprintf("failed to failover application %q", t.Name())
		return t.fail(step, msg, err)
	}
	msg := fmt.Sprintf("Application %q failed over", t.Name())
	return t.pass(step, msg)
}

func (t *Test) Relocate() bool {
	step := &Step{Name: "relocate"}
	if err := dractions.Relocate(t.Context); err != nil {
		msg := fmt.Sprintf("failed to relocate application %q", t.Name())
		return t.fail(step, msg, err)
	}
	msg := fmt.Sprintf("Application %q relocated", t.Name())
	return t.pass(step, msg)
}

func (t *Test) fail(step *Step, msg string, err error) bool {
	step.Status = Failed
	t.Steps = append(t.Steps, step)
	t.Status = Failed
	console.Error(msg)
	t.Logger().Errorf("%s: %s", msg, err)
	return false
}

func (t *Test) pass(step *Step, msg string) bool {
	step.Status = Passed
	t.Steps = append(t.Steps, step)
	console.Pass(msg)
	return true
}
