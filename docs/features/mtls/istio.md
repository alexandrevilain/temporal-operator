# mTLS using istio

The temporal operator supports mTLS using istio.
To use istio and enforce mTLS you only have set `istio` as mTLS provider.

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
# [...]
  mTLS:
    provider: istio
# [...]
```

The Operator creates for each temporal services a `DestinationRule` and a `PeerAuthentication`. They both ensure mutual and strict mTLS.

## Allowing permissive mTLS for metrics

If your metrics collector isn't using Istio or is otherwise unable to connect using mTLS, you can enable permissive mode for the metrics port.

```yaml
spec:
# [...]
  mTLS:
    provider: istio
    permissiveMetrics: true
# [...]
```
