// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report_test

import (
	"maps"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/report"
)

func TestTestSummaryYAMLRoundtrip(t *testing.T) {
	s1 := report.Summary{
		report.SummaryKey("passed"):   5,
		report.SummaryKey("failed"):   2,
		report.SummaryKey("skipped"):  1,
		report.SummaryKey("canceled"): 0,
	}

	data, err := yaml.Marshal(s1)
	if err != nil {
		t.Fatalf("failed to marshal summary: %v", err)
	}

	var s2 report.Summary
	if err := yaml.Unmarshal(data, &s2); err != nil {
		t.Fatalf("failed to unmarshal summary: %v", err)
	}

	if !maps.Equal(s1, s2) {
		t.Errorf("expected %v, got %v", s1, s2)
	}
}

func TestValidationSummaryYAMLRoundtrip(t *testing.T) {
	s1 := report.Summary{
		report.SummaryKey("ok"):      10,
		report.SummaryKey("stale"):   1,
		report.SummaryKey("problem"): 2,
	}

	data, err := yaml.Marshal(s1)
	if err != nil {
		t.Fatalf("failed to marshal summary: %v", err)
	}

	var s2 report.Summary
	if err := yaml.Unmarshal(data, &s2); err != nil {
		t.Fatalf("failed to unmarshal summary: %v", err)
	}

	if !maps.Equal(s1, s2) {
		t.Errorf("expected %v, got %v", s1, s2)
	}
}
