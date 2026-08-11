package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Build is what gear returns from GET /builds/{id}.
type Build struct {
	ID            string `json:"id"`
	BuildNumber   int    `json:"buildNumber"`
	Timestamp     string `json:"timestamp"`
	Status        string `json:"status"`
	RepositoryUrl string `json:"repositoryUrl"`
}

// The Kubernetes resource we create per queued build.
var pipelineRunGVR = schema.GroupVersionResource{
	Group:    "steward.sap.com",
	Version:  "v1alpha1",
	Resource: "pipelineruns",
}

// BuildQueueWatchService polls gear for queued builds and creates a
// PipelineRun for each one.
type BuildQueueWatchService struct {
	gearURL   string
	k8s       dynamic.Interface
	namespace string
}

// Tick runs one poll cycle.
func (s *BuildQueueWatchService) Tick(ctx context.Context) {
	ids, err := s.fetchQueuedBuilds()
	if err != nil {
		log.Printf("failed to list queued builds: %v", err)
		return
	}
	if len(ids) == 0 {
		log.Printf("no queued builds")
		return
	}
	log.Printf("found %d queued build(s)", len(ids))
	for _, id := range ids {
		b, err := s.fetchBuild(id)
		if err != nil {
			log.Printf("failed to fetch build %s: %v", id, err)
			continue
		}
		log.Printf("build: id=%s number=%d repo=%s", b.ID, b.BuildNumber, b.RepositoryUrl)
		if err := s.createPipelineRun(ctx, b); err != nil {
			log.Printf("failed to create PipelineRun for build %s: %v", b.ID, err)
			continue
		}
		log.Printf("created PipelineRun build-%s", b.ID)
	}
}

// GET /builds?status=QUEUED — returns a list of build IDs.
func (s *BuildQueueWatchService) fetchQueuedBuilds() ([]string, error) {
	resp, err := http.Get(s.gearURL + "/builds?status=QUEUED")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	var ids []string
	if err := json.NewDecoder(resp.Body).Decode(&ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// GET /builds/{id} — returns the full build.
func (s *BuildQueueWatchService) fetchBuild(id string) (*Build, error) {
	resp, err := http.Get(s.gearURL + "/builds/" + id)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}
	var b Build
	if err := json.NewDecoder(resp.Body).Decode(&b); err != nil {
		return nil, err
	}
	return &b, nil
}

// Create a PipelineRun custom resource in Kubernetes. If it already exists,
// do nothing (so re-polling the same build is safe).
func (s *BuildQueueWatchService) createPipelineRun(ctx context.Context, b *Build) error {
	pr := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "steward.sap.com/v1alpha1",
			"kind":       "PipelineRun",
			"metadata": map[string]interface{}{
				"name": "build-" + b.ID,
			},
			"spec": map[string]interface{}{
				"buildId":       b.ID,
				"buildNumber":   b.BuildNumber,
				"repositoryUrl": b.RepositoryUrl,
				"timestamp":     b.Timestamp,
				"status":        b.Status,
			},
		},
	}
	_, err := s.k8s.Resource(pipelineRunGVR).Namespace(s.namespace).Create(ctx, pr, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil
	}
	return err
}
