// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
)

func TestSummaryAdd(t *testing.T) {
	s := Summary{}

	s.Add(&report.Validated{State: report.OK})
	s.Add(&report.Validated{State: report.OK})
	s.Add(&report.Validated{State: report.Stale})
	s.Add(&report.Validated{State: report.OK})
	s.Add(&report.Validated{State: report.Stale})
	s.Add(&report.Validated{State: report.Problem})

	expected := Summary{OK: 3, Stale: 2, Problem: 1}
	if s != expected {
		t.Fatalf("expected %+v, got %+v", expected, s)
	}
}

func TestSummaryHasProblems(t *testing.T) {
	cases := []struct {
		name     string
		summary  Summary
		expected bool
	}{
		{"empty", Summary{}, false},
		{"ok", Summary{OK: 5}, false},
		{"only stale", Summary{Stale: 2}, true},
		{"only problem", Summary{Problem: 4}, true},
		{"problem and stale", Summary{Stale: 2, Problem: 3}, true},
	}
	for _, tc := range cases {
		if got := tc.summary.HasIssues(); got != tc.expected {
			t.Errorf("%s: expected %v, got %v", tc.name, tc.expected, got)
		}
	}
}

func TestSummaryString(t *testing.T) {
	s := Summary{OK: 1, Stale: 0, Problem: 2}
	expected := "1 ok, 0 stale, 2 problem"
	if s.String() != expected {
		t.Fatalf("expected %q, got %q", expected, s.String())
	}
}

func TestReportEqual(t *testing.T) {
	helpers.FakeTime(t)
	r1 := &Report{Report: report.NewReport("name", &config.Config{})}
	t.Run("equal to self", func(t *testing.T) {
		r2 := r1
		if !r1.Equal(r2) {
			t.Fatal("report should be equal to itself")
		}
	})
	t.Run("equal reports", func(t *testing.T) {
		r2 := &Report{Report: report.NewReport("name", &config.Config{})}
		if !r1.Equal(r2) {
			t.Fatalf("expected report %+v, got %+v", r1, r2)
		}
	})
}

func TestReportNotEqual(t *testing.T) {
	helpers.FakeTime(t)
	r1 := &Report{Report: report.NewReport("name", &config.Config{})}
	t.Run("nil", func(t *testing.T) {
		if r1.Equal(nil) {
			t.Fatal("report should not be equal to nil")
		}
	})
	t.Run("report", func(t *testing.T) {
		r2 := &Report{Report: report.NewReport("other", &config.Config{})}
		if r1.Equal(r2) {
			t.Fatal("reports with different report should not be equal")
		}
	})
	t.Run("summary", func(t *testing.T) {
		r1 := &Report{Report: report.NewReport("name", &config.Config{})}
		r1.Summary = Summary{OK: 5}
		r2 := &Report{Report: report.NewReport("name", &config.Config{})}
		r2.Summary = Summary{OK: 3}
		if r1.Equal(r2) {
			t.Fatal("reports with different summary should not be equal")
		}
	})
}

func TestReportRoundtrip(t *testing.T) {
	r1 := &Report{
		Report:  report.NewReport("name", &config.Config{}),
		Summary: Summary{OK: 3, Stale: 2, Problem: 1},
	}
	b, err := yaml.Marshal(r1)
	if err != nil {
		t.Fatalf("failed to marshal yaml: %s", err)
	}
	r2 := &Report{}
	if err := yaml.Unmarshal(b, r2); err != nil {
		t.Fatalf("failed to unmarshal yaml: %s", err)
	}
	if !r1.Equal(r2) {
		t.Fatalf("expected report %+v, got %+v", r1, r2)
	}
}
