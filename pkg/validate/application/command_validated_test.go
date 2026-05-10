// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	stdtime "time"

	ramenapi "github.com/ramendr/ramen/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/validate/summary"
)

// Test individual application validation functions without running the full
// command flow.

func TestValidatedDRPCAction(t *testing.T) {
	cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)
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

	cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

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
	cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

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
	cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)
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
	cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

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
	cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

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
	cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

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
	cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

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
	cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

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

func TestValidatedDRPCSchedulingInterval(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)
		writeDRPolicy(t, cmd.DataDir(), "dr-policy-5m", "5m")

		drpc := &ramenapi.DRPlacementControl{
			Spec: ramenapi.DRPlacementControlSpec{
				DRPolicyRef: corev1.ObjectReference{Name: "dr-policy-5m"},
			},
		}
		expected := report.ValidatedDuration{
			Validated: report.Validated{State: report.OK},
			Value:     5 * stdtime.Minute,
		}
		validated := cmd.validatedDRPCSchedulingInterval(drpc)
		if validated != expected {
			t.Fatalf("unexpected result\n%s", helpers.UnifiedDiff(t, expected, validated))
		}
	})

	t.Run("drpolicy not found", func(t *testing.T) {
		cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

		drpc := &ramenapi.DRPlacementControl{
			Spec: ramenapi.DRPlacementControlSpec{
				DRPolicyRef: corev1.ObjectReference{Name: "no-such-policy"},
			},
		}
		expected := report.ValidatedDuration{
			Validated: report.Validated{
				State:       report.Problem,
				Description: `Could not read drpolicy "no-such-policy"`,
			},
		}
		validated := cmd.validatedDRPCSchedulingInterval(drpc)
		if validated != expected {
			t.Fatalf("unexpected result\n%s", helpers.UnifiedDiff(t, expected, validated))
		}
	})

	t.Run("empty scheduling interval", func(t *testing.T) {
		cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)
		writeDRPolicy(t, cmd.DataDir(), "dr-policy-empty", "")

		drpc := &ramenapi.DRPlacementControl{
			Spec: ramenapi.DRPlacementControlSpec{
				DRPolicyRef: corev1.ObjectReference{Name: "dr-policy-empty"},
			},
		}
		expected := report.ValidatedDuration{
			Validated: report.Validated{
				State:       report.Problem,
				Description: `Missing scheduling interval in drpolicy "dr-policy-empty"`,
			},
		}
		validated := cmd.validatedDRPCSchedulingInterval(drpc)
		if validated != expected {
			t.Fatalf("unexpected result\n%s", helpers.UnifiedDiff(t, expected, validated))
		}
	})

	t.Run("invalid scheduling interval", func(t *testing.T) {
		cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)
		writeDRPolicy(t, cmd.DataDir(), "dr-policy-bad", "invalid")

		drpc := &ramenapi.DRPlacementControl{
			Spec: ramenapi.DRPlacementControlSpec{
				DRPolicyRef: corev1.ObjectReference{Name: "dr-policy-bad"},
			},
		}
		expected := report.ValidatedDuration{
			Validated: report.Validated{
				State:       report.Problem,
				Description: `Invalid scheduling interval in drpolicy "dr-policy-bad"`,
			},
		}
		validated := cmd.validatedDRPCSchedulingInterval(drpc)
		if validated != expected {
			t.Fatalf("unexpected result\n%s", helpers.UnifiedDiff(t, expected, validated))
		}
	})
}

