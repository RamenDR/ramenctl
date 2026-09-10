<!--
SPDX-FileCopyrightText: The RamenDR authors
SPDX-License-Identifier: Apache-2.0
-->

# ramenctl init

The init command crates a configuration file required for other *ramenctl*
commands and installs AI agent skills for other *ramenctl* commands.

```console
% ramenctl init -h
Create configuration file and install AI skills

Usage:
  ramenctl init [flags]

Flags:
  -a, --agent string     AI agent to install skills for (bob, claude, codex, cursor, generic) (default "generic")
      --envfile string   ramen testing environment file
  -h, --help             help for init

Global Flags:
  -c, --config string   configuration file (default "config.yaml")
      --interactive     enable interactive features (default auto)
```

## Getting started

The init command creates a configuration file named "config.yaml" in the current
and installs AI skills in the current directory:

```console
$ ramenctl init
⭐ Using config "config.yaml"

🔎 Initializing ...
   ✅ Created config file "config.yaml" - please modify for your clusters
   ✅ Created skills in ".agents/skills/"
   ✅ Created context file "AGENTS.md"
      Instruct your agent to read AGENTS.md

✅ Init completed
```

> [!IMPORTANT]
> Before using the configuration file you need to edit it to match your clusters
> and storage.

Other *ramenctl* commands use "config.yaml" by default.

### AI skills

`ramenctl init` installs AI agent skills alongside the configuration file in the
current directory. After running `init`, the directory is ready for agentic
usage out of the box.

Use the `--agent` (`-a`) flag to install skills in the format expected by your
AI tool (e.g. `ramenctl init -a cursor`). The default generic format works with
any agent.

Running `init` again is safe. Existing skill files and context files are not
overwritten, preserving any user modifications.

For more details on available skills, supported agents, and output directory
conventions, see [AI Skills](skills.md).

## Creating configuration file for a ramen testing environment

When using a ramen testing environment we can create a configuration file
optimized for the testing environment using the `--envfile` option:

```console
$ ramenctl init --envfile ../ramen/test/envs/regional-dr.yaml
⭐ Using config "config.yaml"
⭐ Using envfile "../ramen/test/envs/regional-dr.yaml"

🔎 Initializing ...
   ✅ Created config file "config.yaml" - please modify for your clusters
   ✅ Created skills in ".agents/skills/"
   ✅ Created context file "AGENTS.md"
      Instruct your agent to read AGENTS.md

✅ Init completed
```

You can edit the configuration file to change the default tests.

## Using multiple configuration files

When working with multiple environments or when you want to run different sets
of tests with the same environment, you can create multiple configuration files
and use them with the `--config` option.

Create a configuration file named "myenv.yaml":

```console
$ ramenctl init --config myenv.yaml
⭐ Using config "myenv.yaml"

🔎 Initializing ...
   ✅ Created config file "myenv.yaml" - please modify for your clusters
   ✅ Created skills in ".agents/skills/"
   ✅ Created context file "AGENTS.md"
      Instruct your agent to read AGENTS.md

✅ Init completed
```

To use the configuration file with other commands, specify it with the
`--config` option:

```console
$ ramenctl test run --config myenv.yaml -o test
⭐ Using report "test"
⭐ Using config "myenv.yaml"
...
```

## Configuring common options

All *ramenctl* commands require the `clusters` and `clusterSet` options. For the
validate and gather commands, these are the only options needed.

> [!TIP]
> You can ask your AI agent to configure the options for your environment using
> the skills installed by `ramenctl init`.

