// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report_test

import (
	"reflect"
	"runtime"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/build"
	"github.com/ramendr/ramenctl/pkg/report"
)

func TestHost(t *testing.T) {
	r := report.New()
	expected := report.Host{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
		Cpus: runtime.NumCPU(),
	}
	if !reflect.DeepEqual(r.Host, expected) {
		t.Fatalf("expected host %+v, got %+v", expected, r.Host)
	}
}

func TestBuildInfo(t *testing.T) {
	savedVersion := build.Version
	savedCommit := build.Commit
	defer func() {
		build.Version = savedVersion
		build.Commit = savedCommit
	}()
	t.Run("available", func(t *testing.T) {
		build.Version = "fake-version"
		build.Commit = "fake-commit"
		r := report.New()
		if r.Ramenctl == nil {
			t.Fatalf("ramenctl omitted")
		}
		expected := &report.Ramenctl{
			Version: build.Version,
			Commit:  build.Commit,
		}
		if !reflect.DeepEqual(r.Ramenctl, expected) {
			t.Fatalf("expected ramenctl %+v, got %+v", expected, r.Ramenctl)
		}
	})
	t.Run("missing", func(t *testing.T) {
		build.Version = ""
		build.Commit = ""
		r := report.New()
		if r.Ramenctl != nil {
			t.Fatalf("ramenctl not omitted: %+v", r.Ramenctl)
		}
	})
}

func TestRoundtrip(t *testing.T) {
	r1 := report.New()
	b, err := yaml.Marshal(r1)
	if err != nil {
		t.Fatalf("failed to marshal yaml: %s", err)
	}
	r2 := &report.Report{}
	if err := yaml.Unmarshal(b, r2); err != nil {
		t.Fatalf("failed to unmarshal yaml: %s", err)
	}
	if !reflect.DeepEqual(r1, r2) {
		t.Fatalf("expected report %+v, got %+v", r1, r2)
	}
}
