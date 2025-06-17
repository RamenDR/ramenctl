// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/ramendr/ramenctl/pkg/console"
)

func CreateSampleConfig(filename, commandName, envFile string) error {
	var sample *Sample
	if envFile != "" {
		console.Info("Using envfile %q", envFile)
		env, err := ReadEnvFile(envFile)
		if err != nil {
			return fmt.Errorf("failed to read environment file: %w", err)
		}
		sample = SampleFromEnv(commandName, env)
	} else {
		sample = NewSample(commandName)
	}

	content, err := sample.Bytes()
	if err != nil {
		return fmt.Errorf("failed to create sample config: %w", err)
	}

	if err := createFile(filename, content); err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("configuration file %q already exists", filename)
		}
		return fmt.Errorf("failed to create %q: %w", filename, err)
	}
	return nil
}

func createFile(name string, content []byte) error {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(content); err != nil {
		return err
	}
	return f.Close()
}
