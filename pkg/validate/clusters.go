// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"context"
	"errors"
	"fmt"
	"os"

	ramenapi "github.com/ramendr/ramen/api/v1alpha1"
	"github.com/ramendr/ramen/e2e/types"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/core"
	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/logging"
	"github.com/ramendr/ramenctl/pkg/ramen"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/s3"
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

	// If inspectClustersS3Profiles fails, skip S3 check but continue validation.
	// Missing S3 status will be reported as a problem during validation.
	if profiles, ok := c.inspectClustersS3Profiles(); ok {
		if !c.checkS3(profiles) {
			return c.finishStep()
		}
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

func (c *Command) inspectClustersS3Profiles() ([]*s3.Profile, bool) {
	start := time.Now()
	step := &report.Step{Name: "inspect S3 profiles"}

	c.Logger().Infof("Step %q started", step.Name)

	// Read S3 profiles from the ramen hub configmap, the source of truth
	// synced to managed clusters.
	reader := c.outputReader(c.Env().Hub.Name)
	configMapName := ramen.HubOperatorConfigMapName
	configMapNamespace := c.config.Namespaces.RamenHubNamespace

	profiles, err := ramen.ClusterProfiles(reader, configMapName, configMapNamespace)
	if err != nil {
		step.Duration = time.Since(start).Seconds()
		step.Status = report.Failed
		console.Error("Failed to %s", step.Name)
		c.Logger().Errorf("Step %q failed: %s", step.Name, err)
		c.current.AddStep(step)
		return nil, false
	}

	step.Duration = time.Since(start).Seconds()
	step.Status = report.Passed
	c.current.AddStep(step)

	console.Pass("Inspected S3 profiles")
	c.Logger().Infof("Step %q passed", step.Name)

	return profiles, true
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

	if err := c.validateClustersHub(&s.Hub); err != nil {
		step.Status = report.Failed
		msg := "Failed to validate hub"
		console.Error(msg)
		log.Errorf("%s: %s", msg, err)
		return false
	}

	if err := c.validateClustersClusters(&s.Clusters); err != nil {
		step.Status = report.Failed
		msg := "Failed to validate managed clusters"
		console.Error(msg)
		log.Errorf("%s: %s", msg, err)
		return false
	}

	c.validateClustersS3Status(&s.S3)

	if hasIssues(c.report.Summary) {
		step.Status = report.Failed
		msg := "Issues found during validation"
		console.Error(msg)
		log.Errorf("%s: %s", msg, summaryString(c.report.Summary))
		return false
	}

	step.Status = report.Passed
	console.Pass("Clusters validated")
	return true
}

func (c *Command) validateClustersHub(s *report.ClustersStatusHub) error {
	if err := c.validateClustersDRPolicies(&s.DRPolicies); err != nil {
		return fmt.Errorf("failed to validate drpolicies: %w", err)
	}

	if err := c.validateClustersDRClusters(&s.DRClusters); err != nil {
		return fmt.Errorf("failed to validate drclusters: %w", err)
	}

	hub := c.Env().Hub
	namespace := c.Config().Namespaces.RamenHubNamespace
	if err := c.validateRamen(&s.Ramen, hub, namespace, ramenapi.DRHubType); err != nil {
		return fmt.Errorf("failed to validate ramen: %w", err)
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
		return fmt.Errorf("failed to list drpolicies: %w", err)
	}

	for _, policyName := range drPolicyNames {
		drPolicy, err := ramen.ReadDRPolicy(reader, policyName)
		if err != nil {
			return fmt.Errorf("failed to read drpolicy %q: %w", policyName, err)
		}

		log.Debugf("Read drpolicy %q", drPolicy.Name)
		dps := report.DRPolicySummary{
			Name:               drPolicy.Name,
			SchedulingInterval: drPolicy.Spec.SchedulingInterval,
			DRClusters:         drPolicy.Spec.DRClusters,
			PeerClasses:        c.validatedPeerClasses(drPolicy),
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

	addValidation(c.report.Summary, drPoliciesList)

	return nil
}

func (c *Command) validatedPeerClasses(
	drPolicy *ramenapi.DRPolicy,
) report.ValidatedPeerClassesList {
	peerClassesList := report.ValidatedPeerClassesList{}

	for _, peerClass := range drPolicy.Status.Async.PeerClasses {
		pcs := report.PeerClassesSummary{
			StorageClassName: peerClass.StorageClassName,
			ReplicationID:    peerClass.ReplicationID,
			Grouping:         peerClass.Grouping,
		}
		peerClassesList.Value = append(peerClassesList.Value, pcs)
	}

	if len(peerClassesList.Value) == 0 {
		peerClassesList.State = report.Problem
		peerClassesList.Description = "No peer classes found"
	} else {
		peerClassesList.State = report.OK
	}
	addValidation(c.report.Summary, &peerClassesList)

	return peerClassesList
}

func (c *Command) validateClustersDRClusters(
	drClustersList *report.ValidatedDRClustersList,
) error {
	log := c.Logger()
	reader := c.outputReader(c.Env().Hub.Name)

	drClusterNames, err := ramen.ListDRClusters(reader)
	if err != nil {
		return fmt.Errorf("failed to list drclusters: %w", err)
	}

	for _, drClusterName := range drClusterNames {
		drCluster, err := ramen.ReadDRCluster(reader, drClusterName)
		if err != nil {
			return fmt.Errorf("failed to read drluster %q: %w", drClusterName, err)
		}

		log.Debugf("Read drcluster %q", drCluster.Name)
		dcs := report.DRClusterSummary{
			Name:       drCluster.Name,
			Phase:      string(drCluster.Status.Phase),
			Conditions: c.validatedDRClusterConditions(drCluster),
		}
		drClustersList.Value = append(drClustersList.Value, dcs)
	}

	if len(drClustersList.Value) < 2 {
		drClustersList.State = report.Problem
		drClustersList.Description = fmt.Sprintf("2 DRClusters required, %d found",
			len(drClustersList.Value))
	} else {
		drClustersList.State = report.OK
	}

	addValidation(c.report.Summary, drClustersList)

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

		addValidation(c.report.Summary, &validated)
		conditions = append(conditions, validated)
	}

	return conditions
}

func (c *Command) validateClustersClusters(s *[]report.ClustersStatusCluster) error {
	env := c.Env()
	namespace := c.Config().Namespaces.RamenDRClusterNamespace

	for _, cluster := range env.ManagedClusters() {
		cs := report.ClustersStatusCluster{Name: cluster.Name}
		if err := c.validateRamen(
			&cs.Ramen,
			cluster,
			namespace,
			ramenapi.DRClusterType,
		); err != nil {
			return fmt.Errorf("failed to validate ramen: %w", err)
		}
		*s = append(*s, cs)
	}

	return nil
}

func (c *Command) validateRamen(
	s *report.RamenSummary,
	cluster *types.Cluster,
	namespace string,
	controllerType ramenapi.ControllerType,
) error {
	deploymentName := ramen.OperatorDeploymentName(controllerType)
	configMapName := ramen.OperatorConfigMapName(controllerType)

	if err := c.validateDeployment(
		&s.Deployment,
		cluster,
		deploymentName,
		namespace,
		ramen.OperatorReplicas,
	); err != nil {
		return fmt.Errorf("failed to validate deployment: %w", err)
	}

	if err := c.validateRamenConfigMap(
		&s.ConfigMap,
		cluster,
		configMapName,
		namespace,
		controllerType,
	); err != nil {
		return fmt.Errorf("failed to validate configmap: %w", err)
	}

	return nil
}

func (c *Command) validateRamenConfigMap(
	s *report.ConfigMapSummary,
	cluster *types.Cluster,
	name, namespace string,
	controllerType ramenapi.ControllerType,
) error {
	log := c.Logger()
	reader := c.outputReader(cluster.Name)

	s.Name = name
	s.Namespace = namespace

	configMap, err := core.ReadConfigMap(reader, name, namespace)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to read configmap \"%s/%s\" from cluster %q: %w",
				namespace, name, cluster.Name, err)
		}

		log.Debugf("Configmap \"%s/%s\" does not exist in cluster %q",
			namespace, name, cluster.Name)
		s.Deleted = c.validatedDeleted(nil)
		return nil
	}

	log.Debugf("Read configmap \"%s/%s\" from cluster %q", namespace, name, cluster.Name)
	s.Deleted = c.validatedDeleted(configMap)

	config := &ramenapi.RamenConfig{}
	data := []byte(configMap.Data[ramen.ConfigMapRamenConfigKeyName])
	if err := yaml.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to unmarshal ramen config data: %w\n%s", err, data)
	}

	s.RamenControllerType = c.validatedRamenControllerType(config, controllerType)

	if err := c.validatedS3Profiles(&s.S3StoreProfiles, cluster, config, namespace); err != nil {
		return fmt.Errorf("failed to validate s3 profiles: %w", err)
	}

	// TODO: Validate that configmap is identical to the configmap on the hub except the controller
	// type.

	return nil
}

