# temporal-operator

The Kubernetes Operator to deploy and manage [Temporal](https://temporal.io/) clusters.

Current Status: Work in Progress. The operator can create a basic cluster. Many improvements are needed to make it production ready.

## Roadmap

### Features:
- [x] Deploy a new temporal cluster.
- [x] Ability to deploy multiple clusters.
- [x] Support for SQL datastores.
- [x] Deploy Web UI.
- [x] Deploy admin tools.
- [x] Support for Elastisearch (will be released in [v0.1.0](https://github.com/alexandrevilain/temporal-operator/projects/2)).
- [ ] Support for Cassandra datastore (will be released in [v0.2.0](https://github.com/alexandrevilain/temporal-operator/projects/2)).
- [ ] Cluster version upgrades.
- [ ] Automatic mTLS certificates management (using istio, linkerd or cert-manager).
- [ ] Cluster monitoring.
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
kubectl apply -f https://raw.githubusercontent.com/alexandrevilain/temporal-operator/main/examples/cluster-postgres/00-namespace.yaml
kubectl apply -f https://raw.githubusercontent.com/alexandrevilain/temporal-operator/main/examples/cluster-postgres/01-postgresql.yaml
```

Finish by creating your first temporal cluster:
```yaml
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
    visibilityStore: visibility
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
    - name: visibility
      sql:
        user: temporal
        pluginName: postgres
        databaseName: temporal_visibility
        connectAddr: postgres.demo.svc.cluster.local
        connectProtocol: tcp
      passwordSecretRef:
        name: postgres-password
        key: PASSWORD
```

Apply this file to the cluster.
For more customization options refers to the [api documentation](https://github.com/alexandrevilain/temporal-operator/blob/main/docs/api/v1alpha1.md).

## Examples

Few examples are available to help you get started:
- [Demo cluster with PostgreSQL](https://github.com/alexandrevilain/temporal-operator/blob/main/examples/cluster-postgres)
- [Demo cluster with PostgreSQL & advanced visibility using ElasticSearch](https://github.com/alexandrevilain/temporal-operator/blob/main/examples/cluster-postgres-es)

## License

Temporal Operator is licensed under Apache License Version 2.0. [See LICENSE for more information](https://github.com/alexandrevilain/temporal-operator/blob/main/LICENSE).
