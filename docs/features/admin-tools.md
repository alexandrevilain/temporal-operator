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
  version: 1.24.2
  numHistoryShards: 1
  # [...]
  admintools:
    enabled: true
    # You can specify the admin tools version if needed.
    # Check available tag you can check by the link below
    # https://hub.docker.com/r/temporalio/admin-tools/tags
    version: 1.24.2-tctl-1.18.1-cli-0.13.2
```
