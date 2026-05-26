// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package skills

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"
)

//go:embed templates
var skillsFS embed.FS

// Command identifies the CLI command for skill templates.
type Command struct {
	// Name is the display name (e.g., "ramenctl", "odf dr").
	Name string
	// Slug is the directory-safe form (e.g., "ramenctl", "odf-dr").
	Slug string
}

// Skill holds metadata about one embedded skill template.
type Skill struct {
	// Name is the subcommand name (e.g., "init", "validate-clusters").
	Name string
	// Description is a short one-line summary of the skill.
	Description string
}

// skills is the static list of all embedded skills.
var skills = []Skill{
	{Name: "gather-application", Description: "Gather diagnostic data for an application"},
	{Name: "init", Description: "Configure for your clusters"},
	{Name: "test-clean", Description: "Clean up after test runs"},
	{Name: "test-run", Description: "Run disaster recovery flow tests"},
	{Name: "validate-application", Description: "Validate a DR-protected application"},
	{Name: "validate-clusters", Description: "Validate DR cluster configuration"},
}

// skillData is the data passed to skill templates.
type skillData struct {
	Command Command
	Skill   Skill
}

func renderSkill(skill Skill, cmd Command) ([]byte, error) {
	tmplPath := "templates/skills/" + skill.Name + ".tmpl"

	content, err := skillsFS.ReadFile(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read skill template %q: %w", skill.Name, err)
	}

	t, err := template.New(skill.Name).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse skill template %q: %w", skill.Name, err)
	}

	data := skillData{Command: cmd, Skill: skill}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to render skill template %q: %w", skill.Name, err)
	}

	return buf.Bytes(), nil
}

// contextData is the data passed to agent context templates.
type contextData struct {
	Command Command
	Agent   agent
	Skills  []Skill
}

func renderContextFile(cmd Command, ag agent) ([]byte, error) {
	t, err := template.New("").ParseFS(skillsFS, "templates/agents/*.tmpl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse agent templates: %w", err)
	}

	data := contextData{
		Command: cmd,
		Agent:   ag,
		Skills:  skills,
	}

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, ag.Name+".tmpl", data); err != nil {
		return nil, fmt.Errorf("failed to render context template %q: %w", ag.Name, err)
	}

	return buf.Bytes(), nil
}

// directoryName converts a command name to a form suitable for directory names
// ("odf dr" -> "odf-dr").
func directoryName(name string) string {
	return strings.ReplaceAll(name, " ", "-")
}
