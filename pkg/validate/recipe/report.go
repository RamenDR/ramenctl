// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package recipe

import (
	"github.com/ramendr/ramenctl/pkg/report"
)

// Report is the report for the validate-recipe command.
type Report struct {
	*report.Base
	RecipeReport report.RecipeReport `json:"recipeReport"`
}

// NewReport creates a new recipe validation report.
func NewReport() *Report {
	base := report.NewBase(CommandName)
	base.Summary = &report.Summary{}
	return &Report{
		Base: base,
	}
}
