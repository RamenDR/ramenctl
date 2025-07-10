// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"errors"
	"fmt"

	e2econfig "github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/deployers"
	"github.com/ramendr/ramen/e2e/util"
	"github.com/ramendr/ramen/e2e/workloads"

	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/testing"
	"github.com/ramendr/ramenctl/pkg/time"
)

// Test perform DR opetaions for testing DR flow.
type Test struct {
	*Context
	Backend     testing.Testing
	Status      report.Status
	Steps       []*report.Step
	Duration    float64
	stepStarted time.Time
}

// newTest creates a test from test configuration and command context.
func newTest(tc e2econfig.Test, cmd *Command) *Test {
	pvcSpec, ok := cmd.pvcSpecs[tc.PVCSpec]
	if !ok {
		panic(fmt.Sprintf("unknown pvcSpec %q", tc.PVCSpec))
	}

	workload, err := workloads.New(tc.Workload, cmd.Config().Repo.Branch, pvcSpec)
	if err != nil {
		panic(err)
	}

	deployer, err := deployers.New(tc.Deployer)
	if err != nil {
		panic(err)
	}

	return &Test{
		Context: newContext(cmd, workload, deployer),
		Backend: cmd.backend,
		Status:  report.Passed,
	}
}

func (t *Test) Deploy() bool {
	t.startStep("deploy")
	timedCtx, cancel := t.WithTimeout(util.DeployTimeout)
	defer cancel()
	if err := t.Backend.Deploy(timedCtx); err != nil {
		return t.failStep(err)
	}
	console.Pass("Application %q deployed", t.Name())
	return t.passStep()
}

func (t *Test) Undeploy() bool {
	t.startStep("undeploy")
	timedCtx, cancel := t.WithTimeout(util.UndeployTimeout)
	defer cancel()
	if err := t.Backend.Undeploy(timedCtx); err != nil {
		return t.failStep(err)
	}
	console.Pass("Application %q undeployed", t.Name())
	return t.passStep()
}

func (t *Test) Protect() bool {
	t.startStep("protect")
	timedCtx, cancel := t.WithTimeout(util.EnableTimeout)
	defer cancel()
	if err := t.Backend.Protect(timedCtx); err != nil {
		return t.failStep(err)
	}
	console.Pass("Application %q protected", t.Name())
	return t.passStep()
}

func (t *Test) Unprotect() bool {
	t.startStep("unprotect")
	timedCtx, cancel := t.WithTimeout(util.DisableTimeout)
	defer cancel()
	if err := t.Backend.Unprotect(timedCtx); err != nil {
		return t.failStep(err)
	}
	console.Pass("Application %q unprotected", t.Name())
	return t.passStep()
}

func (t *Test) Failover() bool {
	t.startStep("failover")
	timedCtx, cancel := t.WithTimeout(util.FailoverTimeout)
	defer cancel()
	if err := t.Backend.Failover(timedCtx); err != nil {
		return t.failStep(err)
	}
	console.Pass("Application %q failed over", t.Name())
	return t.passStep()
}

func (t *Test) Relocate() bool {
	t.startStep("relocate")
	timedCtx, cancel := t.WithTimeout(util.RelocateTimeout)
	defer cancel()
	if err := t.Backend.Relocate(timedCtx); err != nil {
		return t.failStep(err)
	}
	console.Pass("Application %q relocated", t.Name())
	return t.passStep()
}

func (t *Test) Purge() bool {
	t.startStep("purge")
	timedCtx, cancel := t.WithTimeout(util.PurgeTimeout)
	defer cancel()
	if err := t.Backend.Purge(timedCtx); err != nil {
		return t.failStep(err)
	}
	console.Pass("Application %q cleaned up", t.Name())
	return t.passStep()
}

func (t *Test) startStep(name string) {
	step := &report.Step{Name: name}
	t.stepStarted = time.Now()
	t.Steps = append(t.Steps, step)
	t.Logger().Infof("Step %q started", step.Name)
}

func (t *Test) failStep(err error) bool {
	step := t.Steps[len(t.Steps)-1]
	step.Duration = time.Since(t.stepStarted).Seconds()
	t.Duration += step.Duration
	if errors.Is(err, context.Canceled) {
		step.Status = report.Canceled
		console.Error("Canceled application %q %s", t.Name(), step.Name)
	} else {
		step.Status = report.Failed
		console.Error("Failed to %s application %q", step.Name, t.Name())
	}
	t.Status = step.Status
	t.Logger().Errorf("Step %q %s: %s", step.Name, step.Status, err)
	return false
}

func (t *Test) passStep() bool {
	step := t.Steps[len(t.Steps)-1]
	step.Duration = time.Since(t.stepStarted).Seconds()
	t.Duration += step.Duration
	step.Status = report.Passed
	t.Logger().Infof("Step %q passed", step.Name)
	return true
}
