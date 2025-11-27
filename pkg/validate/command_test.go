// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	ramenapi "github.com/ramendr/ramen/api/v1alpha1"
	e2econfig "github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/ramen"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/sets"
	"github.com/ramendr/ramenctl/pkg/time"
	"github.com/ramendr/ramenctl/pkg/validation"
)

const (
	validateClusters     = "validate-clusters"
	validateApplication  = "validate-application"
	drpcName             = "appset-deploy-rbd"
	drpcNamespace        = "argocd"
	applicationNamespace = "e2e-appset-deploy-rbd"

	// validateDeleted descriptions.
	resourceDoesNotExist = "Resource does not exist"
	resourceWasDeleted   = "Resource was deleted"
)

// testSystem is a test system such as drenv or ocp clusters.
type testSystem struct {
	name                       string
	config                     *config.Config
	env                        *types.Env
	validateClustersNamespaces []string
}

var (
	testK8s = testSystem{
		name: "k8s",
		config: &config.Config{
			Namespaces: e2econfig.K8sNamespaces,
		},
		env: &types.Env{
			Hub: &types.Cluster{Name: "hub"},
			C1:  &types.Cluster{Name: "dr1"},
			C2:  &types.Cluster{Name: "dr2"},
		},
		validateClustersNamespaces: sets.Sorted([]string{
			e2econfig.K8sNamespaces.RamenHubNamespace,
			e2econfig.K8sNamespaces.RamenDRClusterNamespace,
		}),
	}

	testOcp = testSystem{
		name: "ocp",
		config: &config.Config{
			Namespaces: e2econfig.OcpNamespaces,
		},
		env: &types.Env{
			Hub: &types.Cluster{Name: "hub"},
			C1:  &types.Cluster{Name: "prsurve-s2-c1"},
			C2:  &types.Cluster{Name: "prsurve-s2-c2"},
		},
		validateClustersNamespaces: sets.Sorted([]string{
			e2econfig.OcpNamespaces.RamenHubNamespace,
			e2econfig.OcpNamespaces.RamenDRClusterNamespace,
		}),
	}

	testApplication = &report.Application{
		Name:      drpcName,
		Namespace: drpcNamespace,
	}

	applicationNamespaces = sets.Sorted([]string{
		drpcNamespace,
		applicationNamespace,
	})

	validateApplicationNamespaces = sets.Sorted([]string{
		drpcNamespace,
		applicationNamespace,
	})

	applicationMock = &validation.Mock{
		ApplicationNamespacesFunc: func(validation.Context, string, string) ([]string, error) {
			return applicationNamespaces, nil
		},
	}

	validateConfigFailed = &validation.Mock{
		ValidateFunc: func(ctx validation.Context) error {
			return errors.New("No validate for you!")
		},
	}

	validateConfigCanceled = &validation.Mock{
		ValidateFunc: func(ctx validation.Context) error {
			return context.Canceled
		},
	}

	inspectApplicationFailed = &validation.Mock{
		ApplicationNamespacesFunc: func(validation.Context, string, string) ([]string, error) {
			return nil, errors.New("No namespaces for you!")
		},
	}

	inspectApplicationCanceled = &validation.Mock{
		ApplicationNamespacesFunc: func(validation.Context, string, string) ([]string, error) {
			return nil, context.Canceled
		},
	}

	gatherClusterFailed = &validation.Mock{
		ApplicationNamespacesFunc: applicationMock.ApplicationNamespaces,
		GatherFunc: func(
			ctx validation.Context,
			clusters []*types.Cluster,
			options gathering.Options,
		) <-chan gathering.Result {
			results := make(chan gathering.Result, 3)
			for _, cluster := range clusters {
				if cluster.Name == "hub" {
					results <- gathering.Result{Name: cluster.Name, Err: errors.New("no data for you!")}
				} else {
					results <- gathering.Result{Name: cluster.Name}
				}
			}
			close(results)
			return results
		},
	}
)

// Command common tests

func TestValidatedDeleted(t *testing.T) {
	cmd := testCommand(t, validateApplication, &validation.Mock{}, testK8s)

	t.Run("nil", func(t *testing.T) {
		validated := cmd.validatedDeleted(nil)
		expected := report.ValidatedBool{
			Value: true,
			Validated: report.Validated{
				State:       report.Problem,
				Description: resourceDoesNotExist,
			},
		}
		if validated != expected {
			t.Fatalf("expected %v, got %v", expected, validated)
		}
	})
	t.Run("object deleted", func(t *testing.T) {
		deletedPVC := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				DeletionTimestamp: &metav1.Time{Time: time.Now()},
			},
		}
		validated := cmd.validatedDeleted(deletedPVC)
		expected := report.ValidatedBool{
			Value: true,
			Validated: report.Validated{
				State:       report.Problem,
				Description: resourceWasDeleted,
			},
		}
		if validated != expected {
			t.Fatalf("expected %v, got %v", expected, validated)
		}
	})
	t.Run("object not deleted", func(t *testing.T) {
		pvc := &corev1.PersistentVolumeClaim{}
		validated := cmd.validatedDeleted(pvc)
		expected := report.ValidatedBool{
			Validated: report.Validated{
				State: report.OK,
			},
		}
		if validated != expected {
			t.Fatalf("expected %v, got %v", expected, validated)
		}
	})

	t.Run("update summary", func(t *testing.T) {
		expected := Summary{OK: 1, Problem: 2}
		if cmd.report.Summary != expected {
			t.Fatalf("expected summary %q, got %q", expected, cmd.report.Summary)
		}
	})
}

