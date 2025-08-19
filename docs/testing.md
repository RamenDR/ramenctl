# Testing disaster recovery with ramenctl

## Table of contents

1. [Issue](#issue)
1. [Environment](#environment)
1. [Resolution](#resolution)
   1. [Preparing a configuration file](#preparing-a-configuration-file)
      1. [Configure clusters](#configure-clusters)
      1. [Configure drPolicy](#configure-drpolicy)
      1. [Configure clusterSet](#configure-clusterset)
      1. [Configure pvcSpecs](#configure-pvcspecs)
      1. [Configure tests](#configure-tests)
   1. [Running a test](#running-a-test)
      1. [The test flow](#the-test-flow)
      1. [The test report](#the-test-report)
      1. [The test-run.yaml](#the-test-runyaml)
      1. [The test-run.log](#the-test-runlog)
   1. [Cleaning up](#cleaning-up)
      1. [The clean flow](#the-clean-flow)
   1. [Failed tests](#failed-tests)
      1. [Inspecting gathered data](#inspecting-gathered-data)
   1. [Canceling tests](#canceling-tests)

## Issue

How to test if disaster recovery works in my clusters? Deploying and configuring
clusters for disaster recovery is complicated. The system has many moving parts
and many things can go wrong.

The best way to verify that the system is configured correctly is to deploy a
simple application and test real disaster recovery flow. The *ramenctl* command
makes this task easy.

> [!TIP]
> When working with ODF clusters, use the `odf dr` command instead of `ramenctl`.

## Environment

For disaster recovery you must have a hub cluster and 2 managed clusters,
configured for Regional DR. This document assumes using ODF clusters.

> [!NOTE]
> The `ramenctl` tool is not compatible yet with metro DR.

## Resolution

This section describes how to prepare a configuration file for your ODF clusters
and run a disaster recovery test.

### Preparing a configuration file

`ramenctl` uses a configuration file to access the clusters and the related
resources needed for testing. To create the configuration file run:

```bash
$ ramenctl init

✅ Created config file "config.yaml" - please modify for your clusters
```

The command creates the file `config.yaml` in the current directory. We need to
edit the file to adapt it to our clusters.

#### Configure clusters

Edit the `clusters` section and update the kubeconfig to point to the kubeconfig
files for your clusters:

```yaml
clusters:
  hub:
    kubeconfig: mykubeconfigs/hub
  c1:
    kubeconfig: mykubeconfigs/primary-cluster
  c2:
    kubeconfig: mykubeconfigs/secondary-cluster
```

#### Configure drPolicy

Edit the `drPolicy` section to match your DR configuration

```yaml
drPolicy: drpolicy-1m
```

> [!TIP]
> For quicker test, use a policy with 1 minute internal.

#### Configure clusterSet

Edit `clusterSet` to match your ACM configuration:

```yaml
clusterSet: submariner
```

#### Configure pvcSpecs

Edit the `pvcSpecs` section to use the right storage class names for your
clusters:

```yaml
pvcSpecs:
- name: rbd
  storageClassName: ocs-storagecluster-ceph-rbd
  accessModes: ReadWriteOnce
- name: cephfs
  storageClassName: ocs-storagecluster-cephfs
  accessModes: ReadWriteMany
```

> [!TIP]
> You can add more pvcSpecs for testing other storage classes as needed.
> Modify the test to refer to your own pvcSpec names.

#### Configuring tests

The default tests use a busybox deployment with one *PVC* using *rbd* storage
class, deployed via *ApplicationSet*. You can modify the test to use your
preferred deployment and storage, and add more tests as needed.

The name of the deployer should match the deployer name in the deployers
section.

The available options are:

- workloads
  - `deploy`: busybox deployment with one PVC
- deployers
  - `appset`: ACM managed application deployed using *ApplicationSet*.
  - `subscr`: ACM managed application deployed using *Subscription*.
  - `disapp`: ACM discovered application deployed by the test command.`
- pvcSpecs:
  - `rbd`: *Ceph* *RBD* storage
  - `cephfs`: *CephFS* storage

### Running a test

This section shows how to run a test and inspect the test report.

We will run a test and store the test report in the directory `ramenctl-test`:

```console
$ ramenctl test run -o ramenctl-test
⭐ Using report "ramenctl-test"
⭐ Using config "config.yaml"

🔎 Validate config ...
   ✅ Config validated

🔎 Setup environment ...
   ✅ Environment setup

🔎 Run tests ...
   ✅ Application "appset-deploy-rbd" deployed
   ✅ Application "appset-deploy-rbd" protected
   ✅ Application "appset-deploy-rbd" failed over
   ✅ Application "appset-deploy-rbd" relocated
   ✅ Application "appset-deploy-rbd" unprotected
   ✅ Application "appset-deploy-rbd" undeployed

✅ passed (1 passed, 0 failed, 0 skipped)
```

To clean up after the test use the [clean](#cleaning-up) command.

#### The test flow

When running the run command *ramenctl* prepares the clusters for the tests and
run all tests specified in the configuration file.

Preparing the clusters includes:
1. Create a namespace "test-gitops" on the hub cluster and add a channel for
   https://github.com/RamenDR/ocm-ramen-samples.

For every test specified in the configuration file perform the following steps:
1. **deploy**: Deploy the application in namespace
   "test-{deployer}-{workload}-{pvcSpec}" in the primary cluster.
1. **protect**: Create a *drpc* resource for the application and wait until the
   application is protected.
1. **failover**: Fail over the application to to the secondary cluster and wait
   until the application is protected.
1. **relocate**: Relocate the application back to the primary cluster and wait
   until the application is protected.
1. **unprotect**: Delete the *drpc* resource for the application and wait until
   the *drpc* is deleted.
1. **undeploy**: Undeploy the application from the managed clusters and wait
   until the application is deleted.

#### The test report

The command stores `test-run.yaml` and `test-run.log` in the specified output
directory:

```console
$ tree ramenctl-test
ramenctl-test
├── test-run.log
└── test-run.yaml
```

> [!IMPORTANT]
> When reporting DR related issues, please create an archive with the
> output directory and upload it to the issue tracker.

#### The test-run.yaml

The test-run.yaml is a machine and human readable description of the the
test run:

```yaml
build:
  commit: de6950adf6666a4ff0886e29e99a998615142fe5
  version: v0.4.0-7-gde6950a
config:
  channel:
    name: https-github-com-ramendr-ocm-ramen-samples-git
    namespace: test-gitops
  clusterSet: submariner
  clusters:
    c1:
      kubeconfig: mykubeconfigs/primary-cluster
    c2:
      kubeconfig: mykubeconfigs/secondary-cluster
    hub:
      kubeconfig: mykubeconfigs/hub
  distro: ocp
  drPolicy: drpolicy-1m
  namespaces:
    argocdNamespace: openshift-gitops
    ramenDRClusterNamespace: openshift-dr-system
    ramenHubNamespace: openshift-operators
    ramenOpsNamespace: openshift-dr-ops
  pvcSpecs:
  - accessModes: ReadWriteOnce
    name: rbd
    storageClassName: ocs-storagecluster-ceph-rbd
  - accessModes: ReadWriteMany
    name: cephfs
    storageClassName: ocs-storagecluster-cephfs
  repo:
    branch: main
    url: https://github.com/RamenDR/ocm-ramen-samples.git
  tests:
  - deployer: appset
    pvcSpec: rbd
    workload: deploy
created: "2025-04-24T16:33:28.800757+05:30"
duration: 695.608462543
host:
  arch: arm64
  cpus: 12
  os: darwin
name: test-run
status: passed
steps:
- duration: 0.022823334
  name: validate
  status: passed
- duration: 0.009449584
  name: setup
  status: passed
- duration: 695.576189625
  items:
    duration: 695.576118209
    items:
    - duration: 0.013429167
      name: deploy
      status: passed
    - duration: 95.095957834
      name: protect
      status: passed
    - duration: 270.193688542
      name: failover
      status: passed
    - duration: 270.207244208
      name: relocate
      status: passed
    - duration: 60.052227083
      name: unprotect
      status: passed
    - duration: 0.013571375
      name: undeploy
      status: passed
    name: appset-deploy-rbd
    status: passed
  name: tests
  status: passed
summary:
  canceled: 0
  failed: 0
  passed: 1
  skipped: 0
```

You can query it with tools like yq:

```console
$ yq .status < ramenctl-test/test-run.yaml
passed
```

#### The test-run.log

The `test-run.log` contains detailed logs of the test progress.

To extract single test major events use:

```bash
grep -E '(INFO|ERROR).+appset-deploy-rbd' ramenctl-test/test-run.log
```

Example output

```console
2025-03-29T23:56:24.547+0300	INFO	appset-deploy-rbd	deployers/appset.go:23	Deploying applicationset app "test-appset-deploy-rbd/busybox" in cluster "primary-cluster"
2025-03-29T23:56:25.060+0300	INFO	appset-deploy-rbd	deployers/appset.go:41	Workload deployed
2025-03-29T23:56:25.383+0300	INFO	appset-deploy-rbd	dractions/actions.go:51	Protecting workload "test-appset-deploy-rbd/busybox" in cluster "primary-cluster"
2025-03-29T23:59:16.414+0300	INFO	appset-deploy-rbd	dractions/actions.go:93	Workload protected
2025-03-29T23:59:16.892+0300	INFO	appset-deploy-rbd	dractions/actions.go:157	Failing over workload "test-appset-deploy-rbd/busybox" from cluster "primary-cluster" to cluster "secondary-cluster"
2025-03-30T00:05:03.748+0300	INFO	appset-deploy-rbd	dractions/actions.go:165	Workload failed over
2025-03-30T00:05:04.226+0300	INFO	appset-deploy-rbd	dractions/actions.go:190	Relocating workload "test-appset-deploy-rbd/busybox" from cluster "secondary-cluster" to cluster "primary-cluster"
2025-03-30T00:10:50.940+0300	INFO	appset-deploy-rbd	dractions/actions.go:198	Workload relocated
2025-03-30T00:10:51.260+0300	INFO	appset-deploy-rbd	dractions/actions.go:121	Unprotecting workload "test-appset-deploy-rbd/busybox" in cluster "primary-cluster"
2025-03-30T00:11:17.561+0300	INFO	appset-deploy-rbd	dractions/actions.go:136	Workload unprotected
2025-03-30T00:11:17.882+0300	INFO	appset-deploy-rbd	deployers/appset.go:61	Undeploying applicationset app "test-appset-deploy-rbd/busybox" in cluster "primary-cluster"
2025-03-30T00:11:18.379+0300	INFO	appset-deploy-rbd	deployers/appset.go:80	Workload undeployed
```

### Cleaning up

To clean up after a test, removing resources created by the test, run:

```console
$ ramenctl test clean -o ramenctl-test
⭐ Using report "ramenctl-test"
⭐ Using config "config.yaml"

🔎 Validate config ...
   ✅ Config validated

🔎 Clean tests ...
   ✅ Application "appset-deploy-rbd" unprotected
   ✅ Application "appset-deploy-rbd" undeployed

🔎 Clean environment ...
   ✅ Environment cleaned

✅ passed (1 passed, 0 failed, 0 skipped)
```

The clean command adds `test-clean.log` and `test-clean.yaml` to the
output directory:

```console
$ tree ramenctl-test
ramenctl-test
├── test-clean.log
├── test-clean.yaml
├── test-run.log
└── test-run.yaml
```

#### The clean flow

When running the clean command *ramenctl* deletes all the tests applications
specified in the configuration file and cleans up the clusters.

For every test specified in the configuration file perform the following steps:
1. **unprotect**: Delete the *drpc* resource for the application and wait
   until the *drpc* is deleted.
2. **undeploy**: Undeploy the application from the managed clusters and wait
   until the application is deleted.

Cleaning up the clusters includes:
1. Delete the channel and the namespace "test-gitops" on the hub cluster.

### Failed tests

When a test fails, the test command gathers data related to the failed tests in
the output directory. The gathered data can help you or developers to diagnose the
issue.

The following example shows a test run with a failed test, and how to inspect
the failure.

> [!TIP]
> To fail the test I scaled down the rbd-mirror deployment on the primary
> cluster after the application reached the "protected" state.

Running the test:

```console
$ ramenctl test run -o example-failure
⭐ Using report "example-failure"
⭐ Using config "config.yaml"

🔎 Validate config ...
   ✅ Config validated

🔎 Setup environment ...
   ✅ Environment setup

🔎 Run tests ...
   ✅ Application "appset-deploy-rbd" deployed
   ✅ Application "appset-deploy-rbd" protected
   ❌ failed to failover application "appset-deploy-rbd"

🔎 Gather data ...
   ✅ Gathered data from cluster "hub"
   ✅ Gathered data from cluster "secondary-cluster"
   ✅ Gathered data from cluster "primary-cluster"

❌ failed (0 passed, 1 failed, 0 skipped)
```

The command stores gathered data in the `test-run.data` directory:

```console
$ tree -L2 example-failure
example-failure
├── test-run.data
│   ├── hub
│   ├── primary-cluster
│   └── secondary-cluster
├── test-run.log
└── test-run.yaml
```

> [!IMPORTANT]
> When reporting DR related issues, please create an archive with the output
> directory and upload it to the issue tracker.

#### Inspecting gathered data

The command gathers all the namespaces related to the failed test, and related
cluster scope resources such as storage classes and persistent volumes.

```console
$ tree -L3 example-failure/test-run.data
example-failure/test-run.data
├── hub
│   ├── cluster
│   │   └── namespaces
│   └── namespaces
│       ├── openshift-gitops
│       └── openshift-operators
├── primary-cluster
│   ├── cluster
│   │   ├── namespaces
│   │   ├── persistentvolumes
│   │   └── storage.k8s.io
│   └── namespaces
│       ├── openshift-dr-system
│       ├── openshift-gitops
│       ├── openshift-operators
│       └── test-appset-deploy-rbd
└── secondary-cluster
    ├── cluster
    │   ├── namespaces
    │   ├── persistentvolumes
    │   └── storage.k8s.io
    └── namespaces
        ├── openshift-dr-system
        ├── openshift-gitops
        ├── openshift-operators
        └── test-appset-deploy-rbd
```

Change directory into the gather directory to simplify the next steps:

```console
$ example-failure/test-run.data
```

We can start by looking at the *DRPC*:

```console
$ cat hub/namespaces/openshift-gitops/ramendr.openshift.io/drplacementcontrols/appset-deploy-rbd.yaml
...
status:
  actionStartTime: "2025-04-01T18:02:06Z"
  conditions:
  - lastTransitionTime: "2025-04-01T18:02:41Z"
    message: Completed
    observedGeneration: 2
    reason: FailedOver
    status: "True"
    type: Available
  - lastTransitionTime: "2025-04-01T18:02:06Z"
    message: Started failover to cluster "secondary-cluster"
    observedGeneration: 2
    reason: NotStarted
    status: "False"
    type: PeerReady
  - lastTransitionTime: "2025-04-01T18:03:11Z"
    message: VolumeReplicationGroup (test-appset-deploy-rbd/appset-deploy-rbd) on
      cluster secondary-cluster is not reporting any lastGroupSyncTime as primary, retrying
      till status is met
    observedGeneration: 2
    reason: Progressing
    status: "False"
    type: Protected
  lastUpdateTime: "2025-04-01T18:08:11Z"
  observedGeneration: 2
  phase: FailedOver
  preferredDecision:
    clusterName: primary-cluster
    clusterNamespace: primary-cluster
  progression: Cleaning Up
```

We can see that the application is stuck in "Cleaning Up" progression and the
*VRG* in cluster "secondary-cluster" is not reporting `lastGropSyncTime` value.

Looking at the *VRG* in cluster in cluster "secondary-cluster":

```console
$ cat secondary-cluster/namespaces/test-appset-deploy-rbd/ramendr.openshift.io/volumereplicationgroups/appset-deploy-rbd.yaml
...
status:
  conditions:
  - lastTransitionTime: "2025-04-01T18:02:48Z"
    message: PVCs in the VolumeReplicationGroup are ready for use
    observedGeneration: 2
    reason: Ready
    status: "True"
    type: DataReady
  - lastTransitionTime: "2025-04-01T18:02:37Z"
    message: VolumeReplicationGroup is replicating
    observedGeneration: 2
    reason: Replicating
    status: "False"
    type: DataProtected
  - lastTransitionTime: "2025-04-01T18:02:37Z"
    message: Restored 0 volsync PVs/PVCs and 2 volrep PVs/PVCs
    observedGeneration: 2
    reason: Restored
    status: "True"
    type: ClusterDataReady
  - lastTransitionTime: "2025-04-01T18:02:47Z"
    message: Cluster data of all PVs are protected
    observedGeneration: 2
    reason: Uploaded
    status: "True"
    type: ClusterDataProtected
  - lastTransitionTime: "2025-04-01T18:02:37Z"
    message: Kube objects restored
    observedGeneration: 2
    reason: KubeObjectsRestored
    status: "True"
    type: KubeObjectsReady
  kubeObjectProtection: {}
  lastUpdateTime: "2025-04-01T18:12:48Z"
  observedGeneration: 2
  ...
```

We can see that DataProtected is False.

Looking at the *VR* resource in the same namespace:

```console
$ cat secondary-cluster/namespaces/test-appset-deploy-rbd/replication.storage.openshift.io/volumereplications/busybox-pvc.yaml
...
status:
  conditions:
  - lastTransitionTime: "2025-04-01T18:02:48Z"
    message: volume is promoted to primary and replicating to secondary
    observedGeneration: 1
    reason: Promoted
    status: "True"
    type: Completed
  - lastTransitionTime: "2025-04-01T18:02:48Z"
    message: volume is healthy
    observedGeneration: 1
    reason: Healthy
    status: "False"
    type: Degraded
  - lastTransitionTime: "2025-04-01T18:02:47Z"
    message: volume is not resyncing
    observedGeneration: 1
    reason: NotResyncing
    status: "False"
    type: Resyncing
  - lastTransitionTime: "2025-04-01T18:02:47Z"
    message: volume is validated and met all prerequisites
    observedGeneration: 1
    reason: PrerequisiteMet
    status: "True"
    type: Validated
  lastCompletionTime: "2025-04-01T18:12:48Z"
  message: volume is marked primary
  observedGeneration: 1
  state: Primary
```

We see that the VR is primary and replicating to the other cluster.

We can also inspect ramen-dr-cluster-operator logs:

```console
% tree secondary-cluster/namespaces/openshift-dr-system/pods/ramen-dr-cluster-operator-5dd448864d-78x8l/manager/
secondary-cluster/namespaces/openshift-dr-system/pods/ramen-dr-cluster-operator-5dd448864d-78x8l/manager/
├── current.log
└── previous.log
```

In this case the gathered data tells that ramen is not the root cause, and we
need to inspect the storage.

If more information is needed you can use the standard must-gather with ODF or
ACM images to do a full gather.

When we finished to debug the failed test we need to cleanup up:

```console
$ ramenctl test clean -o example-failure
⭐ Using report "example-failure"
⭐ Using config "config.yaml"

🔎 Validate config ...
   ✅ Config validated

🔎 Clean tests ...
   ✅ Application "appset-deploy-rbd" unprotected
   ✅ Application "appset-deploy-rbd" undeployed

🔎 Clean environment ...
   ✅ Environment cleaned

✅ passed (1 passed, 0 failed, 0 skipped)
```

### Canceling tests

The run or clean command may take up to 10 minutes to complete the current test
step. To get all the information about failed tests you should wait until the
command completes and gathers data for failed tests.

You can cancel the command by pressing `Ctrl+C`. This saves the current tests
progress but does not gather data for incomplete tests.
