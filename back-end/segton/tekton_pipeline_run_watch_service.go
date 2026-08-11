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

// TektonPipelineRunWatchService watches Tekton PipelineRuns and, when one
// created by us changes, mirrors its Succeeded condition back onto the
// owning SegtonPipelineRun's spec.status.
type TektonPipelineRunWatchService struct {
	k8s       dynamic.Interface
	namespace string
}

func (s *TektonPipelineRunWatchService) Run(ctx context.Context) {
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		s.k8s, 0, s.namespace, nil,
	)
	informer := factory.ForResource(tektonPipelineRunGVR).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { s.handle(ctx, obj) },
		UpdateFunc: func(_, obj interface{}) { s.handle(ctx, obj) },
	})

	log.Printf("segton-reverse: watching Tekton PipelineRuns in namespace=%s", s.namespace)
	factory.Start(ctx.Done())
	<-ctx.Done()
}

func (s *TektonPipelineRunWatchService) handle(ctx context.Context, obj interface{}) {
	u := obj.(*unstructured.Unstructured)

	// Only react to Tekton PipelineRuns we created (owned by a SegtonPipelineRun).
	ownerName := findSegtonOwner(u)
	if ownerName == "" {
		return
	}

	projectStatus := tektonConditionToProjectStatus(u)
	if projectStatus == "" {
		return
	}

	// Read the current SegtonPipelineRun.spec.status; skip if it already matches
	// (avoids an infinite update echo through the informer).
	current, err := s.k8s.Resource(segtonPipelineRunGVR).Namespace(s.namespace).Get(ctx, ownerName, metav1.GetOptions{})
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
	_, err = s.k8s.Resource(segtonPipelineRunGVR).Namespace(s.namespace).
		Patch(ctx, ownerName, types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		log.Printf("segton-reverse: failed to patch segton-plr %s: %v", ownerName, err)
		return
	}
	log.Printf("segton-reverse: %s %s → %s", ownerName, currentStatus, projectStatus)
}

// findSegtonOwner returns the name of the SegtonPipelineRun that owns this
// Tekton PipelineRun, or "" if none.
func findSegtonOwner(u *unstructured.Unstructured) string {
	for _, ref := range u.GetOwnerReferences() {
		if ref.Kind == "SegtonPipelineRun" {
			return ref.Name
		}
	}
	return ""
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