// Command application tests

func TestValidatedDRPCAction(t *testing.T) {
	cmd := testCommand(t, validateApplication, &validation.Mock{}, testK8s)
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
		expected := Summary{OK: 3, Problem: 1}
		if cmd.report.Summary != expected {
			t.Fatalf("expected summary %q, got %q", expected, cmd.report.Summary)
		}
	})
}

func TestValidatedDRPCPhaseError(t *testing.T) {
	type testcase struct {
		name   string
		action ramenapi.DRAction
		phase  ramenapi.DRState
	}

	cmd := testCommand(t, validateApplication, &validation.Mock{}, testK8s)

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

	var errors uint
	for _, group := range unstable {
		errors += uint(len(group.cases))
	}
	expected := Summary{Problem: errors}
	if cmd.report.Summary != expected {
		t.Fatalf("expected summary %q, got %q", expected, cmd.report.Summary)
	}
}

func TestValidatedDRPCPhaseOK(t *testing.T) {
	cmd := testCommand(t, validateApplication, &validation.Mock{}, testK8s)

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

	expected := Summary{OK: uint(len(cases))}
	if cmd.report.Summary != expected {
		t.Fatalf("expected summary %q, got %q", expected, cmd.report.Summary)
	}
}

func TestValidatedDRPCProgressionOK(t *testing.T) {
	cmd := testCommand(t, validateApplication, &validation.Mock{}, testK8s)
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

	expected := Summary{OK: 1}
	if cmd.report.Summary != expected {
		t.Fatalf("expected summary %q, got %q", expected, cmd.report.Summary)
	}
}

func TestValidatedDRPCProgressionError(t *testing.T) {
	cmd := testCommand(t, validateApplication, &validation.Mock{}, testK8s)

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

	expected := Summary{Problem: uint(len(progressions))}
	if cmd.report.Summary != expected {
		t.Fatalf("expected summary %q, got %q", expected, cmd.report.Summary)
	}
}

func TestValidatedVRGSTateOK(t *testing.T) {
	cmd := testCommand(t, validateApplication, &validation.Mock{}, testK8s)

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

	expected := Summary{OK: uint(len(cases))}
	if cmd.report.Summary != expected {
		t.Fatalf("expected summary %q, got %q", expected, cmd.report.Summary)
	}
}

func TestValidatedVRGSTateError(t *testing.T) {
	cmd := testCommand(t, validateApplication, &validation.Mock{}, testK8s)

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

	expected := Summary{Problem: uint(len(cases))}
	if cmd.report.Summary != expected {
		t.Fatalf("expected summary %q, got %q", expected, cmd.report.Summary)
	}
}

func TestValidatedProtectedPVCOK(t *testing.T) {
	cmd := testCommand(t, validateApplication, &validation.Mock{}, testK8s)

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

	expected := Summary{OK: 1}
	if cmd.report.Summary != expected {
		t.Fatalf("expected summary %q, got %q", expected, cmd.report.Summary)
	}
}

func TestValidatedProtectedPVCError(t *testing.T) {
	cmd := testCommand(t, validateApplication, &validation.Mock{}, testK8s)

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

	expected := Summary{Problem: uint(len(cases))}
	if cmd.report.Summary != expected {
		t.Fatalf("expected summary %q, got %q", expected, cmd.report.Summary)
	}
}

// Validate clusters tests.

