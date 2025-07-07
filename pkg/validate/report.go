// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/report"
)

// Report created by validate sub commands.
type Report struct {
	*report.Base
	Config *config.Config `json:"config"`
}

func newReport(commandName string, cfg *config.Config) *Report {
	if cfg == nil {
		panic("cfg must not be nil")
	}
	return &Report{
		Base:   report.NewBase(commandName),
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
	return true
}
