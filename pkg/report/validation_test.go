// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report_test

import (
	"fmt"
	"testing"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
)

func TestEmojiRoundtrip(t *testing.T) {
	cases := []struct {
		name  string
		state report.ValidationState
	}{
		{"ok", report.OK},
		{"problem", report.Problem},
		{"stale", report.Stale},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v1 := report.Validated{State: tc.state}
			b, err := yaml.Marshal(&v1)
			if err != nil {
				t.Fatalf("failed to marshal state: %v", err)
			}
			// For inspecting the yaml
			fmt.Print(string(b))
			v2 := report.Validated{}
			if err := yaml.Unmarshal(b, &v2); err != nil {
				t.Fatalf("failed to unmarshal state: %v", err)
			}
			if v1 != v2 {
				t.Fatalf("expected %+v, got %+v", v1, v2)
			}
		})
	}
}

func TestValidatedDurationRoundtrip(t *testing.T) {
	cases := []struct {
		name     string
		original report.ValidatedDuration
		yaml     string
	}{
		{
			name: "ok",
			original: report.ValidatedDuration{
				Validated: report.Validated{State: report.OK},
				Value:     5 * time.Minute,
			},
			yaml: "state: ok ✅\nvalue: 5m0s\n",
		},
		{
			name: "problem",
			original: report.ValidatedDuration{
				Validated: report.Validated{
					State:       report.Problem,
					Description: "Invalid scheduling interval",
				},
			},
			yaml: "description: Invalid scheduling interval\nstate: problem ❌\n",
		},
		{
			name: "zero value",
			original: report.ValidatedDuration{
				Validated: report.Validated{State: report.OK},
			},
			yaml: "state: ok ✅\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tc.original)
			if err != nil {
				t.Fatalf("failed to marshal: %v", err)
			}

			if string(data) != tc.yaml {
				t.Fatalf("unexpected yaml\n%s", helpers.UnifiedDiff(t, tc.yaml, string(data)))
			}

			var decoded report.ValidatedDuration
			if err := yaml.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			if decoded != tc.original {
				t.Fatalf("roundtrip mismatch\n%s", helpers.UnifiedDiff(t, tc.original, decoded))
			}
		})
	}
}
