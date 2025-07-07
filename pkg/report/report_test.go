// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report_test

import (
	"runtime"
	"slices"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/build"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/time"
)

func TestBaseHost(t *testing.T) {
	r := report.NewBase("name")
	expected := report.Host{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
		Cpus: runtime.NumCPU(),
	}
	if r.Host != expected {
		t.Fatalf("expected host %+v, got %+v", expected, r.Host)
	}
}

func TestBuildInfo(t *testing.T) {
	savedVersion := build.Version
	savedCommit := build.Commit
	t.Cleanup(func() {
		build.Version = savedVersion
		build.Commit = savedCommit
	})
	t.Run("available", func(t *testing.T) {
		build.Version = "fake-version"
		build.Commit = "fake-commit"
		r := report.NewBase("name")
		if r.Build == nil {
			t.Fatalf("build info omitted")
		}
		expected := report.Build{
			Version: build.Version,
			Commit:  build.Commit,
		}
		if *r.Build != expected {
			t.Fatalf("expected build info %+v, got %+v", expected, r.Build)
		}
	})
	t.Run("missing", func(t *testing.T) {
		build.Version = ""
		build.Commit = ""
		r := report.NewBase("name")
		if r.Build != nil {
			t.Fatalf("build info not omitted: %+v", r.Build)
		}
	})
}

func TestBaseCreatedTime(t *testing.T) {
	fakeTime(t)
	r := report.NewBase("name")
	if r.Created != time.Now().Local() {
		t.Fatalf("expected %v, got %v", time.Now().Local(), r.Created)
	}
}

func TestBaseRoundtrip(t *testing.T) {
	r1 := report.NewBase("name")
	b, err := yaml.Marshal(r1)
	if err != nil {
		t.Fatalf("failed to marshal yaml: %s", err)
	}
	r2 := &report.Base{}
	if err := yaml.Unmarshal(b, r2); err != nil {
		t.Fatalf("failed to unmarshal yaml: %s", err)
	}
	if !r1.Equal(r2) {
		t.Fatalf("expected report %+v, got %+v", r1, r2)
	}
}

func TestBaseEqual(t *testing.T) {
	fakeTime(t)
	r1 := report.NewBase("name")
	t.Run("equal to self", func(t *testing.T) {
		r2 := r1
		if !r1.Equal(r2) {
			t.Fatal("report should be equal itself")
		}
	})
	t.Run("equal reports", func(t *testing.T) {
		r2 := report.NewBase("name")
		if !r1.Equal(r2) {
			t.Fatalf("expected report %+v, got %+v", r1, r2)
		}
	})
}

func TestBaseNotEqual(t *testing.T) {
	fakeTime(t)
	r1 := report.NewBase("name")
	t.Run("nil", func(t *testing.T) {
		if r1.Equal(nil) {
			t.Fatal("report should not be equal to nil")
		}
	})
	t.Run("created", func(t *testing.T) {
		r2 := report.NewBase("name")
		r2.Created = r2.Created.Add(1)
		if r1.Equal(r2) {
			t.Fatal("reports with different create time should not be equal")
		}
	})
	t.Run("host", func(t *testing.T) {
		r2 := report.NewBase("name")
		r2.Host.OS = "modified"
		if r1.Equal(r2) {
			t.Fatal("reports with different host should not be equal")
		}
	})
	t.Run("build", func(t *testing.T) {
		r2 := report.NewBase("name")
		// Build is either nil or have version and commit, empty Build cannot match.
		r2.Build = &report.Build{}
		if r1.Equal(r2) {
			t.Fatal("reports with different host should not be equal")
		}
	})
	t.Run("name", func(t *testing.T) {
		r2 := report.NewBase("other")
		if r1.Equal(r2) {
			t.Error("reports with different names should not be equal")
		}
	})
	t.Run("status", func(t *testing.T) {
		r2 := report.NewBase("name")
		r2.Status = report.Failed
		if r1.Equal(r2) {
			t.Fatal("reports with different status should not be equal")
		}
	})
	t.Run("duration", func(t *testing.T) {
		r2 := report.NewBase("name")
		r2.Duration += 1.0
		if r1.Equal(r2) {
			t.Error("reports with different duration should not be equal")
		}
	})
	t.Run("steps", func(t *testing.T) {
		r2 := report.NewBase("name")
		r2.Steps = []*report.Step{
			{Name: "new step", Status: report.Passed, Duration: 1.0},
		}
		if r1.Equal(r2) {
			t.Error("reports with different step should not be equal")
		}
	})
}

