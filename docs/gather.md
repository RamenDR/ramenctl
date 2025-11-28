# ramenctl gather

The gather command helps to troubleshoot disaster recovery issues by gathering
data about protected applications.

```console
$ ramenctl gather -h
Collect diagnostic data from your clusters

Usage:
  ramenctl gather [command]

Available Commands:
  application Collect data for a protected application

Flags:
  -h, --help            help for gather
  -o, --output string   output directory

Global Flags:
  -c, --config string   configuration file (default "config.yaml")

Use "ramenctl gather [command] --help" for more information about a command.
```

> [!IMPORTANT]
> The gather command requires a configuration file. See [init](docs/init.md) to
> learn how to create one.

## gather application

The gather application command gathers data for a specific disaster recover
protected application. It gathers entire namespaces and S3 data related to the
protected application across the hub and the managed clusters.

### Looking up applications

To run the gather application command, we need to find the protected
application name and namespace. Run the following command:

```console
$ kubectl get drpc -A --context hub
NAMESPACE   NAME                AGE     PREFERREDCLUSTER   FAILOVERCLUSTER   DESIREDSTATE   CURRENTSTATE
argocd      appset-deploy-rbd   6m16s   dr1                                                 Deployed
```

### Gathering application data

To gather data for the application `appset-deploy-rbd` in namespace `argocd`
run the following command:

```console
$ ramenctl gather application --name appset-deploy-rbd --namespace argocd -o out
⭐ Using config "config.yaml"
⭐ Using report "out"

🔎 Validate config ...
   ✅ Config validated

🔎 Gather application data ...
   ✅ Inspected application
   ✅ Gathered data from cluster "hub"
   ✅ Gathered data from cluster "dr1"
   ✅ Gathered data from cluster "dr2"
   ✅ Inspected S3 profiles
   ✅ Gathered S3 profile "minio-on-dr1"
   ✅ Gathered S3 profile "minio-on-dr2"

✅ Gather completed
```

The command gathered related namespaced from all clusters and stored output
files in the specified output directory:

```console
$ tree -L1 out
out
├── gather-application.data
├── gather-application.log
└── gather-application.yaml
```

### The gather-appplication.data directory

This directory contains resources (namespaced and cluster scoped) and S3 data
related to the protected application. The data depend on the application
deployment type.

```console
$ tree -L3 out/gather-application.data
out/gather-application.data
├── dr1
│   ├── cluster
│   │   ├── namespaces
│   │   ├── persistentvolumes
│   │   └── storage.k8s.io
│   └── namespaces
│       ├── e2e-appset-deploy-rbd
│       └── ramen-system
├── dr2
│   ├── cluster
│   │   └── namespaces
│   └── namespaces
│       ├── e2e-appset-deploy-rbd
│       └── ramen-system
└── s3
    ├── minio-on-dr1
    │   └── e2e-appset-deploy-rbd
    └── minio-on-dr2
        └── e2e-appset-deploy-rbd
```

You can use standard tools to inspect the resources:

```console
$ yq '.status.protectedPVCs[0].conditions' < out/gather-application.data/dr1/namespaces/e2e-appset-deploy-rbd/ramendr.openshift.io/volumereplicationgroups/appset-deploy-rbd.yaml
- lastTransitionTime: "2025-08-17T17:45:41Z"
  message: PVC in the VolumeReplicationGroup is ready for use
  observedGeneration: 1
  reason: Ready
  status: "True"
  type: DataReady
- lastTransitionTime: "2025-08-17T17:45:40Z"
  message: PV cluster data already protected for PVC busybox-pvc
  observedGeneration: 1
  reason: Uploaded
  status: "True"
  type: ClusterDataProtected
- lastTransitionTime: "2025-08-17T17:45:41Z"
  message: PVC in the VolumeReplicationGroup is ready for use
  observedGeneration: 1
  reason: Replicating
  status: "False"
  type: DataProtected
```

You can also inspect ramen logs in all clusters:

```console
$ grep -E 'ERROR.+appset-deploy-rbd' out/gather-application.data/dr1/namespaces/ramen-system/pods/ramen-dr-cluster-operator-67dff877f5-k4gjm/manager/current.log
2025-08-17T17:45:40.644Z	ERROR	vrg	controller/vrg_volrep.go:122	Requeuing due to failure to upload PV object to S3 store(s)	{"vrg": {"name":"appset-deploy-rbd","namespace":"e2e-appset-deploy-rbd"}, "rid": "1c5b6d55", "State": "primary", "pvc": "e2e-appset-deploy-rbd/busybox-pvc", "error": "failed to add archived annotation for PVC (e2e-appset-deploy-rbd/busybox-pvc): failed to update PersistentVolumeClaim (e2e-appset-deploy-rbd/busybox-pvc) annotation (volumereplicationgroups.ramendr.openshift.io/vr-archived) belonging to VolumeReplicationGroup (e2e-appset-deploy-rbd/appset-deploy-rbd), Operation cannot be fulfilled on persistentvolumeclaims \"busybox-pvc\": the object has been modified; please apply your changes to the latest version and try again"}
```

### The gather-application.yaml

The `gather-application.yaml` report is a machine and human readable description
of the command. It can be useful to troubleshoot the gather application command.

### The gather-application.log

This log includes detailed information that may help to troubleshoot the gather
application command. If the command failed, check the error details in the log.
