// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import "github.com/ramendr/ramenctl/pkg/command"

func Clean(configFile string, outputDir string) error {
	cmd, err := command.New("test-clean", configFile, outputDir)
	if err != nil {
		return err
	}
	defer cmd.Stop()

	testCmd := newCommand(cmd)

	if !testCmd.Validate() {
		return testCmd.Failed()
	}

	if !testCmd.CleanTests() {
		return testCmd.Failed()
	}

	if !testCmd.Cleanup() {
		return testCmd.Failed()
	}

	testCmd.Passed()
	return nil
}
