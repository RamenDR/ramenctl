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

const (
	// minS3Profiles is the minimum S3 profiles in configmap required for DR.
	minS3Profiles = 2

	profileNotFoundInHub = "Profile not found in hub"
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

	if !c.checkClustersS3() {
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

// checkClustersS3 inspects S3 profiles and checks access. It returns false only if
// the user cancelled, otherwise true even if there were errors during inspection, as those
// will be reported in the validation results.
func (c *Command) checkClustersS3() bool {
	profiles, err := c.inspectClustersS3Profiles()
	if err != nil {
		return !errors.Is(err, context.Canceled)
	}
	return c.checkS3(profiles)
}

func (c *Command) inspectClustersS3Profiles() ([]*s3.Profile, error) {
	start := time.Now()
	step := &report.Step{Name: "inspect S3 profiles"}

	c.Logger().Infof("Step %q started", step.Name)

	profiles, err := c.clustersS3Info()
	if err != nil {
		step.Duration = time.Since(start).Seconds()
		if errors.Is(err, context.Canceled) {
			step.Status = report.Canceled
			console.Error("Canceled %s", step.Name)
		} else {
			step.Status = report.Failed
			console.Error("Failed to %s", step.Name)
		}
		c.Logger().Errorf("Step %q %s: %s", step.Name, step.Status, err)
		c.current.AddStep(step)
		return nil, err
	}

	step.Duration = time.Since(start).Seconds()
	step.Status = report.Passed
	c.current.AddStep(step)

	console.Pass("Inspected S3 profiles")
	c.Logger().Infof("Step %q passed", step.Name)

	return profiles, nil
}

// clustersS3Info reads S3 profiles and fetches secrets from the hub cluster.
func (c *Command) clustersS3Info() ([]*s3.Profile, error) {
	// Read S3 profiles from the ramen hub configmap, the source of truth
	// synced to managed clusters.
	hub := c.Env().Hub
	reader := c.outputReader(hub.Name)
	configMapName := ramen.HubOperatorConfigMapName
	configMapNamespace := c.config.Namespaces.RamenHubNamespace

	storeProfiles, err := ramen.ClusterProfiles(reader, configMapName, configMapNamespace)
	if err != nil {
		return nil, err
	}

	// Get S3 secrets from live hub cluster since gathered data may contain
	// sanitized secrets. On cancellation, return immediately. On other failures,
	// empty credentials will cause S3 operations to fail during checkS3.
	var profiles []*s3.Profile
	for _, sp := range storeProfiles {
		secret, err := c.backend.GetSecret(c, hub, sp.S3SecretRef.Name, sp.S3SecretRef.Namespace)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return nil, err
			}
			c.Logger().Warnf("Failed to get S3 secret \"%s/%s\" from cluster %q: %s",
				sp.S3SecretRef.Namespace, sp.S3SecretRef.Name, hub.Name, err)
		}
		profiles = append(profiles, ramen.S3ProfileFromStore(sp, secret))
	}

	return profiles, nil
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

	if controllerType == ramenapi.DRHubType {
		if err := c.validatedHubS3Profiles(
			&s.S3StoreProfiles,
			cluster,
			config,
			namespace,
		); err != nil {
			return fmt.Errorf("failed to validate hub s3 profiles: %w", err)
		}
	} else {
		if err := c.validatedManagedClusterS3Profiles(
			&s.S3StoreProfiles,
			cluster,
			config,
			namespace,
		); err != nil {
			return fmt.Errorf("failed to validate managed cluster s3 profiles: %w", err)
		}
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

// validatedHubS3Profiles validates that S3 profile fields in the hub are not empty.
func (c *Command) validatedHubS3Profiles(
	s *report.ValidatedS3StoreProfilesList,
	cluster *types.Cluster,
	config *ramenapi.RamenConfig,
	configNamespace string,
) error {
	for i := range config.S3StoreProfiles {
		profile := &config.S3StoreProfiles[i]

		validatedSecret, err := c.validatedHubSecretRef(
			profile.S3SecretRef,
			cluster,
			configNamespace,
		)
		if err != nil {
			return fmt.Errorf("failed to validate s3 profile %q secret in cluster %q: %w",
				profile.S3ProfileName, cluster.Name, err)
		}

		ps := report.S3StoreProfilesSummary{
			S3ProfileName:        profile.S3ProfileName,
			S3Bucket:             c.validatedRequiredString(profile.S3Bucket),
			S3CompatibleEndpoint: c.validatedRequiredString(profile.S3CompatibleEndpoint),
			S3Region:             c.validatedRequiredString(profile.S3Region),
			CACertificate:        c.validatedCertificateFingerprint(profile.CACertificates),
			S3SecretRef:          validatedSecret,
		}
		s.Value = append(s.Value, ps)
	}

	if len(s.Value) < minS3Profiles {
		s.State = report.Problem
		s.Description = fmt.Sprintf("Found %d S3 profile(s), expected at least %d",
			len(s.Value), minS3Profiles)
	} else {
		s.State = report.OK
	}
	addValidation(c.report.Summary, s)

	return nil
}

// validatedManagedClusterS3Profiles validates that managed cluster S3 profile fields
// are not empty and match the hub profile.
func (c *Command) validatedManagedClusterS3Profiles(
	s *report.ValidatedS3StoreProfilesList,
	cluster *types.Cluster,
	config *ramenapi.RamenConfig,
	configNamespace string,
) error {
	for i := range config.S3StoreProfiles {
		profile := &config.S3StoreProfiles[i]

		hubS3Profile, found := c.lookupHubS3StoreProfileSummary(profile.S3ProfileName)

		validatedSecret, err := c.validatedManagedClusterSecretRef(
			profile.S3SecretRef, cluster, configNamespace, hubS3Profile.S3SecretRef, found)
		if err != nil {
			return fmt.Errorf("failed to validate s3 profile %q secret in cluster %q: %w",
				profile.S3ProfileName, cluster.Name, err)
		}

		ps := report.S3StoreProfilesSummary{
			S3ProfileName: profile.S3ProfileName,
			S3Bucket: c.validatedManagedClusterRequiredString(
				profile.S3Bucket,
				hubS3Profile.S3Bucket,
				found,
			),
			S3CompatibleEndpoint: c.validatedManagedClusterRequiredString(
				profile.S3CompatibleEndpoint,
				hubS3Profile.S3CompatibleEndpoint,
				found,
			),
			S3Region: c.validatedManagedClusterRequiredString(
				profile.S3Region,
				hubS3Profile.S3Region,
				found,
			),
			CACertificate: c.validatedManagedClusterCertificateFingerprint(
				profile.CACertificates,
				hubS3Profile.CACertificate,
				found,
			),
			S3SecretRef: validatedSecret,
		}
		s.Value = append(s.Value, ps)
	}

	hubS3ProfileCount := len(c.report.ClustersStatus.Hub.Ramen.ConfigMap.S3StoreProfiles.Value)
	switch {
	case len(s.Value) < minS3Profiles:
		s.State = report.Problem
		s.Description = fmt.Sprintf("Found %d S3 profile(s), expected at least %d",
			len(s.Value), minS3Profiles)
	case len(s.Value) != hubS3ProfileCount:
		s.State = report.Problem
		s.Description = fmt.Sprintf("Found %d S3 profile(s), hub has %d",
			len(s.Value), hubS3ProfileCount)
	default:
		s.State = report.OK
	}
	addValidation(c.report.Summary, s)

	return nil
}

func (c *Command) lookupHubS3StoreProfileSummary(
	name string,
) (report.S3StoreProfilesSummary, bool) {
	profiles := c.report.ClustersStatus.Hub.Ramen.ConfigMap.S3StoreProfiles.Value
	for i := range profiles {
		if profiles[i].S3ProfileName == name {
			return profiles[i], true
		}
	}
	return report.S3StoreProfilesSummary{}, false
}

func (c *Command) validatedRequiredString(value string) report.ValidatedString {
	validated := report.ValidatedString{Value: value}

	if value == "" {
		validated.State = report.Problem
		validated.Description = "Value is not set"
	} else {
		validated.State = report.OK
	}

	addValidation(c.report.Summary, &validated)
	return validated
}

func (c *Command) validatedManagedClusterRequiredString(
	value string,
	hubValue report.ValidatedString,
	found bool,
) report.ValidatedString {
	validated := report.ValidatedString{Value: value}

	switch {
	case !found:
		validated.State = report.Problem
		validated.Description = profileNotFoundInHub
	case value == "":
		validated.State = report.Problem
		validated.Description = "Value is not set"
	case value != hubValue.Value:
		validated.State = report.Problem
		validated.Description = fmt.Sprintf("Does not match hub: %q", hubValue.Value)
	default:
		validated.State = report.OK
	}

	addValidation(c.report.Summary, &validated)
	return validated
}

func (c *Command) validatedCertificateFingerprint(certPem []byte) report.ValidatedFingerprint {
	validated := report.ValidatedFingerprint{}

	switch {
	case len(certPem) == 0:
		// Empty certificate is validated OK, since it is an optional field.
		validated.State = report.OK
	default:
		fingerprint, err := report.CertificateFingerprint(certPem)
		if err != nil {
			validated.State = report.Problem
			validated.Description = fmt.Sprintf("Invalid certificate: %s", err)
		} else {
			validated.Value = fingerprint
			validated.State = report.OK
		}
	}

	addValidation(c.report.Summary, &validated)
	return validated
}

func (c *Command) validatedManagedClusterCertificateFingerprint(
	certPem []byte,
	hubValue report.ValidatedFingerprint,
	found bool,
) report.ValidatedFingerprint {
	validated := report.ValidatedFingerprint{}

	switch {
	case !found:
		validated.State = report.Problem
		validated.Description = profileNotFoundInHub
	case hubValue.State == report.Problem:
		// Hub has invalid certificate, can't validate against it.
		validated.State = report.Problem
		validated.Description = "Hub certificate is invalid"
	case len(certPem) == 0:
		// Managed cluster has no certificate.
		if hubValue.Value != "" {
			validated.State = report.Problem
			validated.Description = "Missing certificate, but hub has a certificate"
		} else {
			// Validated OK if both Managed cluster and Hub have no certificate.
			validated.State = report.OK
		}
	default:
		// Managed cluster has certificate, compute fingerprint.
		fingerprint, err := report.CertificateFingerprint(certPem)
		if err != nil {
			validated.State = report.Problem
			validated.Description = fmt.Sprintf("Invalid certificate: %s", err)
		} else {
			validated.Value = fingerprint
			switch {
			case hubValue.Value == "":
				validated.State = report.Problem
				validated.Description = "Has certificate, but hub does not have a certificate"
			case fingerprint != hubValue.Value:
				validated.State = report.Problem
				validated.Description = fmt.Sprintf("Does not match hub: %q", hubValue.Value)
			default:
				validated.State = report.OK
			}
		}
	}

	addValidation(c.report.Summary, &validated)
	return validated
}

func (c *Command) validatedHubSecretRef(
	secretRef corev1.SecretReference,
	cluster *types.Cluster,
	configNamespace string,
) (report.S3SecretSummary, error) {
	log := c.Logger()

	validated := report.S3SecretSummary{
		Name:      c.validatedRequiredString(secretRef.Name),
		Namespace: c.validatedSecretNamespaceString(secretRef.Namespace, configNamespace),
	}

	secret, err := c.readSecret(cluster, secretRef.Name, secretRef.Namespace, configNamespace)
	if err != nil {
		return validated, err
	}

	if secret == nil {
		log.Debugf("Secret \"%s/%s\" does not exist in cluster %q",
			secretRef.Namespace, secretRef.Name, cluster.Name)
		validated.Deleted = c.validatedDeleted(nil)
	} else {
		log.Debugf("Read secret \"%s/%s\" from cluster %q",
			secretRef.Namespace, secretRef.Name, cluster.Name)
		validated.Deleted = c.validatedDeleted(secret)
		validated.AWSAccessKeyID = c.validatedSecretKeyFingerprint(secret, "AWS_ACCESS_KEY_ID")
		validated.AWSSecretAccessKey = c.validatedSecretKeyFingerprint(
			secret,
			"AWS_SECRET_ACCESS_KEY",
		)
	}

	return validated, nil
}

func (c *Command) validatedManagedClusterSecretRef(
	secretRef corev1.SecretReference,
	cluster *types.Cluster,
	configNamespace string,
	hubSecret report.S3SecretSummary,
	found bool,
) (report.S3SecretSummary, error) {
	log := c.Logger()

	validated := report.S3SecretSummary{
		Name:      c.validatedManagedClusterRequiredString(secretRef.Name, hubSecret.Name, found),
		Namespace: c.validatedSecretNamespaceString(secretRef.Namespace, configNamespace),
	}

	secret, err := c.readSecret(cluster, secretRef.Name, secretRef.Namespace, configNamespace)
	if err != nil {
		return validated, err
	}

	if secret == nil {
		log.Debugf("Secret \"%s/%s\" does not exist in cluster %q",
			secretRef.Namespace, secretRef.Name, cluster.Name)
		validated.Deleted = c.validatedDeleted(nil)
	} else {
		log.Debugf("Read secret \"%s/%s\" from cluster %q",
			secretRef.Namespace, secretRef.Name, cluster.Name)
		validated.Deleted = c.validatedDeleted(secret)
		validated.AWSAccessKeyID = c.validatedManagedClusterSecretKeyFingerprint(
			secret, "AWS_ACCESS_KEY_ID", hubSecret.AWSAccessKeyID, found)
		validated.AWSSecretAccessKey = c.validatedManagedClusterSecretKeyFingerprint(
			secret, "AWS_SECRET_ACCESS_KEY", hubSecret.AWSSecretAccessKey, found)
	}

	return validated, nil
}

func (c *Command) readSecret(
	cluster *types.Cluster,
	name, namespace, configNamespace string,
) (*corev1.Secret, error) {
	if namespace == "" {
		namespace = configNamespace
	}
	reader := c.outputReader(cluster.Name)
	secret, err := core.ReadSecret(reader, name, namespace)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("failed to read secret \"%s/%s\" from cluster %q: %w",
				namespace, name, cluster.Name, err)
		}
		// Secret doesn't exist
		return nil, nil
	}
	return secret, nil
}