func TestReportDuration(t *testing.T) {
	r := report.NewBase("name")
	steps := []*report.Step{
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

func TestBaseAddDuplicateStep(t *testing.T) {
	r := report.NewBase("name")
	step := &report.Step{Name: "unique", Status: report.Passed, Duration: 1.0}
	r.AddStep(step)

	// Adding another step with the same name should panic.
	dup := &report.Step{Name: "unique", Status: report.Failed, Duration: 1.0}
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected panic when adding duplicate step, but it didn't happen")
		}
	}()
	r.AddStep(dup)
}

func TestReportAddPassedStep(t *testing.T) {
	passedStep := &report.Step{Name: "ok", Status: report.Passed, Duration: 1.0}

	// Adding a passed test should set the report status.
	t.Run("empty", func(t *testing.T) {
		r := report.NewBase("name")
		r.AddStep(passedStep)
		if r.Status != report.Passed {
			t.Errorf("expected status %s, got %s", report.Passed, r.Status)
		}
		if !slices.Equal(r.Steps, []*report.Step{passedStep}) {
			t.Errorf("expected steps to br equal, got %v", r.Steps)
		}
	})

	// Adding a passed test should not modify failed status.
	t.Run("failed", func(t *testing.T) {
		r := report.NewBase("name")
		r.Status = report.Failed
		r.AddStep(passedStep)
		if r.Status != report.Failed {
			t.Errorf("expected status %s, got %s", report.Failed, r.Status)
		}
		if !slices.Equal(r.Steps, []*report.Step{passedStep}) {
			t.Errorf("expected steps to br equal, got %v", r.Steps)
		}
	})

	// Adding a passed test should not modify canceled status.
	t.Run("failed", func(t *testing.T) {
		r := report.NewBase("name")
		r.Status = report.Canceled
		r.AddStep(passedStep)
		if r.Status != report.Canceled {
			t.Errorf("expected status %s, got %s", report.Canceled, r.Status)
		}
		if !slices.Equal(r.Steps, []*report.Step{passedStep}) {
			t.Errorf("expected steps to br equal, got %v", r.Steps)
		}
	})
}

func TestReportAddFailedStep(t *testing.T) {
	failedStep := &report.Step{Name: "fail", Status: report.Failed, Duration: 1.0}

	// Failed status should override existing Passed status.
	t.Run("passed", func(t *testing.T) {
		r := report.NewBase("name")
		r.Status = report.Passed
		r.AddStep(failedStep)
		if r.Status != report.Failed {
			t.Errorf("expected status %s, got %s", report.Failed, r.Status)
		}
		if !slices.Equal(r.Steps, []*report.Step{failedStep}) {
			t.Errorf("expected steps to be equal, got %v", r.Steps)
		}
	})

	// If a report is canceled, adding a failed test should not change the status.
	t.Run("canceled", func(t *testing.T) {
		r := report.NewBase("name")
		r.Status = report.Canceled
		r.AddStep(failedStep)
		if r.Status != report.Canceled {
			t.Errorf("expected status %s, got %s", report.Canceled, r.Status)
		}
		if !slices.Equal(r.Steps, []*report.Step{failedStep}) {
			t.Errorf("expected steps to be equal, got %v", r.Steps)
		}
	})
}

func TestReportAddCanceledStep(t *testing.T) {
	canceledStep := &report.Step{Name: "cancel", Status: report.Canceled, Duration: 1.0}

	// Adding canceled step mark the report as canceled.
	t.Run("failed", func(t *testing.T) {
		r := report.NewBase("name")
		r.Status = report.Failed
		r.AddStep(canceledStep)
		if r.Status != report.Canceled {
			t.Errorf("expected status %s, got %s", report.Canceled, r.Status)
		}
		if !slices.Equal(r.Steps, []*report.Step{canceledStep}) {
			t.Errorf("expected steps to be equal, got %v", r.Steps)
		}
	})

	// Adding canceled step mark the report as canceled.
	t.Run("passed", func(t *testing.T) {
		r := report.NewBase("name")
		r.Status = report.Passed
		r.AddStep(canceledStep)
		if r.Status != report.Canceled {
			t.Errorf("expected status %s, got %s", report.Canceled, r.Status)
		}
		if !slices.Equal(r.Steps, []*report.Step{canceledStep}) {
			t.Errorf("expected steps to be equal, got %v", r.Steps)
		}
	})
}

