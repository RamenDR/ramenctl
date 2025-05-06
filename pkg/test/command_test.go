// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"errors"
	"testing"

	"github.com/ramendr/ramen/e2e/types"
	"github.com/ramendr/ramenctl/pkg/command"
)

var testConfig = types.Config{
	PVCSpecs: []types.PVCSpecConfig{
		{Name: "block", StorageClassName: "block-storage"},
		{Name: "file", StorageClassName: "file-storage"},
	},
	Tests: []types.TestConfig{
		{Workload: "deploy", Deployer: "appset", PVCSpec: "block"},
		{Workload: "deploy", Deployer: "appset", PVCSpec: "file"},
		{Workload: "deploy", Deployer: "subscr", PVCSpec: "block"},
		{Workload: "deploy", Deployer: "subscr", PVCSpec: "file"},
		{Workload: "deploy", Deployer: "disapp", PVCSpec: "block"},
		{Workload: "deploy", Deployer: "disapp", PVCSpec: "file"},
	},
}

var testEnv = types.Env{
	Hub: types.Cluster{Name: "hub"},
	C1:  types.Cluster{Name: "c1"},
	C2:  types.Cluster{Name: "c2"},
}

var testOptions Options

var validateFailed = MockBackend{
	ValidateFunc: func(ctx types.Context) error {
		return errors.New("No validate for you!")
	},
}

var validateCanceled = MockBackend{
	ValidateFunc: func(ctx types.Context) error {
		return context.Canceled
	},
}

var setupFailed = MockBackend{
	SetupFunc: func(ctx types.Context) error {
		return errors.New("No setup for you!")
	},
}
var setupCanceled = MockBackend{
	SetupFunc: func(ctx types.Context) error {
		return context.Canceled
	},
}

var cleanupFailed = MockBackend{
	CleanupFunc: func(ctx types.Context) error {
		return errors.New("No cleanup for you!")
	},
}

var cleanupCanceled = MockBackend{
	CleanupFunc: func(ctx types.Context) error {
		return context.Canceled
	},
}

var failoverFailed = MockBackend{
	FailoverFunc: func(ctx types.TestContext) error {
		return errors.New("No failover for you!")
	},
}

var failoverCanceled = MockBackend{
	FailoverFunc: func(ctx types.TestContext) error {
		return context.Canceled
	},
}

var disappFailoverFailed = MockBackend{
	FailoverFunc: func(ctx types.TestContext) error {
		if ctx.Deployer().IsDiscovered() {
			return errors.New("No failover for you!")
		}
		return nil
	},
}

var unprotectFailed = MockBackend{
	UnprotectFunc: func(ctx types.TestContext) error {
		return errors.New("No unprotect for you!")
	},
}

var unprotectCanceled = MockBackend{
	UnprotectFunc: func(ctx types.TestContext) error {
		return context.Canceled
	},
}

var undeployFailed = MockBackend{
	UndeployFunc: func(ctx types.TestContext) error {
		return errors.New("No undeploy for you!")
	},
}

var undeployCanceled = MockBackend{
	UndeployFunc: func(ctx types.TestContext) error {
		return context.Canceled
	},
}

func TestRunPassed(t *testing.T) {
	outputDir := t.TempDir()
	cmd, err := command.ForTest("test-run", &testConfig, &testEnv, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Close()
	test := newCommand(cmd, &MockBackend{}, testOptions)
	if err := test.Run(); err != nil {
		t.Fatal(err)
	}
	if test.Report.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, test.Report.Status)
	}
	summary := Summary{Passed: len(testConfig.Tests)}
	if test.Report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, test.Report.Summary)
	}
}

func TestRunValidateFailed(t *testing.T) {
	outputDir := t.TempDir()
	cmd, err := command.ForTest("test-run", &testConfig, &testEnv, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Close()
	test := newCommand(cmd, &validateFailed, testOptions)
	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}
	if test.Report.Status != Failed {
		t.Errorf("expected status %q, got %q", Failed, test.Report.Status)
	}
	summary := Summary{}
	if test.Report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, test.Report.Summary)
	}
}

