// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"testing"
)

func TestTemplate(t *testing.T) {
	tmpl, err := Template()
	if err != nil {
		t.Fatalf("Template() error: %v", err)
	}

	// Check that shared templates are defined
	for _, name := range []string{"report.tmpl", "style", "validated"} {
		if tmpl.Lookup(name) == nil {
			t.Errorf("template %q not defined", name)
		}
	}
}
