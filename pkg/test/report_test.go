// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test_test

import (
	"reflect"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramen/e2e/types"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/test"
)

func TestReportEmpty(t *testing.T) {
	r := test.NewReport("test-run")

	// Host and ramenctl info is ready.
	expectedReport := report.New()
	if !reflect.DeepEqual(r.Report, expectedReport) {
		t.Errorf("expected report %+v, got %+v", expectedReport, r.Report)
	}

	// Otherwise nothing was added so the status and steps should be empty, and summary should be all zero.
	if r.Status != "" {
		t.Errorf("non-empty status: %q", r.Status)
	}
	if len(r.Steps) != 0 {
		t.Errorf("unexpected steps %+v", r.Steps)
	}
	if r.Summary.Passed != 0 || r.Summary.Failed != 0 || r.Summary.Skipped != 0 {
		t.Errorf("unexpected summary: %+v", r.Summary)
	}

	// We can marshal and unmarshal the report
	checkRoundtrip(t, r)
}

func TestReportRunSetupFailed(t *testing.T) {
	r := test.NewReport("test-run")
	r.AddSetup(false)

	// Setup failed, so entire report should be failed.
	if r.Status != test.Failed {
		t.Errorf("expected status %q, got %q", test.Failed, r.Status)
	}

	// Setup failed, we should see the failed step.
	if len(r.Steps) != 1 {
		t.Errorf("unexpected steps %+v", r.Steps)
	}
	failedSetup := &test.Step{Name: test.SetupStep, Status: test.Failed}
	if !reflect.DeepEqual(r.Steps[0], failedSetup) {
		t.Fatalf("expected setup %+v, got %+v", r.Steps[0], failedSetup)
	}

	// No test run son counts should be zero.
	if r.Summary.Passed != 0 || r.Summary.Failed != 0 || r.Summary.Skipped != 0 {
		t.Errorf("unexpected summary: %+v", r.Summary)
	}

	// We can marshal and unmarshal the report
	checkRoundtrip(t, r)
}

func TestReportRunSetupPassed(t *testing.T) {
	r := test.NewReport("test-run")
	r.AddSetup(true)

	// Setup succeeded, so entire report should be passed.
	if r.Status != test.Passed {
		t.Errorf("expected status %q, got %q", test.Passed, r.Status)
	}

	// Setup succeeded, we should see the successful step.
	if len(r.Steps) != 1 {
		t.Errorf("unexpected steps %+v", r.Steps)
	}
	passedSetup := &test.Step{Name: test.SetupStep, Status: test.Passed}
	if !reflect.DeepEqual(r.Steps[0], passedSetup) {
		t.Fatalf("expected setup %+v, got %+v", r.Steps[0], passedSetup)
	}

	// No test run son counts should be zero.
	if r.Summary.Passed != 0 || r.Summary.Failed != 0 || r.Summary.Skipped != 0 {
		t.Errorf("unexpected summary: %+v", r.Summary)
	}

	// We can marshal and unmarshal the report
	checkRoundtrip(t, r)
}

func TestReportRunTestFailed(t *testing.T) {
	r := test.NewReport("test-run")
	r.AddSetup(true)
	if r.Status != test.Passed {
		t.Errorf("expected status %q, got %q", test.Passed, r.Status)
	}

	failedTest := &test.Test{
		Config: types.TestConfig{
			Workload: "deploy",
			Deployer: "appset",
			PVCSpec:  "rbd",
		},
		Status: test.Failed,
	}
	r.AddTest(failedTest)
	if r.Status != test.Failed {
		t.Errorf("expected status %q, got %q", test.Failed, r.Status)
	}

	passedTest := &test.Test{
		Config: types.TestConfig{
			Workload: "deploy",
			Deployer: "appset",
			PVCSpec:  "cephfs",
		},
		Status: test.Passed,
	}
	r.AddTest(passedTest)
	if r.Status != test.Failed {
		t.Errorf("expected status %q, got %q", test.Failed, r.Status)
	}

	// We should have a passed setup step, and failed tests step.
	if len(r.Steps) != 2 {
		t.Errorf("unexpected steps %+v", r.Steps)
	}
	setup := r.Steps[0]
	passedSetup := &test.Step{Name: test.SetupStep, Status: test.Passed}
	if !reflect.DeepEqual(setup, passedSetup) {
		t.Fatalf("expected setup %+v, got %+v", setup, passedSetup)
	}
	// One test failed, so the tests step must be failed.
	tests := r.Steps[1]
	if tests.Name != test.TestsStep {
		t.Errorf("expected step name %q, got %q", test.TestsStep, tests.Name)
	}
	if tests.Status != test.Failed {
		t.Errorf("expected step status %q, got %q", test.Failed, tests.Status)
	}

	// The tests setup must have 2 results.
	if len(tests.Tests) != 2 {
		t.Errorf("unexpected tests %+v", r.Steps[1].Tests)
	}

	failedResult := test.Result{
		Workload: failedTest.Config.Workload,
		Deployer: failedTest.Config.Deployer,
		PVCSpec:  failedTest.Config.PVCSpec,
		Status:   failedTest.Status,
	}
	if tests.Tests[0] != failedResult {
		t.Errorf("expected result %+v, got %+v", failedResult, tests.Tests[0])
	}

	passedResult := test.Result{
		Workload: passedTest.Config.Workload,
		Deployer: passedTest.Config.Deployer,
		PVCSpec:  passedTest.Config.PVCSpec,
		Status:   passedTest.Status,
	}
	if tests.Tests[1] != passedResult {
		t.Errorf("expected result %+v, got %+v", failedResult, tests.Tests[1])
	}

	// Counts updated.
	if r.Summary.Passed != 1 || r.Summary.Failed != 1 || r.Summary.Skipped != 0 {
		t.Errorf("unexpected summary: %+v", r.Summary)
	}

	// We can marshal and unmarshal the report
	checkRoundtrip(t, r)
}

