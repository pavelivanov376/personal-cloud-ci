#!/bin/bash

docker context use desktop-linux
# Initially we install Tekton once: kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml
# As well as Kyverno: kubectl create -f https://github.com/kyverno/kyverno/releases/download/v1.13.4/install.yaml
# And the monitoring stack (Prometheus + Grafana, provides the ServiceMonitor CRD):
#   helm repo add prometheus-community https://prometheus-community.github.io/helm-charts && helm repo update
#   helm install monitoring prometheus-community/kube-prometheus-stack --namespace monitoring --create-namespace \
#     --set nodeExporter.enabled=false --set prometheus-node-exporter.hostRootFsMount.enabled=false --wait
#   (node-exporter disabled: it CrashLoopBackOffs on the Docker Desktop VM and we only need app metrics.)

docker build -t personalcicloud/api-service:local  ./front-end/api-service
docker build -t personalcicloud/gear:local         ./front-end/gear
docker build -t personalcicloud/ui:local           ./front-end/ui
docker build -t personalcicloud/build-syncer:local ./back-end/build-syncer
docker build -t personalcicloud/segton:local       ./back-end/segton

# One command installs/upgrades the whole stack. --set rollDate bumps a pod
# annotation each run so the app Deployments always roll to pick up freshly
# built :local images (the Helm equivalent of the old `rollout restart`).
helm upgrade --install cicloud ./chart \
  --namespace cicloud --create-namespace \
  --set rollDate="$(date +%s)" \
  --wait

echo "UI: http://localhost:30500"
