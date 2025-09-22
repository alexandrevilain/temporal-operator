# temporal-operator

The Kubernetes Operator to deploy and manage [Temporal](https://temporal.io/) clusters.

Using this operator, deploying a Temporal Cluster on Kubernetes is as easy as deploying the following manifest:

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
  version: 1.24.3
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

## Documentation

The documentation is available at: [https://temporal-operator.pages.dev/](https://temporal-operator.pages.dev/).

### Quick start

To start using the Operator and deploy you first cluster in a matter of minutes, follow the documentation's [getting started guide](https://temporal-operator.pages.dev/getting-started/).

## Examples

Somes examples are available to help you get started:

- [Temporal Cluster with PostgreSQL](https://github.com/alexandrevilain/temporal-operator/blob/main/examples/cluster-postgres)
- [Temporal Cluster with MySQL](https://github.com/alexandrevilain/temporal-operator/blob/main/examples/cluster-mysql)
- [Temporal Cluster with Cassandra](https://github.com/alexandrevilain/temporal-operator/blob/main/examples/cluster-cassandra)
- [Temporal Cluster with PostgreSQL & advanced visibility using ElasticSearch](https://github.com/alexandrevilain/temporal-operator/blob/main/examples/cluster-postgres-es)
- [Temporal Cluster with mTLS using cert-manager & PostgreSQL as datastore](https://github.com/alexandrevilain/temporal-operator/blob/main/examples/cluster-mtls)
- [Temporal Cluster with mTLS using istio & PostgreSQL as datastore](https://github.com/alexandrevilain/temporal-operator/blob/main/examples/cluster-mtls-istio)
- [Temporal Cluster with mTLS using linkerd & PostgreSQL as datastore](https://github.com/alexandrevilain/temporal-operator/blob/main/examples/cluster-mtls-linkerd)


## Compatibility matrix

The following table shows operator compatibility with Temporal and Kubernetes.
Please note this table only reports end-to-end tests suite coverage, others versions *may* work.

| Temporal Operator      | Temporal           | Kubernetes     |
|------------------------|--------------------|----------------|
| v0.22.x (not released) | v1.24.x to v1.28.x | v1.30 to v1.33 |
| v0.21.x                | v1.20.x to v1.25.x | v1.27 to v1.31 |
| v0.20.x                | v1.19.x to v1.24.x | v1.26 to v1.30 |
| v0.19.x                | v1.19.x to v1.23.x | v1.25 to v1.29 |
| v0.18.x                | v1.19.x to v1.23.x | v1.25 to v1.29 |
| v0.17.x                | v1.18.x to v1.22.x | v1.25 to v1.29 |
| v0.16.x                | v1.18.x to v1.22.x | v1.24 to v1.27 |
| v0.15.x                | v1.18.x to v1.21.x | v1.24 to v1.27 |
| v0.14.x                | v1.18.x to v1.21.x | v1.24 to v1.27 |
| v0.13.x                | v1.18.x to v1.20.x | v1.24 to v1.27 |
| v0.12.x                | v1.18.x to v1.20.x | v1.23 to v1.26 |
| v0.11.x                | v1.17.x to v1.19.x | v1.23 to v1.26 |
| v0.10.x                | v1.17.x to v1.19.x | v1.23 to v1.26 |
| v0.9.x                 | v1.16.x to v1.18.x | v1.22 to v1.25 |

## Roadmap

### Features

- [x] Deploy a new temporal cluster.
- [x] Ability to deploy multiple clusters.
- [x] Support for SQL datastores.
- [x] Deploy Web UI.
- [x] Deploy admin tools.
- [x] Support for Elastisearch.
- [x] Support for Cassandra datastore.
- [x] Automatic mTLS certificates management (using cert-manager).
- [x] Support for integration in meshes: istio & linkerd.
- [x] Namespace management using CRDs.
- [x] Custom search attribute management.
- [x] Cluster version upgrades.
- [x] Cluster monitoring.
- [x] Complete end2end test suite.
- [x] Archival.
- [ ] Auto scaling.
- [ ] Multi cluster replication.

## Contributing

Feel free to contribute to the project ! All issues and PRs are welcome!
To start hacking on the project, you can follow the [local development](https://temporal-operator.pages.dev/contributing/local-development/) documentation page.

## License

Temporal Operator is licensed under Apache License Version 2.0. [See LICENSE for more information](https://github.com/alexandrevilain/temporal-operator/blob/main/LICENSE).
