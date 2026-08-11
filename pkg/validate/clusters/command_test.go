// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package clusters

import (
	"testing"

	"github.com/ramendr/ramenctl/pkg/helpers"
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
	checkError(t, validate.Report, "")
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, testK8s.namespaces)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate clusters",
		Status: report.Passed,
	})

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

	expected := loadClustersStatus(t, "k8s-status.yaml")
	checkClusterStatus(t, validate.Report, expected)

	checkSummary(t, validate.Report, report.Summary{summary.OK: 93})
}

func TestValidateClustersOcp(t *testing.T) {
	validate := testCommand(t, &helpers.ValidationMock{}, testOcp)
	helpers.AddGatheredData(t, validate.DataDir(), ocpTestdata, validate.Report.Name)
	if err := validate.Run(); err != nil {
		dumpCommandLog(t, validate)
		t.Fatal(err)
	}
	checkReport(t, validate, report.Passed)
	checkError(t, validate.Report, "")
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, testOcp.namespaces)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate clusters",
		Status: report.Passed,
	})

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

	expected := loadClustersStatus(t, "ocp-status.yaml")
	checkClusterStatus(t, validate.Report, expected)

	checkSummary(t, validate.Report, report.Summary{summary.OK: 91})
}

func TestValidateClustersValidateFailed(t *testing.T) {
	validate := testCommand(t, helpers.ValidateConfigFailed, testK8s)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkError(t, validate.Report, "Failed to validate config")
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, nil)
	if len(validate.Report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Failed,
		Err:    "Failed to validate config",
	})
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
	checkError(t, validate.Report, "Canceled validate config")
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, nil)
	if len(validate.Report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Canceled,
		Err:    "Canceled validate config",
	})
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
	checkError(t, validate.Report, "Failed to gather data from clusters hub")
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, testK8s.namespaces)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate clusters",
		Status: report.Failed,
		Err:    "Failed to gather data from clusters hub",
	})

	// If gathering data fail for some of the clusters, we skip the validation step.
	items := []*report.Step{
		{
			Name:   "gather \"hub\"",
			Status: report.Failed,
			Err:    `Failed to gather data from cluster "hub"`,
		},
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
	checkError(t, validate.Report,
		"Validation failed (0 ok, 0 warning, 12 problem)")
	checkApplication(t, validate.Report, nil)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate clusters",
		Status: report.Failed,
	})

	// Inspect S3 profiles fails, check S3 is skipped. Validation runs and reports missing S3
	// status as problem.
	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{
			Name:   "inspect S3 profiles",
			Status: report.Failed,
			Err:    "Failed to read S3 profiles from hub",
		},
		{
			Name:   "validate clusters data",
			Status: report.Failed,
			Err:    "Validation failed (0 ok, 0 warning, 12 problem)",
		},
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
	checkError(t, validate.Report, "Canceled inspect S3 profiles")
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, testK8s.namespaces)

	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate clusters",
		Status: report.Canceled,
	})

	// Inspect S3 profiles is canceled, checkS3 and validation are skipped.
	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{
			Name:   "inspect S3 profiles",
			Status: report.Canceled,
			Err:    "Canceled inspect S3 profiles",
		},
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
	checkError(t, validate.Report,
		"Failed to check S3 profiles minio-on-dr1, minio-on-dr2")
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, testK8s.namespaces)

	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate clusters",
		Status: report.Failed,
		Err:    "Failed to check S3 profiles minio-on-dr1, minio-on-dr2",
	})

	// When GetSecret returns an error. The profile will have empty credentials
	// causing checkS3 and validation to fail.
	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{
			Name:   "check S3 profile \"minio-on-dr1\"",
			Status: report.Failed,
			Err:    `Failed to check S3 profile "minio-on-dr1"`,
		},
		{
			Name:   "check S3 profile \"minio-on-dr2\"",
			Status: report.Failed,
			Err:    `Failed to check S3 profile "minio-on-dr2"`,
		},
		{
			Name:   "validate clusters data",
			Status: report.Failed,
			Err:    "Validation failed (91 ok, 0 warning, 2 problem)",
		},
	}
	checkItems(t, validate.Report.Steps[1], items)
	checkSummary(t, validate.Report, report.Summary{summary.OK: 91, summary.Problem: 2})
}

