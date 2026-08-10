// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	e2econfig "github.com/ramendr/ramen/e2e/config"
	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
)

var reportConfig = &e2econfig.Config{
	Distro:     "k8s",
	Repo:       e2econfig.Repo{URL: "https://github.com/org/repo", Branch: "main"},
	DRPolicy:   "dr-policy",
	ClusterSet: "clusterset",
	Clusters: map[string]e2econfig.Cluster{
		"hub": {Kubeconfig: "hub-kubeconfig"},
		"c1":  {Kubeconfig: "c1-kubeconfig"},
		"c2":  {Kubeconfig: "c2-kubeconfig"},
	},
	PVCSpecs: []e2econfig.PVCSpec{
		{Name: "rbd", StorageClassName: "rook-ceph-block", AccessModes: "ReadWriteOnce"},
		{Name: "cephfs", StorageClassName: "rook-cephfs-fs", AccessModes: "ReadWriteMany"},
	},
	Tests: []e2econfig.Test{
		{Workload: "appset", Deployer: "deploy", PVCSpec: "rbd"},
		{Workload: "subscr", Deployer: "deploy", PVCSpec: "rbd"},
		{Workload: "disapp", Deployer: "deploy", PVCSpec: "cephfs"},
	},
	Channel: e2econfig.Channel{
		Name:      "my-channel",
		Namespace: "test-gitops",
	},
	Namespaces: e2econfig.Namespaces{
		RamenHubNamespace:       "ramen-system",
		RamenDRClusterNamespace: "ramen-system",
		RamenOpsNamespace:       "ramen-ops",
		ArgocdNamespace:         "argocd",
	},
}

func TestReportSummary(t *testing.T) {
	r := newReport("test-command", reportConfig)
	testsStep := &report.Step{
		Name:     TestsStep,
		Status:   report.Passed,
		Duration: 1.0,
		Items: []*report.Step{
			{Name: "test1", Status: report.Passed, Duration: 1.0},
			{Name: "test2", Status: report.Failed, Duration: 1.0},
			{Name: "test3", Status: report.Skipped, Duration: 1.0},
			{Name: "test4", Status: report.Canceled, Duration: 1.0},
		},
	}
	r.AddStep(testsStep)

	expectedSummary := &report.Summary{
		Passed:   1,
		Failed:   1,
		Skipped:  1,
		Canceled: 1,
	}
	if !r.Summary.Equal(expectedSummary) {
		t.Errorf("expected summary %+v, got %+v", expectedSummary, r.Summary)
	}
}

func TestReportEqual(t *testing.T) {
	helpers.FakeTime(t)
	// Helper function to create a standard report
	createReport := func() *Report {
		r := newReport("test-command", reportConfig)
		r.Summary = &report.Summary{Passed: 2}
		return r
	}

	r1 := createReport()

	t.Run("equal to self", func(t *testing.T) {
		r2 := r1
		if !r1.Equal(r2) {
			diff := helpers.UnifiedDiff(t, r1, r2)
			t.Fatalf("report should be equal itself\n%s", diff)
		}
	})

	t.Run("not equal to nil", func(t *testing.T) {
		if r1.Equal(nil) {
			t.Error("report should not be equal nil")
		}
	})

	t.Run("equal reports", func(t *testing.T) {
		r2 := createReport()
		if !r1.Equal(r2) {
			diff := helpers.UnifiedDiff(t, r1, r2)
			t.Fatalf("identical reports are not equal\n%s", diff)
		}
	})

	t.Run("different config content", func(t *testing.T) {
		r2 := createReport()
		differentConfig := *reportConfig
		differentConfig.DRPolicy = "different-dr-policy"
		r2.Config = &differentConfig
		if r1.Equal(r2) {
			t.Error("reports with different config content should not be equal")
		}
	})

	t.Run("nil config", func(t *testing.T) {
		r2 := createReport()
		r2.Config = nil
		if r1.Equal(r2) || r2.Equal(r1) {
			t.Error("reports with one nil config should not be equal")
		}
	})

}

func TestReportMarshaling(t *testing.T) {
	helpers.FakeTime(t)
	r := newReport("test-command", reportConfig)
	r.Status = report.Failed
	r.Duration = 2.0
	r.Steps = []*report.Step{
		{
			Name:     "step1",
			Status:   report.Passed,
			Duration: 1.0,
			Items: []*report.Step{
				{Name: "subitem1", Status: report.Passed, Duration: 1.0},
				{Name: "subitem2", Status: report.Passed, Duration: 1.0},
			},
		},
		{
			Name:     "step2",
			Status:   report.Failed,
			Duration: 1.0,
		},
	}
	r.Summary = &report.Summary{Passed: 2, Failed: 1}

	// Test roundtrip marshaling/unmarshaling
	checkRoundtrip(t, r)
}

func TestSummaryString(t *testing.T) {
	summary := &report.Summary{
		Passed:   5,
		Failed:   2,
		Skipped:  3,
		Canceled: 1,
	}
	expectedString := "5 passed, 2 failed, 3 skipped, 1 canceled"
	if summaryString(summary) != expectedString {
		t.Errorf("expected summary string %s, got %s", expectedString, summaryString(summary))
	}
}

func TestSummaryCount(t *testing.T) {
	summary := &report.Summary{}

	// Add multiple tests of different status
	addTest(summary, &report.Step{Status: report.Passed})
	addTest(summary, &report.Step{Status: report.Passed})
	addTest(summary, &report.Step{Status: report.Failed})
	addTest(summary, &report.Step{Status: report.Skipped})
	addTest(summary, &report.Step{Status: report.Canceled})
	addTest(summary, &report.Step{Status: report.Passed})

	expectedSummary := &report.Summary{
		Passed:   3,
		Failed:   1,
		Skipped:  1,
		Canceled: 1,
	}
	if !summary.Equal(expectedSummary) {
		t.Errorf("expected summary %+v, got %+v", expectedSummary, summary)
	}
}

func checkRoundtrip(t *testing.T, r1 *Report) {
	// We must be able to marshal and unmarshal the report
	b, err := yaml.Marshal(r1)
	if err != nil {
		t.Fatalf("failed to marshal report: %s", err)
	}
	r2 := &Report{}
	if err := yaml.Unmarshal(b, r2); err != nil {
		t.Fatalf("failed to unmarshal report: %s", err)
	}
	if !r1.Equal(r2) {
		diff := helpers.UnifiedDiff(t, r1, r2)
		t.Fatalf("unmarshaled reports not equal\n%s", diff)
	}
}
