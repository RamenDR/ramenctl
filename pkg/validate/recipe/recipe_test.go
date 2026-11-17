// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package recipe_test

import (
	"os"
	"strings"
	"testing"

	recipev1alpha1 "github.com/ramendr/recipe/api/v1alpha1"
	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/validate/recipe"
)

func loadRecipe(t *testing.T, path string) *recipev1alpha1.Recipe {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read %q: %v", path, err)
	}
	r := &recipev1alpha1.Recipe{}
	if err := yaml.Unmarshal(data, r); err != nil {
		t.Fatalf("failed to unmarshal %q: %v", path, err)
	}
	return r
}

func TestValidate_Valid(t *testing.T) {
	r := loadRecipe(t, "testdata/valid.yaml")
	status := recipe.Validate(r)

	if status.Name != "valid-recipe" || status.Namespace != "test-ns" {
		t.Fatalf("unexpected metadata in status: %+v", status)
	}
	if len(status.Workflows) != 2 {
		t.Fatalf("expected 2 workflows, got %+v", status.Workflows)
	}
	for _, w := range status.Workflows {
		if w.State != report.OK {
			t.Errorf(
				"expected workflow %q to be OK, got %q (description: %s)",
				w.Name,
				w.State,
				w.Description,
			)
		}
	}
}

func TestValidate_NoWorkflows(t *testing.T) {
	r := loadRecipe(t, "testdata/no-workflows.yaml")
	status := recipe.Validate(r)

	if len(status.Workflows) != 1 {
		t.Fatalf("expected 1 workflow error summary, got %+v", status.Workflows)
	}
	if status.Workflows[0].State != report.Problem {
		t.Errorf("expected problem state, got %q", status.Workflows[0].State)
	}
	if status.Workflows[0].Description != "no workflows defined" {
		t.Errorf("unexpected description: %q", status.Workflows[0].Description)
	}
}

func TestValidate_DuplicateWorkflowName(t *testing.T) {
	r := loadRecipe(t, "testdata/duplicate-workflow.yaml")
	status := recipe.Validate(r)

	for _, w := range status.Workflows {
		if w.Name == "backup" && w.State != report.Problem {
			t.Errorf("expected workflow %q to have Problem state, got %q", w.Name, w.State)
		}
	}
}

func TestValidate_InvalidFailOn(t *testing.T) {
	r := loadRecipe(t, "testdata/invalid-enum.yaml")
	status := recipe.Validate(r)

	for _, w := range status.Workflows {
		if w.Name == "backup" {
			if w.State != report.Problem {
				t.Errorf("expected Problem state for invalid failOn, got %q", w.State)
			}
		}
	}
}

func TestValidate_SequenceUnknownGroup(t *testing.T) {
	r := loadRecipe(t, "testdata/unknown-group.yaml")
	status := recipe.Validate(r)

	for _, w := range status.Workflows {
		if w.Name == "backup" {
			if w.State != report.Problem {
				t.Errorf("expected Problem state for unknown group, got %q", w.State)
			}
			if w.Description != "sequence item \"nonexistent-group\" does not reference a known group" {
				t.Errorf("unexpected description: %q", w.Description)
			}
		}
	}
}

func TestValidate_SequenceUnknownHook(t *testing.T) {
	r := loadRecipe(t, "testdata/unknown-hook.yaml")
	status := recipe.Validate(r)

	for _, w := range status.Workflows {
		if w.Name == "backup" {
			if w.State != report.Problem {
				t.Errorf("expected Problem state for unknown hook, got %q", w.State)
			}
			if w.Description != "sequence item \"nonexistent-hook/op\" does not reference a known hook" {
				t.Errorf("unexpected description: %q", w.Description)
			}
		}
	}
}

func TestValidate_BadSequenceRef(t *testing.T) {
	r := loadRecipe(t, "testdata/bad-sequence-ref.yaml")
	status := recipe.Validate(r)

	for _, w := range status.Workflows {
		if w.Name == "backup" && w.State != report.Problem {
			t.Errorf("expected Problem state for backup workflow, got %s", w.State)
		}
	}
}

func TestValidate_SequenceInvalidKey(t *testing.T) {
	r := loadRecipe(t, "testdata/invalid-sequence-key.yaml")
	status := recipe.Validate(r)

	for _, w := range status.Workflows {
		if w.Name == "backup" {
			if w.State != report.Problem {
				t.Errorf("expected Problem state for invalid sequence key, got %q", w.State)
			}
		}
	}
}

