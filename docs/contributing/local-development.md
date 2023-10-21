# Prerequisites

- [Docker](https://docs.docker.com/engine/install/)
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [tilt](https://docs.tilt.dev/install.html)
- [Golang](https://go.dev/doc/install)
- [kustomize](https://kubectl.docs.kubernetes.io/installation/kustomize/)

# Local Development

Tilt offers a simple way of creating a local development environment.

## Create a kind cluster

Create a kind cluster with a local registry:

```bash
make dev-cluster
```

## Generate

Generate crd and docs when api is modified

```bash
make generate
```

## Run Tilt

Then run:

```bash
tilt up
```

Now, Tilt will automatically reload the deployment to your local cluster every time you make a code change.
