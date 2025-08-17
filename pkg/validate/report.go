// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"fmt"

	"github.com/ramendr/ramenctl/pkg/report"
)

type Summary struct {
	Problem uint `json:"error"`
	Stale   uint `json:"stale"`
	OK      uint `json:"ok"`
}

// Report created by validate sub commands.
type Report struct {
	*report.Report
	Summary Summary `json:"summary"`
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
	if r.Summary != o.Summary {
		return false
	}
	return true
}

// Add a validation to the summary.
func (s *Summary) Add(v report.Validation) {
	switch v.GetState() {
	case report.OK:
		s.OK++
	case report.Stale:
		s.Stale++
	case report.Problem:
		s.Problem++
	}
}

// HasIssues returns true if there are any problems or stale results.
func (s *Summary) HasIssues() bool {
	return s.Stale > 0 || s.Problem > 0
}

// String returns a string representation.
func (s Summary) String() string {
	return fmt.Sprintf("%d ok, %d stale, %d problem", s.OK, s.Stale, s.Problem)
}
