# Admin tools

This page is WIP. Feel free to contribute [on github](https://github.com/alexandrevilain/temporal-operator/edit/main/docs/features/admin-tools.md).

## Enable admintools and set version
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
  admintools:
    enabled: true
    # You can specify ui version if needed.
    # Check available tag you can check by link below
    # https://hub.docker.com/r/temporalio/admin-tools/tags
    version: 1.20.3
```