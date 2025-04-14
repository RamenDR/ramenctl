// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/report"
)

func TestReportEmpty(t *testing.T) {
	fakeTime(t)
	r := newReport("test-run")

	// Host and ramenctl info is ready.
	expectedReport := report.New()
	if !r.Report.Equal(expectedReport) {
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
	fakeTime(t)
	r := newReport("test-run")
	step := &Step{Name: SetupStep, Status: Failed}
	r.AddStep(step)

	// Setup failed, so entire report should be failed.
	if r.Status != Failed {
		t.Errorf("expected status %q, got %q", Failed, r.Status)
	}

	// Setup failed, we should see the failed step.
	if len(r.Steps) != 1 {
		t.Errorf("unexpected steps %+v", r.Steps)
	}
	failedSetup := &Step{Name: SetupStep, Status: Failed}
	if !r.Steps[0].Equal(failedSetup) {
		t.Fatalf("expected setup %+v, got %+v", r.Steps[0], failedSetup)
	}

	// No test run son counts should be zero.
	expectedSummary := Summary{}
	if r.Summary != expectedSummary {
		t.Errorf("expected summary %+v,  %+v", expectedSummary, r.Summary)
	}

	// We can marshal and unmarshal the report
	checkRoundtrip(t, r)
}

func TestReportRunSetupPassed(t *testing.T) {
	fakeTime(t)
	r := newReport("test-run")

	step := &Step{Name: SetupStep, Status: Passed}
	r.AddStep(step)
	if r.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, r.Status)
	}

	// Setup succeeded, we should see the successful step.
	if len(r.Steps) != 1 {
		t.Errorf("unexpected steps %+v", r.Steps)
	}
	passedSetup := &Step{Name: SetupStep, Status: Passed}
	if !r.Steps[0].Equal(passedSetup) {
		t.Fatalf("expected setup %+v, got %+v", passedSetup, r.Steps[0])
	}

	// No test run son counts should be zero.
	expectedSummary := Summary{}
	if r.Summary != expectedSummary {
		t.Errorf("expected summary %+v, got %+v", expectedSummary, r.Summary)
	}

	// We can marshal and unmarshal the report
	checkRoundtrip(t, r)
}

func TestReportRunTestFailed(t *testing.T) {
	fakeTime(t)
	r := newReport("test-run")

	step := &Step{Name: SetupStep, Status: Passed}
	r.AddStep(step)
	if r.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, r.Status)
	}

	failedTest := &Test{
		Context: &Context{name: "appset-deploy-rbd"},
		Config: &types.TestConfig{
			Workload: "deploy",
			Deployer: "appset",
			PVCSpec:  "rbd",
		},
		Status: Failed,
		Steps: []*Step{
			{Name: "deploy", Status: Passed},
			{Name: "protect", Status: Passed},
			{Name: "failover", Status: Failed},
		},
	}
	r.AddTest(failedTest)
	if r.Status != Failed {
		t.Errorf("expected status %q, got %q", Failed, r.Status)
	}

	passedTest := &Test{
		Context: &Context{name: "appset-deploy-cephfs"},
		Config: &types.TestConfig{
			Workload: "deploy",
			Deployer: "appset",
			PVCSpec:  "cephfs",
		},
		Status: Passed,
		Steps: []*Step{
			{Name: "deploy", Status: Passed},
			{Name: "protect", Status: Passed},
			{Name: "failover", Status: Passed},
			{Name: "relocate", Status: Passed},
			{Name: "unprotect", Status: Passed},
			{Name: "undeploy", Status: Passed},
		},
	}
	r.AddTest(passedTest)
	if r.Status != Failed {
		t.Errorf("expected status %q, got %q", Failed, r.Status)
	}

	// We should have a passed setup step, and failed tests step.
	if len(r.Steps) != 2 {
		t.Errorf("unexpected steps %+v", r.Steps)
	}
	setup := r.Steps[0]
	passedSetup := &Step{Name: SetupStep, Status: Passed}
	if !setup.Equal(passedSetup) {
		t.Fatalf("expected setup %+v, got %+v", passedSetup, setup)
	}
	// One test failed, so the tests step must be failed.
	tests := r.Steps[1]
	if tests.Name != TestsStep {
		t.Errorf("expected step name %q, got %q", TestsStep, tests.Name)
	}
	if tests.Status != Failed {
		t.Errorf("expected step status %q, got %q", Failed, tests.Status)
	}

	// The tests setup must have 2 results.
	if len(tests.Items) != 2 {
		t.Errorf("unexpected tests %+v", r.Steps[1].Items)
	}

	failedResult := &Step{
		Name:   failedTest.Name(),
		Config: failedTest.Config,
		Status: failedTest.Status,
		Items:  failedTest.Steps,
	}
	if !tests.Items[0].Equal(failedResult) {
		t.Errorf("expected result %+v, got %+v", failedResult, tests.Items[0])
	}

	passedResult := &Step{
		Name:   passedTest.Name(),
		Config: passedTest.Config,
		Status: passedTest.Status,
		Items:  passedTest.Steps,
	}
	if !tests.Items[1].Equal(passedResult) {
		t.Errorf("expected result %+v, got %+v", passedResult, tests.Items[1])
	}

	// Counts updated.
	expectedSummary := Summary{Passed: 1, Failed: 1}
	if r.Summary != expectedSummary {
		t.Errorf("expected summary %+v, got %+v", expectedSummary, r.Summary)
	}

	// We can marshal and unmarshal the report
	checkRoundtrip(t, r)
}

