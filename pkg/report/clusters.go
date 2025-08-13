// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"slices"
)

// DRClusterSummary is the summary of a DRCluster.
type DRClusterSummary struct {
	Name       string               `json:"name"`
	Phase      string               `json:"phase,omitempty"`
	Conditions []ValidatedCondition `json:"conditions,omitempty"`
}

// DRPolicySummary is the summary of a DRPolicy.
type DRPolicySummary struct {
	Name               string               `json:"name"`
	DRClusters         []string             `json:"drClusters"`
	SchedulingInterval string               `json:"schedulingInterval"`
	Conditions         []ValidatedCondition `json:"conditions,omitempty"`
}

// S3StoreProfilesSummary is the summary of S3 store profiles in the ConfigMap
type S3StoreProfilesSummary struct {
	S3ProfileName string               `json:"s3ProfileName"`
	S3SecretRef   ValidatedS3SecretRef `json:"s3SecretRef"`
}

// ConfigMapSummary is the summary of a Ramen ConfigMap.
type ConfigMapSummary struct {
	Name            string                       `json:"name"`
	Namespace       string                       `json:"namespace"`
	S3StoreProfiles ValidatedS3StoreProfilesList `json:"s3StoreProfiles"`
}

// DeploymentSummary is the summary of a Deployment
type DeploymentSummary struct {
	Name       string               `json:"name"`
	Namespace  string               `json:"namespace"`
	Conditions []ValidatedCondition `json:"conditions,omitempty"`
}

// RamenSummary is the summary of Ramen components.
type RamenSummary struct {
	ConfigMap  ConfigMapSummary  `json:"configmap"`
	Deployment DeploymentSummary `json:"deployment"`
}

// ClustersStatusHub is the cluster status on the hub cluster.
type ClustersStatusHub struct {
	DRClusters []DRClusterSummary `json:"drClusters"`
	DRPolicies []DRPolicySummary  `json:"drPolicies"`
	Ramen      RamenSummary       `json:"ramen"`
}

// ClustersStatusCluster is the cluster status on a managed cluster.
type ClustersStatusCluster struct {
	Name  string       `json:"name"`
	Ramen RamenSummary `json:"ramen"`
}

// ClustersStatus is cluster status in multi-cluster environment.
type ClustersStatus struct {
	Hub      ClustersStatusHub       `json:"hub"`
	Clusters []ClustersStatusCluster `json:"clusters"`
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
	return true
}

func (h *ClustersStatusHub) Equal(o *ClustersStatusHub) bool {
	if h == o {
		return true
	}
	if o == nil {
		return false
	}
	if !slices.EqualFunc(
		h.DRClusters,
		o.DRClusters,
		func(a DRClusterSummary, b DRClusterSummary) bool {
			return a.Equal(&b)
		},
	) {
		return false
	}
	if !slices.EqualFunc(
		h.DRPolicies,
		o.DRPolicies,
		func(a DRPolicySummary, b DRPolicySummary) bool {
			return a.Equal(&b)
		},
	) {
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
	if !slices.Equal(d.Conditions, o.Conditions) {
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
	if s.S3SecretRef != o.S3SecretRef {
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
	if !slices.Equal(d.Conditions, o.Conditions) {
		return false
	}
	return true
}
