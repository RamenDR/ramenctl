// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package clusters

import (
	"testing"

	ramenapi "github.com/ramendr/ramen/api/v1alpha1"

	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/ramen"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/validate/summary"
)

const (
	k8sTestdata = "../../testdata/clusters/k8s"
	ocpTestdata = "../../testdata/clusters/ocp"
)

// Clusters mock instances.

var (
	clustersGatherDataFailed = &helpers.ValidationMock{
		GatherFunc: helpers.GatherDataFailed,
	}

	checkS3Failed = &helpers.ValidationMock{
		CheckS3Func: helpers.CheckS3DataFailed,
	}

	checkS3Canceled = &helpers.ValidationMock{
		CheckS3Func: helpers.CheckS3DataCanceled,
	}
)

// Validate clusters tests.

func TestValidateClustersK8s(t *testing.T) {
	validate := testCommand(t, &helpers.ValidationMock{}, testK8s)
	helpers.AddGatheredData(t, validate.DataDir(), k8sTestdata, validate.Report.Name)
	if err := validate.Run(); err != nil {
		dumpCommandLog(t, validate)
		t.Fatal(err)
	}
	checkReport(t, validate, report.Passed)
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, testK8s.namespaces)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.Report.Steps[1], "validate clusters", report.Passed)

	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{Name: "check S3 profile \"minio-on-dr1\"", Status: report.Passed},
		{Name: "check S3 profile \"minio-on-dr2\"", Status: report.Passed},
		{Name: "validate clusters data", Status: report.Passed},
	}
	checkItems(t, validate.Report.Steps[1], items)

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
					S3StoreProfiles: report.ValidatedS3StoreProfilesList{
						Validated: report.Validated{
							State: report.OK,
						},
						Value: []report.S3StoreProfilesSummary{
							{
								S3ProfileName: "minio-on-dr1",
								S3Bucket: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "bucket",
								},
								S3CompatibleEndpoint: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "http://example-cluster:30000",
								},
								S3Region: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "us-west-1",
								},
								S3SecretRef: report.S3SecretSummary{
									Name: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "ramen-s3-secret-dr1",
									},
									Namespace: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     testK8s.config.Namespaces.RamenHubNamespace,
									},
									Deleted: report.ValidatedBool{
										Validated: report.Validated{State: report.OK},
									},
									AWSAccessKeyID: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
										Value:     helpers.FakeAWSKeyIDFingerprint,
									},
									AWSSecretAccessKey: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
										Value:     helpers.FakeAWSKeyFingerprint,
									},
								},
								// CACertificate is optional, empty is OK if hub also has no cert.
								CACertificate: report.ValidatedFingerprint{
									Validated: report.Validated{State: report.OK},
								},
							},
							{
								S3ProfileName: "minio-on-dr2",
								S3Bucket: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "bucket",
								},
								S3CompatibleEndpoint: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "http://example-cluster:30000",
								},
								S3Region: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "us-east-1",
								},
								S3SecretRef: report.S3SecretSummary{
									Name: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "ramen-s3-secret-dr2",
									},
									Namespace: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     testK8s.config.Namespaces.RamenHubNamespace,
									},
									Deleted: report.ValidatedBool{
										Validated: report.Validated{State: report.OK},
									},
									AWSAccessKeyID: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
										Value:     helpers.FakeAWSKeyIDFingerprint,
									},
									AWSSecretAccessKey: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
										Value:     helpers.FakeAWSKeyFingerprint,
									},
								},
								CACertificate: report.ValidatedFingerprint{
									Validated: report.Validated{State: report.OK},
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
					RamenControllerType: report.ValidatedString{
						Validated: report.Validated{
							State: report.OK,
						},
						Value: string(ramenapi.DRHubType),
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
						S3StoreProfiles: report.ValidatedS3StoreProfilesList{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: []report.S3StoreProfilesSummary{
								{
									S3ProfileName: "minio-on-dr1",
									S3Bucket: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "bucket",
									},
									S3CompatibleEndpoint: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "http://example-cluster:30000",
									},
									S3Region: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "us-west-1",
									},
									S3SecretRef: report.S3SecretSummary{
										Name: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     "ramen-s3-secret-dr1",
										},
										Namespace: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     testK8s.config.Namespaces.RamenHubNamespace,
										},
										Deleted: report.ValidatedBool{
											Validated: report.Validated{State: report.OK},
										},
										AWSAccessKeyID: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyIDFingerprint,
										},
										AWSSecretAccessKey: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyFingerprint,
										},
									},
									CACertificate: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
									},
								},
								{
									S3ProfileName: "minio-on-dr2",
									S3Bucket: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "bucket",
									},
									S3CompatibleEndpoint: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "http://example-cluster:30000",
									},
									S3Region: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "us-east-1",
									},
									S3SecretRef: report.S3SecretSummary{
										Name: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     "ramen-s3-secret-dr2",
										},
										Namespace: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     testK8s.config.Namespaces.RamenHubNamespace,
										},
										Deleted: report.ValidatedBool{
											Validated: report.Validated{State: report.OK},
										},
										AWSAccessKeyID: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyIDFingerprint,
										},
										AWSSecretAccessKey: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyFingerprint,
										},
									},
									CACertificate: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
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
						RamenControllerType: report.ValidatedString{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: string(ramenapi.DRClusterType),
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
						S3StoreProfiles: report.ValidatedS3StoreProfilesList{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: []report.S3StoreProfilesSummary{
								{
									S3ProfileName: "minio-on-dr1",
									S3Bucket: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "bucket",
									},
									S3CompatibleEndpoint: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "http://example-cluster:30000",
									},
									S3Region: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "us-west-1",
									},
									S3SecretRef: report.S3SecretSummary{
										Name: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     "ramen-s3-secret-dr1",
										},
										Namespace: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     testK8s.config.Namespaces.RamenHubNamespace,
										},
										Deleted: report.ValidatedBool{
											Validated: report.Validated{State: report.OK},
										},
										AWSAccessKeyID: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyIDFingerprint,
										},
										AWSSecretAccessKey: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyFingerprint,
										},
									},
									CACertificate: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
									},
								},
								{
									S3ProfileName: "minio-on-dr2",
									S3Bucket: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "bucket",
									},
									S3CompatibleEndpoint: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "http://example-cluster:30000",
									},
									S3Region: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "us-east-1",
									},
									S3SecretRef: report.S3SecretSummary{
										Name: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     "ramen-s3-secret-dr2",
										},
										Namespace: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     testK8s.config.Namespaces.RamenHubNamespace,
										},
										Deleted: report.ValidatedBool{
											Validated: report.Validated{State: report.OK},
										},
										AWSAccessKeyID: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyIDFingerprint,
										},
										AWSSecretAccessKey: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyFingerprint,
										},
									},
									CACertificate: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
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
						RamenControllerType: report.ValidatedString{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: string(ramenapi.DRClusterType),
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
		S3: report.ClustersS3Status{
			Profiles: report.ValidatedClustersS3ProfileStatusList{
				Validated: report.Validated{
					State: report.OK,
				},
				Value: []report.ClustersS3ProfileStatus{
					{
						Name: "minio-on-dr1",
						Accessible: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: true,
						},
					},
					{
						Name: "minio-on-dr2",
						Accessible: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: true,
						},
					},
				},
			},
		},
	}
	checkClusterStatus(t, validate.Report, expected)

	checkSummary(t, validate.Report, report.Summary{summary.OK: 90})
}

