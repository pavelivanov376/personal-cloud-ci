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

// The custom resources we care about.
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

// Core Namespace resource (cluster-scoped, so we call it with no Namespace()).
var namespaceGVR = schema.GroupVersionResource{
	Group:    "",
	Version:  "v1",
	Resource: "namespaces",
}

// Label we stamp on the Tekton PipelineRun so the reverse watcher can find
// the owning SegtonPipelineRun without an ownerReference (owner refs don't
// work across namespaces).
const ownerLabel = "segton.sap.com/owner"

// SegtonPipelineRunWatchService watches SegtonPipelineRuns and, for each one
// with spec.status=QUEUED, creates a fresh namespace segton-run-<buildId>
// and a corresponding Tekton PipelineRun inside it that runs the configured
// Pipeline (fetched from pipelineNamespace via the cluster resolver) against
// spec.repositoryUrl.
type SegtonPipelineRunWatchService struct {
	k8s               dynamic.Interface
	segtonNamespace   string
	pipelineNamespace string
	pipelineName      string
}

func (s *SegtonPipelineRunWatchService) Run(ctx context.Context) {
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		s.k8s, 0, s.segtonNamespace, nil,
	)
	informer := factory.ForResource(segtonPipelineRunGVR).Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { s.handle(ctx, obj) },
		UpdateFunc: func(_, obj interface{}) { s.handle(ctx, obj) },
	})

	log.Printf("segton-forward: watching SegtonPipelineRuns in namespace=%s", s.segtonNamespace)
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

	runNamespace := "segton-run-" + buildID
	if err := s.ensureNamespace(ctx, runNamespace, buildID); err != nil {
		log.Printf("segton-forward: failed to ensure namespace %s: %v", runNamespace, err)
		return
	}

	name := "segton-" + buildID
	if err := s.createTektonPipelineRun(ctx, runNamespace, name, repoURL, u.GetName()); err != nil {
		if errors.IsAlreadyExists(err) {
			return
		}
		log.Printf("segton-forward: failed to create tekton pipelinerun %s/%s: %v", runNamespace, name, err)
		return
	}
	log.Printf("segton-forward: created tekton pipelinerun %s/%s (repo=%s)", runNamespace, name, repoURL)
}

// ensureNamespace creates the per-build namespace, swallowing AlreadyExists.
func (s *SegtonPipelineRunWatchService) ensureNamespace(ctx context.Context, name, buildID string) error {
	ns := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": name,
				"labels": map[string]interface{}{
					"segton.sap.com/build-id": buildID,
				},
			},
		},
	}
	_, err := s.k8s.Resource(namespaceGVR).Create(ctx, ns, metav1.CreateOptions{})
	if errors.IsAlreadyExists(err) {
		return nil
	}
	if err == nil {
		log.Printf("segton-forward: created namespace %s", name)
	}
	return err
}

// createTektonPipelineRun builds a tekton.dev/v1 PipelineRun in runNamespace,
// referencing the Pipeline in s.pipelineNamespace via the cluster resolver.
func (s *SegtonPipelineRunWatchService) createTektonPipelineRun(
	ctx context.Context, runNamespace, name, repoURL, ownerName string,
) error {
	pr := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "tekton.dev/v1",
			"kind":       "PipelineRun",
			"metadata": map[string]interface{}{
				"name": name,
				"labels": map[string]interface{}{
					ownerLabel: ownerName,
				},
			},
			"spec": map[string]interface{}{
				"pipelineRef": map[string]interface{}{
					"resolver": "cluster",
					"params": []interface{}{
						map[string]interface{}{"name": "kind", "value": "pipeline"},
						map[string]interface{}{"name": "name", "value": s.pipelineName},
						map[string]interface{}{"name": "namespace", "value": s.pipelineNamespace},
					},
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
	_, err := s.k8s.Resource(tektonPipelineRunGVR).Namespace(runNamespace).Create(ctx, pr, metav1.CreateOptions{})
	return err
}
