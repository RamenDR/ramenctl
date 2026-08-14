// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package ramen

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/nirs/kubectl-gather/pkg/gather"
	"github.com/ramendr/ramen/api/v1alpha1"
	e2econfig "github.com/ramendr/ramen/e2e/config"
	e2etypes "github.com/ramendr/ramen/e2e/types"
	corev1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/sets"
)

var (
	testConfig = &config.Config{
		Namespaces: e2econfig.K8sNamespaces,
	}
)

const (
	disappName               = "disapp-deploy-rbd"
	disappProtectedNamespace = "e2e-disapp-deploy-rbd"

	testHubName     = "hub"
	testPrimaryName = "dr1"
	testDRPCName    = "my-app"
	testNamespace   = "my-app-ns"
)

// testContext implements ramen.Context for testing ApplicationProfiles.
type testContext struct {
	env     *e2etypes.Env
	config  *config.Config
	dataDir string
}

func (c *testContext) Env() *e2etypes.Env {
	return c.env
}

func (c *testContext) Context() context.Context {
	return context.Background()
}

func (c *testContext) Config() *config.Config {
	return c.config
}

func (c *testContext) OutputReader(cluster string) gathering.OutputReader {
	clusterDir := filepath.Join(c.dataDir, cluster)
	return gather.NewOutputReader(clusterDir)
}

func TestApplicationNamespacesAppSet(t *testing.T) {
	drpc := &v1alpha1.DRPlacementControl{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      "appset-deploy-rbd",
			Namespace: testConfig.Namespaces.ArgocdNamespace,
			Annotations: map[string]string{
				drpcAppNamespaceAnnotation: "e2e-appset-deploy-rbd",
			},
		},
	}

	namespaces := ApplicationNamespaces(drpc)
	expectedNamespaces := sets.Sorted([]string{
		testConfig.Namespaces.ArgocdNamespace,
		"e2e-appset-deploy-rbd",
	})
	checkNamespaces(t, namespaces, expectedNamespaces)
}

func TestApplicationNamespacesSubscription(t *testing.T) {
	drpc := &v1alpha1.DRPlacementControl{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      "subscr-deploy-rbd",
			Namespace: "e2e-subscr-deploy-rbd",
			Annotations: map[string]string{
				drpcAppNamespaceAnnotation: "e2e-subscr-deploy-rbd",
			},
		},
	}

	namespaces := ApplicationNamespaces(drpc)
	expectedNamespaces := []string{"e2e-subscr-deploy-rbd"}
	checkNamespaces(t, namespaces, expectedNamespaces)
}

func TestApplicationNamespacesDiscoveredApp(t *testing.T) {
	drpc := &v1alpha1.DRPlacementControl{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      disappName,
			Namespace: testConfig.Namespaces.RamenOpsNamespace,
			Annotations: map[string]string{
				drpcAppNamespaceAnnotation: testConfig.Namespaces.RamenOpsNamespace,
			},
		},
		Spec: v1alpha1.DRPlacementControlSpec{
			ProtectedNamespaces: &[]string{disappProtectedNamespace},
		},
	}

	namespaces := ApplicationNamespaces(drpc)
	expectedNamespaces := sets.Sorted([]string{
		testConfig.Namespaces.RamenOpsNamespace,
		disappProtectedNamespace,
	})
	checkNamespaces(t, namespaces, expectedNamespaces)
}

func TestApplicationNamespacesDuplicateProtectedNamespaces(t *testing.T) {
	// example drpc for disapp as protected namespaces are part of disapps only.
	drpc := &v1alpha1.DRPlacementControl{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      disappName,
			Namespace: testConfig.Namespaces.RamenOpsNamespace,
			Annotations: map[string]string{
				drpcAppNamespaceAnnotation: testConfig.Namespaces.RamenOpsNamespace,
			},
		},
		Spec: v1alpha1.DRPlacementControlSpec{
			ProtectedNamespaces: &[]string{"duplicate", "duplicate", "unique"},
		},
	}

	namespaces := ApplicationNamespaces(drpc)
	expectedNamespaces := sets.Sorted([]string{
		testConfig.Namespaces.RamenOpsNamespace,
		"duplicate",
		"unique",
	})
	checkNamespaces(t, namespaces, expectedNamespaces)

}

