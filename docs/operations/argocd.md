# ArgoCD caveats

If you're using ArgoCD, you might run into `Out Of Sync` diff issues when using durations in the `spec` of some of the operator's CRDs.  This issue is [documented with ArgoCD](https://github.com/argoproj/argo-cd/discussions/14229) and appears when using durations in the `spec`. For example, in the `mTLS` part:

```yaml
mTLS:
  provider: cert-manager
  internode:
    enabled: true
  frontend:
    enabled: true
  certificatesDuration:
    clientCertificates: 1h
    frontendCertificate: 1h
    intermediateCAsCertificates: 1h30m
    internodeCertificate: 1h
    rootCACertificate: 2h
  renewBefore: 55m
```

In these cases, there are two workarounds:

1. Update your config to match the canonical representation of the durations:
```yaml
spec:
  mTLS:
    provider: cert-manager
    internode:
      enabled: true
    frontend:
      enabled: true
    certificatesDuration:
      clientCertificates: 1h0m0s
      frontendCertificate: 1h0m0s
      intermediateCAsCertificates: 1h30m0s
      internodeCertificate: 1h0m0s
      rootCACertificate: 2h0m0s
    renewBefore: 55m0s
```
2. Disable the validations in the ArgoCD `Application` `spec`:
```yaml
spec:
  syncOptions:
    - RespectIgnoreDifferences=true
  ignoreDifferences:
    - group: temporal.io
      kind: TemporalCluster
      jsonPointers:
        - /spec/mTLS/certificatesDuration
        - /spec/mTLS/renewBefore
```
