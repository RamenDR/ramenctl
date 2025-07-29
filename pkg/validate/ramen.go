// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	ramen "github.com/ramendr/ramen/api/v1alpha1"
	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/gathering"
)

const (
	// TODO: find a way to get this from ramen api. Available in the CRD under spec/names/plural.
	// Should we gather the CRDs from the cluster?
	drPlacementControlPlural = "drplacementcontrols"
)

// readDRPC read a ramen DRPlacementControl resource.
func readDRPC(
	reader gathering.OutputReader,
	name, namespace string,
) (*ramen.DRPlacementControl, error) {
	resource := ramen.GroupVersion.Group + "/" + drPlacementControlPlural
	data, err := reader.ReadResource(namespace, resource, name)
	if err != nil {
		return nil, err
	}
	drpc := &ramen.DRPlacementControl{}
	if err := yaml.Unmarshal(data, drpc); err != nil {
		return nil, err
	}
	return drpc, nil
}