func TestApplicationNamespacesMissingAppNamespaceAnnotation(t *testing.T) {
	drpc := &v1alpha1.DRPlacementControl{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      testConfig.Distro,
			Namespace: testConfig.Namespaces.RamenOpsNamespace,
			// No annotation
		},
		Spec: v1alpha1.DRPlacementControlSpec{
			ProtectedNamespaces: &[]string{disappProtectedNamespace},
		},
	}

	namespaces := ApplicationNamespaces(drpc)
	expectedNamespaces := sets.Sorted([]string{
		testConfig.Namespaces.RamenOpsNamespace,
		disappProtectedNamespace,
	})
	checkNamespaces(t, namespaces, expectedNamespaces)
}

func TestApplicationNamespacesEmptyAppNamespaceAnnotation(t *testing.T) {
	drpc := &v1alpha1.DRPlacementControl{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      disappName,
			Namespace: testConfig.Namespaces.RamenOpsNamespace,
			Annotations: map[string]string{
				drpcAppNamespaceAnnotation: "", // empty!
			},
		},
		Spec: v1alpha1.DRPlacementControlSpec{
			ProtectedNamespaces: &[]string{disappProtectedNamespace},
		},
	}

	namespaces := ApplicationNamespaces(drpc)
	expectedNamespaces := sets.Sorted([]string{
		testConfig.Namespaces.RamenOpsNamespace,
		disappProtectedNamespace,
	})
	checkNamespaces(t, namespaces, expectedNamespaces)
}

func TestParseRamenConfig(t *testing.T) {
	t.Run("valid yaml", func(t *testing.T) {
		configMap := &corev1.ConfigMap{
			Data: map[string]string{
				ConfigMapRamenConfigKeyName: "apiVersion: ramendr.openshift.io/v1alpha1\nkind: RamenConfig\n",
			},
		}
		config, err := ParseRamenConfig(configMap)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if config.Kind != "RamenConfig" {
			t.Fatalf("expected kind %q, got %q", "RamenConfig", config.Kind)
		}
		if config.APIVersion != "ramendr.openshift.io/v1alpha1" {
			t.Fatalf(
				"expected apiVersion %q, got %q",
				"ramendr.openshift.io/v1alpha1",
				config.APIVersion,
			)
		}
	})

	// The error is used as a description in the YAML and HTML validation
	// reports, so it must be a short single line. It must also wrap the
	// underlying error so callers can inspect the cause.
	t.Run("invalid yaml", func(t *testing.T) {
		configMap := &corev1.ConfigMap{
			Data: map[string]string{
				ConfigMapRamenConfigKeyName: "invalid: yaml: data\n",
			},
		}
		_, err := ParseRamenConfig(configMap)
		if err == nil {
			t.Fatal("expected error")
		}
		msg := err.Error()
		if strings.Contains(msg, "\n") {
			t.Errorf("error should be a single line: %q", msg)
		}
		if len(msg) > 256 {
			t.Errorf("error too long for reports (%d chars): %q", len(msg), msg)
		}
		if errors.Unwrap(err) == nil {
			t.Error("error should wrap the underlying yaml error")
		}
	})
}

func TestApplicationProfiles(t *testing.T) {
	tests := []struct {
		name         string
		hubProfiles  []string
		vrgProfiles  []string
		wantProfiles []string
	}{
		{
			name:         "matching profiles",
			hubProfiles:  []string{"profile-1", "profile-2"},
			vrgProfiles:  []string{"profile-1", "profile-2"},
			wantProfiles: []string{"profile-1", "profile-2"},
		},
		{
			name:         "extra profiles in hub config",
			hubProfiles:  []string{"profile-1", "profile-2", "profile-3"},
			vrgProfiles:  []string{"profile-1", "profile-2"},
			wantProfiles: []string{"profile-1", "profile-2"},
		},
		{
			name:         "missing profile in hub config",
			hubProfiles:  []string{"profile-1", "profile-2"},
			vrgProfiles:  []string{"profile-1", "profile-2", "profile-3"},
			wantProfiles: []string{"profile-1", "profile-2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestContext(t)
			writeConfigMap(t, ctx, tt.hubProfiles)
			writeDRPC(t, ctx)
			writeVRG(t, ctx, tt.vrgProfiles)

			profiles, err := ApplicationProfiles(ctx, testDRPCName, testNamespace)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			got := make([]string, len(profiles))
			for i, p := range profiles {
				got[i] = p.S3ProfileName
			}
			if !slices.Equal(got, tt.wantProfiles) {
				t.Errorf("expected profiles %v, got %v", tt.wantProfiles, got)
			}
		})
	}
}

