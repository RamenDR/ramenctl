// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import "github.com/ramendr/ramenctl/pkg/command"

func Run(configFile string, outputDir string) error {
	cmd, err := command.New("test-run", configFile, outputDir)
	if err != nil {
		return err
	}
	defer cmd.Close()

	testCmd := newCommand(cmd)

	if !testCmd.Validate() {
		return testCmd.Failed()
	}

	// NOTE: The environment will be cleaned up by `test clean` command. If a test fail we want to keep the environment
	// as is for inspection.
	if !testCmd.Setup() {
		return testCmd.Failed()
	}

	if !testCmd.RunTests() {
		return testCmd.Failed()
	}

	testCmd.Passed()
	return nil
}