func TestValidateClustersK8s(t *testing.T) {
	validate := testCommand(t, validateClusters, &validation.Mock{}, testK8s)
	addGatheredData(t, validate, "clusters/"+testK8s.name)
	if err := validate.Clusters(); err != nil {
		dumpCommandLog(t, validate)
		t.Fatal(err)
	}
	checkReport(t, validate, report.Passed)
	checkApplication(t, validate.report, nil)
	checkNamespaces(t, validate.report, testK8s.validateClustersNamespaces)
	if len(validate.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.report.Steps[1], "validate clusters", report.Passed)

	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "validate clusters data", Status: report.Passed},
	}
	checkItems(t, validate.report.Steps[1], items)
	checkApplicationStatus(t, validate.report, nil)

	expected := &report.ClustersStatus{
		Hub: report.ClustersStatusHub{
			DRClusters: report.ValidatedDRClustersList{
				Validated: report.Validated{
					State: report.OK,
				},
				Value: []report.DRClusterSummary{
					{
						Name:  "dr1",
						Phase: "Available",
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Fenced",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Clean",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Validated",
							},
						},
					},
					{
						Name:  "dr2",
						Phase: "Available",
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Fenced",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Clean",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Validated",
							},
						},
					},
				},
			},
			DRPolicies: report.ValidatedDRPoliciesList{
				Validated: report.Validated{
					State: report.OK,
				},
				Value: []report.DRPolicySummary{
					{
						Name:               "dr-policy-1m",
						DRClusters:         []string{"dr1", "dr2"},
						SchedulingInterval: "1m",
						PeerClasses: report.ValidatedPeerClassesList{
							Validated: report.Validated{
								State: report.OK,
							},
							// TODO: https://github.com/RamenDR/ramenctl/issues/329
							Value: []report.PeerClassesSummary{
								{
									StorageClassName: "rook-ceph-block",
									ReplicationID:    "rook-ceph-replication-1",
								},
								{
									StorageClassName: "rook-cephfs-fs1",
								},
							},
						},
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Validated",
							},
						},
					},
					{
						Name:               "dr-policy-5m",
						DRClusters:         []string{"dr1", "dr2"},
						SchedulingInterval: "5m",
						PeerClasses: report.ValidatedPeerClassesList{
							Validated: report.Validated{
								State: report.OK,
							},
							// TODO: https://github.com/RamenDR/ramenctl/issues/329
							Value: []report.PeerClassesSummary{
								{
									StorageClassName: "rook-ceph-block",
									ReplicationID:    "rook-ceph-replication-1",
								},
								{
									StorageClassName: "rook-cephfs-fs1",
								},
							},
						},
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Validated",
							},
						},
					},
				},
			},
			Ramen: report.RamenSummary{
				ConfigMap: report.ConfigMapSummary{
					Name:      ramen.HubOperatorConfigMapName,
					Namespace: testK8s.config.Namespaces.RamenHubNamespace,
					Deleted: report.ValidatedBool{
						Validated: report.Validated{
							State: report.OK,
						},
					},
					RamenControllerType: report.ValidatedString{
						Validated: report.Validated{
							State: report.OK,
						},
						Value: string(ramenapi.DRHubType),
					},
					S3StoreProfiles: report.ValidatedS3StoreProfilesList{
						Validated: report.Validated{
							State: report.OK,
						},
						Value: []report.S3StoreProfilesSummary{
							{
								S3ProfileName: "minio-on-dr1",
								S3SecretRef: report.ValidatedS3SecretRef{
									Validated: report.Validated{
										State: report.OK,
									},
									Value: corev1.SecretReference{
										Name:      "ramen-s3-secret-dr1",
										Namespace: testK8s.config.Namespaces.RamenHubNamespace,
									},
								},
							},
							{
								S3ProfileName: "minio-on-dr2",
								S3SecretRef: report.ValidatedS3SecretRef{
									Validated: report.Validated{
										State: report.OK,
									},
									Value: corev1.SecretReference{
										Name:      "ramen-s3-secret-dr2",
										Namespace: testK8s.config.Namespaces.RamenHubNamespace,
									},
								},
							},
						},
					},
				},
				Deployment: report.DeploymentSummary{
					Name:      ramen.HubOperatorName,
					Namespace: testK8s.config.Namespaces.RamenHubNamespace,
					Deleted: report.ValidatedBool{
						Validated: report.Validated{
							State: report.OK,
						},
					},
					Replicas: report.ValidatedInteger{
						Validated: report.Validated{
							State: report.OK,
						},
						Value: 1,
					},
					Conditions: []report.ValidatedCondition{
						{
							Validated: report.Validated{
								State: report.OK,
							},
							Type: "Available",
						},
						{
							Validated: report.Validated{
								State: report.OK,
							},
							Type: "Progressing",
						},
					},
				},
			},
		},
		Clusters: []report.ClustersStatusCluster{
			{
				Name: "dr1",
				Ramen: report.RamenSummary{
					ConfigMap: report.ConfigMapSummary{
						Name:      ramen.DrClusterOperatorConfigMapName,
						Namespace: testK8s.config.Namespaces.RamenDRClusterNamespace,
						Deleted: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
						},
						RamenControllerType: report.ValidatedString{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: string(ramenapi.DRClusterType),
						},
						S3StoreProfiles: report.ValidatedS3StoreProfilesList{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: []report.S3StoreProfilesSummary{
								{
									S3ProfileName: "minio-on-dr1",
									S3SecretRef: report.ValidatedS3SecretRef{
										Validated: report.Validated{
											State: report.OK,
										},
										Value: corev1.SecretReference{
											Name:      "ramen-s3-secret-dr1",
											Namespace: testK8s.config.Namespaces.RamenHubNamespace,
										},
									},
								},
								{
									S3ProfileName: "minio-on-dr2",
									S3SecretRef: report.ValidatedS3SecretRef{
										Validated: report.Validated{
											State: report.OK,
										},
										Value: corev1.SecretReference{
											Name:      "ramen-s3-secret-dr2",
											Namespace: testK8s.config.Namespaces.RamenHubNamespace,
										},
									},
								},
							},
						},
					},
					Deployment: report.DeploymentSummary{
						Name:      ramen.DRClusterOperatorName,
						Namespace: testK8s.config.Namespaces.RamenDRClusterNamespace,
						Deleted: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
						},
						Replicas: report.ValidatedInteger{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: 1,
						},
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Available",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Progressing",
							},
						},
					},
				},
			},
			{
				Name: "dr2",
				Ramen: report.RamenSummary{
					ConfigMap: report.ConfigMapSummary{
						Name:      ramen.DrClusterOperatorConfigMapName,
						Namespace: testK8s.config.Namespaces.RamenDRClusterNamespace,
						Deleted: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
						},
						RamenControllerType: report.ValidatedString{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: string(ramenapi.DRClusterType),
						},
						S3StoreProfiles: report.ValidatedS3StoreProfilesList{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: []report.S3StoreProfilesSummary{
								{
									S3ProfileName: "minio-on-dr1",
									S3SecretRef: report.ValidatedS3SecretRef{
										Validated: report.Validated{
											State: report.OK,
										},
										Value: corev1.SecretReference{
											Name:      "ramen-s3-secret-dr1",
											Namespace: testK8s.config.Namespaces.RamenHubNamespace,
										},
									},
								},
								{
									S3ProfileName: "minio-on-dr2",
									S3SecretRef: report.ValidatedS3SecretRef{
										Validated: report.Validated{
											State: report.OK,
										},
										Value: corev1.SecretReference{
											Name:      "ramen-s3-secret-dr2",
											Namespace: testK8s.config.Namespaces.RamenHubNamespace,
										},
									},
								},
							},
						},
					},
					Deployment: report.DeploymentSummary{
						Name:      ramen.DRClusterOperatorName,
						Namespace: testK8s.config.Namespaces.RamenDRClusterNamespace,
						Deleted: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
						},
						Replicas: report.ValidatedInteger{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: 1,
						},
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Available",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Progressing",
							},
						},
					},
				},
			},
		},
	}
	checkClusterStatus(t, validate.report, expected)

	checkSummary(t, validate.report, Summary{OK: 39})
}

