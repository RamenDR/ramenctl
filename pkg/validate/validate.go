// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"fmt"
	"os"

	recipev1alpha1 "github.com/ramendr/recipe/api/v1alpha1"
	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/console"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/validate/application"
	"github.com/ramendr/ramenctl/pkg/validate/clusters"
	"github.com/ramendr/ramenctl/pkg/validate/recipe"
	"github.com/ramendr/ramenctl/pkg/validation"
)

func Clusters(opts command.Options) error {
	cfg, err := config.ReadConfig(opts.ConfigFile)
	if err != nil {
		return console.Failed(err)
	}

	cmd, err := command.New(clusters.CommandName, cfg.Clusters, opts)
	if err != nil {
		return console.Failed(err)
	}
	defer cmd.Close()

	validate := clusters.NewCommand(cmd, cfg, validation.Backend{})

	var failed error
	if err := validate.Run(); err != nil {
		failed = console.Failed(err)
	}

	if opts.Interactive {
		cmd.BrowseReport()
	}

	return failed
}

func Application(opts command.ApplicationOptions) error {
	cfg, err := config.ReadConfig(opts.ConfigFile)
	if err != nil {
		return console.Failed(err)
	}

	cmd, err := command.New(application.CommandName, cfg.Clusters, opts.Options)
	if err != nil {
		return console.Failed(err)
	}
	defer cmd.Close()

	validate := application.NewCommand(cmd, cfg, validation.Backend{}, opts)

	var failed error
	if err := validate.Run(); err != nil {
		failed = console.Failed(err)
	}

	if opts.Interactive {
		cmd.BrowseReport()
	}

	return failed
}

// Recipe reads the Recipe YAML at path, runs offline semantic checks, and writes
// the recipe validation status as YAML to stdout. It returns a non-nil error when
// any problem-level issue is found or when the file cannot be read or parsed.
func Recipe(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read %q: %w", path, err)
	}

	r := &recipev1alpha1.Recipe{}
	if err := yaml.Unmarshal(data, r); err != nil {
		return fmt.Errorf("failed to parse %q: %w", path, err)
	}

	status := recipe.Validate(r)

	out, err := yaml.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal recipe status: %w", err)
	}
	fmt.Print(string(out))

	for _, w := range status.Workflows {
		if w.State == report.Problem {
			return fmt.Errorf("recipe %q has problems", path)
		}
	}

	return nil
}