func TestValidatedVRGSchedulingInterval(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

		vrg := &ramenapi.VolumeReplicationGroup{
			Spec: ramenapi.VolumeReplicationGroupSpec{
				Async: &ramenapi.VRGAsyncSpec{
					SchedulingInterval: "5m",
				},
			},
		}
		drpcSummary := &report.DRPCSummary{
			DRPolicy: "dr-policy-5m",
			SchedulingInterval: report.ValidatedDuration{
				Validated: report.Validated{State: report.OK},
				Value:     5 * stdtime.Minute,
			},
		}
		expected := report.ValidatedDuration{
			Validated: report.Validated{State: report.OK},
			Value:     5 * stdtime.Minute,
		}
		validated := cmd.validatedVRGSchedulingInterval(vrg, drpcSummary)
		if validated != expected {
			t.Fatalf("unexpected result\n%s", helpers.UnifiedDiff(t, expected, validated))
		}
	})

	t.Run("metro dr", func(t *testing.T) {
		cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

		vrg := &ramenapi.VolumeReplicationGroup{}
		drpcSummary := &report.DRPCSummary{}

		expected := report.ValidatedDuration{}
		validated := cmd.validatedVRGSchedulingInterval(vrg, drpcSummary)
		if validated != expected {
			t.Fatalf("unexpected result\n%s", helpers.UnifiedDiff(t, expected, validated))
		}
	})

	t.Run("empty scheduling interval", func(t *testing.T) {
		cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

		vrg := &ramenapi.VolumeReplicationGroup{
			Spec: ramenapi.VolumeReplicationGroupSpec{
				Async: &ramenapi.VRGAsyncSpec{
					SchedulingInterval: "",
				},
			},
		}
		drpcSummary := &report.DRPCSummary{}

		expected := report.ValidatedDuration{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Missing scheduling interval in vrg",
			},
		}
		validated := cmd.validatedVRGSchedulingInterval(vrg, drpcSummary)
		if validated != expected {
			t.Fatalf("unexpected result\n%s", helpers.UnifiedDiff(t, expected, validated))
		}
	})

	t.Run("invalid scheduling interval", func(t *testing.T) {
		cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

		vrg := &ramenapi.VolumeReplicationGroup{
			Spec: ramenapi.VolumeReplicationGroupSpec{
				Async: &ramenapi.VRGAsyncSpec{
					SchedulingInterval: "invalid",
				},
			},
		}
		drpcSummary := &report.DRPCSummary{}

		expected := report.ValidatedDuration{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Invalid scheduling interval in vrg",
			},
		}
		validated := cmd.validatedVRGSchedulingInterval(vrg, drpcSummary)
		if validated != expected {
			t.Fatalf("unexpected result\n%s", helpers.UnifiedDiff(t, expected, validated))
		}
	})

	t.Run("mismatch with drpolicy", func(t *testing.T) {
		cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

		vrg := &ramenapi.VolumeReplicationGroup{
			Spec: ramenapi.VolumeReplicationGroupSpec{
				Async: &ramenapi.VRGAsyncSpec{
					SchedulingInterval: "1m",
				},
			},
		}
		drpcSummary := &report.DRPCSummary{
			DRPolicy: "dr-policy-5m",
			SchedulingInterval: report.ValidatedDuration{
				Validated: report.Validated{State: report.OK},
				Value:     5 * stdtime.Minute,
			},
		}
		expected := report.ValidatedDuration{
			Validated: report.Validated{
				State: report.Problem,
				Description: fmt.Sprintf(
					"Does not match drpolicy %q interval %s",
					"dr-policy-5m", 5*stdtime.Minute,
				),
			},
			Value: stdtime.Minute,
		}
		validated := cmd.validatedVRGSchedulingInterval(vrg, drpcSummary)
		if validated != expected {
			t.Fatalf("unexpected result\n%s", helpers.UnifiedDiff(t, expected, validated))
		}
	})

	t.Run("drpc interval unknown", func(t *testing.T) {
		cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

		vrg := &ramenapi.VolumeReplicationGroup{
			Spec: ramenapi.VolumeReplicationGroupSpec{
				Async: &ramenapi.VRGAsyncSpec{
					SchedulingInterval: "5m",
				},
			},
		}
		drpcSummary := &report.DRPCSummary{
			SchedulingInterval: report.ValidatedDuration{
				Validated: report.Validated{
					State:       report.Problem,
					Description: "failed to read drpolicy",
				},
			},
		}
		expected := report.ValidatedDuration{
			Validated: report.Validated{State: report.OK},
			Value:     5 * stdtime.Minute,
		}
		validated := cmd.validatedVRGSchedulingInterval(vrg, drpcSummary)
		if validated != expected {
			t.Fatalf("unexpected result\n%s", helpers.UnifiedDiff(t, expected, validated))
		}
	})
}

