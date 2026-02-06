// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0
package gather

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"testing"

	e2econfig "github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/types"
	corev1 "k8s.io/api/core/v1"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/s3"
	"github.com/ramendr/ramenctl/pkg/sets"
	"github.com/ramendr/ramenctl/pkg/validation"
)

const (
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

	validateConfigFailed = &validation.Mock{
		ValidateFunc: func(ctx validation.Context) error {
			return errors.New("No validate for you!")
		},
	}

	validateConfigCanceled = &validation.Mock{
		ValidateFunc: func(ctx validation.Context) error {
			return context.Canceled
		},
	}

	inspectApplicationFailed = &validation.Mock{
		ApplicationNamespacesFunc: func(validation.Context, string, string) ([]string, error) {
			return nil, errors.New("No namespaces for you!")
		},
	}

	gatherClusterFailed = &validation.Mock{
		GatherFunc: func(
			ctx validation.Context,
			clusters []*types.Cluster,
			options gathering.Options,
		) <-chan gathering.Result {
			results := make(chan gathering.Result, 3)
			for _, cluster := range clusters {
				if cluster.Name == "hub" {
					results <- gathering.Result{Name: cluster.Name, Err: errors.New("no data for you!")}
				} else {
					results <- gathering.Result{Name: cluster.Name}
				}
			}
			close(results)
			return results
		},
	}

	inspectS3ProfilesCanceled = &validation.Mock{
		GetSecretFunc: func(ctx validation.Context, cluster *types.Cluster, name, namespace string) (*corev1.Secret, error) {
			return nil, context.Canceled
		},
	}

	gatherS3Failed = &validation.Mock{
		GatherS3Func: func(
			ctx validation.Context,
			profiles []*s3.Profile,
			prefixes []string,
			outputDir string,
		) <-chan s3.Result {
			results := make(chan s3.Result, 2)
			for i, profile := range profiles {
				if i == 0 {
					results <- s3.Result{ProfileName: profile.Name, Err: errors.New("no S3 data for you!")}
				} else {
					results <- s3.Result{ProfileName: profile.Name}
				}
			}
			close(results)
			return results
		},
	}

	gatherS3Canceled = &validation.Mock{
		GatherS3Func: func(
			ctx validation.Context,
			profiles []*s3.Profile,
			prefixes []string,
			outputDir string,
		) <-chan s3.Result {
			results := make(chan s3.Result, 2)
			for i, profile := range profiles {
				if i == 0 {
					results <- s3.Result{ProfileName: profile.Name, Err: context.Canceled}
				} else {
					results <- s3.Result{ProfileName: profile.Name}
				}
			}
			close(results)
			return results
		},
	}
)

func TestGatherApplicationPassed(t *testing.T) {
	cmd := testCommand(t, &validation.Mock{})
	helpers.AddGatheredData(t, cmd.dataDir(), "appset-deploy-rbd", "validate-application")
	if err := cmd.Application(drpcName, drpcNamespace); err != nil {
		t.Fatal(err)
	}
	checkReport(t, cmd.report, report.Passed)
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], "validate config", report.Passed)
	checkStep(t, cmd.report.Steps[1], "gather data", report.Passed)

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
	cmd := testCommand(t, validateConfigFailed)
	if err := cmd.Application(drpcName, drpcNamespace); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Failed)
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], "validate config", report.Failed)
}

func TestGatherApplicationValidateCanceled(t *testing.T) {
	cmd := testCommand(t, validateConfigCanceled)
	if err := cmd.Application(drpcName, drpcNamespace); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Canceled)
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], "validate config", report.Canceled)
}

func TestGatherApplicationInspectFailed(t *testing.T) {
	cmd := testCommand(t, inspectApplicationFailed)
	if err := cmd.Application(drpcName, drpcNamespace); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Failed)
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], "validate config", report.Passed)
	checkStep(t, cmd.report.Steps[1], "gather data", report.Failed)

	items := []*report.Step{
		{Name: "inspect application", Status: report.Failed},
	}
	checkItems(t, cmd.report.Steps[1], items)
}

func TestGatherApplicationGatherClusterFailed(t *testing.T) {
	cmd := testCommand(t, gatherClusterFailed)
	if err := cmd.Application(drpcName, drpcNamespace); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Failed)
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], "validate config", report.Passed)
	checkStep(t, cmd.report.Steps[1], "gather data", report.Failed)

	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
		{Name: "gather \"hub\"", Status: report.Failed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
	}
	checkItems(t, cmd.report.Steps[1], items)
}

