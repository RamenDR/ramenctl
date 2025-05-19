// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"errors"
	"fmt"

	"github.com/ramendr/ramen/e2e/deployers"
	"github.com/ramendr/ramen/e2e/types"
	"github.com/ramendr/ramen/e2e/workloads"
	"golang.org/x/sync/errgroup"

	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/e2e"
	"github.com/ramendr/ramenctl/pkg/time"
)

// Test perform DR opetaions for testing DR flow.
type Test struct {
	*Context
	Backend     e2e.Testing
	Status      Status
	Config      *types.TestConfig
	Steps       []*Step
	Duration    float64
	stepStarted time.Time
}

// newTest creates a test from test configuration and command context.
func newTest(tc types.TestConfig, cmd *Command) *Test {
	pvcSpec, ok := cmd.PVCSpecs[tc.PVCSpec]
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
		Backend: cmd.Backend,
		Status:  Passed,
		Config:  &tc,
	}
}

func (t *Test) Deploy() bool {
	t.startStep("deploy")
	if err := t.Backend.Deploy(t.Context); err != nil {
		return t.failStep(err)
	}
	console.Pass("Application %q deployed", t.Name())
	return t.passStep()
}

func (t *Test) Undeploy() bool {
	t.startStep("undeploy")
	if err := t.Backend.Undeploy(t.Context); err != nil {
		return t.failStep(err)
	}
	console.Pass("Application %q undeployed", t.Name())
	return t.passStep()
}

func (t *Test) Protect() bool {
	t.startStep("protect")
	if err := t.Backend.Protect(t.Context); err != nil {
		return t.failStep(err)
	}
	console.Pass("Application %q protected", t.Name())
	return t.passStep()
}

func (t *Test) Unprotect() bool {
	t.startStep("unprotect")
	if err := t.Backend.Unprotect(t.Context); err != nil {
		return t.failStep(err)
	}
	console.Pass("Application %q unprotected", t.Name())
	return t.passStep()
}

func (t *Test) Failover() bool {
	t.startStep("failover")
	if err := t.Backend.Failover(t.Context); err != nil {
		return t.failStep(err)
	}
	console.Pass("Application %q failed over", t.Name())
	return t.passStep()
}

func (t *Test) Relocate() bool {
	t.startStep("relocate")
	if err := t.Backend.Relocate(t.Context); err != nil {
		return t.failStep(err)
	}
	console.Pass("Application %q relocated", t.Name())
	return t.passStep()
}

func (t *Test) Cleanup() bool {
	var g errgroup.Group

	t.startStep("cleanup")
	g.Go(func() error {
		if err := t.Backend.Unprotect(t.Context); err != nil {
			return err
		}
		return nil
	})
	g.Go(func() error {
		if err := t.Backend.Undeploy(t.Context); err != nil {
			return err
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return t.failStep(err)
	}
	console.Pass("Application %q cleaned up", t.Name())
	return t.passStep()
}

func (t *Test) startStep(name string) {
	step := &Step{Name: name}
	t.stepStarted = time.Now()
	t.Steps = append(t.Steps, step)
	t.Logger().Infof("Step %q started", step.Name)
}

func (t *Test) failStep(err error) bool {
	step := t.Steps[len(t.Steps)-1]
	step.Duration = time.Since(t.stepStarted).Seconds()
	t.Duration += step.Duration
	if errors.Is(err, context.Canceled) {
		step.Status = Canceled
		console.Error("Canceled application %q %s", t.Name(), step.Name)
	} else {
		step.Status = Failed
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
	step.Status = Passed
	t.Logger().Infof("Step %q passed", step.Name)
	return true
}