> [!TIP]
> When using a ramen testing environment, the `--envfile` option configures
> everything for you. See
> [Creating configuration file for a ramen testing environment](#creating-configuration-file-for-a-ramen-testing-environment).

### Configuring clusters

Modify the `clusters` section to match your hub and managed clusters kubeconfig
files. Set `passive-hub` kubeconfig for optional passive hub cluster, or leave
it empty if not using passive hub.

```yaml
clusters:
  hub:
    kubeconfig: my-hub.yaml
  passive-hub:
    kubeconfig: ""
  c1:
    kubeconfig: my-c1.yaml
  c2:
    kubeconfig: my-c2.yaml
```

### Configuring clusterSet

The `clusterSet` option specifies the Open Cluster Management
`ManagedClusterSet` that contains the managed clusters.

To find the clusterSet for your managed clusters, run the following command:

```console
$ kubectl get managedclusters --kubeconfig my-hub.yaml -L cluster.open-cluster-management.io/clusterset
NAME            HUB ACCEPTED   MANAGED CLUSTER URLS                 JOINED   AVAILABLE   AGE   CLUSTERSET
my-c1           true           https://api.my-c1.example.com:6443   True     True        16d   dr-clusters
my-c2           true           https://api.my-c2.example.com:6443   True     True        16d   dr-clusters
local-cluster   true           https://api.my-hub.example.com:6443  True     True        16d   default
```

The `CLUSTERSET` column shows which clusterSet each managed cluster belongs to.
Use this value in the configuration file:

```yaml
clusterSet: dr-clusters
```

### Example common configuration

```yaml
## ramenctl configuration file

clusters:
  hub:
    kubeconfig: my-hub.yaml
  passive-hub:
    kubeconfig: ""
  c1:
    kubeconfig: my-c1.yaml
  c2:
    kubeconfig: my-c2.yaml

clusterSet: dr-clusters
```

## Configuration for the test command

The test command requires the [common options](#configuring-common-options) and
additional test options. The following is a sample configuration file showing
the default values. You must modify it to match your clusters and storage.

```yaml
## ramenctl configuration file

## Common options - used by all commands (validate, gather, test).

## Clusters configuration.
# - Modify clusters "kubeconfig" to match your hub and managed clusters
#   kubeconfig files.
# - Modify "passive-hub" kubeconfig for optional passive hub cluster,
#   leave it empty if not using passive hub.
clusters:
  hub:
    kubeconfig: my-hub.yaml
  passive-hub:
    kubeconfig: ""
  c1:
    kubeconfig: my-c1.yaml
  c2:
    kubeconfig: my-c2.yaml

## ClusterSet with the managed clusters.
# - Modify to match your Open Cluster Management configuration.
clusterSet: default

## Test options - used only by the test command.

## Git repository.
# - Modify "url" to use your own Git repository.
# - Modify "branch" to test a different branch.
repo:
  url: https://github.com/RamenDR/ocm-ramen-samples.git
  branch: main

## DRPolicy.
# - Modify to match actual DRPolicy in the hub cluster.
drPolicy: dr-policy-1m

## PVC specifications.
# - Modify items "storageClassName" to match the actual storage classes in the
#   managed clusters.
# - Add new items for testing more storage types.
pvcSpecs:
- name: rbd
  storageClassName: rook-ceph-block
  accessModes: ReadWriteOnce
- name: cephfs
  storageClassName: rook-cephfs-fs1
  accessModes: ReadWriteMany

## Deployer specifications.
# - Modify items "name" and "type" to match your deployer configurations.
# - Add new items for testing more deployers.
# - Available types: appset, subscr, disapp
deployers:
- name: appset
  type: appset
  description: ApplicationSet deployer for ArgoCD
- name: subscr
  type: subscr
  description: Subscription deployer for OCM subscriptions
- name: disapp
  type: disapp
  description: Discovered Application deployer
- name: disapp-recipe
  type: disapp
  recipe:
    type: generate
  description: Discovered Application deployer with recipe
- name: disapp-recipe-check
  type: disapp
  recipe:
    type: generate
    checkHook: true
  description: Discovered Application deployer with recipe using check hook
- name: disapp-recipe-exec
  type: disapp
  recipe:
    type: generate
    execHook: true
  description: Discovered Application deployer with recipe using exec hook
- name: disapp-recipe-check-exec
  type: disapp
  recipe:
    type: generate
    checkHook: true
    execHook: true
  description: Discovered Application deployer with recipe using check and exec hooks

## Test cases.
# - Modify the test for your preferred workload or deployment type.
# - Add new tests for testing more combinations in parallel.
# - Available workloads: deploy.
# - Available deployers: appset, subscr, disapp, disapp-recipe,
#   disapp-recipe-check, disapp-recipe-exec, disapp-recipe-check-exec.
tests:
- workload: deploy
  deployer: appset
  pvcSpec: rbd
```
