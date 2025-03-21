// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

func Clean(configFile string, outputDir string) error {
	cmd, err := newCommand("test-clean", configFile, outputDir)
	if err != nil {
		return err
	}

	// We want to run all tests in parallel, but for now lets run one test.
	test := newTest(cmd.Config.Tests[0], cmd)
	if err := cmd.CleanTest(test); err != nil {
		return err
	}

	if err := cmd.Cleanup(); err != nil {
		return err
	}

	return nil
}
