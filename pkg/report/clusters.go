// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"slices"
)

// DRClusterSummary is the summary of a DRCluster.
type DRClusterSummary struct {
	// Name is the name of the DRCluster.
	Name string `json:"name"`
	// Phase is the phase of the DRCluster.
	Phase string `json:"phase,omitempty"`
	// Conditions are the validated conditions of the DRCluster.
	Conditions ValidatedConditionList `json:"conditions,omitempty"`
}

// DRPolicySummary is the summary of a DRPolicy.
type DRPolicySummary struct {
	// Name is the name of the DRPolicy.
	Name string `json:"name"`
	// DRClusters is the list of DRCluster names covered by the DRPolicy.
	DRClusters []string `json:"drClusters"`
	// SchedulingInterval is the replication scheduling interval configured
	// on the DRPolicy.
	SchedulingInterval string `json:"schedulingInterval"`
	// PeerClasses is the validated list of peer storage classes for the DRPolicy.
	PeerClasses ValidatedPeerClassesList `json:"peerClasses"`
	// Conditions are the validated conditions of the DRPolicy.
	Conditions ValidatedConditionList `json:"conditions,omitempty"`
}

// PeerClassesSummary is the summary of peerClasses in a DRPolicy.
type PeerClassesSummary struct {
	// StorageClassName is the name of the peer storage class.
	StorageClassName string `json:"storageClassName"`
	// ReplicationID identifies the replication relationship for this storage class.
	ReplicationID string `json:"replicationID,omitempty"`
	// Grouping indicates whether volume grouping is enabled for this storage class.
	Grouping bool `json:"grouping,omitempty"`
}

// S3StoreProfilesSummary is the summary of S3 store profiles in the ConfigMap
type S3StoreProfilesSummary struct {
	// S3ProfileName is the name of the S3 store profile.
	S3ProfileName string `json:"profileName"`
	// S3Bucket is the validated S3 bucket name.
	S3Bucket ValidatedString `json:"bucket"`
	// S3CompatibleEndpoint is the validated S3-compatible endpoint URL.
	S3CompatibleEndpoint ValidatedString `json:"endpoint"`
	// S3Region is the validated S3 region.
	S3Region ValidatedString `json:"region"`
	// CACertificate is the validated fingerprint of the S3 CA certificate.
	CACertificate ValidatedFingerprint `json:"caCertificate"`
	// S3SecretRef is the summary of the S3 credentials secret.
	S3SecretRef S3SecretSummary `json:"secret"`
}

func (p *S3StoreProfilesSummary) AggregateState() ValidationState {
	return aggregateState(
		&p.S3Bucket,
		&p.S3CompatibleEndpoint,
		&p.S3Region,
		&p.CACertificate,
		&p.S3SecretRef,
	)
}

// S3SecretSummary is the summary of S3 Secret in the ConfigMap.
type S3SecretSummary struct {
	// Name is the validated name of the S3 secret.
	Name ValidatedString `json:"name"`
	// Namespace is the validated namespace of the S3 secret.
	Namespace ValidatedString `json:"namespace"`
	// Deleted indicates whether the S3 secret has been deleted.
	Deleted ValidatedBool `json:"deleted"`
	// AWSAccessKeyID is the validated fingerprint of the AWS access key ID.
	AWSAccessKeyID ValidatedFingerprint `json:"awsAccessKeyID"`
	// AWSSecretAccessKey is the validated fingerprint of the AWS secret access key.
	AWSSecretAccessKey ValidatedFingerprint `json:"awsSecretAccessKey"`
}

func (s *S3SecretSummary) AggregateState() ValidationState {
	return aggregateState(
		&s.Name,
		&s.Namespace,
		&s.Deleted,
		&s.AWSAccessKeyID,
		&s.AWSSecretAccessKey,
	)
}

// ConfigMapSummary is the summary of a Ramen ConfigMap.
type ConfigMapSummary struct {
	// Name is the name of the ConfigMap.
	Name string `json:"name"`
	// Namespace is the namespace of the ConfigMap.
	Namespace string `json:"namespace"`
	// Deleted indicates whether the ConfigMap has been deleted.
	Deleted ValidatedBool `json:"deleted"`
	// Parsed indicates whether the ConfigMap was parsed successfully.
	Parsed ValidatedBool `json:"parsed"`
	// S3StoreProfiles is the validated list of S3 store profiles in the ConfigMap.
	S3StoreProfiles ValidatedS3StoreProfilesList `json:"s3StoreProfiles"`
}

