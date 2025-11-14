// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report_test

import (
	"testing"

	ramenapi "github.com/ramendr/ramen/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/ramen"
	"github.com/ramendr/ramenctl/pkg/report"
)

func TestReportClusterStatusEqual(t *testing.T) {
	c1 := testClusterStatus()

	t.Run("equal to self", func(t *testing.T) {
		c2 := c1
		checkClustersEqual(t, c1, c2)
	})
	t.Run("equal clusters", func(t *testing.T) {
		c2 := testClusterStatus()
		checkClustersEqual(t, c1, c2)
	})
}

func TestReportClusterStatusNotEqual(t *testing.T) {
	c1 := testClusterStatus()

	t.Run("not equal to nil", func(t *testing.T) {
		var c2 *report.ClustersStatus
		checkClustersNotEqual(t, c1, c2)
	})

	// Hub drClusters tests

	t.Run("hub drclusters nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters.Value = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drcluster conditions nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters.Value[0].Conditions = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drclusters state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drcluster name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters.Value[0].Name = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drcluster phase", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters.Value[0].Phase = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drcluster conditions", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters.Value[0].Conditions[0].State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})

	// Hub drPolicies tests

	t.Run("hub drpolicies nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy conditions nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value[0].Conditions = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicies state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value[0].Name = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy drclusters", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value[0].DRClusters[0] = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy scheduling interval", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value[0].SchedulingInterval = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy peer classes state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value[0].PeerClasses.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy peer classes storage class name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value[0].PeerClasses.Value[0].StorageClassName = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy peer classes grouping", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value[0].PeerClasses.Value[0].Grouping = false
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy conditions", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies.Value[0].Conditions[0].State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})

	// Hub ramen configmap tests

	t.Run("hub ramen configmap s3storeprofiles value nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.Name = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap namespace", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.Namespace = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap deleted", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.Deleted = report.ValidatedBool{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Resource does not exist",
			},
			Value: true,
		}
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap ramen controller type state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.RamenControllerType.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap ramen controller type", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.RamenControllerType.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3ProfileName = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles secretref state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap s3storeprofiles secretref value name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.Value.Name = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})

	// Hub ramen deployment tests

	t.Run("hub ramen deployment conditions nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Deployment.Conditions = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen deployment name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Deployment.Name = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen deployment namespace", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Deployment.Namespace = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen deployment deleted", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Deployment.Deleted = report.ValidatedBool{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Resource does not exist",
			},
			Value: true,
		}
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen deployment replicas", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Deployment.Replicas = report.ValidatedInteger{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Expecting 1 replica",
			},
			Value: 0,
		}
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen deployment conditions", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Deployment.Conditions[0].State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})

	// Managed cluster tests

	t.Run("clusters nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles value nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen deployment conditions nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.Deployment.Conditions = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Name = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.Name = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap namespace", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.Namespace = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap ramen controller type state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.RamenControllerType.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap ramen controller type", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.RamenControllerType.Value = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles profile name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3ProfileName = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles secretref state", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap s3storeprofiles secretref value name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.S3StoreProfiles.Value[0].S3SecretRef.Value.Name = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen deployment name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.Deployment.Name = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen deployment namespace", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.Deployment.Namespace = helpers.Modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen deployment replicas", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.Deployment.Replicas = report.ValidatedInteger{
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Expecting 1 replica",
			},
			Value: 2,
		}
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen deployment conditions", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.Deployment.Conditions[0].State = report.Problem
		checkClustersNotEqual(t, c1, c2)
	})
}

func TestReportClusterStatusMarshaling(t *testing.T) {
	c1 := testClusterStatus()
	data, err := yaml.Marshal(c1)
	if err != nil {
		t.Fatal(err)
	}
	c2 := &report.ClustersStatus{}
	if err := yaml.Unmarshal(data, c2); err != nil {
		t.Fatal(err)
	}
	checkClustersEqual(t, c1, c2)
}

