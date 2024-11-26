# Temporal Operator Helm Chart

![Version: 0.6.0](https://img.shields.io/badge/Version-0.6.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.20.0](https://img.shields.io/badge/AppVersion-v0.20.0-informational?style=flat-square)

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
| manager.nodeSelector | object | `{}` |  |
| manager.replicas | int | `1` | Number of controller manager replicas to deploy. |
| manager.resources.limits | object | `{"cpu":"500m","memory":"128Mi"}` | Resources limits for the controller manager container. |
| manager.resources.requests | object | `{"cpu":"10m","memory":"64Mi"}` | Resources requests for the controller manager container. |
| manager.serviceAccount | object | `{"annotations":{}}` | Service account settings for the controller manager container. |
| manager.tolerations | list | `[]` |  |
| webhook.certManager | object | `{"certificate":{"enabled":true,"issuerRef":{},"useCustomIssuer":false}}` | Certificate manager settings for the webhook server. |
| webhook.certManager.certificate | object | `{"enabled":true,"issuerRef":{},"useCustomIssuer":false}` | Webhook certificate configuration using cert-manager.  |
| webhook.certManager.certificate.enabled | bool | `true` | Enabled defines if cert-manager should be used to manage the webhook certificate. |
| webhook.certManager.certificate.issuerRef | object | `{}` | Issuer references if you want to use custom issuer In other case will be used selfSigned issuer. |
| webhook.certManager.certificate.useCustomIssuer | bool | `false` | Defines if cert-manager should use self-signed issuer or custom issuer. |
| webhook.containerPort | int | `9443` | The port that the webhook listens on. |
| webhook.hostNetwork | bool | `false` | Set to true if the webhook should be started in hostNetwork mode. This is useful in managed clusters (e.g. AWS EKS) with custom CNI (such as Calico), where the control-plane cannot reach pods' IP CIDR and admission webhooks are not working. `webhook.containerPort` should be adapted in case it conflicts with the host network. |
| webhook.ports | list | `[{"port":443,"protocol":"TCP","targetPort":9443}]` | Service ports settings for the webhook server. |
| webhook.type | string | `"ClusterIP"` | Service type for the webhook server. |
