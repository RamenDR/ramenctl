// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"fmt"
	"maps"

	"github.com/ramendr/ramenctl/pkg/report"
)

// Report created by validate sub commands.
type Report struct {
	*report.Report
	Summary report.Summary `json:"summary"`
}

func (r *Report) Equal(o *Report) bool {
	if r == o {
		return true
	}
	if o == nil {
		return false
	}
	if !r.Report.Equal(o.Report) {
		return false
	}
	if !maps.Equal(r.Summary, o.Summary) {
		return false
	}
	return true
}

// addValidation adds a validation to the summary.
func addValidation(s report.Summary, v report.Validation) {
	switch v.GetState() {
	case report.OK:
		s.Add(report.ValidationOK)
	case report.Stale:
		s.Add(report.ValidationStale)
	case report.Problem:
		s.Add(report.ValidationProblem)
	}
}

// hasIssues returns true if there are any problems or stale results.
func hasIssues(s report.Summary) bool {
	return s.Get(report.ValidationStale) > 0 || s.Get(report.ValidationProblem) > 0
}

// summaryString returns a string representation.
func summaryString(s report.Summary) string {
	return fmt.Sprintf("%d ok, %d stale, %d problem",
		s.Get(report.ValidationOK), s.Get(report.ValidationStale), s.Get(report.ValidationProblem))
}