func (c *Command) validatedSecretNamespaceString(
	namespace string,
	configNamespace string,
) report.ValidatedString {
	validated := report.ValidatedString{Value: namespace}

	// Namespace can be empty (defaults to configmap namespace)
	// But if specified, must match configmap namespace.
	if namespace != "" && namespace != configNamespace {
		validated.State = report.Problem
		validated.Description = fmt.Sprintf("Must be in configmap namespace %q", configNamespace)
	} else {
		validated.State = report.OK
	}

	addValidation(c.report.Summary, &validated)
	return validated
}

func (c *Command) validatedSecretKeyFingerprint(
	secret *corev1.Secret,
	key string,
) report.ValidatedFingerprint {
	validated := report.ValidatedFingerprint{}

	data, exists := secret.Data[key]
	switch {
	case !exists:
		validated.State = report.Problem
		validated.Description = "Key is missing"
	case len(data) == 0:
		validated.State = report.Problem
		validated.Description = "Key is empty"
	default:
		fingerprint, err := report.Fingerprint(data)
		if err != nil {
			panic(fmt.Sprintf("unexpected Fingerprint() error: %v", err))
		}
		validated.Value = fingerprint
		validated.State = report.OK
	}

	addValidation(c.report.Summary, &validated)
	return validated
}

