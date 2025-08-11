// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

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
	Value string `json:"value"`
}

// ValidatedBool is a validated object bool property.
type ValidatedBool struct {
	Validated
	Value bool `json:"value"`
}

// ValidatedCondition is a validated condition.
type ValidatedCondition struct {
	Validated
	Type string `json:"type"`
}

func (v *Validated) GetState() ValidationState {
	return v.State
}
