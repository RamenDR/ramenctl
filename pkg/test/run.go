// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"time"

	"github.com/ramendr/ramenctl/pkg/console"
)

func Run(outputDir string) error {
	app := "ramenctl-test-app"
	primary := "dr1"
	secondary := "dr2"

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
