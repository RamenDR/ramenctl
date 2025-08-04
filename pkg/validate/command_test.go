// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"context"
	"errors"
	"slices"
	"testing"

	e2econfig "github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/sets"
	"github.com/ramendr/ramenctl/pkg/validation"
)

const (
	validateClusters     = "validate-clusters"
	validateApplication  = "validate-application"
	drpcName             = "drpc-name"
	drpcNamespace        = "drpc-namespace"
	applicationNamespace = "application-namespace"
)

var (
	testConfig = &config.Config{
		Namespaces: e2econfig.K8sNamespaces,
	}

	testEnv = &types.Env{
		Hub: &types.Cluster{Name: "hub"},
		C1:  &types.Cluster{Name: "c1"},
		C2:  &types.Cluster{Name: "c2"},
	}

	testApplication = &report.Application{
		Name:      drpcName,
		Namespace: drpcNamespace,
	}

	applicationNamespaces = sets.Sorted([]string{
		drpcNamespace,
		applicationNamespace,
	})

	validateApplicationNamespaces = sets.Sorted([]string{
		drpcNamespace,
		applicationNamespace,
	})

	validateClustersNamespaces = sets.Sorted([]string{
		testConfig.Namespaces.RamenHubNamespace,
		testConfig.Namespaces.RamenDRClusterNamespace,
	})

	applicationMock = &validation.Mock{
		ApplicationNamespacesFunc: func(validation.Context, string, string) ([]string, error) {
			return applicationNamespaces, nil
		},
	}

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

	inspectApplicationCanceled = &validation.Mock{
		ApplicationNamespacesFunc: func(validation.Context, string, string) ([]string, error) {
			return nil, context.Canceled
		},
	}

	gatherClusterFailed = &validation.Mock{
		ApplicationNamespacesFunc: applicationMock.ApplicationNamespaces,
		GatherFunc: func(
			ctx validation.Context,
			clusters []*types.Cluster,
			namespaces []string,
			outputDir string,
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
)

// Validate clusters tests.

func TestValidateClustersPassed(t *testing.T) {
	validate := testCommand(t, validateClusters, &validation.Mock{})
	if err := validate.Clusters(); err != nil {
		t.Fatal(err)
	}
	checkReport(t, validate.report, report.Passed)
	checkApplication(t, validate.report, nil)
	checkNamespaces(t, validate.report, validateClustersNamespaces)
	if len(validate.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.report.Steps[1], "validate clusters", report.Passed)

	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"c1\"", Status: report.Passed},
		{Name: "gather \"c2\"", Status: report.Passed},
		{Name: "validate cluster data", Status: report.Passed},
	}
	checkItems(t, validate.report.Steps[1], items)
}

func TestValidateClustersValidateFailed(t *testing.T) {
	validate := testCommand(t, validateClusters, validateConfigFailed)
	if err := validate.Clusters(); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, validate.report, report.Failed)
	checkApplication(t, validate.report, nil)
	checkNamespaces(t, validate.report, nil)
	if len(validate.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Failed)
}

func TestValidateClustersValidateCanceled(t *testing.T) {
	validate := testCommand(t, validateClusters, validateConfigCanceled)
	if err := validate.Clusters(); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, validate.report, report.Canceled)
	checkApplication(t, validate.report, nil)
	checkNamespaces(t, validate.report, nil)
	if len(validate.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Canceled)
}

func TestValidateClusterGatherClusterFailed(t *testing.T) {
	validate := testCommand(t, validateClusters, gatherClusterFailed)
	if err := validate.Clusters(); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, validate.report, report.Failed)
	checkApplication(t, validate.report, nil)
	checkNamespaces(t, validate.report, validateClustersNamespaces)
	if len(validate.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.report.Steps[1], "validate clusters", report.Failed)

	// If gathering data fail for some of the clusters, we skip the validation step.
	items := []*report.Step{
		{Name: "gather \"hub\"", Status: report.Failed},
		{Name: "gather \"c1\"", Status: report.Passed},
		{Name: "gather \"c2\"", Status: report.Passed},
	}
	checkItems(t, validate.report.Steps[1], items)
}

// Validate application tests.

func TestValidateApplicationPassed(t *testing.T) {
	validate := testCommand(t, validateApplication, applicationMock)
	if err := validate.Application(drpcName, drpcNamespace); err != nil {
		t.Fatal(err)
	}
	checkReport(t, validate.report, report.Passed)
	checkApplication(t, validate.report, testApplication)
	checkNamespaces(t, validate.report, validateApplicationNamespaces)
	if len(validate.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.report.Steps[1], "validate application", report.Passed)

	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
		{Name: "gather \"hub\"", Status: report.Passed},
		{Name: "gather \"c1\"", Status: report.Passed},
		{Name: "gather \"c2\"", Status: report.Passed},
		{Name: "validate data", Status: report.Passed},
	}
	checkItems(t, validate.report.Steps[1], items)
}

