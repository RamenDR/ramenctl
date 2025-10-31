// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/s3"
)

func ClusterNames(clusters []*types.Cluster) []string {
	names := make([]string, 0, len(clusters))
	for _, cluster := range clusters {
		names = append(names, cluster.Name)
	}
	return names
}

func ProfileNames(profiles []*s3.Profile) []string {
	names := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		names = append(names, profile.Name)
	}
	return names
}
