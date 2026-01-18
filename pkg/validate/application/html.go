// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"embed"
	"html/template"
	"io"

	"github.com/ramendr/ramenctl/pkg/report"
)

//go:embed templates/*.tmpl
var templates embed.FS

// templateData wraps the report with template helper methods.
type templateData struct {
	*Report
}

// HeaderData returns data for the report template.
func (d *templateData) HeaderData() report.HeaderData {
	return report.HeaderData{
		Title:    "Application Validation Report",
		Subtitle: d.Application.Namespace + " / " + d.Application.Name,
	}
}

// Template returns the HTML template for this report.
func Template() (*template.Template, error) {
	tmpl, err := report.Template()
	if err != nil {
		return nil, err
	}
	return tmpl.ParseFS(templates, "templates/*.tmpl")
}

// WriteHTML writes the HTML report to the writer.
func (r *Report) WriteHTML(w io.Writer) error {
	tmpl, err := Template()
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, "report.tmpl", &templateData{r})
}
