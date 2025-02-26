// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"time"

	"github.com/ramendr/ramenctl/pkg/console"
)

func Clean(outputDir string) error {
	hub := "hub"
	primary := "dr1"
	secondary := "dr2"

	console.Info("Using report %q", outputDir)

	console.Progress("Cleaning up cluster %q", primary)
	time.Sleep(2 * time.Second)
	console.Completed("Cluster %q cleaned", primary)

	console.Progress("Cleaning up cluster %q", secondary)
	time.Sleep(2 * time.Second)
	console.Completed("Cluster %q cleaned", secondary)

	console.Progress("Cleaning up cluster %q", hub)
	time.Sleep(2 * time.Second)
	console.Completed("Cluster %q cleaned", hub)

	return nil
}
