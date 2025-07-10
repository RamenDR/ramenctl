// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"maps"
	"slices"

	"github.com/ramendr/ramen/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	// Annotation for application namespace on the managed cluster
	// from ramen/internal/controllers/drplacementcontrol.go
	drpcAppNamespaceAnnotation = "drplacementcontrol.ramendr.openshift.io/app-namespace"
)

func drpcNamespaces(ctx Context, drpcName, drpcNamespace string) ([]string, error) {
	drpc, err := getDRPC(ctx, drpcName, drpcNamespace)
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{
		drpcNamespace: {},
	}
	if drpc.Spec.ProtectedNamespaces != nil {
		for _, ns := range *drpc.Spec.ProtectedNamespaces {
			seen[ns] = struct{}{}
		}
	}
	if appNamespace, ok := drpc.Annotations[drpcAppNamespaceAnnotation]; ok {
		seen[appNamespace] = struct{}{}
	}
	return slices.Collect(maps.Keys(seen)), nil
}

func getDRPC(ctx Context, drpcName, drpcNamespace string) (*v1alpha1.DRPlacementControl, error) {
	drpc := &v1alpha1.DRPlacementControl{}
	key := types.NamespacedName{Namespace: drpcNamespace, Name: drpcName}
	err := ctx.Env().Hub.Client.Get(ctx.Context(), key, drpc)
	if err != nil {
		return nil, err
	}
	return drpc, nil
}
