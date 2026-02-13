// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package ramen

import (
	"context"
	"fmt"
	"maps"
	"slices"

	ramenapi "github.com/ramendr/ramen/api/v1alpha1"
	e2etypes "github.com/ramendr/ramen/e2e/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/core"
	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/s3"
)

const (
	// PV data is protected. This means that, the PV data from the storage
	// is in complete sync with its remote peer.
	// https://github.com/RamenDR/ramen/blob/eebc5c0cb46af2eea145e7d40feef09681f6b110/internal/controller/status.go#L25
	VRGConditionTypeDataProtected = "DataProtected"

	// VolSync related conditions. These conditions are only applicable
	// at individual PVCs and not generic VRG conditions.
	// https://github.com/RamenDR/ramen/blob/eebc5c0cb46af2eea145e7d40feef09681f6b110/internal/controller/status.go#L50
	VRGConditionTypeVolSyncPVsRestored = "PVsRestored"

	// ConditionUnused is used on the secondary cluster VRG.
	// https://github.com/RamenDR/ramen/blob/eebc5c0cb46af2eea145e7d40feef09681f6b110/internal/controller/status.go#L55
	VRGConditionReasonUnused = "Unused"

	// HubOperatorName is the name of the deploymentg on the hub.
	// TODO: discover the value from the cluster.
	HubOperatorName = "ramen-hub-operator"

	// DRClusterOperatorName is the name of the deploymentg on the managed clusters.
	// TODO: discover the value from the cluster.
	DRClusterOperatorName = "ramen-dr-cluster-operator"

	// HubOperatorConfigMapName is the name of the ramen configmap on the hub.
	// https://github.com/RamenDR/ramen/blob/bd59a54fa7cdff2e48c1725460cfd76dda9c27e9/internal/controller/ramenconfig.go#L31
	HubOperatorConfigMapName = HubOperatorName + configMapNameSuffix

	// DrClusterOperatorConfigMapName is the name of the ramen configmap on the managed clusters.
	// https://github.com/RamenDR/ramen/blob/bd59a54fa7cdff2e48c1725460cfd76dda9c27e9/internal/controller/ramenconfig.go#L32
	DrClusterOperatorConfigMapName = DRClusterOperatorName + configMapNameSuffix

	// ConfigMapRamenConfigKeyName is the name configuration YAML in the ramen configmap.
	// https://github.com/RamenDR/ramen/blob/ac64bd0bb67bcb194b938d52dc86bd165807987e/internal/controller/ramenconfig.go#L35
	ConfigMapRamenConfigKeyName = "ramen_manager_config.yaml"

	// OperatorReplicas is the number of pods in the ramen operator deployment.
	// TODO: discover the value from the cluster.
	OperatorReplicas = 1
)

const (
	// Annotation for application namespace on the managed cluster
	// from ramen/internal/controllers/drplacementcontrol.go
	drpcAppNamespaceAnnotation = "drplacementcontrol.ramendr.openshift.io/app-namespace"

	// TODO: find a way to get this from ramen api. Available in the CRD under spec/names/plural.
	// Should we gather the CRDs from the cluster?
	drpcPlural      = "drplacementcontrols"
	vrgPlural       = "volumereplicationgroups"
	drPolicyPlural  = "drpolicies"
	drClusterPlural = "drclusters"

	configMapNameSuffix = "-config"
)

// Actions are the valid DRPC and VRG actions.
// NOTE: ramen uses different type for vrg actions with the same values.
var Actions = []string{"", string(ramenapi.ActionFailover), string(ramenapi.ActionRelocate)}

type Context interface {
	Env() *e2etypes.Env
	Context() context.Context
}

// OperatorDeploymentName returns the deployment name for the given controller type.
func OperatorDeploymentName(controllerType ramenapi.ControllerType) string {
	switch controllerType {
	case ramenapi.DRHubType:
		return HubOperatorName
	case ramenapi.DRClusterType:
		return DRClusterOperatorName
	default:
		panic(fmt.Sprintf("Invalid controller type %q", controllerType))
	}
}

// OperatorConfigMapName returns the ramen configmap name for the given controller type.
func OperatorConfigMapName(controllerType ramenapi.ControllerType) string {
	switch controllerType {
	case ramenapi.DRHubType:
		return HubOperatorConfigMapName
	case ramenapi.DRClusterType:
		return DrClusterOperatorConfigMapName
	default:
		panic(fmt.Sprintf("Invalid controller type %q", controllerType))
	}
}

