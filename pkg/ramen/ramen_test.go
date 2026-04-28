// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package ramen

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/ramendr/ramen/api/v1alpha1"
	e2econfig "github.com/ramendr/ramen/e2e/config"
	corev1 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ramendr/ramenctl/pkg/config"
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
)

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

func checkNamespaces(t *testing.T, namespaces []string, expected []string) {
	slices.Sort(namespaces)
	if !slices.Equal(namespaces, expected) {
		t.Fatalf("expected namespaces %q, got %q", expected, namespaces)
	}
}
