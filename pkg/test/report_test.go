// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"slices"
	"testing"

	"sigs.k8s.io/yaml"

	e2econfig "github.com/ramendr/ramen/e2e/config"

	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/time"
)

var reportConfig = &e2econfig.Config{
	Distro:     "k8s",
	Repo:       e2econfig.Repo{URL: "https://github.com/org/repo", Branch: "main"},
	DRPolicy:   "dr-policy",
	ClusterSet: "clusterset",
	Clusters: map[string]e2econfig.Cluster{
		"hub": {Kubeconfig: "hub-kubeconfig"},
		"c1":  {Kubeconfig: "c1-kubeconfig"},
		"c2":  {Kubeconfig: "c2-kubeconfig"},
	},
	PVCSpecs: []e2econfig.PVCSpec{
		{Name: "rbd", StorageClassName: "rook-ceph-block", AccessModes: "ReadWriteOnce"},
		{Name: "cephfs", StorageClassName: "rook-cephfs-fs", AccessModes: "ReadWriteMany"},
	},
	Tests: []e2econfig.Test{
		{Workload: "appset", Deployer: "deploy", PVCSpec: "rbd"},
		{Workload: "subscr", Deployer: "deploy", PVCSpec: "rbd"},
		{Workload: "disapp", Deployer: "deploy", PVCSpec: "cephfs"},
	},
	Channel: e2econfig.Channel{
		Name:      "my-channel",
		Namespace: "test-gitops",
	},
	Namespaces: e2econfig.Namespaces{
		RamenHubNamespace:       "ramen-system",
		RamenDRClusterNamespace: "ramen-system",
		RamenOpsNamespace:       "ramen-ops",
		ArgocdNamespace:         "argocd",
	},
}

func TestReportEmpty(t *testing.T) {
	fakeTime(t)
	r := newReport("test-run", reportConfig)

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
	passedStep := &Step{Name: "passed_step", Status: report.Passed, Duration: 1.0}

	// Adding a passed test should set the report status.
	t.Run("empty", func(t *testing.T) {
		r := newReport("test-command", reportConfig)
		r.AddStep(passedStep)
		if r.Status != report.Passed {
			t.Errorf("expected status %s, got %s", report.Passed, r.Status)
		}
		if !slices.Equal(r.Steps, []*Step{passedStep}) {
			t.Errorf("expected steps to br equal, got %v", r.Steps)
		}
	})

	// Adding a passed test should not modify failed status.
	t.Run("failed", func(t *testing.T) {
		r := newReport("test-command", reportConfig)
		r.Status = report.Failed
		r.AddStep(passedStep)
		if r.Status != report.Failed {
			t.Errorf("expected status %s, got %s", report.Failed, r.Status)
		}
		if !slices.Equal(r.Steps, []*Step{passedStep}) {
			t.Errorf("expected steps to br equal, got %v", r.Steps)
		}
	})

	// Adding a passed test should not modify canceled status.
	t.Run("failed", func(t *testing.T) {
		r := newReport("test-command", reportConfig)
		r.Status = report.Canceled
		r.AddStep(passedStep)
		if r.Status != report.Canceled {
			t.Errorf("expected status %s, got %s", report.Canceled, r.Status)
		}
		if !slices.Equal(r.Steps, []*Step{passedStep}) {
			t.Errorf("expected steps to br equal, got %v", r.Steps)
		}
	})
}

func TestReportAddFailedStep(t *testing.T) {
	fakeTime(t)
	failedStep := &Step{Name: "failed_step", Status: report.Failed, Duration: 1.0}

	// Failed status should override existing Passed status.
	t.Run("passed", func(t *testing.T) {
		r := newReport("test-command", reportConfig)
		r.Status = report.Passed
		r.AddStep(failedStep)
		if r.Status != report.Failed {
			t.Errorf("expected status %s, got %s", report.Failed, r.Status)
		}
		if !slices.Equal(r.Steps, []*Step{failedStep}) {
			t.Errorf("expected steps to be equal, got %v", r.Steps)
		}
	})

	// If a report is canceled, adding a failed test should not change the status.
	t.Run("canceled", func(t *testing.T) {
		r := newReport("test-command", reportConfig)
		r.Status = report.Canceled
		r.AddStep(failedStep)
		if r.Status != report.Canceled {
			t.Errorf("expected status %s, got %s", report.Canceled, r.Status)
		}
		if !slices.Equal(r.Steps, []*Step{failedStep}) {
			t.Errorf("expected steps to be equal, got %v", r.Steps)
		}
	})
}

