// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"runtime"

	"github.com/ramendr/ramenctl/pkg/build"
)

// Host describes the host ramenctl is running on.
type Host struct {
	OS   string `json:"os"`
	Arch string `json:"arch"`
	Cpus int    `json:"cpus"`
}

// Ramenctl describes ramenctl.
type Ramenctl struct {
	Version string `json:"version,omitempty"`
	Commit  string `json:"commit,omitempty"`
}

// Report created by ramenctl command.
type Report struct {
	Host     Host      `json:"host"`
	Ramenctl *Ramenctl `json:"ramenctl,omitempty"`
}

// New create a new generic report. Commands embed the report in the command report.
func New() *Report {
	r := &Report{
		Host: Host{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
			Cpus: runtime.NumCPU(),
		},
	}
	if build.Version != "" || build.Commit != "" {
		r.Ramenctl = &Ramenctl{
			Version: build.Version,
			Commit:  build.Commit,
		}
	}
	return r
}
