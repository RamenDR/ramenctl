// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"context"
	"testing"

	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/sets"
	"github.com/ramendr/ramenctl/pkg/validate/summary"
	"github.com/ramendr/ramenctl/pkg/validation"
)

const applicationTestdata = "../../testdata/appset-deploy-rbd"

var (
	testApplication = &report.Application{
		Name:      drpcName,
		Namespace: drpcNamespace,
	}

	applicationNamespaces = sets.Sorted([]string{
		drpcNamespace,
		applicationNamespace,
	})

	reportNamespaces = sets.Sorted([]string{
		testK8s.config.Namespaces.RamenHubNamespace,
		testK8s.config.Namespaces.RamenDRClusterNamespace,
		drpcNamespace,
		applicationNamespace,
	})

	// Application mock instances.

	applicationMock = &helpers.ValidationMock{
		ApplicationNamespacesFunc: func(validation.Context, string, string) ([]string, error) {
			return applicationNamespaces, nil
		},
	}

	inspectApplicationCanceled = &helpers.ValidationMock{
		ApplicationNamespacesFunc: func(validation.Context, string, string) ([]string, error) {
			return nil, context.Canceled
		},
	}

	gatherDataFailed = &helpers.ValidationMock{
		ApplicationNamespacesFunc: applicationMock.ApplicationNamespaces,
		GatherFunc:                helpers.GatherDataFailed,
	}

	inspectS3ProfilesCanceled = &helpers.ValidationMock{
		ApplicationNamespacesFunc: applicationMock.ApplicationNamespaces,
		GetSecretFunc:             helpers.GetSecretCanceled.GetSecret,
	}

	getSecretFailed = &helpers.ValidationMock{
		ApplicationNamespacesFunc: applicationMock.ApplicationNamespaces,
		GetSecretFunc:             helpers.GetSecretFailed.GetSecret,
	}

	getSecretInvalid = &helpers.ValidationMock{
		ApplicationNamespacesFunc: applicationMock.ApplicationNamespaces,
		GetSecretFunc:             helpers.GetSecretInvalid.GetSecret,
	}

	gatherS3Failed = &helpers.ValidationMock{
		ApplicationNamespacesFunc: applicationMock.ApplicationNamespaces,
		GatherS3Func:              helpers.GatherS3DataFailed,
	}

	gatherS3Canceled = &helpers.ValidationMock{
		ApplicationNamespacesFunc: applicationMock.ApplicationNamespaces,
		GatherS3Func:              helpers.GatherS3DataCanceled,
	}
)

// Validate application tests.

func TestValidateApplicationPassed(t *testing.T) {
	validate := testCommand(t, applicationMock, testK8s)
	helpers.AddGatheredData(t, validate.DataDir(), applicationTestdata, validate.Report.Name)
	if err := validate.Run(); err != nil {
		dumpCommandLog(t, validate)
		t.Fatal(err)
	}
	checkReport(t, validate, report.Passed)
	checkError(t, validate.Report, "")
	checkApplication(t, validate.Report, testApplication)
	checkNamespaces(t, validate.Report, reportNamespaces)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate application",
		Status: report.Passed,
	})

	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{Name: "gather S3 profile \"minio-on-dr1\"", Status: report.Passed},
		{Name: "gather S3 profile \"minio-on-dr2\"", Status: report.Passed},
		{Name: "validate data", Status: report.Passed},
	}
	checkItems(t, validate.Report.Steps[1], items)

	expectedStatus := loadApplicationStatus(t, "appset-deploy-rbd.yaml")

	checkApplicationStatus(t, validate.Report, expectedStatus)

	checkSummary(t, validate.Report, report.Summary{summary.OK: 30})
}

func TestValidateApplicationValidateFailed(t *testing.T) {
	validate := testCommand(t, helpers.ValidateConfigFailed, testK8s)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkError(t, validate.Report, "Failed to validate config")
	checkApplication(t, validate.Report, testApplication)
	checkNamespaces(t, validate.Report, nil)
	if len(validate.Report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Failed,
		Err:    "Failed to validate config",
	})
	checkApplicationStatus(t, validate.Report, &report.ApplicationStatus{})
	checkSummary(t, validate.Report, report.Summary{})
}

func TestValidateApplicationValidateCanceled(t *testing.T) {
	validate := testCommand(t, helpers.ValidateConfigCanceled, testK8s)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Canceled)
	checkError(t, validate.Report, "Canceled validate config")
	checkApplication(t, validate.Report, testApplication)
	checkNamespaces(t, validate.Report, nil)
	if len(validate.Report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Canceled,
		Err:    "Canceled validate config",
	})
	checkApplicationStatus(t, validate.Report, &report.ApplicationStatus{})
	checkSummary(t, validate.Report, report.Summary{})
}

