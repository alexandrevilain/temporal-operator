# Monitoring temporal using prometheus-operator

The operator provides support for prometheus-operator to monitor your Temporal cluster.
When metrics exposition is enabled with prometheus and serviceMonitor, the operator create a ServiceMonitor for each temporal components (frontend, history, matching & worker).


## Enabling metrics exposition to prometheus using prometheus-operator

First of all, your need the prometheus operator running in your cluster, if not you can install a development version on your cluster:

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
  version: 1.24.2
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

That's all, Prometheus is now scraping temporal components. 
You can find some grafana dashboards in the [official temporal dashboard repository](https://github.com/temporalio/dashboards).

## Relabeling metrics

For some use cases, you may want to add relabelConfig to the created `ServiceMonitors`. 
You can use the `spec.metrics.prometheus.scrapeConfig.serviceMonitor.metricRelabelings` field.

For instance, to prefix all metrics:

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
  # [...]
  metrics:
    enabled: true
    prometheus:
      listenPort: 9090
      scrapeConfig:
        serviceMonitor:
          enabled: true
          metricRelabelings:
          - sourceLabels: [__name__]
            targetLabel: __name__
            replacement: temporal_$1
```

To see all the features provided by this field check the `monitoring.coreos.com/v1.RelabelConfig` [API reference](https://prometheus-operator.dev/docs/operator/api/#monitoring.coreos.com/v1.RelabelConfig) on [prometheus-operator website](https://prometheus-operator.dev/).
 