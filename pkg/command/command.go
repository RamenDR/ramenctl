// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"go.uber.org/zap"
	"sigs.k8s.io/yaml"

	e2eenv "github.com/ramendr/ramen/e2e/env"
	"github.com/ramendr/ramen/e2e/types"

	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
)

// Command is a ramenctl command implementing the ramen/e2e/types.Context interface.
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
	log      *zap.SugaredLogger
	closeLog func()

	// context and stop are used for cancellation.
	context context.Context
	stop    context.CancelFunc
}

var _ types.Context = &Command{}

// New creates a new command handling os.Interrupt signal. To close the log and stop the signal handler call Close().
func New(commandName, configFile, outputDir string) (*Command, error) {
	// Create the logger first so we can log early command errors to the command log.
	log, closeLog, err := newLogger(outputDir, commandName+".log")
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

	// Create the context before creating the env so we can cancel the command cleanly if accessing the clusters block
	// for long time. The log will contain the cancelation error.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	env, err := e2eenv.New(ctx, cfg, log)
	if err != nil {
		// Stop the signal handler before we fail.
		stop()
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
		closeLog:  closeLog,
		context:   ctx,
		stop:      stop,
	}, nil
}

// ForTest is a command configured for testing without real clusters. This command does not handle signals and its
// context cannot be cancelled.
func ForTest(commandName string, cfg *types.Config, env *types.Env, outputDir string) (*Command, error) {
	log, closeLog, err := newLogger(outputDir, commandName+".log")
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}
	return &Command{
		name:      commandName,
		outputDir: outputDir,
		config:    cfg,
		env:       env,
		log:       log,
		closeLog:  closeLog,
		context:   context.Background(),
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

func (c *Command) Context() context.Context {
	return c.context
}

// Close log and stop handling signals and mark the command context as done. Calling while a command is running will
// cancel the command.
func (c *Command) Close() {
	if c.stop != nil {
		c.stop()
	}
	_ = c.log.Sync()
	c.closeLog()
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
