package main

import (
	"fmt"
	"log"
	"os"

	"github.com/kubeops/ops-kubernetes/internal/config"
	"github.com/kubeops/ops-kubernetes/internal/api"
)

func main() {
	cfg := config.Load()

	if err := api.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "failed to start server: %v\n", err)
		log.Fatal(err)
	}
}
