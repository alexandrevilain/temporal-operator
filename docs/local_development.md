# Local Development

Tilt offers a simple way of creating a local development environment.

## Create a kind cluster

Create a kind cluster with a local registry:
```bash
make dev-cluster
```

## Run Tilt

Then run:
```bash
tilt up
```

Now, Tilt will automatically reload the deployment to your local cluster every time you make a code change.