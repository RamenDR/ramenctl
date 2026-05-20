// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0
package gather

import (
	"reflect"
	"slices"
	"testing"

	e2econfig "github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/sets"
	"github.com/ramendr/ramenctl/pkg/validation"
)

const (
	applicationTestdata  = "../testdata/appset-deploy-rbd"
	drpcName             = "appset-deploy-rbd"
	drpcNamespace        = "argocd"
	applicationNamespace = "e2e-appset-deploy-rbd"
)

var (
	testConfig = &config.Config{
		Namespaces: e2econfig.K8sNamespaces,
	}

	testEnv = &types.Env{
		Hub: &types.Cluster{Name: "hub"},
		C1:  &types.Cluster{Name: "dr1"},
		C2:  &types.Cluster{Name: "dr2"},
	}

	testApplication = &report.Application{
		Name:      drpcName,
		Namespace: drpcNamespace,
	}

	applicationNamespaces = sets.Sorted([]string{
		drpcNamespace,
		applicationNamespace,
	})

	gatherApplicationNamespaces = sets.Sorted([]string{
		testConfig.Namespaces.RamenHubNamespace,
		testConfig.Namespaces.RamenDRClusterNamespace,
		drpcNamespace,
		applicationNamespace,
	})

	// Mock instances composing shared mock functions and helpers.

	gatherClusterFailed = &helpers.ValidationMock{
		GatherFunc: helpers.GatherDataFailed,
	}

	gatherS3Failed = &helpers.ValidationMock{
		GatherS3Func: helpers.GatherS3DataFailed,
	}

	gatherS3Canceled = &helpers.ValidationMock{
		GatherS3Func: helpers.GatherS3DataCanceled,
	}
)

func TestGatherApplicationPassed(t *testing.T) {
	cmd := testCommand(t, &helpers.ValidationMock{})
	helpers.AddGatheredData(t, cmd.dataDir(), applicationTestdata, "validate-application")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	checkReport(t, cmd.report, report.Passed)
	checkError(t, cmd.report, "")
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], &report.Step{Name: "validate config", Status: report.Passed})
	checkStep(t, cmd.report.Steps[1], &report.Step{Name: "gather data", Status: report.Passed})

	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{Name: "gather S3 profile \"minio-on-dr1\"", Status: report.Passed},
		{Name: "gather S3 profile \"minio-on-dr2\"", Status: report.Passed},
	}
	checkItems(t, cmd.report.Steps[1], items)
}

func TestGatherApplicationValidateFailed(t *testing.T) {
	cmd := testCommand(t, helpers.ValidateConfigFailed)
	if err := cmd.Run(); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Failed)
	checkError(t, cmd.report, "Failed to validate config")
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Failed,
		Err:    "Failed to validate config",
	})
}

func TestGatherApplicationValidateCanceled(t *testing.T) {
	cmd := testCommand(t, helpers.ValidateConfigCanceled)
	if err := cmd.Run(); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Canceled)
	checkError(t, cmd.report, "Canceled validate config")
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], &report.Step{
		Name:   "validate config",
		Status: report.Canceled,
		Err:    "Canceled validate config",
	})
}

func TestGatherApplicationInspectFailed(t *testing.T) {
	cmd := testCommand(t, helpers.InspectApplicationFailed)
	if err := cmd.Run(); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Failed)
	checkError(t, cmd.report,
		`Failed to inspect application "appset-deploy-rbd" in namespace "argocd"`)
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], &report.Step{Name: "validate config", Status: report.Passed})
	checkStep(t, cmd.report.Steps[1], &report.Step{Name: "gather data", Status: report.Failed})

	items := []*report.Step{
		{
			Name:   "inspect application",
			Status: report.Failed,
			Err:    `Failed to inspect application "appset-deploy-rbd" in namespace "argocd"`,
		},
	}
	checkItems(t, cmd.report.Steps[1], items)
}