func TestRunValidateCanceled(t *testing.T) {
	outputDir := t.TempDir()
	cmd, err := command.ForTest("test-run", &testConfig, &testEnv, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Close()
	test := newCommand(cmd, &validateCanceled, testOptions)
	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}
	if test.Report.Status != Canceled {
		t.Errorf("expected status %q, got %q", Failed, test.Report.Status)
	}
	summary := Summary{}
	if test.Report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, test.Report.Summary)
	}
}

func TestRunSetupFailed(t *testing.T) {
	outputDir := t.TempDir()
	cmd, err := command.ForTest("test-run", &testConfig, &testEnv, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Close()
	test := newCommand(cmd, &setupFailed, testOptions)
	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}
	if test.Report.Status != Failed {
		t.Errorf("expected status %q, got %q", Failed, test.Report.Status)
	}
	summary := Summary{}
	if test.Report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, test.Report.Summary)
	}
}

func TestRunSetupCanceled(t *testing.T) {
	outputDir := t.TempDir()
	cmd, err := command.ForTest("test-run", &testConfig, &testEnv, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Close()
	test := newCommand(cmd, &setupCanceled, testOptions)
	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}
	if test.Report.Status != Canceled {
		t.Errorf("expected status %q, got %q", Canceled, test.Report.Status)
	}
	summary := Summary{}
	if test.Report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, test.Report.Summary)
	}
}

func TestRunTestsFailed(t *testing.T) {
	outputDir := t.TempDir()
	cmd, err := command.ForTest("test-run", &testConfig, &testEnv, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Close()
	test := newCommand(cmd, &failoverFailed, testOptions)
	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}
	if test.Report.Status != Failed {
		t.Errorf("expected status %q, got %q", Failed, test.Report.Status)
	}
	summary := Summary{Failed: len(testConfig.Tests)}
	if test.Report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, test.Report.Summary)
	}
}

func TestRunDisappFailed(t *testing.T) {
	outputDir := t.TempDir()
	cmd, err := command.ForTest("test-run", &testConfig, &testEnv, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Close()
	test := newCommand(cmd, &disappFailoverFailed, testOptions)
	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}
	if test.Report.Status != Failed {
		t.Errorf("expected status %q, got %q", Failed, test.Report.Status)
	}
	summary := Summary{Passed: 4, Failed: 2}
	if test.Report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, test.Report.Summary)
	}
}

func TestRunTestsCanceled(t *testing.T) {
	outputDir := t.TempDir()
	cmd, err := command.ForTest("test-run", &testConfig, &testEnv, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Close()
	test := newCommand(cmd, &failoverCanceled, testOptions)
	if err := test.Run(); err == nil {
		t.Fatal("command did not fail")
	}
	if test.Report.Status != Canceled {
		t.Errorf("expected status %q, got %q", Failed, test.Report.Status)
	}
	summary := Summary{Canceled: len(testConfig.Tests)}
	if test.Report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, test.Report.Summary)
	}
}

func TestCleanPassed(t *testing.T) {
	outputDir := t.TempDir()
	cmd, err := command.ForTest("test-run", &testConfig, &testEnv, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Close()
	test := newCommand(cmd, &MockBackend{}, testOptions)
	if err := test.Clean(); err != nil {
		t.Fatal(err)
	}
	if test.Report.Status != Passed {
		t.Errorf("expected status %q, got %q", Passed, test.Report.Status)
	}
	summary := Summary{Passed: 6}
	if test.Report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, test.Report.Summary)
	}
}

func TestCleanValidateFailed(t *testing.T) {
	outputDir := t.TempDir()
	cmd, err := command.ForTest("test-run", &testConfig, &testEnv, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Close()
	test := newCommand(cmd, &validateFailed, testOptions)
	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}
	if test.Report.Status != Failed {
		t.Errorf("expected status %q, got %q", Failed, test.Report.Status)
	}
	summary := Summary{}
	if test.Report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, test.Report.Summary)
	}
}

