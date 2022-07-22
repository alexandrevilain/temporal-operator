# mTLS in the temporal-operator using cert-manager

When you enable mTLS in the operator using the following configuration, the operator asks cert-manager to generate some certificates for you. Cert-manager will then take care to renew them.

```yaml
  mTLS:
    provider: cert-manager
    internode:
      enabled: true
    frontend:
      enabled: true
    certificatesDuration:
      rootCACertificate: 2h
      intermediateCAsCertificates: 1h30m
      clientCertificates: 1h
      frontendCertificate: 1h
      internodeCertificate: 1h
    refreshInterval: 5m
```

Here is a diagram of cert-manager's ressources created by the operator and their hierarchy:

![diagram](/docs/assets/mtls-certmanager.png)

