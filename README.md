# temporal-operator

The Kubernetes Operator to deploy and manage Temporal clusters.

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
