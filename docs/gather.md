# ramenctl gather application

The gather command collects diagnostic data from clusters involved in
a disaster recovery (DR) scenario. It gathers logs, resources, and configuration
from specified namespaces across the hub and managed clusters, 
helping with troubleshooting and support.

```console
$ ramenctl gather application --help
Collect data based on application

Usage:
  ramenctl gather application [flags]

Flags:
  -h, --help               help for application
      --name string        drpc name
  -n, --namespace string   drpc namespace

Global Flags:
  -c, --config string   configuration file (default "config.yaml")
  -o, --output string   output directory
```

## gather application

The gather application command gathers data for a specific DR-protected application
by inspecting its DR placement (DRPC) and collecting resources from relevant
namespaces on the hub and managed clusters.

> [!IMPORTANT]
> The test command requires a configuration file. See [init](docs/init.md) to
> learn how to create one.

```console
$ ramenctl gather application -o gather -c ocp.yaml --name <drpc-name> --namespace <drpc-namespace>
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
- Identifies the application namespaces using the DRPC (drpc-name in drpc-namespace)
- Includes Ramen control plane namespaces (ramen-hub, ramen-dr-cluster)
- Gathers Kubernetes resources and logs from all identified namespaces
- Outputs a structured report and collected data.

The command stores `gather-application.yaml` and `gather-application.log` in the specified output directory:

```console
$ tree gather-output
gather-output
├── gather-application.log
├── gather-application.yaml
└── gather-application.data
    ├── hub
    │   ├── namespace1
    │   └── ramen-hub
    ├── c1
    │   └── namespace1
    └── c2
        └── namespace1
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

## Example usage

```console
# Gather data for a specific DRPC
ramenctl gather application my-drpc -n my-dr-namespace -o /tmp/diag

# Use a custom config file
ramenctl gather application my-drpc -n my-dr-namespace -o /tmp/diag -c /path/to/config.yaml
```