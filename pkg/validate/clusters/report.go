// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package clusters

import (
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/report"
)

// Report is the report for the validate-clusters command.
type Report struct {
	*report.Report
	ClustersStatus report.ClustersStatus `json:"clustersStatus"`
}

// NewReport creates a new clusters validation report.
func NewReport(cfg *config.Config) *Report {
	r := report.NewReport("validate-clusters", cfg)
	r.Summary = &report.Summary{}
	return &Report{
		Report: r,
	}
}
