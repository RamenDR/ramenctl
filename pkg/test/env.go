// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"

	"github.com/ramendr/ramen/e2e/util"

	"github.com/ramendr/ramenctl/pkg/console"
)

func setupEnvironment(cmd *Command) error {
	console.Progress("Setup environment")
	if err := util.EnsureChannel(cmd.Env.Hub, cmd.Config, cmd.Logger); err != nil {
		err := fmt.Errorf("failed to setup environment: %w", err)
		cmd.Logger.Error(err)
		return err
	}
	console.Completed("Environment setup")
	return nil
}

func cleanEnvironment(cmd *Command) error {
	console.Progress("Clean environment")
	if err := util.EnsureChannelDeleted(cmd.Env.Hub, cmd.Config, cmd.Logger); err != nil {
		err := fmt.Errorf("failed to clean environment: %w", err)
		cmd.Logger.Error(err)
		return err
	}
	console.Completed("Environment cleaned")
	return nil
}