func TestReportAddSkippedStep(t *testing.T) {
	skippedStep := &report.Step{Name: "skip", Status: report.Skipped, Duration: 0.0}

	// Skipped step with empty status should result in Passed.
	t.Run("empty", func(t *testing.T) {
		r := report.NewBase("name")
		r.AddStep(skippedStep)
		if r.Status != report.Passed {
			t.Errorf("expected status %s, got %s", report.Passed, r.Status)
		}
		if !slices.Equal(r.Steps, []*report.Step{skippedStep}) {
			t.Errorf("expected steps to be equal, got %v", r.Steps)
		}
	})

	// Failed status should not be overridden by Skipped.
	t.Run("failed", func(t *testing.T) {
		r := report.NewBase("name")
		r.Status = report.Failed
		r.AddStep(skippedStep)
		if r.Status != report.Failed {
			t.Errorf("expected status %s, got %s", report.Failed, r.Status)
		}
		if !slices.Equal(r.Steps, []*report.Step{skippedStep}) {
			t.Errorf("expected steps to be equal, got %v", r.Steps)
		}
	})
}

func TestStepAddPassedStep(t *testing.T) {
	passedStep := &report.Step{
		Status:   report.Passed,
		Duration: 6.0,
		Items: []*report.Step{
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
		s1 := &report.Step{Name: "root"}
		s1.AddStep(passedStep)
		s2 := &report.Step{
			Name:   s1.Name,
			Status: report.Passed,
			Items: []*report.Step{
				{
					Name:     passedStep.Name,
					Status:   passedStep.Status,
					Duration: passedStep.Duration,
					Items:    passedStep.Items,
				},
			},
		}
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})

	// If a report is failed adding new step should not change the status.
	t.Run("failed", func(t *testing.T) {
		s1 := &report.Step{Name: "root", Status: report.Failed}
		s1.AddStep(passedStep)
		s2 := &report.Step{
			Name:   s1.Name,
			Status: report.Failed,
			Items: []*report.Step{
				{
					Name:     passedStep.Name,
					Status:   passedStep.Status,
					Duration: passedStep.Duration,
					Items:    passedStep.Items,
				},
			},
		}
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})

	// If a report is failed adding new step should not change the status.
	t.Run("canceled", func(t *testing.T) {
		s1 := &report.Step{Name: "root", Status: report.Canceled}
		s1.AddStep(passedStep)
		s2 := &report.Step{
			Name:   s1.Name,
			Status: report.Canceled,
			Items: []*report.Step{
				{
					Name:     passedStep.Name,
					Status:   passedStep.Status,
					Duration: passedStep.Duration,
					Items:    passedStep.Items,
				},
			},
		}
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})
}

func TestStepAddFailedStep(t *testing.T) {
	failedStep := &report.Step{
		Status:   report.Failed,
		Duration: 1.0,
		Items: []*report.Step{
			{Name: "undeploy", Status: report.Failed, Duration: 1.0},
		},
	}

	// Adding failed test should set report status.
	t.Run("empty", func(t *testing.T) {
		s1 := &report.Step{Name: "root"}
		s1.AddStep(failedStep)
		s2 := &report.Step{
			Name:   s1.Name,
			Status: report.Failed,
			Items: []*report.Step{
				{
					Name:     failedStep.Name,
					Status:   failedStep.Status,
					Duration: failedStep.Duration,
					Items:    failedStep.Items,
				},
			},
		}
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})

	// Adding failed tests shuuld change report status.
	t.Run("passed", func(t *testing.T) {
		s1 := &report.Step{Name: "root", Status: report.Passed}
		s1.AddStep(failedStep)
		s2 := &report.Step{
			Name: s1.Name,
			// Passed status should be changed to Failed
			Status: report.Failed,
			Items: []*report.Step{
				{
					Name:     failedStep.Name,
					Status:   failedStep.Status,
					Duration: failedStep.Duration,
					Items:    failedStep.Items,
				},
			},
		}
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})

	// Adding failed step should not change canceled report.
	t.Run("canceled", func(t *testing.T) {
		s1 := &report.Step{Name: "root", Status: report.Canceled}
		s1.AddStep(failedStep)
		s2 := &report.Step{
			Name:   s1.Name,
			Status: report.Canceled,
			Items: []*report.Step{
				{
					Name:     failedStep.Name,
					Status:   failedStep.Status,
					Duration: failedStep.Duration,
					Items:    failedStep.Items,
				},
			},
		}
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})
}

