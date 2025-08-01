// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report_test

import (
	"fmt"
	"testing"

	"github.com/ramendr/ramenctl/pkg/report"
	"sigs.k8s.io/yaml"
)

const (
	modified = "modified"
)

func TestReportApplicationEqual(t *testing.T) {
	a1 := testApplication()
	t.Run("equal to self", func(t *testing.T) {
		a2 := a1
		checkApplicationsEqual(t, a1, a2)
	})
	t.Run("equal applications", func(t *testing.T) {
		a2 := testApplication()
		checkApplicationsEqual(t, a1, a2)
	})
	t.Run("equal hub", func(t *testing.T) {
		a2 := testApplication()
		hub := *a1.Hub
		a2.Hub = &hub
		checkApplicationsEqual(t, a1, a2)
	})
	t.Run("equal primary cluster", func(t *testing.T) {
		a2 := testApplication()
		cluster := *a1.PrimaryCluster
		a2.PrimaryCluster = &cluster
		checkApplicationsEqual(t, a1, a2)
	})
	t.Run("equal secondary cluster", func(t *testing.T) {
		a2 := testApplication()
		cluster := *a1.SecondaryCluster
		a2.SecondaryCluster = &cluster
		checkApplicationsEqual(t, a1, a2)
	})
	t.Run("no cluster", func(t *testing.T) {
		a1 := &report.Application{}
		a2 := &report.Application{}
		checkApplicationsEqual(t, a1, a2)
	})
}

