<!--
SPDX-FileCopyrightText: The RamenDR authors
SPDX-License-Identifier: Apache-2.0
-->

# ramenctl contribution guide

We accept contributions via GitHub pull requests. This document outlines
some of the conventions related to development workflow to make it
easier to get your contribution accepted.

## Getting Started

1. Fork this repository on GitHub
1. Run the tests
1. Play with the *ramenctl* tool
1. Open issues, create pull requests

## Contribution Flow

This is a rough outline of what a contributor's workflow looks like:

1. Create a branch from where you want to base your work (usually main).
1. Make your changes, keeping every commit focused on one logical
   change.
1. Make sure your commit messages are in the proper format (see below).
1. Push your changes to the branch in your fork of the repository.
1. Make sure all tests pass, and add any new tests as appropriate.
1. Create a pull request in the original repository.

## Commit message

We follow a rough convention for commit messages that is designed to
answer two questions: what changed and why. The subject line should
feature the what and the body of the commit should describe the why.

```
commands/test: Initial implementation

A minimal implementation of `ramenctl test` working on with ramen
testing environment. We use the ramen/e2e[1] module to deploy a workload
and perform a complete DR flow. This can be used to verify that a
cluster is working properly.

[1] https://github.com/RamenDR/ramen/tree/main/e2e
```

The first line is the subject and should be no longer than 70
characters, the second line is always blank, and other lines should be
wrapped at 80 characters.  This allows the message to be easier to read
on GitHub as well as in various git tools.

## Certificate of Origin

By contributing to this project you agree to the Developer Certificate
of Origin (DCO). This document was created by the Linux Kernel community
and is a simple statement that you, as a contributor, have the legal
right to make the contribution. See the [DCO](DCO) file for details.

Contributors sign-off that they adhere to these requirements by adding a
Signed-off-by line to commit messages. For example:

```
This is my commit message

Signed-off-by: My Name <myname@example.org>
```

You can append this automatically to the commit message using:

```console
git commit -s
```

If you have already made a commit and forgot to include the sign-off,
you can amend your last commit to add the sign-off using:

```console
git commit --amend -s
```

## Coding Style

The *ramenctl* project is written in the Go programming language and follows the
style guidelines of gofmt, gci, and golines tools.

If your editor supports the [EditorConfig](https://editorconfig.org/) file it
will format the code properaly automatically. For some editors you will need to
install an [EditorConfig plugin](https://editorconfig.org/#download).

To format the code automatically before committing a change run:

```console
make fmt
```
