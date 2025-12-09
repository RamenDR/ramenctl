// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

// package helpers provides testings helpers.
package helpers

import (
	"fmt"
	"os"
	"path/filepath"
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

// AddGatheredData adds fake gathered data to the output directory.
func AddGatheredData(t *testing.T, dataDir, name, commandName string) {
	testData := fmt.Sprintf("../testdata/%s/%s.data", name, commandName)
	source, err := filepath.Abs(testData)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(source, dataDir); err != nil {
		t.Fatal(err)
	}
}

func marshal(t *testing.T, obj any) string {
	if s, ok := obj.(string); ok {
		return s
	}
	return MarshalYAML(t, obj)
}
