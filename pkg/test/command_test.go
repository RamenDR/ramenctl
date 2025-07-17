// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	e2econfig "github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/report"
	rtesting "github.com/ramendr/ramenctl/pkg/testing"
)

const (
	testRun   = "test-run"
	testClean = "test-clean"
)

var (
	testConfig = &e2econfig.Config{
		PVCSpecs: []e2econfig.PVCSpec{
			{Name: "block", StorageClassName: "block-storage"},
			{Name: "file", StorageClassName: "file-storage"},
		},
		Tests: []e2econfig.Test{
			{Workload: "deploy", Deployer: "appset", PVCSpec: "block"},
			{Workload: "deploy", Deployer: "appset", PVCSpec: "file"},
			{Workload: "deploy", Deployer: "subscr", PVCSpec: "block"},
			{Workload: "deploy", Deployer: "subscr", PVCSpec: "file"},
			{Workload: "deploy", Deployer: "disapp", PVCSpec: "block"},
			{Workload: "deploy", Deployer: "disapp", PVCSpec: "file"},
		},
	}

	testEnv = &types.Env{
		Hub: &types.Cluster{Name: "hub"},
		C1:  &types.Cluster{Name: "c1"},
		C2:  &types.Cluster{Name: "c2"},
	}

	validateFailed = &rtesting.Mock{
		ValidateFunc: func(ctx types.Context) error {
			return errors.New("No validate for you!")
		},
	}

	validateCanceled = &rtesting.Mock{
		ValidateFunc: func(ctx types.Context) error {
			return context.Canceled
		},
	}

	setupFailed = &rtesting.Mock{
		SetupFunc: func(ctx types.Context) error {
			return errors.New("No setup for you!")
		},
	}

	setupCanceled = &rtesting.Mock{
		SetupFunc: func(ctx types.Context) error {
			return context.Canceled
		},
	}

	cleanupFailed = &rtesting.Mock{
		CleanupFunc: func(ctx types.Context) error {
			return errors.New("No cleanup for you!")
		},
	}

	cleanupCanceled = &rtesting.Mock{
		CleanupFunc: func(ctx types.Context) error {
			return context.Canceled
		},
	}

	failoverFailed = &rtesting.Mock{
		FailoverFunc: func(ctx types.TestContext) error {
			return errors.New("No failover for you!")
		},
	}

	failoverCanceled = &rtesting.Mock{
		FailoverFunc: func(ctx types.TestContext) error {
			return context.Canceled
		},
	}

	disappFailoverFailed = &rtesting.Mock{
		FailoverFunc: func(ctx types.TestContext) error {
			if ctx.Deployer().IsDiscovered() {
				return errors.New("No failover for you!")
			}
			return nil
		},
	}

	purgeFailed = &rtesting.Mock{
		PurgeFunc: func(ctx types.TestContext) error {
			return errors.New("No purge for you!")
		},
	}

	purgeCanceled = &rtesting.Mock{
		PurgeFunc: func(ctx types.TestContext) error {
			return context.Canceled
		},
	}

	runFlow = []string{"deploy", "protect", "failover", "relocate", "unprotect", "undeploy"}
)

func TestRunPassed(t *testing.T) {
	test := testCommand(t, testRun, &rtesting.Mock{})

	if err := test.Run(); err != nil {
		t.Fatal(err)
	}

	checkReport(t, test.report, report.Passed, Summary{Passed: len(testConfig.Tests)})
	if len(test.report.Steps) != 3 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, ValidateStep, report.Passed)
	setup := test.report.Steps[1]
	checkStep(t, setup, SetupStep, report.Passed)
	tests := test.report.Steps[2]
	checkStep(t, tests, TestsStep, report.Passed)
	for i, tc := range testConfig.Tests {
		result := tests.Items[i]
		checkTest(t, result, tc, report.Passed, runFlow...)
	}
}

func TestRunValidateFailed(t *testing.T) {
	test := testCommand(t, testRun, validateFailed)

	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Failed, Summary{})
	if len(test.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, ValidateStep, report.Failed)
}

