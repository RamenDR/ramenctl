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
	Deleted     ValidatedBool        `json:"deleted"`
	Phase       ValidatedString      `json:"phase"`
	Conditions  []ValidatedCondition `json:"conditions,omitempty"`
}

// PVCGroupsSummary represents list of CGs that are protected by the VRG.
type PVCGroupsSummary struct {
	Grouped []string `json:"grouped,omitempty"`
}

// DRPCSummary is the summary of a DRPC.
type DRPCSummary struct {
	Name        string               `json:"name"`
	Namespace   string               `json:"namespace"`
	Deleted     ValidatedBool        `json:"deleted"`
	DRPolicy    string               `json:"drPolicy"`
	Action      ValidatedString      `json:"action"`
	Phase       ValidatedString      `json:"phase"`
	Progression ValidatedString      `json:"progression"`
	Conditions  []ValidatedCondition `json:"conditions,omitempty"`
}

// VRGSummary is the summary of a VRG.
type VRGSummary struct {
	Name          string                `json:"name"`
	Namespace     string                `json:"namespace"`
	Deleted       ValidatedBool         `json:"deleted"`
	State         ValidatedString       `json:"state"`
	Conditions    []ValidatedCondition  `json:"conditions,omitempty"`
	ProtectedPVCs []ProtectedPVCSummary `json:"protectedPVCs,omitempty"`
	PVCGroups     []PVCGroupsSummary    `json:"pvcGroups,omitempty"`
}

// ApplicationHubStaus is the application status on the hub.
type ApplicationStatusHub struct {
	DRPC DRPCSummary `json:"drpc"`
}

// ApplicationHubStaus is the application status on a managed cluster.
type ApplicationStatusCluster struct {
	Name string     `json:"name"`
	VRG  VRGSummary `json:"vrg"`
}

// ApplicationS3ProfileStatus is the status of an S3 profile.
type ApplicationS3ProfileStatus struct {
	Name     string        `json:"name"`
	Gathered ValidatedBool `json:"gathered"`
}

// ApplicationS3Status is the status of all S3 profiles.
type ApplicationS3Status struct {
	Profiles ValidatedApplicationS3ProfileStatusList `json:"profiles"`
}

// ApplicationStatus is protected application status in multi-cluster environment.
type ApplicationStatus struct {
	Hub              ApplicationStatusHub     `json:"hub"`
	PrimaryCluster   ApplicationStatusCluster `json:"primaryCluster"`
	SecondaryCluster ApplicationStatusCluster `json:"secondaryCluster"`
	S3               ApplicationS3Status      `json:"s3"`
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
