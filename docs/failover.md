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

Available Commands:
  dry-run     Test failover without affecting the primary application (DRY-RUN mode)

Flags:
  -h, --help   help for failover

Global Flags:
  -c, --config string   configuration file (default "config.yaml")

Use "ramenctl failover [command] --help" for more information about a command.
```

> [!IMPORTANT]
> The failover command requires a configuration file. See [init](docs/init.md) to
> learn how to create one.

> [!IMPORTANT]
> This command requires Ramen PR [#2416](https://github.com/RamenDR/ramen/pull/2416)
> to be merged. The `DryRun` field is not yet available in the current Ramen API version.

## failover dry-run

The failover dry-run command tests failover to the secondary cluster without
affecting the primary application. This allows you to verify DR readiness without
risk to production workloads.

### What is a dry-run failover?

A dry-run failover is a non-destructive test that:
- Starts the application on the secondary cluster
- Keeps the primary application running
- Validates that failover would work in a real disaster
- Can be safely aborted and reverted

This is achieved by setting `dryRun: true` in the DRPC spec along with
`action: Failover`.

### Looking up applications

To run the dry-run failover command, we need to find the protected application
name and namespace. Run the following command:

```console
$ kubectl get drpc -A --context hub
NAMESPACE   NAME                AGE     PREFERREDCLUSTER   FAILOVERCLUSTER   DESIREDSTATE   CURRENTSTATE
argocd      appset-deploy-rbd   6m16s   dr1                                                 Deployed
```

### Starting a dry-run failover test

To test failover for the application `appset-deploy-rbd` in namespace `argocd`,
run the following command:

```console
$ ramenctl failover dry-run --name appset-deploy-rbd --namespace argocd -o dry-run-test
⭐ Using config "config.yaml"
🔎 Starting DRY-RUN failover test

🧪 DRY-RUN MODE: Testing failover to cluster "dr2" without affecting primary
✅ DRY-RUN failover triggered on cluster "dr2"

🔎 Waiting for DRY-RUN to complete (this may take several minutes)
✅ DRY-RUN: Application "appset-deploy-rbd" is available on cluster "dr2"
✅ DRY-RUN: Primary application remains on original cluster

✅ DRY-RUN failover test passed

💡 To abort this dry-run: ramenctl failover dry-run --abort --name appset-deploy-rbd --namespace argocd
```

The command will:
1. Validate preconditions (not already in dry-run, no active action)
2. Determine the secondary cluster from the DR peer relationship
3. Update the DRPC to trigger dry-run failover
4. Wait for the application to become available on the secondary cluster
5. Generate a test report if `-o` option is provided

### Checking the test report

If you specified an output directory with `-o`, you can examine the test results:

```console
$ tree dry-run-test
dry-run-test
├── failover-dry-run.log
└── failover-dry-run.yaml
```

The YAML report contains the test execution details and timing information.

### Aborting a dry-run failover test

To abort the dry-run test and return the application to its original state:

```console
$ ramenctl failover dry-run --abort --name appset-deploy-rbd --namespace argocd
🔎 Aborting DRY-RUN failover test

⚠️  Aborting dry-run failover for "appset-deploy-rbd"
✅ DRY-RUN aborted

🔎 Waiting for application to return to original state
✅ Application "appset-deploy-rbd" restored to original state

✅ DRY-RUN abort completed
```

The abort command will:
1. Verify the DRPC is in dry-run mode
2. Read the `last-action` label to determine the original state
3. Restore the DRPC spec to its pre-dry-run configuration
4. Wait for the application to return to the original phase

### How abort restores state

The abort logic uses Ramen's `last-action` label to intelligently restore the
DRPC to its state before the dry-run:

| Original State | last-action label | Restored DRPC Spec |
|----------------|-------------------|-------------------|
| Deployed | `""` (empty) | `action=""`, `failoverCluster=""`, `dryRun=false` |
| FailedOver | `"Failover"` | `action="Failover"`, `failoverCluster=preferredCluster`, `dryRun=false` |
| Relocated | `"Relocate"` | `action="Relocate"`, `failoverCluster=""`, `dryRun=false` |

**Important**: The `last-action` label is NOT updated during dry-run (per Ramen
PR #2416), which allows safe state restoration.

## Use cases

### Testing DR readiness

Before a real disaster, test that failover will work:

```console
# Test failover
$ ramenctl failover dry-run --name my-app --namespace argocd -o test-$(date +%Y%m%d)

# Review results
$ cat test-20260406/failover-dry-run.yaml

# Clean up
$ ramenctl failover dry-run --abort --name my-app --namespace argocd
```

### Periodic DR drills

Schedule regular dry-run tests to ensure DR readiness:

```bash
#!/bin/bash
# Monthly DR drill script

APPS=("app1" "app2" "app3")
NAMESPACE="argocd"
REPORT_DIR="dr-drill-$(date +%Y%m)"

for app in "${APPS[@]}"; do
    echo "Testing $app..."
    ramenctl failover dry-run --name "$app" --namespace "$NAMESPACE" -o "$REPORT_DIR/$app"
    sleep 60  # Wait between tests
    ramenctl failover dry-run --abort --name "$app" --namespace "$NAMESPACE"
done
```

### Pre-migration validation

Before performing an actual failover or relocation, verify it will work:

```console
# Test first
$ ramenctl failover dry-run --name critical-app --namespace prod

# If test passes, perform actual failover
$ kubectl patch drpc critical-app -n prod --type merge -p '{"spec":{"action":"Failover","failoverCluster":"dr2"}}'
```

## Troubleshooting

### Error: "DRPC is already in dry-run mode"

You're trying to start a dry-run when one is already active. Abort first:

```console
$ ramenctl failover dry-run --abort --name my-app --namespace argocd
```

### Error: "DRPC has active action"

The DRPC has an ongoing failover or relocate operation. Wait for it to complete
or cancel it before starting a dry-run.

### Dry-run stuck or taking too long

The command waits up to 10 minutes for completion. If stuck:

1. Check DRPC status:
   ```console
   $ kubectl get drpc my-app -n argocd -o yaml
   ```

2. Check Ramen operator logs:
   ```console
   $ kubectl logs -n ramen-system deployment/ramen-hub-operator
   ```

3. Cancel the operation with Ctrl+C and abort:
   ```console
   $ ramenctl failover dry-run --abort --name my-app --namespace argocd
   ```

## Safety features

The dry-run failover command includes several safety features:

1. **Precondition validation**: Prevents starting if already in dry-run or has active action
2. **Non-destructive**: Primary application continues running during test
3. **State preservation**: Uses `last-action` label for safe abort
4. **Timeout protection**: 10-minute timeout prevents indefinite hangs
5. **Error detection**: Monitors DRPC conditions for failures
6. **Cancellation support**: Handles Ctrl+C gracefully

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