func (c *Command) validatedRamenControllerType(
	config *ramenapi.RamenConfig,
	expectedType ramenapi.ControllerType,
) report.ValidatedString {
	validated := report.ValidatedString{Value: string(config.RamenControllerType)}

	if config.RamenControllerType != expectedType {
		validated.State = report.Problem
		validated.Description = fmt.Sprintf("Expecting controller type %q", expectedType)
	} else {
		validated.State = report.OK
	}

	addValidation(c.report.Summary, &validated)
	return validated
}

func (c *Command) validatedS3Profiles(
	s *report.ValidatedS3StoreProfilesList,
	cluster *types.Cluster,
	config *ramenapi.RamenConfig,
	configNamespace string,
) error {
	for i := range config.S3StoreProfiles {
		profile := &config.S3StoreProfiles[i]

		validatedSecret, err := c.validatedSecretRef(profile.S3SecretRef, cluster, configNamespace)
		if err != nil {
			return fmt.Errorf("failed to validate s3 profile %q secret: %w",
				profile.S3ProfileName, err)
		}

		ps := report.S3StoreProfilesSummary{
			S3ProfileName: profile.S3ProfileName,
			S3SecretRef:   validatedSecret,
		}
		s.Value = append(s.Value, ps)
	}

	if len(s.Value) == 0 {
		s.State = report.Problem
		s.Description = "No s3 profiles found"
	} else {
		s.State = report.OK
	}
	addValidation(c.report.Summary, s)

	return nil
}

