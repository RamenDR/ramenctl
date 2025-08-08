// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0
package validation

import (
	"slices"
	"testing"

	"github.com/ramendr/ramen/api/v1alpha1"
	e2etypes "github.com/ramendr/ramen/e2e/types"
	"k8s.io/apimachinery/pkg/apis/meta/internalversion/scheme"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
)

const (
	drpcName      = "test-drpc"
	drpcNamespace = "test-namespace"
	protectedNS1  = "protected-ns-1"
	protectedNS2  = "protected-ns-2"
	appNS         = "app-namespace-from-annotation"
)

type TestContext struct {
	command.Command
}

func (c *TestContext) Config() *config.Config {
	return nil
}

func init() {
	utilruntime.Must(v1alpha1.AddToScheme(scheme.Scheme))
}

// TestApplicationNamespaces tests Backend.ApplicationNamespaces
func TestApplicationNamespaces(t *testing.T) {

	drpc := &v1alpha1.DRPlacementControl{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      drpcName,
			Namespace: drpcNamespace,
			Annotations: map[string]string{
				drpcAppNamespaceAnnotation: appNS,
			},
		},
		Spec: v1alpha1.DRPlacementControlSpec{
			ProtectedNamespaces: &[]string{
				protectedNS1,
				protectedNS2,
			},
		},
	}

	// Create fake client with the DRPC object
	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme.Scheme).
		WithObjects(drpc).
		Build()

	testEnv := &e2etypes.Env{
		Hub: &e2etypes.Cluster{Name: "hub", Client: fakeClient},
		C1:  &e2etypes.Cluster{Name: "c1"},
		C2:  &e2etypes.Cluster{Name: "c2"},
	}
	backend := Backend{}

	cmd, err := command.ForTest("test", testEnv, t.TempDir())
	if err != nil {
		t.Fatalf("failed to create test command: %v", err)
	}

	ctx := &TestContext{Command: *cmd}

	namespaces, err := backend.ApplicationNamespaces(ctx, drpcName, drpcNamespace)
	if err != nil {
		t.Fatal(err)
	}

	// Expected namespaces
	expected := []string{
		drpcNamespace,
		protectedNS1,
		protectedNS2,
		appNS,
	}

	// Sort both for comparison
	slices.Sort(namespaces)
	slices.Sort(expected)

	// Compare lengths
	if len(namespaces) != len(expected) {
		t.Fatalf("expected %d namespaces, got %d: %v", len(expected), len(namespaces), namespaces)
	}

	// Compare each namespace
	for i, ns := range expected {
		if !slices.Equal(namespaces, expected) {
			t.Fatalf("mismatch at index %d: expected %q, got %q", i, ns, namespaces[i])
		}
	}
}
