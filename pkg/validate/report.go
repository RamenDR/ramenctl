// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"fmt"

	"github.com/ramendr/ramenctl/pkg/report"
)

// Summary keys for validation reports.
const (
	OK      = report.SummaryKey("ok")
	Stale   = report.SummaryKey("stale")
	Problem = report.SummaryKey("problem")
)

// addValidation adds a validation to the summary.
func addValidation(s *report.Summary, v report.Validation) {
	switch v.GetState() {
	case report.OK:
		s.Add(OK)
	case report.Stale:
		s.Add(Stale)
	case report.Problem:
		s.Add(Problem)
	}
}

// hasIssues returns true if there are any problems or stale results.
func hasIssues(s *report.Summary) bool {
	return s.Get(Stale) > 0 || s.Get(Problem) > 0
}

// summaryString returns a string representation.
func summaryString(s *report.Summary) string {
	return fmt.Sprintf("%d ok, %d stale, %d problem",
		s.Get(OK), s.Get(Stale), s.Get(Problem))
}
