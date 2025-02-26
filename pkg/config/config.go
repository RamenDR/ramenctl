// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"os"
)

var sampleConfig = `# %s configuration file
---
# Sample cluster configurations:
# Uncomment and edit the following lines to provide the kubeconfig paths
# for your clusters.
# clusters:
#   hub:
#     kubeconfigpath: /kubeconfigs/hub
#   c1:
#     kubeconfigpath: /kubeconfigs/c1
#   c2:
#     kubeconfigpath: /kubeconfigs/c2

# List of PVC specifications for workloads.
# These define storage configurations, such as 'storageClassName' and
# 'accessModes', and are used to kustomize workloads.
pvcspecs:
- name: rbd
  storageclassname: rook-ceph-block
  accessmodes: ReadWriteOnce
- name: cephfs
  storageclassname: rook-cephfs-fs1
  accessmodes: ReadWriteMany
`

func CreateSampleConfig(filename, creator string) error {
	content := fmt.Sprintf(sampleConfig, creator)
	if err := createFile(filename, []byte(content)); err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("configuration file %q already exists", filename)
		}
		return fmt.Errorf("failed to create %q: %w", filename, err)
	}
	return nil
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
