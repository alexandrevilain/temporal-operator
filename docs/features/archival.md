# Archival

Temporal-operator supports configuring archival for Temporal clusters. You can get more informations about Event histories backup on [Temporal documentation](https://docs.temporal.io/clusters#archival).
The operator supports the following providers:
- AWS s3 (and any s3-compatible provider)
- Google Cloud storage
- Filestore

## Set up Archival using S3 on an Amazon EKS cluster

On EKS clusters, to connect and archive data with s3 you first need to create an IAM role with enough permissions to upload files to s3. 
Then create a TemporalCluster with archival enabled and specify the role name you want to use:

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
  version: 1.21.3
  numHistoryShards: 1
  # [...]
  archival:
    enabled: true
    provider:
      s3:
        roleName: my-role
        region: eu-west-1
    history:
      enabled: true
      enabledRead: true
      path: "my-bucket-name"
    visibility:
      enabled: true
      enabledRead: true
      path: "my-bucket-name2"
```

## Set up Archival using S3 on an s3-compatible object storage

If you want to archive data on an s3-compatible object storage like [OVHCloud Object storage](https://www.ovhcloud.com/en-ie/public-cloud/object-storage/) or [minio](https://min.io/) you have provide your credentials using a secret reference and then reference this secret in the TemporalCluster archival specifications. You also need to specify the s3 custom endpoint.

```bash
kubectl create secret generic archival-credentials --from-literal=AWS_ACCESS_KEY_ID=XXXX --from-literal=AWS_SECRET_ACCESS_KEY=XXXX -n demo
```

```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
  namespace: demo
spec:
  version: 1.21.3
  numHistoryShards: 1
  # [...]
  archival:
    enabled: true
    provider:
      s3:
        region: gra
        endpoint: s3.gra.io.cloud.ovh.net
        credentials:
            accessKeyIdRef:
                name: archival-credentials
                key: AWS_ACCESS_KEY_ID
            secretKeyRef:
                name: archival-credentials
                key: AWS_SECRET_ACCESS_KEY
    history:
      enabled: true
      enableRead: true
      path: "dev-temporal-archival"
    visibility:
      enabled: true
      enableRead: true
      path: "dev-temporal-archival-visibility"
```

## Set up Archival using Filestore

Warning: To use your the storage you desired for filestore archival, you'll need to use overrides.


```yaml
apiVersion: temporal.io/v1beta1
kind: TemporalCluster
metadata:
  name: prod
spec:
  version: 1.21.3
  numHistoryShards: 1
  # [...]
  services:
    overrides:
      deployment:
        spec:
          template:
            spec:
              containers:
                - name: service
                  volumeMounts:
                    - name: archival-data
                      mountPath: /etc/archival
              volumes:
                - name: archival-data
                  emptyDir: {}
  archival:
    enabled: true
    provider:
      filestore: {}
    history:
      enabled: true
      enableRead: true
      path: "/etc/archival/history"
    visibility:
      enabled: true
      enableRead: true
      path: "/etc/archival/visibility"
```

