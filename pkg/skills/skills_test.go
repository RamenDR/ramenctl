// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package skills_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ramendr/ramenctl/pkg/skills"
)

var expectedSkills = []string{
	"gather-application",
	"init",
	"test-clean",
	"test-run",
	"validate-application",
	"validate-clusters",
}

// Install per agent.

func TestInstallGeneric(t *testing.T) {
	t.Chdir(t.TempDir())

	if !skills.Install("ramenctl", skills.AgentGeneric) {
		t.Fatal("install failed")
	}

	validateGeneric(t, "ramenctl")
}

func TestInstallCursor(t *testing.T) {
	t.Chdir(t.TempDir())

	if !skills.Install("ramenctl", skills.AgentCursor) {
		t.Fatal("install failed")
	}

	validateCursor(t, "ramenctl")
}

func TestInstallClaude(t *testing.T) {
	t.Chdir(t.TempDir())

	if !skills.Install("ramenctl", skills.AgentClaude) {
		t.Fatal("install failed")
	}

	validateClaude(t, "ramenctl")
}

func TestInstallBob(t *testing.T) {
	t.Chdir(t.TempDir())

	if !skills.Install("ramenctl", skills.AgentBob) {
		t.Fatal("install failed")
	}

	validateBob(t, "ramenctl")
}

// Command name substitution.

func TestInstallCommandName(t *testing.T) {
	t.Chdir(t.TempDir())

	if !skills.Install("odf dr", skills.AgentCursor) {
		t.Fatal("install failed")
	}

	validateSkills(t, ".cursor/skills", "odf-dr")
	validateContextFile(t, ".cursor/rules/ramenctl.mdc", "odf dr")
	assertContextContains(t, ".cursor/skills/odf-dr-init/SKILL.md", "odf dr")
}

// Multiple agents. Users may run init with different agents in the same
// directory. Verify that all agents validate regardless of install order.

func TestInstallAllAgents(t *testing.T) {
	for _, tt := range []struct {
		name   string
		agents []string
	}{
		{"forward", []string{skills.AgentGeneric, skills.AgentCursor, skills.AgentClaude, skills.AgentBob}},
		{"reverse", []string{skills.AgentBob, skills.AgentClaude, skills.AgentCursor, skills.AgentGeneric}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Chdir(t.TempDir())

			for _, agent := range tt.agents {
				if !skills.Install("ramenctl", agent) {
					t.Fatalf("install %s failed", agent)
				}
			}

			validateGeneric(t, "ramenctl")
			validateCursor(t, "ramenctl")
			validateClaude(t, "ramenctl")
			validateBob(t, "ramenctl")
		})
	}
}

// Skills write-once.

func TestInstallAllSkipped(t *testing.T) {
	t.Chdir(t.TempDir())

	if !skills.Install("ramenctl", skills.AgentGeneric) {
		t.Fatal("first install failed")
	}

	// Append to a skill file to simulate user changes.
	path := filepath.Join(".agents", "skills", "ramenctl-init", "SKILL.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content = append(content, "\n# User notes\n"...)
	if err := os.WriteFile(path, content, 0o640); err != nil {
		t.Fatal(err)
	}

	// Second install should succeed but not overwrite existing files.
	if !skills.Install("ramenctl", skills.AgentGeneric) {
		t.Fatal("second install failed")
	}

	validateSkills(t, ".agents/skills", "ramenctl")

	content, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "# User notes") {
		t.Error("install should not overwrite existing skill files")
	}
}

func TestInstallSomeSkipped(t *testing.T) {
	for _, tt := range []struct {
		agent     string
		skillsDir string
	}{
		{skills.AgentGeneric, ".agents/skills"},
		{skills.AgentCursor, ".cursor/skills"},
	} {
		t.Run(tt.agent, func(t *testing.T) {
			t.Chdir(t.TempDir())

			if !skills.Install("ramenctl", tt.agent) {
				t.Fatal("first install failed")
			}

			if err := os.RemoveAll(filepath.Join(tt.skillsDir, "ramenctl-init")); err != nil {
				t.Fatal(err)
			}

			if !skills.Install("ramenctl", tt.agent) {
				t.Fatal("second install failed")
			}

			validateSkills(t, tt.skillsDir, "ramenctl")
		})
	}
}

func TestInstallEmptySkillDir(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.MkdirAll(".agents/skills/ramenctl-init", 0o755); err != nil {
		t.Fatal(err)
	}

	if !skills.Install("ramenctl", skills.AgentGeneric) {
		t.Fatal("install failed")
	}

	validateSkills(t, ".agents/skills", "ramenctl")
}

// Skills failures.

func TestInstallSkillsDirFailure(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.WriteFile(".agents", []byte("block"), 0o640); err != nil {
		t.Fatal(err)
	}

	if skills.Install("ramenctl", skills.AgentGeneric) {
		t.Fatal("install should fail when skills parent directory cannot be created")
	}
}

func TestInstallSkillsFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("chmod does not restrict directory writes on Windows")
	}
	t.Chdir(t.TempDir())

	if err := os.MkdirAll(".agents/skills", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(".agents/skills", 0o555); err != nil {
		t.Fatal(err)
	}
	if skills.Install("ramenctl", skills.AgentGeneric) {
		t.Fatal("install should fail when skills directory is not writable")
	}
}

func TestInstallSkillFileInsteadOfDir(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.MkdirAll(".agents/skills", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(".agents/skills/ramenctl-init", []byte("not a dir"), 0o640); err != nil {
		t.Fatal(err)
	}

	if skills.Install("ramenctl", skills.AgentGeneric) {
		t.Fatal("install should fail when skill path is a file instead of directory")
	}
}

// Context file failures.

func TestInstallContextFileDirFailure(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.MkdirAll(".cursor", 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(".cursor", "rules"), []byte("block"), 0o640); err != nil {
		t.Fatal(err)
	}

	if skills.Install("ramenctl", skills.AgentCursor) {
		t.Fatal("install should fail when context file directory cannot be created")
	}
}

func TestInstallContextFileWriteFailure(t *testing.T) {
	t.Chdir(t.TempDir())

	if err := os.MkdirAll("AGENTS.md", 0o755); err != nil {
		t.Fatal(err)
	}

	if skills.Install("ramenctl", skills.AgentGeneric) {
		t.Fatal("install should fail when context file path is a directory")
	}
}

// Agent validation.

func TestValidateAgent(t *testing.T) {
	for _, agent := range []string{
		skills.AgentCursor, skills.AgentClaude, skills.AgentBob, skills.AgentGeneric,
	} {
		if err := skills.ValidateAgent(agent); err != nil {
			t.Errorf("unexpected error for agent %q: %v", agent, err)
		}
	}
	if err := skills.ValidateAgent("invalid"); err == nil {
		t.Fatal("expected error for invalid agent")
	}
}

// Per-agent validators.

func validateGeneric(t *testing.T, commandSlug string) {
	t.Helper()
	validateSkills(t, ".agents/skills", commandSlug)
	validateContextFile(t, "AGENTS.md", commandSlug)
	validateContextSkillIndex(t, "AGENTS.md", ".agents/skills", commandSlug)
}

func validateCursor(t *testing.T, commandSlug string) {
	t.Helper()
	validateSkills(t, ".cursor/skills", commandSlug)
	validateContextFile(t, ".cursor/rules/ramenctl.mdc", commandSlug)
	assertContextContains(t, ".cursor/rules/ramenctl.mdc", "alwaysApply: true")
}

func validateClaude(t *testing.T, commandSlug string) {
	t.Helper()
	validateSkills(t, ".claude/skills", commandSlug)
	validateContextFile(t, "CLAUDE.md", commandSlug)
}

func validateBob(t *testing.T, commandSlug string) {
	t.Helper()
	validateSkills(t, ".bob/skills", commandSlug)
	validateContextFile(t, ".bob/rules/ramenctl.md", commandSlug)
	assertContextContains(t, ".bob/rules/ramenctl.md", "## Critical workflow rule")
	assertContextContains(t, ".bob/rules/ramenctl.md", "advanced mode")
}

// Helpers.

func assertContextContains(t *testing.T, path, expected string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), expected) {
		t.Errorf("%s should contain %q", path, expected)
	}
}

func validateContextFile(t *testing.T, path, commandName string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("missing context file %s: %v", path, err)
	}

	// Common content from common.tmpl.
	text := string(content)
	if !strings.Contains(text, "# "+commandName) {
		t.Errorf("%s should contain %q heading", path, commandName)
	}
	if !strings.Contains(text, "## Command execution rules") {
		t.Errorf("%s should contain common section %q", path, "## Command execution rules")
	}

	// Check for unexpanded template expressions.
	if strings.Contains(text, "{{") {
		t.Errorf("%s should not contain template syntax", path)
	}
}

func validateContextSkillIndex(t *testing.T, path, skillsDir, commandSlug string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("missing context file %s: %v", path, err)
	}
	text := string(content)
	if !strings.Contains(text, "## Skills") {
		t.Errorf("%s should contain heading %q", path, "## Skills")
	}
	for _, name := range expectedSkills {
		entry := skillsDir + "/" + commandSlug + "-" + name + "/SKILL.md"
		if !strings.Contains(text, entry) {
			t.Errorf("%s should contain skill index entry %q", path, entry)
		}
	}
}

func validateSkills(t *testing.T, skillsDir, commandSlug string) {
	t.Helper()
	for _, name := range expectedSkills {
		path := filepath.Join(skillsDir, commandSlug+"-"+name, "SKILL.md")
		content, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("missing %s: %v", path, err)
			continue
		}
		text := string(content)
		if !strings.HasPrefix(text, "---\n") {
			t.Errorf("%s should have frontmatter", path)
		}
		if !strings.Contains(text, "name: "+commandSlug+"-"+name) {
			t.Errorf("%s should contain skill name in frontmatter", path)
		}
		if strings.Contains(text, "{{") {
			t.Errorf("%s should not contain template syntax", path)
		}
	}
}