// DeploymentSummary is the summary of a Deployment
type DeploymentSummary struct {
	// Name is the name of the Deployment.
	Name string `json:"name"`
	// Namespace is the namespace of the Deployment.
	Namespace string `json:"namespace"`
	// Deleted indicates whether the Deployment has been deleted.
	Deleted ValidatedBool `json:"deleted"`
	// RamenControllerType is the validated type of the Ramen controller
	// running in this Deployment.
	RamenControllerType ValidatedString `json:"ramenControllerType"`
	// Replicas is the validated number of ready replicas for the Deployment.
	Replicas ValidatedInteger `json:"replicas"`
	// Conditions are the validated conditions of the Deployment.
	Conditions ValidatedConditionList `json:"conditions,omitempty"`
}

// RamenSummary is the summary of Ramen components.
type RamenSummary struct {
	// ConfigMap is the summary of the Ramen ConfigMap.
	ConfigMap ConfigMapSummary `json:"configmap"`
	// Deployment is the summary of the Ramen Deployment.
	Deployment DeploymentSummary `json:"deployment"`
}

// ClustersStatusHub is the cluster status on the hub cluster.
type ClustersStatusHub struct {
	// DRClusters is the validated list of DRCluster summaries.
	DRClusters ValidatedDRClustersList `json:"drClusters"`
	// DRPolicies is the validated list of DRPolicy summaries.
	DRPolicies ValidatedDRPoliciesList `json:"drPolicies"`
	// Ramen is the summary of Ramen components on the hub cluster.
	Ramen RamenSummary `json:"ramen"`
}

// ClustersStatusCluster is the cluster status on a managed cluster.
type ClustersStatusCluster struct {
	// Name is the name of the managed cluster.
	Name string `json:"name"`
	// Ramen is the summary of Ramen components on this managed cluster.
	Ramen RamenSummary `json:"ramen"`
}

// ClustersS3ProfileStatus is the status of an S3 profile.
type ClustersS3ProfileStatus struct {
	// Name is the name of the S3 profile.
	Name string `json:"name"`
	// Accessible indicates whether the S3 profile was reachable.
	Accessible ValidatedBool `json:"accessible"`
}

// ClustersS3Status is the status of all S3 profiles.
type ClustersS3Status struct {
	// Profiles is the validated list of S3 profile statuses.
	Profiles ValidatedClustersS3ProfileStatusList `json:"profiles"`
}

// ClustersStatus is cluster status in multi-cluster environment.
type ClustersStatus struct {
	// Hub is the cluster status on the hub cluster.
	Hub ClustersStatusHub `json:"hub"`
	// Clusters is the cluster status on each managed cluster.
	Clusters []ClustersStatusCluster `json:"clusters"`
	// S3 is the status of the cluster's S3 profiles.
	S3 ClustersS3Status `json:"s3"`
}

func (c *ClustersStatus) Equal(o *ClustersStatus) bool {
	if c == o {
		return true
	}
	if o == nil {
		return false
	}
	if !c.Hub.Equal(&o.Hub) {
		return false
	}
	if !slices.EqualFunc(
		c.Clusters,
		o.Clusters,
		func(a ClustersStatusCluster, b ClustersStatusCluster) bool {
			return a.Equal(&b)
		},
	) {
		return false
	}
	if !c.S3.Equal(&o.S3) {
		return false
	}
	return true
}

func (h *ClustersStatusHub) Equal(o *ClustersStatusHub) bool {
	if h == o {
		return true
	}
	if o == nil {
		return false
	}
	if !h.DRClusters.Equal(&o.DRClusters) {
		return false
	}
	if !h.DRPolicies.Equal(&o.DRPolicies) {
		return false
	}
	if !h.Ramen.Equal(&o.Ramen) {
		return false
	}
	return true
}

