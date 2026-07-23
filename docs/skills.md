<!-- SPDX-FileCopyrightText: The RamenDR authors -->

<!-- SPDX-License-Identifier: Apache-2.0 -->

# Using ramenctl with AI agents

*ramenctl* is agentic-ready out of the box. Running `ramenctl init` installs AI
skills that teach your coding assistant how to drive *ramenctl* for disaster
recovery on your clusters — no extra setup needed.

Start with the example session below to see what this feels like in practice.
After that you'll find the list of available skills, where they are installed,
and how to add support for a new agent.

## Example session

The following is a short screenplay-style scene: a human and an agent use the
ramenctl skills. (Monospace block so it reads like a script and renders with a
distinct background on GitHub and most viewers.)

```text
INT. CURSOR CHAT — DAY

USER
        I ran ramenctl init --agent cursor. Help me configure ramenctl
        for my clusters. My kubeconfigs are in ocp/: hub.yaml is the
        hub, c1.yaml and c2.yaml are the managed clusters.

AGENT
        (Edits config.yaml: fills kubeconfigs, reads ClusterClaims on c1
        and c2, matches names on the hub, sets clusterSet.)
        Your config is ready.

USER
        Can you check the clusters?

AGENT
        (Runs ramenctl validate clusters -o ocp/clusters.)
        Validation completed (91 ok, 0 warning, 0 problem).
        (Opens validate-clusters.html in the browser.)

USER
        Validate an application too.

AGENT
        (Lists DRPCs on the hub.)
        I found two applications:
          1. rbd4 (openshift-dr-ops) — FailedOver
          2. appset-deploy-rbd (openshift-gitops) — Deployed
        Which one should I validate?

USER
        rbd4.

AGENT
        (Runs ramenctl validate application --name rbd4 --namespace
        openshift-dr-ops -o ocp/rbd4.)
        All clear. Opening validate-application.html.

USER
        You are a very skillful agent!

AGENT
        Glad it helped.
```

## Available skills

| Skill                           | Description                                        |
| ------------------------------- | -------------------------------------------------- |
| `ramenctl-init`                 | Create a configuration file for your clusters      |
| `ramenctl-validate-clusters`    | Validate disaster recovery cluster configuration   |
| `ramenctl-validate-application` | Validate a DR-protected application                |
| `ramenctl-gather-application`   | Gather diagnostic data for a protected application |
| `ramenctl-test-run`             | Run disaster recovery flow tests                   |
| `ramenctl-test-clean`           | Clean up after test runs                           |

## Where skills are installed

`ramenctl init` installs skills automatically. Use the `--agent` (`-a`) flag to
install in the format expected by your AI tool:

```console
$ ramenctl init -a cursor
```

Supported agents:

| Agent       | Flag        | Skills directory  | Context file                 |
| ----------- | ----------- | ----------------- | ---------------------------- |
| Bob         | `-a bob`    | `.bob/skills/`    | `AGENTS.md`                  |
| Claude Code | `-a claude` | `.claude/skills/` | `CLAUDE.md`                  |
| Codex       | `-a codex`  | `.agents/skills/` | `AGENTS.md`                  |
| Cursor      | `-a cursor` | `.cursor/skills/` | `.cursor/rules/ramenctl.mdc` |
| Generic     | *(default)* | `.agents/skills/` | `AGENTS.md`                  |

> [!TIP]
> - When using the generic format, instruct your AI agent to read `AGENTS.md`
>   for project context and skill discovery.
> - Bob requires advanced mode to discover skills. Use `/mode advanced` in the
>   Bob chat before starting.

See [init](init.md) for more on creating the configuration file.

## Adding a new agent

To add support for a new AI agent:

1. Add a constant (e.g. `AgentMyTool = "my-agent"`) in `pkg/skills/agent.go`.
1. Add an entry to the `agents` map in `pkg/skills/agent.go` with the tool's
   display name, native skills directory, and context file path.
1. Create a context file template `pkg/skills/templates/agents/my-agent.tmpl`.
   The template receives the command name and skill list. Look at existing
   templates for examples.
1. Update the agent table in `docs/skills.md`.
1. Add test cases in `pkg/skills/skills_test.go` for the new agent.
