## Build & start

```sh
# 1. build images
docker context use desktop-linux
docker build -t personalcicloud/api-service:local  ./front-end/api-service
docker build -t personalcicloud/gear:local         ./front-end/gear
docker build -t personalcicloud/ui:local           ./front-end/ui
docker build -t personalcicloud/build-syncer:local ./back-end/build-syncer

# 2. apply manifests
kubectl apply -f k8s/

# 3. wait for everything to be ready
kubectl -n cicloud rollout status statefulset/db
kubectl -n cicloud rollout status deployment/api-service
kubectl -n cicloud rollout status deployment/gear
kubectl -n cicloud rollout status deployment/ui
kubectl -n cicloud rollout status deployment/build-syncer

# 4. open the UI
open http://localhost:30500
```

## Useful commands

```sh
# Watch build-syncer live
kubectl -n cicloud logs -f deployment/build-syncer

# Get a shell in any pod
kubectl -n cicloud exec -it deployment/api-service -- sh

# Rebuild after a code change
docker build -t personalcicloud/<service>:local ./path/to/service
kubectl -n cicloud rollout restart deployment/<service>

# Nuke everything
kubectl delete namespace cicloud
```
