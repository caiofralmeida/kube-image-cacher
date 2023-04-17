
# K.I.C (Kube Image Cacher)

## Development

### Setup local cluster

Install the kind cluster with local registry configuration:
```sh
KIND_CLUSTER_NAME=kic-kind ./kind-with-registry.sh
```

Install certs-manager:
```sh
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.11.0/cert-manager.yaml
```

### Running application locally

Run the command below to start application:
```sh
make up
```

## TODO
[ ] Create app configuration
[ ] Update webhook configuration to v1
[ ] Remove unused files
[ ] Fix RBAC

### Scenarios
As an User, I want to cache container images through the registry to avoid docker.io rate limit and make images pull faster.

### Acceptance Criteria
- It should bypass admission review if the containers image is using the destionation registry
- It should push images outside destination registry and change the pod definition
