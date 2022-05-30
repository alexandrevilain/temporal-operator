# temporal-operator

The Kubernetes Operator to deploy and manage [Temporal](https://temporal.io/) clusters.

Current Status: Work in Progress. The operator can create a basic cluster. Many improvements are needed to make it production ready.

## Roadmap

### Features:
- [x] Deploy a new temporal cluster.
- [x] Ability to deploy multiple clusters.
- [x] Support for SQL datastores.
- [ ] Support for cassandra datastore.
- [ ] Support for Elastisearch.
- [ ] Cluster version upgrades.
- [ ] Automatic mTLS certificates management (using istio, linkerd or cert-manager).
- [ ] Cluster monitoring.
- [ ] Deploy Web UI.
- [ ] Auto scaling.
- [ ] Multi cluster replication.
- [ ] Complete end2end test suite.

## Quick start

First install CRDs on your cluster and the operator:

```
kubectl apply -f https://github.com/alexandrevilain/temporal-operator/releases/latest/download/temporal-operator.crds.yaml
kubectl apply -f https://github.com/alexandrevilain/temporal-operator/releases/latest/download/temporal-operator.yaml
```

Then create the namespace "demo" and create a simple postgresql server:

```
kubectl apply -f https://raw.githubusercontent.com/alexandrevilain/temporal-operator/main/config/samples/namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/alexandrevilain/temporal-operator/main/config/samples/postgresql.yaml
```

Finish by creating your first temporal cluster:
```
apiVersion: apps.alexandrevilain.dev/v1alpha1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
  version: 1.16.0
  numHistoryShards: 1
  persistence:
    defaultStore: default
    visibilityStore: default
  datastores:
    - name: default
      sql:
        user: temporal
        pluginName: postgres
        databaseName: temporal
        connectAddr: postgres.demo.svc.cluster.local
        connectProtocol: tcp
      passwordSecretRef:
        name: postgres-password
        key: PASSWORD
```

Apply this file to the cluster.
For more customization options refers to the [api documentation](https://github.com/alexandrevilain/temporal-operator/blob/main/docs/api/v1alpha1.md).

## License

Temporal Operator is licensed under Apache License Version 2.0. [See LICENSE for more information](https://github.com/alexandrevilain/temporal-operator/blob/main/LICENSE).
