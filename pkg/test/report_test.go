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

func TestStepAddPassedTest(t *testing.T) {
	passedTest := &Test{
		TestContext: &Context{name: "passing_test"},
		Status:      Passed,
		Config:      &config.Tests[0],
		Duration:    6.0,
		Steps: []*Step{
			{Name: "deploy", Status: Passed, Duration: 1.0},
			{Name: "protect", Status: Passed, Duration: 1.0},
			{Name: "failover", Status: Passed, Duration: 1.0},
			{Name: "relocate", Status: Passed, Duration: 1.0},
			{Name: "unprotect", Status: Passed, Duration: 1.0},
			{Name: "undeploy", Status: Passed, Duration: 1.0},
		},
	}
	t.Run("empty initial status", func(t *testing.T) {
		rootStep := &Step{Name: "root"}
		rootStep.AddTest(passedTest)
		expectedStep := &Step{
			Name:   rootStep.Name,
			Status: passedTest.Status,
			Items: []*Step{
				{
					Name:     passedTest.Name(),
					Status:   passedTest.Status,
					Duration: passedTest.Duration,
					Config:   passedTest.Config,
					Items:    passedTest.Steps,
				},
			},
		}
		if !rootStep.Equal(expectedStep) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", rootStep, expectedStep)
		}
	})

	t.Run("failed initial status", func(t *testing.T) {
		rootStep := &Step{Name: "root", Status: Failed}
		rootStep.AddTest(passedTest)
		expectedStep := &Step{
			Name: rootStep.Name,
			// Failed status should not be changed
			Status: Failed,
			Items: []*Step{
				{
					Name:     passedTest.Name(),
					Status:   passedTest.Status,
					Duration: passedTest.Duration,
					Config:   passedTest.Config,
					Items:    passedTest.Steps,
				},
			},
		}
		if !rootStep.Equal(expectedStep) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", rootStep, expectedStep)
		}
	})

	t.Run("canceled initial status", func(t *testing.T) {
		rootStep := &Step{Name: "root", Status: Canceled}
		rootStep.AddTest(passedTest)
		expectedStep := &Step{
			Name: rootStep.Name,
			// Canceled status should not be changed
			Status: Canceled,
			Items: []*Step{
				{
					Name:     passedTest.Name(),
					Status:   passedTest.Status,
					Duration: passedTest.Duration,
					Config:   passedTest.Config,
					Items:    passedTest.Steps,
				},
			},
		}
		if !rootStep.Equal(expectedStep) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", rootStep, expectedStep)
		}
	})
}

func TestStepAddFailedTest(t *testing.T) {
	failedTest := &Test{
		TestContext: &Context{name: "failing_test"},
		Status:      Failed,
		Config:      &config.Tests[0],
		Duration:    1.0,
		Steps: []*Step{
			{Name: "undeploy", Status: Failed, Duration: 1.0},
		},
	}
	t.Run("empty initial status", func(t *testing.T) {
		rootStep := &Step{Name: "root"}
		rootStep.AddTest(failedTest)
		expectedStep := &Step{
			Name: rootStep.Name,
			// Status should be Failed
			Status: Failed,
			Items: []*Step{
				{
					Name:     failedTest.Name(),
					Status:   failedTest.Status,
					Duration: failedTest.Duration,
					Config:   failedTest.Config,
					Items:    failedTest.Steps,
				},
			},
		}
		if !rootStep.Equal(expectedStep) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", rootStep, expectedStep)
		}
	})

	t.Run("passed initial status", func(t *testing.T) {
		rootStep := &Step{Name: "root", Status: Passed}
		rootStep.AddTest(failedTest)
		expectedStep := &Step{
			Name: rootStep.Name,
			// Passed status should be changed to Failed
			Status: Failed,
			Items: []*Step{
				{
					Name:     failedTest.Name(),
					Status:   failedTest.Status,
					Duration: failedTest.Duration,
					Config:   failedTest.Config,
					Items:    failedTest.Steps,
				},
			},
		}
		if !rootStep.Equal(expectedStep) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", rootStep, expectedStep)
		}
	})

	t.Run("canceled initial status", func(t *testing.T) {
		rootStep := &Step{Name: "root", Status: Canceled}
		rootStep.AddTest(failedTest)
		expectedStep := &Step{
			Name: rootStep.Name,
			// Canceled status should not be changed
			Status: Canceled,
			Items: []*Step{
				{
					Name:     failedTest.Name(),
					Status:   failedTest.Status,
					Duration: failedTest.Duration,
					Config:   failedTest.Config,
					Items:    failedTest.Steps,
				},
			},
		}
		if !rootStep.Equal(expectedStep) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", rootStep, expectedStep)
		}
	})
}

