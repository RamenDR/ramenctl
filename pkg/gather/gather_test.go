// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0
package gather

import (
	"path/filepath"
	"testing"

	"github.com/ramendr/ramen/e2e/types"
	"go.uber.org/zap/zaptest"
)

func TestRestConfig(t *testing.T) {
	tests := []struct {
		name          string
		kubeconfig    string
		clusterName   string
		expectedError bool
	}{
		{
			name:          "non existent kubeconfig",
			kubeconfig:    "does_not_exist.yaml",
			clusterName:   "test-cluster",
			expectedError: true,
		},
		{
			name:          "invalid kubeconfig",
			kubeconfig:    "invalid_kubeconfig.yaml",
			clusterName:   "test-cluster",
			expectedError: true,
		},
		{
			name:          "same context name",
			kubeconfig:    "valid_kubeconfig.yaml",
			clusterName:   "admin",
			expectedError: false,
		},
		{
			name:          "different context name",
			kubeconfig:    "valid_kubeconfig.yaml",
			clusterName:   "dr1",
			expectedError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			kubeconfigPath := filepath.Join("testdata", tc.kubeconfig)
			cluster := &types.Cluster{
				Name:       tc.clusterName,
				Kubeconfig: kubeconfigPath,
			}

			log := zaptest.NewLogger(t).Sugar()

			_, err := restConfig(cluster, log)
			if tc.expectedError && err == nil {
				t.Fatal("expected error but got nil")
			} else if !tc.expectedError && err != nil {
				t.Fatalf("expected no error but got: %v", err)
			}
		})
	}
}