func TestValidateLastGroupSyncTime(t *testing.T) {
	cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)
	clusterTime := stdtime.Date(2025, 7, 29, 17, 24, 30, 0, stdtime.UTC)
	schedulingInterval := report.ValidatedDuration{
		Validated: report.Validated{State: report.OK},
		Value:     stdtime.Minute,
	}

	t.Run("nil on primary", func(t *testing.T) {
		expected := report.ValidatedTime{
			Validated: report.Validated{
				State:       report.Stale,
				Description: "Waiting for first volume synchronization",
			},
		}
		validated := cmd.validateLastGroupSyncTime(nil, &clusterTime, schedulingInterval, true)
		if !validated.Equal(&expected) {
			t.Fatalf("unexpected result\n%s", helpers.UnifiedDiff(t, expected, validated))
		}
	})

	t.Run("nil on secondary", func(t *testing.T) {
		expected := report.ValidatedTime{
			Validated: report.Validated{State: report.OK},
		}
		validated := cmd.validateLastGroupSyncTime(nil, &clusterTime, schedulingInterval, false)
		if !validated.Equal(&expected) {
			t.Fatalf("unexpected result\n%s", helpers.UnifiedDiff(t, expected, validated))
		}
	})

	t.Run("no cluster time", func(t *testing.T) {
		syncTime := metav1.NewTime(clusterTime.Add(-5 * stdtime.Minute))
		expected := report.ValidatedTime{Value: &syncTime.Time}
		validated := cmd.validateLastGroupSyncTime(&syncTime, nil, schedulingInterval, true)
		if !validated.Equal(&expected) {
			t.Fatalf("unexpected result\n%s", helpers.UnifiedDiff(t, expected, validated))
		}
	})

	t.Run("no scheduling interval", func(t *testing.T) {
		syncTime := metav1.NewTime(clusterTime.Add(-5 * stdtime.Minute))
		missingInterval := report.ValidatedDuration{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Missing scheduling interval",
			},
		}
		expected := report.ValidatedTime{Value: &syncTime.Time}
		validated := cmd.validateLastGroupSyncTime(&syncTime, &clusterTime, missingInterval, true)
		if !validated.Equal(&expected) {
			t.Fatalf("unexpected result\n%s", helpers.UnifiedDiff(t, expected, validated))
		}
	})

	t.Run("ok", func(t *testing.T) {
		maxOK := 2 * stdtime.Minute
		// Simulate metav1.Time.UnmarshalJSON which converts to local time.
		syncTime := metav1.NewTime(clusterTime.Add(-maxOK).Local())
		expected := report.ValidatedTime{
			Validated: report.Validated{State: report.OK},
			Value:     &syncTime.Time,
		}
		validated := cmd.validateLastGroupSyncTime(
			&syncTime, &clusterTime, schedulingInterval, true)
		if !validated.Equal(&expected) {
			t.Fatalf("unexpected result\n%s", helpers.UnifiedDiff(t, expected, validated))
		}
		if validated.Value.Location() != stdtime.UTC {
			t.Fatalf("expected UTC time, got %v", validated.Value)
		}
	})

	t.Run("stale", func(t *testing.T) {
		maxStale := 2*stdtime.Minute + 59*stdtime.Second
		syncTime := metav1.NewTime(clusterTime.Add(-maxStale))
		expected := report.ValidatedTime{
			Validated: report.Validated{
				State:       report.Stale,
				Description: "Replication is exceeding 2x the scheduling interval",
			},
			Value: &syncTime.Time,
		}
		validated := cmd.validateLastGroupSyncTime(
			&syncTime, &clusterTime, schedulingInterval, true)
		if !validated.Equal(&expected) {
			t.Fatalf("unexpected result\n%s", helpers.UnifiedDiff(t, expected, validated))
		}
	})

	t.Run("problem", func(t *testing.T) {
		minProblem := 3 * stdtime.Minute
		syncTime := metav1.NewTime(clusterTime.Add(-minProblem))
		expected := report.ValidatedTime{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Replication is exceeding 3x the scheduling interval",
			},
			Value: &syncTime.Time,
		}
		validated := cmd.validateLastGroupSyncTime(
			&syncTime, &clusterTime, schedulingInterval, true)
		if !validated.Equal(&expected) {
			t.Fatalf("unexpected result\n%s", helpers.UnifiedDiff(t, expected, validated))
		}
	})
}

// writeDRPolicy writes a DRPolicy to the gathered hub data directory.
func writeDRPolicy(t *testing.T, dataDir, name, schedulingInterval string) {
	t.Helper()

	drPolicy := &ramenapi.DRPolicy{
		TypeMeta: metav1.TypeMeta{
			APIVersion: ramenapi.GroupVersion.String(),
			Kind:       "DRPolicy",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: ramenapi.DRPolicySpec{
			SchedulingInterval: schedulingInterval,
		},
	}

	data, err := yaml.Marshal(drPolicy)
	if err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(dataDir, "hub", "cluster", "ramendr.openshift.io", "drpolicies")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, name+".yaml"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}
