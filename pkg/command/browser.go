// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/ramendr/ramenctl/pkg/console"
)

// BrowseReport opens the HTML report in the default browser. If the report cannot be opened, it
// logs a notice suggesting to open the file manually to view the report.
func (c *Command) BrowseReport() {
	path := filepath.Join(c.outputDir, c.name+".html")

	if _, err := os.Stat(path); err != nil {
		c.log.Warnf("Skipping browse %q: %s", path, err)
		return
	}

	var err error
	var hint string

	switch runtime.GOOS {
	case "darwin":
		// https://keith.github.io/xcode-man-pages/open.1.html
		err = run("open", path)
	case "windows":
		// https://learn.microsoft.com/en-us/windows-server/administration/windows-commands/start
		err = run("cmd", "/c", "start", "", path)
	default:
		// Other platforms (linux, *bsd, etc.) - may not have xdg-open, but contributors
		// can add special cases above for their platform.
		// https://portland.freedesktop.org/doc/xdg-open.html
		if _, e := exec.LookPath("xdg-open"); e != nil {
			err = e
			hint = "Install xdg-open to open the report automatically"
		} else {
			err = run("xdg-open", path)
		}
	}

	if err != nil {
		c.log.Warnf("Cannot open %q in browser: %s", path, err)
		console.Info("Cannot open %q, please open it in a browser to view the report", path)
		if hint != "" {
			console.Hint(hint)
		}
	}
}

func run(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%w: %s", err, stderr.String())
	}
	return nil
}
