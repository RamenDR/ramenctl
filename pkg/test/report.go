// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"slices"

	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/report"
)

type Status string

const (
	Passed  = Status("passed")
	Failed  = Status("failed")
	Skipped = Status("skipped")
)

const (
	ValidateStep = "validate"
	SetupStep    = "setup"
	TestsStep    = "tests"
	CleanupStep  = "cleanup"
)

// A step is a test command step.
type Step struct {
	Name   string            `json:"name"`
	Status Status            `json:"status,omitempty"`
	Config *types.TestConfig `json:"config,omitempty"`
	Items  []*Step           `json:"items,omitempty"`
}

// Summary summaries a test run or clean.
type Summary struct {
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// Report created by test sub commands.
type Report struct {
	*report.Report
	Name    string        `json:"name"`
	Config  *types.Config `json:"config"`
	Steps   []*Step       `json:"steps"`
	Summary Summary       `json:"summary"`
	Status  Status        `json:"status,omitempty"`
}

func newReport(commandName string, config *types.Config) *Report {
	if config == nil {
		panic("config must not be nil")
	}
	return &Report{
		Report: report.New(),
		Name:   commandName,
		Config: config,
	}
}

// AddStep adds a step to the report.
func (r *Report) AddStep(step *Step) {
	if r.findStep(step.Name) != nil {
		panic(fmt.Sprintf("step %q exists", step.Name))
	}
	r.Steps = append(r.Steps, step)

	switch step.Status {
	case Passed, Skipped:
		if r.Status == "" {
			r.Status = Passed
		}
	case Failed:
		r.Status = Failed
	}

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
	if r.Name != o.Name {
		return false
	}
	if r.Config != o.Config {
		if r.Config == nil || o.Config == nil {
			return false
		}
		if !r.Config.Equal(o.Config) {
			return false
		}
	}
	if r.Status != o.Status {
		return false
	}
	if r.Summary != o.Summary {
		return false
	}
	return slices.EqualFunc(r.Steps, o.Steps, func(a *Step, b *Step) bool {
		return a.Equal(b)
	})
}

func (r *Report) findStep(name string) *Step {
	for _, step := range r.Steps {
		if step.Name == name {
			return step
		}
	}
	return nil
}

// AddTest records a completed test. A failed test marks the step as failed.
func (s *Step) AddTest(t *Test) {
	result := &Step{
		Name:   t.Name(),
		Config: t.Config,
		Status: t.Status,
		Items:  t.Steps,
	}

	s.Items = append(s.Items, result)

	switch t.Status {
	case Passed, Skipped:
		if s.Status == "" {
			s.Status = Passed
		}
	case Failed:
		s.Status = Failed
	}
}

// Equal return true if step is equal to other step.
func (s *Step) Equal(o *Step) bool {
	if s == o {
		return true
	}
	if o == nil {
		return false
	}
	if s.Name != o.Name {
		return false
	}
	if s.Status != o.Status {
		return false
	}
	if s.Config != o.Config {
		if s.Config == nil || o.Config == nil {
			return false
		}
		if *s.Config != *o.Config {
			return false
		}
	}
	return slices.EqualFunc(s.Items, o.Items, func(a *Step, b *Step) bool {
		if a == nil {
			return b == nil
		}
		return a.Equal(b)
	})
}

func (s *Summary) AddTest(t *Step) {
	switch t.Status {
	case Passed:
		s.Passed++
	case Failed:
		s.Failed++
	case Skipped:
		s.Skipped++
	}
}