func TestValidateClustersOcp(t *testing.T) {
	validate := testCommand(t, &helpers.ValidationMock{}, testOcp)
	helpers.AddGatheredData(t, validate.DataDir(), ocpTestdata, validate.Report.Name)
	if err := validate.Run(); err != nil {
		dumpCommandLog(t, validate)
		t.Fatal(err)
	}
	checkReport(t, validate, report.Passed)
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, testOcp.namespaces)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.Report.Steps[1], "validate clusters", report.Passed)

	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"c1\"", Status: report.Passed},
		{Name: "gather \"c2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{
			Name:   "check S3 profile \"s3profile-c1-ocs-storagecluster\"",
			Status: report.Passed,
		},
		{
			Name:   "check S3 profile \"s3profile-c2-ocs-storagecluster\"",
			Status: report.Passed,
		},
		{Name: "validate clusters data", Status: report.Passed},
	}
	checkItems(t, validate.Report.Steps[1], items)

	expected := &report.ClustersStatus{
		Hub: report.ClustersStatusHub{
			DRClusters: report.ValidatedDRClustersList{
				Validated: report.Validated{
					State: report.OK,
				},
				Value: []report.DRClusterSummary{
					{
						Name:  "c1",
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
						Name:  "c2",
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
						DRClusters:         []string{"c1", "c2"},
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
					S3StoreProfiles: report.ValidatedS3StoreProfilesList{
						Validated: report.Validated{
							State: report.OK,
						},
						Value: []report.S3StoreProfilesSummary{
							{
								S3ProfileName: "s3profile-c1-ocs-storagecluster",
								S3Bucket: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "odrbucket-244c8f95bf0d",
								},
								S3CompatibleEndpoint: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "https://s3-openshift-storage.apps.c1.example.com",
								},
								S3Region: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "noobaa",
								},
								S3SecretRef: report.S3SecretSummary{
									Name: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "5e88331f09006ac31169b027235b50fd94458b6",
									},
									Namespace: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
									},
									Deleted: report.ValidatedBool{
										Validated: report.Validated{State: report.OK},
									},
									AWSAccessKeyID: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
										Value:     helpers.FakeAWSKeyIDFingerprint,
									},
									AWSSecretAccessKey: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
										Value:     helpers.FakeAWSKeyFingerprint,
									},
								},
								CACertificate: report.ValidatedFingerprint{
									Validated: report.Validated{State: report.OK},
									Value:     caCertificateFingerprint,
								},
							},
							{
								S3ProfileName: "s3profile-c2-ocs-storagecluster",
								S3Bucket: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "odrbucket-244c8f95bf0d",
								},
								S3CompatibleEndpoint: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "https://s3-openshift-storage.apps.c2.example.com",
								},
								S3Region: report.ValidatedString{
									Validated: report.Validated{State: report.OK},
									Value:     "noobaa",
								},
								S3SecretRef: report.S3SecretSummary{
									Name: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "020a140310eb1fce63e2087087d9a0bdf972b93",
									},
									Namespace: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
									},
									Deleted: report.ValidatedBool{
										Validated: report.Validated{State: report.OK},
									},
									AWSAccessKeyID: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
										Value:     helpers.FakeAWSKeyIDFingerprint,
									},
									AWSSecretAccessKey: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
										Value:     helpers.FakeAWSKeyFingerprint,
									},
								},
								CACertificate: report.ValidatedFingerprint{
									Validated: report.Validated{State: report.OK},
									Value:     caCertificateFingerprint,
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
					RamenControllerType: report.ValidatedString{
						Validated: report.Validated{
							State: report.OK,
						},
						Value: string(ramenapi.DRHubType),
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
				Name: "c1",
				Ramen: report.RamenSummary{
					ConfigMap: report.ConfigMapSummary{
						Name:      ramen.DrClusterOperatorConfigMapName,
						Namespace: testOcp.config.Namespaces.RamenDRClusterNamespace,
						Deleted: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
						},
						S3StoreProfiles: report.ValidatedS3StoreProfilesList{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: []report.S3StoreProfilesSummary{
								{
									S3ProfileName: "s3profile-c1-ocs-storagecluster",
									S3Bucket: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "odrbucket-244c8f95bf0d",
									},
									S3CompatibleEndpoint: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "https://s3-openshift-storage.apps.c1.example.com",
									},
									S3Region: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "noobaa",
									},
									S3SecretRef: report.S3SecretSummary{
										Name: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     "5e88331f09006ac31169b027235b50fd94458b6",
										},
										Namespace: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
										},
										Deleted: report.ValidatedBool{
											Validated: report.Validated{State: report.OK},
										},
										AWSAccessKeyID: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyIDFingerprint,
										},
										AWSSecretAccessKey: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyFingerprint,
										},
									},
									CACertificate: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
										Value:     caCertificateFingerprint,
									},
								},
								{
									S3ProfileName: "s3profile-c2-ocs-storagecluster",
									S3Bucket: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "odrbucket-244c8f95bf0d",
									},
									S3CompatibleEndpoint: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "https://s3-openshift-storage.apps.c2.example.com",
									},
									S3Region: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "noobaa",
									},
									S3SecretRef: report.S3SecretSummary{
										Name: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     "020a140310eb1fce63e2087087d9a0bdf972b93",
										},
										Namespace: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
										},
										Deleted: report.ValidatedBool{
											Validated: report.Validated{State: report.OK},
										},
										AWSAccessKeyID: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyIDFingerprint,
										},
										AWSSecretAccessKey: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyFingerprint,
										},
									},
									CACertificate: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
										Value:     caCertificateFingerprint,
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
						RamenControllerType: report.ValidatedString{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: string(ramenapi.DRClusterType),
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
				Name: "c2",
				Ramen: report.RamenSummary{
					ConfigMap: report.ConfigMapSummary{
						Name:      ramen.DrClusterOperatorConfigMapName,
						Namespace: testOcp.config.Namespaces.RamenDRClusterNamespace,
						Deleted: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
						},
						S3StoreProfiles: report.ValidatedS3StoreProfilesList{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: []report.S3StoreProfilesSummary{
								{
									S3ProfileName: "s3profile-c1-ocs-storagecluster",
									S3Bucket: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "odrbucket-244c8f95bf0d",
									},
									S3CompatibleEndpoint: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "https://s3-openshift-storage.apps.c1.example.com",
									},
									S3Region: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "noobaa",
									},
									S3SecretRef: report.S3SecretSummary{
										Name: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     "5e88331f09006ac31169b027235b50fd94458b6",
										},
										Namespace: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
										},
										Deleted: report.ValidatedBool{
											Validated: report.Validated{State: report.OK},
										},
										AWSAccessKeyID: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyIDFingerprint,
										},
										AWSSecretAccessKey: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyFingerprint,
										},
									},
									CACertificate: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
										Value:     caCertificateFingerprint,
									},
								},
								{
									S3ProfileName: "s3profile-c2-ocs-storagecluster",
									S3Bucket: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "odrbucket-244c8f95bf0d",
									},
									S3CompatibleEndpoint: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "https://s3-openshift-storage.apps.c2.example.com",
									},
									S3Region: report.ValidatedString{
										Validated: report.Validated{State: report.OK},
										Value:     "noobaa",
									},
									S3SecretRef: report.S3SecretSummary{
										Name: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
											Value:     "020a140310eb1fce63e2087087d9a0bdf972b93",
										},
										Namespace: report.ValidatedString{
											Validated: report.Validated{State: report.OK},
										},
										Deleted: report.ValidatedBool{
											Validated: report.Validated{State: report.OK},
										},
										AWSAccessKeyID: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyIDFingerprint,
										},
										AWSSecretAccessKey: report.ValidatedFingerprint{
											Validated: report.Validated{State: report.OK},
											Value:     helpers.FakeAWSKeyFingerprint,
										},
									},
									CACertificate: report.ValidatedFingerprint{
										Validated: report.Validated{State: report.OK},
										Value:     caCertificateFingerprint,
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
						RamenControllerType: report.ValidatedString{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: string(ramenapi.DRClusterType),
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
		S3: report.ClustersS3Status{
			Profiles: report.ValidatedClustersS3ProfileStatusList{
				Validated: report.Validated{
					State: report.OK,
				},
				Value: []report.ClustersS3ProfileStatus{
					{
						Name: "s3profile-c1-ocs-storagecluster",
						Accessible: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: true,
						},
					},
					{
						Name: "s3profile-c2-ocs-storagecluster",
						Accessible: report.ValidatedBool{
							Validated: report.Validated{
								State: report.OK,
							},
							Value: true,
						},
					},
				},
			},
		},
	}
	checkClusterStatus(t, validate.Report, expected)

	checkSummary(t, validate.Report, report.Summary{summary.OK: 88})
}

