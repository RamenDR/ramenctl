// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/e2e"
)

func Clean(configFile string, outputDir string) error {
	cmd, err := command.New("test-clean", configFile, outputDir)
	if err != nil {
		return err
	}
	defer cmd.Close()

	test := newCommand(cmd, e2e.Backend{})
	return test.Clean()
}
