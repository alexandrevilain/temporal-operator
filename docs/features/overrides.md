# Applying overrides to resources created by the operator

For some usecases, you may want to override some properites of temporal components. You can use this feature to:

- Set extra properties on created pod like custom resources limits and request
- Add sidecars on temporal services pods
- Add init containers on temporal services pods
- Mount extra volumes 

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

## Overrides per temporal service

Here is a general example:

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
spec:
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