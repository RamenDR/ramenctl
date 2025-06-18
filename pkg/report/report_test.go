// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report_test

import (
	"runtime"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/build"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/time"
)

func TestHost(t *testing.T) {
	r := report.New("name")
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
		r := report.New("name")
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
		r := report.New("name")
		if r.Build != nil {
			t.Fatalf("build info not omitted: %+v", r.Build)
		}
	})
}

func TestCreatedTime(t *testing.T) {
	fakeTime(t)
	r := report.New("name")
	if r.Created != time.Now().Local() {
		t.Fatalf("expected %v, got %v", time.Now().Local(), r.Created)
	}
}

func TestRoundtrip(t *testing.T) {
	r1 := report.New("name")
	b, err := yaml.Marshal(r1)
	if err != nil {
		t.Fatalf("failed to marshal yaml: %s", err)
	}
	r2 := &report.Report{}
	if err := yaml.Unmarshal(b, r2); err != nil {
		t.Fatalf("failed to unmarshal yaml: %s", err)
	}
	if !r1.Equal(r2) {
		t.Fatalf("expected report %+v, got %+v", r1, r2)
	}
}

func TestReportEqual(t *testing.T) {
	fakeTime(t)
	r1 := report.New("name")
	t.Run("equal to self", func(t *testing.T) {
		r2 := r1
		if !r1.Equal(r2) {
			t.Fatal("report should be equal itself")
		}
	})
	t.Run("equal reports", func(t *testing.T) {
		r2 := report.New("name")
		if !r1.Equal(r2) {
			t.Fatalf("expected report %+v, got %+v", r1, r2)
		}
	})
}

func TestReportNotEqual(t *testing.T) {
	fakeTime(t)
	r1 := report.New("name")
	t.Run("nil", func(t *testing.T) {
		if r1.Equal(nil) {
			t.Fatal("report should not be equal to nil")
		}
	})
	t.Run("created", func(t *testing.T) {
		r2 := report.New("name")
		r2.Created = r2.Created.Add(1)
		if r1.Equal(r2) {
			t.Fatal("reports with different create time should not be equal")
		}
	})
	t.Run("host", func(t *testing.T) {
		r2 := report.New("name")
		r2.Host.OS = "modified"
		if r1.Equal(r2) {
			t.Fatal("reports with different host should not be equal")
		}
	})
	t.Run("build", func(t *testing.T) {
		r2 := report.New("name")
		// Build is either nil or have version and commit, empty Build cannot match.
		r2.Build = &report.Build{}
		if r1.Equal(r2) {
			t.Fatal("reports with different host should not be equal")
		}
	})
	t.Run("name", func(t *testing.T) {
		r2 := report.New("other")
		if r1.Equal(r2) {
			t.Error("reports with different names should not be equal")
		}
	})
	t.Run("status", func(t *testing.T) {
		r2 := report.New("name")
		r2.Status = report.Failed
		if r1.Equal(r2) {
			t.Fatal("reports with different status should not be equal")
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
