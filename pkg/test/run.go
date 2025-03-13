// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"
	"time"

	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
)

func Run(configFile string, outputDir string) error {
	config, err := config.ReadConfig(configFile)
	if err != nil {
		return err
	}

	// We want to run all tests in parallel, but for now lets run one test.
	test := config.Tests[0]

	app := fmt.Sprintf("test-%s-%s-%v", test.Deployer, test.Workload, test.PVCSpec)
	primary := config.Clusters["c1"].Name
	secondary := config.Clusters["c2"].Name

	console.Info("Using report %q", outputDir)

	console.Progress("Deploying application %q on cluster %q", app, primary)
	time.Sleep(2 * time.Second)
	console.Completed("Application %q deployed on cluster %q", app, primary)

	console.Progress("Protecting application %q in cluster %q", app, primary)
	time.Sleep(5 * time.Second)
	console.Completed("Application %q protected in cluster %q", app, primary)

	console.Progress("Failing over application %q to cluster %q", app, secondary)
	time.Sleep(10 * time.Second)
	console.Completed("Application %q failed over to cluster %q", app, secondary)

	console.Progress("Unprotecting application %q in cluster %q", app, secondary)
	time.Sleep(5 * time.Second)
	console.Completed("Application %q unprotected in cluster %q", app, secondary)

	console.Progress("Protecting application %q in cluster %q", app, secondary)
	time.Sleep(5 * time.Second)
	console.Completed("Application %q protected in cluster %q", app, secondary)

	console.Progress("Relocating application %q to cluster %q", app, primary)
	time.Sleep(10 * time.Second)
	console.Completed("Application %q relocated to cluster %q", app, primary)

	console.Progress("Unprotecting application %q in cluster %q", app, primary)
	time.Sleep(5 * time.Second)
	console.Completed("Application %q unprotected in cluster %q", app, primary)

	console.Progress("Undeploying application %q from cluster %q", app, primary)
	time.Sleep(2 * time.Second)
	console.Completed("Application %q undeployed from cluster %q", app, primary)

	return nil
}
