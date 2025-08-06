// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/sets"
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
	c.report.Namespaces = namespaces

	options := gathering.Options{
		Namespaces: namespaces,
		Cluster:    true,
		OutputDir:  c.dataDir(),
	}
	if !c.gatherNamespaces(options) {
		return c.finishStep()
	}

	if !c.validateGatheredClusterData() {
		return c.finishStep()
	}

	c.finishStep()
	return true
}

func (c *Command) clustersNamespacesToGather() []string {
	return sets.Sorted([]string{
		c.config.Namespaces.RamenHubNamespace,
		c.config.Namespaces.RamenDRClusterNamespace,
	})
}

func (c *Command) validateGatheredClusterData() bool {
	// TODO: Validate gathered cluster data.
	step := &report.Step{Name: "validate cluster data", Status: report.Passed}
	c.current.AddStep(step)
	console.Pass("Clusters validated")
	return true
}
