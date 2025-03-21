// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import "sync"

func Clean(configFile string, outputDir string) error {
	cmd, err := newCommand("test-clean", configFile, outputDir)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for _, tc := range cmd.Config.Tests {
		test := newTest(tc, cmd)
		wg.Add(1)
		go func() {
			cmd.CleanTest(test)
			wg.Done()
		}()
	}
	wg.Wait()

	if cmd.IsFailed() {
		return cmd.Failed()
	}

	if !cmd.Cleanup() {
		return cmd.Failed()
	}

	cmd.Passed()
	return nil
}
