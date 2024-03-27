# Getting started

First install cert-manager on your cluster. The operator comes with admissions webhooks that requires self-signed certificates.

```
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.10.1/cert-manager.yaml
```
(You can use the installation method you want, see the [cert-manager's documentation](https://cert-manager.io/docs/installation/)). Note that you can use your own certificates if you don't want cert-manager on your cluster.

Then install Temporal Operator's CRDs on your cluster:

```
kubectl apply --server-side -f https://github.com/alexandrevilain/temporal-operator/releases/latest/download/temporal-operator.crds.yaml
```

Then install the operator on your cluster:

```
kubectl apply -f https://github.com/alexandrevilain/temporal-operator/releases/latest/download/temporal-operator.yaml
```

Then create the namespace "demo" and create a sample postgresql server:

```
kubectl apply -f https://raw.githubusercontent.com/alexandrevilain/temporal-operator/main/examples/cluster-postgres/00-namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/alexandrevilain/temporal-operator/main/examples/cluster-postgres/01-postgresql.yaml
```

Finish by creating your first temporal cluster:
```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
  version: 1.23.0
  numHistoryShards: 1
  persistence:
    defaultStore:
      sql:
        user: temporal
        pluginName: postgres
        databaseName: temporal
        connectAddr: postgres.demo.svc.cluster.local:5432
        connectProtocol: tcp
      passwordSecretRef:
        name: postgres-password
        key: PASSWORD
    visibilityStore:
      sql:
        user: temporal
        pluginName: postgres
        databaseName: temporal_visibility
        connectAddr: postgres.demo.svc.cluster.local:5432
        connectProtocol: tcp
      passwordSecretRef:
        name: postgres-password
        key: PASSWORD
```

Apply this file to the cluster.

To try more features the operator provides feel free to navigate in the documentation website or checkout the [examples/](https://github.com/alexandrevilain/temporal-operator/tree/main/examples) directory.