func testClusterStatus() *report.ClustersStatus {
	c := &report.ClustersStatus{
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
							Value: []report.PeerClassesSummary{
								{
									StorageClassName: "rook-ceph-block",
									ReplicationID:    "rook-ceph-replication-1",
									Grouping:         true,
								},
								{
									StorageClassName: "rook-cephfs-fs1",
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
					{
						Name:               "dr-policy-5m",
						DRClusters:         []string{"dr1", "dr2"},
						SchedulingInterval: "5m",
						PeerClasses: report.ValidatedPeerClassesList{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: []report.PeerClassesSummary{
								{
									StorageClassName: "rook-ceph-block",
									ReplicationID:    "rook-ceph-replication-1",
									Grouping:         true,
								},
								{
									StorageClassName: "rook-cephfs-fs1",
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
					Namespace: "ramen-system",
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
								S3ProfileName: "s3-profile-dr1",
								S3SecretRef: report.ValidatedS3SecretRef{
									Validated: report.Validated{
										State: report.OK,
									},
									Value: corev1.SecretReference{
										Name:      "ramen-s3-secret-dr1",
										Namespace: "ramen-system",
									},
								},
							},
							{
								S3ProfileName: "s3-profile-dr2",
								S3SecretRef: report.ValidatedS3SecretRef{
									Validated: report.Validated{
										State: report.OK,
									},
									Value: corev1.SecretReference{
										Name:      "ramen-s3-secret-dr2",
										Namespace: "ramen-system",
									},
								},
							},
						},
					},
				},
				Deployment: report.DeploymentSummary{
					Name:      ramen.HubOperatorName,
					Namespace: "ramen-system",
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
						Namespace: "ramen-system",
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
									S3ProfileName: "s3-profile-dr1",
									S3SecretRef: report.ValidatedS3SecretRef{
										Validated: report.Validated{
											State: report.OK,
										},
										Value: corev1.SecretReference{
											Name:      "ramen-s3-secret-dr1",
											Namespace: "ramen-system",
										},
									},
								},
								{
									S3ProfileName: "s3-profile-dr2",
									S3SecretRef: report.ValidatedS3SecretRef{
										Validated: report.Validated{
											State: report.OK,
										},
										Value: corev1.SecretReference{
											Name:      "ramen-s3-secret-dr2",
											Namespace: "ramen-system",
										},
									},
								},
							},
						},
					},
					Deployment: report.DeploymentSummary{
						Name:      ramen.DRClusterOperatorName,
						Namespace: "ramen-system",
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
						Namespace: "ramen-system",
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
									S3ProfileName: "s3-profile-dr1",
									S3SecretRef: report.ValidatedS3SecretRef{
										Validated: report.Validated{
											State: report.OK,
										},
										Value: corev1.SecretReference{
											Name:      "ramen-s3-secret-dr1",
											Namespace: "ramen-system",
										},
									},
								},
								{
									S3ProfileName: "s3-profile-dr2",
									S3SecretRef: report.ValidatedS3SecretRef{
										Validated: report.Validated{
											State: report.OK,
										},
										Value: corev1.SecretReference{
											Name:      "ramen-s3-secret-dr2",
											Namespace: "ramen-system",
										},
									},
								},
							},
						},
					},
					Deployment: report.DeploymentSummary{
						Name:      ramen.DRClusterOperatorName,
						Namespace: "ramen-system",
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
	return c
}

func checkClustersEqual(t *testing.T, a, b *report.ClustersStatus) {
	if !a.Equal(b) {
		diff := helpers.UnifiedDiff(t, a, b)
		t.Fatalf("clusters statuses are not equal\n%s", diff)
	}
}

func checkClustersNotEqual(t *testing.T, a, b *report.ClustersStatus) {
	if a.Equal(b) {
		t.Fatalf("clusters statuses are equal\n%s", helpers.MarshalYAML(t, a))
	}
}
