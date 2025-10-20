// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

// package helpers provides testings helpers.
package helpers

import (
	"testing"

	"github.com/aymanbagabas/go-udiff"
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

func UnifiedDiff(t *testing.T, expected, actual any) string {
	expectedString := marshal(t, expected)
	actualString := marshal(t, actual)
	return udiff.Unified("expected", "actual", expectedString, actualString)
}

func marshal(t *testing.T, obj any) string {
	if s, ok := obj.(string); ok {
		return s
	}
	return MarshalYAML(t, obj)
}
