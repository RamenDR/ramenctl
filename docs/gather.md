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

### The Gather Report

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

## Inspecting logs

Example of inspecting ramen log on the managed cluter.

```bash
$ head -5 gather/gather-application.data/prsurve-c1-7j/namespaces/openshift-dr-system/pods/ramen-dr-cluster-operator-7cb7d655bf-2bpd2/manager/current.log 

2025-07-21T21:19:27.794Z	ERROR	vrg	controller/vrg_vrgobject.go:49	VRG Kube object protect error	{"vrg": {"name":"appset-deploy-rbd-busybox","namespace":"e2e-appset-deploy-rbd-busybox"}, "rid": "31b4e607", "State": "primary", "profile": "s3profile-prsurve-c1-7j-ocs-storagecluster", "error": "failed to upload data of odrbucket-ebc94e32267b:e2e-appset-deploy-rbd-busybox/appset-deploy-rbd-busybox/v1alpha1.VolumeReplicationGroup/a: code: RequestCanceled, message: request context canceled"}
2025-07-21T21:19:27.794Z	DEBUG	events	recorder/recorder.go:104	failed to upload data of odrbucket-ebc94e32267b:e2e-appset-deploy-rbd-busybox/appset-deploy-rbd-busybox/v1alpha1.VolumeReplicationGroup/a: code: RequestCanceled, message: request context canceled	{"type": "Warning", "object": {"kind":"VolumeReplicationGroup","namespace":"e2e-appset-deploy-rbd-busybox","name":"appset-deploy-rbd-busybox","uid":"efce6b42-cefb-4c8d-bcab-e2dc8ab6d429","apiVersion":"ramendr.openshift.io/v1alpha1","resourceVersion":"32343261"}, "reason": "VrgUploadFailed"}
2025-07-21T21:19:27.794Z	INFO	vrg	controller/vrg_volrep.go:2605	Condition for DataReady	{"vrg": {"name":"appset-deploy-rbd-busybox","namespace":"e2e-appset-deploy-rbd-busybox"}, "rid": "31b4e607", "State": "primary", "cond": "&Condition{Type:DataReady,Status:True,ObservedGeneration:1,LastTransitionTime:2025-07-12 21:09:06 +0000 UTC,Reason:Ready,Message:PVC in the VolumeReplicationGroup is ready for use,}", "protectedPVC": {"namespace":"e2e-appset-deploy-rbd-busybox","name":"busybox-pvc","csiProvisioner":"openshift-storage.rbd.csi.ceph.com","storageID":{"id":"06d422497c1b8a38ba29b4d6d68310c3"},"replicationID":{"id":"93e9e0d4203b3742ccc77bb146af1edf","modes":["Failover"]},"storageClassName":"ocs-storagecluster-ceph-rbd","labels":{"app.kubernetes.io/instance":"appset-deploy-rbd-busybox-prsurve-c1-7j","appname":"busybox"},"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"conditions":[{"type":"DataReady","status":"True","observedGeneration":1,"lastTransitionTime":"2025-07-12T21:09:06Z","reason":"Ready","message":"PVC in the VolumeReplicationGroup is ready for use"},{"type":"ClusterDataProtected","status":"True","observedGeneration":1,"lastTransitionTime":"2025-07-12T21:09:06Z","reason":"Uploaded","message":"PV cluster data already protected for PVC busybox-pvc"},{"type":"DataProtected","status":"False","observedGeneration":1,"lastTransitionTime":"2025-07-12T21:09:06Z","reason":"Replicating","message":"PVC in the VolumeReplicationGroup is ready for use"}],"lastSyncTime":"2025-07-20T13:25:00Z","lastSyncDuration":"0s","volumeMode":"Filesystem"}}
2025-07-21T21:19:27.794Z	INFO	vrg	controller/volumereplicationgroup_controller.go:1869	Marking VRG ready with replicating reason	{"vrg": {"name":"appset-deploy-rbd-busybox","namespace":"e2e-appset-deploy-rbd-busybox"}, "rid": "31b4e607", "State": "primary", "reason": "Ready"}
2025-07-21T21:19:27.794Z	INFO	vrg	controller/volumereplicationgroup_controller.go:1806	merging DataReady condition	{"vrg": {"name":"appset-deploy-rbd-busybox","namespace":"e2e-appset-deploy-rbd-busybox"}, "rid": "31b4e607", "State": "primary", "subconditions": [{"type":"DataReady","status":"True","observedGeneration":1,"lastTransitionTime":null,"reason":"Unused","message":"No PVCs are protected using Volsync scheme"},{"type":"DataReady","status":"True","observedGeneration":1,"lastTransitionTime":null,"reason":"Ready","message":"PVCs in the VolumeReplicationGroup are ready for use"}]}
```

### Application namespaces

```
$ tree gather/gather-application.data/prsurve-c1-7j/namespaces/test-ns-2/
gather/gather-application.data/prsurve-c1-7j/namespaces/test-ns-2/
├── apps
│   ├── deployments
│   │   └── test-dep2.yaml
│   └── replicasets
│       └── test-dep2-5d777fc77d.yaml
├── authorization.openshift.io
│   └── rolebindings
│       ├── admin.yaml
│       ├── system:deployers.yaml
│       ├── system:image-builders.yaml
│       └── system:image-pullers.yaml
├── configmaps
│   ├── kube-root-ca.crt.yaml
│   └── openshift-service-ca.crt.yaml
├── metrics.k8s.io
│   └── pods
│       ├── test-dep2-5d777fc77d-k8wmv.yaml
│       ├── test-dep2-5d777fc77d-l86t5.yaml
│       └── test-dep2-5d777fc77d-wfjhl.yaml
├── operators.coreos.com
│   └── clusterserviceversions
│       ├── odr-cluster-operator.v4.19.0-rhodf.yaml
│       └── openshift-gitops-operator.v1.16.1.yaml
├── pods
│   ├── test-dep2-5d777fc77d-k8wmv
│   │   └── container
│   │       └── current.log
│   ├── test-dep2-5d777fc77d-k8wmv.yaml
│   ├── test-dep2-5d777fc77d-l86t5
│   │   └── container
│   │       └── current.log
│   ├── test-dep2-5d777fc77d-l86t5.yaml
│   ├── test-dep2-5d777fc77d-wfjhl
│   │   └── container
│   │       └── current.log
│   └── test-dep2-5d777fc77d-wfjhl.yaml
├── rbac.authorization.k8s.io
│   └── rolebindings
│       ├── admin.yaml
│       ├── system:deployers.yaml
│       ├── system:image-builders.yaml
│       └── system:image-pullers.yaml
├── secrets
│   ├── builder-dockercfg-f8fmg.yaml
│   ├── default-dockercfg-9h8q8.yaml
│   └── deployer-dockercfg-9x74l.yaml
└── serviceaccounts
    ├── builder.yaml
    ├── default.yaml
    └── deployer.yaml

22 directories, 29 files
```