func TestReportRunAllPassed(t *testing.T) {
	fakeTime(t)
	r := newReport("test-run")

	step := &Step{Name: SetupStep, Status: Passed}
	r.AddStep(step)
	if r.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, r.Status)
	}

	rbdTest := &Test{
		Context: &Context{name: "appset-deploy-rbd"},
		Config: &types.TestConfig{
			Workload: "deploy",
			Deployer: "appset",
			PVCSpec:  "rbd",
		},
		Status: Passed,
		Steps: []*Step{
			{Name: "deploy", Status: Passed},
			{Name: "protect", Status: Passed},
			{Name: "failover", Status: Passed},
			{Name: "relocate", Status: Passed},
			{Name: "unprotect", Status: Passed},
			{Name: "undeploy", Status: Passed},
		},
	}
	r.AddTest(rbdTest)
	if r.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, r.Status)
	}

	cephfsTest := &Test{
		Context: &Context{name: "appset-deploy-cephfs"},
		Config: &types.TestConfig{
			Workload: "deploy",
			Deployer: "appset",
			PVCSpec:  "cephfs",
		},
		Status: Passed,
		Steps: []*Step{
			{Name: "deploy", Status: Passed},
			{Name: "protect", Status: Passed},
			{Name: "failover", Status: Passed},
			{Name: "relocate", Status: Passed},
			{Name: "unprotect", Status: Passed},
			{Name: "undeploy", Status: Passed},
		},
	}
	r.AddTest(cephfsTest)
	if r.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, r.Status)
	}

	// We should have a passed setup and tests steps.
	if len(r.Steps) != 2 {
		t.Errorf("unexpected steps %+v", r.Steps)
	}
	setup := r.Steps[0]
	passedSetup := &Step{Name: SetupStep, Status: Passed}
	if !setup.Equal(passedSetup) {
		t.Fatalf("expected setup %+v, got %+v", passedSetup, setup)
	}

	// All tests passed, so the tests step must be passed.
	tests := r.Steps[1]
	if tests.Name != TestsStep {
		t.Errorf("expected step name %q, got %q", TestsStep, tests.Name)
	}
	if tests.Status != Passed {
		t.Errorf("expected step status %q, got %q", Passed, tests.Status)
	}

	// The tests setup must have 2 passed results.
	if len(tests.Items) != 2 {
		t.Errorf("unexpected tests %+v", tests.Items)
	}

	rbdResult := &Step{
		Name:   rbdTest.Name(),
		Config: rbdTest.Config,
		Status: rbdTest.Status,
		Items:  rbdTest.Steps,
	}
	if !tests.Items[0].Equal(rbdResult) {
		t.Errorf("expected result %+v, got %+v", rbdResult, tests.Items[0])
	}

	cephfsResult := &Step{
		Name:   cephfsTest.Name(),
		Config: cephfsTest.Config,
		Status: cephfsTest.Status,
		Items:  cephfsTest.Steps,
	}
	if !tests.Items[1].Equal(cephfsResult) {
		t.Errorf("expected result %+v, got %+v", rbdResult, tests.Items[1])
	}

	// Counts updated.
	if r.Summary.Passed != 2 || r.Summary.Failed != 0 || r.Summary.Skipped != 0 {
		t.Errorf("unexpected summary: %+v", r.Summary)
	}

	// We can marshal and unmarshal the report
	checkRoundtrip(t, r)
}

