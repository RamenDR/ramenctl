<!-- SPDX-FileCopyrightText: The RamenDR authors -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

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
  clusters    Detect problems in disaster recovery clusters

Flags:
  -h, --help            help for validate
  -o, --output string   output directory

Global Flags:
  -c, --config string   configuration file (default "config.yaml")

Use "ramenctl validate [command] --help" for more information about a command.
```

The command supports the following sub-commands:

* [application](#validate-application)
* [clusters](#validate-clusters)

> [!IMPORTANT]
> The validate command requires a configuration file. See [init](docs/init.md) to
> learn how to create one.

## validate application

The validate application command validates a specific DR-protected application
by gathering related namespaces from all clusters, S3 data and inspecting the
gathered resources.

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
   ✅ Inspected S3 profiles
   ✅ Gathered S3 profile "minio-on-dr1"
   ✅ Gathered S3 profile "minio-on-dr2"
   ✅ Application validated

✅ Validation completed (24 ok, 0 stale, 0 problem)
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

The most important part of the report is the `applicationStatus`:

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
  s3:
    profiles:
      state: ok ✅
      value:
      - gathered:
          state: ok ✅
          value: true
        name: minio-on-dr1
      - gathered:
          state: ok ✅
          value: true
        name: minio-on-dr2
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
├── hub
│   ├── cluster
│   │   └── namespaces
│   └── namespaces
│       └── argocd
└── s3
    ├── minio-on-dr1
    │   └── e2e-appset-deploy-rbd
    └── minio-on-dr2
        └── e2e-appset-deploy-rbd
```

### The validate-application.log

This log includes detailed information that may help to troubleshoot the
validate application command. If the command failed, check the error details in
the log.

## validate clusters

The validate clusters command validates the disaster recovery clusters by
gathering cluster scoped and related ramen resources from all clusters, and
validates that configured S3 endpoints are accessible.

### Validating clusters

To validate the disaster recovery clusters, run the following command:

```console
$ ramenctl validate clusters -o out
⭐ Using config "config.yaml"
⭐ Using report "out"

🔎 Validate config ...
   ✅ Config validated

🔎 Validate clusters ...
   ✅ Gathered data from cluster "hub"
   ✅ Gathered data from cluster "dr1"
   ✅ Gathered data from cluster "dr2"
   ✅ Inspected S3 profiles
   ✅ Checked S3 profile "minio-on-dr2"
   ✅ Checked S3 profile "minio-on-dr1"
   ✅ Clusters validated

✅ Validation completed (90 ok, 0 stale, 0 problem)
```

The command gathered cluster scoped and ramen resources from all clusters,
inspected the resources, and stored output files in the specified output
directory:

```console
$ tree -L1 out
out
├── validate-clusters.data
├── validate-clusters.log
└── validate-clusters.yaml
```

> [!IMPORTANT]
> When reporting DR related issues, please create an archive with the output
> directory and upload it to the issue tracker.

### The validate-clusters.yaml

The `validate-clusters.yaml` report is a machine and human readable description
of the command and the clusters status.

The most important part of the report is the `clustersStatus`:

