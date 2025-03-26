// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"

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

// A step is a test command step. The setup and cleanup steps are modeled as step without tests.
type Step struct {
	Name   string   `json:"name"`
	Status Status   `json:"status,omitempty"`
	Items  []Result `json:"items,omitempty"`
}

// Summary summaries a test run or clean.
type Summary struct {
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// Result describes a single test result.
type Result struct {
	Workload string `json:"workload"`
	Deployer string `json:"deployer"`
	PVCSpec  string `json:"pvcSpec"`
	Status   Status `json:"status,omitempty"`
}

// Report created by test sub commands.
type Report struct {
	*report.Report
	Name    string  `json:"name"`
	Steps   []*Step `json:"steps"`
	Summary Summary `json:"summary"`
	Status  Status  `json:"status,omitempty"`
}

func newReport(commandName string) *Report {
	return &Report{
		Report: report.New(),
		Name:   commandName,
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
}

// AddTest records a completed test. A failed test mark the test step and the report as failed.
func (r *Report) AddTest(t *Test) {
	var step *Step

	// To make it easy to use, we create the tests step automaticlaly when adding the first test.
	if len(r.Steps) == 0 || r.Steps[len(r.Steps)-1].Name != TestsStep {
		step = &Step{Name: TestsStep}
		r.AddStep(step)
	} else {
		step = r.Steps[len(r.Steps)-1]
	}

	step.AddTest(t)

	switch t.Status {
	case Passed:
		r.Summary.Passed++
		if r.Status == "" {
			r.Status = Passed
		}
	case Failed:
		r.Summary.Failed++
		r.Status = Failed
	case Skipped:
		r.Summary.Skipped++
		if r.Status == "" {
			r.Status = Passed
		}
	}
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
	result := Result{
		Workload: t.Config.Workload,
		Deployer: t.Config.Deployer,
		PVCSpec:  t.Config.PVCSpec,
		Status:   t.Status,
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
