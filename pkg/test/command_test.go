// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	e2econfig "github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
	rtesting "github.com/ramendr/ramenctl/pkg/testing"
)

const (
	applicationTestdata = "../testdata/appset-deploy-rbd"
	testRun             = "test-run"
	testClean           = "test-clean"
)

var (
	testConfig = &e2econfig.Config{
		Namespaces: e2econfig.K8sNamespaces,
		PVCSpecs: []e2econfig.PVCSpec{
			{Name: "rbd", StorageClassName: "block-storage"},
			{Name: "cephfs", StorageClassName: "file-storage"},
		},
		Deployers: []e2econfig.Deployer{
			{Name: "appset", Type: "appset"},
			{Name: "subscr", Type: "subscr"},
			{Name: "disapp", Type: "disapp"},
			{Name: "disapp-recipe", Type: "disapp", Recipe: &e2econfig.Recipe{Type: "generate"}},
			{
				Name:   "disapp-recipe-check",
				Type:   "disapp",
				Recipe: &e2econfig.Recipe{Type: "generate", CheckHook: true},
			},
			{
				Name:   "disapp-recipe-exec",
				Type:   "disapp",
				Recipe: &e2econfig.Recipe{Type: "generate", ExecHook: true},
			},
			{
				Name:   "disapp-recipe-check-exec",
				Type:   "disapp",
				Recipe: &e2econfig.Recipe{Type: "generate", CheckHook: true, ExecHook: true},
			},
		},
		Tests: []e2econfig.Test{
			{Workload: "deploy", Deployer: "appset", PVCSpec: "rbd"},
			{Workload: "deploy", Deployer: "appset", PVCSpec: "cephfs"},
			{Workload: "deploy", Deployer: "subscr", PVCSpec: "rbd"},
			{Workload: "deploy", Deployer: "subscr", PVCSpec: "cephfs"},
			{Workload: "deploy", Deployer: "disapp", PVCSpec: "rbd"},
			{Workload: "deploy", Deployer: "disapp", PVCSpec: "cephfs"},
			{Workload: "deploy", Deployer: "disapp-recipe", PVCSpec: "rbd"},
			{Workload: "deploy", Deployer: "disapp-recipe", PVCSpec: "cephfs"},
			{Workload: "deploy", Deployer: "disapp-recipe-check", PVCSpec: "rbd"},
			{Workload: "deploy", Deployer: "disapp-recipe-check", PVCSpec: "cephfs"},
			{Workload: "deploy", Deployer: "disapp-recipe-exec", PVCSpec: "rbd"},
			{Workload: "deploy", Deployer: "disapp-recipe-exec", PVCSpec: "cephfs"},
			{Workload: "deploy", Deployer: "disapp-recipe-check-exec", PVCSpec: "rbd"},
			{Workload: "deploy", Deployer: "disapp-recipe-check-exec", PVCSpec: "cephfs"},
		},
	}

	testEnv = &types.Env{
		Hub: &types.Cluster{Name: "hub"},
		C1:  &types.Cluster{Name: "c1"},
		C2:  &types.Cluster{Name: "c2"},
	}

	validateFailed = &helpers.TestingMock{
		ValidateFunc: func(ctx types.Context) error {
			return errors.New("No validate for you!")
		},
	}

	validateCanceled = &helpers.TestingMock{
		ValidateFunc: func(ctx types.Context) error {
			return context.Canceled
		},
	}

	setupFailed = &helpers.TestingMock{
		SetupFunc: func(ctx types.Context) error {
			return errors.New("No setup for you!")
		},
	}

	setupCanceled = &helpers.TestingMock{
		SetupFunc: func(ctx types.Context) error {
			return context.Canceled
		},
	}

	cleanupFailed = &helpers.TestingMock{
		CleanupFunc: func(ctx types.Context) error {
			return errors.New("No cleanup for you!")
		},
	}

	cleanupCanceled = &helpers.TestingMock{
		CleanupFunc: func(ctx types.Context) error {
			return context.Canceled
		},
	}

	failoverFailed = &helpers.TestingMock{
		FailoverFunc: func(ctx types.TestContext) error {
			return errors.New("No failover for you!")
		},
	}

	failoverCanceled = &helpers.TestingMock{
		FailoverFunc: func(ctx types.TestContext) error {
			return context.Canceled
		},
	}

	disappFailoverFailed = &helpers.TestingMock{
		FailoverFunc: func(ctx types.TestContext) error {
			if ctx.Deployer().IsDiscovered() {
				return errors.New("No failover for you!")
			}
			return nil
		},
	}

	purgeFailed = &helpers.TestingMock{
		PurgeFunc: func(ctx types.TestContext) error {
			return errors.New("No purge for you!")
		},
	}

	purgeCanceled = &helpers.TestingMock{
		PurgeFunc: func(ctx types.TestContext) error {
			return context.Canceled
		},
	}

	runFlow = []string{"deploy", "protect", "failover", "relocate", "unprotect", "undeploy"}
)

