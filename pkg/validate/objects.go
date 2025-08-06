// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/ramendr/ramenctl/pkg/report"
)

func isDeleted(obj client.Object) bool {
	return !obj.GetDeletionTimestamp().IsZero()
}

// conditionStatus maps conditions status to report.ConditionStatus. This works only when the
// expected status is "True".
func conditionStatus(
	obj client.Object,
	condition *metav1.Condition,
	expected metav1.ConditionStatus,
) report.ConditionStatus {
	if condition.ObservedGeneration != obj.GetGeneration() {
		return report.ConditionStale
	}
	if condition.Status != expected {
		return report.ConditionError
	}
	return report.ConditionOK
}
