// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report_test

import (
	"testing"

	"github.com/ramendr/ramenctl/pkg/report"
)

func TestRecipeReportEqual(t *testing.T) {
	r1 := &report.RecipeReport{
		Name:      "my-recipe",
		Namespace: "my-namespace",
		Groups: []report.GroupSummary{
			{
				Name: "group1",
				Type: "volume",
				Validated: report.Validated{
					State: report.OK,
				},
			},
		},
		Hooks: []report.HookSummary{
			{
				Name:      "hook1",
				Type:      "exec",
				Namespace: "default",
				Validated: report.Validated{
					State: report.OK,
				},
			},
		},
		Workflows: []report.WorkflowSummary{
			{
				Name: "backup",
				Validated: report.Validated{
					State: report.OK,
				},
				Sequence: []map[string]string{{"group": "group1"}},
			},
		},
	}
	r2 := &report.RecipeReport{
		Name:      "my-recipe",
		Namespace: "my-namespace",
		Groups: []report.GroupSummary{
			{
				Name: "group1",
				Type: "volume",
				Validated: report.Validated{
					State: report.OK,
				},
			},
		},
		Hooks: []report.HookSummary{
			{
				Name:      "hook1",
				Type:      "exec",
				Namespace: "default",
				Validated: report.Validated{
					State: report.OK,
				},
			},
		},
		Workflows: []report.WorkflowSummary{
			{
				Name: "backup",
				Validated: report.Validated{
					State: report.OK,
				},
				Sequence: []map[string]string{{"group": "group1"}},
			},
		},
	}
	if !r1.Equal(r2) {
		t.Fatal("expected r1 to equal r2")
	}
}

func TestRecipeReportNotEqual(t *testing.T) {
	r1 := &report.RecipeReport{
		Name:      "my-recipe",
		Namespace: "my-namespace",
		Groups: []report.GroupSummary{
			{
				Name: "group1",
				Type: "volume",
			},
		},
		Hooks: []report.HookSummary{
			{
				Name: "hook1",
				Type: "exec",
			},
		},
		Workflows: []report.WorkflowSummary{
			{
				Name: "backup",
				Validated: report.Validated{
					State: report.OK,
				},
			},
		},
	}

	if r1.Equal(nil) {
		t.Fatal("expected r1 not to equal nil")
	}

	r2 := &report.RecipeReport{
		Name:      "different-recipe",
		Namespace: "my-namespace",
	}
	if r1.Equal(r2) {
		t.Fatal("expected r1 not to equal r2 with different name")
	}

	r3 := &report.RecipeReport{
		Name:      "my-recipe",
		Namespace: "different-namespace",
	}
	if r1.Equal(r3) {
		t.Fatal("expected r1 not to equal r3 with different namespace")
	}

	r4 := &report.RecipeReport{
		Name:      "my-recipe",
		Namespace: "my-namespace",
		Groups: []report.GroupSummary{
			{
				Name: "different-group",
			},
		},
	}
	if r1.Equal(r4) {
		t.Fatal("expected r1 not to equal r4 with different groups")
	}

	r5 := &report.RecipeReport{
		Name:      "my-recipe",
		Namespace: "my-namespace",
		Hooks: []report.HookSummary{
			{
				Name: "different-hook",
			},
		},
	}
	if r1.Equal(r5) {
		t.Fatal("expected r1 not to equal r5 with different hooks")
	}

	r6 := &report.RecipeReport{
		Name:      "my-recipe",
		Namespace: "my-namespace",
		Workflows: []report.WorkflowSummary{
			{
				Name: "restore",
				Validated: report.Validated{
					State: report.OK,
				},
			},
		},
	}
	if r1.Equal(r6) {
		t.Fatal("expected r1 not to equal r6 with different workflows")
	}
}

func TestGroupSummaryEqual(t *testing.T) {
	g1 := &report.GroupSummary{
		Name:           "g1",
		Type:           "resource",
		BackupRef:      "b1",
		SelectResource: "deployment",
		Validated: report.Validated{
			State: report.OK,
		},
	}
	g2 := &report.GroupSummary{
		Name:           "g1",
		Type:           "resource",
		BackupRef:      "b1",
		SelectResource: "deployment",
		Validated: report.Validated{
			State: report.OK,
		},
	}
	if !g1.Equal(g2) {
		t.Fatal("expected g1 to equal g2")
	}

	if g1.Equal(nil) {
		t.Fatal("expected g1 not to equal nil")
	}
}

func TestHookSummaryEqual(t *testing.T) {
	h1 := &report.HookSummary{
		Name:           "h1",
		Type:           "exec",
		Namespace:      "default",
		SelectResource: "pod",
		OnError:        "fail",
		Validated: report.Validated{
			State: report.OK,
		},
	}
	h2 := &report.HookSummary{
		Name:           "h1",
		Type:           "exec",
		Namespace:      "default",
		SelectResource: "pod",
		OnError:        "fail",
		Validated: report.Validated{
			State: report.OK,
		},
	}
	if !h1.Equal(h2) {
		t.Fatal("expected h1 to equal h2")
	}

	if h1.Equal(nil) {
		t.Fatal("expected h1 not to equal nil")
	}
}
