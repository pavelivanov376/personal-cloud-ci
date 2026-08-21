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
	segtonNamespace := getenv("SEGTON_NAMESPACE", "cicloud")
	pipelineNamespace := getenv("PIPELINE_NAMESPACE", "tekton-custom-resources")
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
		k8s:               k8s,
		segtonNamespace:   segtonNamespace,
		pipelineNamespace: pipelineNamespace,
		pipelineName:      pipelineName,
	}
	reverse := &TektonPipelineRunWatchService{
		k8s:             k8s,
		segtonNamespace: segtonNamespace,
	}

	log.Printf("segton starting: segton-ns=%s pipeline-ns=%s pipeline=%s",
		segtonNamespace, pipelineNamespace, pipelineName)

	// Expose Prometheus metrics on :9090/metrics.
	StartMetricsServer(":9090")

	ctx := context.Background()
	go forward.Run(ctx)
	go reverse.Run(ctx)
	select {}
}
