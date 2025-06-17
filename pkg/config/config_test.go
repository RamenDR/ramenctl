// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package config_test

import (
	"reflect"
	"testing"

	e2econfig "github.com/ramendr/ramen/e2e/config"

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

func testConfig() *config.Config {
	return &config.Config{
		Clusters: map[string]e2econfig.Cluster{
			"hub": {Kubeconfig: "hub/config"},
			"c1":  {Kubeconfig: "dr1/config"},
			"c2":  {Kubeconfig: "dr2/config"},
		},
		ClusterSet: "default",
	}
}

func TestReadConfigGeneric(t *testing.T) {
	// We read the same config from full test config or the simplified test config.
	c, err := config.ReadConfig("testdata/generic.yaml")
	if err != nil {
		t.Fatal(err)
	}
	expected := testConfig()
	if !c.Equal(expected) {
		t.Fatalf("expected %+v, got %+v", expected, c)
	}
}

func TestReadConfigTest(t *testing.T) {
	// We read the same config from full test config or the simplified test config.
	c1, err := config.ReadConfig("testdata/test.yaml")
	if err != nil {
		t.Fatal(err)
	}
	c2, err := config.ReadConfig("testdata/generic.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !c1.Equal(c2) {
		t.Fatalf("expected %+v, got %+v", c1, c2)
	}
}

func TestConfigEqual(t *testing.T) {
	c1 := testConfig()
	t.Run("equal to itself", func(t *testing.T) {
		c2 := c1
		if !c1.Equal(c2) {
			t.Errorf("config %+v is not equal to itself", c1)
		}
	})
	t.Run("equal to other identical config", func(t *testing.T) {
		c2 := testConfig()
		if !c1.Equal(c2) {
			t.Errorf("config %+v is not equal to other identical config %+v", c1, c2)
		}
	})
}

func TestConfigNotEqual(t *testing.T) {
	c1 := testConfig()
	t.Run("distro", func(t *testing.T) {
		c2 := testConfig()
		c2.Distro = "modified"
		if c1.Equal(c2) {
			t.Fatalf("config %+v non equal config %+v", c1, c2)
		}
	})
	t.Run("clusterset", func(t *testing.T) {
		c2 := testConfig()
		c2.ClusterSet = "modified"
		if c1.Equal(c2) {
			t.Fatalf("config %+v non equal config %+v", c1, c2)
		}
	})
	t.Run("clusters", func(t *testing.T) {
		c2 := testConfig()
		c2.Clusters["c2"] = e2econfig.Cluster{Kubeconfig: "modified"}
		if c1.Equal(c2) {
			t.Fatalf("config %+v non equal config %+v", c1, c2)
		}
	})
}
