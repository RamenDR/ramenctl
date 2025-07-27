// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"maps"
	"slices"

	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/report"
)

func (c *Command) Clusters() error {
	if !c.validateConfig() {
		return c.failed()
	}
	if !c.validateClusters() {
		return c.failed()
	}
	c.passed()
	return nil
}

func (c *Command) validateClusters() bool {
	console.Step("Validate clusters")
	c.startStep("validate clusters")

	namespaces := c.clustersNamespacesToGather()
	if !c.gatherNamespaces(namespaces) {
		return c.finishStep()
	}

	if !c.validateGatheredClusterData() {
		return c.finishStep()
	}

	c.finishStep()
	return true
}

func (c *Command) clustersNamespacesToGather() []string {
	seen := map[string]struct{}{
		c.config.Namespaces.RamenHubNamespace:       {},
		c.config.Namespaces.RamenDRClusterNamespace: {},
	}

	namespaces := slices.Collect(maps.Keys(seen))
	slices.Sort(namespaces)
	return namespaces
}

func (c *Command) validateGatheredClusterData() bool {
	// TODO: Validate gathered cluster data.
	step := &report.Step{Name: "validate cluster data", Status: report.Passed}
	c.current.AddStep(step)
	console.Pass("Clusters validated")
	return true
}
