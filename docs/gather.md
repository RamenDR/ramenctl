# ramenctl gather

The gather command collects diagnostic data from clusters involved in a
disaster recovery (DR) scenario. It gathers logs, resources, and configuration
from specified namespaces across the hub and managed clusters, helping with
troubleshooting and support.

```console
$ ramenctl gather -h
Collect diagnostic data from your clusters

Usage:
  ramenctl gather [command]

Available Commands:
  application Collect data based on application

Flags:
  -h, --help            help for gather
  -o, --output string   output directory

Global Flags:
  -c, --config string   configuration file (default "config.yaml")

Use "ramenctl gather [command] --help" for more information about a command.

```

## gather application

The gather application command gathers data for a specific DR-protected
application by inspecting its DR placement (DRPC) and collecting the namespaces
on the hub and managed clusters.

> [!IMPORTANT]
> The gather command requires a configuration file. See [init](docs/init.md) to
> learn how to create one.

### Looking up application DRPC

In order to execute the gather command, we need to know the DRPC name and
namespaces and these can be achieved with simple command below:

```console
$ oc get drpc -A
NAMESPACE          NAME                        AGE   PREFERREDCLUSTER   FAILOVERCLUSTER   DESIREDSTATE   CURRENTSTATE
openshift-dr-ops   disapp-deploy-rbd-busybox   13d   prsurve-c1-7j                                       Deployed
openshift-dr-ops   test-ns                     14d   prsurve-c1-7j                                       Deployed
openshift-gitops   appset-deploy-rbd-busybox   14d   prsurve-c1-7j                                       Deployed
```

### Gathering application data

Now that we have the DRPC name and namespaces we can run the gather command to
collect required namespaces.

```console
$ ramenctl gather application -o gather -c ocp.yaml --name disapp-deploy-rbd-busybox --namespace openshift-dr-ops
⭐ Using config "ocp.yaml"
⭐ Using report "gather"

🔎 Validate config ...
   ✅ Config validated

🔎 Gather Application data ...
   ✅ Inspected application
   ✅ Gathered data from cluster "prsurve-c2-7j"
   ✅ Gathered data from cluster "hub"
   ✅ Gathered data from cluster "prsurve-c1-7j"

✅ Gather completed
```

This command:

- Validates the configuration and cluster connectivity
- Identifies the application namespaces using the DRPC
- Includes ramen namespaces on the hub and managed cluster to
  collect ramen deployment status and ramen pods logs.
- Gathers Kubernetes resources and logs from all identified namespaces
- Outputs a structured report and collected data.

The command stores `gather-application.yaml` and `gather-application.log` in
the specified output directory:

```console
$ tree -L4 gather/
gather/
├── gather-application.data
│   ├── hub
│   │   ├── cluster
│   │   │   └── namespaces
│   │   └── namespaces
│   │       ├── openshift-dr-ops
│   │       └── openshift-operators
│   ├── prsurve-c1-7j
│   │   ├── cluster
│   │   │   └── namespaces
│   │   └── namespaces
│   │       ├── openshift-dr-ops
│   │       ├── openshift-dr-system
│   │       ├── openshift-operators
│   │       └── test-ns-2
│   └── prsurve-c2-7j
│       ├── cluster
│       │   └── namespaces
│       └── namespaces
│           ├── openshift-dr-ops
│           ├── openshift-dr-system
│           └── openshift-operators
├── gather-application.log
└── gather-application.yaml
```

## Example Report

```console
application:
  name: test-ns
  namespace: openshift-dr-ops
build:
  commit: 1770637cbe1e129786a0ec404a69e7f3b6a42a66
  version: v0.8.0-31-g1770637
config:
  clusterSet: clusterset-submariner-52bbff94cfe4421185
  clusters:
    c1:
      kubeconfig: ocp/c1
    c2:
      kubeconfig: ocp/c2
    hub:
      kubeconfig: ocp/hub
    passive-hub:
      kubeconfig: ""
  distro: ocp
  namespaces:
    argocdNamespace: openshift-gitops
    ramenDRClusterNamespace: openshift-dr-system
    ramenHubNamespace: openshift-operators
    ramenOpsNamespace: openshift-dr-ops
created: "2025-07-22T16:14:43.903524674+05:30"
duration: 141.621068139
host:
  arch: amd64
  cpus: 16
  os: linux
name: gather-application
namespaces:
- openshift-dr-ops
- openshift-dr-system
- openshift-operators
- test-ns-2
status: passed
steps:
- duration: 4.131192067
  name: validate config
  status: passed
- duration: 137.489876072
  items:
  - duration: 0.616132191
    name: inspect application
    status: passed
  - duration: 109.387906106
    name: gather "prsurve-c2-7j"
    status: passed
  - duration: 127.375111889
    name: gather "prsurve-c1-7j"
    status: passed
  - duration: 136.873366241
    name: gather "hub"
    status: passed
  name: gather data
  status: passed
```
