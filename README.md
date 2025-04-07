<!--
SPDX-FileCopyrightText: The RamenDR authors
SPDX-License-Identifier: Apache-2.0
-->

# ramenctl

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

Download the *ramenctl* executable for your operating system and
architecture and install in the PATH.

To install the latest release on Linux and macOS, run:

```console
tag="$(curl -fsSL https://api.github.com/repos/ramendr/ramenctl/releases/latest | jq -r .tag_name)"
os="$(uname | tr '[:upper:]' '[:lower:]')"
machine="$(uname -m)"
if [ "$machine" = "aarch64" ]; then machine="arm64"; fi
if [ "$machine" = "x86_64" ]; then machine="amd64"; fi
curl --location --fail --silent --show-error --output ramenctl \
    https://github.com/ramendr/ramenctl/releases/download/$tag/ramenctl-$tag-$os-$machine
sudo install ramenctl /usr/local/bin/
rm ramenctl
```




## Status

The project is in early development.

## Contributing

- For reporting bugs, suggesting improvements, or requesting new
  features, please open an
  [issue](https://github.com/RamenDR/ramenctl/issues).
- For implementing features or fixing bugs, please see the
  [ramenctl contribution guide](CONTRIBUTING.md)

## License

*ramenctl* is under the [Apache 2.0 license](LICENSE).
