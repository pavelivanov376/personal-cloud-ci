#!/bin/bash

docker context use desktop-linux

docker build -t personalcicloud/api-service:local  ./front-end/api-service
docker build -t personalcicloud/gear:local         ./front-end/gear
docker build -t personalcicloud/ui:local           ./front-end/ui
docker build -t personalcicloud/build-syncer:local ./back-end/build-syncer

kubectl apply -f k8s/

kubectl -n cicloud rollout restart deployment/api-service
kubectl -n cicloud rollout restart deployment/gear
kubectl -n cicloud rollout restart deployment/ui
kubectl -n cicloud rollout restart deployment/build-syncer

kubectl -n cicloud rollout status statefulset/db
kubectl -n cicloud rollout status deployment/api-service
kubectl -n cicloud rollout status deployment/gear
kubectl -n cicloud rollout status deployment/ui
kubectl -n cicloud rollout status deployment/build-syncer

echo "UI: http://localhost:30500"
