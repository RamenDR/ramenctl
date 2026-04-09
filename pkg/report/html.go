// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"embed"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/yosssi/gohtml"
	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/time"
)

//go:embed templates/*.tmpl
var templates embed.FS

//go:embed style.css
var styleCSS []byte

// WriteCSS writes the shared stylesheet to the output directory.
func WriteCSS(dir string) error {
	path := filepath.Join(dir, "style.css")
	return os.WriteFile(path, styleCSS, 0o640)
}

// Template returns a new template set with shared definitions.
func Template() (*template.Template, error) {
	funcs := template.FuncMap{
		"formatTime":     formatTime,
		"formatDuration": formatDuration,
		"formatYAML":     formatYAML,
		"icon":           icon,
		"isProblem":      isProblem,
		"truncate":       truncate,
		"isTruncated":    isTruncated,
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

// truncate shortens a value to n characters, appending ".." if truncated.
func truncate(v any, n int) string {
	s := fmt.Sprint(v)
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// isTruncated returns true if the value would be truncated by truncate().
func isTruncated(v any, n int) bool {
	return len(fmt.Sprint(v)) > n
}

// formatTime formats a time value for display in reports.
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05 -0700")
}

// formatDuration formats a duration in seconds for display.
func formatDuration(seconds float64) string {
	if seconds < 60 {
		return fmt.Sprintf("%.2fs", seconds)
	}
	m := int(seconds) / 60
	s := seconds - float64(m*60)
	return fmt.Sprintf("%dm %.2fs", m, s)
}

// formatYAML marshals a value to YAML for display in a <pre> block.
func formatYAML(v any) (string, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
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
