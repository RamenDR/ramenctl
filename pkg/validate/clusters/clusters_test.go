// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package clusters

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
	"github.com/ramendr/ramenctl/pkg/sets"
	"github.com/ramendr/ramenctl/pkg/validation"
)

// caCertificate fingerprint (SHA-256 hash) for OCP testdata.
const caCertificateFingerprint = "BA:A5:C7:3B:3F:6E:06:27:19:F5:45:FC:6F:07:42:81:3B:F6:4D:61:95:CC:D5:D8:79:22:65:63:35:63:97:00"

// testSystem is a test system such as drenv or ocp clusters.
type testSystem struct {
	name       string
	config     *config.Config
	env        *types.Env
	namespaces []string
}

var (
	testK8s = testSystem{
		name: "k8s",
		config: &config.Config{
			Namespaces: e2econfig.K8sNamespaces,
		},
		env: &types.Env{
			Hub: &types.Cluster{Name: "hub"},
			C1:  &types.Cluster{Name: "dr1"},
			C2:  &types.Cluster{Name: "dr2"},
		},
		namespaces: sets.Sorted([]string{
			e2econfig.K8sNamespaces.RamenHubNamespace,
			e2econfig.K8sNamespaces.RamenDRClusterNamespace,
		}),
	}

	testOcp = testSystem{
		name: "ocp",
		config: &config.Config{
			Namespaces: e2econfig.OcpNamespaces,
		},
		env: &types.Env{
			Hub: &types.Cluster{Name: "hub"},
			C1:  &types.Cluster{Name: "c1"},
			C2:  &types.Cluster{Name: "c2"},
		},
		namespaces: sets.Sorted([]string{
			e2econfig.OcpNamespaces.RamenHubNamespace,
			e2econfig.OcpNamespaces.RamenDRClusterNamespace,
		}),
	}
)

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
	return NewCommand(cmd, system.config, backend)
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
	if !reflect.DeepEqual(expected, r.Application) {
		diff := helpers.UnifiedDiff(t, expected, r.Application)
		t.Fatalf("applications not equal\n%s", diff)
	}
}

func checkNamespaces(t *testing.T, r *Report, expected []string) {
	if !slices.Equal(r.Namespaces, expected) {
		t.Fatalf("expected namespaces %q, got %q", expected, r.Namespaces)
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

func checkClusterStatus(
	t *testing.T,
	r *Report,
	expected *report.ClustersStatus,
) {
	if !r.ClustersStatus.Equal(expected) {
		diff := helpers.UnifiedDiff(t, expected, &r.ClustersStatus)
		t.Fatalf("clusters statuses not equal\n%s", diff)
	}
}

func checkSummary(t *testing.T, r *Report, expected report.Summary) {
	if !r.Summary.Equal(&expected) {
		t.Fatalf("expected summary %v, got %v", expected, *r.Summary)
	}
}

func checkOutputFiles(t *testing.T, cmd *Command) {
	for _, name := range []string{CommandName + ".yaml", CommandName + ".html", "style.css"} {
		path := filepath.Join(cmd.OutputDir(), name)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("output file %q not found: %s", name, err)
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