func TestGatherApplicationGatherClusterFailed(t *testing.T) {
	cmd := testCommand(t, gatherClusterFailed)
	if err := cmd.Run(); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Failed)
	checkError(t, cmd.report, "Failed to gather data from clusters hub")
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], &report.Step{Name: "validate config", Status: report.Passed})
	checkStep(t, cmd.report.Steps[1], &report.Step{
		Name:   "gather data",
		Status: report.Failed,
		Err:    "Failed to gather data from clusters hub",
	})

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
	checkItems(t, cmd.report.Steps[1], items)
}

func TestGatherApplicationNamespaces(t *testing.T) {
	mockBackend := &helpers.ValidationMock{
		ApplicationNamespacesFunc: func(ctx validation.Context, name, namespace string) ([]string, error) {
			if name != drpcName || namespace != drpcNamespace {
				t.Fatalf("unexpected args: name=%s, namespace=%s", drpcName, drpcNamespace)
			}
			return applicationNamespaces, nil
		},
	}

	cmd := testCommand(t, mockBackend)
	helpers.AddGatheredData(t, cmd.dataDir(), applicationTestdata, "validate-application")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !slices.Equal(cmd.report.Namespaces, gatherApplicationNamespaces) {
		diff := helpers.UnifiedDiff(t, gatherApplicationNamespaces, cmd.report.Namespaces)
		t.Fatalf("namespaces not equal\n%s", diff)
	}
}

func TestGatherApplicationInspectS3ProfilesFailed(t *testing.T) {
	cmd := testCommand(t, &helpers.ValidationMock{})
	if err := cmd.Run(); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Failed)
	checkError(t, cmd.report, "Failed to read S3 profiles from hub")
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], &report.Step{Name: "validate config", Status: report.Passed})
	checkStep(t, cmd.report.Steps[1], &report.Step{Name: "gather data", Status: report.Failed})

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
	}
	checkItems(t, cmd.report.Steps[1], items)
}

func TestGatherApplicationInspectS3ProfilesCanceled(t *testing.T) {
	cmd := testCommand(t, helpers.GetSecretCanceled)
	helpers.AddGatheredData(t, cmd.dataDir(), applicationTestdata, "validate-application")
	if err := cmd.Run(); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Canceled)
	checkError(t, cmd.report, "Canceled inspect S3 profiles")
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], &report.Step{Name: "validate config", Status: report.Passed})
	checkStep(t, cmd.report.Steps[1], &report.Step{Name: "gather data", Status: report.Canceled})

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
	checkItems(t, cmd.report.Steps[1], items)
}

func TestGatherApplicationGetSecretFailed(t *testing.T) {
	cmd := testCommand(t, helpers.GetSecretFailed)
	helpers.AddGatheredData(t, cmd.dataDir(), applicationTestdata, "validate-application")
	if err := cmd.Run(); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Failed)
	checkError(t, cmd.report,
		"Failed to gather S3 profiles minio-on-dr1, minio-on-dr2")
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], &report.Step{Name: "validate config", Status: report.Passed})
	checkStep(t, cmd.report.Steps[1], &report.Step{
		Name:   "gather data",
		Status: report.Failed,
		Err:    "Failed to gather S3 profiles minio-on-dr1, minio-on-dr2",
	})

	// When GetSecret returns an error. The profile will have empty credentials
	// causing S3 gather to fail.
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
	}
	checkItems(t, cmd.report.Steps[1], items)
}

func TestGatherApplicationGetSecretInvalid(t *testing.T) {
	cmd := testCommand(t, helpers.GetSecretInvalid)
	helpers.AddGatheredData(t, cmd.dataDir(), applicationTestdata, "validate-application")
	if err := cmd.Run(); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Failed)
	checkError(t, cmd.report,
		"Failed to gather S3 profiles minio-on-dr1, minio-on-dr2")
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], &report.Step{Name: "validate config", Status: report.Passed})
	checkStep(t, cmd.report.Steps[1], &report.Step{
		Name:   "gather data",
		Status: report.Failed,
		Err:    "Failed to gather S3 profiles minio-on-dr1, minio-on-dr2",
	})

	// When GetSecret returns a secret with invalid value, causing S3 gather to fail.
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
	}
	checkItems(t, cmd.report.Steps[1], items)
}

