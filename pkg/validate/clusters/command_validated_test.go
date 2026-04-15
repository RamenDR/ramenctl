// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package clusters

import (
	"fmt"
	"testing"

	ramenapi "github.com/ramendr/ramen/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/ramen"
	"github.com/ramendr/ramenctl/pkg/report"
	"github.com/ramendr/ramenctl/pkg/validate/summary"
)

// Test individual cluster validation functions without running the full
// command flow.

// Modern ramen: controller type is set as env var in the deployment,
// configmap has empty controller type (zero value).
func TestValidateControllerTypeModern(t *testing.T) {
	cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

	deployment := testModernDeployment(string(ramenapi.DRHubType))
	configMap := testConfigMap("")

	s := &report.DeploymentSummary{}
	if err := cmd.validateControllerType(s, deployment, configMap, ramenapi.DRHubType); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := report.ValidatedString{
		Value:     string(ramenapi.DRHubType),
		Validated: report.Validated{State: report.OK},
	}
	if s.RamenControllerType != expected {
		t.Errorf("expected %+v, got %+v", expected, s.RamenControllerType)
	}

	expectedSummary := report.Summary{summary.OK: 1}
	if !cmd.Report.Summary.Equal(&expectedSummary) {
		t.Errorf("expected summary %v, got %v", expectedSummary, *cmd.Report.Summary)
	}
}

// Legacy ramen: no env var in the deployment, controller type is set in the configmap.
func TestValidateControllerTypeLegacy(t *testing.T) {
	cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

	deployment := testLegacyDeployment()
	configMap := testConfigMap(string(ramenapi.DRHubType))

	s := &report.DeploymentSummary{}
	if err := cmd.validateControllerType(s, deployment, configMap, ramenapi.DRHubType); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := report.ValidatedString{
		Value:     string(ramenapi.DRHubType),
		Validated: report.Validated{State: report.OK},
	}
	if s.RamenControllerType != expected {
		t.Errorf("expected %+v, got %+v", expected, s.RamenControllerType)
	}

	expectedSummary := report.Summary{summary.OK: 1}
	if !cmd.Report.Summary.Equal(&expectedSummary) {
		t.Errorf("expected summary %v, got %v", expectedSummary, *cmd.Report.Summary)
	}
}

// Nil deployment: take value from configmap.
func TestValidateControllerTypeNilDeployment(t *testing.T) {
	cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

	configMap := testConfigMap(string(ramenapi.DRHubType))

	s := &report.DeploymentSummary{}
	if err := cmd.validateControllerType(s, nil, configMap, ramenapi.DRHubType); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := report.ValidatedString{
		Value:     string(ramenapi.DRHubType),
		Validated: report.Validated{State: report.OK},
	}
	if s.RamenControllerType != expected {
		t.Errorf("expected %+v, got %+v", expected, s.RamenControllerType)
	}

	expectedSummary := report.Summary{summary.OK: 1}
	if !cmd.Report.Summary.Equal(&expectedSummary) {
		t.Errorf("expected summary %v, got %v", expectedSummary, *cmd.Report.Summary)
	}
}

// Nil configmap: take value from deployment.
func TestValidateControllerTypeNilConfigMap(t *testing.T) {
	cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

	deployment := testModernDeployment(string(ramenapi.DRHubType))

	s := &report.DeploymentSummary{}
	if err := cmd.validateControllerType(s, deployment, nil, ramenapi.DRHubType); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := report.ValidatedString{
		Value:     string(ramenapi.DRHubType),
		Validated: report.Validated{State: report.OK},
	}
	if s.RamenControllerType != expected {
		t.Errorf("expected %+v, got %+v", expected, s.RamenControllerType)
	}

	expectedSummary := report.Summary{summary.OK: 1}
	if !cmd.Report.Summary.Equal(&expectedSummary) {
		t.Errorf("expected summary %v, got %v", expectedSummary, *cmd.Report.Summary)
	}
}

// Nil deployment and nil configmap: controller type is empty and reported as a problem.
func TestValidateControllerTypeNilDeploymentAndConfigMap(t *testing.T) {
	cmd := testCommand(t, &helpers.ValidationMock{}, testK8s)

	s := &report.DeploymentSummary{}
	if err := cmd.validateControllerType(s, nil, nil, ramenapi.DRHubType); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := report.ValidatedString{
		Value: "",
		Validated: report.Validated{
			State:       report.Problem,
			Description: fmt.Sprintf("Expecting controller type %q", ramenapi.DRHubType),
		},
	}
	if s.RamenControllerType != expected {
		t.Errorf("expected %+v, got %+v", expected, s.RamenControllerType)
	}

	expectedSummary := report.Summary{summary.Problem: 1}
	if !cmd.Report.Summary.Equal(&expectedSummary) {
		t.Errorf("expected summary %v, got %v", expectedSummary, *cmd.Report.Summary)
	}
}

func testModernDeployment(controllerType string) *appsv1.Deployment {
	return &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: ramen.ManagerContainerName,
							Env: []corev1.EnvVar{
								{Name: ramen.ControllerTypeEnvName, Value: controllerType},
							},
						},
					},
				},
			},
		},
	}
}

func testLegacyDeployment() *appsv1.Deployment {
	return &appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: ramen.ManagerContainerName,
						},
					},
				},
			},
		},
	}
}

func testConfigMap(controllerType string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		Data: map[string]string{
			ramen.ConfigMapRamenConfigKeyName: "ramenControllerType: " + controllerType + "\n",
		},
	}
}
