// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"os"
	"strings"
	"testing"

	"github.com/ramendr/ramenctl/pkg/helpers"
)

func TestTemplate(t *testing.T) {
	tmpl, err := Template()
	if err != nil {
		t.Fatalf("Template() error: %v", err)
	}

	// Check that shared templates are defined
	for _, name := range []string{"report.tmpl", "style"} {
		if tmpl.Lookup(name) == nil {
			t.Errorf("template %q not defined", name)
		}
	}
}

func TestIncludeIndent(t *testing.T) {
	tmpl, err := Template()
	if err != nil {
		t.Fatalf("Template() error: %v", err)
	}

	// Define trivial templates just for testing includeHTML/indent.
	// Use spaces (not tabs) since indent adds spaces.
	// The template writer indents the first line; indent value must match.
	// Nesting compounds: outer adds 4, inner adds 8, etc.
	tmpl, err = tmpl.Parse(`
{{define "item" -}}
<li>Item</li>
{{- end}}

{{define "list" -}}
<ul>
    {{ includeHTML "item" . | indent 4 }}
</ul>
{{- end}}

{{define "outer" -}}
<div>
    <section>
        {{ includeHTML "list" . | indent 8 }}
    </section>
</div>
{{end}}
`)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	var buf strings.Builder
	err = tmpl.ExecuteTemplate(&buf, "outer", nil)
	if err != nil {
		t.Fatalf("ExecuteTemplate error: %v", err)
	}

	got := buf.String()
	expected, err := os.ReadFile("testdata/include-indent.html")
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if got != string(expected) {
		t.Errorf("output mismatch:\n%s", helpers.UnifiedDiff(t, string(expected), got))
	}
}
