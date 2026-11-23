// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"maps"
	"slices"
)

// GroupSummary is the summary of a group in a recipe.
type GroupSummary struct {
	Name           string `json:"name"`
	Type           string `json:"type,omitempty"`
	BackupRef      string `json:"backupRef,omitempty"`
	SelectResource string `json:"selectResource,omitempty"`
}

// HookSummary is the summary of a hook in a recipe.
type HookSummary struct {
	Name           string `json:"name"`
	Type           string `json:"type,omitempty"`
	Namespace      string `json:"namespace,omitempty"`
	SelectResource string `json:"selectResource,omitempty"`
	OnError        string `json:"onError,omitempty"`
}

// WorkflowSummary is the summary of a workflow in a recipe.
type WorkflowSummary struct {
	Name     string              `json:"name"`
	Sequence []map[string]string `json:"sequence,omitempty"`
}

// RecipeSummary is the recipe validation status report.
type RecipeSummary struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace,omitempty"`
	Groups    []GroupSummary    `json:"groups,omitempty"`
	Hooks     []HookSummary     `json:"hooks,omitempty"`
	Workflows []WorkflowSummary `json:"workflows,omitempty"`
}

// Equal returns true if group summary is equal to other group summary.
func (g *GroupSummary) Equal(o *GroupSummary) bool {
	if g == o {
		return true
	}
	if o == nil {
		return false
	}

	if g.Name != o.Name {
		return false
	}

	if g.Type != o.Type {
		return false
	}

	if g.BackupRef != o.BackupRef {
		return false
	}

	if g.SelectResource != o.SelectResource {
		return false
	}

	return true
}

// Equal returns true if hook summary is equal to other hook summary.
func (h *HookSummary) Equal(o *HookSummary) bool {
	if h == o {
		return true
	}
	if o == nil {
		return false
	}

	if h.Name != o.Name {
		return false
	}

	if h.Type != o.Type {
		return false
	}

	if h.Namespace != o.Namespace {
		return false
	}

	if h.SelectResource != o.SelectResource {
		return false
	}

	if h.OnError != o.OnError {
		return false
	}

	return true
}

// Equal returns true if workflow summary is equal to other workflow summary.
func (w *WorkflowSummary) Equal(o *WorkflowSummary) bool {
	if w == o {
		return true
	}
	if o == nil {
		return false
	}
	if w.Name != o.Name {
		return false
	}

	return slices.EqualFunc(w.Sequence, o.Sequence, maps.Equal)
}

// Equal returns true if recipe report is equal to other recipe report.
func (r *RecipeSummary) Equal(o *RecipeSummary) bool {
	if r == o {
		return true
	}
	if o == nil {
		return false
	}
	if r.Name != o.Name {
		return false
	}

	if r.Namespace != o.Namespace {
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
