// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/time"
)

var config = &types.Config{
	Distro:     "k8s",
	Repo:       types.RepoConfig{URL: "https://github.com/org/repo", Branch: "main"},
	DRPolicy:   "dr-policy",
	ClusterSet: "clusterset",
	Clusters: map[string]types.ClusterConfig{
		"hub": {Kubeconfig: "hub-kubeconfig"},
		"c1":  {Kubeconfig: "c1-kubeconfig"},
		"c2":  {Kubeconfig: "c2-kubeconfig"},
	},
	PVCSpecs: []types.PVCSpecConfig{
		{Name: "rbd", StorageClassName: "rook-ceph-block", AccessModes: "ReadWriteOnce"},
		{Name: "cephfs", StorageClassName: "rook-cephfs-fs", AccessModes: "ReadWriteMany"},
	},
	Tests: []types.TestConfig{
		{Workload: "appset", Deployer: "deploy", PVCSpec: "rbd"},
		{Workload: "subscr", Deployer: "deploy", PVCSpec: "rbd"},
		{Workload: "disapp", Deployer: "deploy", PVCSpec: "cephfs"},
	},
	Channel: types.ChannelConfig{
		Name:      "my-channel",
		Namespace: "test-gitops",
	},
	Namespaces: types.NamespacesConfig{
		RamenHubNamespace:       "ramen-system",
		RamenDRClusterNamespace: "ramen-system",
		RamenOpsNamespace:       "ramen-ops",
		ArgocdNamespace:         "argocd",
	},
}

func TestReportEmpty(t *testing.T) {
	fakeTime(t)
	r := newReport("test-run", config)

	// Host and ramenctl info is ready.
	expectedReport := report.New()
	if !r.Report.Equal(expectedReport) {
		t.Errorf("expected report %+v, got %+v", expectedReport, r.Report)
	}

	// Otherwise nothing was added so the status and steps should be empty, and summary should be
	// all zero.
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
	r := newReport("test-run", config)
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
	r := newReport("test-run", config)

	step := &Step{Name: SetupStep, Status: Passed, Duration: 1.23}
	r.AddStep(step)
	if r.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, r.Status)
	}

	// Setup succeeded, we should see the successful step.
	if len(r.Steps) != 1 {
		t.Errorf("unexpected steps %+v", r.Steps)
	}
	passedSetup := &Step{Name: SetupStep, Status: Passed, Duration: 1.23}
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
	r := newReport("test-run", config)

	step1 := &Step{Name: SetupStep, Status: Passed}
	r.AddStep(step1)
	if r.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, r.Status)
	}

	step2 := &Step{Name: TestsStep}

	failedTest := &Test{
		TestContext: &Context{name: "appset-deploy-rbd"},
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

	// Adding a failed test mark the step as failed.
	step2.AddTest(failedTest)
	if step2.Status != Failed {
		t.Errorf("expected status %q, got %q", Failed, step2.Status)
	}

	passedTest := &Test{
		TestContext: &Context{name: "appset-deploy-cephfs"},
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

	// Adding a passed test does not change the step status if already set.
	step2.AddTest(passedTest)
	if step2.Status != Failed {
		t.Errorf("expected status %q, got %q", Failed, step2.Status)
	}

	// Adding a failed step to the report mark the report as failed.
	r.AddStep(step2)
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
	r := newReport("test-run", config)

	step1 := &Step{Name: SetupStep, Status: Passed}
	r.AddStep(step1)
	if r.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, r.Status)
	}

	step2 := &Step{Name: TestsStep}

	rbdTest := &Test{
		TestContext: &Context{name: "appset-deploy-rbd"},
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

	// Adding pass test set the step status to passed since the status is unset.
	step2.AddTest(rbdTest)
	if step2.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, step2.Status)
	}

	cephfsTest := &Test{
		TestContext: &Context{name: "appset-deploy-cephfs"},
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

	// Adding another passed tests does not change the status.
	step2.AddTest(cephfsTest)
	if step2.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, step2.Status)
	}

	// Adding passed step keep report status as is.
	r.AddStep(step2)
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
	r := newReport("test-clean", config)

	step1 := &Step{Name: TestsStep}

	rbdTest := &Test{
		TestContext: &Context{name: "appset-deploy-rbd"},
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

	// Adding passed tests set the step status.
	step1.AddTest(rbdTest)
	if step1.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, step1.Status)
	}

	cephfsTest := &Test{
		TestContext: &Context{name: "appset-deploy-cephfs"},
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

	// Adding failed test mark the step as failed.
	step1.AddTest(cephfsTest)
	if step1.Status != Failed {
		t.Errorf("expected status %q, got %q", Failed, step1.Status)
	}

	// Adding failed step mark the report as failed.
	r.AddStep(step1)
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
	r := newReport("test-clean", config)

	step1 := &Step{Name: TestsStep}

	rbdTest := &Test{
		TestContext: &Context{name: "appset-deploy-rbd"},
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

	// Adding a passed test mark the step as passed.
	step1.AddTest(rbdTest)
	if step1.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, step1.Status)
	}

	// Adding a passed step marks the report as passed.
	r.AddStep(step1)
	if r.Status != Passed {
		t.Errorf("expected status %q, got %q", Failed, r.Status)
	}

	step2 := &Step{Name: CleanupStep, Status: Failed}

	// Adding a failed step marks the report as failed.
	r.AddStep(step2)
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
	r := newReport("test-clean", config)

	step1 := &Step{Name: TestsStep}

	rbdTest := &Test{
		TestContext: &Context{name: "appset-deploy-rbd"},
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

	// Adding a passed test marks the step as passed.
	step1.AddTest(rbdTest)
	if step1.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, step1.Status)
	}

	cephfsTest := &Test{
		TestContext: &Context{name: "appset-deploy-cephfs"},
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

	// Adding a passed test does not change step status.
	step1.AddTest(cephfsTest)
	if step1.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, step1.Status)
	}

	// Adding a passed step keeps the report passed.
	r.AddStep(step1)
	if r.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, r.Status)
	}

	step2 := &Step{Name: CleanupStep, Status: Passed}

	// Adding a passed step keeps the report passed.
	r.AddStep(step2)
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

