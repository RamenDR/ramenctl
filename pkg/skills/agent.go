// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package skills

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Supported agent tool names.
const (
	AgentCursor  = "cursor"
	AgentClaude  = "claude"
	AgentBob     = "bob"
	AgentGeneric = "generic"
)

type agent struct {
	Name        string
	DisplayName string
	SkillsDir   string
	ContextFile string
	Hint        string
}

var agents = map[string]agent{
	AgentCursor: {
		Name:        AgentCursor,
		DisplayName: "Cursor",
		SkillsDir:   ".cursor/skills",
		ContextFile: ".cursor/rules/ramenctl.mdc",
	},
	AgentClaude: {
		Name:        AgentClaude,
		DisplayName: "Claude Code",
		SkillsDir:   ".claude/skills",
		ContextFile: "CLAUDE.md",
	},
	AgentBob: {
		Name:        AgentBob,
		DisplayName: "Bob",
		SkillsDir:   ".bob/skills",
		ContextFile: ".bob/rules/ramenctl.md",
		Hint:        `Use "/mode advanced" in Bob to enable skills`,
	},
	AgentGeneric: {
		Name:        AgentGeneric,
		SkillsDir:   ".agents/skills",
		ContextFile: "AGENTS.md",
		Hint:        "Instruct your agent to read AGENTS.md",
	},
}

// Agents returns sorted list of supported agent names.
func Agents() []string {
	return slices.Sorted(maps.Keys(agents))
}

// ValidateAgent returns an error if agent is not a supported agent tool name.
func ValidateAgent(agent string) error {
	if _, ok := agents[agent]; !ok {
		return fmt.Errorf(
			"unknown agent tool %q (choose from: %s)",
			agent, strings.Join(Agents(), ", "),
		)
	}
	return nil
}
