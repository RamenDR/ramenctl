<!--
SPDX-FileCopyrightText: The RamenDR authors
SPDX-License-Identifier: Apache-2.0
-->

# ramenctl

[![Actions Status](https://github.com/ramendr/ramenctl/workflows/Test/badge.svg)](https://github.com/ramendr/ramenctl/actions)
[![GoReport Status](https://goreportcard.com/badge/github.com/ramendr/ramenctl)](https://goreportcard.com/report/github.com/ramendr/ramenctl)
[![GitHub All Releases](https://img.shields.io/github/downloads/ramendr/ramenctl/total.svg)](https://github.com/ramendr/ramenctl/releases/latest)
[![Latest Release](https://img.shields.io/github/v/release/ramendr/ramenctl)](https://github.com/ramendr/ramenctl/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/ramendr/ramenctl.svg)](https://pkg.go.dev/github.com/ramendr/ramenctl)

Command line tool and Go module for managing and troubleshooting Ramen.

## Overview

Working with a complicated Kubernetes cluster is not easy. In a typical disaster
recovery environment we have at least 3 connected Kubernetes clusters with many
components. The *ramenctl* project aims to make it easier to manage and
troubleshoot this challenging environment.

## Features

The project provides:

- The *ramenctl* command line tool, managing and troubleshooting ramen.
- The *ramenctl* Go module for integrating the ramenctl commands in other
  projects. This module is used to implement the
  [odf dr](https://github.com/red-hat-storage/odf-cli/blob/main/docs/dr.md)
  command.

## Installing

Download the *ramenctl* executable for your operating system and architecture
and install in the PATH.

To install the latest release on Linux and macOS, run:

```console
os="$(uname | tr '[:upper:]' '[:lower:]')"
machine="$(uname -m)"
if [ "$machine" = "aarch64" ]; then machine="arm64"; fi
if [ "$machine" = "x86_64" ]; then machine="amd64"; fi
curl --location --fail --silent --show-error --output ramenctl \
    https://github.com/ramendr/ramenctl/releases/latest/download/ramenctl-$os-$machine
sudo install ramenctl /usr/local/bin/
rm ramenctl
```

## Examples

Create a configuration file for Regional DR test environment:

```console
$ ramenctl init --envfile ramen/test/envs/regional-dr.yaml
```

Validate the disaster recovery clusters and create an HTML report:

```console
$ ramenctl validate clusters -o validate-clusters
⭐ Using config "config.yaml"
⭐ Using report "validate-clusters"

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

✅ Validation completed (90 ok, 0 warning, 0 problem)
```

Validate a protected application and create an HTML report:

```console
$ ramenctl validate application --name appset-deploy-rbd \
    --namespace argocd -o validate-application
⭐ Using config "config.yaml"
⭐ Using report "validate-application"

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

✅ Validation completed (24 ok, 0 warning, 0 problem)
```

See the [validate documentation](docs/validate.md) for more info.

Please see [Documentation](#documentation) to learn more.

## Documentation

Visit the docs below to learn about *ramenctl* commands:

- [init](docs/init.md)
- [test](docs/test.md)
- [validate](docs/validate.md)
- [gather](docs/gather.md)

Check the guides below to learn more:

- [Testing disaster recovery with ramenctl](docs/testing.md)

## Contributing

- For reporting bugs, suggesting improvements, or requesting new features,
  please open an [issue](https://github.com/RamenDR/ramenctl/issues).
- For implementing features or fixing bugs, please see the
  [ramenctl contribution guide](CONTRIBUTING.md)

## License

*ramenctl* is under the [Apache 2.0 license](LICENSE).
