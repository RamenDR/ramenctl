// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"embed"
	"fmt"
	"html/template"
	"strings"

	"github.com/ramendr/ramenctl/pkg/time"
)

//go:embed templates/*.tmpl
var templates embed.FS

// Template returns a new template set with shared definitions.
func Template() (*template.Template, error) {
	var tmpl *template.Template

	funcs := template.FuncMap{
		"formatTime": formatTime,
		"includeHTML": func(name string, data any) (template.HTML, error) {
			return includeHTML(tmpl, name, data)
		},
		"includeCSS": func(name string, data any) (template.CSS, error) {
			return includeCSS(tmpl, name, data)
		},
		"indent": indentEscaped,
	}

	// Must assign to tmpl since the include closures capture it.
	var err error
	tmpl, err = template.New("").Funcs(funcs).ParseFS(templates, "templates/*.tmpl")

	return tmpl, err
}

// formatTime formats a time value for display in reports.
func formatTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05 MST")
}

// Template functions for including nested templates with proper indentation.
//
// The includeHTML and includeCSS functions execute a named template and return
// the result as a typed safe value. The indent function adds leading spaces to
// all lines except the first, allowing WYSIWYG template indentation:
//
//	<div>
//	    {{ includeHTML "section" . | indent 4 }}
//	</div>
//
// The template writer controls the first line's indentation; indent handles the
// rest. These functions must be used in the correct context:
//   - includeHTML: only in HTML context (element content, attribute values)
//   - includeCSS: only in CSS context (inside <style> tags)
//
// # Safety
//
// The functions return typed values (template.HTML, template.CSS) that
// html/template recognizes as safe for their respective contexts:
//
//   - includeHTML in HTML context: Safe. The included template is executed by
//     html/template which auto-escapes interpolated values during execution.
//     The template.HTML wrapper prevents double-escaping by the outer template.
//
//   - includeCSS in CSS context: Safe. The included template is executed by
//     html/template which applies CSS-appropriate escaping during execution.
//     The template.CSS wrapper prevents double-escaping by the outer template.
//
//   - Wrong context (e.g., includeHTML in <style>): Blocked. html/template's
//     context-aware escaping detects the type mismatch and produces "ZgotmplZ"
//     as a safe placeholder, preventing potential injection.

// includeHTML executes a template and returns safe HTML. See safety discussion above.
func includeHTML(tmpl *template.Template, name string, data any) (template.HTML, error) {
	var buf strings.Builder
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return template.HTML(buf.String()), nil // #nosec G203
}

// includeCSS executes a template and returns safe CSS. See safety discussion above.
func includeCSS(tmpl *template.Template, name string, data any) (template.CSS, error) {
	var buf strings.Builder
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	return template.CSS(buf.String()), nil // #nosec G203
}

// indentEscaped adds leading spaces to all lines except the first. It preserves
// the input type (template.HTML or template.CSS) without escaping. Must be
// called with output from includeHTML or includeCSS.
func indentEscaped(spaces int, s any) (any, error) {
	indent := strings.Repeat(" ", spaces)
	switch v := s.(type) {
	case template.HTML:
		return template.HTML(strings.ReplaceAll(string(v), "\n", "\n"+indent)), nil // #nosec G203
	case template.CSS:
		return template.CSS(strings.ReplaceAll(string(v), "\n", "\n"+indent)), nil // #nosec G203
	default:
		return nil, fmt.Errorf("indent: unsupported type %T", s)
	}
}

// HeaderData provides data for the report template.
type HeaderData struct {
	Title    string // Report title, e.g. "Application Validation Report"
	Subtitle string // Additional context, e.g. "myapp / mynamespace"
}