func TestSummaryString(t *testing.T) {
	summary := Summary{Passed: 5, Failed: 2, Skipped: 3, Canceled: 1}

	expectedString := "5 passed, 2 failed, 3 skipped, 1 canceled"
	if summary.String() != expectedString {
		t.Errorf("expected summary string %s, got %s", expectedString, summary.String())
	}
}

func TestSummaryMarshal(t *testing.T) {
	summary := Summary{Passed: 4, Failed: 3, Skipped: 2, Canceled: 1}

	bytes, err := yaml.Marshal(summary)
	if err != nil {
		t.Fatalf("failed to marshal summary: %v", err)
	}

	var unmarshaledSummary Summary
	err = yaml.Unmarshal(bytes, &unmarshaledSummary)
	if err != nil {
		t.Fatalf("failed to unmarshal summary: %v", err)
	}
	if unmarshaledSummary != summary {
		t.Errorf("unmarshaled summary %+v does not match original summary %+v",
			unmarshaledSummary, summary)
	}
}

func TestSummaryCount(t *testing.T) {
	summary := Summary{}

	// Add multiple tests of different status
	summary.AddTest(&Step{Status: Passed})
	summary.AddTest(&Step{Status: Passed})
	summary.AddTest(&Step{Status: Failed})
	summary.AddTest(&Step{Status: Skipped})
	summary.AddTest(&Step{Status: Canceled})
	summary.AddTest(&Step{Status: Passed})

	expectedSummary := Summary{Passed: 3, Failed: 1, Skipped: 1, Canceled: 1}
	if summary != expectedSummary {
		t.Errorf("expected summary %+v, got %+v", expectedSummary, summary)
	}
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

var fakeNow = time.Now()

func fakeTime(t *testing.T) {
	savedNow := time.Now
	time.Now = func() time.Time {
		return fakeNow
	}
	t.Cleanup(func() {
		time.Now = savedNow
	})
}
