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
	Stale = ValidationState("stale 🟠")

	// Error state such as missing or unexpected value.
	Error = ValidationState("error ❌")
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
	Value []S3StoreProfilesSummary `json:"value"`
}

func (v *Validated) GetState() ValidationState {
	return v.State
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
