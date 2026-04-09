package k8s

import (
	"fmt"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// buildDynamicClient tạo K8s dynamic client.
// Ưu tiên in-cluster config (khi chạy trong Pod với ServiceAccount).
// Fallback về ~/.kube/config cho local development — giống logic Python:
//
//	config.load_incluster_config() → except → config.load_kube_config()
func buildDynamicClient() (dynamic.Interface, error) {
	// Thử in-cluster config trước (chạy trong K8s Pod)
	cfg, err := rest.InClusterConfig()
	if err != nil {
		// Fallback: đọc kubeconfig local (~/.kube/config hoặc KUBECONFIG env)
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			loadingRules,
			&clientcmd.ConfigOverrides{},
		)
		cfg, err = clientConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("build kubeconfig (tried in-cluster and local): %w", err)
		}
	}

	dynClient, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create dynamic client: %w", err)
	}
	return dynClient, nil
}
