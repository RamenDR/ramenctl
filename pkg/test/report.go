// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"

	e2econfig "github.com/ramendr/ramen/e2e/config"

	"github.com/ramendr/ramenctl/pkg/report"
)

type Status string

const (
	ValidateStep = "validate"
	SetupStep    = "setup"
	TestsStep    = "tests"
	CleanupStep  = "cleanup"
)

// Summary summaries a test run or clean.
type Summary struct {
	Passed   int `json:"passed"`
	Failed   int `json:"failed"`
	Skipped  int `json:"skipped"`
	Canceled int `json:"canceled"`
}

// Report created by test sub commands.
type Report struct {
	*report.Report
	Config  *e2econfig.Config `json:"config"`
	Summary Summary           `json:"summary"`
}

func newReport(commandName string, config *e2econfig.Config) *Report {
	if config == nil {
		panic("config must not be nil")
	}
	return &Report{
		Report: report.New(commandName),
		Config: config,
	}
}

// AddStep adds a step to the report.
func (r *Report) AddStep(step *report.Step) {
	r.Report.AddStep(step)

	// Handle the special "tests" step.
	if step.Name == TestsStep {
		for _, t := range step.Items {
			r.Summary.AddTest(t)
		}
	}
}

// Equal return true if report is equal to other report.
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
	if r.Config != nil && o.Config != nil {
		if !r.Config.Equal(o.Config) {
			return false
		}
	} else if r.Config != o.Config {
		return false
	}
	if r.Summary != o.Summary {
		return false
	}
	return true
}

func (s *Summary) AddTest(t *report.Step) {
	switch t.Status {
	case report.Passed:
		s.Passed++
	case report.Failed:
		s.Failed++
	case report.Skipped:
		s.Skipped++
	case report.Canceled:
		s.Canceled++
	}
}

func (s Summary) String() string {
	return fmt.Sprintf("%d passed, %d failed, %d skipped, %d canceled",
		s.Passed, s.Failed, s.Skipped, s.Canceled)
}
