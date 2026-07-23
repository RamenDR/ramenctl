// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package skills

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ramendr/ramenctl/pkg/console"
)

// Install installs AI skills for the given agent tool. Returns true on success.
func Install(commandName, agentName string) bool {
	agent, ok := agents[agentName]
	if !ok {
		panic(fmt.Sprintf("unknown agent tool %q: call ValidateAgent first", agentName))
	}

	cmd := Command{
		Name: commandName,
		Slug: directoryName(commandName),
	}

	if !installSkills(cmd, agent) {
		return false
	}

	return installContextFile(cmd, agent)
}

func installSkills(cmd Command, agent agent) bool {
	if err := os.MkdirAll(agent.SkillsDir, 0o755); err != nil {
		console.Error("failed to create directory %q: %s", agent.SkillsDir, err)
		return false
	}

	var skipped []string

	for _, skill := range skills {
		skillDir := filepath.Join(agent.SkillsDir, cmd.Slug+"-"+skill.Name)

		if err := os.Mkdir(skillDir, 0o755); err != nil {
			if !errors.Is(err, os.ErrExist) {
				console.Error("failed to create directory %q: %s", skillDir, err)
				return false
			}
			if !isDir(skillDir) {
				console.Error("expected directory but found file %q", skillDir)
				return false
			}
		}

		content, err := renderSkill(skill, cmd)
		if err != nil {
			console.Error("%s", err)
			return false
		}

		skillFile := filepath.Join(skillDir, "SKILL.md")
		if err := writeNewFile(skillFile, content); err != nil {
			if errors.Is(err, os.ErrExist) {
				skipped = append(skipped, cmd.Slug+"-"+skill.Name)
				continue
			}
			console.Error("failed to write %q: %s", skillFile, err)
			return false
		}
	}

	switch {
	case len(skipped) == len(skills):
		console.Warn("Skills already exist in \"%s/\"", agent.SkillsDir)
	case len(skipped) > 0:
		if agent.DisplayName != "" {
			console.Warn("Created skills for %s in \"%s/\"", agent.DisplayName, agent.SkillsDir)
		} else {
			console.Warn("Created skills in \"%s/\"", agent.SkillsDir)
		}
		console.StepHint("Skipped existing skills %s", strings.Join(skipped, ", "))
	default:
		if agent.DisplayName != "" {
			console.Pass("Created skills for %s in \"%s/\"", agent.DisplayName, agent.SkillsDir)
		} else {
			console.Pass("Created skills in \"%s/\"", agent.SkillsDir)
		}
	}

	return true
}

func installContextFile(cmd Command, agent agent) bool {
	content, err := renderContextFile(cmd, agent)
	if err != nil {
		console.Error("%s", err)
		return false
	}

	dir := filepath.Dir(agent.ContextFile)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			console.Error("failed to create directory %q: %s", dir, err)
			return false
		}
	}

	if err := writeNewFile(agent.ContextFile, content); err != nil {
		if errors.Is(err, os.ErrExist) && isRegularFile(agent.ContextFile) {
			console.Warn("Context file %q already exists", agent.ContextFile)
		} else {
			console.Error("failed to write %q: %s", agent.ContextFile, err)
			return false
		}
	} else {
		console.Pass("Created context file %q", agent.ContextFile)
	}

	if agent.Hint != "" {
		console.StepHint("%s", agent.Hint)
	}

	return true
}

func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// writeNewFile creates a file only if it does not exist. Returns os.ErrExist
// if the file already exists, matching the write-once ownership model.
func writeNewFile(path string, content []byte) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return err
	}
	// Closing twice is safe; defer handles the error path, explicit close
	// below catches flush errors on the happy path.
	defer f.Close()

	if _, err := f.Write(content); err != nil {
		return err
	}

	return f.Close()
}