func TestStepAddCanceledTest(t *testing.T) {
	t.Run("failed initial status", func(t *testing.T) {
		rootStep := &Step{Name: "root", Status: Failed}
		canceledTest := &Test{
			TestContext: &Context{name: "canceled_test"},
			Status:      Canceled,
			Duration:    1.0,
			Steps: []*Step{
				{Name: "deploy", Status: Canceled, Duration: 1.0},
			},
		}
		rootStep.AddTest(canceledTest)
		expectedStep := &Step{
			Name: rootStep.Name,
			// Failed status should be overridden with Canceled
			Status: canceledTest.Status,
			Items: []*Step{
				{
					Name:     canceledTest.Name(),
					Status:   canceledTest.Status,
					Duration: canceledTest.Duration,
					Items:    canceledTest.Steps,
				},
			},
		}
		if !rootStep.Equal(expectedStep) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", rootStep, expectedStep)
		}
	})

	t.Run("canceled initial status", func(t *testing.T) {
		rootStep := &Step{Name: "root", Status: Canceled}
		failedTest := &Test{
			TestContext: &Context{name: "failed_test"},
			Status:      Failed,
			Duration:    1.0,
			Steps: []*Step{
				{Name: "deploy", Status: Failed, Duration: 1.0},
			},
		}
		rootStep.AddTest(failedTest)
		expectedStep := &Step{
			Name: rootStep.Name,
			// Status should still be Canceled, not overridden by Failed
			Status: Canceled,
			Items: []*Step{
				{
					Name:     failedTest.Name(),
					Status:   failedTest.Status,
					Duration: failedTest.Duration,
					Items:    failedTest.Steps,
				},
			},
		}
		if !rootStep.Equal(expectedStep) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", rootStep, expectedStep)
		}
	})
}

func TestStepAddSkippedTest(t *testing.T) {
	skippedTest := &Test{
		TestContext: &Context{name: "skipped_test"},
		Status:      Skipped,
		Duration:    0.0,
	}
	t.Run("empty initial status", func(t *testing.T) {
		rootStep := &Step{Name: "root"}
		rootStep.AddTest(skippedTest)
		expectedStep := &Step{
			Name: rootStep.Name,
			// Skipped tests get Passed status when parent has no status
			Status: Passed,
			Items: []*Step{
				{
					Name:     skippedTest.Name(),
					Status:   skippedTest.Status,
					Duration: skippedTest.Duration,
				},
			},
		}
		if !rootStep.Equal(expectedStep) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", rootStep, expectedStep)
		}
	})

	t.Run("failed initial status", func(t *testing.T) {
		rootStep := &Step{Name: "root", Status: Failed}
		rootStep.AddTest(skippedTest)
		expectedStep := &Step{
			Name: rootStep.Name,
			// Failed status should not change
			Status: Failed,
			Items: []*Step{
				{
					Name:     skippedTest.Name(),
					Status:   skippedTest.Status,
					Duration: skippedTest.Duration,
				},
			},
		}
		if !rootStep.Equal(expectedStep) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", rootStep, expectedStep)
		}
	})
}
func TestStepMarshal(t *testing.T) {
	step := &Step{
		Name:     "test",
		Status:   Passed,
		Duration: 2.0,
		Config:   &config.Tests[0],
		Items: []*Step{
			{Name: "subtest1", Status: Passed, Duration: 1.0},
			{Name: "subtest2", Status: Failed, Duration: 1.0},
		},
	}

	// Marshal and unmarshal the step
	bytes, err := yaml.Marshal(step)
	if err != nil {
		t.Fatalf("failed to marshal step: %v", err)
	}
	unmarshaledStep := &Step{}
	if err := yaml.Unmarshal(bytes, unmarshaledStep); err != nil {
		t.Fatalf("failed to unmarshal step: %v", err)
	}
	if !step.Equal(unmarshaledStep) {
		t.Fatalf("unmarshalled step %+v, got %+v", step, unmarshaledStep)
	}
}

