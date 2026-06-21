package client

import (
	"os"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	globalClient     kubernetes.Interface
	globalClientOnce sync.Once
)

func GetClient() (kubernetes.Interface, error) {
	var err error
	globalClientOnce.Do(func() {
		var config *rest.Config

		apiServer := os.Getenv("KUBE_API_SERVER")
		token := os.Getenv("KUBE_TOKEN")

		if apiServer != "" && token != "" {
			config = &rest.Config{
				Host:        apiServer,
				BearerToken: token,
				TLSClientConfig: rest.TLSClientConfig{
					Insecure: true,
				},
			}
		} else if os.Getenv("IN_CLUSTER") == "true" || os.Getenv("KUBECONFIG") == "" {
			config, err = rest.InClusterConfig()
		} else {
			config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		}
		if err != nil {
			return
		}

		globalClient, err = kubernetes.NewForConfig(config)
	})
	return globalClient, err
}
