// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validate

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/time"
)

// Command common tests

func TestValidatedDeleted(t *testing.T) {
	cmd := testCommand(t, validateApplication, &helpers.ValidationMock{}, testK8s)

	t.Run("nil", func(t *testing.T) {
		validated := cmd.validatedDeleted(nil)
		expected := report.ValidatedBool{
			Value: true,
			Validated: report.Validated{
				State:       report.Problem,
				Description: resourceDoesNotExist,
			},
		}
		if validated != expected {
			t.Fatalf("expected %v, got %v", expected, validated)
		}
	})
	t.Run("object deleted", func(t *testing.T) {
		deletedPVC := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				DeletionTimestamp: &metav1.Time{Time: time.Now()},
			},
		}
		validated := cmd.validatedDeleted(deletedPVC)
		expected := report.ValidatedBool{
			Value: true,
			Validated: report.Validated{
				State:       report.Problem,
				Description: resourceWasDeleted,
			},
		}
		if validated != expected {
			t.Fatalf("expected %v, got %v", expected, validated)
		}
	})
	t.Run("object not deleted", func(t *testing.T) {
		pvc := &corev1.PersistentVolumeClaim{}
		validated := cmd.validatedDeleted(pvc)
		expected := report.ValidatedBool{
			Validated: report.Validated{
				State: report.OK,
			},
		}
		if validated != expected {
			t.Fatalf("expected %v, got %v", expected, validated)
		}
	})

	t.Run("update summary", func(t *testing.T) {
		expected := report.Summary{OK: 1, Problem: 2}
		if !cmd.report.Summary.Equal(&expected) {
			t.Fatalf("expected summary %v, got %v", expected, *cmd.report.Summary)
		}
	})
}

// TODO: Test gather cancellation when kubectl-gahter supports it:
// https://github.com/nirs/kubectl-gather/issues/88
