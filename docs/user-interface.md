<!--
SPDX-FileCopyrightText: The RamenDR authors
SPDX-License-Identifier: Apache-2.0
-->

# ramenctl user interface

The initial version of the tool will include only the `validate clusters` and
`test` troubleshooting commands. We will add more troubleshooting commands in
the next releases. In a future release, we also plan to add management commands.

This document outlines all the planned commands to make sure that we can extend
the tool in the future.

> [!NOTE]
> The commands names are arguments are work in progress and likely to change as
> we go.

## Initial configuration

### init

```console
ramenctl init [--envfile PATH] [CONFIG_FILE]
```

Create a configuration file containing cluster information, such as kubeconfigs
files for all clusters. This file will be used with other commands to enable
them to inspect and collect data from all clusters.

If the file name is not specified we create the default config file
(config.yaml) that will be used automatically by all other commands.

#### Use cases

- *ramen testing environment*: The default kubeconfig (~/.kube/config) includes
  all the clusters, so we don't need to specify any kubeconfig. The names of the
  clusters are included in the drenv environment file.

- *real clusters*: User obtained kubeconfigs file for every cluster, either via
  OpenShift console, or from automaton system creating the clusters. The
  kubeconfigs may include same names for all clusters, or very long names which
  are hard to use. The config file will allow adding short and easy to use names
  for the clusters.

## Troubleshooting commands

### validate cluasters

```console
ramenctl validate cluasters [--config CONFIG_FILE] [--output REPORT_DIR]
```

Inspect the clusters specified in the configuration file and detect issues. The
report will provide a description of the clusters, list found issues, and
include relevant resources.

### validate application

```console
ramenctl validate application --drpc-name NAME [--namespace NAMESPACE]
    [--config CONFIG_FILE] [--output REPORT_DIR]
```

Inspect a protected application to identify issues. The report will provide an
overview of the protected application, list detected issues, and include related
resources.

#### Use cases

- *Subscription based application*: The drpc is located in specified namespace
  on the hub. The drpc name may clash with other drpc names in other namespaces.
  We must know both the drpc name and namespace to identify the application.

- *ApplicationSet based application*: All drpcs are located in the gitops
  namespace on the hub, and must be unique. We need only the drpc name.

- *OCM discovered application*: Add drpcs are located in the ramen ops namespace
  and must be unique. We need only the drpc name.

### test

```console
ramenctl test [--config CONFIG_FILE] [--output REPORT_DIR]
```

Run an end-to-end test to verify a complete disaster recovery flow using a
a tiny application. The configuration will define the workload, storage,
deployment type, and test flow.

#### Use cases

- *dr-policy*: The user may want to test an existing dr-policy in the cluster,
  or create a new dr-policy for the test.

- *storage*: The user may want to test one or more storage available in the
  system. The config file must speicfy the storage class name and the access
  modes.

- *workload*: The user may want to test Deployment, StatefulSet, DaemonSet, or
  VM. The config file must specify how to get the workload resources. The
  resources must be available in a git repository to allow OCM to deploy the
  resources.

- *deployment*: The user may want to test OCM managed application or OCM
  discovered application. The config file must specify how the workload is
  deployed. It may a Subscription, ApplicationSet, discovered application.

- *multiple configurations*: The user may want to test multiple confugrations.
  The config must allow testing multiple configurations in parallel. Testing
  disaster recovery is very slow (around 10 minutes) so testing in parallel is
  important.

- *cleanup*: If the test fails, the user may want to inspect the system to
  understand the failure, so the test command should not do any cleanup.
  However, after a failure is analyzed, the user will want to remove all
  traces of the test.

### gather application

```console
ramenctl gather application --drpc-name DRPC_NAME [--namespace NAMESPACE]
    [--config CONFIG_FILE] [--output REPORT_FILE]
```

Gather DR protected application resources and logs from relevant namespaces or
cluster scope in all clusters.

### diagnose

Diagnose issues and recommend actions.

## Management commands

The management commands are mainly needed for Kubernetes clusters, but are also
ussul for OpenShist clusters for autoamtion and testing.

### deploy

```console
ramenctl deploy [--config CONFIG_FILE]
```

Deploy the ramen operators across the clusters.

### undeploy

```console
ramenctl undeploy [--config CONFIG_FILE]
```

Undeploy the ramen operators across the clusters.

### config

```console
ramenctl config dr-cluster [--config CONFIG_FILE]
ramenctl config dr-policy [--config CONFIG_FILE]
ramenctl config s3-profile [--config CONFIG_FILE]
```

Create dr-clusters, dr-policies, and s3 profiles.

### protect

```console
ramenctl protect --app-name NAME [--namespace NAMESPACE] [--config CONFIG_FILE]
```

Enable DR protection for an OCM managed or discovered application.

### unprotect

```console
ramenctl unprotect --drpc-name NAME [--namespace NAMESPACE] [--config CONFIG_FILE]
```

Disable DR protection for a protected application.

### failover

```console
ramenctl failover --drpc-name NAME [--namespace NAMESPACE] [--config CONFIG_FILE]
```

Failover a protected application to the secondary cluster.

### relocate

```console
ramenctl relocate --drpc-name NAME [--namespace NAMESPACE] [--config CONFIG_FILE]
```

Relocate a protected application to the other cluster.