func TestRunPassed(t *testing.T) {
	test := testCommand(t, testRun, &helpers.TestingMock{})

	if err := test.Run(); err != nil {
		t.Fatal(err)
	}

	checkReport(
		t,
		test.report,
		report.Passed,
		report.Summary{Passed: len(testConfig.Tests)},
	)
	checkError(t, test.report, "")
	if len(test.report.Steps) != 3 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, &report.Step{Name: ValidateStep, Status: report.Passed})
	setup := test.report.Steps[1]
	checkStep(t, setup, &report.Step{Name: SetupStep, Status: report.Passed})
	tests := test.report.Steps[2]
	checkStep(t, tests, &report.Step{Name: TestsStep, Status: report.Passed})
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

	checkReport(t, test.report, report.Failed, report.Summary{})
	checkError(t, test.report, "Failed to validate")
	if len(test.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, &report.Step{
		Name:   ValidateStep,
		Status: report.Failed,
		Err:    "Failed to validate",
	})
}

func TestRunValidateCanceled(t *testing.T) {
	test := testCommand(t, testRun, validateCanceled)

	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Canceled, report.Summary{})
	checkError(t, test.report, "Canceled validate")
	if len(test.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, &report.Step{
		Name:   ValidateStep,
		Status: report.Canceled,
		Err:    "Canceled validate",
	})
}

func TestRunSetupFailed(t *testing.T) {
	test := testCommand(t, testRun, setupFailed)

	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Failed, report.Summary{})
	checkError(t, test.report, "Failed to setup")
	if len(test.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, &report.Step{Name: ValidateStep, Status: report.Passed})
	setup := test.report.Steps[1]
	checkStep(t, setup, &report.Step{
		Name:   SetupStep,
		Status: report.Failed,
		Err:    "Failed to setup",
	})
}

func TestRunSetupCanceled(t *testing.T) {
	test := testCommand(t, testRun, setupCanceled)

	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Canceled, report.Summary{})
	checkError(t, test.report, "Canceled setup")
	if len(test.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, &report.Step{Name: ValidateStep, Status: report.Passed})
	setup := test.report.Steps[1]
	checkStep(t, setup, &report.Step{
		Name:   SetupStep,
		Status: report.Canceled,
		Err:    "Canceled setup",
	})
}

func TestRunTestsFailed(t *testing.T) {
	test := testCommand(t, testRun, failoverFailed)
	helpers.AddGatheredData(t, test.dataDir(), applicationTestdata, "validate-application")
	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(
		t,
		test.report,
		report.Failed,
		report.Summary{Failed: len(testConfig.Tests)},
	)
	checkError(t, test.report,
		"Test run failed (0 passed, 14 failed, 0 skipped, 0 canceled)")
	if len(test.report.Steps) != 3 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, &report.Step{Name: ValidateStep, Status: report.Passed})
	setup := test.report.Steps[1]
	checkStep(t, setup, &report.Step{Name: SetupStep, Status: report.Passed})
	tests := test.report.Steps[2]
	checkStep(t, tests, &report.Step{
		Name:   TestsStep,
		Status: report.Failed,
		Err:    "Test run failed (0 passed, 14 failed, 0 skipped, 0 canceled)",
	})
	for i, tc := range testConfig.Tests {
		result := tests.Items[i]
		checkTest(t, result, tc, report.Failed, "deploy", "protect", "failover")
	}
}

