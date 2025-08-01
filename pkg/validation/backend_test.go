// validation_backend_test.go
package validation

import (
	"context"
	"slices"
	"testing"

	"github.com/ramendr/ramen/api/v1alpha1"
	"github.com/ramendr/ramen/e2e/types"
	"github.com/ramendr/ramenctl/pkg/config"
	"go.uber.org/zap"

	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type MockContext struct {
	ctx    context.Context
	env    *types.Env
	cfg    *config.Config
	logger *zap.SugaredLogger
}

func (m MockContext) Context() context.Context {
	return m.ctx
}

func (m MockContext) Env() *types.Env {
	return m.env
}

func (m MockContext) Config() *config.Config {
	return m.cfg
}

func (m MockContext) Logger() *zap.SugaredLogger {
	return m.logger
}

// setupScheme initializes a runtime.Scheme with v1alpha1 types
func setupScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	err := v1alpha1.AddToScheme(scheme)
	if err != nil {
		klog.Errorf("failed to add scheme: %v", err)
	}
	return scheme
}

// TestApplicationNamespaces tests Backend.ApplicationNamespaces
func TestApplicationNamespaces(t *testing.T) {
	const (
		drpcName      = "test-drpc"
		drpcNamespace = "test-namespace"
	)

	// Create a DRPlacementControl object
	drpc := &v1alpha1.DRPlacementControl{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      drpcName,
			Namespace: drpcNamespace,
			Annotations: map[string]string{
				drpcAppNamespaceAnnotation: "app-namespace-from-annotation",
			},
		},
		Spec: v1alpha1.DRPlacementControlSpec{
			ProtectedNamespaces: &[]string{
				"protected-ns-1",
				"protected-ns-2",
			},
		},
	}

	// Create fake client with the DRPC object
	fakeClient := fake.NewClientBuilder().
		WithScheme(setupScheme()).
		WithObjects(drpc).
		Build()

	testEnv := &types.Env{
		Hub: &types.Cluster{
			Name:   "hub",
			Client: fakeClient,
		},
		C1: &types.Cluster{Name: "c1"},
		C2: &types.Cluster{Name: "c2"},
	}

	ctx := MockContext{
		ctx:    context.TODO(),
		env:    testEnv,
		cfg:    &config.Config{},
		logger: zap.NewNop().Sugar(),
	}

	backend := Backend{}
	namespaces, err := backend.ApplicationNamespaces(ctx, drpcName, drpcNamespace)
	if err != nil {
		t.Fatalf("ApplicationNamespaces() error = %v", err)
	}

	// Expected namespaces
	expected := []string{
		drpcNamespace,
		"protected-ns-1",
		"protected-ns-2",
		"app-namespace-from-annotation",
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
		if namespaces[i] != ns {
			t.Fatalf("mismatch at index %d: expected %q, got %q", i, ns, namespaces[i])
		}
	}
}
