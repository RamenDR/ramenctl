// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"fmt"
	"io"

	"sigs.k8s.io/yaml"
)

// WriteYAML writes YAML for any report to the writer.
func WriteYAML(w io.Writer, report any) error {
	data, err := yaml.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	_, err = w.Write(data)
	return err
}