func (c *Command) validatedSecretRef(
	secretRef corev1.SecretReference,
	cluster *types.Cluster,
	configNamespace string,
) (report.ValidatedS3SecretRef, error) {
	log := c.Logger()
	reader := c.outputReader(cluster.Name)
	validated := report.ValidatedS3SecretRef{Value: secretRef}

	namespace := secretRef.Namespace
	if namespace == "" {
		namespace = configNamespace
	}

	_, err := core.ReadSecret(reader, secretRef.Name, namespace)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			err := fmt.Errorf("failed to read secret \"%s/%s\" from cluster %q: %w",
				namespace, secretRef.Name, cluster.Name, err)
			return validated, err
		}
		log.Debugf("Secret \"%s/%s\" does not exist in cluster %q",
			namespace, secretRef.Name, cluster.Name)
		validated.State = report.Problem
		validated.Description = "Secret does not exist"
	} else {
		log.Debugf("Read secret \"%s/%s\" from cluster %q",
			namespace, secretRef.Name, cluster.Name)
		// TODO:
		// - Validate secret identical to hub secret?
		// - Validate secret required fields?
		validated.State = report.OK
	}

	addValidation(c.report.Summary, &validated)

	return validated, nil
}

func (c *Command) validateDeployment(
	s *report.DeploymentSummary,
	cluster *types.Cluster,
	name, namespace string,
	expectedReplicas int32,
) error {
	log := c.Logger()
	reader := c.outputReader(cluster.Name)

	s.Name = name
	s.Namespace = namespace

	deployment, err := readDeployment(reader, name, namespace)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to read deployment \"%s/%s\" from cluster %q: %w",
				namespace, name, cluster.Name, err)
		}

		log.Debugf("Deployment \"%s/%s\" does not exist in cluster %q",
			namespace, name, cluster.Name)
		s.Deleted = c.validatedDeleted(nil)
		return nil
	}

	log.Debugf("Read deployment \"%s/%s\" from cluster %q", namespace, name, cluster.Name)
	s.Deleted = c.validatedDeleted(deployment)
	s.Replicas = c.validatedDeploymentReplicas(deployment, expectedReplicas)
	s.Conditions = c.validatedDeploymentConditions(deployment)

	return nil
}

