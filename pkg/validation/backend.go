// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package validation

// Backend performs validation with real clusters.
type Backend struct{}

var _ Validation = &Backend{}

// Validate the environment. Must be called once before calling other functions.
func (b Backend) Validate(ctx Context) error {
	if err := detectDistro(ctx); err != nil {
		return err
	}
	if err := validateClusterset(ctx); err != nil {
		return err
	}
	return nil
}

// ApplicationNamespaces inspects the application DRPC and returns the application namespaces on the
// hub and managed clusters.
func (b Backend) ApplicationNamespaces(
	ctx Context,
	drpcName, drpcNamespace string,
) ([]string, error) {
	return drpcNamespaces(ctx, drpcName, drpcNamespace)
}
