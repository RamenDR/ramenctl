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

	// Base64 decoded secret values from testdata.
	// Both K8s and OCP testdata secrets use the same values.
	AccessKey = "this is not a real secret"
	//nolint:gosec
	SecretKey = "this is not a real key"

	// Invalid credentials for testing GetSecret failure scenarios.
	WrongAccessKey = "wrong-access-key"
	WrongSecretKey = "wrong-secret-key"

	// Secret key fingerprints (SHA-256 hashes) for testdata.
	// Both K8s and OCP testdata secrets have the same data values.
	AccessKeyFingerprint = "F3:1C:B8:5A:2C:33:BA:C3:57:84:22:D5:11:F5:35:40:FF:A8:6A:34:B8:CD:42:AC:86:65:E2:2B:E1:05:EA:23"
	//nolint:gosec
	SecretKeyFingerprint = "BC:42:FE:14:DB:F0:91:1C:91:1F:8F:CF:72:AF:CE:C5:83:5C:AF:93:AC:08:40:CE:31:D8:67:CA:AC:BC:E4:16"
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