```yaml
clustersStatus:
  clusters:
  - name: dr1
    ramen:
      configmap:
        deleted:
          state: ok ✅
        name: ramen-dr-cluster-operator-config
        namespace: ramen-system
        ramenControllerType:
          state: ok ✅
          value: dr-cluster
        s3StoreProfiles:
          state: ok ✅
          value:
          - bucket:
              state: ok ✅
              value: bucket
            caCertificate:
              state: ok ✅
            endpoint:
              state: ok ✅
              value: http://dr1-endpoint:30000
            profileName: minio-on-dr1
            region:
              state: ok ✅
              value: us-west-1
            secret:
              awsAccessKeyID:
                state: ok ✅
                value: 57:91:B6:22:67...
              awsSecretAccessKey:
                state: ok ✅
                value: 17:9C:07:6A:C5...
              deleted:
                state: ok ✅
              name:
                state: ok ✅
                value: ramen-s3-secret-dr1
              namespace:
                state: ok ✅
                value: ramen-system
          - bucket:
              state: ok ✅
              value: bucket
            caCertificate:
              state: ok ✅
            endpoint:
              state: ok ✅
              value: http://dr2-endpoint:30000
            profileName: minio-on-dr2
            region:
              state: ok ✅
              value: us-east-1
            secret:
              awsAccessKeyID:
                state: ok ✅
                value: 57:91:B6:22:67...
              awsSecretAccessKey:
                state: ok ✅
                value: 17:9C:07:6A:C5...
              deleted:
                state: ok ✅
              name:
                state: ok ✅
                value: ramen-s3-secret-dr2
              namespace:
                state: ok ✅
                value: ramen-system
      deployment:
        conditions:
        - state: ok ✅
          type: Available
        - state: ok ✅
          type: Progressing
        deleted:
          state: ok ✅
        name: ramen-dr-cluster-operator
        namespace: ramen-system
        replicas:
          state: ok ✅
          value: 1
  - name: dr2
    ramen:
      configmap:
        deleted:
          state: ok ✅
        name: ramen-dr-cluster-operator-config
        namespace: ramen-system
        ramenControllerType:
          state: ok ✅
          value: dr-cluster
        s3StoreProfiles:
          state: ok ✅
          value:
          - bucket:
              state: ok ✅
              value: bucket
            caCertificate:
              state: ok ✅
            endpoint:
              state: ok ✅
              value: http://dr1-endpoint:30000
            profileName: minio-on-dr1
            region:
              state: ok ✅
              value: us-west-1
            secret:
              awsAccessKeyID:
                state: ok ✅
                value: 57:91:B6:22:67...
              awsSecretAccessKey:
                state: ok ✅
                value: 17:9C:07:6A:C5...
              deleted:
                state: ok ✅
              name:
                state: ok ✅
                value: ramen-s3-secret-dr1
              namespace:
                state: ok ✅
                value: ramen-system
          - bucket:
              state: ok ✅
              value: bucket
            caCertificate:
              state: ok ✅
            endpoint:
              state: ok ✅
              value: http://dr2-endpoint:30000
            profileName: minio-on-dr2
            region:
              state: ok ✅
              value: us-east-1
            secret:
              awsAccessKeyID:
                state: ok ✅
                value: 57:91:B6:22:67...
              awsSecretAccessKey:
                state: ok ✅
                value: 17:9C:07:6A:C5...
              deleted:
                state: ok ✅
              name:
                state: ok ✅
                value: ramen-s3-secret-dr2
              namespace:
                state: ok ✅
                value: ramen-system
      deployment:
        conditions:
        - state: ok ✅
          type: Progressing
        - state: ok ✅
          type: Available
        deleted:
          state: ok ✅
        name: ramen-dr-cluster-operator
        namespace: ramen-system
        replicas:
          state: ok ✅
          value: 1
  hub:
    drClusters:
      state: ok ✅
      value:
      - conditions:
        - state: ok ✅
          type: Fenced
        - state: ok ✅
          type: Clean
        - state: ok ✅
          type: Validated
        name: dr1
        phase: Available
      - conditions:
        - state: ok ✅
          type: Fenced
        - state: ok ✅
          type: Clean
        - state: ok ✅
          type: Validated
        name: dr2
        phase: Available
    drPolicies:
      state: ok ✅
      value:
      - conditions:
        - state: ok ✅
          type: Validated
        drClusters:
        - dr1
        - dr2
        name: dr-policy
        schedulingInterval: 1m
    ramen:
      configmap:
        deleted:
          state: ok ✅
        name: ramen-hub-operator-config
        namespace: ramen-system
        ramenControllerType:
          state: ok ✅
          value: dr-hub
        s3StoreProfiles:
          state: ok ✅
          value:
          - bucket:
              state: ok ✅
              value: bucket
            caCertificate:
              state: ok ✅
            endpoint:
              state: ok ✅
              value: http://dr1-endpoint:30000
            profileName: minio-on-dr1
            region:
              state: ok ✅
              value: us-west-1
            secret:
              awsAccessKeyID:
                state: ok ✅
                value: 57:91:B6:22:67...
              awsSecretAccessKey:
                state: ok ✅
                value: 17:9C:07:6A:C5...
              deleted:
                state: ok ✅
              name:
                state: ok ✅
                value: ramen-s3-secret-dr1
              namespace:
                state: ok ✅
                value: ramen-system
          - bucket:
              state: ok ✅
              value: bucket
            caCertificate:
              state: ok ✅
            endpoint:
              state: ok ✅
              value: http://dr2-endpoint:30000
            profileName: minio-on-dr2
            region:
              state: ok ✅
              value: us-east-1
            secret:
              awsAccessKeyID:
                state: ok ✅
                value: 57:91:B6:22:67...
              awsSecretAccessKey:
                state: ok ✅
                value: 17:9C:07:6A:C5...
              deleted:
                state: ok ✅
              name:
                state: ok ✅
                value: ramen-s3-secret-dr2
              namespace:
                state: ok ✅
                value: ramen-system
      deployment:
        conditions:
        - state: ok ✅
          type: Available
        - state: ok ✅
          type: Progressing
        deleted:
          state: ok ✅
        name: ramen-hub-operator
        namespace: ramen-system
        replicas:
          state: ok ✅
          value: 1
  s3:
    profiles:
      state: ok ✅
      value:
      - accessible:
          state: ok ✅
          value: true
        name: minio-on-dr2
      - accessible:
          state: ok ✅
          value: true
        name: minio-on-dr1
```