func (c *Command) validatedDeploymentReplicas(
	deployment *appsv1.Deployment,
	expectedReplicas int32,
) report.ValidatedInteger {
	validated := report.ValidatedInteger{Value: defaultReplicas}

	if deployment.Spec.Replicas != nil {
		validated.Value = int64(*deployment.Spec.Replicas)
	}

	if validated.Value != int64(expectedReplicas) {
		validated.State = report.Problem
		validated.Description = fmt.Sprintf("Expecting %d replicas", expectedReplicas)
	} else {
		validated.State = report.OK
	}

	addValidation(c.report.Summary, &validated)
	return validated
}

func (c *Command) validatedDeploymentConditions(
	deployment *appsv1.Deployment,
) []report.ValidatedCondition {
	log := c.Logger()
	var conditions []report.ValidatedCondition

	for i := range deployment.Status.Conditions {
		condition := &deployment.Status.Conditions[i]

		var expectedStatus corev1.ConditionStatus
		switch condition.Type {
		case appsv1.DeploymentAvailable, appsv1.DeploymentProgressing:
			expectedStatus = corev1.ConditionTrue
		case appsv1.DeploymentReplicaFailure:
			expectedStatus = corev1.ConditionFalse
		default:
			// Possible if new deployemnt condition is added. We don't have a way to fail during
			// compile time if a new type is introduced.
			log.Warnf("Expecting True status for unexpected deployment condition: %+v", condition)
			expectedStatus = corev1.ConditionTrue
		}

		validated := validatedDeploymentCondition(condition, expectedStatus)
		addValidation(c.report.Summary, &validated)
		conditions = append(conditions, validated)
	}

	return conditions
}

func (c *Command) checkS3(profiles []*s3.Profile) bool {
	start := time.Now()

	c.Logger().Infof("Checking S3 profiles %q", logging.ProfileNames(profiles))

	for r := range c.backend.CheckS3(c, profiles) {
		// Collect results to validate and report S3 status in validateClustersS3Status.
		c.s3Results = append(c.s3Results, r)

		step := &report.Step{
			Name:     fmt.Sprintf("check S3 profile %q", r.ProfileName),
			Duration: r.Duration,
		}
		if r.Err != nil {
			if errors.Is(r.Err, context.Canceled) {
				msg := fmt.Sprintf("Canceled check S3 profile %q", r.ProfileName)
				console.Error(msg)
				c.Logger().Errorf("%s: %s", msg, r.Err)
				step.Status = report.Canceled
			} else {
				msg := fmt.Sprintf("Failed to check S3 profile %q", r.ProfileName)
				console.Error(msg)
				c.Logger().Errorf("%s: %s", msg, r.Err)
				step.Status = report.Failed
			}
		} else {
			step.Status = report.Passed
			console.Pass("Checked S3 profile %q", r.ProfileName)
		}
		c.current.AddStep(step)
	}

	c.Logger().Infof("Checked S3 profiles in %.2f seconds", time.Since(start).Seconds())

	// We want to stop only if the user cancelled. Errors will be
	// reported during validation.
	return c.current.Status != report.Canceled
}

func (c *Command) validateClustersS3Status(s *report.ClustersS3Status) {
	c.validatedClustersS3ProfileStatus(&s.Profiles)
}

func (c *Command) validatedClustersS3ProfileStatus(s *report.ValidatedClustersS3ProfileStatusList) {
	if len(c.s3Results) > 0 {
		// Checked S3 for one or more profiles, validate the results.
		s.State = report.OK
		for _, result := range c.s3Results {
			validated := c.validatedClustersS3Profile(result)
			s.Value = append(s.Value, validated)
		}
	} else {
		// Failed to get S3 profiles from the gathered hub data.
		s.State = report.Problem
		s.Description = "No s3 profiles found"
	}

	addValidation(c.report.Summary, s)
}

func (c *Command) validatedClustersS3Profile(result s3.Result) report.ClustersS3ProfileStatus {
	profileStatus := report.ClustersS3ProfileStatus{
		Name: result.ProfileName,
	}

	if result.Err != nil {
		profileStatus.Accessible = report.ValidatedBool{
			Validated: report.Validated{
				State:       report.Problem,
				Description: result.Err.Error(),
			},
			Value: false,
		}
	} else {
		profileStatus.Accessible = report.ValidatedBool{
			Validated: report.Validated{
				State: report.OK,
			},
			Value: true,
		}
	}

	addValidation(c.report.Summary, &profileStatus.Accessible)
	return profileStatus
}
