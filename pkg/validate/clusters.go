// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import "github.com/ramendr/ramenctl/pkg/console"

func Clusters(outputDir string) error {
	console.Info("Using report %q", outputDir)
	return nil
}