func TestValidateClustersValidateFailed(t *testing.T) {
	validate := testCommand(t, helpers.ValidateConfigFailed, testK8s)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, nil)
	if len(validate.Report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], "validate config", report.Failed)
	checkClusterStatus(t, validate.Report, &report.ClustersStatus{})
	checkSummary(t, validate.Report, report.Summary{})
}

func TestValidateClustersValidateCanceled(t *testing.T) {
	validate := testCommand(t, helpers.ValidateConfigCanceled, testK8s)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Canceled)
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, nil)
	if len(validate.Report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], "validate config", report.Canceled)
	checkClusterStatus(t, validate.Report, &report.ClustersStatus{})
	checkSummary(t, validate.Report, report.Summary{})
}

func TestValidateClusterGatherClusterFailed(t *testing.T) {
	validate := testCommand(t, clustersGatherDataFailed, testK8s)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, testK8s.namespaces)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.Report.Steps[1], "validate clusters", report.Failed)

	// If gathering data fail for some of the clusters, we skip the validation step.
	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Failed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
	}
	checkItems(t, validate.Report.Steps[1], items)
	checkClusterStatus(t, validate.Report, &report.ClustersStatus{})
	checkSummary(t, validate.Report, report.Summary{})
}

func TestValidateClustersInspectS3ProfilesFailed(t *testing.T) {
	validate := testCommand(t, &helpers.ValidationMock{}, testK8s)
	// We don't add test data to cause inspect S3 profiles to fail.
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkApplication(t, validate.Report, nil)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.Report.Steps[1], "validate clusters", report.Failed)

	// Inspect S3 profiles fails, check S3 is skipped. Validation runs and reports missing S3
	// status as problem.
	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Failed},
		{Name: "validate clusters data", Status: report.Failed},
	}
	checkItems(t, validate.Report.Steps[1], items)
	empty := &report.ClustersStatus{}
	if validate.Report.ClustersStatus.Equal(empty) {
		t.Fatal("clusters status is empty")
	}
	checkSummary(t, validate.Report, report.Summary{summary.Problem: 12})
}

