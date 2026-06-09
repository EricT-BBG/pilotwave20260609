# Helm Deployment

This is the operator-facing guide for installing Pilotwave with the Helm chart in `build/helm/pilotwave/`.

Use this file for deployment flows, image handling, and common parameters. The chart package still contains its own `README.md` for packaged handoff artifacts.

## Chart Location

```sh
build/helm/pilotwave
```

Default release settings used by the Makefile:

| Setting | Default |
| --- | --- |
| Release | `pilotwave` |
| Namespace | `pilotwave` |
| Chart path | `build/helm/pilotwave` |
| Local validation context | `colima-legacy-1-18` |

## Validate The Chart

```sh
make helm-lint
make helm-template
```

For a custom namespace or extra values:

```sh
make helm-template HELM_NAMESPACE=pilotwave HELM_ARGS="-f values-prod.yaml"
```

## Install With A Registry Image

Use this path for shared, remote, and production clusters.

```sh
IMAGE_TAG=$(make -s version | awk -F= '/IMAGE_TAG/ {print $2}')

make build-docker-release DOCKER_IMAGE=docker.io/kenduest/brobridge-pilotwave1 IMAGE_TAG=$IMAGE_TAG
make scan-image TRIVY_IMAGE=docker.io/kenduest/brobridge-pilotwave1:$IMAGE_TAG
make push-docker-image DOCKER_IMAGE=docker.io/kenduest/brobridge-pilotwave1 IMAGE_TAG=$IMAGE_TAG

helm upgrade --install pilotwave build/helm/pilotwave \
  --namespace pilotwave --create-namespace \
  --set image.repository=docker.io/kenduest/brobridge-pilotwave1 \
  --set image.tag=$IMAGE_TAG \
  --set image.pullPolicy=Always
```

For private registries, create a pull secret outside Helm and reference it:

```sh
kubectl -n pilotwave create secret docker-registry dockerhub-regcred \
  --docker-server=docker.io \
  --docker-username="$DOCKERHUB_USERNAME" \
  --docker-password="$DOCKERHUB_TOKEN"

helm upgrade --install pilotwave build/helm/pilotwave \
  --namespace pilotwave --create-namespace \
  --set image.repository=docker.io/kenduest/brobridge-pilotwave1 \
  --set image.tag=$IMAGE_TAG \
  --set image.pullPolicy=Always \
  --set imagePullSecret.existingSecret=dockerhub-regcred
```

## Install With A Local Image

Helm does not transfer local images into Kubernetes. The image must already be visible to the node/container runtime, or the cluster must be able to pull it from a registry.

Use this path for Docker Desktop, Colima, Podman, or another local cluster that
shares the container runtime used by the local image build:

```sh
IMAGE_TAG=$(make -s version | awk -F= '/IMAGE_TAG/ {print $2}')

make build-docker-local IMAGE_TAG=$IMAGE_TAG
make helm-upgrade-local-image IMAGE_TAG=$IMAGE_TAG
```

For Podman:

```sh
make build-docker-local CONTAINER_RUNTIME=podman IMAGE_TAG=$IMAGE_TAG
make helm-upgrade-local-image IMAGE_TAG=$IMAGE_TAG
```

`helm-upgrade-local-image` installs with:

| Value | Set to |
| --- | --- |
| `image.repository` | `pilotwave` |
| `image.tag` | `$IMAGE_TAG` |
| `image.pullPolicy` | `IfNotPresent` |

If the local cluster uses a different runtime, save and import the image first:

```sh
make build-docker-archive IMAGE_TAG=$IMAGE_TAG
make load-docker-image IMAGE_TAG=$IMAGE_TAG
```

The default archive path is `build/images/pilotwave-$IMAGE_TAG.tar.gz`. The archive preserves the image name it was built with. If you archive the default release image, deploy with `image.repository=hb.k8sbridge.com/public/pilotwave`; if you need `pilotwave:$IMAGE_TAG`, build the local image with `make build-docker-local` in the target runtime.

## Local Legacy Validation Cluster

The repository's default local validation cluster is `colima-legacy-1-18`.

```sh
eval "$(colima docker-env legacy-pilotware-1-18)"
IMAGE_TAG=$(make -s version | awk -F= '/IMAGE_TAG/ {print $2}')

make build-docker-legacy-local IMAGE_TAG=$IMAGE_TAG
make helm-upgrade-dev IMAGE_TAG=$IMAGE_TAG
```

`helm-upgrade-dev` applies `build/helm/pilotwave/values-legacy-local.example.yaml`, sets the image to `pilotwave:$IMAGE_TAG`, and polls Deployment readiness without `helm --wait`. Avoid `helm --wait` and `kubectl rollout status` on the Kubernetes 1.18 validation cluster because newer client watch behavior can timeout even when the Deployment is ready.

## Production Values File

For production or shared clusters, copy the example file and keep the edited copy outside commits:

