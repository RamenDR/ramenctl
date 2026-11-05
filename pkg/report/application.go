// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"slices"

	"github.com/ramendr/ramenctl/pkg/time"
)

type ReplicationType string

const (
	Volrep   = ReplicationType("volrep")
	Volsync  = ReplicationType("volsync")
	External = ReplicationType("external")
)

// ProtectedPVCSummary is the summary of a protected PVC.
type ProtectedPVCSummary struct {
	// Name is the name of the PVC.
	Name string `json:"name"`
	// Namespace is the namespace of the PVC.
	Namespace string `json:"namespace"`
	// Replication is the replication type used to protect the PVC.
	Replication ReplicationType `json:"replication,omitempty"`
	// Deleted indicates whether the PVC has been deleted.
	Deleted ValidatedBool `json:"deleted"`
	// Phase is the validated phase of the PVC.
	Phase ValidatedString `json:"phase"`
	// Conditions are the validated conditions of the PVC.
	Conditions ValidatedConditionList `json:"conditions,omitempty"`
}

func (p *ProtectedPVCSummary) AggregateState() ValidationState {
	return aggregateState(&p.Phase, &p.Deleted, p.Conditions)
}

// ProtectedPVCList is a list of protected PVC summaries.
type ProtectedPVCList []ProtectedPVCSummary

func (p ProtectedPVCList) AggregateState() ValidationState {
	state := ValidationState("")
	for i := range p {
		state = significantState(state, p[i].AggregateState())
	}
	return state
}

// PVCGroupsSummary represents list of CGs that are protected by the VRG.
type PVCGroupsSummary struct {
	// Grouped is the list of PVC names grouped into this consistency group.
	Grouped []string `json:"grouped,omitempty"`
}

// DRPCSummary is the summary of a DRPC.
type DRPCSummary struct {
	// Name is the name of the DRPC.
	Name string `json:"name"`
	// Namespace is the namespace of the DRPC.
	Namespace string `json:"namespace"`
	// ClusterTime is the API server time when the DRPC was gathered.
	ClusterTime *time.Time `json:"clusterTime,omitempty"`
	// Deleted indicates whether the DRPC has been deleted.
	Deleted ValidatedBool `json:"deleted"`
	// DRPolicy is the name of the DRPolicy used by the DRPC.
	DRPolicy string `json:"drPolicy"`
	// SchedulingInterval is the replication scheduling interval from the DRPolicy.
	SchedulingInterval ValidatedDuration `json:"schedulingInterval"`
	// LastGroupSyncTime is the timestamp of the last successful volume group
	// replication sync, validated for freshness against SchedulingInterval.
	LastGroupSyncTime ValidatedTime `json:"lastGroupSyncTime"`
	// Action is the validated DR action currently set on the DRPC.
	Action ValidatedString `json:"action"`
	// Phase is the validated phase of the DRPC.
	Phase ValidatedString `json:"phase"`
	// Progression is the validated progression of the current DR action.
	Progression ValidatedString `json:"progression"`
	// Conditions are the validated conditions of the DRPC.
	Conditions ValidatedConditionList `json:"conditions,omitempty"`
}

// VRGSummary is the summary of a VRG.
type VRGSummary struct {
	// Name is the name of the VRG.
	Name string `json:"name"`
	// Namespace is the namespace of the VRG.
	Namespace string `json:"namespace"`
	// ClusterTime is the API server time when the VRG was gathered.
	ClusterTime *time.Time `json:"clusterTime,omitempty"`
	// Deleted indicates whether the VRG has been deleted.
	Deleted ValidatedBool `json:"deleted"`
	// SchedulingInterval is the replication scheduling interval for the VRG.
	SchedulingInterval ValidatedDuration `json:"schedulingInterval"`
	// LastGroupSyncTime is the timestamp of the last successful volume group
	// replication sync, validated for freshness against SchedulingInterval.
	LastGroupSyncTime ValidatedTime `json:"lastGroupSyncTime"`
	// State is the validated state of the VRG.
	State ValidatedString `json:"state"`
	// Conditions are the validated conditions of the VRG.
	Conditions ValidatedConditionList `json:"conditions,omitempty"`
	// ProtectedPVCs is the list of PVCs protected by the VRG.
	ProtectedPVCs ProtectedPVCList `json:"protectedPVCs,omitempty"`
	// PVCGroups is the list of consistency groups protected by the VRG.
	PVCGroups []PVCGroupsSummary `json:"pvcGroups,omitempty"`
}