func TestValidateClustersOcp(t *testing.T) {
	validate := testCommand(t, validateClusters, &validation.Mock{}, testOcp)
	addGatheredData(t, validate, "clusters/"+testOcp.name)
	if err := validate.Clusters(); err != nil {
		dumpCommandLog(t, validate)
		t.Fatal(err)
	}
	checkReport(t, validate, report.Passed)
	checkApplication(t, validate.report, nil)
	checkNamespaces(t, validate.report, testOcp.validateClustersNamespaces)
	if len(validate.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.report.Steps[1], "validate clusters", report.Passed)

	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"prsurve-s2-c1\"", Status: report.Passed},
		{Name: "gather \"prsurve-s2-c2\"", Status: report.Passed},
		{Name: "validate clusters data", Status: report.Passed},
	}
	checkItems(t, validate.report.Steps[1], items)
	checkApplicationStatus(t, validate.report, nil)

	expected := &report.ClustersStatus{
		Hub: report.ClustersStatusHub{
			DRClusters: report.ValidatedDRClustersList{
				Validated: report.Validated{
					State: report.OK,
				},
				Value: []report.DRClusterSummary{
					{
						Name:  "prsurve-s2-c1",
						Phase: "Available",
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Fenced",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Clean",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Validated",
							},
						},
					},
					{
						Name:  "prsurve-s2-c2",
						Phase: "Available",
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Fenced",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Clean",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Validated",
							},
						},
					},
				},
			},
			DRPolicies: report.ValidatedDRPoliciesList{
				Validated: report.Validated{
					State: report.OK,
				},
				Value: []report.DRPolicySummary{
					{
						Name:               "odr-policy-5m",
						DRClusters:         []string{"prsurve-s2-c1", "prsurve-s2-c2"},
						SchedulingInterval: "5m",
						PeerClasses: report.ValidatedPeerClassesList{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: []report.PeerClassesSummary{
								{
									StorageClassName: "ocs-storagecluster-ceph-rbd",
									ReplicationID:    "275fb2e9822a88bfbfb96516fd307ff3",
									Grouping:         true,
								},
								{
									StorageClassName: "ocs-storagecluster-ceph-rbd-virtualization",
									ReplicationID:    "275fb2e9822a88bfbfb96516fd307ff3",
									Grouping:         true,
								},
								{
									StorageClassName: "ocs-storagecluster-cephfs",
									Grouping:         true,
								},
							},
						},
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Validated",
							},
						},
					},
				},
			},
			Ramen: report.RamenSummary{
				ConfigMap: report.ConfigMapSummary{
					Name:      ramen.HubOperatorConfigMapName,
					Namespace: testOcp.config.Namespaces.RamenHubNamespace,
					Deleted: report.ValidatedBool{
						Validated: report.Validated{
							State: report.OK,
						},
					},
					RamenControllerType: report.ValidatedString{
						Validated: report.Validated{
							State: report.OK,
						},
						Value: string(ramenapi.DRHubType),
					},
					S3StoreProfiles: report.ValidatedS3StoreProfilesList{
						Validated: report.Validated{
							State: report.OK,
						},
						Value: []report.S3StoreProfilesSummary{
							{
								S3ProfileName: "s3profile-prsurve-s2-c1-ocs-storagecluster",
								S3SecretRef: report.ValidatedS3SecretRef{
									Validated: report.Validated{
										State: report.OK,
									},
									Value: corev1.SecretReference{
										Name: "5e88331f09006ac31169b027235b50fd94458b6",
									},
								},
							},
							{
								S3ProfileName: "s3profile-prsurve-s2-c2-ocs-storagecluster",
								S3SecretRef: report.ValidatedS3SecretRef{
									Validated: report.Validated{
										State: report.OK,
									},
									Value: corev1.SecretReference{
										Name: "020a140310eb1fce63e2087087d9a0bdf972b93",
									},
								},
							},
						},
					},
				},
				Deployment: report.DeploymentSummary{
					Name:      ramen.HubOperatorName,
					Namespace: testOcp.config.Namespaces.RamenHubNamespace,
					Deleted: report.ValidatedBool{
						Validated: report.Validated{
							State: report.OK,
						},
					},
					Replicas: report.ValidatedInteger{
						Validated: report.Validated{
							State: report.OK,
						},
						Value: 1,
					},
					Conditions: []report.ValidatedCondition{
						{
							Validated: report.Validated{
								State: report.OK,
							},
							Type: "Progressing",
						},
						{
							Validated: report.Validated{
								State: report.OK,
							},
							Type: "Available",
						},
					},
				},
			},
		},
		Clusters: []report.ClustersStatusCluster{
			{
				Name: "prsurve-s2-c1",
				Ramen: report.RamenSummary{
					ConfigMap: report.ConfigMapSummary{
						Name:      ramen.DrClusterOperatorConfigMapName,
						Namespace: testOcp.config.Namespaces.RamenDRClusterNamespace,
						Deleted: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
						},
						RamenControllerType: report.ValidatedString{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: string(ramenapi.DRClusterType),
						},
						S3StoreProfiles: report.ValidatedS3StoreProfilesList{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: []report.S3StoreProfilesSummary{
								{
									S3ProfileName: "s3profile-prsurve-s2-c1-ocs-storagecluster",
									S3SecretRef: report.ValidatedS3SecretRef{
										Validated: report.Validated{
											State: report.OK,
										},
										Value: corev1.SecretReference{
											Name: "5e88331f09006ac31169b027235b50fd94458b6",
										},
									},
								},
								{
									S3ProfileName: "s3profile-prsurve-s2-c2-ocs-storagecluster",
									S3SecretRef: report.ValidatedS3SecretRef{
										Validated: report.Validated{
											State: report.OK,
										},
										Value: corev1.SecretReference{
											Name: "020a140310eb1fce63e2087087d9a0bdf972b93",
										},
									},
								},
							},
						},
					},
					Deployment: report.DeploymentSummary{
						Name:      ramen.DRClusterOperatorName,
						Namespace: testOcp.config.Namespaces.RamenDRClusterNamespace,
						Deleted: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
						},
						Replicas: report.ValidatedInteger{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: 1,
						},
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Available",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Progressing",
							},
						},
					},
				},
			},
			{
				Name: "prsurve-s2-c2",
				Ramen: report.RamenSummary{
					ConfigMap: report.ConfigMapSummary{
						Name:      ramen.DrClusterOperatorConfigMapName,
						Namespace: testOcp.config.Namespaces.RamenDRClusterNamespace,
						Deleted: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
						},
						RamenControllerType: report.ValidatedString{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: string(ramenapi.DRClusterType),
						},
						S3StoreProfiles: report.ValidatedS3StoreProfilesList{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: []report.S3StoreProfilesSummary{
								{
									S3ProfileName: "s3profile-prsurve-s2-c1-ocs-storagecluster",
									S3SecretRef: report.ValidatedS3SecretRef{
										Validated: report.Validated{
											State: report.OK,
										},
										Value: corev1.SecretReference{
											Name: "5e88331f09006ac31169b027235b50fd94458b6",
										},
									},
								},
								{
									S3ProfileName: "s3profile-prsurve-s2-c2-ocs-storagecluster",
									S3SecretRef: report.ValidatedS3SecretRef{
										Validated: report.Validated{
											State: report.OK,
										},
										Value: corev1.SecretReference{
											Name: "020a140310eb1fce63e2087087d9a0bdf972b93",
										},
									},
								},
							},
						},
					},
					Deployment: report.DeploymentSummary{
						Name:      ramen.DRClusterOperatorName,
						Namespace: testOcp.config.Namespaces.RamenDRClusterNamespace,
						Deleted: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
						},
						Replicas: report.ValidatedInteger{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: 1,
						},
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Available",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "Progressing",
							},
						},
					},
				},
			},
		},
	}
	checkClusterStatus(t, validate.report, expected)

	checkSummary(t, validate.report, Summary{OK: 37})
}

