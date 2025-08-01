// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"maps"
	"slices"
)

// ProtectedPVCSummary is the summary of a protected PVC.
type ProtectedPVCSummary struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Deleted    bool              `json:"deleted,omitempty"`
	Phase      string            `json:"phase,omitempty"`
	Conditions map[string]string `json:"conditions,omitempty"`
}

// DRPCSummary is the summary of a DRPC.
type DRPCSummary struct {
	Deleted     bool              `json:"deleted,omitempty"`
	DRPolicy    string            `json:"drPolicy"`
	Action      string            `json:"action,omitempty"`
	Phase       string            `json:"phase"`
	Progression string            `json:"progression"`
	Conditions  map[string]string `json:"conditions,omitempty"`
}

// VRGSummary is the summary of a VRG.
type VRGSummary struct {
	Namespace     string                `json:"namespace"`
	Deleted       bool                  `json:"deleted,omitempty"`
	State         string                `json:"state"`
	Conditions    map[string]string     `json:"conditions,omitempty"`
	ProtectedPVCs []ProtectedPVCSummary `json:"protectedPVCs,omitempty"`
}

// ApplicationHubStaus is the application status on the hub.
type ApplicationHubStatus struct {
	DRPC DRPCSummary `json:"drpc"`
}

// ApplicationHubStaus is the application status on a managed cluster.
type ApplicationClusterStatus struct {
	Name string     `json:"name"`
	VRG  VRGSummary `json:"vrg"`
}

// Application is application info.
type Application struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`

	// Optional info for `validate application` command.
	Hub              *ApplicationHubStatus     `json:"hub,omitempty"`
	PrimaryCluster   *ApplicationClusterStatus `json:"primary-cluster,omitempty"`
	SecondaryCluster *ApplicationClusterStatus `json:"secondary-cluster,omitempty"`
}

func (a *Application) Equal(o *Application) bool {
	if a == o {
		return true
	}
	if o == nil {
		return false
	}
	if a.Name != o.Name {
		return false
	}
	if a.Namespace != o.Namespace {
		return false
	}
	if a.Hub != nil {
		if !a.Hub.Equal(o.Hub) {
			return false
		}
	} else if o.Hub != nil {
		return false
	}
	if a.PrimaryCluster != nil {
		if !a.PrimaryCluster.Equal(o.PrimaryCluster) {
			return false
		}
	} else if o.PrimaryCluster != nil {
		return false
	}
	if a.SecondaryCluster != nil {
		if !a.SecondaryCluster.Equal(o.SecondaryCluster) {
			return false
		}
	} else if o.SecondaryCluster != nil {
		return false
	}
	return true
}

func (h *ApplicationHubStatus) Equal(o *ApplicationHubStatus) bool {
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

func (c *ApplicationClusterStatus) Equal(o *ApplicationClusterStatus) bool {
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
	if !maps.Equal(d.Conditions, o.Conditions) {
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
	if v.Namespace != o.Namespace {
		return false
	}
	if v.Deleted != o.Deleted {
		return false
	}
	if v.State != o.State {
		return false
	}
	if !maps.Equal(v.Conditions, o.Conditions) {
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
	if p.Deleted != o.Deleted {
		return false
	}
	if p.Phase != o.Phase {
		return false
	}
	if !maps.Equal(p.Conditions, o.Conditions) {
		return false
	}
	return true
}