func TestGatherApplicationS3DataFailed(t *testing.T) {
	cmd := testCommand(t, gatherS3Failed)
	helpers.AddGatheredData(t, cmd.dataDir(), applicationTestdata, "validate-application")
	if err := cmd.Run(); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Failed)
	checkError(t, cmd.report, "Failed to gather S3 profiles minio-on-dr1")
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], &report.Step{Name: "validate config", Status: report.Passed})
	checkStep(t, cmd.report.Steps[1], &report.Step{
		Name:   "gather data",
		Status: report.Failed,
		Err:    "Failed to gather S3 profiles minio-on-dr1",
	})

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
	}
	checkItems(t, cmd.report.Steps[1], items)
}

func TestGatherApplicationS3DataCanceled(t *testing.T) {
	cmd := testCommand(t, gatherS3Canceled)
	helpers.AddGatheredData(t, cmd.dataDir(), applicationTestdata, "validate-application")
	if err := cmd.Run(); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Canceled)
	checkError(t, cmd.report, "Canceled gather S3 profiles")
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], &report.Step{Name: "validate config", Status: report.Passed})
	checkStep(t, cmd.report.Steps[1], &report.Step{
		Name:   "gather data",
		Status: report.Canceled,
		Err:    "Canceled gather S3 profiles",
	})

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
	checkItems(t, cmd.report.Steps[1], items)
}

// Helpers

func testCommand(t *testing.T, backend validation.Validation) *Command {
	cmd, err := command.ForTest("gather-application", testEnv, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cmd.Close()
	})
	opts := command.ApplicationOptions{
		DRPCName:      drpcName,
		DRPCNamespace: drpcNamespace,
	}
	return newCommand(cmd, testConfig, backend, opts)
}

func checkReport(t *testing.T, report *report.Report, status report.Status) {
	if report.Status != status {
		t.Fatalf("expected status %q, got %q", status, report.Status)
	}
	if !report.Config.Equal(testConfig) {
		t.Fatalf("expected config %q, got %q", testConfig, report.Config)
	}
	duration := totalDuration(report.Steps)
	if report.Duration != duration {
		t.Fatalf("expected duration %v, got %v", duration, report.Duration)
	}
}

func checkApplication(t *testing.T, report *report.Report, expected *report.Application) {
	if !reflect.DeepEqual(expected, report.Application) {
		diff := helpers.UnifiedDiff(t, expected, report.Application)
		t.Fatalf("applications are not equal\n%s", diff)
	}
}

// We cannot check duration since it may be zero on windows.
func checkStep(t *testing.T, got *report.Step, expected *report.Step) {
	if got.Name != expected.Name {
		t.Fatalf("expected step %q, got %q", expected.Name, got.Name)
	}
	if got.Status != expected.Status {
		t.Fatalf("expected step %q status %q, got %q", expected.Name, expected.Status, got.Status)
	}
	if got.Err != expected.Err {
		t.Fatalf("expected step %q error %q, got %q", expected.Name, expected.Err, got.Err)
	}
}

func checkError(t *testing.T, r *report.Report, expected string) {
	if got := r.Error(); got != expected {
		t.Fatalf("expected error %q, got %q", expected, got)
	}
}

func checkItems(t *testing.T, step *report.Step, expected []*report.Step) {
	if len(expected) != len(step.Items) {
		t.Fatalf("expected %d items, got %d", len(expected), len(step.Items))
	}
	for i, item := range expected {
		checkStep(t, step.Items[i], item)
	}
}

func totalDuration(steps []*report.Step) float64 {
	var total float64
	for _, step := range steps {
		total += step.Duration
	}
	return total
}
