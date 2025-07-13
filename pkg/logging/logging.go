// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"github.com/ramendr/ramen/e2e/types"
)

func ClusterNames(clusters []*types.Cluster) []string {
	names := make([]string, 0, len(clusters))
	for _, cluster := range clusters {
		names = append(names, cluster.Name)
	}
	return names
}
