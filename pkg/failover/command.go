// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

// Package failover implements dry-run failover testing commands.
//
// Requires Ramen v0.17.0 or later with dry-run support.
//
// # Abort restore logic
//
// The abort logic uses Ramen's last-action label to restore the DRPC to its
// state before the dry-run:
//
//	Original State | last-action label | Restored DRPC Spec
//	---------------|-------------------|-------------------
//	Deployed       | "" (empty)        | action="", failoverCluster="", dryRun=false
//	FailedOver     | "Failover"        | action="Failover", failoverCluster=preferredCluster,
//	               |                   | dryRun=false
//	Relocated      | "Relocate"        | action="Relocate", failoverCluster="", dryRun=false
//
// The last-action label is NOT updated during dry-run, which allows safe state
// restoration after aborting the test.

package failover

import (
	"context"
	"errors"
	"fmt"
	stdtime "time"

	ramenapi "github.com/ramendr/ramen/api/v1alpha1"
	"github.com/ramendr/ramen/e2e/types"
	"go.uber.org/zap"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/ramen"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/time"
)

const (
	// Label key for last action set by Ramen controller
	lastActionLabel = "ramendr.openshift.io/last-action"

	// Polling interval for waiting on DRPC status changes
	pollInterval = 5 * stdtime.Second

	// Timeout for dry-run operations (similar to failover timeout in tests)
	dryRunTimeout = 10 * stdtime.Minute
)

// Command is a ramenctl failover command.
type Command struct {
	// command is the generic command used by all ramenctl commands.
	command *command.Command

	// config is the config for this command.
	config *config.Config

	// context is used to set deadlines.
	context context.Context

	// report describes the command execution (only for test dry-run).
	report *report.Report

	// current step.
	current        *report.Step
	currentStarted time.Time
}

// Ensure that command implements ramen.Context.
var _ ramen.Context = &Command{}

// ramen.Context interface.

func (c *Command) Env() *types.Env {
	return c.command.Env()
}

func (c *Command) Logger() *zap.SugaredLogger {
	return c.command.Logger()
}

func (c *Command) Context() context.Context {
	return c.context
}

func newCommand(cmd *command.Command, cfg *config.Config) *Command {
	return &Command{
		command: cmd,
		config:  cfg,
		context: cmd.Context(),
		report:  report.NewReport(cmd.Name(), cfg),
	}
}

// TestDryRun triggers a dry-run failover test for the specified DRPC.
func (c *Command) TestDryRun(drpcName, drpcNamespace string) error {
	c.report.Application = &report.Application{
		Name:      drpcName,
		Namespace: drpcNamespace,
	}

	console.Step("validating config")

	// Check if Ramen supports dry-run failover
	if err := c.checkDryRunSupport(); err != nil {
		return err
	}

	// Get current DRPC state
	drpc, err := c.getDRPC(drpcName, drpcNamespace)
	if err != nil {
		return fmt.Errorf("failed to get DRPC: %w", err)
	}

	console.Step("failover dry run")
	c.startStep("failover dry run")

	// Validate preconditions
	alreadyInDryRun, err := c.validateTestPreconditions(drpc)
	if err != nil {
		return c.failStep(err)
	}

	// If already in dry-run, skip to waiting
	if alreadyInDryRun {
		console.Info("DRPC already in dry-run mode")
	} else {
		// Calculate secondary cluster
		secondaryCluster, err := ramen.SecondaryCluster(c, drpc)
		if err != nil {
			return c.failStep(fmt.Errorf("failed to determine secondary cluster: %w", err))
		}

		c.Logger().Infof("Starting dry-run failover: DRPC=%s/%s, failoverCluster=%s",
			drpcNamespace, drpcName, secondaryCluster.Name)

		// Update DRPC to trigger dry-run
		if err := c.triggerDryRun(drpc, secondaryCluster.Name); err != nil {
			return c.failStep(fmt.Errorf("failed to trigger dry-run: %w", err))
		}
	}

	// Wait for dry-run to complete
	if err := c.waitForDryRunComplete(drpcName, drpcNamespace); err != nil {
		return c.failStep(fmt.Errorf("dry-run failed: %w", err))
	}

	c.passStep()

	// Write report if output directory was provided
	if c.command.OutputDir() != "" {
		c.command.WriteYAMLReport(c.report)
	}

	return nil
}

