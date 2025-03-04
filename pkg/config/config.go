// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"text/template"
)

var sampleConfig = `## {{.CommandName}} configuration file

## Clusters configuration.
# - Modify clusters "kubeconfigpath" and "name" to match your hub and managed
#   clusters names and path to the kubeconfig file.
clusters:
  hub:
    name: hub
    kubeconfigpath: hub/config
  c1:
    name: primary
    kubeconfigpath: primary/config
  c2:
    name: secondary
    kubeconfigpath: secondary/config

## Git repository for test command.
# - Modify "url" to use your own Git repository.
# - Modify "branch" to test a different branch.
repo:
  url: https://github.com/RamenDR/ocm-ramen-samples.git
  branch: main

## DRPolicy for test command.
# - Modify to match actual DRPolicy in the hub cluster.
drpolicy: dr-policy

## ClusterSet for test command".
# - Modify to match your Open Cluster Management configuration.
clusterset: default

## PVC specifications for test command.
# - Modify items "storageclassname" to match the actual storage classes in the
#   managed clusters.
# - Add new items for testing more storage types.
pvcspecs:
- name: rbd
  storageclassname: rook-ceph-block
  accessmodes: ReadWriteOnce
- name: cephfs
  storageclassname: rook-cephfs-fs1
  accessmodes: ReadWriteMany

## Tests cases for test command.
# - Modify the test for your preferred workload or deployment type.
# - Add new tests for testing more combinations in parallel.
# - Available workloads: deploy
# - Available deployers: appset, subscr, disapp
tests:
- workload: deploy
  deployer: appset
  pvcspec: rbd
`

func CreateSampleConfig(filename, commandName string) error {
	sample := Sample{CommandName: commandName}
	content, err := sample.Bytes()
	if err != nil {
		return fmt.Errorf("failed to create sample config: %w", err)
	}
	if err := createFile(filename, content); err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("configuration file %q already exists", filename)
		}
		return fmt.Errorf("failed to create %q: %w", filename, err)
	}
	return nil
}

type Sample struct {
	CommandName string
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

func createFile(name string, content []byte) error {
	f, err := os.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(content); err != nil {
		return err
	}
	return f.Close()
}
