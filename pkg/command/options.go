// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package command

// Options shared by all commands except init.
type Options struct {
	ConfigFile string
	OutputDir  string
}

// ApplicationOptions shared by commands operating on a protected application.
type ApplicationOptions struct {
	Options
	DRPCName      string
	DRPCNamespace string
}
