// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import "slices"

// GroupSummary is the summary of a group in a recipe.
type GroupSummary struct {
	Validated
	Name           string `json:"name"`
	Type           string `json:"type,omitempty"`
	BackupRef      string `json:"backupRef,omitempty"`
	SelectResource string `json:"selectResource,omitempty"`
}

// Equal returns true if group summary is equal to other group summary.
func (g *GroupSummary) Equal(o *GroupSummary) bool {
	if g == o {
		return true
	}
	if g == nil || o == nil {
		return false
	}
	return g.Validated == o.Validated &&
		g.Name == o.Name &&
		g.Type == o.Type &&
		g.BackupRef == o.BackupRef &&
		g.SelectResource == o.SelectResource
}

// HookSummary is the summary of a hook in a recipe.
type HookSummary struct {
	Validated
	Name           string `json:"name"`
	Type           string `json:"type,omitempty"`
	Namespace      string `json:"namespace,omitempty"`
	SelectResource string `json:"selectResource,omitempty"`
	OnError        string `json:"onError,omitempty"`
}

// Equal returns true if hook summary is equal to other hook summary.
func (h *HookSummary) Equal(o *HookSummary) bool {
	if h == o {
		return true
	}
	if h == nil || o == nil {
		return false
	}
	return h.Validated == o.Validated &&
		h.Name == o.Name &&
		h.Type == o.Type &&
		h.Namespace == o.Namespace &&
		h.SelectResource == o.SelectResource &&
		h.OnError == o.OnError
}

// WorkflowSummary is the summary of a workflow in a recipe.
type WorkflowSummary struct {
	Validated
	Name     string              `json:"name"`
	Sequence []map[string]string `json:"sequence,omitempty"`
}

// Equal returns true if workflow summary is equal to other workflow summary.
func (w *WorkflowSummary) Equal(o *WorkflowSummary) bool {
	if w == o {
		return true
	}
	if w == nil || o == nil {
		return false
	}
	if w.Validated != o.Validated || w.Name != o.Name {
		return false
	}
	if len(w.Sequence) != len(o.Sequence) {
		return false
	}
	return slices.EqualFunc(w.Sequence, o.Sequence, func(a, b map[string]string) bool {
		if len(a) != len(b) {
			return false
		}
		for k, v := range a {
			if b[k] != v {
				return false
			}
		}
		return true
	})
}

// RecipeReport is the recipe validation status report.
type RecipeReport struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace,omitempty"`
	Groups    []GroupSummary    `json:"groups,omitempty"`
	Hooks     []HookSummary     `json:"hooks,omitempty"`
	Workflows []WorkflowSummary `json:"workflows,omitempty"`
}

// Equal returns true if recipe report is equal to other recipe report.
func (r *RecipeReport) Equal(o *RecipeReport) bool {
	if r == o {
		return true
	}
	if r == nil || o == nil {
		return false
	}
	if r.Name != o.Name || r.Namespace != o.Namespace {
		return false
	}
	if !slices.EqualFunc(r.Groups, o.Groups, func(a, b GroupSummary) bool {
		return a.Equal(&b)
	}) {
		return false
	}
	if !slices.EqualFunc(r.Hooks, o.Hooks, func(a, b HookSummary) bool {
		return a.Equal(&b)
	}) {
		return false
	}
	return slices.EqualFunc(r.Workflows, o.Workflows, func(a, b WorkflowSummary) bool {
		return a.Equal(&b)
	})
}