func TestReportCleanTestFailed(t *testing.T) {
	fakeTime(t)
	r := newReport("test-clean")

	rbdTest := &Test{
		Context: &Context{name: "appset-deploy-rbd"},
		Config: &types.TestConfig{
			Workload: "deploy",
			Deployer: "appset",
			PVCSpec:  "rbd",
		},
		Status: Passed,
		Steps: []*Step{
			{Name: "unprotect", Status: Passed},
			{Name: "undeploy", Status: Passed},
		},
	}
	r.AddTest(rbdTest)
	if r.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, r.Status)
	}

	cephfsTest := &Test{
		Context: &Context{name: "appset-deploy-cephfs"},
		Config: &types.TestConfig{
			Workload: "deploy",
			Deployer: "appset",
			PVCSpec:  "cephfs",
		},
		Status: Failed,
		Steps: []*Step{
			{Name: "unprotect", Status: Failed},
		},
	}
	r.AddTest(cephfsTest)
	if r.Status != Failed {
		t.Errorf("expected status %q, got %q", Failed, r.Status)
	}

	// We should have a failed tests step.
	if len(r.Steps) != 1 {
		t.Errorf("unexpected steps %+v", r.Steps)
	}

	// One test failed, so the tests step must be failed.
	tests := r.Steps[0]
	if tests.Name != TestsStep {
		t.Errorf("expected step name %q, got %q", TestsStep, tests.Name)
	}
	if tests.Status != Failed {
		t.Errorf("expected step status %q, got %q", Failed, tests.Status)
	}

	// The tests setup must have 2 results.
	if len(tests.Items) != 2 {
		t.Errorf("unexpected tests %+v", tests.Items)
	}

	rbdResult := &Step{
		Name:   rbdTest.Name(),
		Config: rbdTest.Config,
		Status: rbdTest.Status,
		Items:  rbdTest.Steps,
	}
	if !tests.Items[0].Equal(rbdResult) {
		t.Errorf("expected result %+v, got %+v", rbdResult, tests.Items[0])
	}

	cephfsResult := &Step{
		Name:   cephfsTest.Name(),
		Config: cephfsTest.Config,
		Status: cephfsTest.Status,
		Items:  cephfsTest.Steps,
	}
	if !tests.Items[1].Equal(cephfsResult) {
		t.Errorf("expected result %+v, got %+v", cephfsResult, tests.Items[1])
	}

	// Counts updated.
	if r.Summary.Passed != 1 || r.Summary.Failed != 1 || r.Summary.Skipped != 0 {
		t.Errorf("unexpected summary: %+v", r.Summary)
	}

	// We can marshal and unmarshal the report
	checkRoundtrip(t, r)
}

func TestReportCleanFailed(t *testing.T) {
	fakeTime(t)
	r := newReport("test-clean")

	rbdTest := &Test{
		Context: &Context{name: "appset-deploy-rbd"},
		Config: &types.TestConfig{
			Workload: "deploy",
			Deployer: "appset",
			PVCSpec:  "rbd",
		},
		Status: Passed,
		Steps: []*Step{
			{Name: "unprotect", Status: Passed},
			{Name: "undeploy", Status: Passed},
		},
	}
	r.AddTest(rbdTest)
	if r.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, r.Status)
	}

	step := &Step{Name: CleanupStep, Status: Failed}
	r.AddStep(step)
	if r.Status != Failed {
		t.Errorf("expected status %q, got %q", Failed, r.Status)
	}

	// We should have a passed tests and failed cleanup steps.
	if len(r.Steps) != 2 {
		t.Errorf("unexpected steps %+v", r.Steps)
	}

	// All tests passed, so the tests step must be passed.
	tests := r.Steps[0]
	if tests.Name != TestsStep {
		t.Errorf("expected step name %q, got %q", TestsStep, tests.Name)
	}
	if tests.Status != Passed {
		t.Errorf("expected step status %q, got %q", Passed, tests.Status)
	}

	// The cleanup step failed, the step must be passed.
	cleanup := r.Steps[1]
	failedCleanup := &Step{Name: CleanupStep, Status: Failed}
	if !cleanup.Equal(failedCleanup) {
		t.Fatalf("expected setup %+v, got %+v", cleanup, failedCleanup)
	}

	// Counts updated.
	expectedSummary := Summary{Passed: 1}
	if r.Summary != expectedSummary {
		t.Errorf("expected summary %+v, got %+v", expectedSummary, r.Summary)
	}

	// We can marshal and unmarshal the report
	checkRoundtrip(t, r)
}