```sh
cp build/helm/pilotwave/values-production.example.yaml values-prod.yaml

helm upgrade --install pilotwave build/helm/pilotwave \
  --namespace pilotwave --create-namespace \
  -f values-prod.yaml \
  --set image.repository=docker.io/kenduest/brobridge-pilotwave1 \
  --set image.tag=$IMAGE_TAG
```

Do not commit live registry tokens, TLS private keys, kubeconfigs, or cluster-specific values.

## Production Persistence

Pilotwave currently defaults to SQLite, so production persistence should be explicit.

Recommended production rules:

- Keep `replicaCount=1` while using SQLite.
- If the cluster has a default or named StorageClass, set `persistence.enabled=true` and optionally `persistence.storageClassName=<class>`.
- If the production cluster has no StorageClass, have the platform/storage team create the PV/PVC outside this chart, then set `persistence.existingClaim=<claim-name>`.
- Do not make the application chart responsible for provisioning production PVs.

Dynamic provisioning with a StorageClass:

```sh
helm upgrade --install pilotwave build/helm/pilotwave \
  --namespace pilotwave --create-namespace \
  -f values-prod.yaml \
  --set replicaCount=1 \
  --set persistence.enabled=true \
  --set persistence.storageClassName=<storage-class> \
  --set persistence.size=1Gi
```

No StorageClass, using a pre-created PVC:

```sh
helm upgrade --install pilotwave build/helm/pilotwave \
  --namespace pilotwave --create-namespace \
  -f values-prod.yaml \
  --set replicaCount=1 \
  --set persistence.enabled=true \
  --set persistence.existingClaim=pilotwave-data
```

When `persistence.existingClaim` is set, Helm does not create the PVC. The named claim must already exist in the release namespace.

## Common Parameters

### Image

| Value | Purpose | Example |
| --- | --- | --- |
| `image.repository` | Container image repository | `docker.io/kenduest/brobridge-pilotwave1` |
| `image.tag` | Container image tag | `v1.3-20260522123045` |
| `image.pullPolicy` | Kubernetes pull policy | `Always`, `IfNotPresent`, `Never` |
| `imagePullSecret.existingSecret` | Existing registry pull secret name | `dockerhub-regcred` |
| `imagePullSecret.create` | Let Helm create a pull secret | `false` |
| `imagePullSecret.registry` | Registry server for generated secret | `docker.io` |
| `imagePullSecret.username` | Registry username for generated secret | `$DOCKERHUB_USERNAME` |
| `imagePullSecret.password` | Registry password/token for generated secret | `$DOCKERHUB_TOKEN` |

Prefer `imagePullSecret.existingSecret` for production so registry credentials are not stored in Helm release values.

### Service

| Value | Purpose | Example |
| --- | --- | --- |
| `service.type` | Kubernetes Service type | `ClusterIP`, `NodePort`, `LoadBalancer` |
| `service.port` | Public service port | `80` |
| `service.targetPort` | Pilotwave container port | `22112` |
| `service.nodePort` | Fixed NodePort, or `0` for auto | `32112` |
| `service.loadBalancerIP` | Provider-specific static LB IP | `192.168.1.50` |
| `service.loadBalancerSourceRangesText` | CLI-friendly CIDR list | `192.168.1.0/24 10.0.0.0/8` |
| `service.externalTrafficPolicy` | External traffic policy | `Cluster`, `Local` |

### Ingress

| Value | Purpose | Example |
| --- | --- | --- |
| `ingress.enabled` | Render Kubernetes Ingress | `true` |
| `ingress.className` | Ingress class | `nginx` |
| `ingress.host` | Single-host shortcut | `pilotwave.example.com` |
| `ingress.path` | Single-host path | `/` |
| `ingress.pathType` | Path matching mode | `Prefix` |
| `ingress.tls.enabled` | Add TLS entry | `true` |
| `ingress.tls.createSecret` | Let chart create TLS secret | `false` |
| `ingress.tls.secretName` | Existing TLS secret name | `pilotwave-tls` |

For real TLS data, prefer an existing secret or `--set-file` instead of inline values.

### OpenShift Route

| Value | Purpose | Example |
| --- | --- | --- |
| `route.enabled` | Render OpenShift Route | `true` |
| `route.host` | Route host | `pilotwave.apps.example.com` |
| `route.tls.enabled` | Enable route TLS | `true` |
| `route.tls.termination` | Route TLS termination | `edge`, `passthrough`, `reencrypt` |
| `route.tls.key` | Route TLS private key | use `--set-file` |
| `route.tls.certificate` | Route TLS certificate | use `--set-file` |
| `route.tls.caCertificate` | Optional CA certificate | use `--set-file` |
| `route.tls.destinationCACertificate` | Optional destination CA | use `--set-file` |

### Persistence And Config