func TestCleanValidateCanceled(t *testing.T) {
	outputDir := t.TempDir()
	cmd, err := command.ForTest("test-run", &testConfig, &testEnv, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Close()
	test := newCommand(cmd, &validateCanceled, testOptions)
	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}
	if test.Report.Status != Canceled {
		t.Errorf("expected status %q, got %q", Canceled, test.Report.Status)
	}
	summary := Summary{}
	if test.Report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, test.Report.Summary)
	}
}

func TestCleanUnprotectFailed(t *testing.T) {
	outputDir := t.TempDir()
	cmd, err := command.ForTest("test-run", &testConfig, &testEnv, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Close()
	test := newCommand(cmd, &unprotectFailed, testOptions)
	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}
	if test.Report.Status != Failed {
		t.Errorf("expected status %q, got %q", Failed, test.Report.Status)
	}
	summary := Summary{Failed: len(testConfig.Tests)}
	if test.Report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, test.Report.Summary)
	}
}

func TestCleanUndeployFailed(t *testing.T) {
	outputDir := t.TempDir()
	cmd, err := command.ForTest("test-run", &testConfig, &testEnv, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Close()
	test := newCommand(cmd, &undeployFailed, testOptions)
	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}
	if test.Report.Status != Failed {
		t.Errorf("expected status %q, got %q", Failed, test.Report.Status)
	}
	summary := Summary{Failed: len(testConfig.Tests)}
	if test.Report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, test.Report.Summary)
	}
}

func TestCleanUnprotectCanceled(t *testing.T) {
	outputDir := t.TempDir()
	cmd, err := command.ForTest("test-run", &testConfig, &testEnv, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Close()
	test := newCommand(cmd, &unprotectCanceled, testOptions)
	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}
	if test.Report.Status != Canceled {
		t.Errorf("expected status %q, got %q", Canceled, test.Report.Status)
	}
	summary := Summary{Canceled: len(testConfig.Tests)}
	if test.Report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, test.Report.Summary)
	}
}

func TestCleanUndeployCanceled(t *testing.T) {
	outputDir := t.TempDir()
	cmd, err := command.ForTest("test-run", &testConfig, &testEnv, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Close()
	test := newCommand(cmd, &undeployCanceled, testOptions)
	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}
	if test.Report.Status != Canceled {
		t.Errorf("expected status %q, got %q", Canceled, test.Report.Status)
	}
	summary := Summary{Canceled: len(testConfig.Tests)}
	if test.Report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, test.Report.Summary)
	}
}

func TestCleanCleanupFailed(t *testing.T) {
	outputDir := t.TempDir()
	cmd, err := command.ForTest("test-run", &testConfig, &testEnv, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Close()
	test := newCommand(cmd, &cleanupFailed, testOptions)
	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}
	if test.Report.Status != Failed {
		t.Errorf("expected status %q, got %q", Failed, test.Report.Status)
	}
	summary := Summary{Passed: len(testConfig.Tests)}
	if test.Report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, test.Report.Summary)
	}
}

func TestCleanCleanupCanceled(t *testing.T) {
	outputDir := t.TempDir()
	cmd, err := command.ForTest("test-run", &testConfig, &testEnv, outputDir)
	if err != nil {
		t.Fatal(err)
	}
	defer cmd.Close()
	test := newCommand(cmd, &cleanupCanceled, testOptions)
	if err := test.Clean(); err == nil {
		t.Fatal("command did not fail")
	}
	if test.Report.Status != Canceled {
		t.Errorf("expected status %q, got %q", Failed, test.Report.Status)
	}
	summary := Summary{Passed: len(testConfig.Tests)}
	if test.Report.Summary != summary {
		t.Errorf("expected summary %+v, got %+v", summary, test.Report.Summary)
	}
}
