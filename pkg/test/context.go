// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package test

import (
	"context"
	"fmt"
	stdtime "time"

	"go.uber.org/zap"

	e2econfig "github.com/ramendr/ramen/e2e/config"
	"github.com/ramendr/ramen/e2e/types"
)

// Context implements types.TestContext interface.
type Context struct {
	parent   types.Context
	context  context.Context
	workload types.Workload
	deployer types.Deployer
	name     string
	logger   *zap.SugaredLogger
}

var _ types.TestContext = &Context{}

func newContext(parent types.Context, workload types.Workload, deployer types.Deployer) *Context {
	name := fmt.Sprintf("%s-%s", deployer.GetName(), workload.GetName())
	return &Context{
		parent:   parent,
		context:  parent.Context(),
		workload: workload,
		deployer: deployer,
		name:     name,
		logger:   parent.Logger().Named(name),
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
	return namespacePrefix + c.name
}

func (c *Context) Logger() *zap.SugaredLogger {
	return c.logger
}

func (c *Context) Env() *types.Env {
	return c.parent.Env()
}

func (c *Context) Config() *e2econfig.Config {
	return c.parent.Config()
}

func (c *Context) Context() context.Context {
	return c.context
}

// WithTimeout returns a derived context with a deadline. Call cancel to release resources
// associated with the context as soon as the operation running in the context complete.
func (c Context) WithTimeout(d stdtime.Duration) (*Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(c.context, d)
	c.context = ctx
	return &c, cancel
}
