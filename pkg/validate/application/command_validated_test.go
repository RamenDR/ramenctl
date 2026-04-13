// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"fmt"
	"testing"

	ramenapi "github.com/ramendr/ramen/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"

	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/validate/summary"
)

// Test individual application validation functions without running the full
// command flow.

func TestValidatedDRPCAction(t *testing.T) {
	cmd := testCommand(t, validateApplication, &helpers.ValidationMock{}, testK8s)
	known := []struct {
		name   string
		action string
	}{
		{"empty action", ""},
		{"failover action", string(ramenapi.ActionFailover)},
		{"relocate action", string(ramenapi.ActionRelocate)},
	}
	for _, tc := range known {
		t.Run(tc.name, func(t *testing.T) {
			expected := report.ValidatedString{
				Value: tc.action,
				Validated: report.Validated{
					State: report.OK,
				},
			}
			validated := cmd.validatedDRPCAction(tc.action)
			if validated != expected {
				t.Errorf("expected action %+v, got %+v", expected, validated)
			}
		})
	}

	t.Run("unknown action", func(t *testing.T) {
		action := "Failback"
		expected := report.ValidatedString{
			Value: action,
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Unknown action \"Failback\"",
			},
		}
		validated := cmd.validatedDRPCAction(action)
		if validated != expected {
			t.Fatalf("expected action %+v, got %+v", expected, validated)
		}
	})

	t.Run("update summary", func(t *testing.T) {
		expected := report.Summary{summary.OK: 3, summary.Problem: 1}
		if !cmd.Report.Summary.Equal(&expected) {
			t.Fatalf("expected summary %v, got %v", expected, *cmd.Report.Summary)
		}
	})
}

func TestValidatedDRPCPhaseError(t *testing.T) {
	type testcase struct {
		name   string
		action ramenapi.DRAction
		phase  ramenapi.DRState
	}

	cmd := testCommand(t, validateApplication, &helpers.ValidationMock{}, testK8s)

	unstable := []struct {
		stable ramenapi.DRState
		cases  []testcase
	}{
		// No action error phases.
		{
			stable: ramenapi.Deployed,
			cases: []testcase{
				{"empty initiating", "", ramenapi.Initiating},
				{"empty deleting", "", ramenapi.Deploying},
				{"empty deleting", "", ramenapi.Deleting},
				{"empty failed over", "", ramenapi.FailedOver},
				{"empty relocated", "", ramenapi.Relocated},
			},
		},
		// Error failover phases.
		{
			stable: ramenapi.FailedOver,
			cases: []testcase{
				{"failover failing over", ramenapi.ActionFailover, ramenapi.FailingOver},
				{"failover wait for user", ramenapi.ActionFailover, ramenapi.WaitForUser},
				{"failover deleting", ramenapi.ActionFailover, ramenapi.Deleting},
				{"failover deployed", ramenapi.ActionFailover, ramenapi.Deployed},
				{"failover relocated", ramenapi.ActionFailover, ramenapi.Relocated},
			},
		},
		// Error relocate phases.
		{
			stable: ramenapi.Relocated,
			cases: []testcase{
				{"relocate relocating", ramenapi.ActionRelocate, ramenapi.Relocating},
				{"relocate wait for user", ramenapi.ActionRelocate, ramenapi.WaitForUser},
				{"relocate deleting", ramenapi.ActionRelocate, ramenapi.Deleting},
				{"relocate deployed", ramenapi.ActionRelocate, ramenapi.Deployed},
				{"relocate failed over", ramenapi.ActionRelocate, ramenapi.FailedOver},
			},
		},
	}

	for _, group := range unstable {
		for _, tc := range group.cases {
			t.Run(tc.name, func(t *testing.T) {
				drpc := &ramenapi.DRPlacementControl{
					Spec: ramenapi.DRPlacementControlSpec{
						Action: tc.action,
					},
					Status: ramenapi.DRPlacementControlStatus{
						Phase: tc.phase,
					},
				}
				expected := report.ValidatedString{
					Validated: report.Validated{
						State:       report.Problem,
						Description: fmt.Sprintf("Waiting for stable phase %q", group.stable),
					},
					Value: string(tc.phase),
				}
				validated := cmd.validatedDRPCPhase(drpc)
				if validated != expected {
					t.Errorf("expected phase %+v, got %+v", expected, validated)
				}
			})
		}
	}

	var errors int
	for _, group := range unstable {
		errors += len(group.cases)
	}
	expected := report.Summary{summary.Problem: errors}
	if !cmd.Report.Summary.Equal(&expected) {
		t.Fatalf("expected summary %v, got %v", expected, *cmd.Report.Summary)
	}
}

