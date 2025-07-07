// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"context"
	"errors"
	"testing"

	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/validation"
)

const (
	validateClusters    = "validate-clusters"
	validateApplication = "validate-application"
	drpcName            = "drpc-name"
	drpcNamespace       = "drpc-namespace"
)

var (
	testConfig = &config.Config{}

	testEnv = &types.Env{
		Hub: &types.Cluster{Name: "hub"},
		C1:  &types.Cluster{Name: "c1"},
		C2:  &types.Cluster{Name: "c2"},
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
)

// Validate clusters tests.

func TestValidateClustersPassed(t *testing.T) {
	validate := testCommand(t, validateClusters, &validation.Mock{})
	if err := validate.Clusters(); err != nil {
		t.Fatal(err)
	}
	checkReport(t, validate.report, report.Passed)
	if len(validate.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.report.Steps[1], "validate clusters", report.Passed)
}

func TestValidateClustersValidateFailed(t *testing.T) {
	validate := testCommand(t, validateClusters, validateConfigFailed)
	if err := validate.Clusters(); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, validate.report, report.Failed)
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
	if len(validate.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Canceled)
}

// Validate application tests.

func TestValidateApplicationPassed(t *testing.T) {
	validate := testCommand(t, validateApplication, &validation.Mock{})
	if err := validate.Application(drpcName, drpcNamespace); err != nil {
		t.Fatal(err)
	}
	checkReport(t, validate.report, report.Passed)
	if len(validate.report.Steps) != 2 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Passed)
	checkStep(t, validate.report.Steps[1], "validate application", report.Passed)
}

func TestValidateApplicationValidateFailed(t *testing.T) {
	validate := testCommand(t, validateApplication, validateConfigFailed)
	if err := validate.Clusters(); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, validate.report, report.Failed)
	if len(validate.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Failed)
}

func TestValidateApplicationValidateCanceled(t *testing.T) {
	validate := testCommand(t, validateApplication, validateConfigCanceled)
	if err := validate.Clusters(); err == nil {
		t.Fatal("command did not fail")
	}
	checkReport(t, validate.report, report.Canceled)
	if len(validate.report.Steps) != 1 {
		t.Fatalf("unexpected steps %+v", validate.report.Steps)
	}
	checkStep(t, validate.report.Steps[0], "validate config", report.Canceled)
}

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

func checkStep(t *testing.T, step *report.Step, name string, status report.Status) {
	if name != step.Name {
		t.Fatalf("expected step %q, got %q", name, step.Name)
	}
	if status != step.Status {
		t.Fatalf("expected status %q, got %q", status, step.Status)
	}
	// We cannot check duration since it may be zero on windows.
}

func totalDuration(steps []*report.Step) float64 {
	var total float64
	for _, step := range steps {
		total += step.Duration
	}
	return total
}