func ApplicationNamespaces(drpc *ramenapi.DRPlacementControl) []string {
	seen := map[string]struct{}{
		drpc.Namespace: {},
	}
	if drpc.Spec.ProtectedNamespaces != nil {
		for _, ns := range *drpc.Spec.ProtectedNamespaces {
			if ns != "" {
				seen[ns] = struct{}{}
			}
		}
	}
	if appNamespace := drpc.Annotations[drpcAppNamespaceAnnotation]; appNamespace != "" {
		seen[appNamespace] = struct{}{}
	}
	return slices.Collect(maps.Keys(seen))
}

func VRGNamespace(drpc *ramenapi.DRPlacementControl) string {
	return drpc.Annotations[drpcAppNamespaceAnnotation]
}

// PrimaryCluster returns the desired cluster for the application. During failover or relocate it
// may take few minutes until the application is placed on this cluster.
func PrimaryCluster(ctx Context, drpc *ramenapi.DRPlacementControl) (*e2etypes.Cluster, error) {
	return ctx.Env().GetCluster(primaryClusterName(drpc))
}

// SecondaryCluster returns the desired secondary cluster for the application. During failover or
// relocate it may take few minutes until the application is moved out of the secondary cluster.
func SecondaryCluster(ctx Context, drpc *ramenapi.DRPlacementControl) (*e2etypes.Cluster, error) {
	clusterName := primaryClusterName(drpc)

	// TODO: Use the dr policy to match ramen behavior.
	env := ctx.Env()
	switch clusterName {
	case env.C1.Name:
		return env.GetCluster(env.C2.Name)
	case env.C2.Name:
		return env.GetCluster(env.C1.Name)
	default:
		return nil, fmt.Errorf("primary cluster %q unknown", clusterName)
	}
}

func StablePhase(action ramenapi.DRAction) (ramenapi.DRState, error) {
	switch action {
	case "":
		return ramenapi.Deployed, nil
	case ramenapi.ActionFailover:
		return ramenapi.FailedOver, nil
	case ramenapi.ActionRelocate:
		return ramenapi.Relocated, nil
	default:
		return "", fmt.Errorf("unknown action %q", action)
	}
}

func GetDRPC(ctx Context, drpcName, drpcNamespace string) (*ramenapi.DRPlacementControl, error) {
	drpc := &ramenapi.DRPlacementControl{}
	key := types.NamespacedName{Namespace: drpcNamespace, Name: drpcName}
	err := ctx.Env().Hub.Client.Get(ctx.Context(), key, drpc)
	if err != nil {
		return nil, err
	}
	return drpc, nil
}

// ReadDRPC reads a ramen DRPlacementControl from the output directory.
func ReadDRPC(
	reader gathering.OutputReader,
	name, namespace string,
) (*ramenapi.DRPlacementControl, error) {
	resource := ramenapi.GroupVersion.Group + "/" + drpcPlural
	data, err := reader.ReadResource(namespace, resource, name)
	if err != nil {
		return nil, err
	}
	drpc := &ramenapi.DRPlacementControl{}
	if err := yaml.Unmarshal(data, drpc); err != nil {
		return nil, err
	}
	return drpc, nil
}

// ReadVRG reads a ramen VolumeReplicationGroup from the output directory.
func ReadVRG(
	reader gathering.OutputReader,
	name, namespace string,
) (*ramenapi.VolumeReplicationGroup, error) {
	resource := ramenapi.GroupVersion.Group + "/" + vrgPlural
	data, err := reader.ReadResource(namespace, resource, name)
	if err != nil {
		return nil, err
	}
	vrg := &ramenapi.VolumeReplicationGroup{}
	if err := yaml.Unmarshal(data, vrg); err != nil {
		return nil, err
	}
	return vrg, nil
}

// ReadDRPolicy reads a ramen DRPolicy from the output directory.
func ReadDRPolicy(reader gathering.OutputReader, name string) (*ramenapi.DRPolicy, error) {
	resource := ramenapi.GroupVersion.Group + "/" + drPolicyPlural
	data, err := reader.ReadResource("", resource, name)
	if err != nil {
		return nil, err
	}
	drPolicy := &ramenapi.DRPolicy{}
	if err := yaml.Unmarshal(data, drPolicy); err != nil {
		return nil, err
	}
	return drPolicy, nil
}