func TestValidateClustersInspectS3ProfilesCanceled(t *testing.T) {
	validate := testCommand(t, helpers.GetSecretCanceled, testK8s)
	helpers.AddGatheredData(t, validate.DataDir(), k8sTestdata, validate.Report.Name)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Canceled)
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, testK8s.namespaces)

	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.Report.Steps[1], "validate clusters", report.Canceled)

	// Inspect S3 profiles is canceled, checkS3 and validation are skipped.
	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Canceled},
	}
	checkItems(t, validate.Report.Steps[1], items)
	checkClusterStatus(t, validate.Report, &report.ClustersStatus{})
	checkSummary(t, validate.Report, report.Summary{})
}

func TestValidateClustersGetSecretFailed(t *testing.T) {
	validate := testCommand(t, helpers.GetSecretFailed, testK8s)
	helpers.AddGatheredData(t, validate.DataDir(), k8sTestdata, validate.Report.Name)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, testK8s.namespaces)

	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.Report.Steps[1], "validate clusters", report.Failed)

	// When GetSecret returns an error. The profile will have empty credentials
	// causing checkS3 and validation to fail.
	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{Name: "check S3 profile \"minio-on-dr1\"", Status: report.Failed},
		{Name: "check S3 profile \"minio-on-dr2\"", Status: report.Failed},
		{Name: "validate clusters data", Status: report.Failed},
	}
	checkItems(t, validate.Report.Steps[1], items)
	checkSummary(t, validate.Report, report.Summary{summary.OK: 88, summary.Problem: 2})
}