func TestValidateApplicationValidateFailed(t *testing.T) {
	validate := testCommand(t, validateApplication, validateConfigFailed)
	if err := validate.Application(drpcName, drpcNamespace); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, validate.report, report.Failed)
	checkApplication(t, validate.report, testApplication)
	checkNamespaces(t, validate.report, nil)
	if len(validate.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Failed)
}

func TestValidateApplicationValidateCanceled(t *testing.T) {
	validate := testCommand(t, validateApplication, validateConfigCanceled)
	if err := validate.Application(drpcName, drpcNamespace); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, validate.report, report.Canceled)
	checkApplication(t, validate.report, testApplication)
	checkNamespaces(t, validate.report, nil)
	if len(validate.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Canceled)
}

func TestValidateApplicationInspectApplicationFailed(t *testing.T) {
	validate := testCommand(t, validateApplication, inspectApplicationFailed)
	if err := validate.Application(drpcName, drpcNamespace); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, validate.report, report.Failed)
	checkApplication(t, validate.report, testApplication)
	checkNamespaces(t, validate.report, nil)
	if len(validate.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.report.Steps[1], "validate application", report.Failed)

	// If inspecting the application has failed we skip the gather step.
	items := []*report.Step{
		{Name: "inspect application", Status: report.Failed},
	}
	checkItems(t, validate.report.Steps[1], items)
}

func TestValidateApplicationInspectApplicationCanceled(t *testing.T) {
	validate := testCommand(t, validateApplication, inspectApplicationCanceled)
	if err := validate.Application(drpcName, drpcNamespace); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, validate.report, report.Canceled)
	checkApplication(t, validate.report, testApplication)
	checkNamespaces(t, validate.report, nil)
	if len(validate.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.report.Steps[1], "validate application", report.Canceled)

	// If inspecting the application has been canceled we skip the gather step.
	items := []*report.Step{
		{Name: "inspect application", Status: report.Canceled},
	}
	checkItems(t, validate.report.Steps[1], items)
}

func TestValidateApplicationGatherClusterFailed(t *testing.T) {
	validate := testCommand(t, validateApplication, gatherClusterFailed)
	if err := validate.Application(drpcName, drpcNamespace); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, validate.report, report.Failed)
	checkApplication(t, validate.report, testApplication)
	checkNamespaces(t, validate.report, validateApplicationNamespaces)
	if len(validate.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.report.Steps[1], "validate application", report.Failed)

	// If gathering data fail for some of the clusters, we skip the validation step.
	items := []*report.Step{
		{Name: "inspect application", Status: report.Passed},
		{Name: "gather \"hub\"", Status: report.Failed},
		{Name: "gather \"c1\"", Status: report.Passed},
		{Name: "gather \"c2\"", Status: report.Passed},
	}
	checkItems(t, validate.report.Steps[1], items)
}

// TODO: Test gather cancellation when kubectl-gahter supports it:
// https://github.com/nirs/kubectl-gather/issues/88

// Helpers.

func testCommand(t *testing.T, name string, backend validation.Validation) *Command {
	cmd, err := command.ForTest(name, testEnv, t.TempDir())
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
	if report.Application != nil && expected != nil {
		if *report.Application != *expected {
			t.Fatalf("expected application %+v, got %+v", expected, report.Application)
		}
	} else if report.Application != expected {
		t.Fatalf("expected application %+v, got %+v", expected, report.Application)
	}
}

func checkNamespaces(t *testing.T, report *report.Report, expected []string) {
	if !slices.Equal(report.Namespaces, expected) {
		t.Fatalf("expected namespaces %q, got %q", expected, report.Namespaces)
	}
}

func checkStep(t *testing.T, step *report.Step, name string, status report.Status) {
	if name != step.Name {
		t.Fatalf("expected step %q, got %q", name, step.Name)
	}
	if status != step.Status {
		t.Fatalf("expected status %q, got %q", status, step.Status)
	}
	// We cannot check duration since it may be zero on windows.
}

func checkItems(t *testing.T, step *report.Step, expected []*report.Step) {
	if len(expected) != len(step.Items) {
		t.Fatalf("expected items %+v, got %+v", expected, step.Items)
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
