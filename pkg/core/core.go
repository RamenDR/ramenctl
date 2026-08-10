// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package core

import (
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/gathering"
)

const (
	pvcPlural       = "persistentvolumeclaims"
	configMapPlural = "configmaps"
	secretPlural    = "secrets"
)

func ReadPVC(
	reader gathering.OutputReader,
	name, namespace string,
) (*v1.PersistentVolumeClaim, error) {
	data, err := reader.ReadResource(namespace, pvcPlural, name)
	if err != nil {
		return nil, err
	}
	pvc := &v1.PersistentVolumeClaim{}
	if err := yaml.Unmarshal(data, pvc); err != nil {
		return nil, err
	}
	return pvc, nil
}

func ReadConfigMap(
	reader gathering.OutputReader,
	name, namespace string,
) (*v1.ConfigMap, error) {
	data, err := reader.ReadResource(namespace, configMapPlural, name)
	if err != nil {
		return nil, err
	}
	configMap := &v1.ConfigMap{}
	if err := yaml.Unmarshal(data, configMap); err != nil {
		return nil, err
	}
	return configMap, nil
}

func ReadSecret(
	reader gathering.OutputReader,
	name, namespace string,
) (*v1.Secret, error) {
	data, err := reader.ReadResource(namespace, secretPlural, name)
	if err != nil {
		return nil, err
	}
	secret := &v1.Secret{}
	if err := yaml.Unmarshal(data, secret); err != nil {
		return nil, err
	}
	return secret, nil
}
