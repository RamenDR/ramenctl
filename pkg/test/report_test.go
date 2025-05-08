// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"slices"
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

func TestReportAddPassedStep(t *testing.T) {
	fakeTime(t)
	r := newReport("test-command", config)
	passedStep := &Step{Name: "passed_step", Status: Passed, Duration: 1.0}
	r.AddStep(passedStep)

	if !slices.Equal(r.Steps, []*Step{passedStep}) {
		t.Errorf("expected steps to br equal, got %v", r.Steps)
	}
	if r.Status != Passed {
		t.Errorf("expected status %s, got %s", Passed, r.Status)
	}
}

func TestReportAddFailedStep(t *testing.T) {
	fakeTime(t)
	r := newReport("test-command", config)
	r.Status = Passed
	failedStep := &Step{Name: "failed_step", Status: Failed, Duration: 1.0}
	r.AddStep(failedStep)

	if !slices.Equal(r.Steps, []*Step{failedStep}) {
		t.Errorf("expected steps to be equal, got %v", r.Steps)
	}

	// Failed status should override existing Passed status
	if r.Status != Failed {
		t.Errorf("expected status %s, got %s", Failed, r.Status)
	}
}

func TestReportAddCanceledStep(t *testing.T) {
	fakeTime(t)
	t.Run("failed initial status", func(t *testing.T) {
		r := newReport("test-command", config)
		r.Status = Failed
		canceledStep := &Step{Name: "canceled_step", Status: Canceled, Duration: 1.0}
		r.AddStep(canceledStep)

		if !slices.Equal(r.Steps, []*Step{canceledStep}) {
			t.Errorf("expected steps to be equal, got %v", r.Steps)
		}

		// Canceled status should override Failed status
		if r.Status != Canceled {
			t.Errorf("expected status %s, got %s", Canceled, r.Status)
		}
	})

	t.Run("canceled initial status", func(t *testing.T) {
		r := newReport("test-command", config)
		r.Status = Canceled
		failedStep := &Step{Name: "failed_step", Status: Failed, Duration: 1.0}
		r.AddStep(failedStep)

		if !slices.Equal(r.Steps, []*Step{failedStep}) {
			t.Errorf("expected steps to be equal, got %v", r.Steps)
		}

		// Status should remain Canceled, not overridden by Failed
		if r.Status != Canceled {
			t.Errorf("expected status %s, got %s", Canceled, r.Status)
		}
	})
}

func TestReportAddSkippedStep(t *testing.T) {
	fakeTime(t)
	skippedStep := &Step{Name: "skipped-step", Status: Skipped, Duration: 0.0}
	t.Run("empty initial status", func(t *testing.T) {
		r := newReport("test-command", config)
		r.AddStep(skippedStep)

		if !slices.Equal(r.Steps, []*Step{skippedStep}) {
			t.Errorf("expected steps to be equal, got %v", r.Steps)
		}

		// Skipped step with empty status should result in Passed
		if r.Status != Passed {
			t.Errorf("expected status %s, got %s", Passed, r.Status)
		}
	})

	t.Run("failed initial status", func(t *testing.T) {
		r := newReport("test-command", config)
		r.Status = Failed
		r.AddStep(skippedStep)

		if !slices.Equal(r.Steps, []*Step{skippedStep}) {
			t.Errorf("expected steps to be equal, got %v", r.Steps)
		}

		// Failed status should not be overridden by Skipped
		if r.Status != Failed {
			t.Errorf("expected status %s, got %s", Failed, r.Status)
		}
	})
}

func TestReportDuration(t *testing.T) {
	fakeTime(t)
	r := newReport("test-command", config)
	steps := []*Step{
		{Name: "step1", Status: Passed, Duration: 1.0},
		{Name: "step2", Status: Passed, Duration: 1.0},
		{Name: "step3", Status: Skipped, Duration: 0.0},
	}
	for _, step := range steps {
		r.AddStep(step)
	}

	// Verify cumulative duration
	expectedDuration := steps[0].Duration + steps[1].Duration + steps[2].Duration
	if r.Duration != expectedDuration {
		t.Errorf("expected duration %f, got %f", expectedDuration, r.Duration)
	}

	// Add a failing step and verify duration updates
	failStep := &Step{Name: "step4", Status: Failed, Duration: 1.0}
	r.AddStep(failStep)
	expectedDuration += failStep.Duration
	if r.Duration != expectedDuration {
		t.Errorf("expected duration %f, got %f", expectedDuration, r.Duration)
	}
}

func TestReportAddDuplicateStep(t *testing.T) {
	fakeTime(t)
	r := newReport("test-command", config)
	step := &Step{Name: "unique-step", Status: Passed, Duration: 1.0}
	r.AddStep(step)

	// Add another step with the same name
	duplicateStep := &Step{Name: "unique-step", Status: Failed, Duration: 1.0}

	// AddStep should panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic when adding duplicate step, but it didn't happen")
		}
	}()
	r.AddStep(duplicateStep)
}

