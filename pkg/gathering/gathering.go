// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package gathering

import (
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

// Namespaces gathers namespaces from all clusters storing data in outputDir. Returns a channel for
// getting gather results. The channel is closed when all clusters are gathered.
func Namespaces(
	clusters []*types.Cluster,
	namespaces []string,
	outputDir string,
	log *zap.SugaredLogger,
) <-chan Result {
	results := make(chan Result)
	var wg sync.WaitGroup

	// Start gathering in parallel for all clusters.
	for _, cluster := range clusters {
		wg.Add(1)
		go func() {
			defer wg.Done()
			start := time.Now()
			err := gatherCluster(cluster, namespaces, outputDir, log)
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

func gatherCluster(
	cluster *types.Cluster,
	namespaces []string,
	outputDir string,
	log *zap.SugaredLogger,
) error {
	start := time.Now()
	log.Infof("Gather namespaces from cluster %q", cluster.Name)

	config, err := restConfig(cluster.Kubeconfig)
	if err != nil {
		return err
	}

	clusterDir := filepath.Join(outputDir, cluster.Name)
	options := gather.Options{
		Kubeconfig: cluster.Kubeconfig,
		Namespaces: namespaces,
		Log:        log.Named(cluster.Name),
	}

	g, err := gather.New(config, clusterDir, options)
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
