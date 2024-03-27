# Using temporal server dynamic config.

For some usecases, you may want to use temporal server's dynamic config.
You can set all your dynamic config under the field `spec.dynamicconfig.values`, the operator will save them in a configmap as-is, without applying any validation nor mutations.

Example:
```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
  version: 1.23.0
  numHistoryShards: 1
  # [...]
  dynamicConfig:
    pollInterval: 10s
    values:
      matching.numTaskqueueReadPartitions:
      - value: 5
        constraints: {}
      matching.numTaskqueueWritePartitions:
      - value: 5
        constraints: {}
```