func TestReportApplicationNotEqual(t *testing.T) {
	a1 := testApplication()
	t.Run("not equal to nil", func(t *testing.T) {
		var a2 *report.Application
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("name", func(t *testing.T) {
		a2 := testApplication()
		a2.Name = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("namespace", func(t *testing.T) {
		a2 := testApplication()
		a2.Namespace = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("hub nil", func(t *testing.T) {
		a2 := testApplication()
		a2.Hub = nil
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("hub drpc deleted", func(t *testing.T) {
		a2 := testApplication()
		a2.Hub.DRPC.Deleted = true
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("hub drpc action", func(t *testing.T) {
		a2 := testApplication()
		a2.Hub.DRPC.Action = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("hub drpc drPolicy", func(t *testing.T) {
		a2 := testApplication()
		a2.Hub.DRPC.DRPolicy = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("hub drpc phase", func(t *testing.T) {
		a2 := testApplication()
		a2.Hub.DRPC.Phase = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("hub drpc progression", func(t *testing.T) {
		a2 := testApplication()
		a2.Hub.DRPC.Progression = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("hub drpc conditions nil", func(t *testing.T) {
		a2 := testApplication()
		a2.Hub.DRPC.Conditions = nil
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("hub drpc conditions", func(t *testing.T) {
		a2 := testApplication()
		a2.Hub.DRPC.Conditions["Protected"] = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster nil", func(t *testing.T) {
		a2 := testApplication()
		a2.PrimaryCluster = nil
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster name", func(t *testing.T) {
		a2 := testApplication()
		a2.PrimaryCluster.Name = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg deleted", func(t *testing.T) {
		a2 := testApplication()
		a2.PrimaryCluster.VRG.Deleted = true
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg namespace", func(t *testing.T) {
		a2 := testApplication()
		a2.PrimaryCluster.VRG.Namespace = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg state", func(t *testing.T) {
		a2 := testApplication()
		a2.PrimaryCluster.VRG.State = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg conditions nil", func(t *testing.T) {
		a2 := testApplication()
		a2.PrimaryCluster.VRG.Conditions = nil
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg conditions", func(t *testing.T) {
		a2 := testApplication()
		a2.PrimaryCluster.VRG.Conditions["dataReady"] = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg protectedpvcs nil", func(t *testing.T) {
		a2 := testApplication()
		a2.PrimaryCluster.VRG.ProtectedPVCs = nil
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg protectedpvcs name", func(t *testing.T) {
		a2 := testApplication()
		a2.PrimaryCluster.VRG.ProtectedPVCs[0].Name = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg protectedpvcs namespace", func(t *testing.T) {
		a2 := testApplication()
		a2.PrimaryCluster.VRG.ProtectedPVCs[0].Namespace = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg protectedpvcs phase", func(t *testing.T) {
		a2 := testApplication()
		a2.PrimaryCluster.VRG.ProtectedPVCs[0].Phase = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg protectedpvcs conditions nil", func(t *testing.T) {
		a2 := testApplication()
		a2.PrimaryCluster.VRG.ProtectedPVCs[0].Conditions = nil
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("primary cluster vrg protectedpvcs conditions", func(t *testing.T) {
		a2 := testApplication()
		a2.PrimaryCluster.VRG.ProtectedPVCs[0].Conditions["dataReady"] = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("secondary cluster nil", func(t *testing.T) {
		a2 := testApplication()
		a2.SecondaryCluster = nil
		checkApplicationsNotEqual(t, a1, a2)
		checkApplicationsNotEqual(t, a2, a1)
	})
	t.Run("secondary cluster name", func(t *testing.T) {
		a2 := testApplication()
		a2.SecondaryCluster.Name = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("secondary cluster vrg deleted", func(t *testing.T) {
		a2 := testApplication()
		a2.SecondaryCluster.VRG.Deleted = true
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("secondary cluster vrg namespace", func(t *testing.T) {
		a2 := testApplication()
		a2.SecondaryCluster.VRG.Namespace = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("secondary cluster vrg state", func(t *testing.T) {
		a2 := testApplication()
		a2.SecondaryCluster.VRG.State = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("secondary cluster vrg conditions nil", func(t *testing.T) {
		a2 := testApplication()
		a2.SecondaryCluster.VRG.Conditions = nil
		checkApplicationsNotEqual(t, a1, a2)
	})
	t.Run("secondary cluster vrg conditions", func(t *testing.T) {
		a2 := testApplication()
		a2.SecondaryCluster.VRG.Conditions["noClusterDataConflict"] = modified
		checkApplicationsNotEqual(t, a1, a2)
	})
}

func TestReportApplicationMarshaling(t *testing.T) {
	a1 := testApplication()
	data, err := yaml.Marshal(a1)
	if err != nil {
		t.Fatal(err)
	}
	// For inspecting the generated yaml.
	fmt.Print(string(data))
	a2 := &report.Application{}
	if err := yaml.Unmarshal(data, a2); err != nil {
		t.Fatal(err)
	}
	checkApplicationsEqual(t, a1, a2)
}

func testApplication() *report.Application {
	a := &report.Application{
		Name:      "drpc",
		Namespace: "drpc-namespace",
		Hub: &report.ApplicationHubStatus{
			DRPC: report.DRPCSummary{
				DRPolicy:    "dr-policy-1m",
				Phase:       "Deployed",
				Progression: "completed",
				Conditions: map[string]string{
					"available": "ok",
					"peerReady": "ok",
					"protected": "ok",
				},
			},
		},
		PrimaryCluster: &report.ApplicationClusterStatus{
			Name: "dr1",
			VRG: report.VRGSummary{
				Namespace: "app-namespace",
				State:     "Primary",
				Conditions: map[string]string{
					"dataReady":             "ok",
					"dataProtected":         "ok",
					"clusterDataReady":      "ok",
					"clusterDataProtected":  "ok",
					"kubeObjectsReady":      "ok",
					"noClusterDataConflict": "ok",
				},
				ProtectedPVCs: []report.ProtectedPVCSummary{
					{
						Name:      "pvc1",
						Namespace: "app-namespace",
						Phase:     "Bound",
						Conditions: map[string]string{
							"dataReady":            "ok",
							"clusterDataProtected": "ok",
							"dataProtected":        "ok",
						},
					},
				},
			},
		},
		SecondaryCluster: &report.ApplicationClusterStatus{
			Name: "dr2",
			VRG: report.VRGSummary{
				Namespace: "app-namespace",
				State:     "Secondary",
				Conditions: map[string]string{
					"noClusterDataConflict": "ok",
				},
			},
		},
	}
	return a
}

func checkApplicationsEqual(t *testing.T, a, b *report.Application) {
	if !a.Equal(b) {
		t.Fatalf(
			"applications are not equal\n%s\n%s",
			marshalApplication(t, a),
			marshalApplication(t, b),
		)
	}
}

func checkApplicationsNotEqual(t *testing.T, a, b *report.Application) {
	if a.Equal(b) {
		t.Fatalf("applications are equal\n%s\n%s",
			marshalApplication(t, a),
			marshalApplication(t, b),
		)
	}
}

func marshalApplication(t *testing.T, a *report.Application) string {
	data, err := yaml.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
