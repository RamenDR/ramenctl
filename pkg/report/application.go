// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"slices"
)

type ReplicationType string

const (
	Volrep   = ReplicationType("volrep")
	Volsync  = ReplicationType("volsync")
	External = ReplicationType("external")
)

// ProtectedPVCSummary is the summary of a protected PVC.
type ProtectedPVCSummary struct {
	Name        string               `json:"name"`
	Namespace   string               `json:"namespace"`
	Replication ReplicationType      `json:"replication,omitempty"`
	Deleted     bool                 `json:"deleted,omitempty"`
	Phase       string               `json:"phase,omitempty"`
	Conditions  []ValidatedCondition `json:"conditions,omitempty"`
}

// DRPCSummary is the summary of a DRPC.
type DRPCSummary struct {
	Name        string               `json:"name"`
	Namespace   string               `json:"namespace"`
	Deleted     bool                 `json:"deleted,omitempty"`
	DRPolicy    string               `json:"drPolicy"`
	Action      string               `json:"action,omitempty"`
	Phase       string               `json:"phase"`
	Progression string               `json:"progression"`
	Conditions  []ValidatedCondition `json:"conditions,omitempty"`
}

// VRGSummary is the summary of a VRG.
type VRGSummary struct {
	Name          string                `json:"name"`
	Namespace     string                `json:"namespace"`
	Deleted       bool                  `json:"deleted,omitempty"`
	State         string                `json:"state"`
	Conditions    []ValidatedCondition  `json:"conditions,omitempty"`
	ProtectedPVCs []ProtectedPVCSummary `json:"protectedPVCs,omitempty"`
}

// ApplicationHubStaus is the application status on the hub.
type HubApplicationStatus struct {
	DRPC DRPCSummary `json:"drpc"`
}

// ApplicationHubStaus is the application status on a managed cluster.
type ClusterApplicationStatus struct {
	Name string     `json:"name"`
	VRG  VRGSummary `json:"vrg"`
}

// ApplicationStatus is protected application status in multi-cluster environment.
type ApplicationStatus struct {
	Hub              HubApplicationStatus     `json:"hub"`
	PrimaryCluster   ClusterApplicationStatus `json:"primaryCluster"`
	SecondaryCluster ClusterApplicationStatus `json:"secondaryCluster"`
}

func (a *ApplicationStatus) Equal(o *ApplicationStatus) bool {
	if a == o {
		return true
	}
	if o == nil {
		return false
	}
	if !a.Hub.Equal(&o.Hub) {
		return false
	}
	if !a.PrimaryCluster.Equal(&o.PrimaryCluster) {
		return false
	}
	if !a.SecondaryCluster.Equal(&o.SecondaryCluster) {
		return false
	}
	return true
}

func (h *HubApplicationStatus) Equal(o *HubApplicationStatus) bool {
	if h == o {
		return true
	}
	if o == nil {
		return false
	}
	if !h.DRPC.Equal(&o.DRPC) {
		return false
	}
	return true
}

func (c *ClusterApplicationStatus) Equal(o *ClusterApplicationStatus) bool {
	if c == o {
		return true
	}
	if o == nil {
		return false
	}
	if c.Name != o.Name {
		return false
	}
	if !c.VRG.Equal(&o.VRG) {
		return false
	}
	return true
}

func (d *DRPCSummary) Equal(o *DRPCSummary) bool {
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
	if d.DRPolicy != o.DRPolicy {
		return false
	}
	if d.Action != o.Action {
		return false
	}
	if d.Phase != o.Phase {
		return false
	}
	if d.Progression != o.Progression {
		return false
	}
	if !slices.Equal(d.Conditions, o.Conditions) {
		return false
	}
	return true
}

func (v *VRGSummary) Equal(o *VRGSummary) bool {
	if v == o {
		return true
	}
	if o == nil {
		return false
	}
	if v.Name != o.Name {
		return false
	}
	if v.Namespace != o.Namespace {
		return false
	}
	if v.Deleted != o.Deleted {
		return false
	}
	if v.State != o.State {
		return false
	}
	if !slices.Equal(v.Conditions, o.Conditions) {
		return false
	}
	if !slices.EqualFunc(
		v.ProtectedPVCs,
		o.ProtectedPVCs,
		func(a ProtectedPVCSummary, b ProtectedPVCSummary) bool {
			return a.Equal(&b)
		},
	) {
		return false
	}
	return true
}

func (p *ProtectedPVCSummary) Equal(o *ProtectedPVCSummary) bool {
	if p == o {
		return true
	}
	if o == nil {
		return false
	}
	if p.Name != o.Name {
		return false
	}
	if p.Namespace != o.Namespace {
		return false
	}
	if p.Replication != o.Replication {
		return false
	}
	if p.Deleted != o.Deleted {
		return false
	}
	if p.Phase != o.Phase {
		return false
	}
	if !slices.Equal(p.Conditions, o.Conditions) {
		return false
	}
	return true
}