func TestReportAddCanceledStep(t *testing.T) {
	fakeTime(t)
	canceledStep := &Step{Name: "canceled_step", Status: report.Canceled, Duration: 1.0}

	// Adding canceled step mark the report as canceled.
	t.Run("failed", func(t *testing.T) {
		r := newReport("test-command", reportConfig)
		r.Status = report.Failed
		r.AddStep(canceledStep)
		if r.Status != report.Canceled {
			t.Errorf("expected status %s, got %s", report.Canceled, r.Status)
		}
		if !slices.Equal(r.Steps, []*Step{canceledStep}) {
			t.Errorf("expected steps to be equal, got %v", r.Steps)
		}
	})

	// Adding canceled step mark the report as canceled.
	t.Run("passed", func(t *testing.T) {
		r := newReport("test-command", reportConfig)
		r.Status = report.Passed
		r.AddStep(canceledStep)
		if r.Status != report.Canceled {
			t.Errorf("expected status %s, got %s", report.Canceled, r.Status)
		}
		if !slices.Equal(r.Steps, []*Step{canceledStep}) {
			t.Errorf("expected steps to be equal, got %v", r.Steps)
		}
	})
}

func TestReportAddSkippedStep(t *testing.T) {
	fakeTime(t)
	skippedStep := &Step{Name: "skipped-step", Status: report.Skipped, Duration: 0.0}

	// Skipped step with empty status should result in Passed.
	t.Run("empty", func(t *testing.T) {
		r := newReport("test-command", reportConfig)
		r.AddStep(skippedStep)
		if r.Status != report.Passed {
			t.Errorf("expected status %s, got %s", report.Passed, r.Status)
		}
		if !slices.Equal(r.Steps, []*Step{skippedStep}) {
			t.Errorf("expected steps to be equal, got %v", r.Steps)
		}
	})

	// Failed status should not be overridden by Skipped.
	t.Run("failed", func(t *testing.T) {
		r := newReport("test-command", reportConfig)
		r.Status = report.Failed
		r.AddStep(skippedStep)
		if r.Status != report.Failed {
			t.Errorf("expected status %s, got %s", report.Failed, r.Status)
		}
		if !slices.Equal(r.Steps, []*Step{skippedStep}) {
			t.Errorf("expected steps to be equal, got %v", r.Steps)
		}
	})
}

func TestReportDuration(t *testing.T) {
	r := newReport("test-command", reportConfig)
	steps := []*Step{
		{Name: "step1", Status: report.Passed, Duration: 1.0},
		{Name: "step2", Status: report.Passed, Duration: 1.0},
		{Name: "step3", Status: report.Skipped, Duration: 1.0},
		{Name: "step4", Status: report.Failed, Duration: 1.0},
		{Name: "step5", Status: report.Canceled, Duration: 1.0},
	}
	for _, step := range steps {
		r.AddStep(step)
	}
	var expectedDuration float64
	for _, s := range r.Steps {
		expectedDuration += s.Duration
	}
	if r.Duration != expectedDuration {
		t.Errorf("expected duration %f, got %f", expectedDuration, r.Duration)
	}
}

func TestReportAddDuplicateStep(t *testing.T) {
	r := newReport("test-command", reportConfig)
	step := &Step{Name: "unique-step", Status: report.Passed, Duration: 1.0}
	r.AddStep(step)

	// Adding another step with the same name should panic.
	dup := &Step{Name: "unique-step", Status: report.Failed, Duration: 1.0}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic when adding duplicate step, but it didn't happen")
		}
	}()
	r.AddStep(dup)
}

