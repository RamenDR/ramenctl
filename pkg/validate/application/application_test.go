// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"testing"

	e2econfig "github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/types"

	basecmd "github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/validation"
)

const (
	drpcName             = "appset-deploy-rbd"
	drpcNamespace        = "argocd"
	applicationNamespace = "e2e-appset-deploy-rbd"
)

// testSystem is a test system such as drenv or ocp clusters.
type testSystem struct {
	name   string
	config *config.Config
	env    *types.Env
}

var testK8s = testSystem{
	name: "k8s",
	config: &config.Config{
		Namespaces: e2econfig.K8sNamespaces,
	},
	env: &types.Env{
		Hub: &types.Cluster{Name: "hub"},
		C1:  &types.Cluster{Name: "dr1"},
		C2:  &types.Cluster{Name: "dr2"},
	},
}

func testCommand(
	t *testing.T,
	backend validation.Validation,
	system testSystem,
) *Command {
	cmd, err := basecmd.ForTest(CommandName, system.env, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cmd.Close()
	})
	opts := basecmd.ApplicationOptions{
		DRPCName:      drpcName,
		DRPCNamespace: drpcNamespace,
	}
	return NewCommand(cmd, system.config, backend, opts)
}

func checkReport(t *testing.T, cmd *Command, status report.Status) {
	if cmd.Report.Status != status {
		t.Fatalf("expected status %q, got %q", status, cmd.Report.Status)
	}
	if !cmd.Report.Config.Equal(cmd.Config()) {
		t.Fatalf("expected config %q, got %q", cmd.Config(), cmd.Report.Config)
	}
	duration := totalDuration(cmd.Report.Steps)
	if cmd.Report.Duration != duration {
		t.Fatalf("expected duration %v, got %v", duration, cmd.Report.Duration)
	}
	checkOutputFiles(t, cmd)
}

func checkApplication(t *testing.T, r *Report, expected *report.Application) {
	if !reflect.DeepEqual(expected, &r.Application) {
		diff := helpers.UnifiedDiff(t, expected, &r.Application)
		t.Fatalf("applications not equal\n%s", diff)
	}
}

func checkNamespaces(t *testing.T, r *Report, expected []string) {
	if !slices.Equal(r.Namespaces, expected) {
		t.Fatalf("expected namespaces %q, got %q", expected, r.Namespaces)
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

func checkError(t *testing.T, r *Report, expected string) {
	if got := r.Error(); got != expected {
		t.Fatalf("expected error %q, got %q", expected, got)
	}
}

func checkItems(t *testing.T, step *report.Step, expected []*report.Step) {
	if len(expected) != len(step.Items) {
		t.Fatalf("expected items %+v, got %+v", expected, step.Items)
	}
	for i, item := range expected {
		checkStep(t, step.Items[i], item)
	}
}

func checkApplicationStatus(
	t *testing.T,
	r *Report,
	expected *report.ApplicationStatus,
) {
	if !r.ApplicationStatus.Equal(expected) {
		diff := helpers.UnifiedDiff(t, expected, &r.ApplicationStatus)
		t.Fatalf("application statuses not equal\n%s", diff)
	}
}

func checkSummary(t *testing.T, r *Report, expected report.Summary) {
	if !r.Summary.Equal(&expected) {
		t.Fatalf("expected summary %v, got %v", expected, *r.Summary)
	}
}

func checkOutputFiles(t *testing.T, cmd *Command) {
	for _, path := range []string{
		cmd.ReportFile("yaml"),
		cmd.ReportFile("html"),
		filepath.Join(cmd.OutputDir(), "style.css"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("output file %q not found: %s", path, err)
		}
	}
}

func totalDuration(steps []*report.Step) float64 {
	var total float64
	for _, step := range steps {
		total += step.Duration
	}
	return total
}

func dumpCommandLog(t *testing.T, cmd *Command) {
	log, err := os.ReadFile(cmd.LogFile())
	if err != nil {
		t.Logf("Failed to read command log: %s", err)
		return
	}
	t.Logf("Command log:\n%s", log)
}