func (c *Command) validatedManagedClusterSecretKeyFingerprint(
	secret *corev1.Secret,
	key string,
	hubValue report.ValidatedFingerprint,
	found bool,
) report.ValidatedFingerprint {
	validated := report.ValidatedFingerprint{}

	switch {
	case !found:
		validated.State = report.Problem
		validated.Description = profileNotFoundInHub
	case hubValue.Value == "":
		validated.State = report.Problem
		validated.Description = "Hub key is missing"
	default:
		data, exists := secret.Data[key]
		switch {
		case !exists:
			validated.State = report.Problem
			validated.Description = "Key is missing"
		case len(data) == 0:
			validated.State = report.Problem
			validated.Description = "Key is empty"
		default:
			fingerprint, err := report.Fingerprint(data)
			if err != nil {
				panic(fmt.Sprintf("unexpected Fingerprint() error: %v", err))
			}
			validated.Value = fingerprint
			if fingerprint != hubValue.Value {
				validated.State = report.Problem
				validated.Description = fmt.Sprintf("Does not match hub: %q", hubValue.Value)
			} else {
				validated.State = report.OK
			}
		}
	}

	addValidation(c.report.Summary, &validated)
	return validated
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

// checkS3 checks S3 access for the given profiles by verifying bucket connectivity.
// Returns false only if the user cancelled, otherwise true even if there were errors,
// as those will be reported during validation.
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
