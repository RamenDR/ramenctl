// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"slices"
	"testing"

	"github.com/ramendr/ramen/api/v1alpha1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ramendr/ramenctl/pkg/ramen"
	"github.com/ramendr/ramenctl/pkg/sets"
)

const (
	drpcAppNamespaceAnnotation = "drplacementcontrol.ramendr.openshift.io/app-namespace"
	appsetDRPCNamespace        = "argocd"
	appsetAppNamespace         = "e2e-appset-deploy-rbd"
	subscrDRPCNamespace        = "e2e-subscr-deploy-rbd"
	subscrAppNamespace         = "e2e-subscr-deploy-rbd"
	disappDRPCNamespace        = "ramen-ops"
	disappAppNamespace         = "ramen-ops"
	protectedNS1               = "protected-ns-1"
	protectedNS2               = "protected-ns-2"
	protectedNS3               = "protected-ns-3"
)

func checkNamespaces(t *testing.T, namespaces []string, expected []string) {
	if !slices.Equal(namespaces, expected) {
		t.Fatalf("expected namespaces %q, got %q", expected, namespaces)
	}
}

func TestApplicationNamespacesAppSet(t *testing.T) {
	drpc := &v1alpha1.DRPlacementControl{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      "appset-drpc",
			Namespace: appsetDRPCNamespace,
			Annotations: map[string]string{
				drpcAppNamespaceAnnotation: appsetAppNamespace,
			},
		},
		Spec: v1alpha1.DRPlacementControlSpec{
			ProtectedNamespaces: &[]string{"default", "kube-system"},
		},
	}

	namespaces := ramen.ApplicationNamespaces(drpc)

	expectedNamespaces := sets.Sorted([]string{appsetDRPCNamespace,
		appsetAppNamespace,
		"default",
		"kube-system"})

	checkNamespaces(t, sets.Sorted(namespaces), expectedNamespaces)
}

func TestApplicationNamespacesSubscription(t *testing.T) {
	drpc := &v1alpha1.DRPlacementControl{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      "subscr-drpc",
			Namespace: subscrDRPCNamespace,
			Annotations: map[string]string{
				drpcAppNamespaceAnnotation: subscrAppNamespace,
			},
		},
		Spec: v1alpha1.DRPlacementControlSpec{
			ProtectedNamespaces: &[]string{"monitoring", "logging"},
		},
	}

	namespaces := ramen.ApplicationNamespaces(drpc)
	expectedNamespaces := sets.Sorted([]string{subscrDRPCNamespace,
		subscrAppNamespace,
		"monitoring",
		"logging"})

	checkNamespaces(t, sets.Sorted(namespaces), expectedNamespaces)
}

func TestApplicationNamespacesDiscoveredApp(t *testing.T) {
	drpc := &v1alpha1.DRPlacementControl{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      "disapp-drpc",
			Namespace: disappDRPCNamespace,
			Annotations: map[string]string{
				drpcAppNamespaceAnnotation: disappAppNamespace,
			},
		},
		Spec: v1alpha1.DRPlacementControlSpec{
			ProtectedNamespaces: &[]string{protectedNS1, protectedNS2, protectedNS3},
		},
	}

	namespaces := ramen.ApplicationNamespaces(drpc)
	expectedNamespaces := sets.Sorted([]string{disappDRPCNamespace,
		disappAppNamespace,
		protectedNS1,
		protectedNS2,
		protectedNS3})

	checkNamespaces(t, sets.Sorted(namespaces), expectedNamespaces)
}

func TestApplicationNamespacesDuplicateProtectedNamespaces(t *testing.T) {
	drpc := &v1alpha1.DRPlacementControl{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      "duplicate-drpc",
			Namespace: "test-ns",
			Annotations: map[string]string{
				drpcAppNamespaceAnnotation: "app-ns",
			},
		},
		Spec: v1alpha1.DRPlacementControlSpec{
			ProtectedNamespaces: &[]string{"duplicate", "duplicate", "unique"},
		},
	}
	expectedNamespaces := sets.Sorted([]string{"test-ns", "app-ns", "duplicate", "unique"})
	namespaces := ramen.ApplicationNamespaces(drpc)
	checkNamespaces(t, sets.Sorted(namespaces), expectedNamespaces)

}

func TestApplicationNamespacesMissingAppNamespaceAnnotation(t *testing.T) {
	drpc := &v1alpha1.DRPlacementControl{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      "missing-annotation-drpc",
			Namespace: "test-ns",
			// No annotation
		},
		Spec: v1alpha1.DRPlacementControlSpec{
			ProtectedNamespaces: &[]string{"app-ns"},
		},
	}

	namespaces := ramen.ApplicationNamespaces(drpc)
	expectedNamespaces := sets.Sorted([]string{"test-ns", "app-ns"})
	checkNamespaces(t, sets.Sorted(namespaces), expectedNamespaces)
}

func TestApplicationNamespacesEmptyAppNamespaceAnnotation(t *testing.T) {
	drpc := &v1alpha1.DRPlacementControl{
		ObjectMeta: v1meta.ObjectMeta{
			Name:      "empty-annotation-drpc",
			Namespace: "test-ns",
			Annotations: map[string]string{
				drpcAppNamespaceAnnotation: "", // empty!
			},
		},
		Spec: v1alpha1.DRPlacementControlSpec{
			ProtectedNamespaces: &[]string{"app-ns"},
		},
	}

	namespaces := ramen.ApplicationNamespaces(drpc)
	expectedNamespaces := sets.Sorted([]string{"", "test-ns", "app-ns"})
	checkNamespaces(t, sets.Sorted(namespaces), expectedNamespaces)
}
