# Temporal UI

This page is WIP. Feel free to contribute [on github](https://github.com/alexandrevilain/temporal-operator/edit/main/docs/features/temporal-ui.md).

## Enable UI and set version
Example:
```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
  version: 1.20.0
  numHistoryShards: 1
  # [...]
  ui:
    enabled: true
    # You can specify ui version if needed.
    # Check available tag you can check by link below
    # https://hub.docker.com/r/temporalio/ui/tags
    version: 2.15.0
```

## Create Ingress

Example:
```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
  version: 1.20.0
  numHistoryShards: 1
  # [...]
  ui:
    enabled: true
    version: 2.15.0
    ingress:
      hosts:
        - example.com
      annotations:
        <annotations>
```