func TestReportSummary(t *testing.T) {
	fakeTime(t)
	r := newReport("test-command", config)
	testsStep := &Step{
		Name:     TestsStep,
		Status:   Passed,
		Duration: 1.0,
		Items: []*Step{
			{Name: "test1", Status: Passed, Duration: 1.0},
			{Name: "test2", Status: Failed, Duration: 1.0},
			{Name: "test3", Status: Skipped, Duration: 1.0},
			{Name: "test4", Status: Canceled, Duration: 1.0},
		},
	}
	r.AddStep(testsStep)

	expectedSummary := Summary{Passed: 1, Failed: 1, Skipped: 1, Canceled: 1}
	if r.Summary != expectedSummary {
		t.Errorf("expected summary %+v, got %+v", expectedSummary, r.Summary)
	}
}

func TestReportEqual(t *testing.T) {
	fakeTime(t)
	// Helper function to create a standard report
	createReport := func() *Report {
		r := newReport("test-command", config)
		r.Status = Passed
		r.Duration = 1.0
		r.Steps = []*Step{
			{Name: "step1", Status: Passed, Duration: 1.0},
			{Name: "step2", Status: Passed, Duration: 1.0},
		}
		r.Summary = Summary{Passed: 2}
		return r
	}

	r1 := createReport()

	// Intentionally comparing report to itself
	//nolint:gocritic
	t.Run("equal to self", func(t *testing.T) {
		if !r1.Equal(r1) {
			t.Error("report should be equal itself")
		}
	})

	t.Run("not equal to nil", func(t *testing.T) {
		if r1.Equal(nil) {
			t.Error("report should not be equal nil")
		}
	})

	t.Run("equal reports", func(t *testing.T) {
		equalReport := createReport()
		if !r1.Equal(equalReport) {
			t.Error("reports with identical content should be equal")
		}
	})

	t.Run("different name", func(t *testing.T) {
		r2 := createReport()
		r2.Name = "different-command"
		if r1.Equal(r2) {
			t.Error("reports with different names should not be equal")
		}
	})

	t.Run("different config reference", func(t *testing.T) {
		r2 := createReport()
		if !r1.Equal(r2) {
			t.Error("reports with equal configs should be equal")
		}
	})

	t.Run("different config content", func(t *testing.T) {
		r2 := createReport()
		differentConfig := *config
		differentConfig.DRPolicy = "differentDrPolicy"
		r2.Config = &differentConfig
		if r1.Equal(r2) {
			t.Error("reports with different config content should not be equal")
		}
	})

	t.Run("nil config", func(t *testing.T) {
		r2 := createReport()
		r2.Config = nil
		if r1.Equal(r2) || r2.Equal(r1) {
			t.Error("reports with one nil config should not be equal")
		}
	})

	t.Run("different status", func(t *testing.T) {
		r2 := createReport()
		r2.Status = Failed
		if r1.Equal(r2) {
			t.Error("reports with different status should not be equal")
		}
	})

	t.Run("different summary", func(t *testing.T) {
		r2 := createReport()
		r2.Summary = Summary{Passed: 1, Failed: 1}
		if r1.Equal(r2) {
			t.Error("reports with different summary should not be equal")
		}
	})

	t.Run("different duration", func(t *testing.T) {
		r2 := createReport()
		r2.Duration = 4.0
		if r1.Equal(r2) {
			t.Error("reports with different duration should not be equal")
		}
	})

	t.Run("different step length", func(t *testing.T) {
		r2 := createReport()
		r2.Steps = []*Step{
			{Name: "step1", Status: Passed, Duration: 1.0},
		}
		if r1.Equal(r2) {
			t.Error("reports with different step counts should not be equal")
		}
	})

	t.Run("different steps content", func(t *testing.T) {
		r2 := createReport()
		r2.Steps = []*Step{
			{Name: "step1", Status: Passed, Duration: 1.0},
			{Name: "different", Status: Passed, Duration: 2.0},
		}
		if r1.Equal(r2) {
			t.Error("reports with different step content should not be equal")
		}
	})
}

func TestReportMarshaling(t *testing.T) {
	fakeTime(t)
	r := newReport("test-command", config)
	r.Status = Failed
	r.Duration = 2.0
	r.Steps = []*Step{
		{
			Name:     "step1",
			Status:   Passed,
			Duration: 1.0,
			Items: []*Step{
				{Name: "subitem1", Status: Passed, Duration: 1.0},
				{Name: "subitem2", Status: Passed, Duration: 1.0},
			},
		},
		{
			Name:     "step2",
			Status:   Failed,
			Duration: 1.0,
		},
	}
	r.Summary = Summary{Passed: 2, Failed: 1}

	// Test roundtrip marshaling/unmarshaling
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
