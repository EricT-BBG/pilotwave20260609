# Pilotwave Helm Chart

完整的 operator-facing Helm 部署流程、image handling 與 values 參數說明集中在 [`../../../docs/HELM.md`](../../../docs/HELM.md)。這份 chart-local README 保留給 Helm package handoff 與 chart 範例參考。環境差異應放在 Helm values、CI/CD 或部署自動化中，不要直接修改 raw Kubernetes manifests。

## 目錄

- [使用方式](#使用方式)
- [最小安裝](#最小安裝)
- [Production / Shared Cluster 部署](#production--shared-cluster-部署)
- [產生給部署者的發佈包](#產生給部署者的發佈包)
- [本機舊版驗證叢集部署](#本機舊版驗證叢集部署)
- [常用 Values](#常用-values)
- [私有映像 Registry](#私有映像-registry)
- [Service 與外部存取](#service-與外部存取)
- [Ingress](#ingress)
- [OpenShift Route](#openshift-route)
- [Monitoring](#monitoring)
- [Workload Runtime Options](#workload-runtime-options)
- [Istio CRD Preflight](#istio-crd-preflight)
- [驗證與解除安裝](#驗證與解除安裝)
- [Raw YAML 狀態](#raw-yaml-狀態)

## 使用方式

本文件中的命令預設在 `build/helm/pilotwave` 目錄內執行：

```sh
helm upgrade --install pilotwave . --namespace pilotwave --create-namespace
```

如果從 repository root 執行，chart path 請改用 `build/helm/pilotwave`：

```sh
helm upgrade --install pilotwave build/helm/pilotwave \
  --namespace pilotwave --create-namespace
```

一般規則：

- production/shared clusters 使用 Helm 直接部署，cluster selection 由 shell 或 CI/CD 控制。
- repo Makefile 的 `helm-upgrade-dev` 是本機 `colima-legacy-1-18` 驗證叢集專用便利 wrapper。
- secrets、tokens、kubeconfigs、private keys 不要提交到 values 檔。
- 複雜 maps 或 lists 使用 values file；簡單 scalar 才用 `--set`。

## 最小安裝

chart 可以不帶 custom values 安裝：

```sh
helm upgrade --install pilotwave . \
  --namespace pilotwave --create-namespace
```

預設會建立：

- `ServiceAccount`
- `ClusterRole`
- `ClusterRoleBinding`
- built-in auth secret
- runtime config `ConfigMap`
- `Deployment`
- `ClusterIP` `Service`，service port `80`，container target port `22112`

預設不會建立：

- PersistentVolumeClaim
- registry pull secret
- ServiceMonitor
- Grafana dashboard ConfigMap
- PrometheusRule
- LoadBalancer / NodePort Service
- Ingress / OpenShift Route

內建 admin 帳號只會在 database 初次初始化時建立為 `admin/admin`。production/shared environments 不要啟用 `PILOTWAVE_DEV_RESET_ADMIN_PASSWORD=true`；那是本機開發啟動時重設密碼用的旗標。

緊急重設密碼時，請在使用相同 ConfigMap 與 database volume 的 Pod 或一次性 Job 中執行：

```sh
kubectl -n pilotwave exec deploy/pilotwave -- \
  /pilotwave --config /config.toml admin reset-password --username admin --password 'new-password'
```

shared environments 建議用短生命週期 Job 從 Kubernetes Secret 讀取密碼，並透過 `PILOTWAVE_ADMIN_PASSWORD` 傳入。

## Production / Shared Cluster 部署

production 建議從範例 values file 開始：

```sh
cp values-production.example.yaml values-prod.yaml
```

調整 `values-prod.yaml` 後部署：

```sh
helm upgrade --install pilotwave . \
  --namespace pilotwave --create-namespace \
  -f values-prod.yaml \
  --set image.repository=docker.io/kenduest/brobridge-pilotwave1 \
  --set image.tag=$IMAGE_TAG \
  --set imagePullSecret.existingSecret=dockerhub-regcred
```

部署前建議從 repository root 執行 chart 檢查：

```sh
make helm-lint
make helm-template
```

production/shared cluster 的基本原則：

- 使用 environment-owned values file，不依賴 Makefile dev defaults。
- 私有 registry token 先建立成 namespace-local Kubernetes Secret，再用 `imagePullSecret.existingSecret` 引用。
- persistence、resources、readiness/liveness probes、monitoring resources 應依環境需求在 values file 中明確設定。若 production 沒有 StorageClass，請由 platform/storage team 先建立 PVC，再用 `persistence.existingClaim` 引用；SQLite 模式維持 `replicaCount=1`。
- kube context selection 放在 shell、CI job 或 Helm `--kube-context`，不要寫死在 repo 文件以外的通用流程。

## 產生給部署者的發佈包

從 repository root 執行：

```sh
make dist
```

預設輸出在 `build/dist/`：

- `pilotwave-<chart-version>.tgz`：Helm chart package。
- `values-production.example.yaml`：production-oriented values template。
- `INSTALL.md`：部署者最短安裝步驟。
- `HELM_README.md`：完整 Helm 說明。
- `SHA256SUMS`：發佈檔案 checksum。

對外交付時，通常給部署者 `build/dist/` 的檔案，加上一個已推送到 registry 的 container image tag。部署者可直接用 chart package 安裝：

```sh
helm upgrade --install pilotwave ./pilotwave-<chart-version>.tgz \
  --namespace pilotwave --create-namespace \
  -f values-prod.yaml \
  --set image.repository=<image-repository> \
  --set image.tag=<image-tag> \
  --set imagePullSecret.existingSecret=<pull-secret-name>
```

## 本機舊版驗證叢集部署

本機舊版驗證叢集預設 context 是 `colima-legacy-1-18`。先固定本次 `IMAGE_TAG`，再把 ARM image build 進該 Colima profile 的 Docker engine：

```sh
eval "$(colima docker-env legacy-pilotware-1-18)"
IMAGE_TAG=$(make -s version | awk -F= '/IMAGE_TAG/ {print $2}')
make build-docker-legacy-local IMAGE_TAG=$IMAGE_TAG
```

部署時把同一個 `IMAGE_TAG` 傳給 dev wrapper：

```sh
make helm-upgrade-dev IMAGE_TAG=$IMAGE_TAG
```

`helm-upgrade-dev` 會套用 `values-legacy-local.example.yaml`，使用 `image.repository=pilotwave`、`image.pullPolicy=IfNotPresent`，並在 Helm apply 後用 `kubectl get` 輪詢 Deployment readiness。

在 Kubernetes 1.18 測試叢集上，避免使用 `helm --wait` 與 `kubectl rollout status`。較新版 Helm/client-go 的 watch bookmark 行為可能 timeout，即使 Deployment 已經 ready。

## 本機 Image 與 Registry Image

Helm 只會把 image 名稱寫進 Deployment，不會把本機 image 傳進 Kubernetes。安裝時有兩種常用路徑：

1. 推到 registry，讓叢集節點 pull：

```sh
IMAGE_TAG=$(make -s version | awk -F= '/IMAGE_TAG/ {print $2}')
make build-docker-release IMAGE_TAG=$IMAGE_TAG
make push-docker-image IMAGE_TAG=$IMAGE_TAG
make helm-upgrade IMAGE_TAG=$IMAGE_TAG \
  HELM_ARGS="--set image.repository=hb.k8sbridge.com/public/pilotwave --set image.tag=$IMAGE_TAG --set image.pullPolicy=Always"
```

2. 本機 cluster runtime 已經看得到 image 時，直接指定本機 image：

```sh
IMAGE_TAG=$(make -s version | awk -F= '/IMAGE_TAG/ {print $2}')
make build-docker-local IMAGE_TAG=$IMAGE_TAG
make helm-upgrade-local-image IMAGE_TAG=$IMAGE_TAG
```

第二種方式適合 Docker Desktop、Colima profile，或已經用 `docker load` / runtime import 把 image 放進目標節點的情境。若 Kubernetes 節點不是使用同一個本機 Docker runtime，請改用 registry push 或先把 image 匯入該 cluster runtime。

## 常用 Values

適合用 CLI 傳入的 scalar values：

```sh
--set image.repository=docker.io/kenduest/brobridge-pilotwave1
--set image.tag=$IMAGE_TAG
--set image.pullPolicy=Always
--set imagePullSecret.existingSecret=dockerhub-regcred
--set imagePullSecret.create=true
--set imagePullSecret.registry=docker.io
--set imagePullSecret.username="$DOCKERHUB_USERNAME"
--set imagePullSecret.password="$DOCKERHUB_TOKEN"
--set service.type=ClusterIP
--set service.type=NodePort
--set service.nodePort=32112
--set service.type=LoadBalancer
--set service.loadBalancerIP=192.168.1.50
--set ingress.enabled=true
--set ingress.host=pilotwave.example.com
--set ingress.className=nginx
--set ingress.tls.enabled=true
--set ingress.tls.secretName=pilotwave-tls
--set route.enabled=true
--set route.host=pilotwave.apps.example.com
--set route.tls.enabled=true
--set route.tls.termination=edge
--set metrics.enabled=true
--set serviceMonitor.enabled=true
--set serviceMonitor.namespace=monitoring
--set grafanaDashboard.enabled=true
--set grafanaDashboard.namespace=monitoring
--set prometheusRule.enabled=true
--set prometheusRule.namespace=monitoring
--set gateway.tlsSecretNamespace=istio-system
--set istio.required=true
--set persistence.enabled=true
--set persistence.existingClaim=pilotwave-data
--set readinessProbe.enabled=true
--set livenessProbe.enabled=true
```

請用 values file 管理下列複雜欄位：`service.annotations`、`ingress.annotations`、`ingress.hosts`、`ingress.tlsEntries`、inline TLS certificate fields、`route.annotations`、`serviceMonitor.labels`、`grafanaDashboard.labels`、`prometheusRule.labels`、`resources`、`podSecurityContext`、`securityContext`、`nodeSelector`、`tolerations`、`affinity`。

## 私有映像 Registry

production/shared clusters 建議使用既有 Secret，避免 token 存入 Helm release values：

```sh
kubectl -n pilotwave create secret docker-registry dockerhub-regcred \
  --docker-server=docker.io \
  --docker-username="$DOCKERHUB_USERNAME" \
  --docker-password="$DOCKERHUB_TOKEN"

helm upgrade --install pilotwave . \
  --namespace pilotwave --create-namespace \
  --set image.repository=docker.io/kenduest/brobridge-pilotwave1 \
  --set image.tag=$IMAGE_TAG \
  --set imagePullSecret.existingSecret=dockerhub-regcred
```

只有 disposable dev/test clusters 才建議讓 chart 建立 registry Secret：

```sh
helm upgrade --install pilotwave . \
  --namespace pilotwave --create-namespace \
  --set image.repository=docker.io/kenduest/brobridge-pilotwave1 \
  --set image.tag=$IMAGE_TAG \
  --set imagePullSecret.create=true \
  --set imagePullSecret.registry=docker.io \
  --set imagePullSecret.username="$DOCKERHUB_USERNAME" \
  --set imagePullSecret.password="$DOCKERHUB_TOKEN"
```

這會在 release namespace 建立 `kubernetes.io/dockerconfigjson` Secret，並掛到 Deployment 的 `imagePullSecrets`。Secret 是 namespace-scoped，不是全叢集設定，也不會跨 namespace 生效。

## Service 與外部存取

chart 支援三種 Service type：

```yaml
service:
  type: ClusterIP
  port: 80
  targetPort: 22112
```

- `ClusterIP`：叢集內存取、port-forward、Ingress 或 Istio Gateway 前置時使用。
- `NodePort`：只建議用於 local labs 或可接受 node direct access 的叢集。
- `LoadBalancer`：叢集需有 cloud provider controller、MetalLB 或其他 LB implementation。

`LoadBalancer` 不是 vanilla Kubernetes 內建能力；沒有 controller 時可能停在 `EXTERNAL-IP: <pending>`。provider-specific 設定請透過 `service.annotations` 放在 values file。

## Ingress

叢集有 Ingress controller，且希望 chart 建立 application ingress 時啟用：

```sh
helm upgrade --install pilotwave . \
  --namespace pilotwave --create-namespace \
  --set service.type=ClusterIP \
  --set ingress.enabled=true \
  --set ingress.host=pilotwave.example.com \
  --set ingress.className=nginx
```

TLS 建議先由 cert-manager 或外部流程建立 Secret，再讓 chart 引用：

```sh
helm upgrade --install pilotwave . \
  --namespace pilotwave --create-namespace \
  --set ingress.enabled=true \
  --set ingress.host=pilotwave.example.com \
  --set ingress.tls.enabled=true \
  --set ingress.tls.secretName=pilotwave-tls
```

dev/test 可用 `--set-file` 讓 chart 建立 TLS Secret，但 certificate material 會進入 Helm release data，不適合 shared/production environments：

```sh
helm upgrade --install pilotwave . \
  --namespace pilotwave --create-namespace \
  --set ingress.enabled=true \
  --set ingress.host=pilotwave.example.com \
  --set ingress.tls.enabled=true \
  --set ingress.tls.createSecret=true \
  --set ingress.tls.secretName=pilotwave-tls \
  --set-file ingress.tls.key=./tls.key \
  --set-file ingress.tls.certificate=./tls.crt
```

chart 會在較新叢集 render `networking.k8s.io/v1` Ingress，並在 Kubernetes 1.18 類型舊叢集 fallback 到 `networking.k8s.io/v1beta1`。

## OpenShift Route

OpenShift/OCP 使用 Route 而不是 Kubernetes Ingress 時啟用：

```sh
helm upgrade --install pilotwave . \
  --namespace pilotwave --create-namespace \
  --set service.type=ClusterIP \
  --set route.enabled=true \
  --set route.host=pilotwave.apps.example.com
```

edge TLS termination：

```sh
helm upgrade --install pilotwave . \
  --namespace pilotwave --create-namespace \
  --set route.enabled=true \
  --set route.host=pilotwave.apps.example.com \
  --set route.tls.enabled=true \
  --set route.tls.termination=edge
```

`route.enabled` 與 `ingress.enabled` 互斥。chart 會在兩者同時啟用、缺少 `route.host`、或非 OpenShift cluster 啟用 Route 時 fail fast。

shared/production OCP 建議使用 router-managed certificates、external secret automation 或受保護的 values file；不要提交 private keys。inline Route certificate fields 會存入 Helm release data。

## Monitoring

Pilotwave 在 HTTP service port 暴露 `/metrics`。chart 預設會在 Service 上加入 Prometheus scrape annotations：

```yaml
metrics:
  enabled: true
  path: /metrics
```

如果叢集使用 Prometheus Operator，建議啟用 `ServiceMonitor`：

```yaml
serviceMonitor:
  enabled: true
  namespace: monitoring
  honorLabels: true
  labels:
    # Must match your Prometheus serviceMonitorSelector.
    release: prometheus-release
```

`serviceMonitor.namespace` 控制 `ServiceMonitor` 物件建立在哪個 namespace；它仍會選取 Helm release namespace 內的 Pilotwave Service。`honorLabels: true` 會保留 Pilotwave metrics 裡的 Istio `namespace` label，避免被 Prometheus target namespace 覆蓋成 `exported_namespace`。

chart 也可選擇建立 Grafana dashboard ConfigMap 與 PrometheusRule：

```yaml
grafanaDashboard:
  enabled: true
  namespace: monitoring
  labels:
    grafana_dashboard: "1"

prometheusRule:
  enabled: true
  namespace: monitoring
  labels:
    # Must match your Prometheus ruleSelector.
    release: prometheus-release
```

PrometheusRule 包含下列 alerts。沒有 Prometheus Operator CRDs 的叢集請保持關閉。

| Alert | Severity | Condition | For | Metrics source |
| --- | --- | --- | --- | --- |
| `PilotwaveIstioTLSCertificateExpiringSoon` | warning | Gateway TLS certificate remaining days `< 30` | 30m | Pilotwave `/metrics` |
| `PilotwaveIstioTLSCertificateExpired` | critical | Gateway TLS certificate is already expired | 5m | Pilotwave `/metrics` |
| `PilotwaveIstioGatewayTLSSecretMissing` | critical | Gateway references a missing TLS credential secret | 10m | Pilotwave `/metrics` |
| `PilotwaveIstioGatewayTLSSecretInvalid` | critical | Gateway references an invalid TLS credential secret | 10m | Pilotwave `/metrics` |
| `PilotwaveKubernetesWriteConflicts` | warning | More than 5 Kubernetes write conflicts for the same resource/verb over 10m | 10m | Pilotwave `/metrics` |
| `PilotwaveHighHTTP5xxRate` | critical | More than 5% Pilotwave HTTP 5xx rate while request rate is above 0.1 req/s | 10m | Pilotwave `/metrics` |
| `PilotwaveHighP95Latency` | warning | Pilotwave HTTP p95 latency above 1s while request rate is above 0.1 req/s | 10m | Pilotwave `/metrics` |

Grafana dashboard 的 Pilotwave app health、Gateway TLS 狀態、certificate expiry panels 只需要 Pilotwave `/metrics`。Istio traffic panels 則需要 Prometheus 另外 scrape Istio proxy metrics，也就是 `istio_requests_total` 與 `istio_request_duration_milliseconds_bucket`。

如果叢集已經有全域 Istio metrics scraping，保持 `istioPodMonitor.enabled=false`，避免重複 scrape。若沒有，chart 可以建立 ingressgateway `PodMonitor`：

```yaml
istioPodMonitor:
  enabled: true
  namespace: monitoring
  labels:
    # Must match your Prometheus podMonitorSelector.
    release: prometheus-release
  targetNamespace: istio-system
  selector:
    app: istio-ingressgateway
  port: http-envoy-prom
  path: /stats/prometheus
```

這個 `PodMonitor` 預設抓 `istio-system` 裡 `app=istio-ingressgateway` 的 Envoy metrics port。若正式環境的 ingressgateway label、namespace 或 port name 不同，請在 values 裡覆寫 `targetNamespace`、`selector` 或 `port`。

## Workload Runtime Options

chart 預設會建立 ServiceAccount。若 RBAC 或 cloud identity binding 由外部管理，請改用既有 ServiceAccount：

```yaml
serviceAccount:
  create: false
  name: pilotwave-sa
```

security contexts、probes、resources、node placement 建議放在 values file：

```yaml
resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi

podSecurityContext:
  runAsNonRoot: true

securityContext:
  allowPrivilegeEscalation: false
  readOnlyRootFilesystem: true

readinessProbe:
  enabled: true

livenessProbe:
  enabled: true
```

container image 使用 distroless `nonroot` base image。OpenShift/OCP 上建議不要固定 `runAsUser`、`runAsGroup`、`fsGroup`，除非 SCC 或 storage policy 明確要求；OCP 可以注入允許的 random UID。

## Istio CRD Preflight

預設 `istio.required=false`，即使 Istio CRDs 不存在也不阻擋安裝。若目標叢集必須已具備 Pilotwave 管理的 Istio APIs，可啟用：

```sh
helm upgrade --install pilotwave . \
  --namespace pilotwave --create-namespace \
  --set istio.required=true
```

預設檢查的 CRDs：

```yaml
istio:
  required: false
  requiredCRDs:
    - gateways.networking.istio.io
    - virtualservices.networking.istio.io
    - destinationrules.networking.istio.io
    - authorizationpolicies.security.istio.io
    - requestauthentications.security.istio.io
```

此檢查使用 Helm `lookup`，只反映 target cluster connection 可見的 CRDs。未連線叢集的 `helm template` 無法滿足 lookup；若同時設定 `istio.required=true`，offline render 會像 CRDs 缺失一樣失敗。

## 驗證與解除安裝

repository root 的本機驗證 targets：

```sh
make helm-lint
make helm-template
make helm-upgrade-dev IMAGE_TAG=$IMAGE_TAG
make helm-wait-ready-dev
make smoke-istio
make e2e-web-istio
```

production/shared cluster 可用 Helm history 與 rollback：

```sh
helm history pilotwave --namespace pilotwave
helm rollback pilotwave <revision> --namespace pilotwave
```

解除安裝：

```sh
helm uninstall pilotwave --namespace pilotwave
```

## Raw YAML 狀態

`manifests/rbac.yaml` 與 `manifests/portal-deploy.yaml` 只保留給直接 `kubectl apply` smoke tests、troubleshooting，或與 chart output 比對。新安裝與可重複部署請優先使用 Helm chart。
