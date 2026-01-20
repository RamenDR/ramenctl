// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"embed"
	"html/template"

	"github.com/yosssi/gohtml"

	"github.com/ramendr/ramenctl/pkg/time"
)

//go:embed templates/*.tmpl
var templates embed.FS

// Template returns a new template set with shared definitions.
func Template() (*template.Template, error) {
	funcs := template.FuncMap{
		"formatTime": formatTime,
		"icon":       icon,
		"isProblem":  isProblem,
	}
	return template.New("").Funcs(funcs).ParseFS(templates, "templates/*.tmpl")
}

// isProblem returns true if the validation state is Problem.
func isProblem(s ValidationState) bool {
	return s == Problem
}

// icon returns the icon for a validation state.
func icon(s ValidationState) string {
	switch s {
	case OK:
		return "✅"
	case Stale:
		return "⭕"
	case Problem:
		return "❌"
	default:
		return ""
	}
}

// formatTime formats a time value for display in reports.
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05 MST")
}

// FormatHTML formats HTML for readability. Use this to format generated HTML
// for comparison with golden files in tests.
func FormatHTML(html string) string {
	gohtml.Condense = true
	return gohtml.Format(html) + "\n"
}

// HeaderData provides data for the report template.
type HeaderData struct {
	Title    string // Report title, e.g. "Application Validation Report"
	Subtitle string // Additional context, e.g. "myapp / mynamespace"
}
