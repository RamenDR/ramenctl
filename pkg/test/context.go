// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/ramendr/ramen/e2e/types"
)

// Context implements types.Context interface.
type Context struct {
	workload types.Workload
	deployer types.Deployer
	name     string
	logger   *zap.SugaredLogger
	cmd      *Command
}

var _ types.Context = &Context{}

func newContext(workload types.Workload, deployer types.Deployer, cmd *Command) *Context {
	name := fmt.Sprintf("%s-%s", deployer.GetName(), workload.GetName())
	return &Context{
		workload: workload,
		deployer: deployer,
		name:     name,
		logger:   cmd.Logger().Named(name),
		cmd:      cmd,
	}
}

func (c *Context) Deployer() types.Deployer {
	return c.deployer
}

func (c *Context) Workload() types.Workload {
	return c.workload
}

func (c *Context) Name() string {
	return c.name
}

func (c *Context) ManagementNamespace() string {
	if ns := c.deployer.GetNamespace(c); ns != "" {
		return ns
	}
	return c.AppNamespace()
}

func (c *Context) AppNamespace() string {
	return c.cmd.NamespacePrefix + c.name
}

func (c *Context) Logger() *zap.SugaredLogger {
	return c.logger
}

func (c *Context) Env() *types.Env {
	return c.cmd.Env()
}

func (c *Context) Config() *types.Config {
	return c.cmd.Config()
}
