package main

import (
	"context"
	"log"
	"os"

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
	namespace := getenv("NAMESPACE", "cicloud")
	pipelineName := getenv("PIPELINE_NAME", "spring-pipeline")

	// Talk to the Kubernetes API using the pod's ServiceAccount.
	cfg, err := rest.InClusterConfig()
	if err != nil {
		log.Fatalf("kubernetes config: %v", err)
	}
	k8s, err := dynamic.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("kubernetes client: %v", err)
	}

	forward := &SegtonPipelineRunWatchService{
		k8s:          k8s,
		namespace:    namespace,
		pipelineName: pipelineName,
	}
	reverse := &TektonPipelineRunWatchService{
		k8s:       k8s,
		namespace: namespace,
	}

	log.Printf("segton starting: namespace=%s pipeline=%s", namespace, pipelineName)

	ctx := context.Background()
	go forward.Run(ctx)
	go reverse.Run(ctx)
	select {}
}
