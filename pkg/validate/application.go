// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"context"
	"errors"
	"fmt"
	"slices"

	ramenapi "github.com/ramendr/ramen/api/v1alpha1"
	e2etypes "github.com/ramendr/ramen/e2e/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/ramen"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/time"
)

func (c *Command) Application(drpcName, drpcNamespace string) error {
	c.report.Application = &report.Application{
		Name:      drpcName,
		Namespace: drpcNamespace,
	}
	if !c.validateConfig() {
		return c.failed()
	}
	if !c.validateApplication(drpcName, drpcNamespace) {
		return c.failed()
	}
	c.passed()
	return nil
}

func (c *Command) validateApplication(drpcName, drpcNamespace string) bool {
	console.Step("Validate application")
	c.startStep("validate application")

	namespaces, ok := c.inspectApplication(drpcName, drpcNamespace)
	if !ok {
		return c.finishStep()
	}

	c.report.Namespaces = namespaces

	options := gathering.Options{
		Namespaces: namespaces,
		OutputDir:  c.dataDir(),
	}
	if !c.gatherNamespaces(options) {
		return c.finishStep()
	}

	if !c.validateGatheredApplicationData(drpcName, drpcNamespace) {
		return c.finishStep()
	}

	c.finishStep()
	return true
}

func (c *Command) inspectApplication(drpcName, drpcNamespace string) ([]string, bool) {
	start := time.Now()
	step := &report.Step{Name: "inspect application"}
	c.Logger().Infof("Step %q started", step.Name)

	namespaces, err := c.backend.ApplicationNamespaces(c, drpcName, drpcNamespace)
	if err != nil {
		step.Duration = time.Since(start).Seconds()
		if errors.Is(err, context.Canceled) {
			console.Error("Canceled %s", step.Name)
			step.Status = report.Canceled
		} else {
			console.Error("Failed to %s", step.Name)
			step.Status = report.Failed
		}
		c.Logger().Errorf("Step %q %s: %s", c.current.Name, step.Status, err)
		c.current.AddStep(step)

		return nil, false
	}

	step.Duration = time.Since(start).Seconds()
	step.Status = report.Passed
	c.current.AddStep(step)

	console.Pass("Inspected application")
	c.Logger().Infof("Step %q passed", step.Name)

	// For consistent gather order and report.
	slices.Sort(namespaces)

	return namespaces, true
}

func (c *Command) validateGatheredApplicationData(drpcName, drpcNamespace string) bool {
	log := c.Logger()

	start := time.Now()
	step := &report.Step{Name: "validate data"}
	defer func() {
		step.Duration = time.Since(start).Seconds()
		c.current.AddStep(step)
	}()

	s := &report.ApplicationStatus{}
	c.report.ApplicationStatus = s

	drpc, err := c.validateHub(&s.Hub, drpcName, drpcNamespace)
	if err != nil {
		step.Status = report.Failed
		console.Error("Failed to validate hub")
		log.Error(err)
		return false
	}

	if err := c.validatePrimaryCluster(&s.PrimaryCluster, drpc); err != nil {
		step.Status = report.Failed
		console.Error("Failed to validate primary cluster")
		log.Error(err)
		return false
	}

	if err := c.validateSecondaryCluster(&s.SecondaryCluster, drpc); err != nil {
		step.Status = report.Failed
		console.Error("Failed to validate primary cluster")
		log.Error(err)
		return false
	}

	if c.report.Summary.HasProblems() {
		step.Status = report.Failed
		msg := "Problems found during validation"
		console.Error(msg)
		log.Errorf("%s: %s", msg, c.report.Summary)
		return false
	}

	step.Status = report.Passed
	console.Pass("Application validated")
	return true
}

func (c *Command) validateHub(
	s *report.HubApplicationStatus,
	drpcName, drpcNamespace string,
) (*ramenapi.DRPlacementControl, error) {
	log := c.Logger()
	reader := c.outputReader(c.Env().Hub.Name)
	drpc, err := ramen.ReadDRPC(reader, drpcName, drpcNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to read drpc: %w", err)
	}
	log.Debugf("Read drpc \"%s/%s\"", drpc.Namespace, drpc.Name)
	c.validateDRPC(&s.DRPC, drpc)
	return drpc, nil
}

func (c *Command) validatePrimaryCluster(
	s *report.ClusterApplicationStatus,
	drpc *ramenapi.DRPlacementControl,
) error {
	cluster, err := ramen.PrimaryCluster(c, drpc)
	if err != nil {
		return fmt.Errorf("failed to find primary cluster: %w", err)
	}
	s.Name = cluster.Name
	return c.validateVRG(&s.VRG, cluster, drpc)
}

func (c *Command) validateSecondaryCluster(
	s *report.ClusterApplicationStatus,
	drpc *ramenapi.DRPlacementControl,
) error {
	cluster, err := ramen.SecondaryCluster(c, drpc)
	if err != nil {
		return fmt.Errorf("failed to find secondary cluster: %w", err)
	}
	s.Name = cluster.Name
	return c.validateVRG(&s.VRG, cluster, drpc)
}

func (c *Command) validateDRPC(
	s *report.DRPCSummary,
	drpc *ramenapi.DRPlacementControl,
) {
	s.Name = drpc.Name
	s.Namespace = drpc.Namespace
	s.Deleted = isDeleted(drpc)
	s.DRPolicy = drpc.Spec.DRPolicyRef.Name
	s.Action = string(drpc.Spec.Action)
	s.Phase = string(drpc.Status.Phase)
	s.Progression = string(drpc.Status.Progression)
	s.Conditions = c.validatedDRPCConditions(drpc)
}

