# ramenctl test

The test command tests disaster recovery flow to validate that an application
can be protected by *ramen*, fail over to to the secondary cluster, or relocated
between the clusters.

```console
$ ramenctl test -h
Test disaster recovery flow in your clusters

Usage:
  ramenctl test [command]

Available Commands:
  clean       Delete test artifacts
  run         Run disaster recovery flow

Flags:
  -h, --help            help for test
  -o, --output string   output directory

Global Flags:
  -c, --config string   configuration file (default "config.yaml")

Use "ramenctl test [command] --help" for more information about a command.
```

The command supports the following sub-commands:

* [run](#test-run)
* [clean](#test-clean)

> [!IMPORTANT]
> The test command requires a configuration file. See [init](docs/init.md) to
> learn how to create one.

## test run

The run command runs a disaster recovery flow with one or more tiny applications
specified in the configuration file.

```console
$ ramenctl test run -o test
⭐ Using report "test"
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

The command stores `test-run.yaml` and `test-run.log` in the specified output
directory:

```console
$ tree test
test
├── test-run.log
└── test-run.yaml
```

> [!IMPORTANT]
> When reporting DR related issues, please create an archive with the output
> directory and upload it to the issue tracker.

To clean up after the test use the [clean](#test-clean) command.

## test clean

The clean command delete resources created by the [run](#test-run) command.

```console
$ ramenctl test clean -o test
⭐ Using report "test"
⭐ Using config "config.yaml"

🔎 Validate config ...
   ✅ Config validated

🔎 Clean tests ...
   ✅ Application "appset-deploy-rbd" unprotected
   ✅ Application "appset-deploy-rbd" undeployed

🔎 Clean environment ...
   ✅ Environment cleaned

✅ passed (1 passed, 0 failed, 0 skipped)
```

The command stores `test-clean.yaml` and `test-clean.log` in the specified
output directory:

```bash
$ tree test
test
├── test-clean.log
├── test-clean.yaml
├── test-run.log
└── test-run.yaml
```
