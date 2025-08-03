// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package sets

import (
	"maps"
	"slices"
)

// Sorted returns sorted set of values removing duplicates.
func Sorted(values []string) []string {
	set := map[string]struct{}{}
	for _, v := range values {
		set[v] = struct{}{}
	}
	return slices.Sorted(maps.Keys(set))
}
