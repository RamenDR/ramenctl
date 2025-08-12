// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report_test

import (
	"fmt"
	"testing"

	"sigs.k8s.io/yaml"

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
		c2.Hub.DRClusters = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drcluster conditions nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters[0].Conditions = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drcluster name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters[0].Name = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drcluster phase", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters[0].Phase = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drcluster conditions", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters[0].Conditions[0].State = report.Error
		checkClustersNotEqual(t, c1, c2)
	})

	// Hub drPolicies tests

	t.Run("hub drpolicies nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy conditions nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies[0].Conditions = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies[0].Name = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy drclusters", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies[0].DRClusters[0] = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy scheduling interval", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies[0].SchedulingInterval = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy conditions", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies[0].Conditions[0].State = report.Error
		checkClustersNotEqual(t, c1, c2)
	})

	// Hub ramen configmap tests

	t.Run("hub ramen configmap name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.Name = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen configmap namespace", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.Namespace = modified
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
		c2.Hub.Ramen.Deployment.Name = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen deployment namespace", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Deployment.Namespace = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen deployment conditions", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Deployment.Conditions[0].State = report.Error
		checkClustersNotEqual(t, c1, c2)
	})

	// Managed cluster tests

	t.Run("clusters nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen deployment conditions nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.Deployment.Conditions = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Name = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.Name = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen configmap namespace", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.ConfigMap.Namespace = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen deployment name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.Deployment.Name = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen deployment namespace", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.Deployment.Namespace = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("cluster ramen deployment conditions", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Clusters[0].Ramen.Deployment.Conditions[0].State = report.Error
		checkClustersNotEqual(t, c1, c2)
	})
}

func TestReportClusterStatusMarshaling(t *testing.T) {
	c1 := testClusterStatus()
	data, err := yaml.Marshal(c1)
	if err != nil {
		t.Fatal(err)
	}
	// For inspecting the generated yaml.
	fmt.Print(string(data))
	c2 := &report.ClustersStatus{}
	if err := yaml.Unmarshal(data, c2); err != nil {
		t.Fatal(err)
	}
	checkClustersEqual(t, c1, c2)
}

func testClusterStatus() *report.ClustersStatus {
	c := &report.ClustersStatus{
		Hub: report.ClustersStatusHub{
			DRClusters: []report.DRClusterSummary{
				{
					Name:  "dr1",
					Phase: "Available",
					Conditions: []report.ValidatedCondition{
						{
							Type: "Fenced",
							Validated: report.Validated{
								State: report.OK,
							},
						},
						{
							Type: "Clean",
							Validated: report.Validated{
								State: report.OK,
							},
						},
						{
							Type: "Validated",
							Validated: report.Validated{
								State: report.OK,
							},
						},
					},
				},
				{
					Name:  "dr2",
					Phase: "Available",
					Conditions: []report.ValidatedCondition{
						{
							Type: "Fenced",
							Validated: report.Validated{
								State: report.OK,
							},
						},
						{
							Type: "Clean",
							Validated: report.Validated{
								State: report.OK,
							},
						},
						{
							Type: "Validated",
							Validated: report.Validated{
								State: report.OK,
							},
						},
					},
				},
			},
			DRPolicies: []report.DRPolicySummary{
				{
					Name:               "dr-policy-1m",
					DRClusters:         []string{"dr1", "dr2"},
					SchedulingInterval: "1m",
					Conditions: []report.ValidatedCondition{
						{
							Type: "Validated",
							Validated: report.Validated{
								State: report.OK,
							},
						},
					},
				},
				{
					Name:               "dr-policy-5m",
					DRClusters:         []string{"dr1", "dr2"},
					SchedulingInterval: "5m",
					Conditions: []report.ValidatedCondition{
						{
							Type: "Validated",
							Validated: report.Validated{
								State: report.OK,
							},
						},
					},
				},
			},
			Ramen: report.RamenSummary{
				ConfigMap: report.ConfigMapSummary{
					Name:      "ramen-hub-operator-config",
					Namespace: "ramen-system",
				},
				Deployment: report.DeploymentSummary{
					Name:      "ramen-hub-operator",
					Namespace: "ramen-system",
					Conditions: []report.ValidatedCondition{
						{
							Type: "Available",
							Validated: report.Validated{
								State: report.OK,
							},
						},
						{
							Type: "Progressing",
							Validated: report.Validated{
								State: report.OK,
							},
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
						Name:      "ramen-dr-cluster-operator-config",
						Namespace: "ramen-system",
					},
					Deployment: report.DeploymentSummary{
						Name:      "ramen-dr-cluster-operator",
						Namespace: "ramen-system",
						Conditions: []report.ValidatedCondition{
							{
								Type: "Available",
								Validated: report.Validated{
									State: report.OK,
								},
							},
							{
								Type: "Progressing",
								Validated: report.Validated{
									State: report.OK,
								},
							},
						},
					},
				},
			},
			{
				Name: "dr2",
				Ramen: report.RamenSummary{
					ConfigMap: report.ConfigMapSummary{
						Name:      "ramen-dr-cluster-operator-config",
						Namespace: "ramen-system",
					},
					Deployment: report.DeploymentSummary{
						Name:      "ramen-dr-cluster-operator",
						Namespace: "ramen-system",
						Conditions: []report.ValidatedCondition{
							{
								Type: "Available",
								Validated: report.Validated{
									State: report.OK,
								},
							},
							{
								Type: "Progressing",
								Validated: report.Validated{
									State: report.OK,
								},
							},
						},
					},
				},
			},
		},
	}
	return c
}

func checkClustersEqual(t *testing.T, c1, c2 *report.ClustersStatus) {
	if !c1.Equal(c2) {
		t.Fatalf("clusters are not equal\n%s\n%s", marshal(t, c1), marshal(t, c2))
	}
}

func checkClustersNotEqual(t *testing.T, c1, c2 *report.ClustersStatus) {
	if c1.Equal(c2) {
		t.Fatalf("clusters are equal\n%s\n%s", marshal(t, c1), marshal(t, c2))
	}
}
