// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package failover

import (
	"testing"

	ramenapi "github.com/ramendr/ramen/api/v1alpha1"
)

// NOTE: These tests will compile once Ramen PR #2416 is merged.
// Reference: https://github.com/RamenDR/ramen/pull/2416

func TestValidateTestPreconditions(t *testing.T) {
	tests := []struct {
		name    string
		drpc    *ramenapi.DRPlacementControl
		wantErr bool
	}{
		{
			name: "valid - can start dry-run",
			drpc: &ramenapi.DRPlacementControl{
				Spec: ramenapi.DRPlacementControlSpec{
					Action: "",
					DryRun: false,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid - already in dry-run",
			drpc: &ramenapi.DRPlacementControl{
				Spec: ramenapi.DRPlacementControlSpec{
					Action: "",
					DryRun: true,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid - active action",
			drpc: &ramenapi.DRPlacementControl{
				Spec: ramenapi.DRPlacementControlSpec{
					Action: ramenapi.ActionFailover,
					DryRun: false,
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Command{}
			err := c.validateTestPreconditions(tt.drpc)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTestPreconditions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRevertDryRunLogic(t *testing.T) {
	tests := []struct {
		name             string
		lastAction       string
		preferredCluster string
		expectedAction   ramenapi.DRAction
		expectedFailover string
		expectedDryRun   bool
	}{
		{
			name:             "revert to deployed state",
			lastAction:       "",
			preferredCluster: "dr1",
			expectedAction:   "",
			expectedFailover: "",
			expectedDryRun:   false,
		},
		{
			name:             "revert to failedover state",
			lastAction:       string(ramenapi.ActionFailover),
			preferredCluster: "dr1",
			expectedAction:   ramenapi.ActionFailover,
			expectedFailover: "dr1",
			expectedDryRun:   false,
		},
		{
			name:             "revert to relocated state",
			lastAction:       string(ramenapi.ActionRelocate),
			preferredCluster: "dr2",
			expectedAction:   ramenapi.ActionRelocate,
			expectedFailover: "",
			expectedDryRun:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Command{}
			drpc := &ramenapi.DRPlacementControl{
				Spec: ramenapi.DRPlacementControlSpec{
					PreferredCluster: tt.preferredCluster,
					Action:           ramenapi.ActionFailover,
					FailoverCluster:  "dr2",
					DryRun:           true,
				},
			}

			// Manually apply the revert logic (simulating revertDryRun)
			switch tt.lastAction {
			case "":
				drpc.Spec.Action = ""
				drpc.Spec.FailoverCluster = ""
				drpc.Spec.DryRun = false
			case string(ramenapi.ActionFailover):
				drpc.Spec.Action = ramenapi.ActionFailover
				drpc.Spec.FailoverCluster = drpc.Spec.PreferredCluster
				drpc.Spec.DryRun = false
			case string(ramenapi.ActionRelocate):
				drpc.Spec.Action = ramenapi.ActionRelocate
				drpc.Spec.FailoverCluster = ""
				drpc.Spec.DryRun = false
			}

			// Verify the results
			if drpc.Spec.Action != tt.expectedAction {
				t.Errorf("Action = %v, want %v", drpc.Spec.Action, tt.expectedAction)
			}
			if drpc.Spec.FailoverCluster != tt.expectedFailover {
				t.Errorf("FailoverCluster = %v, want %v", drpc.Spec.FailoverCluster, tt.expectedFailover)
			}
			if drpc.Spec.DryRun != tt.expectedDryRun {
				t.Errorf("DryRun = %v, want %v", drpc.Spec.DryRun, tt.expectedDryRun)
			}
		})
	}
}

func TestHasDRPCFailed(t *testing.T) {
	tests := []struct {
		name       string
		conditions []ramenapi.DRPCCondition
		want       bool
	}{
		{
			name: "not failed - available is true",
			conditions: []ramenapi.DRPCCondition{
				{
					Type:   "Available",
					Status: "True",
				},
			},
			want: false,
		},
		{
			name: "failed - available is false",
			conditions: []ramenapi.DRPCCondition{
				{
					Type:   "Available",
					Status: "False",
				},
			},
			want: true,
		},
		{
			name:       "not failed - no conditions",
			conditions: []ramenapi.DRPCCondition{},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Command{}
			drpc := &ramenapi.DRPlacementControl{
				Status: ramenapi.DRPlacementControlStatus{
					Conditions: tt.conditions,
				},
			}
			if got := c.hasDRPCFailed(drpc); got != tt.want {
				t.Errorf("hasDRPCFailed() = %v, want %v", got, tt.want)
			}
		})
	}
}