func TestValidateClustersValidateFailed(t *testing.T) {
	validate := testCommand(t, validateClusters, validateConfigFailed, testK8s)
	if err := validate.Clusters(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkApplication(t, validate.report, nil)
	checkNamespaces(t, validate.report, nil)
	if len(validate.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Failed)
	checkApplicationStatus(t, validate.report, nil)
	checkClusterStatus(t, validate.report, nil)
	checkSummary(t, validate.report, Summary{})
}

func TestValidateClustersValidateCanceled(t *testing.T) {
	validate := testCommand(t, validateClusters, validateConfigCanceled, testK8s)
	if err := validate.Clusters(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Canceled)
	checkApplication(t, validate.report, nil)
	checkNamespaces(t, validate.report, nil)
	if len(validate.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Canceled)
	checkApplicationStatus(t, validate.report, nil)
	checkClusterStatus(t, validate.report, nil)
	checkSummary(t, validate.report, Summary{})
}

func TestValidateClusterGatherClusterFailed(t *testing.T) {
	validate := testCommand(t, validateClusters, gatherClusterFailed, testK8s)
	if err := validate.Clusters(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkApplication(t, validate.report, nil)
	checkNamespaces(t, validate.report, testK8s.validateClustersNamespaces)
	if len(validate.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.report.Steps[1], "validate clusters", report.Failed)

	// If gathering data fail for some of the clusters, we skip the validation step.
	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Failed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
	}
	checkItems(t, validate.report.Steps[1], items)
	checkApplicationStatus(t, validate.report, nil)
	checkClusterStatus(t, validate.report, nil)
	checkSummary(t, validate.report, Summary{})
}

// Validate application tests.

func TestValidateApplicationPassed(t *testing.T) {
	validate := testCommand(t, validateApplication, applicationMock, testK8s)
	addGatheredData(t, validate, "appset-deploy-rbd")
	if err := validate.Application(drpcName, drpcNamespace); err != nil {
		dumpCommandLog(t, validate)
		t.Fatal(err)
	}
	checkReport(t, validate, report.Passed)
	checkApplication(t, validate.report, testApplication)
	checkNamespaces(t, validate.report, validateApplicationNamespaces)
	if len(validate.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.report.Steps[1], "validate application", report.Passed)

	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "validate data", Status: report.Passed},
	}
	checkItems(t, validate.report.Steps[1], items)

	expectedStatus := &report.ApplicationStatus{
		Hub: report.ApplicationStatusHub{
			DRPC: report.DRPCSummary{
				Name:      drpcName,
				Namespace: drpcNamespace,
				Deleted: report.ValidatedBool{
					Validated: report.Validated{
						State: report.OK,
					},
				},
				Action: report.ValidatedString{
					Validated: report.Validated{
						State: report.OK,
					},
				},
				DRPolicy: "dr-policy",
				Phase: report.ValidatedString{
					Validated: report.Validated{
						State: report.OK,
					},
					Value: string(ramenapi.Deployed),
				},
				Progression: report.ValidatedString{
					Validated: report.Validated{
						State: report.OK,
					},
					Value: string(ramenapi.ProgressionCompleted),
				},
				Conditions: []report.ValidatedCondition{
					{
						Validated: report.Validated{
							State: report.OK,
						},
						Type: "Available",
					},
					{
						Validated: report.Validated{
							State: report.OK,
						},
						Type: "PeerReady",
					},
					{
						Validated: report.Validated{
							State: report.OK,
						},
						Type: "Protected",
					},
				},
			},
		},
		PrimaryCluster: report.ApplicationStatusCluster{
			Name: "dr1",
			VRG: report.VRGSummary{
				Name:      drpcName,
				Namespace: applicationNamespace,
				Deleted: report.ValidatedBool{
					Validated: report.Validated{
						State: report.OK,
					},
				},
				State: report.ValidatedString{
					Validated: report.Validated{
						State: report.OK,
					},
					Value: string(ramenapi.PrimaryState),
				},
				Conditions: []report.ValidatedCondition{
					{
						Validated: report.Validated{
							State: report.OK,
						},
						Type: "DataReady",
					},
					{
						Validated: report.Validated{
							State: report.OK,
						},
						Type: "ClusterDataReady",
					},
					{
						Validated: report.Validated{
							State: report.OK,
						},
						Type: "ClusterDataProtected",
					},
					{
						Validated: report.Validated{
							State: report.OK,
						},
						Type: "KubeObjectsReady",
					},
					{
						Validated: report.Validated{
							State: report.OK,
						},
						Type: "NoClusterDataConflict",
					},
				},
				ProtectedPVCs: []report.ProtectedPVCSummary{
					{
						Name:      "busybox-pvc",
						Namespace: "e2e-appset-deploy-rbd",
						Deleted: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
						},
						Replication: report.Volrep,
						Phase: report.ValidatedString{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: string(corev1.ClaimBound),
						},
						Conditions: []report.ValidatedCondition{
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "DataReady",
							},
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "ClusterDataProtected",
							},
						},
					},
				},
				// TODO: https://github.com/RamenDR/ramenctl/issues/330
			},
		},
		SecondaryCluster: report.ApplicationStatusCluster{
			Name: "dr2",
			VRG: report.VRGSummary{
				Name:      drpcName,
				Namespace: applicationNamespace,
				Deleted: report.ValidatedBool{
					Validated: report.Validated{
						State: report.OK,
					},
				},
				State: report.ValidatedString{
					Validated: report.Validated{
						State: report.OK,
					},
					Value: string(ramenapi.SecondaryState),
				},
				Conditions: []report.ValidatedCondition{
					{
						Validated: report.Validated{
							State: report.OK,
						},
						Type: "NoClusterDataConflict",
					},
				},
			},
		},
	}
	checkApplicationStatus(t, validate.report, expectedStatus)

	checkSummary(t, validate.report, Summary{OK: 21})
}

