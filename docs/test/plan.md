<!-- SPDX-FileCopyrightText: The RamenDR authors -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

# `odf dr` test plan

## Table of contents

1. [Overview](#overview)
1. [Environment](#environment)
   1. [Clusters](#clusters)
   1. [Software](#software)
   1. [Storage](#storage)
1. [Test cases](#test-cases)
   1. [CLI basics](#cli-basics)
      1. [Version (upstream only)](#version-upstream-only)
      1. [Help](#help)
      1. [Subcommand help](#subcommand-help)
      1. [Missing required flags](#missing-required-flags)
      1. [Missing config file](#missing-config-file)
   1. [Init command](#init-command)
      1. [Create default config](#create-default-config)
      1. [Create named config](#create-named-config)
      1. [Create config from envfile (upstream only)](#create-config-from-envfile-upstream-only)
      1. [Config already exists](#config-already-exists)
   1. [Validate clusters command](#validate-clusters-command)
      1. [Validate fully configured clusters](#validate-fully-configured-clusters)
      1. [Validate clusters ramen not deployed](#validate-clusters-ramen-not-deployed)
      1. [Validate clusters ramen not configured](#validate-clusters-ramen-not-configured)
      1. [Validate clusters S3 endpoint unreachable](#validate-clusters-s3-endpoint-unreachable)
      1. [Validate clusters missing S3 secret](#validate-clusters-missing-s3-secret)
      1. [Validate clusters ramen deployment is down](#validate-clusters-ramen-deployment-is-down)
   1. [Validate application command](#validate-application-command)
      1. [Validate healthy application](#validate-healthy-application)
      1. [Validate application during failover](#validate-application-during-failover)
      1. [Validate degraded application](#validate-degraded-application)
      1. [Validate application missing S3 secret](#validate-application-missing-s3-secret)
      1. [Validate application missing required flags](#validate-application-missing-required-flags)
      1. [Validate nonexistent application](#validate-nonexistent-application)
   1. [Gather command](#gather-command)
      1. [Gather application data](#gather-application-data)
      1. [Gather degraded application](#gather-degraded-application)
      1. [Gather application missing required flags](#gather-application-missing-required-flags)
      1. [Gather nonexistent application](#gather-nonexistent-application)
   1. [Test command](#test-command)
      1. [Default application (appset + rbd)](#default-application-appset--rbd)
      1. [Subscription-based application](#subscription-based-application)
      1. [Discovered application](#discovered-application)
      1. [Multiple applications](#multiple-applications)
      1. [Failing test](#failing-test)
      1. [Cancelling a test](#cancelling-a-test)
      1. [Cleaning up after tests](#cleaning-up-after-tests)

## Overview

This test plan covers functional validation of the `odf dr` command in a
real ODF + Ramen Regional DR environment. The `odf dr` command is delivered
as part of the odf-cli project and is based on the upstream ramenctl project.
The goal is to verify that all commands produce correct output, detect real
problems, and generate usable reports.

Each test case includes:
- **Preconditions**: what must be true before the test
- **Steps**: exact user actions
- **Expected result**: what the user should observe
- **Automation**: (optional) notes about automating this test

Some features are available only in the upstream `ramenctl` tool and are not
included in `odf dr`. These tests are marked with **(upstream only)** and use
the `ramenctl` command.

## Environment

### Clusters

Regional DR environment with 3 clusters:

| Cluster | Description |
|---------|-------------|
| Hub | ACM hub cluster |
| c1 | Managed cluster |
| c2 | Managed cluster |

Different tests require different cluster configuration levels. Each test
case specifies its own preconditions. The main levels are:

| Level | Description |
|-------|-------------|
| Base clusters | Hub with ACM/OCM, managed clusters joined, no Ramen or DR configuration |
| Ramen deployed | Ramen hub and DR cluster operators deployed, not configured |
| DR configured | DRPolicy, DRClusters, S3 profiles, and storage replication configured |
| Application protected | A DR-protected application deployed and in a stable state |
| Application in-progress | A DR action (failover or relocate) is in progress; the application is moving between clusters |
| Application degraded | Application is protected but a component is unhealthy (e.g., rbd-mirror daemon is down) |

Some tests (e.g., validate clusters on unconfigured clusters) specifically
require clusters that are NOT fully configured, to verify that `odf dr`
correctly detects and reports missing components. Other tests require the
application to be in a transient or degraded state to verify that `odf dr`
correctly reports problems during DR operations or when replication
components are unhealthy.

### Software

| Component | Notes |
|-----------|-------|
| odf-cli | Built from the version under test, provides the `odf dr` command |
| ramenctl | For testing upstream only features |
| oc | For verifying cluster state before and after tests (not a dependency of odf dr) |
| yq | For inspecting YAML reports |

### Storage

ODF creates storage classes during installation. The default names used by
`odf dr init` may differ from the actual names in your clusters. Check the
available storage classes and update the `pvcSpecs` section in the
configuration file accordingly:

```bash
oc get storageclass
```

Common storage types used for DR testing:

| Type | Access Mode | Example Storage Class |
|------|-------------|-----------------------|
| RBD | ReadWriteOnce | `ocs-storagecluster-ceph-rbd` |
| CephFS | ReadWriteMany | `ocs-storagecluster-cephfs` |

## Test cases

### CLI basics

#### Version (upstream only)

**Steps:**

```bash
ramenctl --version
```

**Expected result:** Prints the version string (e.g., `v0.17.0`) and exits
with code 0.

#### Help

**Steps:**

```bash
odf dr --help
```

**Expected result:** Shows usage with available commands: `init`, `validate`,
`test`. Exits with code 0.

#### Subcommand help

**Steps:**

```bash
odf dr validate --help
odf dr test --help
```

**Expected result:** Each shows its subcommands and flags. Exits with code 0.

#### Missing required flags

**Steps:**

```bash
odf dr validate clusters
```

**Expected result:** Error message indicating `--output` flag is required.
Exits with non-zero code.

#### Missing config file

**Steps:**

```bash
odf dr validate clusters -c nonexistent.yaml -o out/invalid-config
```

**Expected result:** Error message about missing config file. Exits with
non-zero code.


### Init command

#### Create default config

**Preconditions:** No `config.yaml` in current directory.

**Steps:**

```bash
odf dr init
```

**Expected result:**
- Prints `Created config file "config.yaml" - please modify for your clusters`.
- File `config.yaml` is created with documented sections: `clusters`, `repo`,
  `drPolicy`, `clusterSet`, `pvcSpecs`, `deployers`, `tests`.
- File contains helpful comments explaining each section.
- File includes notes about options that user may need to modify

#### Create named config

**Preconditions:** No `myenv.yaml` in current directory.

**Steps:**

```bash
odf dr init --config myenv.yaml
```

**Expected result:**
- Prints `Created config file "myenv.yaml" - please modify for your clusters`.
- File `myenv.yaml` is created.

#### Create config from envfile (upstream only)

**Preconditions:** A ramen testing environment file exists.

**Steps:**

```bash
ramenctl init --envfile ../ramen/test/envs/regional-dr.yaml
```

**Expected result:**
- Prints the envfile path and success message.
- `config.yaml` is populated with cluster names, storage class names and distro relevant to ramen testing environment.

#### Config already exists

**Preconditions:** `config.yaml` already exists.

**Steps:**

```bash
odf dr init
```

**Expected result:** Error indicating the file already exists. Does not
overwrite the existing file.


### Validate clusters command

**Preconditions for all tests in this section:**
- `config.yaml` is configured with correct kubeconfig paths.

#### Validate fully configured clusters

**Preconditions:**
- Cluster level: DR configured.

**Steps:**

```bash
odf dr validate clusters -o out/validate-clusters
```

**Expected result:**
- Console output shows progress with checkmarks for each cluster and S3
  profile.
- Final line: `Validation completed (N ok, 0 stale, 0 problem)`.
- Exits with code 0.

**Output files:**

```bash
ls out/validate-clusters/
```

Expected directory contents:
- `validate-clusters.yaml` - machine-readable report
- `validate-clusters.log` - detailed log
- `validate-clusters.data/` - gathered resources
- `validate-clusters.html` - human-readable HTML report

**YAML report content:**

```bash
yq < out/validate-clusters/validate-clusters.yaml
```

The YAML report contains common fields and command-specific data.

**Common fields** (present in all reports):

- `name` - command name (e.g., `validate-clusters`)
- `host` - OS, architecture, and CPU count of the machine running `odf dr`
- `build` - version and commit (upstream only, omitted in `odf dr`)
- `created` - timestamp when the report was created
- `config` - the configuration used (clusters, namespaces, drPolicy, etc.)
- `namespaces` - namespaces gathered during the command
- `status` - overall result (`passed` or `failed`)
- `duration` - total duration in seconds
- `steps` - list of steps with individual status and duration
- `summary` - counts of ok, stale, and problem validations

**Command-specific field** (`clustersStatus`):

The `clustersStatus` section contains validations for the hub, managed
clusters, and S3 profiles. Every validated item shows `state: ok ✅` on a
healthy cluster. The report validates the following:

**For each managed cluster:**

- Ramen configmap
  - Exists (not deleted)
  - Controller type is `dr-cluster`
  - S3 store profiles are configured, and for each profile:
    - Bucket name
    - CA certificate (if configured)
    - S3 endpoint URL
    - Region
    - Secret: exists, access key ID (fingerprint), secret key (fingerprint),
      name, namespace
- Ramen deployment
  - Exists (not deleted)
  - Conditions: `Available`, `Progressing`
  - Replica count

**For the hub cluster:**

- DRClusters: exist, and for each cluster:
  - Conditions: `Fenced`, `Clean`, `Validated`
  - Phase is `Available`
- DRPolicies: exist, and for each policy:
  - Condition: `Validated`
  - Associated DR clusters
  - Scheduling interval
- Ramen configmap (same checks as managed clusters, controller type is
  `dr-hub`)
- Ramen deployment (same checks as managed clusters)

**S3 profiles:**

- For each profile: accessible from the machine running `odf dr`

**Gathered data structure:**

```bash
tree out/validate-clusters/validate-clusters.data/
```

One directory per cluster. Each contains `cluster/` for cluster-scoped
resources and `namespaces/` for namespaced resources. The general structure
is:

```
<cluster-name>/
├── cluster/
│   └── <group>/
│       └── <type>/
│           └── <name>.yaml
└── namespaces/
    └── <namespace>/
        ├── <group>/
        │   └── <type>/
        │       └── <name>.yaml
        └── pods/
            ├── <pod-name>.yaml
            └── <pod-name>/
                └── <container>/
                    ├── current.log
                    └── previous.log
```

For the validate clusters command:

- **Cluster-scoped resources:** all cluster-scoped resources are gathered
- **Namespaces:** `openshift-operators` and `openshift-dr-system` on all
  clusters

**HTML report:**

Open `out/validate-clusters/validate-clusters.html` in a browser.

- Shows the same information in the yaml in a more compact and easier to
  read form.
- All items show green/ok status.

#### Validate clusters ramen not deployed

**Preconditions:**
- Cluster level: Base clusters (ACM/OCM configured, managed clusters joined).
- Ramen operators are NOT deployed on any cluster.

**Steps:**

```bash
odf dr validate clusters -o out/no-ramen
```

**Expected result:**
- Command completes without crashing.
- Report shows problems for missing ramen deployments and configmaps on all
  clusters.
- Problem count > 0 in the summary.

#### Validate clusters ramen not configured

**Preconditions:**
- Cluster level: Ramen deployed.
- No DRPolicy, DRClusters, or S3 profiles configured on the hub.
- No S3 store profiles in ramen configmaps on managed clusters.

**Steps:**

```bash
odf dr validate clusters -o out/no-config
```

**Expected result:**
- Report shows ramen deployments as ok.
- Report shows problems for missing DRPolicy, DRClusters on the hub.
- Report shows problems for missing or empty S3 store profiles.

#### Validate clusters S3 endpoint unreachable

**Preconditions:**
- Cluster level: DR configured.
- Scale down the S3 endpoint controller (minio in upstream).

**Steps:**

```bash
odf dr validate clusters -o out/s3-down
```

**Expected result:**
- S3 profile check reports a problem.
- The report shows the S3 profile as not accessible.

**Cleanup:** Restore the S3 endpoint to normal operation.

#### Validate clusters missing S3 secret

**Preconditions:**
- Cluster level: DR configured.
- Delete or rename the S3 secret in one managed cluster's openshift-dr-system
  namespace.

**Steps:**

```bash
odf dr validate clusters -o out/secret-problem
```

**Expected result:**
- Report shows a problem for the affected S3 profile's secret.

**Cleanup:** Restore the secret.

#### Validate clusters ramen deployment is down

**Preconditions:**
- Cluster level: DR configured.
- Scale down the ramen-dr-cluster-operator deployment on one managed cluster:

```bash
oc scale deployment ramen-dr-cluster-operator -n openshift-dr-system --replicas=0 --context c1
```

**Steps:**

```bash
odf dr validate clusters -o out/operator-down
```

**Expected result:**
- Console output completes but final summary shows problem count > 0.
- `validate-clusters.yaml` report shows a non-ok state for the affected
  cluster's deployment.

**Cleanup:** Restore the operator:

```bash
oc scale deployment ramen-dr-cluster-operator -n openshift-dr-system --replicas=1 --context c1
```

### Validate application command

#### Validate healthy application

**Preconditions:**
- Cluster level: Application protected.
- The application is in a healthy stable state (progression `Completed`).

**Steps:**

```bash
oc get drpc -A --context hub
odf dr validate application --name <drpc-name> --namespace <namespace> -o out/validate-app
```

**Expected result:**
- Console output shows progress: inspected application, gathered data from
  each cluster, inspected S3 profiles.
- Final line: `Validation completed (N ok, 0 stale, 0 problem)`.
- Exits with code 0.

**YAML report content:**

```bash
yq < out/validate-app/validate-application.yaml
```

The YAML report contains common fields and command-specific data.

**Common fields** (present in all reports):

- `name` - command name (e.g., `validate-application`)
- `host` - OS, architecture, and CPU count of the machine running `odf dr`
- `build` - version and commit (upstream only, omitted in `odf dr`)
- `created` - timestamp when the report was created
- `config` - the configuration used (clusters, namespaces, drPolicy, etc.)
- `namespaces` - namespaces gathered during the command
- `application` - the application name and namespace
- `status` - overall result (`passed` or `failed`)
- `duration` - total duration in seconds
- `steps` - list of steps with individual status and duration
- `summary` - counts of ok, stale, and problem validations

**Command-specific field** (`applicationStatus`):

The `applicationStatus` section contains validations for the hub, primary
cluster, secondary cluster, and S3 profiles. Every validated item shows
`state: ok ✅` on a healthy application. The report validates the
following:

**Hub (DRPC):**

- Exists (not deleted)
- DR policy reference
- Action (e.g., `Failover`, `Relocate`)
- Phase (must be the stable phase for the current action)
- Progression (must be `Completed`)
- Conditions

**Primary cluster (VRG):**

- Cluster name
- VRG exists (not deleted)
- VRG state (must be `Primary`)
- VRG conditions (excluding `DataProtected` and stale conditions)
- Protected PVCs, and for each PVC:
  - Exists (not deleted)
  - Phase (must be `Bound`)
  - Replication type (`volrep` or `volsync`)
  - Conditions
- PVC groups (consistency groups, if configured)

**Secondary cluster (VRG):**

- Cluster name
- VRG exists (not deleted)
- VRG state (must be `Secondary`)
- VRG conditions (excluding unused and `DataProtected` conditions)
- Protected PVCs are not validated on the secondary cluster

**S3 profiles:**

- For each profile: application data gathered successfully from S3

**Gathered data structure:**

```bash
tree out/validate-app/validate-application.data/
```

One directory per cluster plus an `s3/` directory. Each cluster directory
has the same structure as validate clusters. The general structure is:

```
<cluster-name>/
├── cluster/
│   └── <group>/
│       └── <type>/
│           └── <name>.yaml
└── namespaces/
    └── <namespace>/
        ├── <group>/
        │   └── <type>/
        │       └── <name>.yaml
        └── pods/
            ├── <pod-name>.yaml
            └── <pod-name>/
                └── <container>/
                    ├── current.log
                    └── previous.log
s3/
└── <profile-name>/
    └── <s3-prefix>/
        └── ...
```

For the validate application command:

- **Cluster-scoped resources:** some related cluster scope resources are
  gathered (PersistentVolume, StorageClass).
- **Namespaces:** application namespaces on all clusters are gathered, depending
  on the type of application.
- **S3 data:** application-specific S3 objects from each profile. Content
  depends on the application type.

#### Validate application during failover

**Preconditions:**
- Cluster level: Application in-progress.
- Application is during failover; run the validate command about one minute
  after starting the failover.

**Steps:**

```bash
oc get drpc -A --context hub
odf dr validate application --name <drpc-name> --namespace <namespace> -o out/app-failover
```

**Expected result:**
- Report shows problems in DRPC progression (not `Completed`).
- VRG conditions may show issues.
- Problem count > 0 in the summary.

#### Validate degraded application

**Preconditions:**
- Cluster level: DR configured.
- Scale down the rbd-mirror daemon on the secondary cluster before
  protecting the application. Then deploy and protect the application. The
  protection will get stuck because replication cannot progress.

**Steps:**

```bash
oc get drpc -A --context hub
odf dr validate application --name <drpc-name> --namespace <namespace> -o out/app-degraded
```

**Expected result:**
- Report shows problems in protected PVC conditions (protection condition
  is not true).
- Problem count > 0 in the summary.

**Cleanup:** Restore the rbd-mirror daemon and wait for the application to
become fully protected.

#### Validate application missing S3 secret

**Preconditions:**
- Cluster level: Application protected.
- Delete or rename the S3 secret on one managed cluster.

**Steps:**

```bash
oc get drpc -A --context hub
odf dr validate application --name <drpc-name> --namespace <namespace> -o out/app-s3-problem
```

**Expected result:**
- Gathering S3 application data from the affected profile fails.
- Report shows a problem for the affected S3 profile's `gathered` field.
- Problem count > 0 in the summary.

**Cleanup:** Restore the S3 secret.

#### Validate application missing required flags

**Steps:**

```bash
odf dr validate application -o out/validate-app
```

**Expected result:** Error indicating `--name` and `--namespace` flags are
required.

#### Validate nonexistent application

**Steps:**

```bash
odf dr validate application --name nonexistent --namespace default -o out/app-noexist
```

**Expected result:** Error indicating the application was not found. Exits
with non-zero code.

### Gather command

#### Gather application data

**Preconditions:**
- Cluster level: Application protected.

**Steps:**

```bash
oc get drpc -A --context hub
odf dr gather application --name <drpc-name> --namespace <namespace> -o out/gather-app
```

**Expected result:**
- Console shows progress: inspected application, gathered data from each
  cluster and S3 profiles.
- Prints `Gather completed`.
- Exits with code 0.
- Output directory contains:
  - `gather-application.yaml` - report
  - `gather-application.log` - detailed log
  - `gather-application.data/` - gathered resources
- `gather-application.data/` contains directories for hub, each managed
  cluster, and `s3/`.
- Each cluster directory contains the application namespace and
  `openshift-operators`/`openshift-dr-system`. Ramen operator logs are
  gathered in `<namespace>/pods/<operator-pod>/manager/`.
- `s3/` contains one directory per S3 profile with the application's S3
  data.

#### Gather degraded application

**Preconditions:**
- Cluster level: DR configured.
- Scale down the rbd-mirror daemon on the secondary cluster before
  protecting the application. Then deploy and protect the application. The
  protection will get stuck because replication cannot progress.

**Steps:**

```bash
oc get drpc -A --context hub
odf dr gather application --name <drpc-name> --namespace <namespace> -o out/gather-degraded
```

**Expected result:**
- Gather completes successfully.
- Gathered data includes ramen operator logs that contain errors related
  to the replication failure.

**Cleanup:** Restore the rbd-mirror daemon and wait for the application to
become fully protected.

#### Gather application missing required flags

**Steps:**

```bash
odf dr gather application -o out/gather-app
```

**Expected result:** Error indicating `--name` and `--namespace` flags are
required.

#### Gather nonexistent application

**Steps:**

```bash
odf dr gather application --name nonexistent --namespace default -o out/gather-noexist
```

**Expected result:** Error indicating the application was not found.


### Test command

**Preconditions for all tests in this section:**
- Cluster level: DR configured.
- `config.yaml` configured with correct kubeconfigs, drPolicy, clusterSet,
  storage classes, and test definitions.

> **Important:** Every test must run both `test run` and `test clean`.
> Skipping `test clean` may leave resources in the clusters and cause false
> results for subsequent tests.

#### Default application (appset + rbd)

**Steps:**

```bash
odf dr test run -o out/test-rbd
odf dr test clean -o out/test-rbd
```

**Expected result:**
- Console shows progress through all DR phases: deployed, protected, failed
  over, relocated, unprotected, undeployed.
- Final line: `passed (1 passed, 0 failed, 0 skipped)`.
- Exits with code 0.
- Typical duration: 10-15 minutes depending on environment.

**Output files:**

```bash
ls out/test-rbd/
```

Expected directory contents:
- `test-run.yaml` - machine-readable test report
- `test-run.log` - detailed log
- `test-clean.yaml` - clean report
- `test-clean.log` - clean log
- No `test-run.data/` directory (data is only gathered on failure)

**YAML report:**

```bash
yq '.status' < out/test-rbd/test-run.yaml
```

Expected: `passed`

```bash
yq '.steps[-1].items.items[].name' < out/test-rbd/test-run.yaml
```

Expected: lists the DR phases: `deploy`, `protect`, `failover`,
`relocate`, `unprotect`, `undeploy`.

```bash
yq '.summary' < out/test-rbd/test-run.yaml
```

Expected: `passed: 1, failed: 0, skipped: 0, canceled: 0`.

#### Subscription-based application

**Preconditions:** Config file has a test using `subscr` deployer.

**Steps:**

```bash
odf dr test run -o out/test-subscr
odf dr test clean -o out/test-subscr
```

**Expected result:** Test passes using Subscription deployment type.

#### Discovered application

**Preconditions:** Config file has a test using `disapp` deployer.

**Steps:**

```bash
odf dr test run -o out/test-disapp
odf dr test clean -o out/test-disapp
```

**Expected result:** Test passes using discovered application deployment type.

#### Multiple applications

**Preconditions:** Config file has multiple test entries (e.g., appset+rbd
and subscr+rbd).

**Steps:**

```bash
odf dr test run -o out/test-multi
odf dr test clean -o out/test-multi
```

**Expected result:**
- Both tests run (may run in parallel).
- Final line: `passed (2 passed, 0 failed, 0 skipped)`.

#### Failing test

**Preconditions:**
- Config ready. Plan to break rbd-mirror after the `protected` step.

**Steps:**

```bash
odf dr test run -o out/test-fail
# After "protected" step appears, scale down rbd-mirror on c1:
#   oc scale deployment rbd-mirror -n ... --replicas=0 --context c1
```

**Expected result:**
- Console shows `❌` for the failing step.
- Command gathers data from all clusters and S3.
- Final line: `failed (0 passed, 1 failed, 0 skipped)`.
- Exits with non-zero code.

**Gathered data on failure:**

```bash
ls out/test-fail/
```

Expected:
- `test-run.yaml` - report with `status: failed`
- `test-run.log` - detailed log
- `test-run.data/` - gathered resources from all clusters and S3

```bash
ls out/test-fail/test-run.data/
```

Expected: directories for hub, both managed clusters, and `s3/`
with data for each S3 profile.

**YAML report shows failure details:**

```bash
yq '.status' < out/test-fail/test-run.yaml
yq '.steps[-1].items.items[].status' < out/test-fail/test-run.yaml
```

Overall status is `failed`. Individual step statuses show which step
failed.

**Cleanup:** Restore rbd-mirror first, then clean up test resources.
Cleaning up before restoring rbd-mirror may fail or time out and leave
leftovers in the clusters.

```bash
# Restore rbd-mirror on c1:
#   oc scale deployment rbd-mirror -n ... --replicas=1 --context c1
odf dr test clean -o out/test-fail
```

#### Cancelling a test

**Steps:**

```bash
odf dr test run -o out/test-cancel
# Press Ctrl+C while a DR step is in progress
```

**Expected result:**
- Command stops gracefully after completing or timing out the current step.
- `test-run.yaml` is written with partial results.
- Does NOT gather data for incomplete tests.

**Cleanup:**

```bash
odf dr test clean -o out/test-cancel
```

#### Cleaning up after tests

**Preconditions:** A test was run (passed or failed) and resources remain.

**Steps:**

```bash
odf dr test run -o out/test-rbd
odf dr test clean -o out/test-rbd
oc get drpc -A --context hub | grep test-
oc get ns --context c1 | grep test-
oc get ns --context c2 | grep test-
oc get channel https-github-com-ramendr-ocm-ramen-samples-git -n test-gitops --context hub
oc get ns test-gitops --context hub
```

**Expected result:**
- Console shows applications being unprotected and undeployed.
- Environment cleaned.
- Final line: `passed (1 passed, 0 failed, 0 skipped)`.
- No test-related DRPC resources or namespaces remain in any cluster.
- The channel `https-github-com-ramendr-ocm-ramen-samples-git` and
  namespace `test-gitops` are deleted from the hub.

