# Prerequisites

- [Docker](https://docs.docker.com/engine/install/) and [docker-buildx](https://github.com/docker/buildx)
- [Helm](https://helm.sh/)
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [tilt](https://docs.tilt.dev/install.html)
- [Golang](https://go.dev/doc/install)
- [kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/)

# Local Development

Tilt offers a simple way of creating a local development environment.

## Create local k8s cluster

```bash
make deploy-dev
```

Open the `tilt` browser UI by pressing the "space" key when prompted. Note that `tilt` sets up a directory watcher and will automatically redeploy any code changes!

You can check the "temporal-operator-controller-manager" Pod status with `kubectl get pods -n temporal-system -w`.

## Apply Manifests

Once the local cluster is created, start applying your desired manifests and let the temporal operator handle reconciliation:

```bash
# example
make artifacts
kubectl apply -f examples/cluster-postgres/00-namespace.yaml
kubectl apply -f examples/cluster-postgres/01-postgresql.yaml
kubectl apply -f examples/cluster-postgres/02-temporal-cluster.yaml
kubectl apply -f examples/cluster-postgres/03-temporal-namespace.yaml
```

Note: if you wish to interact with the Temporal Web UI or frontend gRPC service, you should port forward the services to localhost.

## Generate

Generate crd and docs when api is modified

```bash
make generate
```

## Test

Run tests with coverage report:

```bash
make test
```

Run end-to-end tests:

```bash
make test-e2e
```

Run end-to-end tests on development computer using `kind`:

```bash
make test-e2e-dev
```

## Gracefully Shutdown k8s Cluster

```bash
make clean-dev-cluster
```
