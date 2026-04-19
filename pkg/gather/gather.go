// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package gather

import (
	"fmt"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/validation"
)

func Gather(opts command.ApplicationOptions) error {
	config, err := config.ReadConfig(opts.ConfigFile)
	if err != nil {
		return fmt.Errorf("unable to read config: %w", err)
	}

	cmd, err := command.New("gather-application", config.Clusters, opts.Options)
	if err != nil {
		return err
	}
	defer cmd.Close()

	gather := newCommand(cmd, config, validation.Backend{}, opts)
	return gather.Run()
}
