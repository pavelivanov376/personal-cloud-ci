#!/bin/bash

docker context use desktop-linux
# Initially we install Tekton once: kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml

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
