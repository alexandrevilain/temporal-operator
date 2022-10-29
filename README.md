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
- [x] Support for Elastisearch.
- [x] Support for Cassandra datastore.
- [x] Automatic mTLS certificates management (using cert-manager).
- [x] Support for integration in meshes: istio & linkerd.
- [x] Namespace management using CRDs.
- [x] Cluster version upgrades.
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
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
  version: 1.17.4
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
For more customization options refers to the [api documentation](https://github.com/alexandrevilain/temporal-operator/blob/main/docs/api/v1beta1.md).

## Examples

Few examples are available to help you get started:
- [Demo cluster with PostgreSQL](https://github.com/alexandrevilain/temporal-operator/blob/main/examples/cluster-postgres)
- [Demo cluster with PostgreSQL & advanced visibility using ElasticSearch](https://github.com/alexandrevilain/temporal-operator/blob/main/examples/cluster-postgres-es)
- [Demo cluster with Cassandra](https://github.com/alexandrevilain/temporal-operator/blob/main/examples/cluster-cassandra)
- [Demo cluster with mTLS using cert-manager & PostgreSQL as datastore](https://github.com/alexandrevilain/temporal-operator/blob/main/examples/cluster-mtls)

## License

Temporal Operator is licensed under Apache License Version 2.0. [See LICENSE for more information](https://github.com/alexandrevilain/temporal-operator/blob/main/LICENSE).

## Getting Started 🤩🤗

1. Fork this repository.
2. Clone on the repository to your local machine

    ```
    git clone <git repo>
    ```

3. Navigate to cloned repository.
    ```
    cd filename
    ```

4. Create a new branch to work on with
    ```markdown
    git checkout -b my-new-branch
    ```

    >  Time to make some changes to the cloned repository on your local machine.

5. Add your work with
    ```
    git add .
    ```

6. Save your work, by commiting it.
    ```markdown
    git commit -m "first commit"
    ```

7. Let's try pushing it on remote repository, in our case github!
    ```
    git push origin my-new-branch
    ```

8. Ready to share your work with others? So let's generate a Pull Request!
    To do this, go to your forked github repository and `Create a Pull Request`.
