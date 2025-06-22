// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"testing"

	"sigs.k8s.io/yaml"

	e2econfig "github.com/ramendr/ramen/e2e/config"

	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/time"
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

	expectedSummary := Summary{Passed: 1, Failed: 1, Skipped: 1, Canceled: 1}
	if r.Summary != expectedSummary {
		t.Errorf("expected summary %+v, got %+v", expectedSummary, r.Summary)
	}
}

func TestReportEqual(t *testing.T) {
	fakeTime(t)
	// Helper function to create a standard report
	createReport := func() *Report {
		r := newReport("test-command", reportConfig)
		r.Summary = Summary{Passed: 2}
		return r
	}

	r1 := createReport()

	// Intentionally comparing report to itself
	//nolint:gocritic
	t.Run("equal to self", func(t *testing.T) {
		if !r1.Equal(r1) {
			t.Error("report should be equal itself")
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
			t.Error("reports with identical content should be equal")
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

	t.Run("different summary", func(t *testing.T) {
		r2 := createReport()
		r2.Summary = Summary{Passed: 1, Failed: 1}
		if r1.Equal(r2) {
			t.Error("reports with different summary should not be equal")
		}
	})
}

func TestReportMarshaling(t *testing.T) {
	fakeTime(t)
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
	r.Summary = Summary{Passed: 2, Failed: 1}

	// Test roundtrip marshaling/unmarshaling
	checkRoundtrip(t, r)
}

func TestSummaryString(t *testing.T) {
	summary := Summary{Passed: 5, Failed: 2, Skipped: 3, Canceled: 1}
	expectedString := "5 passed, 2 failed, 3 skipped, 1 canceled"
	if summary.String() != expectedString {
		t.Errorf("expected summary string %s, got %s", expectedString, summary.String())
	}
}

func TestSummaryMarshal(t *testing.T) {
	summary := Summary{Passed: 4, Failed: 3, Skipped: 2, Canceled: 1}

	bytes, err := yaml.Marshal(summary)
	if err != nil {
		t.Fatalf("failed to marshal summary: %v", err)
	}

	var unmarshaledSummary Summary
	err = yaml.Unmarshal(bytes, &unmarshaledSummary)
	if err != nil {
		t.Fatalf("failed to unmarshal summary: %v", err)
	}
	if unmarshaledSummary != summary {
		t.Errorf("unmarshaled summary %+v does not match original summary %+v",
			unmarshaledSummary, summary)
	}
}

func TestSummaryCount(t *testing.T) {
	summary := Summary{}

	// Add multiple tests of different status
	summary.AddTest(&report.Step{Status: report.Passed})
	summary.AddTest(&report.Step{Status: report.Passed})
	summary.AddTest(&report.Step{Status: report.Failed})
	summary.AddTest(&report.Step{Status: report.Skipped})
	summary.AddTest(&report.Step{Status: report.Canceled})
	summary.AddTest(&report.Step{Status: report.Passed})

	expectedSummary := Summary{Passed: 3, Failed: 1, Skipped: 1, Canceled: 1}
	if summary != expectedSummary {
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
		t.Fatalf("expected report %+v, got %+v", r1, r2)
	}
}

var fakeNow = time.Now()

func fakeTime(t *testing.T) {
	savedNow := time.Now
	time.Now = func() time.Time {
		return fakeNow
	}
	t.Cleanup(func() {
		time.Now = savedNow
	})
}