func TestStepEqual(t *testing.T) {
	baseStep := Step{Name: "base_test", Status: Passed, Duration: 1.0, Config: &config.Tests[0]}

	t.Run("equal to self", func(t *testing.T) {
		if !baseStep.Equal(&baseStep) {
			t.Fatalf("step should be equal to itself")
		}
	})

	t.Run("not equal to nil", func(t *testing.T) {
		if baseStep.Equal(nil) {
			t.Fatalf("step should not be equal to nil")
		}
	})

	t.Run("different name", func(t *testing.T) {
		differentStep := baseStep
		differentStep.Name = "new_test"
		if baseStep.Equal(&differentStep) {
			t.Fatalf("steps with different names should not be equal")
		}
	})

	t.Run("different status", func(t *testing.T) {
		differentStep := baseStep
		differentStep.Status = Failed
		if baseStep.Equal(&differentStep) {
			t.Fatalf("steps with different status should not be equal")
		}
	})

	t.Run("different duration", func(t *testing.T) {
		differentStep := baseStep
		differentStep.Duration = 2.0
		if baseStep.Equal(&differentStep) {
			t.Fatalf("steps with different duration should not be equal")
		}
	})

	t.Run("different config", func(t *testing.T) {
		differentStep := baseStep
		differentStep.Config = &config.Tests[1]
		if baseStep.Equal(&differentStep) {
			t.Fatalf("steps with different config should not be equal")
		}
	})

	t.Run("one nil config", func(t *testing.T) {
		differentStep := baseStep
		differentStep.Config = nil
		if baseStep.Equal(&differentStep) {
			t.Fatalf("step with config should not be equal to step without config")
		}
	})

	t.Run("both nil config", func(t *testing.T) {
		stepA := Step{Name: "test", Status: Passed, Duration: 1.0, Config: nil}
		stepB := stepA
		if !stepA.Equal(&stepB) {
			t.Fatalf("steps with nil config should be equal")
		}
	})

	t.Run("equal subitems", func(t *testing.T) {
		stepA := Step{
			Name:     "parent",
			Status:   Passed,
			Duration: 1.0,
			Items:    []*Step{{Name: "child1", Status: Passed}, {Name: "child2", Status: Failed}},
		}
		stepB := stepA
		if !stepA.Equal(&stepB) {
			t.Fatalf("steps with equal subitems should be equal")
		}
	})

	t.Run("different subitem name", func(t *testing.T) {
		stepA := Step{
			Name:     "parent",
			Status:   Passed,
			Duration: 1.0,
			Items:    []*Step{{Name: "child1", Status: Passed}, {Name: "child2", Status: Failed}},
		}
		stepB := stepA
		stepB.Items = []*Step{{Name: "child1", Status: Passed}, {Name: "different", Status: Failed}}
		if stepA.Equal(&stepB) {
			t.Fatalf("steps with different subitem names should not be equal")
		}
	})

	t.Run("different number of subitems", func(t *testing.T) {
		stepA := Step{
			Name:     "parent",
			Status:   Passed,
			Duration: 1.0,
			Items:    []*Step{{Name: "child1", Status: Passed}, {Name: "child2", Status: Failed}},
		}
		stepB := stepA
		stepB.Items = []*Step{{Name: "child1", Status: Passed}}
		if stepA.Equal(&stepB) {
			t.Fatalf("steps with different number of subitems should not be equal")
		}
	})
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
