// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

// package helpers provides testings helpers.
package helpers

import (
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/time"
)

const (
	Modified = "modified"
)

func MarshalYAML(t *testing.T, a any) string {
	data, err := yaml.Marshal(a)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func FakeTime(t *testing.T) {
	fakeTime := time.Now()
	savedNow := time.Now
	time.Now = func() time.Time {
		return fakeTime
	}
	t.Cleanup(func() {
		time.Now = savedNow
	})
}
