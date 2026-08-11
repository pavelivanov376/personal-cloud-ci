package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

// SharedInformerService holds a persistent watch on SegtonPipelineRuns in a
// namespace. When a SegtonPipelineRun changes, it reads its status and PATCHes
// gear's /builds/{id}/status endpoint.
type SharedInformerService struct {
	gearURL   string
	k8s       dynamic.Interface
	namespace string
}

// Run starts the informer and blocks until ctx is cancelled. Under the hood
// this opens a single long-lived HTTP watch connection to the Kubernetes API
// server and dispatches Add/Update events as they arrive.
func (s *SharedInformerService) Run(ctx context.Context) {
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		s.k8s, 0, s.namespace, nil,
	)
	informer := factory.ForResource(segtonPipelineRunGVR).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { s.handle(obj, "add") },
		UpdateFunc: func(_, obj interface{}) { s.handle(obj, "update") },
	})

	log.Printf("shared-informer: watching SegtonPipelineRuns in namespace=%s", s.namespace)
	factory.Start(ctx.Done())
	<-ctx.Done()
}

// handle extracts (buildId, status) from a SegtonPipelineRun and forwards it to gear.
func (s *SharedInformerService) handle(obj interface{}, event string) {
	u := obj.(*unstructured.Unstructured)
	buildID, _, _ := unstructured.NestedString(u.Object, "spec", "buildId")
	status, _, _ := unstructured.NestedString(u.Object, "spec", "status")
	if buildID == "" || status == "" {
		return
	}

	log.Printf("shared-informer: %s segtonpipelinerun=%s buildId=%s status=%s",
		event, u.GetName(), buildID, status)

	if err := s.updateBuildStatus(buildID, status); err != nil {
		log.Printf("shared-informer: failed to update build %s: %v", buildID, err)
	}
}

// PATCH /builds/{id}/status?status={status}
func (s *SharedInformerService) updateBuildStatus(buildID, status string) error {
	url := fmt.Sprintf("%s/builds/%s/status?status=%s", s.gearURL, buildID, status)
	req, err := http.NewRequest(http.MethodPatch, url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("gear returned status %d", resp.StatusCode)
	}
	return nil
}