func TestValidateClustersGetSecretInvalid(t *testing.T) {
	validate := testCommand(t, helpers.GetSecretInvalid, testK8s)
	helpers.AddGatheredData(t, validate.DataDir(), k8sTestdata, validate.Report.Name)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkError(t, validate.Report,
		"Failed to check S3 profiles minio-on-dr1, minio-on-dr2")
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, testK8s.namespaces)

	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate clusters",
		Status: report.Failed,
		Err:    "Failed to check S3 profiles minio-on-dr1, minio-on-dr2",
	})

	// When GetSecret returns a secret with invalid value, causing checkS3 and
	// validation to fail.
	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{
			Name:   "check S3 profile \"minio-on-dr1\"",
			Status: report.Failed,
			Err:    `Failed to check S3 profile "minio-on-dr1"`,
		},
		{
			Name:   "check S3 profile \"minio-on-dr2\"",
			Status: report.Failed,
			Err:    `Failed to check S3 profile "minio-on-dr2"`,
		},
		{
			Name:   "validate clusters data",
			Status: report.Failed,
			Err:    "Validation failed (91 ok, 0 warning, 2 problem)",
		},
	}
	checkItems(t, validate.Report.Steps[1], items)
	checkSummary(t, validate.Report, report.Summary{summary.OK: 91, summary.Problem: 2})
}

func TestValidateClustersCheckS3Failed(t *testing.T) {
	validate := testCommand(t, checkS3Failed, testK8s)
	helpers.AddGatheredData(t, validate.DataDir(), k8sTestdata, validate.Report.Name)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkError(t, validate.Report, "Failed to check S3 profiles minio-on-dr1")
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, testK8s.namespaces)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate clusters",
		Status: report.Failed,
		Err:    "Failed to check S3 profiles minio-on-dr1",
	})

	// Check s3 fails for one profile, other profile succeeds. Validation runs and reports the
	// failed profile as problem.
	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{
			Name:   "check S3 profile \"minio-on-dr1\"",
			Status: report.Failed,
			Err:    `Failed to check S3 profile "minio-on-dr1"`,
		},
		{Name: "check S3 profile \"minio-on-dr2\"", Status: report.Passed},
		{
			Name:   "validate clusters data",
			Status: report.Failed,
			Err:    "Validation failed (92 ok, 0 warning, 1 problem)",
		},
	}
	checkItems(t, validate.Report.Steps[1], items)
	empty := &report.ClustersStatus{}
	if validate.Report.ClustersStatus.Equal(empty) {
		t.Fatal("clusters status is empty")
	}
	checkSummary(
		t,
		validate.Report,
		report.Summary{summary.OK: 92, summary.Problem: 1},
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
	checkError(t, validate.Report, "Canceled check S3 profiles")
	checkApplication(t, validate.Report, nil)
	checkNamespaces(t, validate.Report, testK8s.namespaces)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate clusters",
		Status: report.Canceled,
		Err:    "Canceled check S3 profiles",
	})

	// Check S3 is canceled, validation is skipped.
	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{
			Name:   "check S3 profile \"minio-on-dr1\"",
			Status: report.Canceled,
			Err:    "Canceled check S3 profile \"minio-on-dr1\"",
		},
		{
			Name:   "check S3 profile \"minio-on-dr2\"",
			Status: report.Canceled,
			Err:    "Canceled check S3 profile \"minio-on-dr2\"",
		},
	}
	checkItems(t, validate.Report.Steps[1], items)
	checkClusterStatus(t, validate.Report, &report.ClustersStatus{})
	checkSummary(t, validate.Report, report.Summary{})
}