func TestReportRunAllPassed(t *testing.T) {
	r := test.NewReport("test-run")
	r.AddSetup(true)
	if r.Status != test.Passed {
		t.Errorf("expected status %q, got %q", test.Passed, r.Status)
	}

	rbdTest := &test.Test{
		Config: types.TestConfig{
			Workload: "deploy",
			Deployer: "appset",
			PVCSpec:  "rbd",
		},
		Status: test.Passed,
	}
	r.AddTest(rbdTest)
	if r.Status != test.Passed {
		t.Errorf("expected status %q, got %q", test.Passed, r.Status)
	}

	cephfsTest := &test.Test{
		Config: types.TestConfig{
			Workload: "deploy",
			Deployer: "appset",
			PVCSpec:  "cephfs",
		},
		Status: test.Passed,
	}
	r.AddTest(cephfsTest)
	if r.Status != test.Passed {
		t.Errorf("expected status %q, got %q", test.Passed, r.Status)
	}

	// We should have a passed setup and tests steps.
	if len(r.Steps) != 2 {
		t.Errorf("unexpected steps %+v", r.Steps)
	}
	setup := r.Steps[0]
	passedSetup := &test.Step{Name: test.SetupStep, Status: test.Passed}
	if !reflect.DeepEqual(setup, passedSetup) {
		t.Fatalf("expected setup %+v, got %+v", setup, passedSetup)
	}

	// All tests passed, so the tests step must be passed.
	tests := r.Steps[1]
	if tests.Name != test.TestsStep {
		t.Errorf("expected step name %q, got %q", test.TestsStep, tests.Name)
	}
	if tests.Status != test.Passed {
		t.Errorf("expected step status %q, got %q", test.Passed, tests.Status)
	}

	// The tests setup must have 2 passed results.
	if len(tests.Tests) != 2 {
		t.Errorf("unexpected tests %+v", tests.Tests)
	}

	rbdResult := test.Result{
		Workload: rbdTest.Config.Workload,
		Deployer: rbdTest.Config.Deployer,
		PVCSpec:  rbdTest.Config.PVCSpec,
		Status:   rbdTest.Status,
	}
	if tests.Tests[0] != rbdResult {
		t.Errorf("expected result %+v, got %+v", rbdResult, tests.Tests[0])
	}

	cephfsResult := test.Result{
		Workload: cephfsTest.Config.Workload,
		Deployer: cephfsTest.Config.Deployer,
		PVCSpec:  cephfsTest.Config.PVCSpec,
		Status:   cephfsTest.Status,
	}
	if tests.Tests[1] != cephfsResult {
		t.Errorf("expected result %+v, got %+v", rbdResult, tests.Tests[1])
	}

	// Counts updated.
	if r.Summary.Passed != 2 || r.Summary.Failed != 0 || r.Summary.Skipped != 0 {
		t.Errorf("unexpected summary: %+v", r.Summary)
	}

	// We can marshal and unmarshal the report
	checkRoundtrip(t, r)
}

