# PretzelAI Operator

Lightweight Kubernetes Operator to manage PretzelAI instances (Jupyter-based notebooks packaged as PretzelAI). This repository contains the operator source code, CRD definitions, and kustomize manifests used to deploy the operator locally or in a cluster.

## Overview

The operator watches PretzelAI custom resources and ensures a Deployment and Service are present according to the CR spec. It includes basic features such as:
- Pod/Service reconciliation
- ConfigMap support
- Finalizer handling
- RBAC markers for operator permissions (including leader election `leases`)

The operator expects the application (PretzelAI) to listen on container port `8888` by default.

## Requirements

- Go 1.20+ (or version required by controller-runtime used in this project)
- kubectl
- A Kubernetes cluster (kind is recommended for local testing)
- docker (to build local images) if using kind

## Quickstart (local cluster using kind)

1. Create a kind cluster (if you don't have one):

```bash
kind create cluster --name pretzelai
```

2. Build the application image used by the operator (if you changed the app):

```bash
# from repo root
docker build -t pretzelai:local .
# if using kind, load the image into the cluster
kind load docker-image pretzelai:local --name pretzelai
```

3. Generate manifests (CRD and RBAC) and deploy the operator manager:

```bash
make manifests
make deploy IMG=pretzelai:local
```

4. (Optional) Apply a sample PretzelAI resource (moved to `archive/` in this repo). To test, you can apply `archive/config/samples/pretzelai_v1alpha1_pretzelai.example.yaml` after adjusting `image`/`replicas`.

```bash
kubectl apply -f archive/config/samples/pretzelai_v1alpha1_pretzelai.example.yaml
```

5. Port-forward to the PretzelAI pod (assuming the pod is running and listens on 8888):

```bash
kubectl get pods -n pretzelai-operator-system
kubectl port-forward pod/<pretzelai-pod-name> 8888:8888 -n <namespace>
```

Then open `http://localhost:8888` and use the token printed in the pod logs.

## Project structure

- `api/` — CRD Go types and markers
- `controllers/` — reconciliation logic
- `config/` — kustomize manifests for CRD, RBAC, samples and manager
- `hack/` — helper scripts and boilerplate used by code generation (do not remove)
- `archive/` — moved example and placeholder files (archived during cleanup)

## Development & Testing

Format and vet the code:

```bash
gofmt -w .
go vet ./...
go mod tidy
```

Regenerate manifests after changing markers or API types:

```bash
make manifests
```

Deploy locally:

```bash
make deploy IMG=pretzelai:local
```

Check operator logs:

```bash
kubectl logs -n pretzelai-operator-system -l control-plane=controller-manager -f
```

## Notes

- The original placeholder Dockerfile and sample CRs have been moved to `archive/` during cleanup. Use them as references only.
- Keep `hack/boilerplate.go.txt` as it is required by controller-gen header injection.

## License

Licensed under the Apache License, Version 2.0. See `LICENSE` or https://www.apache.org/licenses/LICENSE-2.0 for details.


## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:
	
```sh
make docker-build docker-push IMG=<some-registry>/pretzelai-operator:tag
```
	
3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/pretzelai-operator:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller to the cluster:

```sh
make undeploy
```

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources untile the desired state is reached on the cluster 

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.


# pretzelai-operator
