# Temporal UI

This page is WIP. Feel free to contribute [on github](https://github.com/alexandrevilain/temporal-operator/edit/main/docs/features/temporal-ui.md).

Temporal-operator supports configuring web UI for Temporal clusters. You can get more information about web UI on [Temporal documentation](https://docs.temporal.io/web-ui).

## Enable UI and set version

Example:

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
  version: 1.24.3
  numHistoryShards: 1
  # [...]
  ui:
    enabled: true
    # You can specify ui version if needed.
    # Check available tag you can check by link below
    # https://hub.docker.com/r/temporalio/ui/tags
    version: 2.25.0
```

## Create Ingress

Ingress is an optional ingress configuration for the UI. If leaved empty, no ingress configuration will be created and the UI will only by available through ClusterIP service.

Example:

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
  version: 1.24.3
  numHistoryShards: 1
  # [...]
  ui:
    enabled: true
    version: 2.25.0
    ingress:
      hosts:
        - example.com
      annotations:
        <annotations>
```

## Set UI replicas and resources

Example:

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
  version: 1.24.3
  numHistoryShards: 1
  # [...]
  ui:
    enabled: true
    version: 2.25.0
    replicas: 1
    resources:
      limits:
        cpu: 10m
        memory: 20Mi
      requests:
        cpu: 10m
        memory: 20Mi
```

## Override UI deployment

Web UI overrides can be used to set [web UI environment variables](https://docs.temporal.io/references/web-ui-environment-variables).

Example:

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
  version: 1.24.3
  numHistoryShards: 1
  ui:
    enabled: true
    overrides:
      deployment:
        spec:
          template:
            spec:
              containers:
                - name: ui
                  env:
                    - name: TEMPORAL_SHOW_TEMPORAL_SYSTEM_NAMESPACE
                      value: "true"
                    # Allows the UI to be served from a subpath
                    - name: TEMPORAL_UI_PUBLIC_PATH
                      value: /temporal
```
