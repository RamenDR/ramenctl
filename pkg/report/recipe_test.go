// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report_test

import (
	"slices"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
)

func TestGroupSummaryEqual(t *testing.T) {
	g1 := testGroupSummary()
	t.Run("equal to self", func(t *testing.T) {
		g2 := g1
		if !g1.Equal(g2) {
			t.Error("group summary not equal to self")
		}
	})
	t.Run("not equal to nil", func(t *testing.T) {
		if g1.Equal(nil) {
			t.Error("group summary equal to nil")
		}
	})
	t.Run("nil equal to nil", func(t *testing.T) {
		var nilGroup *report.GroupSummary
		if !nilGroup.Equal(nil) {
			t.Error("nil group summary not equal to nil")
		}
	})

	cases := []struct {
		name   string
		modify func(g *report.GroupSummary)
		equal  bool
	}{
		{"equal values", func(*report.GroupSummary) {}, true},
		{"name", func(g *report.GroupSummary) {
			g.Name = helpers.Modified
		}, false},
		{"type", func(g *report.GroupSummary) {
			g.Type = helpers.Modified
		}, false},
		{"backupRef", func(g *report.GroupSummary) {
			g.BackupRef = helpers.Modified
		}, false},
		{"selectResource", func(g *report.GroupSummary) {
			g.SelectResource = helpers.Modified
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g2 := testGroupSummary()
			tc.modify(g2)
			if got := g1.Equal(g2); got != tc.equal {
				t.Errorf("g1.Equal(g2) = %v, want %v", got, tc.equal)
			}
			if got := g2.Equal(g1); got != tc.equal {
				t.Errorf("g2.Equal(g1) = %v, want %v", got, tc.equal)
			}
		})
	}
}

func TestHookSummaryEqual(t *testing.T) {
	h1 := testHookSummary()
	t.Run("equal to self", func(t *testing.T) {
		h2 := h1
		if !h1.Equal(h2) {
			t.Error("hook summary not equal to self")
		}
	})
	t.Run("not equal to nil", func(t *testing.T) {
		if h1.Equal(nil) {
			t.Error("hook summary equal to nil")
		}
	})
	t.Run("nil equal to nil", func(t *testing.T) {
		var nilHook *report.HookSummary
		if !nilHook.Equal(nil) {
			t.Error("nil hook summary not equal to nil")
		}
	})

	cases := []struct {
		name   string
		modify func(h *report.HookSummary)
		equal  bool
	}{
		{"equal values", func(*report.HookSummary) {}, true},
		{"name", func(h *report.HookSummary) {
			h.Name = helpers.Modified
		}, false},
		{"type", func(h *report.HookSummary) {
			h.Type = helpers.Modified
		}, false},
		{"namespace", func(h *report.HookSummary) {
			h.Namespace = helpers.Modified
		}, false},
		{"selectResource", func(h *report.HookSummary) {
			h.SelectResource = helpers.Modified
		}, false},
		{"onError", func(h *report.HookSummary) {
			h.OnError = helpers.Modified
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h2 := testHookSummary()
			tc.modify(h2)
			if got := h1.Equal(h2); got != tc.equal {
				t.Errorf("h1.Equal(h2) = %v, want %v", got, tc.equal)
			}
			if got := h2.Equal(h1); got != tc.equal {
				t.Errorf("h2.Equal(h1) = %v, want %v", got, tc.equal)
			}
		})
	}
}

func TestWorkflowSummaryEqual(t *testing.T) {
	w1 := testWorkflowSummary()
	t.Run("equal to self", func(t *testing.T) {
		w2 := w1
		if !w1.Equal(w2) {
			t.Error("workflow summary not equal to self")
		}
	})
	t.Run("not equal to nil", func(t *testing.T) {
		if w1.Equal(nil) {
			t.Error("workflow summary equal to nil")
		}
	})
	t.Run("nil equal to nil", func(t *testing.T) {
		var nilWorkflow *report.WorkflowSummary
		if !nilWorkflow.Equal(nil) {
			t.Error("nil workflow summary not equal to nil")
		}
	})

	cases := []struct {
		name   string
		modify func(w *report.WorkflowSummary)
		equal  bool
	}{
		{"equal values", func(*report.WorkflowSummary) {}, true},
		{"name", func(w *report.WorkflowSummary) {
			w.Name = helpers.Modified
		}, false},
		{"sequence entry added", func(w *report.WorkflowSummary) {
			w.Sequence = append(w.Sequence, map[string]string{"hook": "cleanup"})
		}, false},
		{"sequence entry removed", func(w *report.WorkflowSummary) {
			w.Sequence = w.Sequence[:1]
		}, false},
		{"sequence order", func(w *report.WorkflowSummary) {
			slices.Reverse(w.Sequence)
		}, false},
		{"sequence entry value", func(w *report.WorkflowSummary) {
			w.Sequence[0]["group"] = helpers.Modified
		}, false},
		{"sequence entry key", func(w *report.WorkflowSummary) {
			w.Sequence[0] = map[string]string{"hook": "data"}
		}, false},
		{"sequence entry extra key", func(w *report.WorkflowSummary) {
			w.Sequence[0]["hook"] = "check"
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w2 := testWorkflowSummary()
			tc.modify(w2)
			if got := w1.Equal(w2); got != tc.equal {
				t.Errorf("w1.Equal(w2) = %v, want %v", got, tc.equal)
			}
			if got := w2.Equal(w1); got != tc.equal {
				t.Errorf("w2.Equal(w1) = %v, want %v", got, tc.equal)
			}
		})
	}
}

