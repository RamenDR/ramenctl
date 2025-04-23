// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"sigs.k8s.io/yaml"

	e2eenv "github.com/ramendr/ramen/e2e/env"
	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
)

// Command is a ramenctl command.
type Command struct {
	// name is the command name (e.g. "test-run")
	name string
	// outputDir contains the command log, summary, and gathered files.
	outputDir string
	// config loaded from configFile.
	config *types.Config
	// env loaded from the config.
	env *types.Env
	// log logging to the command log.
	log *zap.SugaredLogger
}

func New(commandName, configFile, outputDir string) (*Command, error) {
	// Create the logger first so we can log early command errors to the command log.
	log, err := newLogger(outputDir, commandName+".log")
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	console.Info("Using report %q", outputDir)

	cfg, err := config.ReadConfig(configFile)
	if err != nil {
		err := fmt.Errorf("failed to read config %q: %w", configFile, err)
		log.Error(err)
		return nil, err
	}

	console.Info("Using config %q", configFile)

	env, err := e2eenv.New(cfg, log)
	if err != nil {
		err := fmt.Errorf("failed to create env: %w", err)
		log.Error(err)
		return nil, err
	}

	return &Command{
		name:      commandName,
		outputDir: outputDir,
		config:    cfg,
		env:       env,
		log:       log,
	}, nil
}

func (c *Command) Name() string {
	return c.name
}

func (c *Command) OutputDir() string {
	return c.outputDir
}

// ramen/e2e/types.Context interface

func (c *Command) Logger() *zap.SugaredLogger {
	return c.log
}

func (c *Command) Config() *types.Config {
	return c.config
}

func (c *Command) Env() *types.Env {
	return c.env
}

// WriteReport writes report in yaml format to the command output directory.
func (c *Command) WriteReport(report any) error {
	data, err := yaml.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	path := filepath.Join(c.outputDir, c.name+".yaml")
	return os.WriteFile(path, data, 0o640)
}
