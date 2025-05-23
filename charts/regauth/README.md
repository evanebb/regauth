# regauth

An API-driven container registry authorization server.

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.1.0](https://img.shields.io/badge/AppVersion-0.1.0-informational?style=flat-square)

## Installing

To install the latest version of the chart with release name `my-release`:
```shell
helm install my-release oci://ghcr.io/evanebb/charts/regauth
```

Note that this will not install and configure the database or container registry.
You must install those yourself, and configure this application to use them using the [chart values](#Values).

There are a few excellent options to run a PostgreSQL database within Kubernetes, to name a few:
- [Bitnami PostgreSQL chart](https://artifacthub.io/packages/helm/bitnami/postgresql)
- [CloudNativePG](https://cloudnative-pg.io/)

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| autoscaling.enabled | bool | `false` |  |
| autoscaling.maxReplicas | int | `100` |  |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| database.host | string | `""` | PostgreSQL database host. |
| database.name | string | `"regauth"` | Database name. |
| database.port | int | `5432` | Database port. |
| database.secretName | string | `""` | Name of the secret containing the database credentials. |
| database.secretPasswordKey | string | `"password"` | Key of the password value within the secret. |
| database.secretUsernameKey | string | `"username"` | Key of the username value within the secret. |
| fullnameOverride | string | `""` |  |
| httpRoute.annotations | object | `{}` |  |
| httpRoute.enabled | bool | `false` |  |
| httpRoute.hostnames | list | `[]` |  |
| httpRoute.parentRefs | list | `[]` |  |
| httpRoute.rules | list | `[]` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"ghcr.io/evanebb/regauth"` |  |
| image.tag | string | `""` |  |
| imagePullSecrets | list | `[]` |  |
| ingress.annotations | object | `{}` |  |
| ingress.className | string | `""` |  |
| ingress.enabled | bool | `false` |  |
| ingress.hosts | list | `[]` |  |
| ingress.tls | list | `[]` |  |
| initialadmin.secretName | string | `""` | Name of the secret containing the credentials for the initial admin user. This will only be used on first application start-up, and can be removed afterward. |
| initialadmin.secretPasswordKey | string | `"password"` | Key of the password value within the secret. |
| initialadmin.secretUsernameKey | string | `"username"` | Key of the username value within the secret. |
| livenessProbe.httpGet.path | string | `"/health"` |  |
| livenessProbe.httpGet.port | string | `"http"` |  |
| log.formatter | string | `"text"` | Format of the log entries, can be either 'text' or 'json'. |
| log.level | string | `"info"` | Log level, can be 'debug', 'info', 'warn' or 'error' |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` |  |
| podAnnotations | object | `{}` |  |
| podLabels | object | `{}` |  |
| podSecurityContext.runAsGroup | int | `65532` |  |
| podSecurityContext.runAsNonRoot | bool | `true` |  |
| podSecurityContext.runAsUser | int | `65532` |  |
| podSecurityContext.seccompProfile.type | string | `"RuntimeDefault"` |  |
| readinessProbe.httpGet.path | string | `"/health"` |  |
| readinessProbe.httpGet.port | string | `"http"` |  |
| replicaCount | int | `1` |  |
| resources | object | `{}` |  |
| securityContext.allowPrivilegeEscalation | bool | `false` |  |
| securityContext.capabilities.drop[0] | string | `"ALL"` |  |
| securityContext.readOnlyRootFilesystem | bool | `true` |  |
| service.port | int | `8000` |  |
| service.type | string | `"ClusterIP"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.automount | bool | `true` |  |
| serviceAccount.create | bool | `true` |  |
| serviceAccount.name | string | `""` |  |
| token.alg | string | `""` | Signing algorithm to use when signing tokens. All asymmetric JWT signing algorithms are supported. This must match the configured private key. |
| token.certFilename | string | `""` | Filename/key of the certificate within the secret. |
| token.certKeyFilename | string | `""` | Filename/key of the certificate private key within the secret. |
| token.issuer | string | `"regauth"` | Issuer set in the tokens generated for the registry. Must match the registry configuration. |
| token.secretName | string | `""` | Name of the secret containing the signing key/certificate. |
| token.service | string | `"registry"` | Audience/service set in the tokens generated for the registry. Must match the registry configuration. |
| tolerations | list | `[]` |  |
| volumeMounts | list | `[]` |  |
| volumes | list | `[]` |  |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.14.2](https://github.com/norwoodj/helm-docs/releases/v1.14.2)