func TestReportCleanTestFailed(t *testing.T) {
	r := test.NewReport("test-clean")

	rbdTest := &test.Test{
		Config: types.TestConfig{
			Workload: "deploy",
			Deployer: "appset",
			PVCSpec:  "rbd",
		},
		Status: test.Passed,
	}
	r.AddTest(rbdTest)
	if r.Status != test.Passed {
		t.Errorf("expected status %q, got %q", test.Passed, r.Status)
	}

	cephfsTest := &test.Test{
		Config: types.TestConfig{
			Workload: "deploy",
			Deployer: "appset",
			PVCSpec:  "cephfs",
		},
		Status: test.Failed,
	}
	r.AddTest(cephfsTest)
	if r.Status != test.Failed {
		t.Errorf("expected status %q, got %q", test.Failed, r.Status)
	}

	// We should have a failed tests step.
	if len(r.Steps) != 1 {
		t.Errorf("unexpected steps %+v", r.Steps)
	}

	// One test failed, so the tests step must be failed.
	tests := r.Steps[0]
	if tests.Name != test.TestsStep {
		t.Errorf("expected step name %q, got %q", test.TestsStep, tests.Name)
	}
	if tests.Status != test.Failed {
		t.Errorf("expected step status %q, got %q", test.Failed, tests.Status)
	}

	// The tests setup must have 2 results.
	if len(tests.Tests) != 2 {
		t.Errorf("unexpected tests %+v", tests.Tests)
	}

	rbdResult := test.Result{
		Workload: rbdTest.Config.Workload,
		Deployer: rbdTest.Config.Deployer,
		PVCSpec:  rbdTest.Config.PVCSpec,
		Status:   rbdTest.Status,
	}
	if tests.Tests[0] != rbdResult {
		t.Errorf("expected result %+v, got %+v", rbdResult, tests.Tests[0])
	}

	cephfsResult := test.Result{
		Workload: cephfsTest.Config.Workload,
		Deployer: cephfsTest.Config.Deployer,
		PVCSpec:  cephfsTest.Config.PVCSpec,
		Status:   cephfsTest.Status,
	}
	if tests.Tests[1] != cephfsResult {
		t.Errorf("expected result %+v, got %+v", rbdResult, tests.Tests[1])
	}

	// Counts updated.
	if r.Summary.Passed != 1 || r.Summary.Failed != 1 || r.Summary.Skipped != 0 {
		t.Errorf("unexpected summary: %+v", r.Summary)
	}

	// We can marshal and unmarshal the report
	checkRoundtrip(t, r)
}

func TestReportCleanAllPassed(t *testing.T) {
	r := test.NewReport("test-clean")

	rbdTest := &test.Test{
		Config: types.TestConfig{
			Workload: "deploy",
			Deployer: "appset",
			PVCSpec:  "rbd",
		},
		Status: test.Passed,
	}
	r.AddTest(rbdTest)
	if r.Status != test.Passed {
		t.Errorf("expected status %q, got %q", test.Passed, r.Status)
	}

	cephfsTest := &test.Test{
		Config: types.TestConfig{
			Workload: "deploy",
			Deployer: "appset",
			PVCSpec:  "cephfs",
		},
		Status: test.Passed,
	}
	r.AddTest(cephfsTest)
	if r.Status != test.Passed {
		t.Errorf("expected status %q, got %q", test.Passed, r.Status)
	}

	r.AddCleanup(true)
	if r.Status != test.Passed {
		t.Errorf("expected status %q, got %q", test.Passed, r.Status)
	}

	// We should have a passed tests and cleanup steps.
	if len(r.Steps) != 2 {
		t.Errorf("unexpected steps %+v", r.Steps)
	}

	// All tests passed, so the tests step must be passed.
	tests := r.Steps[0]
	if tests.Name != test.TestsStep {
		t.Errorf("expected step name %q, got %q", test.TestsStep, tests.Name)
	}
	if tests.Status != test.Passed {
		t.Errorf("expected step status %q, got %q", test.Passed, tests.Status)
	}

	// The tests setup must have 2 passed results.
	if len(tests.Tests) != 2 {
		t.Errorf("unexpected tests %+v", tests.Tests)
	}

	rbdResult := test.Result{
		Workload: rbdTest.Config.Workload,
		Deployer: rbdTest.Config.Deployer,
		PVCSpec:  rbdTest.Config.PVCSpec,
		Status:   rbdTest.Status,
	}
	if tests.Tests[0] != rbdResult {
		t.Errorf("expected result %+v, got %+v", rbdResult, tests.Tests[0])
	}

	cephfsResult := test.Result{
		Workload: cephfsTest.Config.Workload,
		Deployer: cephfsTest.Config.Deployer,
		PVCSpec:  cephfsTest.Config.PVCSpec,
		Status:   cephfsTest.Status,
	}
	if tests.Tests[1] != cephfsResult {
		t.Errorf("expected result %+v, got %+v", rbdResult, tests.Tests[1])
	}

	// The cleanup step passed, the step must be passed.
	cleanup := r.Steps[1]
	passedCleanup := &test.Step{Name: test.CleanupStep, Status: test.Passed}
	if !reflect.DeepEqual(cleanup, passedCleanup) {
		t.Fatalf("expected setup %+v, got %+v", cleanup, passedCleanup)
	}

	// Counts updated.
	if r.Summary.Passed != 2 || r.Summary.Failed != 0 || r.Summary.Skipped != 0 {
		t.Errorf("unexpected summary: %+v", r.Summary)
	}

	// We can marshal and unmarshal the report
	checkRoundtrip(t, r)
}

func checkRoundtrip(t *testing.T, r1 *test.Report) {
	// We must be able to marshal and unmarshal the report
	b, err := yaml.Marshal(r1)
	if err != nil {
		t.Fatalf("failed to marshal report: %s", err)
	}
	r2 := &test.Report{}
	if err := yaml.Unmarshal(b, r2); err != nil {
		t.Fatalf("failed to unmarshal report: %s", err)
	}
	if !reflect.DeepEqual(r1, r2) {
		t.Fatalf("expected report %+v, got %+v", r1, r2)
	}
}