func TestValidateApplicationInspectApplicationFailed(t *testing.T) {
	validate := testCommand(t, helpers.InspectApplicationFailed, testK8s)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkError(t, validate.Report,
		`Failed to inspect application "appset-deploy-rbd" in namespace "argocd"`)
	checkApplication(t, validate.Report, testApplication)
	checkNamespaces(t, validate.Report, nil)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate application",
		Status: report.Failed,
	})

	// If inspecting the application has failed we skip the gather step.
	items := []*report.Step{
		{
			Name:   "inspect application",
			Status: report.Failed,
			Err:    `Failed to inspect application "appset-deploy-rbd" in namespace "argocd"`,
		},
	}
	checkItems(t, validate.Report.Steps[1], items)
	checkApplicationStatus(t, validate.Report, &report.ApplicationStatus{})
	checkSummary(t, validate.Report, report.Summary{})
}

func TestValidateApplicationInspectApplicationCanceled(t *testing.T) {
	validate := testCommand(t, inspectApplicationCanceled, testK8s)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Canceled)
	checkError(t, validate.Report, "Canceled inspect application")
	checkApplication(t, validate.Report, testApplication)
	checkNamespaces(t, validate.Report, nil)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate application",
		Status: report.Canceled,
	})

	// If inspecting the application has been canceled we skip the gather step.
	items := []*report.Step{
		{
			Name:   "inspect application",
			Status: report.Canceled,
			Err:    "Canceled inspect application",
		},
	}
	checkItems(t, validate.Report.Steps[1], items)
	checkApplicationStatus(t, validate.Report, &report.ApplicationStatus{})
	checkSummary(t, validate.Report, report.Summary{})
}

func TestValidateApplicationGatherClusterFailed(t *testing.T) {
	validate := testCommand(t, gatherDataFailed, testK8s)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkError(t, validate.Report, "Failed to gather data from clusters hub")
	checkApplication(t, validate.Report, testApplication)
	checkNamespaces(t, validate.Report, reportNamespaces)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate application",
		Status: report.Failed,
		Err:    "Failed to gather data from clusters hub",
	})

	// If gathering data fail for some of the clusters, we skip the validation step.
	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
		{
			Name:   "gather \"hub\"",
			Status: report.Failed,
			Err:    `Failed to gather data from cluster "hub"`,
		},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
	}
	checkItems(t, validate.Report.Steps[1], items)
	checkApplicationStatus(t, validate.Report, &report.ApplicationStatus{})
	checkSummary(t, validate.Report, report.Summary{})
}

func TestValidateApplicationInspectS3ProfilesFailed(t *testing.T) {
	validate := testCommand(t, applicationMock, testK8s)
	// We don't add test data to cause inspect application s3 to fail.
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkError(t, validate.Report, "Failed to validate hub")
	checkApplication(t, validate.Report, testApplication)
	checkNamespaces(t, validate.Report, reportNamespaces)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate application",
		Status: report.Failed,
	})

	// Inspect S3 profiles fails, S3 gathering is skipped.
	// Validation runs and reports missing S3 data as problem.
	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{
			Name:   "inspect S3 profiles",
			Status: report.Failed,
			Err:    "Failed to read S3 profiles from hub",
		},
		{
			Name:   "validate data",
			Status: report.Failed,
			Err:    "Failed to validate hub",
		},
	}
	checkItems(t, validate.Report.Steps[1], items)
	checkSummary(t, validate.Report, report.Summary{})
}

func TestValidateApplicationInspectS3ProfilesCanceled(t *testing.T) {
	validate := testCommand(t, inspectS3ProfilesCanceled, testK8s)
	helpers.AddGatheredData(t, validate.DataDir(), applicationTestdata, validate.Report.Name)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Canceled)
	checkError(t, validate.Report, "Canceled inspect S3 profiles")
	checkApplication(t, validate.Report, testApplication)
	checkNamespaces(t, validate.Report, reportNamespaces)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate application",
		Status: report.Canceled,
	})

	// Inspect S3 profiles is canceled, gatherS3 and validation are skipped.
	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
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
	checkSummary(t, validate.Report, report.Summary{})
}

