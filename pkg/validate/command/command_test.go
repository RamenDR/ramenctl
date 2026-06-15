// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"testing"

	"github.com/ramendr/ramen/e2e/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	basecmd "github.com/ramendr/ramenctl/pkg/command"
	"github.com/ramendr/ramenctl/pkg/config"
	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/time"
	"github.com/ramendr/ramenctl/pkg/validate/summary"
)

func TestValidatedDeleted(t *testing.T) {
	cmd := testCommand(t)

	t.Run("nil", func(t *testing.T) {
		validated := cmd.ValidatedDeleted(nil)
		expected := report.ValidatedBool{
			Value: true,
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Resource does not exist",
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
		validated := cmd.ValidatedDeleted(deletedPVC)
		expected := report.ValidatedBool{
			Value: true,
			Validated: report.Validated{
				State:       report.Problem,
				Description: "Resource was deleted",
			},
		}
		if validated != expected {
			t.Fatalf("expected %v, got %v", expected, validated)
		}
	})
	t.Run("object not deleted", func(t *testing.T) {
		pvc := &corev1.PersistentVolumeClaim{}
		validated := cmd.ValidatedDeleted(pvc)
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
		expected := report.Summary{summary.OK: 1, summary.Problem: 2}
		if !cmd.Report.Summary.Equal(&expected) {
			t.Fatalf("expected summary %v, got %v", expected, *cmd.Report.Summary)
		}
	})
}

// Helpers.

func testCommand(t *testing.T) *Command {
	helpers.FakeTime(t)
	env := &types.Env{
		Hub: &types.Cluster{Name: "hub"},
		C1:  &types.Cluster{Name: "dr1"},
		C2:  &types.Cluster{Name: "dr2"},
	}
	cmd, err := basecmd.ForTest("test", env, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cmd.Close()
	})
	r := report.NewReport("test", &config.Config{})
	r.Summary = &report.Summary{}
	return New(cmd, &config.Config{}, &helpers.ValidationMock{}, r)
}
