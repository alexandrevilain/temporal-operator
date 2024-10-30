# Monitoring temporal using Prometheus

The operator provides support for prometheus to monitor your Temporal cluster.
When metrics exposition is enabled, the operator adds prometheus service discovery annotations on each temporal components:

- `prometheus.io/scrape`
- `prometheus.io/scheme`
- `prometheus.io/path`
- `prometheus.io/port`

If you're using prometheus-operator, please check the dedicated documentation page: [Monitoring temporal using prometheus-operator](/features/monitoring/prometheus-operator/)

## Enabling metrics exposition using prometheus annotations

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
  metrics:
    enabled: true
    prometheus:
      listenPort: 9090
      scrapeConfig:
        annotations: true
```