func TestValidateApplicationValidateFailed(t *testing.T) {
	validate := testCommand(t, validateApplication, validateConfigFailed, testK8s)
	if err := validate.Application(drpcName, drpcNamespace); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkApplication(t, validate.report, testApplication)
	checkNamespaces(t, validate.report, nil)
	if len(validate.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Failed)
	checkApplicationStatus(t, validate.report, nil)
	checkSummary(t, validate.report, Summary{})
}

func TestValidateApplicationValidateCanceled(t *testing.T) {
	validate := testCommand(t, validateApplication, validateConfigCanceled, testK8s)
	if err := validate.Application(drpcName, drpcNamespace); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Canceled)
	checkApplication(t, validate.report, testApplication)
	checkNamespaces(t, validate.report, nil)
	if len(validate.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Canceled)
	checkApplicationStatus(t, validate.report, nil)
	checkSummary(t, validate.report, Summary{})
}

func TestValidateApplicationInspectApplicationFailed(t *testing.T) {
	validate := testCommand(t, validateApplication, inspectApplicationFailed, testK8s)
	if err := validate.Application(drpcName, drpcNamespace); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkApplication(t, validate.report, testApplication)
	checkNamespaces(t, validate.report, nil)
	if len(validate.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.report.Steps[1], "validate application", report.Failed)

	// If inspecting the application has failed we skip the gather step.
	items := []*report.Step{
		{Name: "inspect application", Status: report.Failed},
	}
	checkItems(t, validate.report.Steps[1], items)
	checkApplicationStatus(t, validate.report, nil)
	checkSummary(t, validate.report, Summary{})
}

