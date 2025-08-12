// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report_test

import (
	"testing"

	"sigs.k8s.io/yaml"
)

func marshal(t *testing.T, a any) string {
	data, err := yaml.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
