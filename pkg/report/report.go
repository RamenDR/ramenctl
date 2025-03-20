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
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// Report created by ramenctl command.
type Report struct {
	Host     Host     `json:"host"`
	Ramenctl Ramenctl `json:"ramenctl"`
}

// New create a new generic report. Commands embed the report in the command report.
func New() *Report {
	return &Report{
		Host: Host{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
			Cpus: runtime.NumCPU(),
		},
		Ramenctl: Ramenctl{
			Version: build.Version,
			Commit:  build.Commit,
		},
	}
}
