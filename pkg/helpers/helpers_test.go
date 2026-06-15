// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package helpers_test

import (
	"testing"

	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
)

func TestUnifiedDiff(t *testing.T) {
	expected := report.Build{Version: "v0.12.0", Commit: "269c8c474c35e17768abd5f9831d3fde642d20c3"}

	t.Run("same", func(t *testing.T) {
		diff := helpers.UnifiedDiff(t, expected, expected)
		if diff != "" {
			t.Fatalf("expected no diff, got %s", diff)
		}
	})

	t.Run("modified property", func(t *testing.T) {
		actual := report.Build{Version: "v0.13.0", Commit: expected.Commit}
		actualDiff := helpers.UnifiedDiff(t, expected, actual)
		expectedDiff := `--- expected
+++ actual
@@ -1,2 +1,2 @@
 commit: 269c8c474c35e17768abd5f9831d3fde642d20c3
-version: v0.12.0
+version: v0.13.0
`
		if actualDiff != expectedDiff {
			t.Fatalf("expected\n%s\ngot:\n%s", expectedDiff, actualDiff)
		}
	})

	t.Run("empty property", func(t *testing.T) {
		actual := report.Build{Version: expected.Version}
		actualDiff := helpers.UnifiedDiff(t, expected, actual)
		expectedDiff := `--- expected
+++ actual
@@ -1,2 +1 @@
-commit: 269c8c474c35e17768abd5f9831d3fde642d20c3
 version: v0.12.0
`
		if actualDiff != expectedDiff {
			t.Fatalf("expected:\n%s\ngot:\n%s", expectedDiff, actualDiff)
		}
	})

	t.Run("nill", func(t *testing.T) {
		actualDiff := helpers.UnifiedDiff(t, &expected, nil)
		expectedDiff := `--- expected
+++ actual
@@ -1,2 +1 @@
-commit: 269c8c474c35e17768abd5f9831d3fde642d20c3
-version: v0.12.0
+null
`
		if actualDiff != expectedDiff {
			t.Fatalf("expected:\n%s\ngot:\n%s", expectedDiff, actualDiff)
		}
	})

	t.Run("strings", func(t *testing.T) {
		a := `line 1
line 2
`
		b := `line 1
modified
`
		actualDiff := helpers.UnifiedDiff(t, a, b)
		expectedDiff := `--- expected
+++ actual
@@ -1,2 +1,2 @@
 line 1
-line 2
+modified
`
		if actualDiff != expectedDiff {
			t.Fatalf("expected:\n%s\ngot:\n%s", expectedDiff, actualDiff)
		}
	})
}
