// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import "sync"

func Run(configFile string, outputDir string) error {
	cmd, err := newCommand("test-run", configFile, outputDir)
	if err != nil {
		return err
	}

	// NOTE: The environment will be cleaned up by `test clean` command. If a test fail we want to keep the environment
	// as is for inspection.
	if !cmd.Setup() {
		return cmd.Failed()
	}

	var wg sync.WaitGroup
	for _, tc := range cmd.Config.Tests {
		test := newTest(tc, cmd)
		wg.Add(1)
		go func() {
			cmd.RunTest(test)
			wg.Done()
		}()
	}
	wg.Wait()

	if cmd.IsFailed() {
		return cmd.Failed()
	}

	cmd.Passed()
	return nil
}