func TestReportCleanAllPassed(t *testing.T) {
	fakeTime(t)
	r := newReport("test-clean")

	rbdTest := &Test{
		Context: &Context{name: "appset-deploy-rbd"},
		Config: &types.TestConfig{
			Workload: "deploy",
			Deployer: "appset",
			PVCSpec:  "rbd",
		},
		Status: Passed,
		Steps: []*Step{
			{Name: "unprotect", Status: Passed},
			{Name: "undeploy", Status: Passed},
		},
	}
	r.AddTest(rbdTest)
	if r.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, r.Status)
	}

	cephfsTest := &Test{
		Context: &Context{name: "appset-deploy-cephfs"},
		Config: &types.TestConfig{
			Workload: "deploy",
			Deployer: "appset",
			PVCSpec:  "cephfs",
		},
		Status: Passed,
		Steps: []*Step{
			{Name: "unprotect", Status: Passed},
			{Name: "undeploy", Status: Passed},
		},
	}
	r.AddTest(cephfsTest)
	if r.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, r.Status)
	}

	step := &Step{Name: CleanupStep, Status: Passed}
	r.AddStep(step)
	if r.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, r.Status)
	}

	// We should have a passed tests and cleanup steps.
	if len(r.Steps) != 2 {
		t.Errorf("unexpected steps %+v", r.Steps)
	}

	// All tests passed, so the tests step must be passed.
	tests := r.Steps[0]
	if tests.Name != TestsStep {
		t.Errorf("expected step name %q, got %q", TestsStep, tests.Name)
	}
	if tests.Status != Passed {
		t.Errorf("expected step status %q, got %q", Passed, tests.Status)
	}

	// The tests setup must have 2 passed results.
	if len(tests.Items) != 2 {
		t.Errorf("unexpected tests %+v", tests.Items)
	}

	rbdResult := &Step{
		Name:   rbdTest.Name(),
		Config: rbdTest.Config,
		Status: rbdTest.Status,
		Items:  rbdTest.Steps,
	}
	if !tests.Items[0].Equal(rbdResult) {
		t.Errorf("expected result %+v, got %+v", rbdResult, tests.Items[0])
	}

	cephfsResult := &Step{
		Name:   cephfsTest.Name(),
		Config: cephfsTest.Config,
		Status: cephfsTest.Status,
		Items:  cephfsTest.Steps,
	}
	if !tests.Items[1].Equal(cephfsResult) {
		t.Errorf("expected result %+v, got %+v", cephfsResult, tests.Items[1])
	}

	// The cleanup step passed, the step must be passed.
	cleanup := r.Steps[1]
	passedCleanup := &Step{Name: CleanupStep, Status: Passed}
	if !cleanup.Equal(passedCleanup) {
		t.Fatalf("expected setup %+v, got %+v", cleanup, passedCleanup)
	}

	// Counts updated.
	expectedSummary := Summary{Passed: 2}
	if r.Summary != expectedSummary {
		t.Errorf("expected %+v, got %+v", expectedSummary, r.Summary)
	}

	// We can marshal and unmarshal the report
	checkRoundtrip(t, r)
}

func checkRoundtrip(t *testing.T, r1 *Report) {
	// We must be able to marshal and unmarshal the report
	b, err := yaml.Marshal(r1)
	if err != nil {
		t.Fatalf("failed to marshal report: %s", err)
	}
	r2 := &Report{}
	if err := yaml.Unmarshal(b, r2); err != nil {
		t.Fatalf("failed to unmarshal report: %s", err)
	}
	if !r1.Equal(r2) {
		t.Fatalf("expected report %+v, got %+v", r1, r2)
	}
}

var fakeNow = report.Now()

func fakeTime(t *testing.T) {
	savedNow := report.Now
	report.Now = func() time.Time {
		return fakeNow
	}
	t.Cleanup(func() {
		report.Now = savedNow
	})
}
