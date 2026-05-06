<!-- SPDX-FileCopyrightText: The RamenDR authors -->
<!-- SPDX-License-Identifier: Apache-2.0 -->

# ramenctl failover

The failover command manages application failover operations for disaster recovery
protected applications.

```console
$ ramenctl failover -h
Manage application failover operations

Usage:
  ramenctl failover [command]

Flags:
      --abort         abort the dry-run failover test and revert to original state
      --dry-run       perform dry-run failover test (required)
  -n, --name string       name of the DRPlacementControl resource
      --namespace string  namespace of the DRPlacementControl resource
  -o, --output string     output directory for test report (default: ./failover-dry-run-<name>-<timestamp>)

Flags:
  -h, --help   help for failover

Global Flags:
  -c, --config string   configuration file (default "config.yaml")

Use "ramenctl failover [command] --help" for more information about a command.
```

> [!IMPORTANT]
> The failover command requires a configuration file. See [init](docs/init.md) to
> learn how to create one.

## failover --dry-run

The failover dry-run command tests failover to the secondary cluster while
keeping the primary application running. This allows you to verify DR readiness,
but has significant implications for data synchronization.

### What is a dry-run failover?

A dry-run failover is a test that:
- Starts the application on the secondary cluster
- Keeps the primary application running (creating a temporary split-brain)
- Validates that failover would work in a real disaster
- Can be aborted, but requires full data resynchronization afterward

> [!WARNING]
> **Data Sync Implications**: Aborting a dry-run creates a split-brain scenario
> where both clusters had the application running. After abort, a full data sync
> is required from the primary to secondary cluster to ensure consistency. This
> sync can take significant time depending on data volume.

> [!WARNING]
> **Not Risk-Free**: While the primary application continues running during the
> test, the abort operation requires full resynchronization of all data, which
> can impact performance and take considerable time.

### Preconditions

Before running a dry-run failover, the following must be true:

1. **No active action**: The DRPC must not have an ongoing action (empty `spec.action`)
2. **Progression completed**: The DRPC `status.progression` must be `Completed` (not stuck in cleanup or other operation)
3. **Ramen version**: Must have Ramen with dry-run support (v0.17.0+)

The command validates all preconditions and fails with a clear error if not met. 
If already in dry-run mode, the command continues the existing dry-run instead of failing.

> [!NOTE]
> Dry-run is only supported with `action: Failover`. There is no dry-run mode for relocate operations.

### Looking up applications

To run the dry-run failover command, we need to find the protected application
name and namespace. Run the following command:

```console
$ kubectl get drpc -A --context hub
NAMESPACE   NAME                AGE     PREFERREDCLUSTER   FAILOVERCLUSTER   DESIREDSTATE   CURRENTSTATE
argocd      appset-deploy-rbd   6m16s   dr1                                                 Deployed
```

### Starting a dry-run failover test

To test failover for the application `appset-deploy-rbd` in namespace `argocd`:

```console
$ ramenctl failover --dry-run --name appset-deploy-rbd --namespace argocd
validating config
failover dry run
```

This starts the application on the secondary cluster while keeping the primary
running, allowing you to verify that failover would work in a real disaster.
The command waits for the test to complete and generates a test report.

The report is automatically saved to `./failover-dry-run-<name>-<timestamp>/`.
You can specify a custom location with `-o`:

```console
$ ramenctl failover --dry-run --name appset-deploy-rbd --namespace argocd -o my-report
```

### Checking the test report

Examine the test results in the auto-generated output directory:

```console
$ tree failover-dry-run-appset-deploy-rbd-20260415-143022
failover-dry-run-appset-deploy-rbd-20260415-143022
├── failover-dry-run.log
└── failover-dry-run.yaml
```

The YAML report contains the test execution details and timing information.

### Aborting a dry-run failover test

To abort the dry-run test and return the application to its original state:

```console
$ ramenctl failover --dry-run --abort --name appset-deploy-rbd --namespace argocd
validating config
abort dry run
```

The abort operation stops the test and returns the application to the state it was
in before the dry-run started. After abort completes, a full data resynchronization
occurs from the primary to secondary cluster.

## Use cases

### Testing DR readiness

Users can use this feature to test failover in advance and verify that they are prepared for a real disaster.

```console
# Test failover (report auto-generated in ./failover-dry-run-my-app-<timestamp>/)
$ ramenctl failover --dry-run --name my-app --namespace argocd

# Review results
$ ls failover-dry-run-my-app-*/
$ cat failover-dry-run-my-app-*/failover-dry-run.yaml

# Clean up
$ ramenctl failover --dry-run --abort --name my-app --namespace argocd
```

## Troubleshooting

### Error: "dry-run failover is not supported"

Your Ramen installation does not support dry-run failover. This feature requires
Ramen v0.17.0 or later. Upgrade Ramen to use this command.

### Error: "DRPC has active action"

The DRPC has an ongoing failover or relocate operation. Wait for it to complete
or cancel it before starting a dry-run.

### Dry-run stuck or taking too long

The command waits up to 10 minutes for completion. If stuck:

1. Check DRPC status:
   ```console
   $ kubectl get drpc my-app -n argocd --context hub -o yaml
   ```

2. Check Ramen operator logs:
   ```console
   $ kubectl logs -n ramen-system deployment/ramen-hub-operator --context hub
   ```

3. Cancel the operation with Ctrl+C and abort:
   ```console
   $ ramenctl failover --dry-run --abort --name my-app --namespace argocd
   ```

## Risks and Implications

**Split-Brain During Dry-Run**: Both clusters run the application as PRIMARY and write data 
independently. This creates a true split-brain scenario where data diverges between sites.

**Full Resynchronization Required on Abort**: When aborting the dry-run:
- All data written on the secondary during the test is discarded
- The entire dataset must be resynchronized from primary to secondary
- This operation consumes significant time, network bandwidth, and may impact performance
- There is no shortcut - the full resync is unavoidable

Use this command only when you understand and accept the resync cost for your data volume.

## Comparison with actual failover

| Feature | Dry-Run Failover | Actual Failover |
|---------|------------------|-----------------|
| Primary app | Keeps running | Stopped |
| Secondary app | Started | Started |
| Data sync | Read-only on secondary | Read-write on secondary |
| Production impact | None | Full failover |
| Reversible | Yes (via abort) | Requires relocate |
| Use case | Testing | Disaster recovery |
| `dryRun` flag | `true` | `false` |
