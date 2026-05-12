// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package summary

import (
	"fmt"

	"github.com/ramendr/ramenctl/pkg/report"
)

// Summary keys for validation reports.
const (
	OK      = report.SummaryKey("ok")
	Warning = report.SummaryKey("warning")
	Problem = report.SummaryKey("problem")
)

// AddValidation adds a validation to the summary.
func AddValidation(s *report.Summary, v report.Validation) {
	switch v.GetState() {
	case report.OK:
		s.Add(OK)
	case report.Warning:
		s.Add(Warning)
	case report.Problem:
		s.Add(Problem)
	}
}

// HasIssues returns true if there are any problems or warning results.
func HasIssues(s *report.Summary) bool {
	return s.Get(Warning) > 0 || s.Get(Problem) > 0
}

// String returns a string representation of a validation summary.
func String(s *report.Summary) string {
	return fmt.Sprintf("%d ok, %d warning, %d problem",
		s.Get(OK), s.Get(Warning), s.Get(Problem))
}
