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
	s := &report.Summary{}

	addValidation(s, &report.Validated{State: report.OK})
	addValidation(s, &report.Validated{State: report.OK})
	addValidation(s, &report.Validated{State: report.Stale})
	addValidation(s, &report.Validated{State: report.OK})
	addValidation(s, &report.Validated{State: report.Stale})
	addValidation(s, &report.Validated{State: report.Problem})

	expected := report.Summary{
		OK:      3,
		Stale:   2,
		Problem: 1,
	}
	if !s.Equal(&expected) {
		t.Fatalf("expected %+v, got %+v", expected, *s)
	}
}

func TestSummaryHasProblems(t *testing.T) {
	cases := []struct {
		name     string
		summary  *report.Summary
		expected bool
	}{
		{"empty", &report.Summary{}, false},
		{"ok", &report.Summary{OK: 5}, false},
		{"only stale", &report.Summary{Stale: 2}, true},
		{"only problem", &report.Summary{Problem: 4}, true},
		{
			"problem and stale",
			&report.Summary{Stale: 2, Problem: 3},
			true,
		},
	}
	for _, tc := range cases {
		if got := hasIssues(tc.summary); got != tc.expected {
			t.Errorf("%s: expected %v, got %v", tc.name, tc.expected, got)
		}
	}
}

func TestSummaryString(t *testing.T) {
	s := &report.Summary{
		OK:      1,
		Problem: 2,
	}
	expected := "1 ok, 0 stale, 2 problem"
	if summaryString(s) != expected {
		t.Fatalf("expected %q, got %q", expected, summaryString(s))
	}
}

func TestReportEqual(t *testing.T) {
	helpers.FakeTime(t)
	r1 := report.NewReport("name", &config.Config{})
	r1.Summary = &report.Summary{}
	t.Run("equal to self", func(t *testing.T) {
		r2 := r1
		if !r1.Equal(r2) {
			diff := helpers.UnifiedDiff(t, r1, r2)
			t.Fatalf("report not equal to itself\n%s", diff)
		}
	})
	t.Run("equal reports", func(t *testing.T) {
		r2 := report.NewReport("name", &config.Config{})
		r2.Summary = &report.Summary{}
		if !r1.Equal(r2) {
			diff := helpers.UnifiedDiff(t, r1, r2)
			t.Fatalf("reports not equal\n%s", diff)
		}
	})
}

func TestReportNotEqual(t *testing.T) {
	helpers.FakeTime(t)
	r1 := report.NewReport("name", &config.Config{})
	r1.Summary = &report.Summary{}
	t.Run("nil", func(t *testing.T) {
		if r1.Equal(nil) {
			t.Fatal("report should not be equal to nil")
		}
	})
	t.Run("report", func(t *testing.T) {
		r2 := report.NewReport("other", &config.Config{})
		r2.Summary = &report.Summary{}
		if r1.Equal(r2) {
			t.Fatal("reports with different report should not be equal")
		}
	})
}

func TestReportRoundtrip(t *testing.T) {
	r1 := report.NewReport("name", &config.Config{})
	r1.Summary = &report.Summary{
		OK:      3,
		Stale:   2,
		Problem: 1,
	}
	b, err := yaml.Marshal(r1)
	if err != nil {
		t.Fatalf("failed to marshal yaml: %s", err)
	}
	r2 := &report.Report{}
	if err := yaml.Unmarshal(b, r2); err != nil {
		t.Fatalf("failed to unmarshal yaml: %s", err)
	}
	if !r1.Equal(r2) {
		diff := helpers.UnifiedDiff(t, r1, r2)
		t.Fatalf("unmarshaled report not equal\n%s", diff)
	}
}