func (c *Command) validateVRG(
	s *report.VRGSummary,
	cluster *e2etypes.Cluster,
	drpc *ramenapi.DRPlacementControl,
) error {
	log := c.Logger()

	reader := c.outputReader(cluster.Name)
	vrg, err := ramen.ReadVRG(reader, drpc.Name, ramen.VRGNamespace(drpc))
	if err != nil {
		// TODO: present missing vrg in the report instead of failing. Can happen in early
		// deployment if ramen deployment is down.
		return fmt.Errorf("failed to read vrg from cluster %q: %w", cluster.Name, err)
	}

	log.Debugf("Read vrg \"%s/%s\" from cluster %q", vrg.Namespace, vrg.Name, cluster.Name)

	s.Name = vrg.Name
	s.Namespace = vrg.Namespace
	s.Conditions = c.validatedVRGConditions(drpc, vrg)
	s.ProtectedPVCs = c.validatedProtectedPVCs(cluster, drpc, vrg)

	// TODO: Mark as an error if unknown or not primary on the primary cluster.
	s.State = string(vrg.Status.State)

	return nil
}

func (c *Command) validatedProtectedPVCs(
	cluster *e2etypes.Cluster,
	drpc *ramenapi.DRPlacementControl,
	vrg *ramenapi.VolumeReplicationGroup,
) []report.ProtectedPVCSummary {
	log := c.Logger()
	reader := c.outputReader(cluster.Name)

	var protectedPVCs []report.ProtectedPVCSummary
	for i := range vrg.Status.ProtectedPVCs {
		ppvc := &vrg.Status.ProtectedPVCs[i]
		ps := report.ProtectedPVCSummary{
			Name:        ppvc.Name,
			Namespace:   ppvc.Namespace,
			Replication: c.protectedPVCReplication(ppvc),
			Conditions:  c.validatedProtectedPVCConditions(drpc, vrg, ppvc),
		}

		if pvc, err := readPVC(reader, ppvc.Name, ppvc.Namespace); err != nil {
			log.Warnf("failed to read pvc \"%s/%s\" from cluster %q: %s",
				ppvc.Namespace, ppvc.Name, cluster.Name, err)
		} else {
			log.Debugf("Read pvc \"%s/%s\" from cluster %q", pvc.Namespace, pvc.Name, cluster.Name)
			ps.Phase = string(pvc.Status.Phase)
			ps.Deleted = isDeleted(pvc)
		}

		protectedPVCs = append(protectedPVCs, ps)
	}

	return protectedPVCs
}

func (c *Command) validatedDRPCConditions(
	drpc *ramenapi.DRPlacementControl,
) []report.ValidatedCondition {
	var conditions []report.ValidatedCondition
	for i := range drpc.Status.Conditions {
		condition := &drpc.Status.Conditions[i]
		validated := validatedCondition(drpc, condition, metav1.ConditionTrue)
		c.report.Summary.Add(&validated)
		conditions = append(conditions, validated)
	}
	return conditions
}

func (c *Command) validatedVRGConditions(
	drpc *ramenapi.DRPlacementControl,
	vrg *ramenapi.VolumeReplicationGroup,
) []report.ValidatedCondition {
	var conditions []report.ValidatedCondition
	for i := range vrg.Status.Conditions {
		condition := &vrg.Status.Conditions[i]
		// On the secondary cluster most conditions are unused.
		if condition.Reason == ramen.VRGConditionReasonUnused {
			continue
		}
		var validated report.ValidatedCondition
		if condition.Type == ramen.VRGConditionTypeDataProtected {
			validated = validatedDataProtectedCondition(drpc, vrg, condition)
		} else {
			validated = validatedCondition(vrg, condition, metav1.ConditionTrue)
		}
		c.report.Summary.Add(&validated)
		conditions = append(conditions, validated)
	}
	return conditions
}

func (c *Command) protectedPVCReplication(ppvc *ramenapi.ProtectedPVC) report.ReplicationType {
	// TODO: report external replication.
	if ppvc.ProtectedByVolSync {
		return report.Volsync
	}
	return report.Volrep
}

func (c *Command) validatedProtectedPVCConditions(
	drpc *ramenapi.DRPlacementControl,
	vrg *ramenapi.VolumeReplicationGroup,
	ppvc *ramenapi.ProtectedPVC,
) []report.ValidatedCondition {
	var conditions []report.ValidatedCondition
	for i := range ppvc.Conditions {
		condition := &ppvc.Conditions[i]
		var validated report.ValidatedCondition
		if condition.Type == ramen.VRGConditionTypeDataProtected {
			validated = validatedDataProtectedCondition(drpc, vrg, condition)
		} else {
			validated = validatedCondition(vrg, condition, metav1.ConditionTrue)
		}
		c.report.Summary.Add(&validated)
		conditions = append(conditions, validated)
	}
	return conditions
}

// validatedDataProtectedCondition returns the status for the special DataProtected contion. The
// status depends on the action. Most of the time the expected status is False, but it should be
// True before the application is placed on the secondary cluster, and it becomes False when we
// start to replicate again from the secondary cluster to the primary.
func validatedDataProtectedCondition(
	drpc *ramenapi.DRPlacementControl,
	vrg *ramenapi.VolumeReplicationGroup,
	condition *metav1.Condition,
) report.ValidatedCondition {
	// TODO: Needs testing and probably limit to some progression values, but progression values are
	// undocumnted.
	if drpc.Spec.Action == ramenapi.ActionRelocate && drpc.Status.Phase == ramenapi.Relocating {
		return validatedCondition(vrg, condition, metav1.ConditionTrue)
	}
	return validatedCondition(vrg, condition, metav1.ConditionFalse)
}
