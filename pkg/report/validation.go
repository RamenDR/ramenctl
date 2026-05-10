// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"encoding/json"
	"slices"
	"time"
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

// ValidatedDuration is a validated object duration property.
// Value marshals as a duration string (e.g. "1m0s") instead of nanoseconds.
type ValidatedDuration struct {
	Validated
	Value time.Duration `json:"value,omitempty"`
}

func (d ValidatedDuration) MarshalJSON() ([]byte, error) {
	var value string
	if d.Value != 0 {
		value = d.Value.String()
	}

	return json.Marshal(ValidatedString{Validated: d.Validated, Value: value})
}

func (d *ValidatedDuration) UnmarshalJSON(data []byte) error {
	var s ValidatedString
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	d.Validated = s.Validated
	d.Value = 0

	if s.Value != "" {
		duration, err := time.ParseDuration(s.Value)
		if err != nil {
			return err
		}

		d.Value = duration
	}

	return nil
}

// ValidatedTime is a validated object time property.
// Value is a pointer to properly omit zero values in JSON/YAML.
type ValidatedTime struct {
	Validated
	Value *time.Time `json:"value,omitempty"`
}

func (t *ValidatedTime) Equal(o *ValidatedTime) bool {
	if t == o {
		return true
	}
	if o == nil {
		return false
	}
	if t.Validated != o.Validated {
		return false
	}

	if t.Value != nil && o.Value != nil {
		if !t.Value.Equal(*o.Value) {
			return false
		}
	} else if t.Value != o.Value {
		return false
	}

	return true
}

// ValidatedCondition is a validated condition.
type ValidatedCondition struct {
	Validated
	Type string `json:"type"`
}

// ValidatedFingerprint is a validated fingerprint property.
type ValidatedFingerprint struct {
	Validated
	Value string `json:"value,omitempty"`
}

// ValidatedS3StoreProfilesList is a validated list of S3 store profiles.
type ValidatedS3StoreProfilesList struct {
	Validated
	Value []S3StoreProfilesSummary `json:"value,omitempty"`
}

// ValidatedApplicationS3ProfileStatusList is a validated list of S3 profile statuses.
type ValidatedApplicationS3ProfileStatusList struct {
	Validated
	Value []ApplicationS3ProfileStatus `json:"value,omitempty"`
}

// ValidatedClustersS3ProfileStatusList is a validated list of S3 profile statuses.
type ValidatedClustersS3ProfileStatusList struct {
	Validated
	Value []ClustersS3ProfileStatus `json:"value,omitempty"`
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

// MaxLen returns the maximum display length for truncation.
func (v *Validated) MaxLen() int {
	return 32
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

func (v *ValidatedApplicationS3ProfileStatusList) Equal(
	o *ValidatedApplicationS3ProfileStatusList,
) bool {
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
		func(a ApplicationS3ProfileStatus, b ApplicationS3ProfileStatus) bool {
			return a.Equal(&b)
		},
	) {
		return false
	}
	return true
}

func (v *ValidatedClustersS3ProfileStatusList) Equal(o *ValidatedClustersS3ProfileStatusList) bool {
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
		func(a ClustersS3ProfileStatus, b ClustersS3ProfileStatus) bool {
			return a.Equal(&b)
		},
	) {
		return false
	}
	return true
}
