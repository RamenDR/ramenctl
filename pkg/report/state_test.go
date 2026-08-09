// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"testing"
)

func TestIsIssue(t *testing.T) {
	cases := []struct {
		state ValidationState
		want  bool
	}{
		{OK, false},
		{Warning, true},
		{Problem, true},
		{ValidationState(""), false},
	}
	for _, tc := range cases {
		t.Run(string(tc.state), func(t *testing.T) {
			if got := tc.state.IsIssue(); got != tc.want {
				t.Errorf("IsIssue() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSignificantState(t *testing.T) {
	cases := []struct {
		name string
		a, b ValidationState
		want ValidationState
	}{
		{"empty+empty", "", "", ""},
		{"empty+ok", "", OK, OK},
		{"ok+empty", OK, "", OK},
		{"ok+ok", OK, OK, OK},
		{"ok+warning", OK, Warning, Warning},
		{"warning+ok", Warning, OK, Warning},
		{"ok+problem", OK, Problem, Problem},
		{"problem+ok", Problem, OK, Problem},
		{"warning+problem", Warning, Problem, Problem},
		{"problem+warning", Problem, Warning, Problem},
		{"warning+warning", Warning, Warning, Warning},
		{"problem+problem", Problem, Problem, Problem},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := significantState(tc.a, tc.b); got != tc.want {
				t.Errorf("significantState(%q, %q) = %q, want %q", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestAggregateState(t *testing.T) {
	empty := Validated{}
	ok := Validated{State: OK}
	warning := Validated{State: Warning}
	problem := Validated{State: Problem}

	cases := []struct {
		name  string
		items []StateAggregator
		want  ValidationState
	}{
		{"nil", nil, ""},
		{"some empty", []StateAggregator{&empty, &empty}, ""},
		{"empty and ok", []StateAggregator{&empty, &ok}, OK},
		{"ok and warning", []StateAggregator{&ok, &ok, &warning}, Warning},
		{"ok and problem", []StateAggregator{&ok, &ok, &problem}, Problem},
		{"ok warning and problem", []StateAggregator{&ok, &warning, &problem}, Problem},
		// Ensure earlier significant state is not overridden by later ok.
		{"problem then ok", []StateAggregator{&problem, &ok}, Problem},
		{"warning then ok", []StateAggregator{&warning, &ok}, Warning},
		// Ensure earlier significant state is not overridden by later warning.
		{"problem then warning", []StateAggregator{&problem, &warning}, Problem},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := aggregateState(tc.items...); got != tc.want {
				t.Errorf("aggregateState() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestValidatedConditionListAggregateState(t *testing.T) {
	cases := []struct {
		name string
		list ValidatedConditionList
		want ValidationState
	}{
		{"empty", nil, ""},
		{"all ok", ValidatedConditionList{
			{Validated: Validated{State: OK}, Type: "Ready"},
			{Validated: Validated{State: OK}, Type: "Available"},
		}, OK},
		{"one warning", ValidatedConditionList{
			{Validated: Validated{State: OK}, Type: "Ready"},
			{Validated: Validated{State: Warning}, Type: "Available"},
		}, Warning},
		{"one problem", ValidatedConditionList{
			{Validated: Validated{State: OK}, Type: "Ready"},
			{Validated: Validated{State: Problem}, Type: "Available"},
		}, Problem},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.list.AggregateState(); got != tc.want {
				t.Errorf("AggregateState() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestProtectedPVCListAggregateState(t *testing.T) {
	cases := []struct {
		name string
		list ProtectedPVCList
		want ValidationState
	}{
		{"empty", nil, ""},
		{"all ok", ProtectedPVCList{
			{
				Phase: ValidatedString{Validated: Validated{State: OK}},
				Conditions: ValidatedConditionList{
					{Validated: Validated{State: OK}, Type: "DataReady"},
				},
			},
			{
				Phase: ValidatedString{Validated: Validated{State: OK}},
				Conditions: ValidatedConditionList{
					{Validated: Validated{State: OK}, Type: "DataReady"},
				},
			},
		}, OK},
		{"pvc problem", ProtectedPVCList{
			{
				Phase: ValidatedString{Validated: Validated{State: Problem}},
				Conditions: ValidatedConditionList{
					{Validated: Validated{State: OK}, Type: "DataReady"},
				},
			},
		}, Problem},
		{"condition problem", ProtectedPVCList{
			{
				Phase: ValidatedString{Validated: Validated{State: OK}},
				Conditions: ValidatedConditionList{
					{Validated: Validated{State: OK}, Type: "DataReady"},
					{Validated: Validated{State: Problem}, Type: "ClusterDataProtected"},
				},
			},
		}, Problem},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.list.AggregateState(); got != tc.want {
				t.Errorf("AggregateState() = %q, want %q", got, tc.want)
			}
		})
	}
}
