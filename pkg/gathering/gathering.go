// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package gathering

import (
	"context"
	"path/filepath"
	"sync"
	"time"

	"github.com/nirs/kubectl-gather/pkg/gather"
	"github.com/ramendr/ramen/e2e/types"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Result struct {
	Name     string
	Err      error
	Duration float64
}

type Options struct {
	Namespaces []string
	OutputDir  string
}

// Context provides logging and Go context access for gathering operations.
type Context interface {
	Logger() *zap.SugaredLogger
	Context() context.Context
}

// OutputReader is the interface for reading gathered data from the output directory.
type OutputReader interface {
	ListResources(namespace, resource string) ([]string, error)
	ReadResource(namespace, resource, name string) ([]byte, error)
}

// Namespaces gathers namespaces from all clusters storing data in outputDir. Returns a channel for
// getting gather results. The channel is closed when all clusters are gathered.
func Namespaces(ctx Context, clusters []*types.Cluster, options Options) <-chan Result {
	results := make(chan Result)
	var wg sync.WaitGroup

	// Start gathering in parallel for all clusters.
	for _, cluster := range clusters {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			err := gatherCluster(ctx, cluster, options)
			results <- Result{Name: cluster.Name, Err: err, Duration: time.Since(start).Seconds()}
		}()
	}

	// Close results channel when done to make client code nicer.
	go func() {
		wg.Wait()
		close(results)
	}()

	return results
}

func gatherCluster(ctx Context, cluster *types.Cluster, options Options) error {
	start := time.Now()
	log := ctx.Logger()

	log.Infof("Gathering from cluster %q", cluster.Name)

	config, err := restConfig(cluster.Kubeconfig)
	if err != nil {
		return err
	}

	clusterDir := filepath.Join(options.OutputDir, cluster.Name)
	gatherOptions := gather.Options{
		Kubeconfig: cluster.Kubeconfig,
		Namespaces: options.Namespaces,
		Log:        log.Named(cluster.Name),
	}

	g, err := gather.New(config, clusterDir, gatherOptions)
	if err != nil {
		return err
	}

	if err := g.Gather(); err != nil {
		return err
	}

	log.Infof("Gathered %d resources from cluster %q in %.2f seconds",
		g.Count(), cluster.Name, time.Since(start).Seconds())

	return nil
}

func restConfig(kubeconfig string) (*rest.Config, error) {
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return nil, err
	}

	return clientcmd.NewNonInteractiveClientConfig(*config, "", nil, nil).ClientConfig()
}
