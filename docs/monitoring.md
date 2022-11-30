# Monitoring temporal using prometheus-operator

First of all, your need the prometheus operator running in your cluster:

```
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install prometheus-operator prometheus-community/kube-prometheus-stack --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false --set=prometheusOperator.namespaces.additional={demo,default,}
```

Then create your temporal cluster and enable the `ServiceMonitor` creation:

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
  version: 1.17.4
  numHistoryShards: 1
  # [...]
  metrics:
    enabled: true
    prometheus:
      listenPort: 9090
      scrapeConfig:
        serviceMonitor:
          enabled: true
```