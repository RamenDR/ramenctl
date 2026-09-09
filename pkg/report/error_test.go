// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report_test

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"

	"github.com/ramendr/ramenctl/pkg/helpers"
	"github.com/ramendr/ramenctl/pkg/report"
)

func TestSetErrorDetailed(t *testing.T) {
	v := report.Validated{State: report.Problem}
	err := &report.DetailedError{
		Message:    "failed to download objects from profile \"s3profile\"",
		Reason:     "certificate signed by unknown authority",
		Suggestion: "configure CACertificates in the S3 store profile to trust the endpoint certificate",
		URL:        "https://ramendr.github.io/ramen/s3-store-profile/#cacertificates",
	}
	report.SetError(&v, err)

	expected := report.Validated{
		State:       report.Problem,
		Description: err.Message,
		Reason:      err.Reason,
		Suggestion:  err.Suggestion,
		URL:         err.URL,
	}
	if v != expected {
		t.Fatalf("unexpected validated\n%s", helpers.UnifiedDiff(t, expected, v))
	}
}

func TestSetErrorPlain(t *testing.T) {
	v := report.Validated{State: report.Problem}
	report.SetError(&v, errors.New("connection refused"))

	expected := report.Validated{
		State:       report.Problem,
		Description: "connection refused",
	}
	if v != expected {
		t.Fatalf("unexpected validated\n%s", helpers.UnifiedDiff(t, expected, v))
	}
}

func TestSetErrorWrappedDetailed(t *testing.T) {
	inner := &report.DetailedError{
		Message:    "failed to list objects",
		Reason:     "certificate signed by unknown authority",
		Suggestion: "configure CACertificates",
		URL:        "https://example.com",
	}
	v := report.Validated{State: report.Problem}
	report.SetError(&v, fmt.Errorf("outer: %w", inner))

	if v.Description != inner.Message {
		t.Fatalf("expected description %q, got %q", inner.Message, v.Description)
	}
	if v.Reason != inner.Reason || v.Suggestion != inner.Suggestion || v.URL != inner.URL {
		t.Fatalf("detailed fields not copied: %+v", v)
	}
}

func TestPromoteDetailed(t *testing.T) {
	inner := &report.DetailedError{
		Message:    "failed to list objects in bucket \"odrbucket\" with prefix \"app/\"",
		Reason:     "certificate signed by unknown authority",
		Suggestion: "configure CACertificates in the S3 store profile to trust the endpoint certificate",
		URL:        "https://ramendr.github.io/ramen/s3-store-profile/#cacertificates",
	}
	err := report.Promote("failed to download objects from profile \"s3profile\"", inner)

	var detailed *report.DetailedError
	if !errors.As(err, &detailed) {
		t.Fatalf("expected DetailedError, got %T: %v", err, err)
	}
	expected := &report.DetailedError{
		Message:    "failed to download objects from profile \"s3profile\"",
		Reason:     inner.Reason,
		Suggestion: inner.Suggestion,
		URL:        inner.URL,
	}
	if *detailed != *expected {
		t.Fatalf("unexpected error\n%s", helpers.UnifiedDiff(t, expected, detailed))
	}
}

func TestPromotePlain(t *testing.T) {
	err := report.Promote("failed to download objects from profile \"s3profile\"",
		errors.New("failed to list objects"))
	expected := "failed to download objects from profile \"s3profile\": failed to list objects"
	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
	var detailed *report.DetailedError
	if errors.As(err, &detailed) {
		t.Fatalf("unexpected DetailedError: %+v", detailed)
	}
}

func TestValidatedDetailedYAMLRoundtrip(t *testing.T) {
	original := report.Validated{
		State:       report.Problem,
		Description: "failed to download objects from profile \"s3profile\"",
		Reason:      "certificate signed by unknown authority",
		Suggestion:  "configure CACertificates in the S3 store profile to trust the endpoint certificate",
		URL:         "https://ramendr.github.io/ramen/s3-store-profile/#cacertificates",
	}

	data, err := yaml.Marshal(&original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{
		"description: failed to download objects from profile \"s3profile\"",
		"reason: certificate signed by unknown authority",
		"state: problem ❌",
		"suggestion:",
		"configure CACertificates in the S3 store profile to trust the endpoint",
		"url: https://ramendr.github.io/ramen/s3-store-profile/#cacertificates",
	} {
		if !strings.Contains(string(data), want) {
			t.Fatalf("yaml missing %q\n%s", want, data)
		}
	}

	var decoded report.Validated
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if decoded != original {
		t.Fatalf("roundtrip mismatch\n%s", helpers.UnifiedDiff(t, original, decoded))
	}
}

func TestValidatedTemplateDetailed(t *testing.T) {
	tmpl, err := report.Template()
	if err != nil {
		t.Fatalf("Template(): %v", err)
	}
	validated := &report.ValidatedBool{
		Validated: report.Validated{
			State:       report.Problem,
			Description: "failed to download objects from profile \"s3profile\"",
			Reason:      "certificate signed by unknown authority",
			Suggestion:  "configure CACertificates in the S3 store profile to trust the endpoint certificate",
			URL:         "https://ramendr.github.io/ramen/s3-store-profile/#cacertificates",
		},
		Value: false,
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "validated", validated); err != nil {
		t.Fatalf("ExecuteTemplate: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		`class="description">failed to download objects from profile &#34;s3profile&#34;`,
		`class="reason">certificate signed by unknown authority`,
		`href="https://ramendr.github.io/ramen/s3-store-profile/#cacertificates">configure CACertificates in the S3 store profile to trust the endpoint certificate</a>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n%s", want, out)
		}
	}
}