// AbortDryRun reverts a dry-run failover test for the specified DRPC.
func (c *Command) AbortDryRun(drpcName, drpcNamespace string) error {
	console.Step("validating config")

	// Check if Ramen supports dry-run failover
	if err := c.checkDryRunSupport(); err != nil {
		return err
	}

	// Get current DRPC state
	drpc, err := c.getDRPC(drpcName, drpcNamespace)
	if err != nil {
		return fmt.Errorf("failed to get DRPC: %w", err)
	}

	// Validate that DRPC is in dry-run mode
	if !drpc.Spec.DryRun {
		return fmt.Errorf("DRPC %s/%s is not in dry-run mode", drpcNamespace, drpcName)
	}

	console.Step("abort dry run")

	c.Logger().Infof("Aborting dry-run: DRPC=%s/%s", drpcNamespace, drpcName)

	// Get the original state from last-action label
	lastAction := drpc.Labels[lastActionLabel]
	c.Logger().Infof("Original state from last-action label: %q", lastAction)

	// Revert DRPC to original state
	if err := c.revertDryRun(drpc, lastAction); err != nil {
		return fmt.Errorf("failed to revert dry-run: %w", err)
	}

	// Wait for revert to complete
	if err := c.waitForRevertComplete(drpcName, drpcNamespace, lastAction); err != nil {
		return fmt.Errorf("revert failed: %w", err)
	}

	return nil
}

// validateTestPreconditions checks if dry-run can be started.
// Returns true if already in dry-run mode (idempotent), false if starting new dry-run.
func (c *Command) validateTestPreconditions(
	drpc *ramenapi.DRPlacementControl,
) (alreadyInDryRun bool, err error) {
	// Check if already in dry-run mode (idempotent - allow re-running)
	if drpc.Spec.DryRun {
		c.Logger().Info("DRPC is already in dry-run mode, continuing...")
		return true, nil
	}

	// Check if there's an active non-dry-run action
	if drpc.Spec.Action != "" {
		return false, fmt.Errorf(
			"DRPC has active action %q, cannot start dry-run test",
			drpc.Spec.Action,
		)
	}

	// Check if DRPC is in a completed state (not stuck in cleanup or other operation)
	if drpc.Status.Progression != ramenapi.ProgressionCompleted {
		return false, fmt.Errorf(
			"DRPC progression is %q, must be %q before starting dry-run",
			drpc.Status.Progression,
			ramenapi.ProgressionCompleted,
		)
	}

	c.Logger().Infof(
		"Preconditions validated: dryRun=%v, action=%q, progression=%q",
		drpc.Spec.DryRun,
		drpc.Spec.Action,
		drpc.Status.Progression,
	)

	return false, nil
}

// triggerDryRun updates the DRPC to start dry-run failover.
func (c *Command) triggerDryRun(drpc *ramenapi.DRPlacementControl, failoverCluster string) error {
	drpc.Spec.Action = ramenapi.ActionFailover
	drpc.Spec.FailoverCluster = failoverCluster
	drpc.Spec.DryRun = true

	return c.updateDRPC(drpc)
}

