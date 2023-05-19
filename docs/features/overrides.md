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
                labels:
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

### Example: mount an extra volume to the frontend pod

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
                    volumeMounts:
                      - name: extra-volume
                        mountPath: /etc/extra
                volumes:
                  - name: extra-volume
                    configMap:
                      name: extra-config
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

## Overrides container resources all services

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