func TestApplicationProfilesErrors(t *testing.T) {
	tests := []struct {
		name        string
		hubProfiles []string
		vrgProfiles []string
	}{
		{
			name:        "no profiles match",
			hubProfiles: []string{"profile-1", "profile-2"},
			vrgProfiles: []string{"non-existent"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := newTestContext(t)
			writeConfigMap(t, ctx, tt.hubProfiles)
			writeDRPC(t, ctx)
			writeVRG(t, ctx, tt.vrgProfiles)

			profiles, err := ApplicationProfiles(ctx, testDRPCName, testNamespace)
			if err == nil {
				t.Fatal("expected error")
			}
			if profiles != nil {
				t.Errorf("expected nil profiles, got %v", profiles)
			}
		})
	}
}

func checkNamespaces(t *testing.T, namespaces []string, expected []string) {
	slices.Sort(namespaces)
	if !slices.Equal(namespaces, expected) {
		t.Fatalf("expected namespaces %q, got %q", expected, namespaces)
	}
}

func newTestContext(t *testing.T) *testContext {
	t.Helper()
	dataDir := t.TempDir()
	return &testContext{
		env: &e2etypes.Env{
			Hub: &e2etypes.Cluster{Name: testHubName},
			C1:  &e2etypes.Cluster{Name: testPrimaryName},
			C2:  &e2etypes.Cluster{Name: "dr2"},
		},
		config:  testConfig,
		dataDir: dataDir,
	}
}

func writeConfigMap(t *testing.T, ctx *testContext, profileNames []string) {
	t.Helper()
	ramenConfig := &v1alpha1.RamenConfig{
		S3StoreProfiles: make([]v1alpha1.S3StoreProfile, len(profileNames)),
	}
	for i, name := range profileNames {
		ramenConfig.S3StoreProfiles[i] = v1alpha1.S3StoreProfile{
			S3ProfileName: name,
		}
	}
	configData, err := yaml.Marshal(ramenConfig)
	if err != nil {
		t.Fatal(err)
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      HubOperatorConfigMapName,
			Namespace: ctx.config.Namespaces.RamenHubNamespace,
		},
		Data: map[string]string{
			ConfigMapRamenConfigKeyName: string(configData),
		},
	}
	writeResource(t, filepath.Join(ctx.dataDir, testHubName), cm)
}

func writeDRPC(t *testing.T, ctx *testContext) {
	t.Helper()
	drpc := &v1alpha1.DRPlacementControl{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      testDRPCName,
			Namespace: testNamespace,
			Annotations: map[string]string{
				drpcAppNamespaceAnnotation: testNamespace,
			},
		},
		Spec: v1alpha1.DRPlacementControlSpec{
			PreferredCluster: testPrimaryName,
		},
	}
	writeResource(t, filepath.Join(ctx.dataDir, testHubName), drpc)
}

func writeVRG(t *testing.T, ctx *testContext, s3Profiles []string) {
	t.Helper()
	vrg := &v1alpha1.VolumeReplicationGroup{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      testDRPCName,
			Namespace: testNamespace,
		},
		Spec: v1alpha1.VolumeReplicationGroupSpec{
			S3Profiles: s3Profiles,
		},
	}
	writeResource(t, filepath.Join(ctx.dataDir, testPrimaryName), vrg)
}

func writeResource(t *testing.T, clusterDir string, obj v1meta.Object) {
	t.Helper()

	var resource string
	switch obj.(type) {
	case *corev1.ConfigMap:
		resource = "configmaps"
	case *v1alpha1.DRPlacementControl:
		resource = v1alpha1.GroupVersion.Group + "/" + drpcPlural
	case *v1alpha1.VolumeReplicationGroup:
		resource = v1alpha1.GroupVersion.Group + "/" + vrgPlural
	default:
		t.Fatalf("unsupported resource type: %T", obj)
	}

	data, err := yaml.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(clusterDir, "namespaces", obj.GetNamespace(), resource)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, obj.GetName()+".yaml")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}
