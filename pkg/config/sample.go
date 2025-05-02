// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	_ "embed"
	"text/template"
)

//go:embed sample.yaml
var sampleConfig string

type Sample struct {
	CommandName            string
	HubKubeconfig          string
	PrimaryKubeconfig      string
	SecondaryKubeconfig    string
	RBDStorageClassName    string
	CephFSStorageClassName string
}

func NewSample(commandName string) *Sample {
	sample := &Sample{
		CommandName:         commandName,
		HubKubeconfig:       "hub/config",
		PrimaryKubeconfig:   "primary/config",
		SecondaryKubeconfig: "secondary/config",
	}

	// When running as `odf dr init` we optimize for ODF cluster.
	// TODO: look up available storage classes in the cluster and let the user choose?
	if commandName == "odf dr" {
		sample.RBDStorageClassName = "ocs-storagecluster-ceph-rbd"
		sample.CephFSStorageClassName = "ocs-storagecluster-cephfs"
	} else {
		sample.RBDStorageClassName = "rook-ceph-block"
		sample.CephFSStorageClassName = "rook-cephfs-fs1"
	}

	return sample
}

func SampleFromEnv(commandName string, env *EnvFile) *Sample {
	// Using drenv envfile: use drenv storage classes.
	return &Sample{
		CommandName:         commandName,
		HubKubeconfig:       env.MustGetKubeCfgPath(env.Ramen.Hub),
		PrimaryKubeconfig:   env.MustGetKubeCfgPath(env.Ramen.Clusters[0]),
		SecondaryKubeconfig: env.MustGetKubeCfgPath(env.Ramen.Clusters[1]),

		// TODO: Get the info from the envfile instead of hard-coding.
		RBDStorageClassName:    "rook-ceph-block",
		CephFSStorageClassName: "rook-cephfs-fs1",
	}
}

func (s *Sample) Bytes() ([]byte, error) {
	t, err := template.New("sample").Parse(sampleConfig)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