// revertDryRun restores DRPC to its pre-dry-run state based on last-action label.
func (c *Command) revertDryRun(drpc *ramenapi.DRPlacementControl, lastAction string) error {
	switch lastAction {
	case "":
		// Was in Deployed state
		drpc.Spec.Action = ""
		drpc.Spec.FailoverCluster = ""
		drpc.Spec.DryRun = false
		c.Logger().Info("Restoring to Deployed state")

	case string(ramenapi.ActionFailover):
		// Was in FailedOver state
		drpc.Spec.Action = ramenapi.ActionFailover
		drpc.Spec.FailoverCluster = drpc.Spec.PreferredCluster
		drpc.Spec.DryRun = false
		c.Logger().Infof(
			"Restoring to FailedOver state with failoverCluster=%s",
			drpc.Spec.PreferredCluster,
		)

	case string(ramenapi.ActionRelocate):
		// Was in Relocated state
		drpc.Spec.Action = ramenapi.ActionRelocate
		drpc.Spec.FailoverCluster = ""
		drpc.Spec.DryRun = false
		c.Logger().Info("Restoring to Relocated state")

	default:
		return fmt.Errorf("unknown last-action value: %q", lastAction)
	}

	return c.updateDRPC(drpc)
}

// getDRPC fetches the DRPC from the hub cluster.
func (c *Command) getDRPC(name, namespace string) (*ramenapi.DRPlacementControl, error) {
	drpc := &ramenapi.DRPlacementControl{}
	key := k8stypes.NamespacedName{Namespace: namespace, Name: name}
	err := c.Env().Hub.Client.Get(c.Context(), key, drpc)
	if err != nil {
		return nil, err
	}
	return drpc, nil
}

// checkDryRunSupport verifies that the installed Ramen version supports dry-run failover.
func (c *Command) checkDryRunSupport() error {
	crd := &apiextensionsv1.CustomResourceDefinition{}
	key := k8stypes.NamespacedName{Name: "drplacementcontrols.ramendr.openshift.io"}
	err := c.Env().Hub.Client.Get(c.Context(), key, crd)
	if err != nil {
		return fmt.Errorf("failed to get DRPlacementControl CRD: %w", err)
	}

	// Check if the CRD has a version with the dryRun field in the spec
	for _, version := range crd.Spec.Versions {
		if version.Schema == nil || version.Schema.OpenAPIV3Schema == nil {
			continue
		}

		// Navigate to spec.dryRun in the schema
		spec, ok := version.Schema.OpenAPIV3Schema.Properties["spec"]
		if !ok {
			continue
		}

		if _, hasDryRun := spec.Properties["dryRun"]; hasDryRun {
			c.Logger().Debugf("Dry-run support detected in CRD version %s", version.Name)
			return nil
		}
	}

	return fmt.Errorf(
		"dry-run failover is not supported by the installed Ramen version",
	)
}

// updateDRPC updates the DRPC on the hub cluster.
func (c *Command) updateDRPC(drpc *ramenapi.DRPlacementControl) error {
	c.Logger().Infof("Updating DRPC: action=%q, failoverCluster=%q, dryRun=%v",
		drpc.Spec.Action, drpc.Spec.FailoverCluster, drpc.Spec.DryRun)

	err := c.Env().Hub.Client.Update(c.Context(), drpc)
	if err != nil {
		return fmt.Errorf("failed to update DRPC: %w", err)
	}

	return nil
}

// waitForDryRunComplete waits for DRPC to reach ProgressionTestingFailover status.
func (c *Command) waitForDryRunComplete(name, namespace string) error {
	ctx, cancel := context.WithTimeout(c.Context(), dryRunTimeout)
	defer cancel()

	return wait.PollUntilContextCancel(
		ctx,
		pollInterval,
		true,
		func(ctx context.Context) (bool, error) {
			drpc, err := c.getDRPC(name, namespace)
			if err != nil {
				// If context is cancelled, return the cancellation error
				if errors.Is(err, context.Canceled) {
					return false, err
				}
				// For other errors, log and retry
				c.Logger().Warnf("Failed to get DRPC status (will retry): %s", err)
				return false, nil
			}

			c.Logger().Debugf("DRPC progression: %s", drpc.Status.Progression)

			// Check if dry-run completed successfully
			if drpc.Status.Progression == "ProgressionTestingFailover" {
				c.Logger().Info("DRY-RUN completed: progression=ProgressionTestingFailover")
				return true, nil
			}

			// Check for failures
			if c.hasDRPCFailed(drpc) {
				return false, fmt.Errorf(
					"DRPC entered failed state: progression=%s",
					drpc.Status.Progression,
				)
			}

			// Still progressing
			return false, nil
		},
	)
}

