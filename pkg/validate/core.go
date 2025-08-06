// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/gathering"
)

const (
	pvcPlural = "persistentvolumeclaims"
)

func readPVC(
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
