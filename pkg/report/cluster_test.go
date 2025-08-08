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
		var c2 *report.ClusterStatus
		checkClustersNotEqual(t, c1, c2)
	})

	// Hub DRClusters tests
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
	t.Run("hub drcluster conditions nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters[0].Conditions = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drcluster conditions", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters[0].Conditions["Validated"] = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drclusters nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRClusters = nil
		checkClustersNotEqual(t, c1, c2)
	})

	// Hub DRPolicies tests
	t.Run("hub drpolicy name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies[0].Name = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy drclusters", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies[0].DRClusters = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy conditions nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies[0].Conditions = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicy conditions", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies[0].Conditions["Validated"] = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub drpolicies nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.DRPolicies = nil
		checkClustersNotEqual(t, c1, c2)
	})

	// Hub Ramen ConfigMap tests
	t.Run("hub ramen configmap name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.ConfigMap.Name = modified
		checkClustersNotEqual(t, c1, c2)
	})

	// Hub Ramen Operator tests
	t.Run("hub ramen operator name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Operator.Name = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen operator conditions nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Operator.Conditions = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("hub ramen operator conditions", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.Hub.Ramen.Operator.Conditions["Available"] = modified
		checkClustersNotEqual(t, c1, c2)
	})

	// DRClusters tests
	t.Run("drcluster name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.DRClusters[0].Name = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("drcluster ramen configmap name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.DRClusters[0].Ramen.ConfigMap.Name = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("drcluster ramen operator name", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.DRClusters[0].Ramen.Operator.Name = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("drcluster ramen operator conditions nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.DRClusters[0].Ramen.Operator.Conditions = nil
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("drcluster ramen operator conditions", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.DRClusters[0].Ramen.Operator.Conditions["Available"] = modified
		checkClustersNotEqual(t, c1, c2)
	})
	t.Run("drclusters nil", func(t *testing.T) {
		c2 := testClusterStatus()
		c2.DRClusters = nil
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
	c2 := &report.ClusterStatus{}
	if err := yaml.Unmarshal(data, c2); err != nil {
		t.Fatal(err)
	}
	checkClustersEqual(t, c1, c2)
}

func testClusterStatus() *report.ClusterStatus {
	c := &report.ClusterStatus{
		Hub: report.HubClusterStatus{
			DRClusters: []report.DRClusterSummary{
				{
					Name:  "dr1",
					Phase: "Available",
					Conditions: map[string]string{
						"Validated": "ok",
					},
				},
				{
					Name:  "dr2",
					Phase: "Available",
					Conditions: map[string]string{
						"Validated": "ok",
					},
				},
			},
			DRPolicies: []report.DRPolicySummary{
				{
					Name:       "dr-policy-5m",
					DRClusters: "ok",
					Conditions: map[string]string{
						"Validated": "ok",
					},
				},
				{
					Name:       "dr-policy-7m",
					DRClusters: "ok",
					Conditions: map[string]string{
						"Validated": "ok",
					},
				},
			},
			Ramen: report.RamenSummary{
				ConfigMap: report.ConfigMapSummary{
					Name: "ramen-hub-operator-config",
				},
				Operator: report.OperatorSummary{
					Name: "ramen-hub-operator",
					Conditions: map[string]string{
						"Available": "ok",
					},
				},
			},
		},
		DRClusters: []report.DRClusterStatus{
			{
				Name: "dr1",
				Ramen: report.RamenSummary{
					ConfigMap: report.ConfigMapSummary{
						Name: "ramen-dr-cluster-operator-config",
					},
					Operator: report.OperatorSummary{
						Name: "ramen-dr-cluster-operator",
						Conditions: map[string]string{
							"Available": "ok",
						},
					},
				},
			},
			{
				Name: "dr2",
				Ramen: report.RamenSummary{
					ConfigMap: report.ConfigMapSummary{
						Name: "ramen-dr-cluster-operator-config",
					},
					Operator: report.OperatorSummary{
						Name: "ramen-dr-cluster-operator",
						Conditions: map[string]string{
							"Available": "ok",
						},
					},
				},
			},
		},
	}
	return c
}

func checkClustersEqual(t *testing.T, c1, c2 *report.ClusterStatus) {
	if !c1.Equal(c2) {
		t.Fatalf("clusters are not equal\n%s\n%s", marshal(t, c1), marshal(t, c2))
	}
}

func checkClustersNotEqual(t *testing.T, c1, c2 *report.ClusterStatus) {
	if c1.Equal(c2) {
		t.Fatalf("clusters are equal\n%s\n%s", marshal(t, c1), marshal(t, c2))
	}
}