func TestRunValidateCanceled(t *testing.T) {
	test := testCommand(t, testRun, validateCanceled)

	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Canceled, Summary{})
	if len(test.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, ValidateStep, report.Canceled)
}

func TestRunSetupFailed(t *testing.T) {
	test := testCommand(t, testRun, setupFailed)

	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Failed, Summary{})
	if len(test.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, ValidateStep, report.Passed)
	setup := test.report.Steps[1]
	checkStep(t, setup, SetupStep, report.Failed)
}

func TestRunSetupCanceled(t *testing.T) {
	test := testCommand(t, testRun, setupCanceled)

	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Canceled, Summary{})
	if len(test.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, ValidateStep, report.Passed)
	setup := test.report.Steps[1]
	checkStep(t, setup, SetupStep, report.Canceled)
}

func TestRunTestsFailed(t *testing.T) {
	test := testCommand(t, testRun, failoverFailed)

	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Failed, Summary{Failed: len(testConfig.Tests)})
	if len(test.report.Steps) != 3 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, ValidateStep, report.Passed)
	setup := test.report.Steps[1]
	checkStep(t, setup, SetupStep, report.Passed)
	tests := test.report.Steps[2]
	checkStep(t, tests, TestsStep, report.Failed)
	for i, tc := range testConfig.Tests {
		result := tests.Items[i]
		checkTest(t, result, tc, report.Failed, "deploy", "protect", "failover")
	}
}

func TestRunDisappFailed(t *testing.T) {
	test := testCommand(t, testRun, disappFailoverFailed)

	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Failed, Summary{Passed: 4, Failed: 2})
	if len(test.report.Steps) != 3 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, ValidateStep, report.Passed)
	setup := test.report.Steps[1]
	checkStep(t, setup, SetupStep, report.Passed)
	tests := test.report.Steps[2]
	checkStep(t, tests, TestsStep, report.Failed)
	for i, tc := range testConfig.Tests {
		result := tests.Items[i]
		if tc.Deployer == "disapp" {
			checkTest(t, result, tc, report.Failed, "deploy", "protect", "failover")
		} else {
			checkTest(t, result, tc, report.Passed, runFlow...)
		}
	}
}

func TestRunTestsCanceled(t *testing.T) {
	test := testCommand(t, testRun, failoverCanceled)

	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Canceled, Summary{Canceled: len(testConfig.Tests)})
	if len(test.report.Steps) != 3 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, ValidateStep, report.Passed)
	setup := test.report.Steps[1]
	checkStep(t, setup, SetupStep, report.Passed)
	tests := test.report.Steps[2]
	checkStep(t, tests, TestsStep, report.Canceled)
	for i, tc := range testConfig.Tests {
		result := tests.Items[i]
		checkTest(t, result, tc, report.Canceled, "deploy", "protect", "failover")
	}
}

func TestCleanPassed(t *testing.T) {
	test := testCommand(t, testClean, &rtesting.Mock{})

	if err := test.Clean(); err != nil {
		t.Fatal(err)
	}

	checkReport(t, test.report, report.Passed, Summary{Passed: 6})
	if len(test.report.Steps) != 3 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, ValidateStep, report.Passed)
	tests := test.report.Steps[1]
	checkStep(t, tests, TestsStep, report.Passed)
	for i, tc := range testConfig.Tests {
		result := tests.Items[i]
		checkTest(t, result, tc, report.Passed, "purge")
	}
	cleanup := test.report.Steps[2]
	checkStep(t, cleanup, CleanupStep, report.Passed)
}

func TestCleanValidateFailed(t *testing.T) {
	test := testCommand(t, testClean, validateFailed)

	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Failed, Summary{})
	if len(test.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, ValidateStep, report.Failed)
}

func TestCleanValidateCanceled(t *testing.T) {
	test := testCommand(t, testClean, validateCanceled)

	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Canceled, Summary{})
	if len(test.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, ValidateStep, report.Canceled)
}

