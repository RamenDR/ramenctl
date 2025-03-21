// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

func Run(configFile string, outputDir string) error {
	cmd, err := newCommand("test-run", configFile, outputDir)
	if err != nil {
		return err
	}

	// NOTE: The environment will be cleaned up by `test clean` command. If a test fail we want to keep the environment
	// as is for inspection.
	if err := cmd.Setup(); err != nil {
		return err
	}

	// We want to run all tests in parallel, but for now lets run one test.
	test := newTest(cmd.Config.Tests[0], cmd)

	if err := test.Deploy(); err != nil {
		return err
	}

	if err := test.Protect(); err != nil {
		return err
	}

	if err := test.Failover(); err != nil {
		return err
	}

	if err := test.Relocate(); err != nil {
		return err
	}

	if err := test.Unprotect(); err != nil {
		return err
	}

	if err := test.Undeploy(); err != nil {
		return err
	}

	return nil
}