func TestWorkflowSummaryEqualMissingKey(t *testing.T) {
	cases := []struct {
		name string
		a, b map[string]string
	}{
		{
			name: "missing key, empty value",
			a:    map[string]string{"group": ""},
			b:    map[string]string{"hook": "check"},
		},
		{
			name: "different keys, both empty",
			a:    map[string]string{"group": ""},
			b:    map[string]string{"hook": ""},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w1 := &report.WorkflowSummary{Name: "backup", Sequence: []map[string]string{tc.a}}
			w2 := &report.WorkflowSummary{Name: "backup", Sequence: []map[string]string{tc.b}}
			if w1.Equal(w2) {
				t.Error("w1.Equal(w2) = true, want false")
			}
			if w2.Equal(w1) {
				t.Error("w2.Equal(w1) = true, want false")
			}
		})
	}
}

func TestRecipeSummaryEqual(t *testing.T) {
	r1 := testRecipeSummary()
	t.Run("equal to self", func(t *testing.T) {
		r2 := r1
		if !r1.Equal(r2) {
			t.Error("recipe summary not equal to self")
		}
	})
	t.Run("not equal to nil", func(t *testing.T) {
		if r1.Equal(nil) {
			t.Error("recipe summary equal to nil")
		}
	})
	t.Run("nil equal to nil", func(t *testing.T) {
		var nilRecipe *report.RecipeSummary
		if !nilRecipe.Equal(nil) {
			t.Error("nil recipe summary not equal to nil")
		}
	})

	cases := []struct {
		name   string
		modify func(r *report.RecipeSummary)
		equal  bool
	}{
		{"equal values", func(*report.RecipeSummary) {}, true},
		{"name", func(r *report.RecipeSummary) {
			r.Name = helpers.Modified
		}, false},
		{"namespace", func(r *report.RecipeSummary) {
			r.Namespace = helpers.Modified
		}, false},
		{"group", func(r *report.RecipeSummary) {
			r.Groups[0].Name = helpers.Modified
		}, false},
		{"hook", func(r *report.RecipeSummary) {
			r.Hooks[0].OnError = helpers.Modified
		}, false},
		{"workflow", func(r *report.RecipeSummary) {
			r.Workflows[0].Name = helpers.Modified
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r2 := testRecipeSummary()
			tc.modify(r2)
			if got := r1.Equal(r2); got != tc.equal {
				t.Errorf("r1.Equal(r2) = %v, want %v", got, tc.equal)
			}
			if got := r2.Equal(r1); got != tc.equal {
				t.Errorf("r2.Equal(r1) = %v, want %v", got, tc.equal)
			}
		})
	}
}

func TestRecipeSummaryMarshaling(t *testing.T) {
	r1 := testRecipeSummary()
	data, err := yaml.Marshal(r1)
	if err != nil {
		t.Fatal(err)
	}
	r2 := &report.RecipeSummary{}
	if err := yaml.Unmarshal(data, r2); err != nil {
		t.Fatal(err)
	}
	if !r1.Equal(r2) {
		t.Fatalf("recipe summaries are not equal\n%s", helpers.UnifiedDiff(t, r1, r2))
	}
}

func testGroupSummary() *report.GroupSummary {
	return &report.GroupSummary{
		Name:           "data",
		Type:           "volume",
		BackupRef:      "backup",
		SelectResource: "pvc",
	}
}

func testHookSummary() *report.HookSummary {
	return &report.HookSummary{
		Name:           "check",
		Type:           "exec",
		Namespace:      "app",
		SelectResource: "deployment",
		OnError:        "fail",
	}
}

func testWorkflowSummary() *report.WorkflowSummary {
	return &report.WorkflowSummary{
		Name: "backup",
		Sequence: []map[string]string{
			{"group": "data"},
			{"hook": "check"},
		},
	}
}

func testRecipeSummary() *report.RecipeSummary {
	return &report.RecipeSummary{
		Name:      "recipe",
		Namespace: "app",
		Groups:    []report.GroupSummary{*testGroupSummary()},
		Hooks:     []report.HookSummary{*testHookSummary()},
		Workflows: []report.WorkflowSummary{*testWorkflowSummary()},
	}
}
