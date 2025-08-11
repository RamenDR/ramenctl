// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/ramendr/ramenctl/pkg/report"
)

func isDeleted(obj client.Object) bool {
	return !obj.GetDeletionTimestamp().IsZero()
}

func validatedCondition(
	obj client.Object,
	condition *metav1.Condition,
	expectedStatus metav1.ConditionStatus,
) report.ValidatedCondition {
	validated := report.ValidatedCondition{Type: condition.Type}

	if condition.ObservedGeneration != obj.GetGeneration() {
		validated.State = report.Stale
		validated.Description = fmt.Sprintf(
			"Observed generation %d does not match object generation %d",
			condition.ObservedGeneration,
			obj.GetGeneration(),
		)
		return validated
	}

	if condition.Status != expectedStatus {
		validated.State = report.Error
		validated.Description = condition.Message
		return validated
	}

	validated.State = report.OK
	return validated
}
