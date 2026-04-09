// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"os"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
)

func TestTemplate(t *testing.T) {
	tmpl, err := Template()
	if err != nil {
		t.Fatalf("Template() error: %v", err)
	}

	expected := []string{
		// Shared templates from pkg/report.
		"conditions",
		"report.tmpl",
		"validated",
		// Command templates.
		"content",
		"drpc",
		"pvc",
		"s3",
		"vrg",
	}
	for _, name := range expected {
		if tmpl.Lookup(name) == nil {
			t.Errorf("template %q not defined", name)
		}
	}
}

func TestWriteHTML(t *testing.T) {
	for _, name := range []string{"ok", "problem"} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile("testdata/" + name + ".yaml")
			if err != nil {
				t.Fatalf("ReadFile() error: %v", err)
			}

			r := &Report{}
			if err := yaml.Unmarshal(data, r); err != nil {
				t.Fatalf("Unmarshal() error: %v", err)
			}

			var buf strings.Builder
			err = r.WriteHTML(&buf)
			if err != nil {
				t.Fatalf("WriteHTML() error: %v", err)
			}

			actual := report.FormatHTML(buf.String())

			expected, err := os.ReadFile("testdata/" + name + ".html")
			if err != nil {
				t.Fatalf("ReadFile() error: %v", err)
			}

			if actual != string(expected) {
				t.Fatalf("output mismatch.\n%s", helpers.UnifiedDiff(t, string(expected), actual))
			}
		})
	}
}

func TestHeaderData(t *testing.T) {
	r := &Report{
		Report: &report.Report{
			Base: &report.Base{
				Name: "validate-application",
			},
		},
		Application: report.Application{
			Name:      "appset-deploy-rbd",
			Namespace: "argocd",
		},
	}

	d := &templateData{r}
	actual := d.HeaderData()

	expected := report.HeaderData{
		Title:    "Application Validation Report",
		Subtitle: "argocd / appset-deploy-rbd",
	}

	if actual != expected {
		t.Fatalf("mismatch.\n%s", helpers.UnifiedDiff(t, expected, actual))
	}
}