func TestValidate_MissingBackup(t *testing.T) {
	r := loadRecipe(t, "testdata/missing-backup.yaml")
	status := recipe.Validate(r)

	foundBackup := false
	for _, w := range status.Workflows {
		if w.Name == "backup" {
			foundBackup = true
			if w.State != report.Warning {
				t.Errorf("expected Warning state for missing backup workflow, got %q", w.State)
			}
		}
	}
	if !foundBackup {
		t.Error("expected missing backup workflow to be reported in workflows")
	}
}

func TestValidate_MissingRestore(t *testing.T) {
	r := loadRecipe(t, "testdata/missing-restore.yaml")
	status := recipe.Validate(r)

	foundRestore := false
	for _, w := range status.Workflows {
		if w.Name == "restore" {
			foundRestore = true
			if w.State != report.Warning {
				t.Errorf("expected Warning state for missing restore workflow, got %q", w.State)
			}
		}
	}
	if !foundRestore {
		t.Error("expected missing restore workflow to be reported in workflows")
	}
}

func TestValidate_ParamWarning(t *testing.T) {
	r := loadRecipe(t, "testdata/param-warning.yaml")
	status := recipe.Validate(r)

	foundRestore := false
	for _, w := range status.Workflows {
		if w.Name == "restore" {
			foundRestore = true
			if w.State != report.Warning {
				t.Errorf("expected Warning state for missing restore workflow, got %s", w.State)
			}
		}
	}
	if !foundRestore {
		t.Error("expected missing restore workflow to be reported")
	}
}

func TestValidate_MultipleProblemsInWorkflow(t *testing.T) {
	r := &recipev1alpha1.Recipe{
		Spec: recipev1alpha1.RecipeSpec{
			Workflows: []*recipev1alpha1.Workflow{
				{
					Name:   "backup",
					FailOn: "invalid-fail-on",
					Sequence: []map[string]string{
						{"group": "unknown-group"},
						{"hook": "unknown-hook"},
						{"invalid-kind": "something"},
					},
				},
				{
					Name: "restore",
				},
			},
		},
	}
	status := recipe.Validate(r)

	var backupWorkflow *report.WorkflowSummary
	for i := range status.Workflows {
		if status.Workflows[i].Name == "backup" {
			backupWorkflow = &status.Workflows[i]
			break
		}
	}

	if backupWorkflow == nil {
		t.Fatal("expected backup workflow")
	}
	if backupWorkflow.State != report.Problem {
		t.Fatalf("expected Problem state, got %v", backupWorkflow.State)
	}

	expectedSubstrings := []string{
		`failOn must be one of any-error, essential-error, full-error; got "invalid-fail-on"`,
		`sequence item "unknown-group" does not reference a known group`,
		`sequence item "unknown-hook" does not reference a known hook`,
		`sequence step key must be "group" or "hook", got "invalid-kind"`,
	}
	for _, expected := range expectedSubstrings {
		if !strings.Contains(backupWorkflow.Description, expected) {
			t.Errorf(
				"expected description to contain %q, got %q",
				expected,
				backupWorkflow.Description,
			)
		}
	}
}

func TestValidate_UnknownHookOperation(t *testing.T) {
	r := &recipev1alpha1.Recipe{
		Spec: recipev1alpha1.RecipeSpec{
			Hooks: []*recipev1alpha1.Hook{
				{
					Name: "app-hook",
					Ops: []*recipev1alpha1.Operation{
						{Name: "quiesce"},
					},
					Chks: []*recipev1alpha1.Check{
						{Name: "ready-chk"},
					},
					Jobs: []*recipev1alpha1.JobDetails{
						{Name: "backup-job"},
					},
				},
			},
			Workflows: []*recipev1alpha1.Workflow{
				{
					Name: "backup",
					Sequence: []map[string]string{
						{"hook": "app-hook/quiesce"},
						{"hook": "app-hook/ready-chk"},
						{"hook": "app-hook/backup-job"},
						{"hook": "app-hook/nonexistent-op"},
						{"hook": "app-hook"},
					},
				},
				{
					Name: "restore",
				},
			},
		},
	}
	status := recipe.Validate(r)

	var backupWorkflow *report.WorkflowSummary
	for i := range status.Workflows {
		if status.Workflows[i].Name == "backup" {
			backupWorkflow = &status.Workflows[i]
			break
		}
	}

	if backupWorkflow == nil {
		t.Fatal("expected backup workflow")
	}
	if backupWorkflow.State != report.Problem {
		t.Fatalf("expected Problem state, got %v", backupWorkflow.State)
	}

	expectedSubstrings := []string{
		`sequence item "app-hook/nonexistent-op" does not reference a known hook operation`,
		`sequence item "app-hook" is missing a hook operation`,
	}
	for _, expected := range expectedSubstrings {
		if !strings.Contains(backupWorkflow.Description, expected) {
			t.Errorf("expected description to contain %q, got %q", expected, backupWorkflow.Description)
		}
	}
}
