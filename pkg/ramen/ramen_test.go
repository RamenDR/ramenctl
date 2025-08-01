// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package ramen

import (
	"slices"
	"testing"

	"github.com/ramendr/ramen/api/v1alpha1"
	e2econfig "github.com/ramendr/ramen/e2e/config"
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

func checkNamespaces(t *testing.T, namespaces []string, expected []string) {
	slices.Sort(namespaces)
	if !slices.Equal(namespaces, expected) {
		t.Fatalf("expected namespaces %q, got %q", expected, namespaces)
	}
}
