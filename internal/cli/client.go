package cli

import (
	"fmt"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/ogormans-deptstack/kubectl-generate/pkg/openapi"
)

// LoadClusterDoc builds the OpenAPI document from the cluster's discovery API.
func LoadClusterDoc(kubeconfigPath string) (*openapi.Document, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeconfigPath != "" {
		loadingRules.ExplicitPath = kubeconfigPath
	}

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	).ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("build kubeconfig: %w", err)
	}

	disc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("create discovery client: %w", err)
	}

	fetcher := openapi.NewSchemaFetcher(disc.OpenAPIV3())
	return fetcher.FetchAll()
}