func TestCleanPurgeFailed(t *testing.T) {
	test := testCommand(t, testClean, purgeFailed)

	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Failed, Summary{Failed: len(testConfig.Tests)})
	if len(test.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, ValidateStep, report.Passed)
	tests := test.report.Steps[1]
	checkStep(t, tests, TestsStep, report.Failed)
	for i, tc := range testConfig.Tests {
		result := tests.Items[i]
		checkTest(t, result, tc, report.Failed, "purge")
	}
}

func TestCleanPurgeCanceled(t *testing.T) {
	test := testCommand(t, testClean, purgeCanceled)

	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Canceled, Summary{Canceled: len(testConfig.Tests)})
	if len(test.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, ValidateStep, report.Passed)
	tests := test.report.Steps[1]
	checkStep(t, tests, TestsStep, report.Canceled)
	for i, tc := range testConfig.Tests {
		result := tests.Items[i]
		checkTest(t, result, tc, report.Canceled, "purge")
	}
}

func TestCleanCleanupFailed(t *testing.T) {
	test := testCommand(t, testClean, cleanupFailed)

	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Failed, Summary{Passed: len(testConfig.Tests)})
	if len(test.report.Steps) != 3 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, ValidateStep, report.Passed)
	tests := test.report.Steps[1]
	checkStep(t, tests, TestsStep, report.Passed)
	for i, tc := range testConfig.Tests {
		result := tests.Items[i]
		checkTest(t, result, tc, report.Passed, "purge")
	}
	cleanup := test.report.Steps[2]
	checkStep(t, cleanup, CleanupStep, report.Failed)
}

func TestCleanCleanupCanceled(t *testing.T) {
	test := testCommand(t, testClean, cleanupCanceled)

	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Canceled, Summary{Passed: len(testConfig.Tests)})
	if len(test.report.Steps) != 3 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, ValidateStep, report.Passed)
	tests := test.report.Steps[1]
	checkStep(t, tests, TestsStep, report.Passed)
	for i, tc := range testConfig.Tests {
		result := tests.Items[i]
		checkTest(t, result, tc, report.Passed, "purge")
	}
	cleanup := test.report.Steps[2]
	checkStep(t, cleanup, CleanupStep, report.Canceled)
}

// Test helpers.

func testCommand(t *testing.T, name string, backend rtesting.Testing) *Command {
	cmd, err := command.ForTest(name, testEnv, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cmd.Close()
	})
	return newCommand(cmd, testConfig, backend)
}

func checkReport(t *testing.T, report *Report, status report.Status, summary Summary) {
	if report.Status != status {
		t.Errorf("expected status %q, got %q", status, report.Status)
	}
	if report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, report.Summary)
	}
	duration := totalDuration(report.Steps)
	if report.Duration != duration {
		t.Fatalf("expected duration %v, got %v", duration, report.Duration)
	}
}

func checkStep(t *testing.T, step *report.Step, name string, status report.Status) {
	if name != step.Name {
		t.Fatalf("expected step %q, got %q", name, step.Name)
	}
	if status != step.Status {
		t.Fatalf("expected status %q, got %q", status, step.Status)
	}
	// We cannot check duration since it may be zero on windows.
}

func checkTest(
	t *testing.T,
	test *report.Step,
	tc e2econfig.Test,
	status report.Status,
	flow ...string,
) {
	name := fmt.Sprintf("%s-%s-%s", tc.Deployer, tc.Workload, tc.PVCSpec)
	if name != test.Name {
		t.Fatalf("expected step %q, got %q", name, test.Name)
	}
	if status != test.Status {
		t.Fatalf("expected status %q, got %q", status, test.Status)
	}
	duration := totalDuration(test.Items)
	if test.Duration != duration {
		t.Fatalf("expected duration %v, got %v", duration, test.Duration)
	}
	if len(flow) != len(test.Items) {
		t.Fatalf("test %q steps %+v do not match flow %q", test.Name, test.Items, flow)
	}
	last := len(flow) - 1
	for i, name := range flow[:last] {
		checkStep(t, test.Items[i], name, report.Passed)
	}
	checkStep(t, test.Items[last], flow[last], test.Status)
}

func totalDuration(steps []*report.Step) float64 {
	var total float64
	for _, step := range steps {
		total += step.Duration
	}
	return total
}
