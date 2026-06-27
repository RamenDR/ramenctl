<!--
SPDX-FileCopyrightText: The RamenDR authors
SPDX-License-Identifier: Apache-2.0
-->

# Pinning GitHub Actions

We pin GitHub Actions to specific commit hashes to prevent supply chain attacks.
A compromised or mutable tag (like `v6`) could point to different code at any
time, but a commit hash is immutable.

## Format

Pin actions using the full commit hash with the full version in a comment:

```yaml
- uses: actions/checkout@df4cb1c069e1874edd31b4311f1884172cec0e10 # v6.0.3
```

> [!WARNING]
> Always use a full version tag (e.g., `v6.0.3`) in the comment, not a major
> version tag (e.g., `v6`). Major version tags are mutable — maintainers move
> them to the latest release. If you look up a hash from `v6`, you may get the
> hash after an attacker has already replaced the tag. Full version tags are
> immutable by convention and safer to verify.

## Verifying hashes

To verify that a pinned hash matches the expected version tag, use
`git ls-remote`:

```console
$ git ls-remote --tags https://github.com/OWNER/REPO.git TAG 'TAG^{}'
```

For a lightweight tag, only one hash is shown:

```console
$ git ls-remote --tags https://github.com/actions/checkout.git v6.0.3 'v6.0.3^{}'
df4cb1c069e1874edd31b4311f1884172cec0e10	refs/tags/v6.0.3
```

For an annotated tag, two hashes are shown: the tag object hash and the commit
hash it points to (also known as the "peeled" tag, shown as `TAG^{}`):

```console
$ git ls-remote --tags https://github.com/golangci/golangci-lint-action.git v9.2.1 'v9.2.1^{}'
db582008a42febd596419635a5abc9d9815daa9c	refs/tags/v9.2.1
82606bf257cbaff209d206a39f5134f0cfbfd2ee	refs/tags/v9.2.1^{}
```

If two hashes are shown, use the commit hash (`TAG^{}`). The tag object hash can
change if the tag is recreated, but the commit hash is truly immutable.

## Updating hashes

When updating a pinned action:

1. Find the latest full version tag on the action's releases page.

1. Look up the commit hash for the tag:

   ```console
   $ git ls-remote --tags https://github.com/OWNER/REPO.git TAG 'TAG^{}'
   ```

1. If two hashes are shown, use the peeled hash (`TAG^{}`). Otherwise use the
   single hash shown.

1. Update the hash and the version comment in the workflow file.
