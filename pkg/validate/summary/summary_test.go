// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package summary

import (
	"testing"

	"github.com/ramendr/ramenctl/pkg/report"
)

func TestSummaryAdd(t *testing.T) {
	s := &report.Summary{}

	AddValidation(s, &report.Validated{State: report.OK})
	AddValidation(s, &report.Validated{State: report.OK})
	AddValidation(s, &report.Validated{State: report.Stale})
	AddValidation(s, &report.Validated{State: report.OK})
	AddValidation(s, &report.Validated{State: report.Stale})
	AddValidation(s, &report.Validated{State: report.Problem})

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
		if got := HasIssues(tc.summary); got != tc.expected {
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
	if String(s) != expected {
		t.Fatalf("expected %q, got %q", expected, String(s))
	}
}
