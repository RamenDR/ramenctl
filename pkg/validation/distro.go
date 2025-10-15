// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"fmt"

	e2econfig "github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/types"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// detectDistro detects the cluster distribution. If distro is set in config, it uses the user
// provided distro. Otherwise, it determines the distro, sets config.Distro and config.Namespaces.
// Returns an error if distro detection fails or if clusters have inconsistent distributions.
func detectDistro(ctx Context) error {
	cfg := ctx.Config()
	env := ctx.Env()
	log := ctx.Logger()

	if cfg.Distro != "" {
		return nil
	}

	clusters := []*types.Cluster{env.Hub, env.C1, env.C2}
	if env.PassiveHub != nil {
		clusters = append(clusters, env.PassiveHub)
	}

	var detectedDistro string
	for _, cluster := range clusters {
		distro, err := probeDistro(ctx, cluster)
		if err != nil {
			return fmt.Errorf("failed to determine distro for cluster %q: %w", cluster.Name, err)
		}

		if detectedDistro == "" {
			detectedDistro = distro
		} else if detectedDistro != distro {
			return fmt.Errorf("clusters have inconsistent distributions, cluster %q has distro %q, expected %q",
				cluster.Name, distro, detectedDistro)
		}
	}

	cfg.SetDistro(detectedDistro)

	log.Infof("Detected kubernetes distribution: %q", cfg.Distro)
	log.Infof("Using namespaces: %+v", cfg.Namespaces)

	return nil
}

// probeDistro determines the distribution of the given Kubernetes cluster.
// Returns the detected distribution name ("ocp" or "k8s") or an error if detection fails.
func probeDistro(ctx Context, cluster *types.Cluster) (string, error) {
	list := &unstructured.Unstructured{}
	list.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "config.openshift.io",
		Version: "v1",
		Kind:    "ClusterVersion",
	})
	err := cluster.Client.List(ctx.Context(), list)
	if err != nil {
		if !meta.IsNoMatchError(err) {
			return "", err
		}
		// api server says no match for OpenShift only resource type,
		// it is not OpenShift
		return e2econfig.DistroK8s, nil
	}
	// found OpenShift only resource type, it is OpenShift
	return e2econfig.DistroOcp, nil
}
