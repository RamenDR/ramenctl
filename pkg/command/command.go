// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"

	"go.uber.org/zap"

	e2eenv "github.com/ramendr/ramen/e2e/env"
	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
)

// Command is a ramenctl command.
type Command struct {
	// Name is the command name (e.g. "test-run")
	Name string
	// OutputDir contains the command log, summary, and gathered files.
	OutputDir string
	// Config loaded from configFile.
	Config *types.Config
	// Env loaded from the config.
	Env *types.Env
	// Logger logging to the command log.
	Logger *zap.SugaredLogger
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
		err := fmt.Errorf("failed to read config %q: %w", err)
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
		Name:      commandName,
		OutputDir: outputDir,
		Config:    cfg,
		Env:       env,
		Logger:    log,
	}, nil
}