func TestValidateClustersGetSecretInvalid(t *testing.T) {
	validate := testCommand(t, helpers.GetSecretInvalid, testK8s)
	helpers.AddGatheredData(t, validate.DataDir(), k8sTestdata, validate.Report.Name)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, testK8s.namespaces)

	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.Report.Steps[1], "validate clusters", report.Failed)

	// When GetSecret returns a secret with invalid value, causing checkS3 and
	// validation to fail.
	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{Name: "check S3 profile \"minio-on-dr1\"", Status: report.Failed},
		{Name: "check S3 profile \"minio-on-dr2\"", Status: report.Failed},
		{Name: "validate clusters data", Status: report.Failed},
	}
	checkItems(t, validate.Report.Steps[1], items)
	checkSummary(t, validate.Report, report.Summary{summary.OK: 88, summary.Problem: 2})
}

func TestValidateClustersCheckS3Failed(t *testing.T) {
	validate := testCommand(t, checkS3Failed, testK8s)
	helpers.AddGatheredData(t, validate.DataDir(), k8sTestdata, validate.Report.Name)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, testK8s.namespaces)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.Report.Steps[1], "validate clusters", report.Failed)

	// Check s3 fails for one profile, other profile succeeds. Validation runs and reports the
	// failed profile as problem.
	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{Name: "check S3 profile \"minio-on-dr1\"", Status: report.Failed},
		{Name: "check S3 profile \"minio-on-dr2\"", Status: report.Passed},
		{Name: "validate clusters data", Status: report.Failed},
	}
	checkItems(t, validate.Report.Steps[1], items)
	empty := &report.ClustersStatus{}
	if validate.Report.ClustersStatus.Equal(empty) {
		t.Fatal("clusters status is empty")
	}
	checkSummary(
		t,
		validate.Report,
		report.Summary{summary.OK: 89, summary.Problem: 1},
	)
}

func TestValidateClustersCheckS3Canceled(t *testing.T) {
	validate := testCommand(t, checkS3Canceled, testK8s)
	helpers.AddGatheredData(t, validate.DataDir(), k8sTestdata, validate.Report.Name)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Canceled)
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, testK8s.namespaces)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.Report.Steps[1], "validate clusters", report.Canceled)

	// Check S3 is canceled, validation is skipped.
	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{Name: "check S3 profile \"minio-on-dr1\"", Status: report.Canceled},
		{Name: "check S3 profile \"minio-on-dr2\"", Status: report.Canceled},
	}
	checkItems(t, validate.Report.Steps[1], items)
	checkClusterStatus(t, validate.Report, &report.ClustersStatus{})
	checkSummary(t, validate.Report, report.Summary{})
}
