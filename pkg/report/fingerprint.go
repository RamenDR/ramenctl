// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import (
	"crypto/sha256"
	"encoding/pem"
	"fmt"
	"strings"
)

// Fingerprint returns the colon separated SHA-256 fingerprint of data.
func Fingerprint(data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("empty input data")
	}

	sum := sha256.Sum256(data)

	parts := make([]string, len(sum))
	for i, b := range sum {
		parts[i] = fmt.Sprintf("%02X", b)
	}

	return strings.Join(parts, ":"), nil
}

// CertificateFingerprint returns the SHA-256 fingerprint of a PEM certificate.
func CertificateFingerprint(certPem []byte) (string, error) {
	block, _ := pem.Decode(certPem)
	if block == nil || block.Type != "CERTIFICATE" {
		return "", fmt.Errorf("failed to decode PEM, not a valid certificate")
	}

	return Fingerprint(block.Bytes)
}