func TestValidateApplicationGetSecretFailed(t *testing.T) {
	validate := testCommand(t, getSecretFailed, testK8s)
	helpers.AddGatheredData(t, validate.DataDir(), applicationTestdata, validate.Report.Name)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkError(t, validate.Report,
		"Failed to gather S3 profiles minio-on-dr1, minio-on-dr2")
	checkApplication(t, validate.Report, testApplication)
	checkNamespaces(t, validate.Report, reportNamespaces)

	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate application",
		Status: report.Failed,
		Err:    "Failed to gather S3 profiles minio-on-dr1, minio-on-dr2",
	})

	// When GetSecret returns an error. The profile will have empty credentials
	// causing S3 gather and validation to fail.
	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{
			Name:   "gather S3 profile \"minio-on-dr1\"",
			Status: report.Failed,
			Err:    `Failed to gather S3 profile "minio-on-dr1"`,
		},
		{
			Name:   "gather S3 profile \"minio-on-dr2\"",
			Status: report.Failed,
			Err:    `Failed to gather S3 profile "minio-on-dr2"`,
		},
		{
			Name:   "validate data",
			Status: report.Failed,
			Err:    "Validation failed (28 ok, 0 warning, 2 problem)",
		},
	}
	checkItems(t, validate.Report.Steps[1], items)
	checkSummary(t, validate.Report, report.Summary{summary.OK: 28, summary.Problem: 2})
}

func TestValidateApplicationGetSecretInvalid(t *testing.T) {
	validate := testCommand(t, getSecretInvalid, testK8s)
	helpers.AddGatheredData(t, validate.DataDir(), applicationTestdata, validate.Report.Name)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkError(t, validate.Report,
		"Failed to gather S3 profiles minio-on-dr1, minio-on-dr2")
	checkApplication(t, validate.Report, testApplication)
	checkNamespaces(t, validate.Report, reportNamespaces)

	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate application",
		Status: report.Failed,
		Err:    "Failed to gather S3 profiles minio-on-dr1, minio-on-dr2",
	})

	// When GetSecret returns a secret with invalid value, causing S3 gather and
	// validation to fail.
	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{
			Name:   "gather S3 profile \"minio-on-dr1\"",
			Status: report.Failed,
			Err:    `Failed to gather S3 profile "minio-on-dr1"`,
		},
		{
			Name:   "gather S3 profile \"minio-on-dr2\"",
			Status: report.Failed,
			Err:    `Failed to gather S3 profile "minio-on-dr2"`,
		},
		{
			Name:   "validate data",
			Status: report.Failed,
			Err:    "Validation failed (28 ok, 0 warning, 2 problem)",
		},
	}
	checkItems(t, validate.Report.Steps[1], items)
	checkSummary(t, validate.Report, report.Summary{summary.OK: 28, summary.Problem: 2})
}

func TestValidateApplicationGatherS3Failed(t *testing.T) {
	validate := testCommand(t, gatherS3Failed, testK8s)
	helpers.AddGatheredData(t, validate.DataDir(), applicationTestdata, validate.Report.Name)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Failed)
	checkError(t, validate.Report, "Failed to gather S3 profiles minio-on-dr1")
	checkApplication(t, validate.Report, testApplication)
	checkNamespaces(t, validate.Report, reportNamespaces)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate application",
		Status: report.Failed,
		Err:    "Failed to gather S3 profiles minio-on-dr1",
	})

	// S3 gather fails for one profile, other profile succeeds.
	// Validation runs and reports the failed profile as problem.
	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{
			Name:   "gather S3 profile \"minio-on-dr1\"",
			Status: report.Failed,
			Err:    `Failed to gather S3 profile "minio-on-dr1"`,
		},
		{Name: "gather S3 profile \"minio-on-dr2\"", Status: report.Passed},
		{
			Name:   "validate data",
			Status: report.Failed,
			Err:    "Validation failed (29 ok, 0 warning, 1 problem)",
		},
	}
	checkItems(t, validate.Report.Steps[1], items)
	checkSummary(
		t,
		validate.Report,
		report.Summary{summary.OK: 29, summary.Problem: 1},
	)
}

func TestValidateApplicationGatherS3Canceled(t *testing.T) {
	validate := testCommand(t, gatherS3Canceled, testK8s)
	helpers.AddGatheredData(t, validate.DataDir(), applicationTestdata, validate.Report.Name)
	if err := validate.Run(); err == nil {
		dumpCommandLog(t, validate)
		t.Fatal("command did not fail")
	}
	checkReport(t, validate, report.Canceled)
	checkError(t, validate.Report, "Canceled gather S3 profiles")
	checkApplication(t, validate.Report, testApplication)
	checkNamespaces(t, validate.Report, reportNamespaces)
	if len(validate.Report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.Report.Steps)
	}
	checkStep(t, validate.Report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Passed,
	})
	checkStep(t, validate.Report.Steps[1], &report.Step{
		Name:   "validate application",
		Status: report.Canceled,
		Err:    "Canceled gather S3 profiles",
	})

	// S3 gather is canceled, validation is skipped.
	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{
			Name:   "gather S3 profile \"minio-on-dr1\"",
			Status: report.Canceled,
			Err:    "Canceled gather S3 profile \"minio-on-dr1\"",
		},
		{Name: "gather S3 profile \"minio-on-dr2\"", Status: report.Passed},
	}
	checkItems(t, validate.Report.Steps[1], items)
	checkSummary(t, validate.Report, report.Summary{})
}