func TestValidatedDRPCPhaseOK(t *testing.T) {
	cmd := testCommand(t, validateApplication, &helpers.ValidationMock{}, testK8s)

	cases := []struct {
		name   string
		action ramenapi.DRAction
		phase  ramenapi.DRState
	}{
		{"empty deployed", "", ramenapi.Deployed},
		{"failover failed over", ramenapi.ActionFailover, ramenapi.FailedOver},
		{"relocate relocated", ramenapi.ActionRelocate, ramenapi.Relocated},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			drpc := &ramenapi.DRPlacementControl{
				Spec: ramenapi.DRPlacementControlSpec{
					Action: tc.action,
				},
				Status: ramenapi.DRPlacementControlStatus{
					Phase: tc.phase,
				},
			}
			expected := report.ValidatedString{
				Validated: report.Validated{
					State: report.OK,
				},
				Value: string(tc.phase),
			}
			validated := cmd.validatedDRPCPhase(drpc)
			if validated != expected {
				t.Errorf("expected phase %+v, got %+v", expected, validated)
			}
		})
	}

	expected := report.Summary{summary.OK: len(cases)}
	if !cmd.Report.Summary.Equal(&expected) {
		t.Fatalf("expected summary %v, got %v", expected, *cmd.Report.Summary)
	}
}

func TestValidatedDRPCProgressionOK(t *testing.T) {
	cmd := testCommand(t, validateApplication, &helpers.ValidationMock{}, testK8s)
	progression := ramenapi.ProgressionCompleted

	t.Run(string(progression), func(t *testing.T) {
		drpc := &ramenapi.DRPlacementControl{
			Status: ramenapi.DRPlacementControlStatus{
				Progression: progression,
			},
		}
		expected := report.ValidatedString{
			Validated: report.Validated{
				State: report.OK,
			},
			Value: string(progression),
		}
		validated := cmd.validatedDRPCProgression(drpc)
		if validated != expected {
			t.Errorf("expected phase %+v, got %+v", expected, validated)
		}
	})

	expected := report.Summary{summary.OK: 1}
	if !cmd.Report.Summary.Equal(&expected) {
		t.Fatalf("expected summary %v, got %v", expected, *cmd.Report.Summary)
	}
}

func TestValidatedDRPCProgressionError(t *testing.T) {
	cmd := testCommand(t, validateApplication, &helpers.ValidationMock{}, testK8s)

	progressions := []ramenapi.ProgressionStatus{
		ramenapi.ProgressionCreatingMW,
		ramenapi.ProgressionUpdatingPlRule,
		ramenapi.ProgressionWaitForReadiness,
		ramenapi.ProgressionCleaningUp,
		ramenapi.ProgressionWaitOnUserToCleanUp,
		ramenapi.ProgressionCheckingFailoverPrerequisites,
		ramenapi.ProgressionFailingOverToCluster,
		ramenapi.ProgressionWaitForFencing,
		ramenapi.ProgressionWaitForStorageMaintenanceActivation,
		ramenapi.ProgressionPreparingFinalSync,
		ramenapi.ProgressionClearingPlacement,
		ramenapi.ProgressionRunningFinalSync,
		ramenapi.ProgressionFinalSyncComplete,
		ramenapi.ProgressionEnsuringVolumesAreSecondary,
		ramenapi.ProgressionWaitingForResourceRestore,
		ramenapi.ProgressionUpdatedPlacement,
		ramenapi.ProgressionEnsuringVolSyncSetup,
		ramenapi.ProgressionSettingupVolsyncDest,
		ramenapi.ProgressionDeleting,
		ramenapi.ProgressionDeleted,
		ramenapi.ProgressionActionPaused,
	}

	for _, progression := range progressions {
		t.Run(string(progression), func(t *testing.T) {
			drpc := &ramenapi.DRPlacementControl{
				Status: ramenapi.DRPlacementControlStatus{
					Progression: progression,
				},
			}
			expected := report.ValidatedString{
				Validated: report.Validated{
					State: report.Problem,
					Description: fmt.Sprintf(
						"Waiting for progression %q",
						ramenapi.ProgressionCompleted,
					),
				},
				Value: string(drpc.Status.Progression),
			}
			validated := cmd.validatedDRPCProgression(drpc)
			if validated != expected {
				t.Errorf("expected phase %+v, got %+v", expected, validated)
			}
		})
	}

	expected := report.Summary{summary.Problem: len(progressions)}
	if !cmd.Report.Summary.Equal(&expected) {
		t.Fatalf("expected summary %v, got %v", expected, *cmd.Report.Summary)
	}
}

