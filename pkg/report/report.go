// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"runtime"
	"time"

	"github.com/ramendr/ramenctl/pkg/build"
)

// Used to provide a custom clock for testing.
var Now func() time.Time

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

// Report created by ramenctl command.
type Report struct {
	Host    Host      `json:"host"`
	Build   *Build    `json:"build,omitempty"`
	Created time.Time `json:"created"`
}

// New create a new generic report. Commands embed the report in the command report.
func New() *Report {
	r := &Report{
		Host: Host{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
			Cpus: runtime.NumCPU(),
		},
		Created: Now(),
	}
	if build.Version != "" || build.Commit != "" {
		r.Build = &Build{
			Version: build.Version,
			Commit:  build.Commit,
		}
	}
	return r
}

// marshalableTime return a time value without monotonic time info. This makes it possible to marshal and unmarshal the
// time value.
func marshalableTime() time.Time {
	return time.Now().Local()
}

func init() {
	Now = marshalableTime
}
