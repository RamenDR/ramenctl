// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMergeMaps tests the mergeMaps function
func TestMergeMaps(t *testing.T) {
	tests := []struct {
		name        string
		baseMap     map[string]interface{}
		overrideMap map[string]interface{}
		expected    map[string]interface{}
	}{
		{
			name: "Override single value",
			baseMap: map[string]interface{}{
				"cluster1": ClusterInfo{Name: "baseCluster", KubeconfigPath: "/base/path"},
			},
			overrideMap: map[string]interface{}{
				"cluster1": ClusterInfo{Name: "overrideCluster", KubeconfigPath: "/override/path"},
			},
			expected: map[string]interface{}{
				"cluster1": ClusterInfo{Name: "overrideCluster", KubeconfigPath: "/override/path"},
			},
		},
		{
			name: "Add new key",
			baseMap: map[string]interface{}{
				"cluster1": ClusterInfo{Name: "baseCluster", KubeconfigPath: "/base/path"},
			},
			overrideMap: map[string]interface{}{
				"cluster2": ClusterInfo{Name: "newCluster", KubeconfigPath: "/new/path"},
			},
			expected: map[string]interface{}{
				"cluster1": ClusterInfo{Name: "baseCluster", KubeconfigPath: "/base/path"},
				"cluster2": ClusterInfo{Name: "newCluster", KubeconfigPath: "/new/path"},
			},
		},
		{
			name:        "Override empty base map",
			baseMap:     map[string]interface{}{},
			overrideMap: map[string]interface{}{"cluster1": ClusterInfo{Name: "newCluster", KubeconfigPath: "/new/path"}},
			expected:    map[string]interface{}{"cluster1": ClusterInfo{Name: "newCluster", KubeconfigPath: "/new/path"}},
		},
		{
			name: "Empty override map",
			baseMap: map[string]interface{}{
				"cluster1": ClusterInfo{Name: "baseCluster", KubeconfigPath: "/base/path"},
			},
			overrideMap: map[string]interface{}{},
			expected: map[string]interface{}{
				"cluster1": ClusterInfo{Name: "baseCluster", KubeconfigPath: "/base/path"},
			},
		},
		{
			name:        "Both maps empty",
			baseMap:     map[string]interface{}{},
			overrideMap: map[string]interface{}{},
			expected:    map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeMaps(tt.baseMap, tt.overrideMap)
			assert.Equal(t, tt.expected, result)
		})
	}
}