func TestGatherApplicationNamespaces(t *testing.T) {
	mockBackend := &validation.Mock{
		ApplicationNamespacesFunc: func(ctx validation.Context, name, namespace string) ([]string, error) {
			if name != drpcName || namespace != drpcNamespace {
				t.Fatalf("unexpected args: name=%s, namespace=%s", drpcName, drpcNamespace)
			}
			return applicationNamespaces, nil
		},
	}

	cmd := testCommand(t, mockBackend)
	helpers.AddGatheredData(t, cmd.dataDir(), "appset-deploy-rbd", "validate-application")
	err := cmd.Application(drpcName, drpcNamespace)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !slices.Equal(cmd.report.Namespaces, gatherApplicationNamespaces) {
		diff := helpers.UnifiedDiff(t, gatherApplicationNamespaces, cmd.report.Namespaces)
		t.Fatalf("namespaces not equal\n%s", diff)
	}
}

func TestGatherApplicationInspectS3ProfilesFailed(t *testing.T) {
	cmd := testCommand(t, &validation.Mock{})
	if err := cmd.Application(drpcName, drpcNamespace); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Failed)
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], "validate config", report.Passed)
	checkStep(t, cmd.report.Steps[1], "gather data", report.Failed)

	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Failed},
	}
	checkItems(t, cmd.report.Steps[1], items)
}

func TestGatherApplicationInspectS3ProfilesCanceled(t *testing.T) {
	cmd := testCommand(t, inspectS3ProfilesCanceled)
	helpers.AddGatheredData(t, cmd.dataDir(), "appset-deploy-rbd", "validate-application")
	if err := cmd.Application(drpcName, drpcNamespace); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Canceled)
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], "validate config", report.Passed)
	checkStep(t, cmd.report.Steps[1], "gather data", report.Canceled)

	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Canceled},
	}
	checkItems(t, cmd.report.Steps[1], items)
}

func TestGatherApplicationS3DataFailed(t *testing.T) {
	cmd := testCommand(t, gatherS3Failed)
	helpers.AddGatheredData(t, cmd.dataDir(), "appset-deploy-rbd", "validate-application")
	if err := cmd.Application(drpcName, drpcNamespace); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Failed)
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], "validate config", report.Passed)
	checkStep(t, cmd.report.Steps[1], "gather data", report.Failed)

	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{Name: "gather S3 profile \"minio-on-dr1\"", Status: report.Failed},
		{Name: "gather S3 profile \"minio-on-dr2\"", Status: report.Passed},
	}
	checkItems(t, cmd.report.Steps[1], items)
}

func TestGatherApplicationS3DataCanceled(t *testing.T) {
	cmd := testCommand(t, gatherS3Canceled)
	helpers.AddGatheredData(t, cmd.dataDir(), "appset-deploy-rbd", "validate-application")
	if err := cmd.Application(drpcName, drpcNamespace); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, cmd.report, report.Canceled)
	checkApplication(t, cmd.report, testApplication)

	if len(cmd.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", cmd.report.Steps)
	}
	checkStep(t, cmd.report.Steps[0], "validate config", report.Passed)
	checkStep(t, cmd.report.Steps[1], "gather data", report.Canceled)

	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"dr1\"", Status: report.Passed},
		{Name: "gather \"dr2\"", Status: report.Passed},
		{Name: "inspect S3 profiles", Status: report.Passed},
		{Name: "gather S3 profile \"minio-on-dr1\"", Status: report.Canceled},
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
	return newCommand(cmd, testConfig, backend)
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

func checkStep(t *testing.T, step *report.Step, name string, status report.Status) {
	if name != step.Name {
		t.Fatalf("expected step %q, got %q", name, step.Name)
	}
	if status != step.Status {
		t.Fatalf("expected status %q, got %q", status, step.Status)
	}
}

func checkItems(t *testing.T, step *report.Step, expected []*report.Step) {
	if len(expected) != len(step.Items) {
		t.Fatalf("expected %d items, got %d", len(expected), len(step.Items))
	}
	for i, item := range expected {
		checkStep(t, step.Items[i], item.Name, item.Status)
	}
}

func totalDuration(steps []*report.Step) float64 {
	var total float64
	for _, step := range steps {
		total += step.Duration
	}
	return total
}
