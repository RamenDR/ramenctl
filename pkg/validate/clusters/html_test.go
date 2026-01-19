// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package clusters

import (
	"os"
	"strings"
	"testing"

	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
)

func TestTemplate(t *testing.T) {
	tmpl, err := Template()
	if err != nil {
		t.Fatalf("Template() error: %v", err)
	}

	// Check that shared templates and command templates are defined
	for _, name := range []string{"report.tmpl", "style", "content"} {
		if tmpl.Lookup(name) == nil {
			t.Errorf("template %q not defined", name)
		}
	}
}

func TestWriteHTML(t *testing.T) {
	r := &Report{
		Report: &report.Report{
			Base: &report.Base{
				Name:   "validate-clusters",
				Status: report.Passed,
			},
		},
	}

	var buf strings.Builder
	err := r.WriteHTML(&buf)
	if err != nil {
		t.Fatalf("WriteHTML() error: %v", err)
	}

	actual := buf.String()

	expected, err := os.ReadFile("testdata/report.html")
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}

	if actual != string(expected) {
		t.Fatalf("output mismatch.\n%s", helpers.UnifiedDiff(t, string(expected), actual))
	}
}

func TestHeaderData(t *testing.T) {
	r := &Report{
		Report: &report.Report{
			Base: &report.Base{
				Name: "validate-clusters",
			},
		},
	}

	d := &templateData{r}
	actual := d.HeaderData()

	expected := report.HeaderData{
		Title: "Clusters Validation Report",
	}

	if actual != expected {
		t.Fatalf("mismatch.\n%s", helpers.UnifiedDiff(t, expected, actual))
	}
}