func TestRunDisappFailed(t *testing.T) {
	test := testCommand(t, testRun, disappFailoverFailed)
	helpers.AddGatheredData(t, test.dataDir(), applicationTestdata, "validate-application")
	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(
		t,
		test.report,
		report.Failed,
		report.Summary{Passed: 4, Failed: 10},
	)
	checkError(t, test.report,
		"Test run failed (4 passed, 10 failed, 0 skipped, 0 canceled)")
	if len(test.report.Steps) != 3 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, &report.Step{Name: ValidateStep, Status: report.Passed})
	setup := test.report.Steps[1]
	checkStep(t, setup, &report.Step{Name: SetupStep, Status: report.Passed})
	tests := test.report.Steps[2]
	checkStep(t, tests, &report.Step{
		Name:   TestsStep,
		Status: report.Failed,
		Err:    "Test run failed (4 passed, 10 failed, 0 skipped, 0 canceled)",
	})
	for i, tc := range testConfig.Tests {
		result := tests.Items[i]
		if strings.HasPrefix(tc.Deployer, "disapp") {
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

	checkReport(
		t,
		test.report,
		report.Canceled,
		report.Summary{Canceled: len(testConfig.Tests)},
	)
	checkError(t, test.report, "Test run canceled")
	if len(test.report.Steps) != 3 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, &report.Step{Name: ValidateStep, Status: report.Passed})
	setup := test.report.Steps[1]
	checkStep(t, setup, &report.Step{Name: SetupStep, Status: report.Passed})
	tests := test.report.Steps[2]
	checkStep(t, tests, &report.Step{
		Name:   TestsStep,
		Status: report.Canceled,
		Err:    "Test run canceled",
	})
	for i, tc := range testConfig.Tests {
		result := tests.Items[i]
		checkTest(t, result, tc, report.Canceled, "deploy", "protect", "failover")
	}
}

func TestCleanPassed(t *testing.T) {
	test := testCommand(t, testClean, &helpers.TestingMock{})

	if err := test.Clean(); err != nil {
		t.Fatal(err)
	}

	checkReport(t, test.report, report.Passed, report.Summary{Passed: 14})
	checkError(t, test.report, "")
	if len(test.report.Steps) != 3 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, &report.Step{Name: ValidateStep, Status: report.Passed})
	tests := test.report.Steps[1]
	checkStep(t, tests, &report.Step{Name: TestsStep, Status: report.Passed})
	for i, tc := range testConfig.Tests {
		result := tests.Items[i]
		checkTest(t, result, tc, report.Passed, "purge")
	}
	cleanup := test.report.Steps[2]
	checkStep(t, cleanup, &report.Step{Name: CleanupStep, Status: report.Passed})
}

func TestCleanValidateFailed(t *testing.T) {
	test := testCommand(t, testClean, validateFailed)

	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Failed, report.Summary{})
	checkError(t, test.report, "Failed to validate")
	if len(test.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, &report.Step{
		Name:   ValidateStep,
		Status: report.Failed,
		Err:    "Failed to validate",
	})
}

func TestCleanValidateCanceled(t *testing.T) {
	test := testCommand(t, testClean, validateCanceled)

	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(t, test.report, report.Canceled, report.Summary{})
	checkError(t, test.report, "Canceled validate")
	if len(test.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, &report.Step{
		Name:   ValidateStep,
		Status: report.Canceled,
		Err:    "Canceled validate",
	})
}

