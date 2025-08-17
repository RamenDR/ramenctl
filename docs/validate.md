# ramenctl validate

The validate commands help to troubleshoot disaster recovery problems. They
gathers data from the clusters and detects problems in configuration and the
current status of the clusters or protected applications.

```console
$ ramenctl validate -h
Detect disaster recovery problems

Usage:
  ramenctl validate [command]

Available Commands:
  application Detect problems in disaster recovery protected application

Flags:
  -h, --help            help for validate
  -o, --output string   output directory

Global Flags:
  -c, --config string   configuration file (default "config.yaml")

Use "ramenctl validate [command] --help" for more information about a command.
```

The command supports the following sub-commands:

* [application](#validate-application)

> [!IMPORTANT]
> The validate command requires a configuration file. See [init](docs/init.md) to
> learn how to create one.

## validate application

The validate application command validates a specific DR-protected application
by gathering related namespaces from all clusters and inspecting the gathered
resources.

### Looking up applications

To run the validate application command, we need to find the protected
application name and namespace. Run the following command:

```console
$ kubectl get drpc -A --context hub
NAMESPACE   NAME                   AGE   PREFERREDCLUSTER   FAILOVERCLUSTER   DESIREDSTATE   CURRENTSTATE
argocd      appset-deploy-rbd      69m   dr1                dr2               Relocate       Relocated
```

### Validating an application

To validate the application `appset-deploy-rbd` in namespace `argocd` run the
following command:

```console
$ ramenctl validate application --name appset-deploy-rbd --namespace argocd -o out
⭐ Using config "config.yaml"
⭐ Using report "out"

🔎 Validate config ...
   ✅ Config validated

🔎 Validate application ...
   ✅ Inspected application
   ✅ Gathered data from cluster "dr2"
   ✅ Gathered data from cluster "dr1"
   ✅ Gathered data from cluster "hub"
   ✅ Application validated

✅ Validation completed (21 ok, 0 stale, 0 problem)
```

The command gathered related namespaced from all clusters, inspected the
resources, and stored output files in the specified output directory:

```console
$ tree -L1 out
out
├── validate-application.data
├── validate-application.log
└── validate-application.yaml
```

> [!IMPORTANT]
> When reporting DR related issues, please create an archive with the output
> directory and upload it to the issue tracker.

### The validate-application.yaml

The `validate-application.yaml` report is a machine and human readable
description of the command and the application status.

The most important part of the report is the applicationStatus:

```yaml
applicationStatus:
  hub:
    drpc:
      action:
        state: ok ✅
        value: Relocate
      conditions:
      - state: ok ✅
        type: Available
      - state: ok ✅
        type: PeerReady
      - state: ok ✅
        type: Protected
      deleted:
        state: ok ✅
      drPolicy: dr-policy
      name: appset-deploy-rbd
      namespace: argocd
      phase:
        state: ok ✅
        value: Relocated
      progression:
        state: ok ✅
        value: Completed
  primaryCluster:
    name: dr1
    vrg:
      conditions:
      - state: ok ✅
        type: DataReady
      - state: ok ✅
        type: ClusterDataReady
      - state: ok ✅
        type: ClusterDataProtected
      - state: ok ✅
        type: KubeObjectsReady
      - state: ok ✅
        type: NoClusterDataConflict
      deleted:
        state: ok ✅
      name: appset-deploy-rbd
      namespace: e2e-appset-deploy-rbd
      protectedPVCs:
      - conditions:
        - state: ok ✅
          type: DataReady
        - state: ok ✅
          type: ClusterDataProtected
        deleted:
          state: ok ✅
        name: busybox-pvc
        namespace: e2e-appset-deploy-rbd
        phase:
          state: ok ✅
          value: Bound
        replication: volrep
      state:
        state: ok ✅
        value: Primary
  secondaryCluster:
    name: dr2
    vrg:
      conditions:
      - state: ok ✅
        type: NoClusterDataConflict
      deleted:
        state: ok ✅
      name: appset-deploy-rbd
      namespace: e2e-appset-deploy-rbd
      state:
        state: ok ✅
        value: Secondary
```

### The validate-application.data directory

This directory contains all data gathered during validation. The data depend on
the application deployment type. Use the gathered data to investigate the
problems reported in the `validate-application.yaml` report.

```console
$ tree -L3 out/validate-application.data
out/validate-application.data
├── dr1
│   ├── cluster
│   │   ├── namespaces
│   │   ├── persistentvolumes
│   │   └── storage.k8s.io
│   └── namespaces
│       └── e2e-appset-deploy-rbd
├── dr2
│   ├── cluster
│   │   └── namespaces
│   └── namespaces
│       └── e2e-appset-deploy-rbd
└── hub
    ├── cluster
    │   └── namespaces
    └── namespaces
        └── argocd
```

### The validate-application.log

This log includes detailed information that may help to troubleshoot the
validate application command. If the command failed, check the error details in
the log.
