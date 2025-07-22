# ramenctl init

The init command crates a configuration file required for other
*ramenctl* commands.

```console
% ramenctl init -h
Create configuration file for your clusters

Usage:
  ramenctl init [flags]

Flags:
      --envfile string   ramen testing environment file
  -h, --help             help for init

Global Flags:
  -c, --config string   configuration file (default "config.yaml")
```

## Creating a configuration file

The init command creates a configuration file named "config.yaml" in the
current directory:

```console
$ ramenctl init

✅ Created config file "config.yaml" - please modify for your clusters
```

> [!IMPORTANT]
> Before using the configuration file you need to edit it to match your
> clusters and storage.

Other *ramenctl* commands use "config.yaml" by default.

## Creating configuration file for a ramen testing environment

When using a ramen testing environment we can create a configuration file
optimized for the testing environment using the `--envfile` option:

```console
$ ramenctl init --envfile ../ramen/test/envs/regional-dr.yaml
⭐ Using envfile "../ramen/test/envs/regional-dr.yaml"

✅ Created config file "config.yaml" - please modify for your clusters
```

You can edit the configuration file to change the default tests.

## Using multiple configuration files

When working with multiple environments or when you want to run
different sets of tests with the same environment, you can create
multiple configuration files and use them with the `--config` option.

Create a configuration file named "myenv.yaml":

```console
$ ramenctl init --config myenv.yaml

✅ Created config file "myenv.yaml" - please modify for your clusters
```

To use the configuration file with other commands, specify it with the
`--config` option:

```console
$ ramenctl test run --config myenv.yaml -o test
⭐ Using report "test"
⭐ Using config "myenv.yaml"
...
```

## Sample configuration file

The following is a sample configuration file showing the default values. You
must modify it to match your clusters and storage.

```yaml
## ramenctl configuration file

## Clusters configuration.
# - Modify clusters "kubeconfig" to match your hub and managed clusters
#   kubeconfig files.
# - Modify "passive-hub" kubeconfig for optional passive hub cluster,
#   leave it empty if not using passive hub.
clusters:
  hub:
    kubeconfig: hub/config
  passive-hub:
    kubeconfig: ""
  c1:
    kubeconfig: primary/config
  c2:
    kubeconfig: secondary/config

## Git repository for test command.
# - Modify "url" to use your own Git repository.
# - Modify "branch" to test a different branch.
repo:
  url: https://github.com/RamenDR/ocm-ramen-samples.git
  branch: main

## DRPolicy for test command.
# - Modify to match actual DRPolicy in the hub cluster.
drPolicy: dr-policy

## ClusterSet for test command.
# - Modify to match your Open Cluster Management configuration.
clusterSet: default

## PVC specifications for test command.
# - Modify items "storageclassname" to match the actual storage classes in the
#   managed clusters.
# - Add new items for testing more storage types.
pvcSpecs:
- name: rbd
  storageClassName: rook-ceph-block
  accessModes: ReadWriteOnce
- name: cephfs
  storageClassName: rook-cephfs-fs1
  accessModes: ReadWriteMany

## Deployer specifications for test command.
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
  description: Discovered Application deployer for discovered applications

## Tests cases for test command.
# - Modify the test for your preferred workload or deployment type.
# - Add new tests for testing more combinations in parallel.
# - Available workloads: deploy
# - Available deployers: appset, subscr, disapp
tests:
- workload: deploy
  deployer: appset
  pvcSpec: rbd
```
