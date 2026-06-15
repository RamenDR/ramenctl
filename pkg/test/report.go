// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"

	e2econfig "github.com/ramendr/ramen/e2e/config"

	"github.com/ramendr/ramenctl/pkg/report"
)

const (
	ValidateStep = "validate"
	SetupStep    = "setup"
	TestsStep    = "tests"
	CleanupStep  = "cleanup"
)

// Summary keys for test reports.
const (
	Passed   = report.SummaryKey("passed")
	Failed   = report.SummaryKey("failed")
	Skipped  = report.SummaryKey("skipped")
	Canceled = report.SummaryKey("canceled")
)

// Report created by test sub commands.
type Report struct {
	*report.Base
	Config *e2econfig.Config `json:"config"`
}

func newReport(commandName string, config *e2econfig.Config) *Report {
	if config == nil {
		panic("config must not be nil")
	}
	base := report.NewBase(commandName)
	base.Summary = &report.Summary{}
	return &Report{
		Base:   base,
		Config: config,
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
	return true
}

func addTest(s *report.Summary, t *report.Step) {
	switch t.Status {
	case report.Passed:
		s.Add(Passed)
	case report.Failed:
		s.Add(Failed)
	case report.Skipped:
		s.Add(Skipped)
	case report.Canceled:
		s.Add(Canceled)
	}
}

func summaryString(s *report.Summary) string {
	return fmt.Sprintf("%d passed, %d failed, %d skipped, %d canceled",
		s.Get(Passed), s.Get(Failed),
		s.Get(Skipped), s.Get(Canceled))
}
