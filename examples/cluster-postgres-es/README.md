# cluster-postgres-es

To run this example you have to start the eck-operator first:

```
helm repo add elastic https://helm.elastic.co
helm repo update
helm install elastic-operator elastic/eck-operator -n elastic-system --create-namespace --version v2.8.0 --wait
```

Then apply the manifests:
```
kubectl apply -f 00-namespace.yaml
kubectl apply -f 01-postgresql.yaml
kubectl apply -f 02-elasticsearch.yaml
kubectl apply -f 03-temporal-cluster.yaml
```