| Value | Purpose | Example |
| --- | --- | --- |
| `persistence.enabled` | Enable PVC-backed app data | `true` |
| `persistence.existingClaim` | Use an existing PVC | `pilotwave-data` |
| `persistence.size` | PVC size | `1Gi` |
| `persistence.storageClassName` | StorageClass name | `standard` |
| `persistence.accessModes` | PVC access modes | `ReadWriteOnce` |
| `persistence.mountPath` | Application data mount path | `/pilotwaveDB` |
| `database.type` | Database driver | `sqlite3` |
| `database.dbpath` | SQLite DB path | `./pilotwaveDB/pilotwave.db` |
| `database.dbname` | Database name | `pilotwave` |
| `database.debugMode` | Enable DB debug logging | `false` |
| `auth.method` | Auth backend | `built-in` |
| `auth.secret.value` | Inline auth secret | avoid in production |
| `auth.secret.existingSecret` | Existing auth secret name | `pilotwave-auth` |
| `auth.secret.key` | Key in auth secret | `auth-secret` |

### Grafana And Istio Integration

| Value | Purpose | Example |
| --- | --- | --- |
| `grafana.provider` | Grafana provider mode | `grafana`, `prometheus` |
| `grafana.host` | Grafana or Prometheus host | `grafana.monitoring.svc` |
| `grafana.port` | Integration service port | `80` |
| `grafana.token` | Grafana API token | use a secret in production |
| `grafana.datasourceId` | Grafana datasource ID | `1` |
| `grafana.tls` | Use HTTPS for integration | `true` |
| `grafana.skipTlsVerify` | Skip upstream TLS verification | `false` |
| `gateway.tlsSecretNamespace` | Namespace for Gateway TLS secrets | `istio-system` |
| `istio.required` | Fail install if Istio CRDs are missing | `true` |
| `istio.requiredCRDs` | CRDs checked by preflight | see `values.yaml` |

### Monitoring

| Value | Purpose | Example |
| --- | --- | --- |
| `metrics.enabled` | Expose `/metrics` | `true` |
| `metrics.path` | Metrics endpoint | `/metrics` |
| `serviceMonitor.enabled` | Render ServiceMonitor | `true` |
| `serviceMonitor.namespace` | ServiceMonitor namespace | `monitoring` |
| `serviceMonitor.labels.release` | Prometheus selector label | `kube-prometheus-stack` |
| `istioPodMonitor.enabled` | Render ingressgateway PodMonitor | `true` |
| `grafanaDashboard.enabled` | Render dashboard ConfigMap | `true` |
| `prometheusRule.enabled` | Render alert rules | `true` |

### Workload Runtime

| Value | Purpose | Example |
| --- | --- | --- |
| `replicaCount` | Deployment replica count | `1` |
| `serviceAccount.create` | Create a ServiceAccount | `true` |
| `serviceAccount.name` | Existing or generated ServiceAccount name | `pilotwave` |
| `serviceAccount.annotations` | ServiceAccount annotations | cloud IAM annotations |
| `podAnnotations` | Pod annotations | scrape or policy annotations |
| `resources` | Container resource requests/limits | see Kubernetes docs |
| `podSecurityContext` | Pod security context | `{}` |
| `securityContext` | Container security context | `{}` |
| `nodeSelector` | Node placement selector | `{}` |
| `tolerations` | Pod tolerations | `[]` |
| `affinity` | Pod affinity rules | `{}` |

### Probes

| Value | Purpose | Default |
| --- | --- | --- |
| `readinessProbe.enabled` | Enable readiness probe | `false` |
| `readinessProbe.path` | Readiness HTTP path | `/metrics` |
| `readinessProbe.port` | Readiness port name/number | `pilotwaveport` |
| `readinessProbe.scheme` | HTTP scheme | `HTTP` |
| `readinessProbe.initialDelaySeconds` | Initial delay | `5` |
| `readinessProbe.periodSeconds` | Probe period | `10` |
| `readinessProbe.timeoutSeconds` | Probe timeout | `3` |
| `readinessProbe.failureThreshold` | Failure threshold | `3` |
| `livenessProbe.enabled` | Enable liveness probe | `false` |
| `livenessProbe.path` | Liveness HTTP path | `/metrics` |
| `livenessProbe.port` | Liveness port name/number | `pilotwaveport` |
| `livenessProbe.scheme` | HTTP scheme | `HTTP` |
| `livenessProbe.initialDelaySeconds` | Initial delay | `15` |
| `livenessProbe.periodSeconds` | Probe period | `20` |
| `livenessProbe.timeoutSeconds` | Probe timeout | `3` |
| `livenessProbe.failureThreshold` | Failure threshold | `3` |

## Packaging For Handoff

```sh
make dist IMAGE_TAG=$IMAGE_TAG
```

Output goes to `build/dist/`:

- `pilotwave-<chart-version>.tgz`
- `values-production.example.yaml`
- `INSTALL.md`
- `HELM_README.md`
- `SHA256SUMS`

Hand off `build/dist/` with the pushed image tag or a compressed image archive.

## Readiness And Rollback

Recheck readiness without applying a release:

```sh
make helm-wait-ready HELM_NAMESPACE=pilotwave
```

For the local validation cluster:

```sh
make helm-wait-ready-dev
```

Rollback:

```sh
helm history pilotwave --namespace pilotwave
helm rollback pilotwave <revision> --namespace pilotwave
make helm-wait-ready HELM_NAMESPACE=pilotwave
```
