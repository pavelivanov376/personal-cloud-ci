package main

import (
	"context"
	"encoding/json"
	"log"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

// TektonPipelineRunWatchService watches Tekton PipelineRuns cluster-wide
// (they now live in per-build namespaces segton-run-<buildId>). When one
// that carries our owner label changes, it mirrors the Succeeded condition
// back onto the owning SegtonPipelineRun's spec.status in segtonNamespace.
type TektonPipelineRunWatchService struct {
	k8s             dynamic.Interface
	segtonNamespace string
}

func (s *TektonPipelineRunWatchService) Run(ctx context.Context) {
	// Cluster-wide informer — a per-build namespace can't be watched
	// individually because it doesn't exist yet at startup.
	factory := dynamicinformer.NewDynamicSharedInformerFactory(s.k8s, 0)
	informer := factory.ForResource(tektonPipelineRunGVR).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { s.handle(ctx, obj) },
		UpdateFunc: func(_, obj interface{}) { s.handle(ctx, obj) },
	})

	log.Printf("segton-reverse: watching Tekton PipelineRuns cluster-wide")
	factory.Start(ctx.Done())
	<-ctx.Done()
}

func (s *TektonPipelineRunWatchService) handle(ctx context.Context, obj interface{}) {
	u := obj.(*unstructured.Unstructured)

	// Only react to Tekton PipelineRuns we created (identified by our owner label).
	ownerName := u.GetLabels()[ownerLabel]
	if ownerName == "" {
		return
	}

	projectStatus := tektonConditionToProjectStatus(u)
	if projectStatus == "" {
		return
	}

	// Read the current SegtonPipelineRun.spec.status; skip if it already matches
	// (avoids an infinite update echo through the informer).
	current, err := s.k8s.Resource(segtonPipelineRunGVR).Namespace(s.segtonNamespace).Get(ctx, ownerName, metav1.GetOptions{})
	if err != nil {
		log.Printf("segton-reverse: failed to get segton-plr %s: %v", ownerName, err)
		return
	}
	currentStatus, _, _ := unstructured.NestedString(current.Object, "spec", "status")
	if currentStatus == projectStatus {
		return
	}

	// JSON merge patch to spec.status.
	patch, _ := json.Marshal(map[string]interface{}{
		"spec": map[string]interface{}{"status": projectStatus},
	})
	_, err = s.k8s.Resource(segtonPipelineRunGVR).Namespace(s.segtonNamespace).
		Patch(ctx, ownerName, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		log.Printf("segton-reverse: failed to patch segton-plr %s: %v", ownerName, err)
		return
	}
	log.Printf("segton-reverse: %s %s → %s", ownerName, currentStatus, projectStatus)
}

// tektonConditionToProjectStatus reads status.conditions[type=Succeeded] and
// maps it to a project status string (RUNNING / FINISHED / FAILED). Returns
// "" if there's no usable condition yet.
func tektonConditionToProjectStatus(u *unstructured.Unstructured) string {
	conditions, _, _ := unstructured.NestedSlice(u.Object, "status", "conditions")
	for _, c := range conditions {
		m, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		if m["type"] != "Succeeded" {
			continue
		}
		switch m["status"] {
		case "Unknown":
			return "RUNNING"
		case "True":
			return "FINISHED"
		case "False":
			return "FAILED"
		}
	}
	return ""
}
