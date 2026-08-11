#!/bin/bash

docker context use desktop-linux
# Initially we install Tekton once: kubectl apply --filename https://storage.googleapis.com/tekton-releases/pipeline/latest/release.yaml

docker build -t personalcicloud/api-service:local  ./front-end/api-service
docker build -t personalcicloud/gear:local         ./front-end/gear
docker build -t personalcicloud/ui:local           ./front-end/ui
docker build -t personalcicloud/build-syncer:local ./back-end/build-syncer
docker build -t personalcicloud/segton:local       ./back-end/segton

kubectl apply -f k8s/

kubectl -n cicloud rollout restart deployment/api-service
kubectl -n cicloud rollout restart deployment/gear
kubectl -n cicloud rollout restart deployment/ui
kubectl -n cicloud rollout restart deployment/build-syncer
kubectl -n cicloud rollout restart deployment/segton

kubectl -n cicloud rollout status statefulset/db
kubectl -n cicloud rollout status deployment/api-service
kubectl -n cicloud rollout status deployment/gear
kubectl -n cicloud rollout status deployment/ui
kubectl -n cicloud rollout status deployment/build-syncer
kubectl -n cicloud rollout status deployment/segton

echo "UI: http://localhost:30500"
