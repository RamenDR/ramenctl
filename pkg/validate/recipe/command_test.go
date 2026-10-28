// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package recipe_test

import (
	"os"
	"path/filepath"
	"testing"

	"sigs.k8s.io/yaml"

	basecmd "github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/validate/recipe"
	"github.com/ramendr/ramenctl/pkg/validate/summary"
)

func testCommand(t *testing.T, recipeFile string) (*recipe.Command, string) {
	t.Helper()
	outputDir := t.TempDir()
	cmd, err := basecmd.ForTest(recipe.CommandName, nil, outputDir)
	if err != nil {
		t.Fatalf("failed to create command: %v", err)
	}
	return recipe.NewCommand(cmd, recipeFile), outputDir
}

func TestCommand_Valid(t *testing.T) {
	cmd, outputDir := testCommand(t, "testdata/valid.yaml")
	if err := cmd.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cmd.Report.Summary.Get(summary.Problem) != 0 {
		t.Errorf("expected 0 problems, got %d", cmd.Report.Summary.Get(summary.Problem))
	}
	if cmd.Report.Summary.Get(summary.OK) != 2 {
		t.Errorf("expected 2 ok workflows, got %d", cmd.Report.Summary.Get(summary.OK))
	}

	reportPath := filepath.Join(outputDir, "validate-recipe.yaml")
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatalf("failed to read report file: %v", err)
	}

	var rep recipe.Report
	if err := yaml.Unmarshal(data, &rep); err != nil {
		t.Fatalf("failed to unmarshal report: %v", err)
	}
	if rep.Status != report.Passed {
		t.Errorf("expected status %v, got %v", report.Passed, rep.Status)
	}
}

func TestCommand_Problem(t *testing.T) {
	cmd, outputDir := testCommand(t, "testdata/no-workflows.yaml")
	if err := cmd.Run(); err == nil {
		t.Fatal("expected error, got nil")
	}

	if cmd.Report.Summary.Get(summary.Problem) != 1 {
		t.Errorf("expected 1 problem, got %d", cmd.Report.Summary.Get(summary.Problem))
	}

	reportPath := filepath.Join(outputDir, "validate-recipe.yaml")
	if _, err := os.Stat(reportPath); err != nil {
		t.Fatalf("report file should exist even on failure: %v", err)
	}
}

func TestCommand_MissingFile(t *testing.T) {
	cmd, outputDir := testCommand(t, "testdata/nonexistent.yaml")
	if err := cmd.Run(); err == nil {
		t.Fatal("expected error, got nil")
	}

	reportPath := filepath.Join(outputDir, "validate-recipe.yaml")
	if _, err := os.Stat(reportPath); err != nil {
		t.Fatalf("report file should exist even on failure: %v", err)
	}
}