func TestStepAddCanceledStep(t *testing.T) {
	canceledStep := &report.Step{
		Status:   report.Canceled,
		Duration: 1.0,
		Items: []*report.Step{
			{Name: "deploy", Status: report.Canceled, Duration: 1.0},
		},
	}

	// Adding canceled test should override failed status.
	t.Run("failed", func(t *testing.T) {
		s1 := &report.Step{Name: "root", Status: report.Failed}
		s1.AddStep(canceledStep)
		s2 := &report.Step{
			Name:   s1.Name,
			Status: report.Canceled,
			Items: []*report.Step{
				{
					Name:     canceledStep.Name,
					Status:   canceledStep.Status,
					Duration: canceledStep.Duration,
					Items:    canceledStep.Items,
				},
			},
		}
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})

	// Adding canceled test should override passed status.
	t.Run("passed", func(t *testing.T) {
		s1 := &report.Step{Name: "root", Status: report.Passed}
		s1.AddStep(canceledStep)
		s2 := &report.Step{
			Name:   s1.Name,
			Status: report.Canceled,
			Items: []*report.Step{
				{
					Name:     canceledStep.Name,
					Status:   canceledStep.Status,
					Duration: canceledStep.Duration,
					Items:    canceledStep.Items,
				},
			},
		}
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})
}

func TestStepAddSkippedStep(t *testing.T) {
	skippedStep := &report.Step{
		Status:   report.Skipped,
		Duration: 0.0,
	}

	// Adding skipped test should set status to passed.
	t.Run("empty", func(t *testing.T) {
		s1 := &report.Step{Name: "root"}
		s1.AddStep(skippedStep)
		s2 := &report.Step{
			Name:   s1.Name,
			Status: report.Passed,
			Items: []*report.Step{
				{
					Name:     skippedStep.Name,
					Status:   skippedStep.Status,
					Duration: skippedStep.Duration,
				},
			},
		}
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})

	// Adding skipped test should not modify failed status.
	t.Run("failed", func(t *testing.T) {
		s1 := &report.Step{Name: "root", Status: report.Failed}
		s1.AddStep(skippedStep)
		s2 := &report.Step{
			Name:   s1.Name,
			Status: report.Failed,
			Items: []*report.Step{
				{
					Name:     skippedStep.Name,
					Status:   skippedStep.Status,
					Duration: skippedStep.Duration,
				},
			},
		}
		if !s1.Equal(s2) {
			t.Errorf("rootStep %+v doesn't match expectedStep %+v", s1, s2)
		}
	})
}

func TestStepMarshal(t *testing.T) {
	step := &report.Step{
		Name:     "test",
		Status:   report.Passed,
		Duration: 2.0,
		Items: []*report.Step{
			{Name: "subtest1", Status: report.Passed, Duration: 1.0},
			{Name: "subtest2", Status: report.Failed, Duration: 1.0},
		},
	}

	// Marshal and unmarshal the step
	bytes, err := yaml.Marshal(step)
	if err != nil {
		t.Fatalf("failed to marshal step: %v", err)
	}
	unmarshaledStep := &report.Step{}
	if err := yaml.Unmarshal(bytes, unmarshaledStep); err != nil {
		t.Fatalf("failed to unmarshal step: %v", err)
	}
	if !step.Equal(unmarshaledStep) {
		t.Fatalf("unmarshalled step %+v, got %+v", step, unmarshaledStep)
	}
}

func TestStepEqual(t *testing.T) {
	s1 := report.Step{Name: "base_test", Status: report.Passed, Duration: 1.0}

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

	t.Run("equal subitems", func(t *testing.T) {
		s1 := report.Step{
			Name:     "parent",
			Status:   report.Passed,
			Duration: 1.0,
			Items: []*report.Step{
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
		s1 := report.Step{
			Name:     "parent",
			Status:   report.Passed,
			Duration: 1.0,
			Items: []*report.Step{
				{Name: "child1", Status: report.Passed},
				{Name: "child2", Status: report.Failed},
			},
		}
		s2 := s1
		s2.Items = []*report.Step{
			{Name: "child1", Status: report.Passed},
			{Name: "different", Status: report.Failed},
		}
		if s1.Equal(&s2) {
			t.Fatalf("steps with different subitem names should not be equal")
		}
	})

	t.Run("different number of subitems", func(t *testing.T) {
		s1 := report.Step{
			Name:     "parent",
			Status:   report.Passed,
			Duration: 1.0,
			Items: []*report.Step{
				{Name: "child1", Status: report.Passed},
				{Name: "child2", Status: report.Failed},
			},
		}
		s2 := s1
		s2.Items = []*report.Step{{Name: "child1", Status: report.Passed}}
		if s1.Equal(&s2) {
			t.Fatalf("steps with different number of subitems should not be equal")
		}
	})
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