// ApplicationHubStaus is the application status on the hub.
type ApplicationStatusHub struct {
	// DRPC is the summary of the application's DRPC on the hub.
	DRPC DRPCSummary `json:"drpc"`
}

// ApplicationHubStaus is the application status on a managed cluster.
type ApplicationStatusCluster struct {
	// Name is the name of the managed cluster.
	Name string `json:"name"`
	// VRG is the summary of the application's VRG on this cluster.
	VRG VRGSummary `json:"vrg"`
}

// ApplicationS3ProfileStatus is the status of an S3 profile.
type ApplicationS3ProfileStatus struct {
	// Name is the name of the S3 profile.
	Name string `json:"name"`
	// Gathered indicates whether data was gathered from this S3 profile.
	Gathered ValidatedBool `json:"gathered"`
}

// ApplicationS3Status is the status of all S3 profiles.
type ApplicationS3Status struct {
	// Profiles is the validated list of S3 profile statuses.
	Profiles ValidatedApplicationS3ProfileStatusList `json:"profiles"`
}

// ApplicationStatus is protected application status in multi-cluster environment.
type ApplicationStatus struct {
	// Hub is the application status on the hub cluster.
	Hub ApplicationStatusHub `json:"hub"`
	// PrimaryCluster is the application status on the primary managed cluster.
	PrimaryCluster ApplicationStatusCluster `json:"primaryCluster"`
	// SecondaryCluster is the application status on the secondary managed cluster.
	SecondaryCluster ApplicationStatusCluster `json:"secondaryCluster"`
	// S3 is the status of the application's S3 profiles.
	S3 ApplicationS3Status `json:"s3"`
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
	if !a.S3.Equal(&o.S3) {
		return false
	}
	return true
}

func (h *ApplicationStatusHub) Equal(o *ApplicationStatusHub) bool {
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

func (c *ApplicationStatusCluster) Equal(o *ApplicationStatusCluster) bool {
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
	if d.ClusterTime != nil && o.ClusterTime != nil {
		if !d.ClusterTime.Equal(*o.ClusterTime) {
			return false
		}
	} else if d.ClusterTime != o.ClusterTime {
		return false
	}
	if d.Deleted != o.Deleted {
		return false
	}
	if d.DRPolicy != o.DRPolicy {
		return false
	}
	if d.SchedulingInterval != o.SchedulingInterval {
		return false
	}
	if !d.LastGroupSyncTime.Equal(&o.LastGroupSyncTime) {
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
	if v.ClusterTime != nil && o.ClusterTime != nil {
		if !v.ClusterTime.Equal(*o.ClusterTime) {
			return false
		}
	} else if v.ClusterTime != o.ClusterTime {
		return false
	}
	if v.Deleted != o.Deleted {
		return false
	}
	if v.SchedulingInterval != o.SchedulingInterval {
		return false
	}
	if !v.LastGroupSyncTime.Equal(&o.LastGroupSyncTime) {
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
	if !slices.EqualFunc(
		v.PVCGroups,
		o.PVCGroups,
		func(a PVCGroupsSummary, b PVCGroupsSummary) bool {
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

func (p *PVCGroupsSummary) Equal(o *PVCGroupsSummary) bool {
	if p == o {
		return true
	}
	if o == nil {
		return false
	}
	if !slices.Equal(p.Grouped, o.Grouped) {
		return false
	}
	return true
}

func (s *ApplicationS3ProfileStatus) Equal(o *ApplicationS3ProfileStatus) bool {
	if s == o {
		return true
	}
	if o == nil {
		return false
	}
	if s.Name != o.Name {
		return false
	}
	if s.Gathered != o.Gathered {
		return false
	}
	return true
}

func (s *ApplicationS3Status) Equal(o *ApplicationS3Status) bool {
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
