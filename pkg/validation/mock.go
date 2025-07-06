// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validation

type ContextFunc func(Context) error

// Mock implements the Validation interface. All operations succeed without accessing the
// clusters. To cause operations to fail, set a function returning an error.
type Mock struct {
	ValidateFunc ContextFunc
}

var _ Validation = &Mock{}

func (m *Mock) Validate(ctx Context) error {
	if m.ValidateFunc != nil {
		return m.ValidateFunc(ctx)
	}
	return nil
}
