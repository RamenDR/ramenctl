// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/report"
)

// Report is the report for the validate-application command.
type Report struct {
	*report.Report
	Application       report.Application       `json:"application"`
	ApplicationStatus report.ApplicationStatus `json:"applicationStatus"`
}

// NewReport creates a new application validation report.
func NewReport(cfg *config.Config) *Report {
	r := report.NewReport("validate-application", cfg)
	r.Summary = &report.Summary{}
	return &Report{
		Report: r,
	}
}
