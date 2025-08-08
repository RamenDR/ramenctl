// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"maps"
	"slices"
)

// DRClusterSummary is the summary of a DRCluster.
type DRClusterSummary struct {
	Name       string            `json:"name"`
	Phase      string            `json:"phase"`
	Conditions map[string]string `json:"conditions,omitempty"`
}

// DRPolicySummary is the summary of a DRPolicy.
type DRPolicySummary struct {
	Name       string            `json:"name"`
	DRClusters string            `json:"drClusters"`
	Conditions map[string]string `json:"conditions,omitempty"`
}

// ConfigMapSummary is the summary of a Ramen ConfigMap.
type ConfigMapSummary struct {
	Name string `json:"name"`
	// TODO: Add relevant fields to validate
}

// OperatorSummary is the summary of a Ramen Operator.
type OperatorSummary struct {
	Name       string            `json:"name"`
	Conditions map[string]string `json:"conditions,omitempty"`
}

// RamenSummary is the summary of Ramen components.
type RamenSummary struct {
	ConfigMap ConfigMapSummary `json:"configmap"`
	Operator  OperatorSummary  `json:"operator"`
}

// HubClusterStatus is the cluster status on the hub.
type HubClusterStatus struct {
	DRClusters []DRClusterSummary `json:"drClusters"`
	DRPolicies []DRPolicySummary  `json:"drPolicies"`
	Ramen      RamenSummary       `json:"ramen"`
}

// DRClusterStatus is the cluster status on a managed cluster.
type DRClusterStatus struct {
	Name  string       `json:"name"`
	Ramen RamenSummary `json:"ramen"`
}

// ClusterStatus is cluster status in multi-cluster environment.
type ClusterStatus struct {
	Hub        HubClusterStatus  `json:"hub"`
	DRClusters []DRClusterStatus `json:"drClusters"`
}

func (c *ClusterStatus) Equal(o *ClusterStatus) bool {
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
		c.DRClusters,
		o.DRClusters,
		func(a DRClusterStatus, b DRClusterStatus) bool {
			return a.Equal(&b)
		},
	) {
		return false
	}
	return true
}

func (h *HubClusterStatus) Equal(o *HubClusterStatus) bool {
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

func (m *DRClusterStatus) Equal(o *DRClusterStatus) bool {
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
	if !maps.Equal(d.Conditions, o.Conditions) {
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
	if d.DRClusters != o.DRClusters {
		return false
	}
	if !maps.Equal(d.Conditions, o.Conditions) {
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
	if !r.Operator.Equal(&o.Operator) {
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
	return true
}

func (op *OperatorSummary) Equal(o *OperatorSummary) bool {
	if op == o {
		return true
	}
	if o == nil {
		return false
	}
	if op.Name != o.Name {
		return false
	}
	if !maps.Equal(op.Conditions, o.Conditions) {
		return false
	}
	return true
}
