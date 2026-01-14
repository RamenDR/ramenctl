// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"maps"

	e2econfig "github.com/ramendr/ramen/e2e/config"

	"github.com/ramendr/ramenctl/pkg/report"
)

const (
	ValidateStep = "validate"
	SetupStep    = "setup"
	TestsStep    = "tests"
	CleanupStep  = "cleanup"
)

// Report created by test sub commands.
type Report struct {
	*report.Base
	Config  *e2econfig.Config `json:"config"`
	Summary report.Summary    `json:"summary"`
}

func newReport(commandName string, config *e2econfig.Config) *Report {
	if config == nil {
		panic("config must not be nil")
	}
	return &Report{
		Base:    report.NewBase(commandName),
		Config:  config,
		Summary: report.Summary{},
	}
}

// AddStep adds a step to the report.
func (r *Report) AddStep(step *report.Step) {
	r.Base.AddStep(step)

	// Handle the special "tests" step.
	if step.Name == TestsStep {
		for _, t := range step.Items {
			addTest(r.Summary, t)
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
	if !r.Base.Equal(o.Base) {
		return false
	}
	if r.Config != nil && o.Config != nil {
		if !r.Config.Equal(o.Config) {
			return false
		}
	} else if r.Config != o.Config {
		return false
	}
	if !maps.Equal(r.Summary, o.Summary) {
		return false
	}
	return true
}

func addTest(s report.Summary, t *report.Step) {
	switch t.Status {
	case report.Passed:
		s.Add(report.TestPassed)
	case report.Failed:
		s.Add(report.TestFailed)
	case report.Skipped:
		s.Add(report.TestSkipped)
	case report.Canceled:
		s.Add(report.TestCanceled)
	}
}

func summaryString(s report.Summary) string {
	return fmt.Sprintf("%d passed, %d failed, %d skipped, %d canceled",
		s.Get(report.TestPassed), s.Get(report.TestFailed),
		s.Get(report.TestSkipped), s.Get(report.TestCanceled))
}
