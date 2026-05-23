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

Working with a complicated Kubernetes cluster is not easy.  In a typical
disaster recovery environment we have at least 3 connected Kubernetes
clusters with many components. The *ramenctl* project aims to make it
easier to manage and troubleshoot this challenging environment.

## Features

The project provides:

- The *ramenctl* command line tool, managing and troubleshooting ramen.
- The *ramenctl* Go module for integrating the ramenctl commands in other
  projects. This module is used to implement the
  [odf dr](https://github.com/red-hat-storage/odf-cli/blob/main/docs/dr.md) command.

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

Run disaster recovery tests:

```console
$ ramenctl test run -o rdr-test
⭐ Using report "rdr-test"
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

Your system is ready for disaster recovery!

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

- For reporting bugs, suggesting improvements, or requesting new
  features, please open an
  [issue](https://github.com/RamenDR/ramenctl/issues).
- For implementing features or fixing bugs, please see the
  [ramenctl contribution guide](CONTRIBUTING.md)

## License

*ramenctl* is under the [Apache 2.0 license](LICENSE).
