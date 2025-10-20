// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report_test

import (
	"testing"

	ramenapi "github.com/ramendr/ramen/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
)

func TestReportApplicationStatusEqual(t *testing.T) {
	a1 := testApplicationStatus()
	t.Run("equal to self", func(t *testing.T) {
		a2 := a1
		checkApplicationsEqual(t, a1, a2)
	})
	t.Run("equal applications", func(t *testing.T) {
		a2 := testApplicationStatus()
		checkApplicationsEqual(t, a1, a2)
	})
}

func TestReportApplicationStatusNotEqual(t *testing.T) {
	a1 := testApplicationStatus()
	t.Run("not equal to nil", func(t *testing.T) {
		var a2 *report.ApplicationStatus
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("hub drpc name", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.Hub.DRPC.Name = helpers.Modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("hub drpc namespace", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.Hub.DRPC.Namespace = helpers.Modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("hub drpc deleted", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.Hub.DRPC.Deleted = report.ValidatedBool{
			Value: true,
			Validated: report.Validated{
				State:       report.Problem,
				Description: "DRPC does not exist",
			},
		}
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("hub drpc action", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.Hub.DRPC.Action = report.ValidatedString{
			Validated: report.Validated{
				State: report.OK,
			},
			Value: string(ramenapi.ActionFailover),
		}
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("hub drpc drPolicy", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.Hub.DRPC.DRPolicy = helpers.Modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("hub drpc phase", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.Hub.DRPC.Phase = report.ValidatedString{
			Validated: report.Validated{
				State: report.OK,
			},
			Value: string(ramenapi.FailedOver),
		}
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("hub drpc progression", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.Hub.DRPC.Progression = report.ValidatedString{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Waiting for stable progression",
			},
			Value: string(ramenapi.ProgressionFailingOverToCluster),
		}
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("hub drpc conditions nil", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.Hub.DRPC.Conditions = nil
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("hub drpc conditions", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.Hub.DRPC.Conditions[0].State = report.Problem
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster name", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.PrimaryCluster.Name = helpers.Modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg name", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.PrimaryCluster.VRG.Name = helpers.Modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg namespace", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.PrimaryCluster.VRG.Namespace = helpers.Modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg deleted", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.PrimaryCluster.VRG.Deleted = report.ValidatedBool{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "VRG does not exist",
			},
			Value: true,
		}
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg state", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.PrimaryCluster.VRG.State = report.ValidatedString{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Waiting to become \"Primary\"",
			},
			Value: string(ramenapi.SecondaryState),
		}
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg conditions nil", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.PrimaryCluster.VRG.Conditions = nil
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg conditions", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.PrimaryCluster.VRG.Conditions[0].State = report.Problem
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg protectedpvcs nil", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.PrimaryCluster.VRG.ProtectedPVCs = nil
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg protectedpvcs name", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.PrimaryCluster.VRG.ProtectedPVCs[0].Name = helpers.Modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg protectedpvcs namespace", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.PrimaryCluster.VRG.ProtectedPVCs[0].Namespace = helpers.Modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg protectedpvcs replication", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.PrimaryCluster.VRG.ProtectedPVCs[0].Replication = report.Volsync
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg protectedpvcs deleted", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.PrimaryCluster.VRG.ProtectedPVCs[0].Deleted = report.ValidatedBool{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "PVC does not exist",
			},
			Value: true,
		}
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg protectedpvcs phase", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.PrimaryCluster.VRG.ProtectedPVCs[0].Phase = report.ValidatedString{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "PVC is not \"Bound\"",
			},
			Value: "Terminating",
		}
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg protectedpvcs conditions nil", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.PrimaryCluster.VRG.ProtectedPVCs[0].Conditions = nil
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg protectedpvcs conditions", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.PrimaryCluster.VRG.ProtectedPVCs[0].Conditions[0].State = report.Problem
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg pvcgroups nil", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.PrimaryCluster.VRG.PVCGroups = nil
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg pvcgroups grouped", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.PrimaryCluster.VRG.PVCGroups[0].Grouped = []string{"different-pvc"}
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("secondary cluster name", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.SecondaryCluster.Name = helpers.Modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("secondary cluster vrg name", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.SecondaryCluster.VRG.Name = helpers.Modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("secondary cluster vrg namespace", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.SecondaryCluster.VRG.Namespace = helpers.Modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("secondary cluster vrg deleted", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.SecondaryCluster.VRG.Deleted = report.ValidatedBool{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "VRG does not exist",
			},
			Value: true,
		}
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("secondary cluster vrg state", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.SecondaryCluster.VRG.State = report.ValidatedString{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Waiting to become \"Secondary\"",
			},
			Value: string(ramenapi.PrimaryState),
		}
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("secondary cluster vrg conditions nil", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.SecondaryCluster.VRG.Conditions = nil
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("secondary cluster vrg conditions", func(t *testing.T) {
		a2 := testApplicationStatus()
		a2.SecondaryCluster.VRG.Conditions[0].State = report.Problem
		checkApplicationsNotEqual(t, a1, a2)
	})
}

func TestReportApplicationStatusMarshaling(t *testing.T) {
	a1 := testApplicationStatus()
	data, err := yaml.Marshal(a1)
	if err != nil {
		t.Fatal(err)
	}
	a2 := &report.ApplicationStatus{}
	if err := yaml.Unmarshal(data, a2); err != nil {
		t.Fatal(err)
	}
	checkApplicationsEqual(t, a1, a2)
}

func testApplicationStatus() *report.ApplicationStatus {
	a := &report.ApplicationStatus{
		Hub: report.ApplicationStatusHub{
			DRPC: report.DRPCSummary{
				Name:      "drpc-name",
				Namespace: "drpc-namespace",
				Deleted: report.ValidatedBool{
					Validated: report.Validated{
						State: report.OK,
					},
				},
				DRPolicy: "dr-policy-1m",
				Action: report.ValidatedString{
					Validated: report.Validated{
						State: report.OK,
					},
				},
				Phase: report.ValidatedString{
					Validated: report.Validated{
						State: report.OK,
					},
					Value: "Deployed",
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
				Name:      "vrg-name",
				Namespace: "vrg-namespace",
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
						Type: "DataProtected",
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
						Name:        "pvc-name",
						Namespace:   "app-namespace",
						Replication: report.Volrep,
						Deleted: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
						},
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
							{
								Validated: report.Validated{
									State: report.OK,
								},
								Type: "DataProtected",
							},
						},
					},
				},
				PVCGroups: []report.PVCGroupsSummary{
					{
						Grouped: []string{"pvc-name"},
					},
				},
			},
		},
		SecondaryCluster: report.ApplicationStatusCluster{
			Name: "dr2",
			VRG: report.VRGSummary{
				Name:      "vrg-name",
				Namespace: "vrg-namespace",
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
	return a
}

func checkApplicationsEqual(t *testing.T, a, b *report.ApplicationStatus) {
	if !a.Equal(b) {
		diff := helpers.UnifiedDiff(t, a, b)
		t.Fatalf("application statuses are not equal\n%s", diff)
	}
}

func checkApplicationsNotEqual(t *testing.T, a, b *report.ApplicationStatus) {
	if a.Equal(b) {
		t.Fatalf("application statuses are equal\n%s", helpers.MarshalYAML(t, a))
	}
}
