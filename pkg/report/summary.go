// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package report

import "maps"

// SummaryKey is a typed key for Summary counters.
// Each package defines its own keys (e.g., validate.OK, test.Passed).
type SummaryKey string

// Summary is a counter for report summaries.
type Summary map[SummaryKey]int

// Add increments the counter for the given key.
func (s Summary) Add(key SummaryKey) {
	s[key]++
}

// Get returns the count for the given key.
func (s Summary) Get(key SummaryKey) int {
	return s[key]
}

// Equal returns true if both summaries are equal.
// Handles nil pointers: two nil summaries are equal, nil and non-nil are not.
func (s *Summary) Equal(o *Summary) bool {
	if s == o {
		return true
	}
	if s == nil || o == nil {
		return false
	}
	return maps.Equal(*s, *o)
}
