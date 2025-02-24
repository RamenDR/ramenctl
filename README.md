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

The project will provides:

- The *ramenctl* command line tool.
- The *ramenctl* Go module for integrating the commands in other
  projects.

## Status

The project is in early development.

## Contributing

- For reporting bugs, suggesting improvements, or requesting new
  features, please open an
  [issue](https://github.com/RamenDR/ramenctl/issues).
- For implementing features or fixing bugs, please see the
  [ramenctl contribution guide](CONTRIBUTING.md)
- For testing *ramenctl* with your environment please download the
  [latest release](https://github.com/RamenDR/ramenctl/releases/latest).
- To integrated *ramenctl* with your project please check the
  [module documentation](https://pkg.go.dev/github.com/ramendr/ramenctl)

## License

*ramenctl* is under the [Apache 2.0 license](/LICENSE).
