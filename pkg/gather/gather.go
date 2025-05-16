// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package gather

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/nirs/kubectl-gather/pkg/gather"
	"github.com/ramendr/ramen/e2e/types"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/ramendr/ramenctl/pkg/console"
)

// Namespaces gathers namespaces from all clusters storing data in outputDir.
func Namespaces(env *types.Env, namespaces []string, outputDir string, log *zap.SugaredLogger) {
	start := time.Now()
	log.Infof("Gather namespaces %q from all clusters", namespaces)

	var wg sync.WaitGroup
	for _, cluster := range []*types.Cluster{&env.Hub, &env.C1, &env.C2} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := gatherCluster(cluster, namespaces, outputDir, log); err != nil {
				msg := fmt.Sprintf("Failed to gather data from cluster %q", cluster.Name)
				console.Error(msg)
				log.Errorf("%s: %s", msg, err)
				return
			}
			console.Pass("Gathered data from cluster %q", cluster.Name)
		}()
	}
	wg.Wait()

	log.Infof("Gathered all clusters in %.2f seconds", time.Since(start).Seconds())
}

func gatherCluster(
	cluster *types.Cluster,
	namespaces []string,
	outputDir string,
	log *zap.SugaredLogger,
) error {
	start := time.Now()
	log.Infof("Gather namespaces from cluster %q", cluster.Name)

	config, err := restConfig(cluster, log)
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

func restConfig(cluster *types.Cluster, log *zap.SugaredLogger) (*rest.Config, error) {
	config, err := clientcmd.LoadFromFile(cluster.Kubeconfig)
	if err != nil {
		return nil, err
	}

	// On OpenShift platforms, kubeconfig files can have long or generic context names
	// (e.g., "system:admin/https://api-cluster-example-com:6443" or simply "admin" across all
	// clusters). To ensure the context name matches the cluster name, we rename the context
	// accordingly.
	if config.CurrentContext != cluster.Name {
		log.Infof("Renaming context %q to %q", config.CurrentContext, cluster.Name)
		config.Contexts[cluster.Name] = config.Contexts[config.CurrentContext]
		delete(config.Contexts, config.CurrentContext)
		config.CurrentContext = cluster.Name
	}

	return clientcmd.NewNonInteractiveClientConfig(*config, cluster.Name, nil, nil).ClientConfig()
}
