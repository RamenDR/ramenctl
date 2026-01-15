# HTML Report Generation

This document describes the design for generating HTML reports from command
output.

## Overview

Each command (`validate-application`, `validate-clusters`, `test-run`,
`gather-application`) generates a YAML report. We want to also generate HTML
reports that are human-readable and shareable.

## Requirements

1. Each command has its own report structure with specific data
2. HTML reports share common styling and layout
3. External tools can unmarshal YAML reports without custom parsing
4. Templates should be standard Go templates for maintainability
5. HTML reports are self-contained (embedded CSS)

## Report Types

Each command has its own report type with clear structure:

| Command | Report Type | Specific Data |
|---------|-------------|---------------|
| validate-application | `application.Report` | Application, ApplicationStatus |
| validate-clusters | `clusters.Report` | ClustersStatus |
| test-run | `test.Report` | Config, TestResults |
| gather-application | `gather.Report` | Application, GatherResults |

All reports embed `report.Report` (for validate/gather) or `report.Base` (for
test) which contains common fields.

Report-specific fields use value types (not pointers) since they are always
present:

```go
// validate/application/report.go
type Report struct {
    *report.Report
    Application       report.Application       `json:"application"`
    ApplicationStatus report.ApplicationStatus `json:"applicationStatus"`
}
```

The `Name` field identifies the report type for unmarshaling (similar to
Kubernetes `kind`). We may rename it to `Kind` in the future.

## YAML Unmarshaling

External tools can unmarshal reports by:

1. Using the specific report type directly if known
2. Using a helper function that detects the type from the `Name` field

## HTML Template Structure

### File Organization

Templates use `.tmpl` extension for proper IDE support (Go template syntax
highlighting).

```
pkg/report/
    templates/
        report.tmpl       # Main report structure (shared)
        style.tmpl        # Embedded CSS
    html.go               # Template(), custom functions, HeaderData

pkg/validate/application/
    report.go             # Report type
    templates/
        content.tmpl      # Defines "content" template
    html.go               # templateData, HeaderData, WriteHTML
```

### Main Report Template

The shared `report.tmpl` defines the complete HTML structure. Each command
only needs to define its `content` template:

```html
<!DOCTYPE html>
<html>
<head>
    <title>{{.HeaderData.Title}}</title>
    <style>
        {{ includeCSS "style" . | indent 8 }}
    </style>
</head>
<body>
<header>
    <h1>{{.HeaderData.Title}}</h1>
</header>
<main>
    <section>
        {{ includeHTML "content" . | indent 8 }}
    </section>
    <section>
        <h2>Report Details</h2>
        <p>Common information for all reports.</p>
    </section>
</main>
</body>
</html>
```

### Command-Specific Content

Each command defines only its unique content:

```html
{{define "content" -}}
<h2>Application Status</h2>
<p>Hub, primary cluster, and secondary cluster validation details here.</p>
{{- end}}
```

### Including Templates with Proper Indentation

Go's built-in `{{template}}` doesn't support indentation - included content
appears at column 0. We provide custom functions for WYSIWYG indentation:

- `includeHTML "name" data` - executes a template, returns `template.HTML`
- `includeCSS "name" data` - executes a template, returns `template.CSS`
- `indent N` - adds N spaces after each newline (preserves type)

The indent value matches the visual column position in the template source:

```html
<section>
    {{ includeHTML "content" . | indent 4 }}
</section>
```

### Template Data

Each report uses a `templateData` wrapper that provides `HeaderData()`:

```go
// validate/application/html.go

type templateData struct {
    *Report
}

func (d *templateData) HeaderData() report.HeaderData {
    return report.HeaderData{
        Title:    "Application Validation Report",
        Subtitle: fmt.Sprintf("%s / %s", d.Application.Name, d.Application.Namespace),
    }
}
```

### The shared report template

The report package provides the `Template()` function loading the common
`report.tmpl` template and functions such as `includeHTML` and `indent`.

```go
//go:embed templates/*.tmpl
var templates embed.FS

func Template() (*template.Template, error) {
	var tmpl *template.Template

	funcs := template.FuncMap{
		"includeHTML": func(name string, data any) (template.HTML, error) {
			return includeHTML(tmpl, name, data)
		},
		"includeCSS": func(name string, data any) (template.CSS, error) {
			return includeCSS(tmpl, name, data)
		},
		"indent": indentEscaped,
	}

	var err error
	tmpl, err = template.New("").Funcs(funcs).ParseFS(templates, "templates/*.tmpl")

	return tmpl, err
}
```

### Command specific template

Command package provide `Template()` function loading the report template and
adding the command specific templates.

```go
//go:embed templates/*.tmpl
var templates embed.FS

func Template() (*template.Template, error) {
	tmpl, err := report.Template()
	if err != nil {
		return nil, err
	}
	return tmpl.ParseFS(templates, "templates/*.tmpl")
}
```

### Command HTML writing

Command report provides the `WriteHTML()` function using the command template
to write report HTML.

```go
func (r *Report) WriteHTML(w io.Writer) error {
	tmpl, err := Template()
	if err != nil {
		return err
	}
	return tmpl.ExecuteTemplate(w, "report.tmpl", &templateData{r})
}
```

## Adding a New Report

To add HTML support for a new command (e.g., `gather-application`):

1. Create `templates/content.tmpl` defining the `content` template
2. Create `html.go` with `templateData` type and `Template()` function
3. Add `WriteHTML()` method to the report type
