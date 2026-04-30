// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindNextSuffix(t *testing.T) {
	const commandName = "validate-clusters"

	tests := []struct {
		name   string
		files  []string
		suffix string
	}{
		{
			name:   "missing output dir",
			suffix: "",
		},
		{
			name:   "empty output dir",
			files:  []string{},
			suffix: "",
		},
		{
			name: "second run",
			files: []string{
				"validate-clusters.log",
				"validate-clusters.yaml",
				"validate-clusters.html",
				"validate-clusters.data",
			},
			suffix: "-2",
		},
		{
			name: "third run",
			files: []string{
				"validate-clusters.log",
				"validate-clusters.yaml",
				"validate-clusters.html",
				"validate-clusters.data",
				"validate-clusters-2.log",
				"validate-clusters-2.yaml",
				"validate-clusters-2.html",
				"validate-clusters-2.data",
			},
			suffix: "-3",
		},
		{
			name: "partial previous run",
			files: []string{
				"validate-clusters.log",
			},
			suffix: "-2",
		},
		{
			name: "ignored files",
			files: []string{
				"validate-clusters.log",
				"validate-clusters.yaml",
				"validate-clusters.html",
				"validate-clusters.data",
				"validate-clusters.yaml.bak",
				"validate-clusters-error.log",
			},
			suffix: "-2",
		},
		{
			name: "gap in sequence numbers",
			files: []string{
				"validate-clusters.log",
				"validate-clusters-3.log",
			},
			suffix: "-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outDir := createOutDir(t, tt.files)
			suffix, err := findNextSuffix(outDir, commandName)
			if err != nil {
				t.Fatal(err)
			}
			if suffix != tt.suffix {
				t.Errorf("expected %q, got %q", tt.suffix, suffix)
			}
		})
	}
}

// createOutDir creates an output directory with the given files. If files is nil, returns a
// non-existent directory path.
func createOutDir(t *testing.T, files []string) string {
	outDir := t.TempDir()
	if files == nil {
		return filepath.Join(outDir, "missing")
	}
	for _, name := range files {
		if err := os.WriteFile(filepath.Join(outDir, name), nil, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return outDir
}