// ReadDRCluster reads a ramen DRCluster from the output directory.
func ReadDRCluster(reader gathering.OutputReader, name string) (*ramenapi.DRCluster, error) {
	resource := ramenapi.GroupVersion.Group + "/" + drClusterPlural
	data, err := reader.ReadResource("", resource, name)
	if err != nil {
		return nil, err
	}
	drCluster := &ramenapi.DRCluster{}
	if err := yaml.Unmarshal(data, drCluster); err != nil {
		return nil, err
	}
	return drCluster, nil
}

// ListDRPolicies lists ramen DRPolicies from the output directory.
func ListDRPolicies(reader gathering.OutputReader) ([]string, error) {
	resource := ramenapi.GroupVersion.Group + "/" + drPolicyPlural
	return reader.ListResources("", resource)
}

// ListDRClusters lists ramen DRClusters from the output directory.
func ListDRClusters(reader gathering.OutputReader) ([]string, error) {
	resource := ramenapi.GroupVersion.Group + "/" + drClusterPlural
	return reader.ListResources("", resource)
}

// ClusterProfiles extracts S3 store profiles from the ramen configmap.
func ClusterProfiles(
	reader gathering.OutputReader,
	configMapName, configMapNamespace string,
) ([]*ramenapi.S3StoreProfile, error) {
	configData, err := getRamenConfigMapData(reader, configMapName, configMapNamespace)
	if err != nil {
		return nil, err
	}
	if len(configData.S3StoreProfiles) == 0 {
		return nil, fmt.Errorf("no S3 profiles found in ramen config")
	}
	profiles := make([]*ramenapi.S3StoreProfile, 0, len(configData.S3StoreProfiles))
	for i := range configData.S3StoreProfiles {
		profile := &configData.S3StoreProfiles[i]
		// Default empty namespace to configmap namespace.
		if profile.S3SecretRef.Namespace == "" {
			profile.S3SecretRef.Namespace = configMapNamespace
		}
		profiles = append(profiles, profile)
	}
	return profiles, nil
}

// S3ProfileFromStore creates an s3.Profile from a ramen S3StoreProfile and secret.
// If secret is nil, the profile will have empty credentials.
func S3ProfileFromStore(storeProfile *ramenapi.S3StoreProfile, secret *corev1.Secret) *s3.Profile {
	var accessKey, secretKey string
	if secret != nil {
		accessKey = string(secret.Data["AWS_ACCESS_KEY_ID"])
		secretKey = string(secret.Data["AWS_SECRET_ACCESS_KEY"])
	}
	return &s3.Profile{
		Name:          storeProfile.S3ProfileName,
		Bucket:        storeProfile.S3Bucket,
		Region:        storeProfile.S3Region,
		Endpoint:      storeProfile.S3CompatibleEndpoint,
		CACertificate: storeProfile.CACertificates,
		AccessKey:     accessKey,
		SecretKey:     secretKey,
	}
}

// ApplicationS3Prefix returns the s3 object prefix for an application's s3 data.
func ApplicationS3Prefix(
	reader gathering.OutputReader,
	drpcName, drpcNamespace string,
) (string, error) {
	drpc, err := ReadDRPC(reader, drpcName, drpcNamespace)
	if err != nil {
		return "", fmt.Errorf("failed to read drpc \"%s/%s\": %w",
			drpcNamespace, drpcName, err)
	}
	vrgNamespace := VRGNamespace(drpc)
	if vrgNamespace == "" {
		return "", fmt.Errorf("drpc \"%s/%s\" annotation %q not found",
			drpc.Namespace, drpc.Name, drpcAppNamespaceAnnotation)
	}
	return fmt.Sprintf("%s/%s/", vrgNamespace, drpc.Name), nil
}

// getRamenConfigMapData reads and parse the ramen operator configmap data.
func getRamenConfigMapData(
	reader gathering.OutputReader,
	name, namespace string,
) (*ramenapi.RamenConfig, error) {
	configMap, err := core.ReadConfigMap(reader, name, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to read ramen configmap \"%s/%s\": %w",
			namespace, name, err)
	}
	configData := &ramenapi.RamenConfig{}
	data := []byte(configMap.Data[ConfigMapRamenConfigKeyName])
	if err := yaml.Unmarshal(data, configData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal ramen configmap data: %w\n%s", err, data)
	}
	return configData, nil
}

func primaryClusterName(drpc *ramenapi.DRPlacementControl) string {
	if drpc.Spec.Action == ramenapi.ActionFailover {
		return drpc.Spec.FailoverCluster
	}
	return drpc.Spec.PreferredCluster
}