### The validate-clusters.data directory

This directory contains all data gathered during validation. Use the gathered
data to investigate the problems reported in the `validate-clusters.yaml` report.

```console
$ tree -L3 out/validate-clusters.data
out/validate-clusters.data
├── dr1
│   ├── cluster
│   │   ├── apiextensions.k8s.io
│   │   ├── apiregistration.k8s.io
│   │   ├── cluster.open-cluster-management.io
│   │   ├── flowcontrol.apiserver.k8s.io
│   │   ├── namespaces
│   │   ├── networking.k8s.io
│   │   ├── nodes
│   │   ├── operator.open-cluster-management.io
│   │   ├── operators.coreos.com
│   │   ├── persistentvolumes
│   │   ├── ramendr.openshift.io
│   │   ├── rbac.authorization.k8s.io
│   │   ├── replication.storage.openshift.io
│   │   ├── scheduling.k8s.io
│   │   ├── snapshot.storage.k8s.io
│   │   ├── storage.k8s.io
│   │   ├── submariner.io
│   │   └── work.open-cluster-management.io
│   └── namespaces
│       └── ramen-system
├── dr2
│   ├── cluster
│   │   ├── apiextensions.k8s.io
│   │   ├── apiregistration.k8s.io
│   │   ├── cluster.open-cluster-management.io
│   │   ├── flowcontrol.apiserver.k8s.io
│   │   ├── namespaces
│   │   ├── networking.k8s.io
│   │   ├── nodes
│   │   ├── operator.open-cluster-management.io
│   │   ├── operators.coreos.com
│   │   ├── persistentvolumes
│   │   ├── ramendr.openshift.io
│   │   ├── rbac.authorization.k8s.io
│   │   ├── replication.storage.openshift.io
│   │   ├── scheduling.k8s.io
│   │   ├── snapshot.storage.k8s.io
│   │   ├── storage.k8s.io
│   │   ├── submariner.io
│   │   └── work.open-cluster-management.io
│   └── namespaces
│       └── ramen-system
└── hub
    ├── cluster
    │   ├── addon.open-cluster-management.io
    │   ├── admissionregistration.k8s.io
    │   ├── apiextensions.k8s.io
    │   ├── apiregistration.k8s.io
    │   ├── cluster.open-cluster-management.io
    │   ├── flowcontrol.apiserver.k8s.io
    │   ├── namespaces
    │   ├── networking.k8s.io
    │   ├── nodes
    │   ├── operator.open-cluster-management.io
    │   ├── operators.coreos.com
    │   ├── ramendr.openshift.io
    │   ├── rbac.authorization.k8s.io
    │   ├── scheduling.k8s.io
    │   └── storage.k8s.io
    └── namespaces
        └── ramen-system
```

### The validate-clusters.log

This log includes detailed information that may help to troubleshoot the
validate clusters command. If the command failed, check the error details in
the log.
