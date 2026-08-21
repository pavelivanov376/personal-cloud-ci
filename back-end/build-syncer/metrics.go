package main

import (
	"log"
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// StartMetricsServer serves the default Go runtime metrics (goroutines, memory,GC) at /metrics
func StartMetricsServer(addr string) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	go func() {
		log.Printf("metrics server listening on %s/metrics", addr)
		if err := http.ListenAndServe(addr, mux); err != nil {
			log.Printf("metrics server stopped: %v", err)
		}
	}()
}
