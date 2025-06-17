// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"reflect"
	"testing"

	"github.com/ramendr/ramenctl/pkg/config"
)

func TestSample(t *testing.T) {
	sample := config.NewSample("ramenctl")
	expected := &config.Sample{
		CommandName:            "ramenctl",
		HubKubeconfig:          "hub/config",
		PrimaryKubeconfig:      "primary/config",
		SecondaryKubeconfig:    "secondary/config",
		RBDStorageClassName:    "rook-ceph-block",
		CephFSStorageClassName: "rook-cephfs-fs1",
	}
	if !reflect.DeepEqual(expected, sample) {
		t.Fatalf("expected %+v, got %+v", expected, sample)
	}
}

func TestSampleForODF(t *testing.T) {
	sample := config.NewSample("odf dr")
	expected := &config.Sample{
		CommandName:            "odf dr",
		HubKubeconfig:          "hub/config",
		PrimaryKubeconfig:      "primary/config",
		SecondaryKubeconfig:    "secondary/config",
		RBDStorageClassName:    "ocs-storagecluster-ceph-rbd",
		CephFSStorageClassName: "ocs-storagecluster-cephfs",
	}
	if !reflect.DeepEqual(expected, sample) {
		t.Fatalf("expected %+v, got %+v", expected, sample)
	}
}

func TestSampleFromEnv(t *testing.T) {
	env := &config.EnvFile{
		Name: "rdr",
		Ramen: config.Ramen{
			Hub:      "hub",
			Clusters: []string{"dr1", "dr2"},
		},
	}
	sample := config.SampleFromEnv("ramenctl", env)
	expected := &config.Sample{
		CommandName:            "ramenctl",
		HubKubeconfig:          env.KubeconfigPath("hub"),
		PrimaryKubeconfig:      env.KubeconfigPath("dr1"),
		SecondaryKubeconfig:    env.KubeconfigPath("dr2"),
		RBDStorageClassName:    "rook-ceph-block",
		CephFSStorageClassName: "rook-cephfs-fs1",
	}
	if !reflect.DeepEqual(expected, sample) {
		t.Fatalf("expected %+v, got %+v", expected, sample)
	}
}
