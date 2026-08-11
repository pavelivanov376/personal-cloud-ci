package main

import (
	"context"
	"log"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

// The two custom resources we care about.
var segtonPipelineRunGVR = schema.GroupVersionResource{
	Group:    "segton.sap.com",
	Version:  "v1alpha1",
	Resource: "segtonpipelineruns",
}

var tektonPipelineRunGVR = schema.GroupVersionResource{
	Group:    "tekton.dev",
	Version:  "v1",
	Resource: "pipelineruns",
}

// SegtonPipelineRunWatchService watches SegtonPipelineRuns and, for each one
// with spec.status=QUEUED, creates a corresponding Tekton PipelineRun that
// runs the configured Pipeline against spec.repositoryUrl.
type SegtonPipelineRunWatchService struct {
	k8s          dynamic.Interface
	namespace    string
	pipelineName string
}

func (s *SegtonPipelineRunWatchService) Run(ctx context.Context) {
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		s.k8s, 0, s.namespace, nil,
	)
	informer := factory.ForResource(segtonPipelineRunGVR).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { s.handle(ctx, obj) },
		UpdateFunc: func(_, obj interface{}) { s.handle(ctx, obj) },
	})

	log.Printf("segton-forward: watching SegtonPipelineRuns in namespace=%s", s.namespace)
	factory.Start(ctx.Done())
	<-ctx.Done()
}

func (s *SegtonPipelineRunWatchService) handle(ctx context.Context, obj interface{}) {
	u := obj.(*unstructured.Unstructured)
	status, _, _ := unstructured.NestedString(u.Object, "spec", "status")
	buildID, _, _ := unstructured.NestedString(u.Object, "spec", "buildId")
	repoURL, _, _ := unstructured.NestedString(u.Object, "spec", "repositoryUrl")

	// Only translate freshly queued builds. Anything RUNNING/FINISHED/FAILED
	// has already been picked up (or was replayed by the informer on restart).
	if status != "QUEUED" || buildID == "" {
		return
	}

	name := "segton-" + buildID
	if err := s.createTektonPipelineRun(ctx, name, repoURL, u); err != nil {
		if errors.IsAlreadyExists(err) {
			return
		}
		log.Printf("segton-forward: failed to create tekton pipelinerun %s: %v", name, err)
		return
	}
	log.Printf("segton-forward: created tekton pipelinerun %s (repo=%s)", name, repoURL)
}

// createTektonPipelineRun builds a tekton.dev/v1 PipelineRun that mirrors the
// shape of the smoke-test in k8s/04-tekton.yaml.
func (s *SegtonPipelineRunWatchService) createTektonPipelineRun(
	ctx context.Context, name, repoURL string, owner *unstructured.Unstructured,
) error {
	pr := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "tekton.dev/v1",
			"kind":       "PipelineRun",
			"metadata": map[string]interface{}{
				"name": name,
				"ownerReferences": []interface{}{
					map[string]interface{}{
						"apiVersion": "segton.sap.com/v1alpha1",
						"kind":       "SegtonPipelineRun",
						"name":       owner.GetName(),
						"uid":        string(owner.GetUID()),
					},
				},
			},
			"spec": map[string]interface{}{
				"pipelineRef": map[string]interface{}{
					"name": s.pipelineName,
				},
				"params": []interface{}{
					map[string]interface{}{"name": "git-url", "value": repoURL},
					map[string]interface{}{"name": "git-revision", "value": "main"},
				},
				"workspaces": []interface{}{
					map[string]interface{}{
						"name": "source",
						"volumeClaimTemplate": map[string]interface{}{
							"spec": map[string]interface{}{
								"accessModes":      []interface{}{"ReadWriteOnce"},
								"storageClassName": "hostpath",
								"resources": map[string]interface{}{
									"requests": map[string]interface{}{"storage": "1Gi"},
								},
							},
						},
					},
				},
			},
		},
	}
	_, err := s.k8s.Resource(tektonPipelineRunGVR).Namespace(s.namespace).Create(ctx, pr, metav1.CreateOptions{})
	return err
}