func TestReportSummary(t *testing.T) {
	r := newReport("test-command", reportConfig)
	testsStep := &Step{
		Name:     TestsStep,
		Status:   report.Passed,
		Duration: 1.0,
		Items: []*Step{
			{Name: "test1", Status: report.Passed, Duration: 1.0},
			{Name: "test2", Status: report.Failed, Duration: 1.0},
			{Name: "test3", Status: report.Skipped, Duration: 1.0},
			{Name: "test4", Status: report.Canceled, Duration: 1.0},
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
		r := newReport("test-command", reportConfig)
		r.Status = report.Passed
		r.Duration = 1.0
		r.Steps = []*Step{
			{Name: "step1", Status: report.Passed, Duration: 1.0},
			{Name: "step2", Status: report.Passed, Duration: 1.0},
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
		r2 := createReport()
		if !r1.Equal(r2) {
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

	t.Run("same config reference", func(t *testing.T) {
		r2 := createReport()
		if !r1.Equal(r2) {
			t.Error("reports with equal configs should be equal")
		}
	})

	t.Run("different config content", func(t *testing.T) {
		r2 := createReport()
		differentConfig := *reportConfig
		differentConfig.DRPolicy = "different-dr-policy"
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

	t.Run("different summary", func(t *testing.T) {
		r2 := createReport()
		r2.Summary = Summary{Passed: 1, Failed: 1}
		if r1.Equal(r2) {
			t.Error("reports with different summary should not be equal")
		}
	})

	t.Run("different duration", func(t *testing.T) {
		r2 := createReport()
		r2.Duration += 1.0
		if r1.Equal(r2) {
			t.Error("reports with different duration should not be equal")
		}
	})

	t.Run("different step length", func(t *testing.T) {
		r2 := createReport()
		r2.Steps = []*Step{
			{Name: "step1", Status: report.Passed, Duration: 1.0},
		}
		if r1.Equal(r2) {
			t.Error("reports with different step counts should not be equal")
		}
	})

	t.Run("different steps content", func(t *testing.T) {
		r2 := createReport()
		r2.Steps = []*Step{
			{Name: "step1", Status: report.Passed, Duration: 1.0},
			{Name: "different", Status: report.Passed, Duration: 2.0},
		}
		if r1.Equal(r2) {
			t.Error("reports with different step content should not be equal")
		}
	})
}

func TestReportMarshaling(t *testing.T) {
	fakeTime(t)
	r := newReport("test-command", reportConfig)
	r.Status = report.Failed
	r.Duration = 2.0
	r.Steps = []*Step{
		{
			Name:     "step1",
			Status:   report.Passed,
			Duration: 1.0,
			Items: []*Step{
				{Name: "subitem1", Status: report.Passed, Duration: 1.0},
				{Name: "subitem2", Status: report.Passed, Duration: 1.0},
			},
		},
		{
			Name:     "step2",
			Status:   report.Failed,
			Duration: 1.0,
		},
	}
	r.Summary = Summary{Passed: 2, Failed: 1}

	// Test roundtrip marshaling/unmarshaling
	checkRoundtrip(t, r)
}

func TestStepAddPassedTest(t *testing.T) {
	passedTest := &Test{
		Context:  &Context{name: "passing_test"},
		Status:   report.Passed,
		Config:   &reportConfig.Tests[0],
		Duration: 6.0,
		Steps: []*Step{
			{Name: "deploy", Status: report.Passed, Duration: 1.0},
			{Name: "protect", Status: report.Passed, Duration: 1.0},
			{Name: "failover", Status: report.Passed, Duration: 1.0},
			{Name: "relocate", Status: report.Passed, Duration: 1.0},
			{Name: "unprotect", Status: report.Passed, Duration: 1.0},
			{Name: "undeploy", Status: report.Passed, Duration: 1.0},
		},
	}

	// Adding passed step should set step status.
	t.Run("empty", func(t *testing.T) {
		s1 := &Step{Name: "root"}
		s1.AddTest(passedTest)
		s2 := &Step{
			Name:   s1.Name,
			Status: report.Passed,
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
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})

	// If a report is failed adding new step should not change the status.
	t.Run("failed", func(t *testing.T) {
		s1 := &Step{Name: "root", Status: report.Failed}
		s1.AddTest(passedTest)
		s2 := &Step{
			Name:   s1.Name,
			Status: report.Failed,
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
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})

	// If a report is failed adding new step should not change the status.
	t.Run("canceled", func(t *testing.T) {
		s1 := &Step{Name: "root", Status: report.Canceled}
		s1.AddTest(passedTest)
		s2 := &Step{
			Name:   s1.Name,
			Status: report.Canceled,
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
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})
}

func TestStepAddFailedTest(t *testing.T) {
	failedTest := &Test{
		Context:  &Context{name: "failing_test"},
		Status:   report.Failed,
		Config:   &reportConfig.Tests[0],
		Duration: 1.0,
		Steps: []*Step{
			{Name: "undeploy", Status: report.Failed, Duration: 1.0},
		},
	}

	// Adding failed test should set report status.
	t.Run("empty", func(t *testing.T) {
		s1 := &Step{Name: "root"}
		s1.AddTest(failedTest)
		s2 := &Step{
			Name:   s1.Name,
			Status: report.Failed,
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
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})

	// Adding failed tests shuuld change report status.
	t.Run("passed", func(t *testing.T) {
		s1 := &Step{Name: "root", Status: report.Passed}
		s1.AddTest(failedTest)
		s2 := &Step{
			Name: s1.Name,
			// Passed status should be changed to Failed
			Status: report.Failed,
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
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})

	// Adding failed step should not change canceled report.
	t.Run("canceled", func(t *testing.T) {
		s1 := &Step{Name: "root", Status: report.Canceled}
		s1.AddTest(failedTest)
		s2 := &Step{
			Name:   s1.Name,
			Status: report.Canceled,
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
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})
}

func TestStepAddCanceledTest(t *testing.T) {
	canceledTest := &Test{
		Context:  &Context{name: "canceled_test"},
		Status:   report.Canceled,
		Duration: 1.0,
		Steps: []*Step{
			{Name: "deploy", Status: report.Canceled, Duration: 1.0},
		},
	}

	// Adding canceled test should override failed status.
	t.Run("failed", func(t *testing.T) {
		s1 := &Step{Name: "root", Status: report.Failed}
		s1.AddTest(canceledTest)
		s2 := &Step{
			Name:   s1.Name,
			Status: report.Canceled,
			Items: []*Step{
				{
					Name:     canceledTest.Name(),
					Status:   canceledTest.Status,
					Duration: canceledTest.Duration,
					Items:    canceledTest.Steps,
				},
			},
		}
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})

	// Adding canceled test should override passed status.
	t.Run("passed", func(t *testing.T) {
		s1 := &Step{Name: "root", Status: report.Passed}
		s1.AddTest(canceledTest)
		s2 := &Step{
			Name:   s1.Name,
			Status: report.Canceled,
			Items: []*Step{
				{
					Name:     canceledTest.Name(),
					Status:   canceledTest.Status,
					Duration: canceledTest.Duration,
					Items:    canceledTest.Steps,
				},
			},
		}
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})
}

func TestStepAddSkippedTest(t *testing.T) {
	skippedTest := &Test{
		Context:  &Context{name: "skipped_test"},
		Status:   report.Skipped,
		Duration: 0.0,
	}

	// Adding skipped test should set status to passed.
	t.Run("empty", func(t *testing.T) {
		s1 := &Step{Name: "root"}
		s1.AddTest(skippedTest)
		s2 := &Step{
			Name:   s1.Name,
			Status: report.Passed,
			Items: []*Step{
				{
					Name:     skippedTest.Name(),
					Status:   skippedTest.Status,
					Duration: skippedTest.Duration,
				},
			},
		}
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})

	// Adding skipped test should not modify failed status.
	t.Run("failed", func(t *testing.T) {
		s1 := &Step{Name: "root", Status: report.Failed}
		s1.AddTest(skippedTest)
		s2 := &Step{
			Name:   s1.Name,
			Status: report.Failed,
			Items: []*Step{
				{
					Name:     skippedTest.Name(),
					Status:   skippedTest.Status,
					Duration: skippedTest.Duration,
				},
			},
		}
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})
}
func TestStepMarshal(t *testing.T) {
	step := &Step{
		Name:     "test",
		Status:   report.Passed,
		Duration: 2.0,
		Config:   &reportConfig.Tests[0],
		Items: []*Step{
			{Name: "subtest1", Status: report.Passed, Duration: 1.0},
			{Name: "subtest2", Status: report.Failed, Duration: 1.0},
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
	s1 := Step{
		Name:     "base_test",
		Status:   report.Passed,
		Duration: 1.0,
		Config:   &reportConfig.Tests[0],
	}

	t.Run("equal to self", func(t *testing.T) {
		if !s1.Equal(&s1) {
			t.Fatalf("step should be equal to itself")
		}
	})

	t.Run("not equal to nil", func(t *testing.T) {
		if s1.Equal(nil) {
			t.Fatalf("step should not be equal to nil")
		}
	})

	t.Run("different name", func(t *testing.T) {
		s2 := s1
		s2.Name = "new_test"
		if s1.Equal(&s2) {
			t.Fatalf("steps with different names should not be equal")
		}
	})

	t.Run("different status", func(t *testing.T) {
		s2 := s1
		s2.Status = report.Failed
		if s1.Equal(&s2) {
			t.Fatalf("steps with different status should not be equal")
		}
	})

	t.Run("different duration", func(t *testing.T) {
		s2 := s1
		s2.Duration = 2.0
		if s1.Equal(&s2) {
			t.Fatalf("steps with different duration should not be equal")
		}
	})

	t.Run("different config", func(t *testing.T) {
		s2 := s1
		s2.Config = &reportConfig.Tests[1]
		if s1.Equal(&s2) {
			t.Fatalf("steps with different config should not be equal")
		}
	})

	t.Run("one nil config", func(t *testing.T) {
		s2 := s1
		s2.Config = nil
		if s1.Equal(&s2) {
			t.Fatalf("step with config should not be equal to step without config")
		}
	})

	t.Run("both nil config", func(t *testing.T) {
		s1 := Step{Name: "test", Status: report.Passed, Duration: 1.0, Config: nil}
		s2 := s1
		if !s1.Equal(&s2) {
			t.Fatalf("steps with nil config should be equal")
		}
	})

	t.Run("equal subitems", func(t *testing.T) {
		s1 := Step{
			Name:     "parent",
			Status:   report.Passed,
			Duration: 1.0,
			Items: []*Step{
				{Name: "child1", Status: report.Passed},
				{Name: "child2", Status: report.Failed},
			},
		}
		s2 := s1
		if !s1.Equal(&s2) {
			t.Fatalf("steps with equal subitems should be equal")
		}
	})

	t.Run("different subitem name", func(t *testing.T) {
		s1 := Step{
			Name:     "parent",
			Status:   report.Passed,
			Duration: 1.0,
			Items: []*Step{
				{Name: "child1", Status: report.Passed},
				{Name: "child2", Status: report.Failed},
			},
		}
		s2 := s1
		s2.Items = []*Step{
			{Name: "child1", Status: report.Passed},
			{Name: "different", Status: report.Failed},
		}
		if s1.Equal(&s2) {
			t.Fatalf("steps with different subitem names should not be equal")
		}
	})

	t.Run("different number of subitems", func(t *testing.T) {
		s1 := Step{
			Name:     "parent",
			Status:   report.Passed,
			Duration: 1.0,
			Items: []*Step{
				{Name: "child1", Status: report.Passed},
				{Name: "child2", Status: report.Failed},
			},
		}
		s2 := s1
		s2.Items = []*Step{{Name: "child1", Status: report.Passed}}
		if s1.Equal(&s2) {
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
	summary.AddTest(&Step{Status: report.Passed})
	summary.AddTest(&Step{Status: report.Passed})
	summary.AddTest(&Step{Status: report.Failed})
	summary.AddTest(&Step{Status: report.Skipped})
	summary.AddTest(&Step{Status: report.Canceled})
	summary.AddTest(&Step{Status: report.Passed})

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