func (m *ClustersStatusCluster) Equal(o *ClustersStatusCluster) bool {
	if m == o {
		return true
	}
	if o == nil {
		return false
	}
	if m.Name != o.Name {
		return false
	}
	if !m.Ramen.Equal(&o.Ramen) {
		return false
	}
	return true
}

func (d *DRClusterSummary) Equal(o *DRClusterSummary) bool {
	if d == o {
		return true
	}
	if o == nil {
		return false
	}
	if d.Name != o.Name {
		return false
	}
	if d.Phase != o.Phase {
		return false
	}
	if !slices.Equal(d.Conditions, o.Conditions) {
		return false
	}
	return true
}

func (d *DRPolicySummary) Equal(o *DRPolicySummary) bool {
	if d == o {
		return true
	}
	if o == nil {
		return false
	}
	if d.Name != o.Name {
		return false
	}
	if !slices.Equal(d.DRClusters, o.DRClusters) {
		return false
	}
	if d.SchedulingInterval != o.SchedulingInterval {
		return false
	}
	if !d.PeerClasses.Equal(&o.PeerClasses) {
		return false
	}
	if !slices.Equal(d.Conditions, o.Conditions) {
		return false
	}
	return true
}

func (p *PeerClassesSummary) Equal(o *PeerClassesSummary) bool {
	if p == o {
		return true
	}
	if o == nil {
		return false
	}
	if p.StorageClassName != o.StorageClassName {
		return false
	}
	if p.ReplicationID != o.ReplicationID {
		return false
	}
	if p.Grouping != o.Grouping {
		return false
	}
	return true
}

func (r *RamenSummary) Equal(o *RamenSummary) bool {
	if r == o {
		return true
	}
	if o == nil {
		return false
	}
	if !r.ConfigMap.Equal(&o.ConfigMap) {
		return false
	}
	if !r.Deployment.Equal(&o.Deployment) {
		return false
	}
	return true
}

func (c *ConfigMapSummary) Equal(o *ConfigMapSummary) bool {
	if c == o {
		return true
	}
	if o == nil {
		return false
	}
	if c.Name != o.Name {
		return false
	}
	if c.Namespace != o.Namespace {
		return false
	}
	if c.Deleted != o.Deleted {
		return false
	}
	if c.Parsed != o.Parsed {
		return false
	}
	if !c.S3StoreProfiles.Equal(&o.S3StoreProfiles) {
		return false
	}
	return true
}

func (s *S3StoreProfilesSummary) Equal(o *S3StoreProfilesSummary) bool {
	if s == o {
		return true
	}
	if o == nil {
		return false
	}
	if s.S3ProfileName != o.S3ProfileName {
		return false
	}
	if s.S3Bucket != o.S3Bucket {
		return false
	}
	if s.S3CompatibleEndpoint != o.S3CompatibleEndpoint {
		return false
	}
	if s.S3Region != o.S3Region {
		return false
	}
	if s.S3SecretRef != o.S3SecretRef {
		return false
	}
	if s.CACertificate != o.CACertificate {
		return false
	}
	return true
}

func (s *ClustersS3ProfileStatus) Equal(o *ClustersS3ProfileStatus) bool {
	if s == o {
		return true
	}
	if o == nil {
		return false
	}
	if s.Name != o.Name {
		return false
	}
	if s.Accessible != o.Accessible {
		return false
	}
	return true
}

func (s *ClustersS3Status) Equal(o *ClustersS3Status) bool {
	if s == o {
		return true
	}
	if o == nil {
		return false
	}
	if !s.Profiles.Equal(&o.Profiles) {
		return false
	}
	return true
}

func (d *DeploymentSummary) Equal(o *DeploymentSummary) bool {
	if d == o {
		return true
	}
	if o == nil {
		return false
	}
	if d.Name != o.Name {
		return false
	}
	if d.Namespace != o.Namespace {
		return false
	}
	if d.Deleted != o.Deleted {
		return false
	}
	if d.RamenControllerType != o.RamenControllerType {
		return false
	}
	if d.Replicas != o.Replicas {
		return false
	}
	if !slices.Equal(d.Conditions, o.Conditions) {
		return false
	}
	return true
}
