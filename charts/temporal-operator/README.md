# Temporal Operator Helm Chart

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.16.1](https://img.shields.io/badge/AppVersion-v0.16.1-informational?style=flat-square)

This Helm chart deploys the Temporal Operator to manage a Temporal Cluster in a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.22+
- Helm 3+

## Get repository

To add the Temporal Operator repository, use the following Helm command:

```bash
helm repo add temporal-operator https://alexandrevilain.github.io/temporal-operator
helm repo update
```

## Install Chart

To install the chart, use the following command:

```bash
helm install [RELEASE_NAME] temporal-operator/temporal-operator
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| imagePullSecrets | list | `[]` | Image pull secrets for accessing private image repositories. |
| kubernetesClusterDomain | string | `"cluster.local"` | Domain for the cluster. |
| manager.args | list | `["--leader-elect"]` | Arguments to be passed to the controller manager container. |
| manager.containerSecurityContext | object | `{"allowPrivilegeEscalation":false}` | Security context for the controller manager container. |
| manager.containerSecurityContext.allowPrivilegeEscalation | bool | `false` | Disallow privilege escalation for the container. |
| manager.image.repository | string | `"ghcr.io/alexandrevilain/temporal-operator"` | Docker image repository for the controller manager container. |
| manager.replicas | int | `1` | Number of controller manager replicas to deploy. |
| manager.resources.limits | object | `{"cpu":"500m","memory":"128Mi"}` | Resources limits for the controller manager container. |
| manager.resources.requests | object | `{"cpu":"10m","memory":"64Mi"}` | Resources requests for the controller manager container. |
| manager.serviceAccount | object | `{"annotations":{}}` | Service account settings for the controller manager container. |
| webhook.certManager | object | `{"certificate":{"enabled":true,"issuerRef":{},"useCustomIssuer":false}}` | Certificate manager settings for the webhook server. |
| webhook.certManager.certificate | object | `{"enabled":true,"issuerRef":{},"useCustomIssuer":false}` | Webhook certificate configuration using cert-manager.  |
| webhook.certManager.certificate.enabled | bool | `true` | Enabled defines if cert-manager should be used to manage the webhook certificate. |
| webhook.certManager.certificate.issuerRef | object | `{}` | Issuer references if you want to use custom issuer In other case will be used selfSigned issuer. |
| webhook.certManager.certificate.useCustomIssuer | bool | `false` | Defines if cert-manager should use self-signed issuer or custom issuer. |
| webhook.ports[0].port | int | `443` |  |
| webhook.ports[0].protocol | string | `"TCP"` |  |
| webhook.ports[0].targetPort | int | `9443` |  |
| webhook.type | string | `"ClusterIP"` |  |
