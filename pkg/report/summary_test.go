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
		report.TestPassed:   5,
		report.TestFailed:   2,
		report.TestSkipped:  1,
		report.TestCanceled: 0,
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
		report.ValidationOK:      10,
		report.ValidationStale:   1,
		report.ValidationProblem: 2,
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
