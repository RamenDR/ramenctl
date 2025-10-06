// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"slices"

	corev1 "k8s.io/api/core/v1"
)

type ValidationState string

const (
	// OK is an expected value.
	OK = ValidationState("ok ✅")

	// Condition Generation does not match object generation.
	Stale = ValidationState("stale ⭕")

	// Problem state such as missing or unexpected value.
	Problem = ValidationState("problem ❌")
)

type Validation interface {
	GetState() ValidationState
}

type Validated struct {
	// State is the validation state (one of OK, Stale, Error).
	State ValidationState `json:"state"`
	// Description explains why the value is not OK.
	Description string `json:"description,omitempty"`
}

// ValidatedString is a validated object string property.
type ValidatedString struct {
	Validated
	Value string `json:"value,omitempty"`
}

// ValidatedBool is a validated object bool property.
type ValidatedBool struct {
	Validated
	Value bool `json:"value,omitempty"`
}

// ValidatedInteger is a validated object integer property.
type ValidatedInteger struct {
	Validated
	Value int64 `json:"value"`
}

// ValidatedCondition is a validated condition.
type ValidatedCondition struct {
	Validated
	Type string `json:"type"`
}

// ValidatedS3SecretRef is a validated S3 secret reference.
type ValidatedS3SecretRef struct {
	Validated
	Value corev1.SecretReference `json:"value"`
}

// ValidatedS3StoreProfilesList is a validated list of S3 store profiles.
type ValidatedS3StoreProfilesList struct {
	Validated
	Value []S3StoreProfilesSummary `json:"value,omitempty"`
}

// ValidatedDRClustersList is a validated list of DR clusters.
type ValidatedDRClustersList struct {
	Validated
	Value []DRClusterSummary `json:"value,omitempty"`
}

// ValidatedDRPoliciesList is a validated list of DR policies.
type ValidatedDRPoliciesList struct {
	Validated
	Value []DRPolicySummary `json:"value,omitempty"`
}

// ValidatedPeerClassesList is a validated list of peerClasses in a DRPolicy.
type ValidatedPeerClassesList struct {
	Validated
	Value []PeerClassesSummary `json:"value,omitempty"`
}

func (v *Validated) GetState() ValidationState {
	return v.State
}

func (v *ValidatedDRClustersList) Equal(o *ValidatedDRClustersList) bool {
	if v == o {
		return true
	}
	if o == nil {
		return false
	}
	if v.State != o.State {
		return false
	}
	if v.Description != o.Description {
		return false
	}
	if !slices.EqualFunc(
		v.Value,
		o.Value,
		func(a DRClusterSummary, b DRClusterSummary) bool {
			return a.Equal(&b)
		},
	) {
		return false
	}
	return true
}

func (v *ValidatedDRPoliciesList) Equal(o *ValidatedDRPoliciesList) bool {
	if v == o {
		return true
	}
	if o == nil {
		return false
	}
	if v.State != o.State {
		return false
	}
	if v.Description != o.Description {
		return false
	}
	if !slices.EqualFunc(
		v.Value,
		o.Value,
		func(a DRPolicySummary, b DRPolicySummary) bool {
			return a.Equal(&b)
		},
	) {
		return false
	}
	return true
}

func (v *ValidatedPeerClassesList) Equal(o *ValidatedPeerClassesList) bool {
	if v == o {
		return true
	}
	if o == nil {
		return false
	}
	if v.State != o.State {
		return false
	}
	if v.Description != o.Description {
		return false
	}
	if !slices.EqualFunc(
		v.Value,
		o.Value,
		func(a PeerClassesSummary, b PeerClassesSummary) bool {
			return a.Equal(&b)
		},
	) {
		return false
	}
	return true
}

func (v *ValidatedS3StoreProfilesList) Equal(o *ValidatedS3StoreProfilesList) bool {
	if v == o {
		return true
	}
	if o == nil {
		return false
	}
	if v.State != o.State {
		return false
	}
	if v.Description != o.Description {
		return false
	}
	if !slices.EqualFunc(
		v.Value,
		o.Value,
		func(a S3StoreProfilesSummary, b S3StoreProfilesSummary) bool {
			return a.Equal(&b)
		},
	) {
		return false
	}
	return true
}