func TestValidatedVRGSTateOK(t *testing.T) {
	cmd := testCommand(t, validateApplication, &helpers.ValidationMock{}, testK8s)

	cases := []struct {
		name        string
		stableState ramenapi.State
	}{
		{"primary", ramenapi.PrimaryState},
		{"secondary", ramenapi.SecondaryState},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vrg := &ramenapi.VolumeReplicationGroup{
				Status: ramenapi.VolumeReplicationGroupStatus{
					State: tc.stableState,
				},
			}
			expected := report.ValidatedString{
				Validated: report.Validated{
					State: report.OK,
				},
				Value: string(vrg.Status.State),
			}
			validated := cmd.validatedVRGState(vrg, tc.stableState)
			if validated != expected {
				t.Errorf("expected state %+v, got %+v", expected, validated)
			}
		})
	}

	expected := report.Summary{summary.OK: len(cases)}
	if !cmd.Report.Summary.Equal(&expected) {
		t.Fatalf("expected summary %v, got %v", expected, *cmd.Report.Summary)
	}
}

func TestValidatedVRGSTateError(t *testing.T) {
	cmd := testCommand(t, validateApplication, &helpers.ValidationMock{}, testK8s)

	cases := []struct {
		name        string
		state       ramenapi.State
		stableState ramenapi.State
	}{
		{"primary empty", "", ramenapi.PrimaryState},
		{"primary unknown", ramenapi.UnknownState, ramenapi.PrimaryState},
		{"primary secondary", ramenapi.SecondaryState, ramenapi.PrimaryState},
		{"secondary empty", "", ramenapi.SecondaryState},
		{"secondary unknown", ramenapi.UnknownState, ramenapi.SecondaryState},
		{"secondary primary", ramenapi.PrimaryState, ramenapi.SecondaryState},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			vrg := &ramenapi.VolumeReplicationGroup{
				Status: ramenapi.VolumeReplicationGroupStatus{
					State: tc.state,
				},
			}
			expected := report.ValidatedString{
				Validated: report.Validated{
					State:       report.Problem,
					Description: fmt.Sprintf("Waiting to become %q", tc.stableState),
				},
				Value: string(vrg.Status.State),
			}
			validated := cmd.validatedVRGState(vrg, tc.stableState)
			if validated != expected {
				t.Errorf("expected state %+v, got %+v", expected, validated)
			}
		})
	}

	expected := report.Summary{summary.Problem: len(cases)}
	if !cmd.Report.Summary.Equal(&expected) {
		t.Fatalf("expected summary %v, got %v", expected, *cmd.Report.Summary)
	}
}

func TestValidatedProtectedPVCOK(t *testing.T) {
	cmd := testCommand(t, validateApplication, &helpers.ValidationMock{}, testK8s)

	t.Run("bound", func(t *testing.T) {
		pvc := &corev1.PersistentVolumeClaim{
			Status: corev1.PersistentVolumeClaimStatus{
				Phase: corev1.ClaimBound,
			},
		}
		expected := report.ValidatedString{
			Validated: report.Validated{
				State: report.OK,
			},
			Value: string(pvc.Status.Phase),
		}
		validated := cmd.validatedProtectedPVCPhase(pvc)
		if validated != expected {
			t.Errorf("expected phase %+v, got %+v", expected, validated)
		}
	})

	expected := report.Summary{summary.OK: 1}
	if !cmd.Report.Summary.Equal(&expected) {
		t.Fatalf("expected summary %v, got %v", expected, *cmd.Report.Summary)
	}
}

func TestValidatedProtectedPVCError(t *testing.T) {
	cmd := testCommand(t, validateApplication, &helpers.ValidationMock{}, testK8s)

	cases := []struct {
		name  string
		phase corev1.PersistentVolumeClaimPhase
	}{
		{"empty", ""},
		{"pending", corev1.ClaimPending},
		{"lost", corev1.ClaimLost},
		{"terminating", "Terminating"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			pvc := &corev1.PersistentVolumeClaim{
				Status: corev1.PersistentVolumeClaimStatus{
					Phase: tc.phase,
				},
			}
			expected := report.ValidatedString{
				Validated: report.Validated{
					State:       report.Problem,
					Description: fmt.Sprintf("PVC is not %q", corev1.ClaimBound),
				},
				Value: string(pvc.Status.Phase),
			}
			validated := cmd.validatedProtectedPVCPhase(pvc)
			if validated != expected {
				t.Errorf("expected phase %+v, got %+v", expected, validated)
			}
		})
	}

	expected := report.Summary{summary.Problem: len(cases)}
	if !cmd.Report.Summary.Equal(&expected) {
		t.Fatalf("expected summary %v, got %v", expected, *cmd.Report.Summary)
	}
}
