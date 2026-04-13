// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/report"
	validatecmd "github.com/ramendr/ramenctl/pkg/validate/command"
	"github.com/ramendr/ramenctl/pkg/validation"
)

type Command struct {
	*validatecmd.Command
}

func newCommand(cmd *command.Command, cfg *config.Config, backend validation.Validation) *Command {
	r := report.NewReport(cmd.Name(), cfg)
	r.Summary = &report.Summary{}
	return &Command{
		Command: validatecmd.New(cmd, cfg, backend, r),
	}
}
