package main

import (
	"context"
	"log"
	"os"
	"time"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	gearURL := getenv("GEAR_URL", "http://localhost:8081")
	namespace := getenv("PIPELINERUN_NAMESPACE", "cicloud")
	interval, err := time.ParseDuration(getenv("POLL_INTERVAL", "10s"))
	if err != nil {
		log.Fatalf("invalid POLL_INTERVAL: %v", err)
	}

	// Talk to the Kubernetes API using the pod's ServiceAccount.
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("kubernetes config: %v", err)
	}
	k8s, err := dynamic.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("kubernetes client: %v", err)
	}

	svc := &BuildQueueWatchService{
		gearURL:   gearURL,
		k8s:       k8s,
		namespace: namespace,
	}

	log.Printf("build-syncer starting: gear=%s interval=%s namespace=%s", gearURL, interval, namespace)

	// Poll once immediately, then every `interval`.
	ctx := context.Background()
	svc.Tick(ctx)
	for range time.Tick(interval) {
		svc.Tick(ctx)
	}
}
