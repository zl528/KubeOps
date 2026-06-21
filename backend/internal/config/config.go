package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port         int
	Mode         string
	KubeConfig   string
	InCluster    bool
	KubeAPIServer string
	KubeToken     string
}

func Load() *Config {
	port := 8080
	if v := os.Getenv("PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			port = p
		}
	}

	mode := os.Getenv("MODE")
	if mode == "" {
		mode = "release"
	}

	kubeConfig := os.Getenv("KUBECONFIG")

	inCluster := os.Getenv("IN_CLUSTER") == "true"

	kubeAPIServer := os.Getenv("KUBE_API_SERVER")
	kubeToken := os.Getenv("KUBE_TOKEN")

	return &Config{
		Port:          port,
		Mode:          mode,
		KubeConfig:    kubeConfig,
		InCluster:     inCluster,
		KubeAPIServer: kubeAPIServer,
		KubeToken:     kubeToken,
	}
}