func TestValidateApplicationInspectApplicationCanceled(t *testing.T) {
	validate := testCommand(t, validateApplication, inspectApplicationCanceled, testK8s)
	if err := validate.Application(drpcName, drpcNamespace); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Canceled)
	checkApplication(t, validate.report, testApplication)
	checkNamespaces(t, validate.report, nil)
	if len(validate.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.report.Steps[1], "validate application", report.Canceled)

	// If inspecting the application has been canceled we skip the gather step.
	items := []*report.Step{
		{Name: "inspect application", Status: report.Canceled},
	}
	checkItems(t, validate.report.Steps[1], items)
	checkApplicationStatus(t, validate.report, nil)
	checkSummary(t, validate.report, Summary{})
}

func TestValidateApplicationGatherClusterFailed(t *testing.T) {
	validate := testCommand(t, validateApplication, gatherClusterFailed, testK8s)
	if err := validate.Application(drpcName, drpcNamespace); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkApplication(t, validate.report, testApplication)
	checkNamespaces(t, validate.report, validateApplicationNamespaces)
	if len(validate.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.report.Steps[1], "validate application", report.Failed)

	// If gathering data fail for some of the clusters, we skip the validation step.
	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
		{Name: "gather \"hub\"", Status: report.Failed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
	}
	checkItems(t, validate.report.Steps[1], items)
	checkApplicationStatus(t, validate.report, nil)
	checkSummary(t, validate.report, Summary{})
}

