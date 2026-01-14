// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"fmt"
	"runtime"
	"slices"

	"github.com/ramendr/ramenctl/pkg/build"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/time"
)

type Status string

const (
	Passed   = Status("passed")
	Failed   = Status("failed")
	Skipped  = Status("skipped")
	Canceled = Status("canceled")
)

// Host describes the host ramenctl is running on.
type Host struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
	Cpus int    `json:"cpus"`
}

// Build describes ramenctl build.
type Build struct {
	Version string `json:"version,omitempty"`
	Commit  string `json:"commit,omitempty"`
}

type Step struct {
	Name     string  `json:"name"`
	Status   Status  `json:"status,omitempty"`
	Duration float64 `json:"duration,omitempty"`
	Items    []*Step `json:"items,omitempty"`
}

// Base report for ramenctl commands report.
type Base struct {
	Host     Host      `json:"host"`
	Build    *Build    `json:"build,omitempty"`
	Created  time.Time `json:"created"`
	Name     string    `json:"name"`
	Status   Status    `json:"status,omitempty"`
	Duration float64   `json:"duration,omitempty"`
	Steps    []*Step   `json:"steps"`

	// Summary is set by `validate` and `test` commands.
	Summary *Summary `json:"summary,omitempty"`
}

// Application is application info.
type Application struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// Report is used by all ramenctl commands except the test commands.
type Report struct {
	*Base
	Config *config.Config `json:"config"`

	// Namespaces is set by `validate` and `gather` commands.
	Namespaces []string `json:"namespaces,omitempty"`

	// Application is set by `validate application` and `gather application` commands.
	Application *Application `json:"application,omitempty"`

	// ApplicationStatus is set by the `validate application` commmnad.
	ApplicationStatus *ApplicationStatus `json:"applicationStatus,omitempty"`

	// ClustersStatus is set by the `validate clusters` command.
	ClustersStatus *ClustersStatus `json:"clustersStatus,omitempty"`
}

// NewBase create a new base report for ramenctl commands reports.
func NewBase(commandName string) *Base {
	r := &Base{
		Name: commandName,
		Host: Host{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
			Cpus: runtime.NumCPU(),
		},
		// time value without monotonic clock reading
		Created: time.Now().Local(),
	}
	if build.Version != "" || build.Commit != "" {
		r.Build = &Build{
			Version: build.Version,
			Commit:  build.Commit,
		}
	}
	return r
}

// Equal returns true if report is equal to other report.
func (r *Base) Equal(o *Base) bool {
	if r == o {
		return true
	}
	if o == nil {
		return false
	}
	if r.Host != o.Host {
		return false
	}
	if !r.Created.Equal(o.Created) {
		return false
	}
	if r.Build != nil && o.Build != nil {
		if *r.Build != *o.Build {
			return false
		}
	} else if r.Build != o.Build {
		return false
	}
	if r.Name != o.Name {
		return false
	}
	if r.Status != o.Status {
		return false
	}
	if r.Duration != o.Duration {
		return false
	}
	if !slices.EqualFunc(r.Steps, o.Steps, func(a *Step, b *Step) bool {
		return a.Equal(b)
	}) {
		return false
	}
	if !r.Summary.Equal(o.Summary) {
		return false
	}
	return true
}

// AddStep adds a step to the report.
func (r *Base) AddStep(step *Step) {
	if findStep(r.Steps, step.Name) != nil {
		panic(fmt.Sprintf("step %q exists", step.Name))
	}
	r.Steps = append(r.Steps, step)
	r.Duration += step.Duration

	switch step.Status {
	case Passed, Skipped:
		if r.Status == "" {
			r.Status = Passed
		}
	case Failed:
		if r.Status != Canceled {
			r.Status = step.Status
		}
	case Canceled:
		r.Status = step.Status
	}
}

func NewReport(commandName string, cfg *config.Config) *Report {
	if cfg == nil {
		panic("cfg must not be nil")
	}
	return &Report{
		Base:   NewBase(commandName),
		Config: cfg,
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
	if r.Application != nil && o.Application != nil {
		if *r.Application != *o.Application {
			return false
		}
	} else if r.Application != o.Application {
		return false
	}
	if !slices.Equal(r.Namespaces, o.Namespaces) {
		return false
	}
	return true
}

// AddStep records a completed sub step.
func (s *Step) AddStep(sub *Step) {
	if findStep(s.Items, sub.Name) != nil {
		panic(fmt.Sprintf("step %q exists", sub.Name))
	}
	s.Items = append(s.Items, sub)

	switch sub.Status {
	case Passed, Skipped:
		if s.Status == "" {
			s.Status = Passed
		}
	case Failed:
		if s.Status != Canceled {
			s.Status = sub.Status
		}
	case Canceled:
		s.Status = sub.Status
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
	if s.Duration != o.Duration {
		return false
	}
	return slices.EqualFunc(s.Items, o.Items, func(a *Step, b *Step) bool {
		if a == nil {
			return b == nil
		}
		return a.Equal(b)
	})
}

func findStep(steps []*Step, name string) *Step {
	for _, step := range steps {
		if step.Name == name {
			return step
		}
	}
	return nil
}