// waitForRevertComplete waits for DRPC to return to its original state after abort.
func (c *Command) waitForRevertComplete(name, namespace, lastAction string) error {
	ctx, cancel := context.WithTimeout(c.Context(), dryRunTimeout)
	defer cancel()

	var expectedPhase ramenapi.DRState
	switch lastAction {
	case "":
		expectedPhase = ramenapi.Deployed
	case string(ramenapi.ActionFailover):
		expectedPhase = ramenapi.FailedOver
	case string(ramenapi.ActionRelocate):
		expectedPhase = ramenapi.Relocated
	default:
		return fmt.Errorf("unknown last-action: %q", lastAction)
	}

	c.Logger().Infof("Waiting for DRPC to return to phase: %s", expectedPhase)

	return wait.PollUntilContextCancel(
		ctx,
		pollInterval,
		true,
		func(ctx context.Context) (bool, error) {
			drpc, err := c.getDRPC(name, namespace)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return false, err
				}
				c.Logger().Warnf("Failed to get DRPC status (will retry): %s", err)
				return false, nil
			}

			c.Logger().Debugf(
				"DRPC phase: %s, progression: %s",
				drpc.Status.Phase,
				drpc.Status.Progression,
			)

			// Check if returned to expected phase
			if drpc.Status.Phase == expectedPhase &&
				drpc.Status.Progression == ramenapi.ProgressionCompleted {
				c.Logger().Infof(
					"Revert completed: phase=%s, progression=%s",
					drpc.Status.Phase,
					drpc.Status.Progression,
				)
				return true, nil
			}

			// Check for failures
			if c.hasDRPCFailed(drpc) {
				return false, fmt.Errorf(
					"DRPC entered failed state during revert: progression=%s",
					drpc.Status.Progression,
				)
			}

			return false, nil
		},
	)
}

// hasDRPCFailed checks if DRPC has entered a failed state.
func (c *Command) hasDRPCFailed(drpc *ramenapi.DRPlacementControl) bool {
	// Check if Available condition is False
	availableCond := meta.FindStatusCondition(drpc.Status.Conditions, "Available")
	if availableCond != nil && availableCond.Status == metav1.ConditionFalse {
		c.Logger().Warnf("DRPC Available condition is False: %s", availableCond.Message)
		return true
	}

	// Could add more failure detection logic here if needed
	return false
}

// Step management methods (similar to validate/command.go).

func (c *Command) startStep(name string) {
	c.current = &report.Step{Name: name}
	c.currentStarted = time.Now()
	c.Logger().Infof("Step %q started", c.current.Name)
}

func (c *Command) failStep(err error) error {
	c.current.Duration = time.Since(c.currentStarted).Seconds()
	if errors.Is(err, context.Canceled) {
		c.current.Status = report.Canceled
		console.Error("Canceled %s", c.current.Name)
	} else {
		c.current.Status = report.Failed
		console.Error("Failed: %s", err)
	}
	c.Logger().Errorf("Step %q %s: %s", c.current.Name, c.current.Status, err)
	c.report.AddStep(c.current)
	c.current = nil

	// Write report on failure if output directory was provided
	if c.command.OutputDir() != "" {
		c.command.WriteYAMLReport(c.report)
	}

	return err
}

func (c *Command) passStep() {
	c.current.Duration = time.Since(c.currentStarted).Seconds()
	c.current.Status = report.Passed
	c.Logger().Infof("Step %q passed", c.current.Name)
	c.report.AddStep(c.current)
	c.current = nil
}