// TODO: Test gather cancellation when kubectl-gahter supports it:
// https://github.com/nirs/kubectl-gather/issues/88

// Helpers.

func testCommand(
	t *testing.T,
	name string,
	backend validation.Validation,
	system testSystem,
) *Command {
	cmd, err := command.ForTest(name, system.env, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cmd.Close()
	})
	return newCommand(cmd, system.config, backend)
}

// addGatheredData adds fake gathered data to the output directory.
func addGatheredData(t *testing.T, cmd *Command, name string) {
	testData := fmt.Sprintf("../testdata/%s/%s.data", name, cmd.report.Name)
	source, err := filepath.Abs(testData)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(source, cmd.dataDir()); err != nil {
		t.Fatal(err)
	}
}

func checkReport(t *testing.T, cmd *Command, status report.Status) {
	if cmd.report.Status != status {
		t.Fatalf("expected status %q, got %q", status, cmd.report.Status)
	}
	if !cmd.report.Config.Equal(cmd.Config()) {
		t.Fatalf("expected config %q, got %q", cmd.Config(), cmd.report.Config)
	}
	duration := totalDuration(cmd.report.Steps)
	if cmd.report.Duration != duration {
		t.Fatalf("expected duration %v, got %v", duration, cmd.report.Duration)
	}
}

func checkApplication(t *testing.T, report *Report, expected *report.Application) {
	if !reflect.DeepEqual(expected, report.Application) {
		diff := helpers.UnifiedDiff(t, expected, report.Application)
		t.Fatalf("applications not equal\n%s", diff)
	}
}

func checkNamespaces(t *testing.T, report *Report, expected []string) {
	if !slices.Equal(report.Namespaces, expected) {
		t.Fatalf("expected namespaces %q, got %q", expected, report.Namespaces)
	}
}

func checkStep(t *testing.T, step *report.Step, name string, status report.Status) {
	if name != step.Name {
		t.Fatalf("expected step %q, got %q", name, step.Name)
	}
	if status != step.Status {
		t.Fatalf("expected status %q, got %q", status, step.Status)
	}
	// We cannot check duration since it may be zero on windows.
}

func checkItems(t *testing.T, step *report.Step, expected []*report.Step) {
	if len(expected) != len(step.Items) {
		t.Fatalf("expected items %+v, got %+v", expected, step.Items)
	}
	for i, item := range expected {
		checkStep(t, step.Items[i], item.Name, item.Status)
	}
}

func checkApplicationStatus(
	t *testing.T,
	report *Report,
	expected *report.ApplicationStatus,
) {
	if expected != nil {
		if !expected.Equal(report.ApplicationStatus) {
			diff := helpers.UnifiedDiff(t, expected, report.ApplicationStatus)
			t.Fatalf("application statuses not equal\n%s", diff)
		}
	} else if report.ApplicationStatus != nil {
		t.Fatalf("application status not nil\n%s",
			helpers.MarshalYAML(t, report.ApplicationStatus))
	}
}

func checkClusterStatus(
	t *testing.T,
	report *Report,
	expected *report.ClustersStatus,
) {
	if expected != nil {
		if !expected.Equal(report.ClustersStatus) {
			diff := helpers.UnifiedDiff(t, expected, report.ClustersStatus)
			t.Fatalf("clusters statuses not equal\n%s", diff)
		}
	} else if report.ClustersStatus != nil {
		t.Fatalf("clusters status not nil\n%s",
			helpers.MarshalYAML(t, report.ClustersStatus))
	}
}

func checkSummary(t *testing.T, report *Report, expected Summary) {
	if report.Summary != expected {
		t.Fatalf("expected summary %q, got %q", expected, report.Summary)
	}
}

func totalDuration(steps []*report.Step) float64 {
	var total float64
	for _, step := range steps {
		total += step.Duration
	}
	return total
}

func dumpCommandLog(t *testing.T, cmd *Command) {
	log, err := os.ReadFile(cmd.command.LogFile())
	if err != nil {
		t.Logf("Failed to read command log: %s", err)
		return
	}
	t.Logf("Command log:\n%s", log)
}
