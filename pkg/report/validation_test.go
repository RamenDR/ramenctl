// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report_test

import (
	"fmt"
	"testing"

	"sigs.k8s.io/yaml"

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
