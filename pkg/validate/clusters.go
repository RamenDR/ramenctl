// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"fmt"

	ramenapi "github.com/ramendr/ramen/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/ramen"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/sets"
	"github.com/ramendr/ramenctl/pkg/time"
)

func (c *Command) Clusters() error {
	if !c.validateConfig() {
		return c.failed()
	}
	if !c.validateClusters() {
		return c.failed()
	}
	c.passed()
	return nil
}

func (c *Command) validateClusters() bool {
	console.Step("Validate clusters")
	c.startStep("validate clusters")

	namespaces := c.clustersNamespacesToGather()
	c.report.Namespaces = namespaces

	options := gathering.Options{
		Namespaces: namespaces,
		Cluster:    true,
		OutputDir:  c.dataDir(),
	}
	if !c.gatherNamespaces(options) {
		return c.finishStep()
	}

	if !c.validateGatheredClustersData() {
		return c.finishStep()
	}

	c.finishStep()
	return true
}

func (c *Command) clustersNamespacesToGather() []string {
	return sets.Sorted([]string{
		c.config.Namespaces.RamenHubNamespace,
		c.config.Namespaces.RamenDRClusterNamespace,
	})
}

func (c *Command) validateGatheredClustersData() bool {
	log := c.Logger()

	start := time.Now()
	step := &report.Step{Name: "validate clusters data"}
	defer func() {
		step.Duration = time.Since(start).Seconds()
		c.current.AddStep(step)
	}()

	s := &report.ClustersStatus{}
	c.report.ClustersStatus = s

	err := c.validateClustersHub(&s.Hub)
	if err != nil {
		step.Status = report.Failed
		console.Error("Failed to validate hub")
		log.Error(err)
		return false
	}

	step.Status = report.Passed
	console.Pass("Clusters validated")
	return true
}

func (c *Command) validateClustersHub(s *report.ClustersStatusHub) error {
	if err := c.validateClustersDRPolicies(&s.DRPolicies); err != nil {
		return fmt.Errorf("failed to validate DRPolicies: %w", err)
	}

	if err := c.validateClustersDRClusters(&s.DRClusters); err != nil {
		return fmt.Errorf("failed to validate DRClusters: %w", err)
	}

	return nil
}

func (c *Command) validateClustersDRPolicies(
	drPoliciesList *report.ValidatedDRPoliciesList,
) error {
	log := c.Logger()
	reader := c.outputReader(c.Env().Hub.Name)

	drPolicyNames, err := ramen.ListDRPolicies(reader)
	if err != nil {
		return err
	}

	for _, policyName := range drPolicyNames {
		drPolicy, err := ramen.ReadDRPolicy(reader, policyName)
		if err != nil {
			return err
		}
		log.Debugf("Read DRPolicy %q", drPolicy.Name)

		dps := report.DRPolicySummary{
			Name:               drPolicy.Name,
			SchedulingInterval: drPolicy.Spec.SchedulingInterval,
			DRClusters:         drPolicy.Spec.DRClusters,
			Conditions:         c.validatedConditions(drPolicy, drPolicy.Status.Conditions),
		}
		drPoliciesList.Value = append(drPoliciesList.Value, dps)
	}

	if len(drPoliciesList.Value) == 0 {
		drPoliciesList.State = report.Problem
		drPoliciesList.Description = "No DRPolicies found"
	} else {
		drPoliciesList.State = report.OK
	}

	c.report.Summary.Add(drPoliciesList)

	return nil
}

func (c *Command) validateClustersDRClusters(
	drClustersList *report.ValidatedDRClustersList,
) error {
	log := c.Logger()
	reader := c.outputReader(c.Env().Hub.Name)

	drClusterNames, err := ramen.ListDRClusters(reader)
	if err != nil {
		return err
	}

	for _, drClusterName := range drClusterNames {
		drCluster, err := ramen.ReadDRCluster(reader, drClusterName)
		if err != nil {
			return err
		}
		log.Debugf("Read DRCluster %q", drCluster.Name)

		dcs := report.DRClusterSummary{
			Name:       drCluster.Name,
			Phase:      string(drCluster.Status.Phase),
			Conditions: c.validatedDRClusterConditions(drCluster),
		}
		drClustersList.Value = append(drClustersList.Value, dcs)
	}

	if len(drClustersList.Value) < 2 {
		drClustersList.State = report.Problem
		drClustersList.Description = fmt.Sprintf("2 DRClusters required %d found",
			len(drClustersList.Value))
	} else {
		drClustersList.State = report.OK
	}

	c.report.Summary.Add(drClustersList)

	return nil
}

func (c *Command) validatedDRClusterConditions(
	drCluster *ramenapi.DRCluster,
) []report.ValidatedCondition {
	var conditions []report.ValidatedCondition
	for i := range drCluster.Status.Conditions {
		condition := &drCluster.Status.Conditions[i]

		var validated report.ValidatedCondition
		if condition.Type == ramenapi.DRClusterConditionTypeFenced {
			// For Fenced condition, "False" is the expected status.
			validated = validatedCondition(drCluster, condition, metav1.ConditionFalse)
		} else {
			// For Clean & Validated conditions, "True" is the expected status.
			validated = validatedCondition(drCluster, condition, metav1.ConditionTrue)
		}

		c.report.Summary.Add(&validated)
		conditions = append(conditions, validated)
	}
	return conditions
}
