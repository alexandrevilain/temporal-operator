# Multi-cluster replication

Temporal supports multi-cluster replication. This feature allows you to replicate specific temporal namespaces to a different temporal cluster. This is useful for disaster recovery, or to have a temporal cluster in a different region for latency reasons, or if you want to upgrade the temporal history shard count.

## How it works

To set up multi-cluster replication using the temporal operator, you must first enable global namespaces on the clusters you wish to support, and then assign them a unique failover version.
This can be configured via the `spec.replicaton` of the `TemporalCluster` resource. Temporal operator automatically configures the remaining fields, and currently hard-codes the failover
increment to 10, meaning you can have at most one leader and 9 followers. If a cluster fails, the remaining clusters will elect a new primary cluster based with the lowest failover version. The original cluster, if it comes back online, will
be assigned a new failover version, which is always the lowest multiple of the failover increment (+ initialFailoverVersion) that is greater than the leader cluster's failover version.

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
spec:
  replication:
    enableGlobalNamespace: true
    initialFailoverVersion: 1
```

For example, in a setup with a leader with `initialFailoverVersion` 1, and a follower with `initialFailoverVersion` 2, since the increment is set to 10 a failure in the leader will flip control to the follower, and increment the leader's failover version to 11.

## Starting replication

Once the two clusters are configured, simply set up connections between them using the temporal CLI. In the future, there may be a way to do this via the operator.

```bash
# port forward to the frontend of the primary cluster
kubectl port-forward primary-frontend 7233:7233

temporal operator cluster upsert --frontend-address secondary-cluster.namespace.svc.cluster.local:7233 --enable-connection true

# port forward to the frontend of the secondary cluster
kubectl port-forward secondary-frontend 7233:7233

temporal operator cluster upsert --frontend-address primary-cluster.namespace.svc.cluster.local:7233 --enable-connection true
```

## Replicating namespaces

Once the clusters are connected, you can replicate namespaces between them. This can be done via the temporal CLI, or via the operator. Simply create a new namespace resource for just the primary clusterRef, and add the secondary cluster to the list of clusters. The namespace will be added to the secondary cluster automatically, with all the same settings, and start receiving updates.

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalNamespace
metadata:
  name: primary-namespace
spec:
  clusterRef:
    name: primary
  clusters:
    - primary
    - secondary
  activeClusterName: secondary
  isGlobalNamespace: true
```

| **🚨 Note**: Enabling replication will not automatically replicate old workflows. It only replicates workflows as they are interacted with. For cases like trying to increase the shard count, this is important as you need to make sure each workflow has been evaluated at least once after replication has been set up.

## A mechanism for increasing the history shard count

Since temporal 1.20, replicated clusters do not require the same number of history shards. This means it is a viable method for migrating a cluster that has outgrown its shard count. To do this, simply have your secondary cluster use a higher shard count than the primary cluster. The only requirement is that the shard count on the secondary cluster is an even multiple of the first. So if you have 512 shards on the primary cluster, you can have 1024 shards (or any other multiple) on the secondary cluster, but not 1023.

When replication is complete, simply take down the old cluster, and flip your clients over to the new cluster. This can all be achieved with very little downtime.