func TestCleanPurgeFailed(t *testing.T) {
	test := testCommand(t, testClean, purgeFailed)
	helpers.AddGatheredData(t, test.dataDir(), applicationTestdata, "validate-application")
	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(
		t,
		test.report,
		report.Failed,
		report.Summary{Failed: len(testConfig.Tests)},
	)
	checkError(t, test.report,
		"Test clean failed (0 passed, 14 failed, 0 skipped, 0 canceled)")
	if len(test.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, &report.Step{Name: ValidateStep, Status: report.Passed})
	tests := test.report.Steps[1]
	checkStep(t, tests, &report.Step{
		Name:   TestsStep,
		Status: report.Failed,
		Err:    "Test clean failed (0 passed, 14 failed, 0 skipped, 0 canceled)",
	})
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

	checkReport(
		t,
		test.report,
		report.Canceled,
		report.Summary{Canceled: len(testConfig.Tests)},
	)
	checkError(t, test.report, "Test clean canceled")
	if len(test.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, &report.Step{Name: ValidateStep, Status: report.Passed})
	tests := test.report.Steps[1]
	checkStep(t, tests, &report.Step{
		Name:   TestsStep,
		Status: report.Canceled,
		Err:    "Test clean canceled",
	})
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

	checkReport(
		t,
		test.report,
		report.Failed,
		report.Summary{Passed: len(testConfig.Tests)},
	)
	checkError(t, test.report, "Failed to cleanup")
	if len(test.report.Steps) != 3 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, &report.Step{Name: ValidateStep, Status: report.Passed})
	tests := test.report.Steps[1]
	checkStep(t, tests, &report.Step{Name: TestsStep, Status: report.Passed})
	for i, tc := range testConfig.Tests {
		result := tests.Items[i]
		checkTest(t, result, tc, report.Passed, "purge")
	}
	cleanup := test.report.Steps[2]
	checkStep(t, cleanup, &report.Step{
		Name:   CleanupStep,
		Status: report.Failed,
		Err:    "Failed to cleanup",
	})
}

func TestCleanCleanupCanceled(t *testing.T) {
	test := testCommand(t, testClean, cleanupCanceled)

	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}

	checkReport(
		t,
		test.report,
		report.Canceled,
		report.Summary{Passed: len(testConfig.Tests)},
	)
	checkError(t, test.report, "Canceled cleanup")
	if len(test.report.Steps) != 3 {
		t.Fatalf("unexpected steps %+v", test.report.Steps)
	}
	validate := test.report.Steps[0]
	checkStep(t, validate, &report.Step{Name: ValidateStep, Status: report.Passed})
	tests := test.report.Steps[1]
	checkStep(t, tests, &report.Step{Name: TestsStep, Status: report.Passed})
	for i, tc := range testConfig.Tests {
		result := tests.Items[i]
		checkTest(t, result, tc, report.Passed, "purge")
	}
	cleanup := test.report.Steps[2]
	checkStep(t, cleanup, &report.Step{
		Name:   CleanupStep,
		Status: report.Canceled,
		Err:    "Canceled cleanup",
	})
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

func checkReport(t *testing.T, r *Report, status report.Status, summary report.Summary) {
	if r.Status != status {
		t.Errorf("expected status %q, got %q", status, r.Status)
	}
	if !r.Summary.Equal(&summary) {
		t.Errorf("expected summary %+v, got %+v", summary, *r.Summary)
	}
	duration := totalDuration(r.Steps)
	if r.Duration != duration {
		t.Fatalf("expected duration %v, got %v", duration, r.Duration)
	}
}

// We cannot check duration since it may be zero on windows.
func checkStep(t *testing.T, got *report.Step, expected *report.Step) {
	if got.Name != expected.Name {
		t.Fatalf("expected step %q, got %q", expected.Name, got.Name)
	}
	if got.Status != expected.Status {
		t.Fatalf("expected step %q status %q, got %q", expected.Name, expected.Status, got.Status)
	}
	if got.Err != expected.Err {
		t.Fatalf("expected step %q error %q, got %q", expected.Name, expected.Err, got.Err)
	}
}

func checkError(t *testing.T, r *Report, expected string) {
	if got := r.Error(); got != expected {
		t.Fatalf("expected error %q, got %q", expected, got)
	}
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
		t.Fatalf("expected step %q status %q, got %q", name, status, test.Status)
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
		checkStep(t, test.Items[i], &report.Step{Name: name, Status: report.Passed})
	}
	lastStep := &report.Step{Name: flow[last], Status: test.Status}
	switch test.Status {
	case report.Failed:
		lastStep.Err = fmt.Sprintf("Failed to %s application %q", flow[last], name)
	case report.Canceled:
		lastStep.Err = fmt.Sprintf("Canceled %s application %q", flow[last], name)
	}
	checkStep(t, test.Items[last], lastStep)
}

func totalDuration(steps []*report.Step) float64 {
	var total float64
	for _, step := range steps {
		total += step.Duration
	}
	return total
}
