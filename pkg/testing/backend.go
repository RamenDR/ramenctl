// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package testing

import (
	"github.com/ramendr/ramen/e2e/dractions"
	"github.com/ramendr/ramen/e2e/types"
	"github.com/ramendr/ramen/e2e/util"
	"github.com/ramendr/ramen/e2e/validate"
	corev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"

	"github.com/ramendr/ramenctl/pkg/gathering"
	"github.com/ramendr/ramenctl/pkg/s3"
)

type Backend struct{}

var _ Testing = &Backend{}

func (Backend) Validate(ctx types.Context) error {
	return validate.TestConfig(ctx)
}

func (Backend) Setup(ctx types.Context) error {
	return util.EnsureChannel(ctx)
}

func (Backend) Cleanup(ctx types.Context) error {
	return util.EnsureChannelDeleted(ctx)
}

func (Backend) Deploy(ctx types.TestContext) error {
	return ctx.Deployer().Deploy(ctx)
}

func (Backend) Undeploy(ctx types.TestContext) error {
	return ctx.Deployer().Undeploy(ctx)
}

func (Backend) Protect(ctx types.TestContext) error {
	return dractions.EnableProtection(ctx)
}

func (Backend) Unprotect(ctx types.TestContext) error {
	return dractions.DisableProtection(ctx)
}

func (Backend) Failover(ctx types.TestContext) error {
	return dractions.Failover(ctx)
}

func (Backend) Relocate(ctx types.TestContext) error {
	return dractions.Relocate(ctx)
}

func (Backend) Purge(ctx types.TestContext) error {
	return dractions.Purge(ctx)
}

func (b Backend) Gather(
	ctx types.Context,
	clusters []*types.Cluster,
	options gathering.Options,
) <-chan gathering.Result {
	return gathering.Namespaces(ctx, clusters, options)
}

func (b Backend) GetSecret(
	ctx types.Context,
	cluster *types.Cluster,
	name, namespace string,
) (*corev1.Secret, error) {
	secret := &corev1.Secret{}
	key := k8stypes.NamespacedName{Namespace: namespace, Name: name}
	if err := cluster.Client.Get(ctx.Context(), key, secret); err != nil {
		return nil, err
	}
	return secret, nil
}

func (b Backend) GatherS3(
	ctx types.Context,
	profiles []*s3.Profile,
	prefixes []string,
	outputDir string,
) <-chan s3.Result {
	return s3.Gather(ctx.Context(), profiles, prefixes, outputDir, ctx.Logger())
}
