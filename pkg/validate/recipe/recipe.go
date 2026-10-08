// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package recipe

import (
	"fmt"
	"strings"

	recipev1alpha1 "github.com/ramendr/recipe/api/v1alpha1"

	"github.com/ramendr/ramenctl/pkg/report"
)

// Validate runs all semantic checks on r and returns the validated RecipeStatus.
func Validate(r *recipev1alpha1.Recipe) *report.RecipeReport {
	status := &report.RecipeReport{
		Name:      r.Name,
		Namespace: r.Namespace,
		Workflows: []report.WorkflowSummary{},
	}

	knownGroups := map[string]bool{}
	if v := r.Spec.Volumes; v != nil {
		knownGroups[v.Name] = true
	}
	for _, g := range r.Spec.Groups {
		if g != nil {
			knownGroups[g.Name] = true
		}
	}

	knownHooks := map[string]bool{}
	for _, h := range r.Spec.Hooks {
		if h != nil {
			knownHooks[h.Name] = true
		}
	}

	// No workflows defined is a problem.
	if len(r.Spec.Workflows) == 0 {
		status.Workflows = append(status.Workflows, report.WorkflowSummary{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "no workflows defined",
			},
		})
		return status
	}

	seenWorkflows := map[string]int{}
	for _, w := range r.Spec.Workflows {
		if w == nil {
			continue
		}
		seenWorkflows[w.Name]++
		summary := validateWorkflow(w, knownGroups, knownHooks)
		status.Workflows = append(status.Workflows, summary)
	}

	// Check if duplicate workflow names exist.
	for i := range status.Workflows {
		wName := status.Workflows[i].Name
		if seenWorkflows[wName] > 1 {
			status.Workflows[i].State = report.Problem
			status.Workflows[i].Description = fmt.Sprintf("duplicate workflow name %q", wName)
		}
	}

	// If "backup" is missing, add a warning entry.
	if seenWorkflows[recipev1alpha1.BackupWorkflowName] == 0 {
		status.Workflows = append(status.Workflows, report.WorkflowSummary{
			Name: recipev1alpha1.BackupWorkflowName,
			Validated: report.Validated{
				State: report.Warning,
				Description: fmt.Sprintf(
					"workflow %q is missing",
					recipev1alpha1.BackupWorkflowName,
				),
			},
		})
	}

	// If "restore" is missing, add a warning entry.
	if seenWorkflows[recipev1alpha1.RestoreWorkflowName] == 0 {
		status.Workflows = append(status.Workflows, report.WorkflowSummary{
			Name: recipev1alpha1.RestoreWorkflowName,
			Validated: report.Validated{
				State: report.Warning,
				Description: fmt.Sprintf(
					"workflow %q is missing",
					recipev1alpha1.RestoreWorkflowName,
				),
			},
		})
	}

	return status
}

func validateWorkflow(
	w *recipev1alpha1.Workflow,
	knownGroups, knownHooks map[string]bool,
) report.WorkflowSummary {
	summary := report.WorkflowSummary{
		Name:     w.Name,
		Sequence: w.Sequence,
		Validated: report.Validated{
			State: report.OK,
		},
	}

	if w.FailOn != "" {
		switch w.FailOn {
		case "any-error", "essential-error", "full-error":
			// valid
		default:
			summary.State = report.Problem
			summary.Description = fmt.Sprintf(
				"failOn must be one of any-error, essential-error, full-error; got %q",
				w.FailOn,
			)
			return summary
		}
	}

	for _, step := range w.Sequence {
		for kind, name := range step {
			switch kind {
			case "group":
				if !knownGroups[name] {
					summary.State = report.Problem
					summary.Description = fmt.Sprintf(
						"sequence item %q does not reference a known group",
						name,
					)
					return summary
				}
			case "hook":
				hookName, _, _ := strings.Cut(name, "/")
				if !knownHooks[hookName] {
					summary.State = report.Problem
					summary.Description = fmt.Sprintf(
						"sequence item %q does not reference a known hook",
						name,
					)
					return summary
				}
			default:
				summary.State = report.Problem
				summary.Description = fmt.Sprintf(
					"sequence step key must be \"group\" or \"hook\", got %q",
					kind,
				)
				return summary
			}
		}
	}

	return summary
}
