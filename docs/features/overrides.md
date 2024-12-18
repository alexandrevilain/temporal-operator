# Applying overrides to resources created by the operator

For some usecases, you may want to override some properites of temporal components. You can use this feature to:

- Set extra properties on created pod like custom resources limits and request
- Add sidecars on temporal services pods
- Add init containers on temporal services pods
- Mount extra volumes
- Get environment variable for secretRef

Overrides allows you to override every fields you want in temporal services deployments.

The API provides you the ability to apply your overrides:

- per temporal service (using `spec.services.[frontend|history|matching|worker].overrides`)
- for all services (using `spec.services.overrides`)

There are two ways of performing overrides, one is via StrategicPatchMerge and one using RFC6902 JSON patches. You can find examples of both below. If working with certain fields that aren't handled by StrategicPatchMerge properly (i.e., arrays that don't have go struct tags for merging valid for your use case), you may want to consider using JSON patches.

## Overrides for all services

Here is a general example:

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
spec:
  # [...]
  services:
    overrides:
      deployment:
        metadata:
          labels: {}
          annotations: {}
        spec:
          template:
            spec:
              containers:
                - name: service
                # anything you want
```

### Example: mount an extra volume to all pods

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
spec:
  services:
    overrides:
      deployment:
        spec:
          template:
            spec:
              containers:
                - name: service
                  volumeMounts:
                    - name: extra-volume
                      mountPath: /etc/extra
              volumes:
                - name: extra-volume
                  configMap:
                    name: extra-config
```

### Example: add sidecar to all pods

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
spec:
  # [...]
  services:
    overrides:
      deployment:
        spec:
          template:
            spec:
              containers:
                - name: my-sidecar
                  image: busybox
                  command: ["sh","-c","while true; do echo 'Hello from sidecar'; sleep 30; done"]
```

### Example: add init container to all pods

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
  # [...]
  services:
    overrides:
      deployment:
        spec:
          template:
            spec:
              initContainers:
              - name: init-myservice
                image: busybox:1.28
                command: ['sh', '-c', "echo My example init container"]
```

## Example: Override containers resources

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
spec:
  # [...]
  services:
    overrides:
      deployment:
        spec:
          template:
            spec:
              containers:
                - name: service
                  resources:
                    limits:
                      cpu: 500m
                      memory: 500Mi
                    requests:
                      cpu: 500m
                      memory: 500Mi
```

## Overrides per temporal service

Here is a general example:

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
spec:
  # [...]
  services:
    [service name]:
      overrides:
        deployment:
          metadata:
            labels: {}
            annotations: {}
          spec:
            template:
              spec:
                containers:
                  - name: service
                    # anything you want
```

### Example: Add labels to the frontend pod

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
spec:
  # [...]
  services:
    frontend:
      overrides:
        deployment:
          spec:
            template:
              metadata:
                annotations:
                    ad.datadoghq.com/<CONTAINER_IDENTIFIER>.logs: '[{ "source": "golang", "service": "<CONTAINER_IDENTIFIER>" }]'
                    ad.datadoghq.com/<CONTAINER_IDENTIFIER>.checks: |
                    {
                        "<INTEGRATION_NAME>": {
                        "init_config": <INIT_CONFIG>,
                        "instances": [<INSTANCE_CONFIG>]
                        }
                    }
```

### Example: Add an environment variable to the worker pod

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
spec:
  # [...]
  services:
    worker:
      overrides:
        deployment:
          spec:
            template:
              spec:
                containers:
                  - name: service
                    env:
                      - name: HTTP_PROXY
                        value: example.com
```

### Example: Mount an extra secret volume to the frontend pod

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
spec:
  # [...]
  services:
    frontend:
      overrides:
        deployment:
          jsonPatch:
            - op: add
              path: /spec/template/spec/containers/0/volumeMounts/-
              value:
                name: extra-volume
                mountPath: /etc/extra
            - op: add
              path: /spec/template/spec/volumes/-
              value:
                name: extra-volume
                secret:
                  secretName: test-secret
```

### Example: Add an environment variable from secretRef to the frontend pod

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
spec:
  # [...]
  services:
    frontend:
      overrides:
        deployment:
          spec:
            template:
              spec:
                containers:
                  - name: service
                    envFrom:
                      - secretRef:
                          name: frontend
```

### Example: Replace default liveness probe

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
spec:
  # [...]
  services:
    frontend:
      overrides:
        deployment:
          spec:
            template:
              spec:
                containers:
                  - name: service
                    livenessProbe:
                      $patch: replace
                      tcpSocket: null
                      grpc:
                        port: 7233
                        service: frontend.temporal.temporal.svc.cluster.local
```

### Example: Add environment variable from a secret to frontend pod
```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
spec:
  # [...]
  services:
    frontend:
      overrides:
        deployment:
          jsonPatch:
            - op: add
              path: /spec/template/spec/containers/0/env/-
              value:
                name: TEST
                valueFrom:
                  secretKeyRef:
                    name: test-secret
                    key: test
```

Read more in [Strategic Merge Patch](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-api-machinery/strategic-merge-patch.md#strategic-merge-patch).

## Override UI deployment

See [Temporal UI / Override UI deployment](../temporal-ui/#override-ui-deployment)
