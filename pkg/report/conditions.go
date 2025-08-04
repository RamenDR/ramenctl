// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

type ConditionStatus string

const (
	// Condition status is expected.
	ConditionOK = ConditionStatus("ok")

	// Condition generation does not match object generation.
	ConditionStale = ConditionStatus("stale")

	// Condition status is not expected.
	ConditionError = ConditionStatus("error")
)
