// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package application

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
				Name:   "validate-application",
				Status: report.Passed,
			},
		},
		Application: report.Application{
			Name:      "myapp",
			Namespace: "mynamespace",
		},
		ApplicationStatus: report.ApplicationStatus{
			Hub: report.ApplicationStatusHub{
				DRPC: report.DRPCSummary{
					Name:      "myapp-drpc",
					Namespace: "myapp-ns",
					DRPolicy:  "my-dr-policy",
					Action: report.ValidatedString{
						Validated: report.Validated{State: report.OK},
						Value:     "Failover",
					},
					Phase: report.ValidatedString{
						Validated: report.Validated{State: report.OK},
						Value:     "FailedOver",
					},
					Progression: report.ValidatedString{
						Validated: report.Validated{
							State:       report.Problem,
							Description: "Waiting for progression \"Completed\"",
						},
						Value: "WaitForReadiness",
					},
				},
			},
			PrimaryCluster: report.ApplicationStatusCluster{
				Name: "dr1",
			},
			SecondaryCluster: report.ApplicationStatusCluster{
				Name: "dr2",
			},
			S3: report.ApplicationS3Status{
				Profiles: report.ValidatedApplicationS3ProfileStatusList{
					Validated: report.Validated{State: report.OK},
					Value: []report.ApplicationS3ProfileStatus{
						{
							Name: "s3-profile-1",
							Gathered: report.ValidatedBool{
								Validated: report.Validated{State: report.OK},
								Value:     true,
							},
						},
						{
							Name: "s3-profile-2",
							Gathered: report.ValidatedBool{
								Validated: report.Validated{
									State:       report.Problem,
									Description: "failed to connect to endpoint",
								},
								Value: false,
							},
						},
					},
				},
			},
		},
	}

	var buf strings.Builder
	err := r.WriteHTML(&buf)
	if err != nil {
		t.Fatalf("WriteHTML() error: %v", err)
	}

	actual := report.FormatHTML(buf.String())

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
				Name: "validate-application",
			},
		},
		Application: report.Application{
			Name:      "testapp",
			Namespace: "testns",
		},
	}

	d := &templateData{r}
	actual := d.HeaderData()

	expected := report.HeaderData{
		Title:    "Application Validation Report",
		Subtitle: "testns / testapp",
	}

	if actual != expected {
		t.Fatalf("mismatch.\n%s", helpers.UnifiedDiff(t, expected, actual))
	}
}
