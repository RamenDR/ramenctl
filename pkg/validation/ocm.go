// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"
	"slices"

	ocmv1 "open-cluster-management.io/api/cluster/v1"
	ocmv1b2 "open-cluster-management.io/api/cluster/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func validateClusterset(ctx Context) error {
	env := ctx.Env()
	cfg := ctx.Config()
	log := ctx.Logger()
	if _, err := getClusterSet(ctx, cfg.ClusterSet); err != nil {
		return err
	}
	clusterNames, err := listManagedClustersForClusterSet(ctx, cfg.ClusterSet)
	if err != nil {
		return err
	}
	clusters := env.ManagedClusters()
	for _, cluster := range clusters {
		if !slices.Contains(clusterNames, cluster.Name) {
			return fmt.Errorf(
				"cluster %q is not defined in clusterSet %q, clusters in ClusterSet: %q",
				cluster.Name,
				cfg.ClusterSet,
				clusterNames,
			)
		}
	}
	log.Infof(
		"Validated clusters [%q, %q] in clusterSet %q",
		clusters[0].Name,
		clusters[1].Name,
		cfg.ClusterSet,
	)
	return nil
}

func getClusterSet(ctx Context, clusterSetName string) (*ocmv1b2.ManagedClusterSet, error) {
	hub := ctx.Env().Hub
	clusterSet := &ocmv1b2.ManagedClusterSet{}
	key := client.ObjectKey{Name: clusterSetName}
	if err := hub.Client.Get(ctx.Context(), key, clusterSet); err != nil {
		return nil, fmt.Errorf("failed to get clusterSet %q: %w", clusterSetName, err)
	}
	return clusterSet, nil
}

func listManagedClustersForClusterSet(ctx Context, clusterSetName string) ([]string, error) {
	hub := ctx.Env().Hub
	list := &ocmv1.ManagedClusterList{}
	labelSelector := client.MatchingLabels{
		"cluster.open-cluster-management.io/clusterset": clusterSetName,
	}
	if err := hub.Client.List(ctx.Context(), list, labelSelector); err != nil {
		return nil, fmt.Errorf(
			"failed to list ManagedClusters for clusterSet %q: %w",
			clusterSetName,
			err,
		)
	}
	if len(list.Items) == 0 {
		return nil, fmt.Errorf("no clusters found for clusterSet %q", clusterSetName)
	}
	clusterNames := make([]string, 0, len(list.Items))
	for _, cluster := range list.Items {
		clusterNames = append(clusterNames, cluster.Name)
	}
	return clusterNames, nil
}
