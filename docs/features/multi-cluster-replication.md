# Multi-cluster replication

Temporal can replicate global namespaces from one cluster to another. Use it for disaster recovery, to serve another region, or to move to a cluster with more history shards.

The operator renders each cluster's `clusterMetadata` server config from `spec.clusterMetadata`. Connecting the clusters and replicating namespaces are runtime operations, done with the `temporal` CLI and `TemporalNamespace` resources.

## Configure each cluster

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: primary
spec:
  clusterMetadata:
    enableGlobalNamespace: true
    initialFailoverVersion: 1
    # Optional: noop (default), selected-apis-forwarding or all-apis-forwarding.
    dcRedirectionPolicy: selected-apis-forwarding
```

- `enableGlobalNamespace` lets the cluster host global namespaces.
- `initialFailoverVersion` must be different on every cluster that replicates with this one. The failover version increment is fixed at 10, so the value is between 1 and 9, and at most 9 clusters can replicate together. Default: 1.
- `dcRedirectionPolicy` makes the frontend forward requests for a namespace that is active in another cluster.

Without `spec.clusterMetadata`, the cluster is standalone and its config is the same as before this field existed.

Temporal stores `enableGlobalNamespace` and `initialFailoverVersion` in its database the first time global namespaces are enabled, and ignores later config changes to them. So once `enableGlobalNamespace` is `true`, the webhook rejects disabling it and changing `initialFailoverVersion`. Before that, you can change both.

## Connect the clusters and replicate a namespace

These are runtime operations. Run them with the `temporal` CLI, for example from an admin tools pod. `primary` and `secondary` are the two TemporalClusters, and `orders` is the namespace to replicate.

```bash
# On primary, then on secondary with primary's address:
temporal operator cluster upsert --frontend-address secondary-frontend.<namespace>:7233 \
  --enable-connection --enable-replication

# On primary: make the namespace global and add the second cluster.
# "Invalid cluster name" can appear for about a minute after the upsert; retry.
temporal operator namespace update -n orders --promote-global
temporal operator namespace update -n orders --cluster primary --cluster secondary
```

New and changed workflow executions now replicate as they run. Executions that do not change are copied by Temporal's force-replication workflow:

```bash
# On primary:
temporal workflow execute -n temporal-system --task-queue default-worker-tq \
  --type force-replication --workflow-id fr-orders --input '{"Namespace":"orders",
  "Query":"StartTime > \"1970-01-01T00:00:00Z\"","ConcurrentActivityCount":4,
  "OverallRps":50,"EnableVerification":true,"TargetClusterName":"secondary",
  "TargetClusterEndpoint":"secondary-frontend.<namespace>:7233"}'
```

To fail over, once replication has caught up:

```bash
temporal operator namespace update -n orders --active-cluster secondary
```

With `dcRedirectionPolicy: all-apis-forwarding` on the primary, clients that still call the primary keep working after the failover.

## Changing the number of history shards

`numHistoryShards` cannot change on a running cluster. Since Temporal 1.20, clusters with different shard counts can replicate, provided one count is a multiple of the other (4 and 512 work; 512 and 1000 do not). To change it, create a new cluster with the new count and a different `initialFailoverVersion`, replicate every namespace to it, fail over, move the clients, then remove the old cluster from the namespace and delete it.
