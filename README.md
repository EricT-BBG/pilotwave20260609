# Pilotwave

Pilotwave 是一套服務網格管理工具，用於微服務路由、流量控制，以及 Istio 資源操作。

> **一鍵部署：** 在已有 Kubernetes 叢集的 VM 上拉下本 repo 後執行 `./deploy`，會自動安裝 Istio、從原始碼建置映像並以 NodePort 部署。詳見 [`DEPLOY.md`](DEPLOY.md)。

## 目錄

- [系統需求](#系統需求)
- [建置需求](#建置需求)
- [專案文件](#專案文件)
- [安裝與建置](#安裝與建置)
  - [完整本機驗證](#完整本機驗證)
  - [前端](#前端)
  - [後端](#後端)
- [執行](#執行)
  - [前端開發模式](#前端開發模式)
- [容器映像](#容器映像)
- [Istio 管理 UI](#istio-管理-ui)
- [發佈給部署者](#發佈給部署者)
- [Helm 部署](#helm-部署)
- [監控](#監控)
- [疑難排解](#疑難排解)
- [測試](#測試)
- [Swagger API 文件](#swagger-api-文件)

## 系統需求

* Kubernetes
* Istio
* Helm 3，用於封裝後的 Kubernetes 安裝流程
* Docker Buildx 或 Podman，用於映像建置、保存、載入與推送流程
* Trivy，用於映像安全掃描

## 建置需求

* Go 1.25+，新版 Go 相依圖需要此版本
* 建議使用 Node.js 20+，搭配 Vue 3 + Vite 前端
* 前端執行期使用 Vite 風格的 `VITE_*` 環境變數，並保留 `VUE_APP_*` 名稱給既有本機腳本相容使用

## 專案文件

其他專案筆記位於 [`docs/`](docs/)：

* [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) 說明後端、前端架構與資料流。
* [`docs/AI_CHANGE_PLAYBOOK.md`](docs/AI_CHANGE_PLAYBOOK.md) 記錄安全使用 AI 輔助修改的指引。
* [`docs/MAKEFILE.md`](docs/MAKEFILE.md) 說明 Makefile targets、變數與常用 workflow。
* [`docs/HELM.md`](docs/HELM.md) 說明 Helm 安裝、image workflow、values 參數與 rollback。
* [`docs/LOCAL_LEGACY_K8S_ENVIRONMENT.md`](docs/LOCAL_LEGACY_K8S_ENVIRONMENT.md) 記錄獨立的 Colima/k3s/Istio 1.7.5 本機驗證環境。
* [`docs/RELEASE_CHECKLIST.md`](docs/RELEASE_CHECKLIST.md) 說明建置、掃描、推送、Helm 部署、驗證與回復步驟。
* [`docs/TODO.md`](docs/TODO.md) 追蹤目前的 Istio 相容性與安全更新待辦。
* [`docs/FRONTEND.md`](docs/FRONTEND.md) 記錄目前 Vue 3 + Vite 前端架構、開發方式與驗證命令。
* [`docs/DEPENDENCY_SECURITY.md`](docs/DEPENDENCY_SECURITY.md) 記錄相依套件安全狀態與驗證命令。

## 安裝與建置

可以使用下列命令編譯 Pilotwave。

#### 完整本機驗證

```shell
make verify
```

此流程會執行前端安裝、lint、測試、Vite 建置、靜態資產產生，以及 Go 後端建置。

#### 前端

```shell
cd web
npm ci --include=dev
npm run lint
npm run build
npm test
```

#### 後端

```shell
go generate ./cmd/pilotwave
go build ./cmd/pilotwave
```

目前 source repository 位於 GitHub：

```shell
git clone https://github.com/BrobridgeOrg/pilotwave.git
```

## 執行

```shell
KUBECONFIG=YOUR_KUBECONFIG ./pilotwave
KUBECONFIG=YOUR_KUBECONFIG ./pilotwave --config /path/to/config.toml
```

透過 Makefile 進行本機開發：

```shell
make run-server
```

`make run-server` 是 `make run-server-cluster` 的別名。它會建置前端、重新產生內嵌靜態資產、為開發 context 建立暫時 kubeconfig，並使用本機舊版 Kubernetes/Istio 叢集啟動 Go server。預設 Kubernetes context 是 `colima-legacy-1-18`。

當執行模式需要明確指定時，請使用下列 targets：

```shell
make run-server-cluster
make run-server-mock
```

`make run-server-mock` 會以 `PILOTWAVE_CLUSTER_DISABLED=true` 啟動，適合 UI-only smoke tests。它不會連線到 Kubernetes 或 Istio。若要使用其他實際叢集且不變更全域 current context，請傳入 `KUBE_CONTEXT`：

```shell
make run-server-cluster KUBE_CONTEXT=my-context
```

若要建立或維護本機舊版 Kubernetes/Istio 驗證環境，請使用儲存庫內的 wrapper：

```shell
make legacy-env-create
make legacy-env-install-istio
make install-prometheus
make status-prometheus
make legacy-env-verify-pilotwave
```

Prometheus/Grafana 是獨立安裝步驟，因為 `legacy-env-create` 只會建立 Kubernetes baseline。完整的 Colima/k3s/Istio/monitoring 設定與 dry-run 命令，請參考 [`docs/LOCAL_LEGACY_K8S_ENVIRONMENT.md`](docs/LOCAL_LEGACY_K8S_ENVIRONMENT.md)。

新初始化 SQLite DB 的預設內建登入帳號：

```text
username: admin
password: admin
```

僅限本機開發時，`make run-server`、`make run-server-cluster` 與 `make run-server-mock` 會設定 `PILOTWAVE_DEV_RESET_ADMIN_PASSWORD=true`，讓內建 admin 帳號在啟動時回復為 `admin/admin`。請勿在 production 啟用此旗標。

若需要明確執行緊急重設，請使用同一份 config 與 database 執行 CLI：

```shell
PILOTWAVE_ADMIN_PASSWORD='new-password' ./pilotwave --config /path/to/config.toml admin reset-password --username admin
./pilotwave --config /path/to/config.toml admin reset-password --username admin --password-file /path/to/password.txt
```

未指定 `--config` 時，Pilotwave 會依序尋找 `./config.toml` 與 `./configs/config.toml`。在 Kubernetes 中，請於 Pilotwave Pod 內執行相同命令，或使用掛載相同 config 與 database volume 的一次性 Job。

#### 前端開發模式

```shell
cd web
npm run dev
npm run lint
```

使用 Vite dev server 時，請視需要將前端 API 環境變數指向 Go server。設定可使用 `VITE_API_URL` / `VITE_API_VERSION`，既有本機腳本仍可用 `VUE_APP_API_URL` / `VUE_APP_API_VERSION`。

## 容器映像

儲存庫版本記錄於 `VERSION`，目前建置版本為 `v1.3`。預設映像 tag 為 `$(APP_VERSION)-$(BUILD_TIMESTAMP)`，例如 `v1.3-20260522123045`。

Makefile 將 registry release image、本機 cluster image、legacy Colima image，以及 compressed image archive 分成不同 targets。完整 target 與變數說明請看 [`docs/MAKEFILE.md`](docs/MAKEFILE.md)。

常用 release path：

```shell
make version
IMAGE_TAG=$(make -s version | awk -F= '/IMAGE_TAG/ {print $2}')
make build-docker-release DOCKER_IMAGE=docker.io/kenduest/brobridge-pilotwave1 IMAGE_TAG=$IMAGE_TAG
make scan-image TRIVY_IMAGE=docker.io/kenduest/brobridge-pilotwave1:$IMAGE_TAG
make push-docker-image DOCKER_IMAGE=docker.io/kenduest/brobridge-pilotwave1 IMAGE_TAG=$IMAGE_TAG
```

預設 container runtime 是 Docker；若要用 Podman，對同一組 target 加上
`CONTAINER_RUNTIME=podman`：

```shell
make build-docker-release CONTAINER_RUNTIME=podman DOCKER_IMAGE=docker.io/kenduest/brobridge-pilotwave1 IMAGE_TAG=$IMAGE_TAG
make push-docker-image CONTAINER_RUNTIME=podman DOCKER_IMAGE=docker.io/kenduest/brobridge-pilotwave1 IMAGE_TAG=$IMAGE_TAG
```

Compressed archive path：

```shell
make build-docker-archive IMAGE_TAG=$IMAGE_TAG
```

Helm 不會自動把本機 image 傳進 Kubernetes；Deployment 只會引用 chart values 中的 image 名稱。registry、本機 image、archive import 與 legacy Colima 的完整使用方式請看 [`docs/HELM.md`](docs/HELM.md)。

## Istio 管理 UI

Pilotwave 前端是 Vue 3 + Vite SPA，主要頁面依 Istio/Kubernetes 資源分成：

* `Gateway 管理`：管理 Istio `Gateway`，可指定 namespace、server/host/port、TLS credential，以及 `spec.selector` labels。未指定 selector 時，後端仍預設使用 `istio=ingressgateway`；若在 UI 指定 selector labels，Pilotwave 會保留使用者設定，不再強制注入 `istio` label。
* `VirtualService 管理`：管理 Istio `VirtualService`。內部仍有部分 `router` module/API 命名，使用者可見 UI 統一顯示為 VirtualService。建立時可選擇關聯 Gateway；detail 頁預設為檢視模式，按「編輯」才會修改基本設定或路由規則。
* `API 認證機制`：管理 Istio `RequestAuthentication`。
* `黑白名單`：管理 Istio `AuthorizationPolicy`。
* `TLS 憑證管理`：盤點 Gateway 參照的 TLS credential secret 與憑證到期狀態。Istio ingress gateway 讀取 credential secret 的 namespace 取決於 ingress gateway workload 所在 namespace；預設 Helm 設定是 `gateway.tlsSecretNamespace=istio-system`。

Gateway、VirtualService、API 認證機制、黑白名單 detail 頁都採用「先檢視、再編輯」的互動模型，避免使用者進入 detail 就直接改到 cluster-backed resource。namespace 下拉會顯示 Istio sidecar injection 狀態；建立 cluster-backed Istio 資源時，若 namespace 尚未啟用 injection，UI 會先提醒再送出。

## 發佈給部署者

若要產生給其他人部署用的 Helm 發佈包，使用：

```shell
make dist
```

預設輸出在 `build/dist/`，包含：

* `pilotwave-<chart-version>.tgz`：Helm chart package。
* `values-production.example.yaml`：production values 範本。
* `INSTALL.md`：給部署者的最短安裝步驟。
* `HELM_DEPLOYMENT.md`：完整 Helm 部署流程、image workflow 與參數說明。
* `HELM_README.md`：chart-local Helm 範例與 template 參考。
* `SHA256SUMS`：發佈檔案 checksum。

對外交付時，通常提供 `build/dist/` 內的檔案與已推送的 container image tag。部署者不需要整個 source repository。

## Helm 部署

[`build/helm/pilotwave/`](build/helm/pilotwave/) 中的 Helm chart 是建議的 Kubernetes 封裝方式。完整安裝、production/shared cluster 部署、private registry、本機 image、Ingress、OpenShift Route、monitoring values、rollback 與常用參數集中在 [`docs/HELM.md`](docs/HELM.md)。

最小 registry image 範例：

```shell
helm upgrade --install pilotwave build/helm/pilotwave \
  --namespace pilotwave --create-namespace \
  --set image.repository=docker.io/kenduest/brobridge-pilotwave1 \
  --set image.tag=$IMAGE_TAG
```

常用檢查：

```shell
make helm-lint
make helm-template
make package-helm
```

## 監控

Pilotwave 會在 `/metrics` 暴露 application 與 Go runtime metrics。chart 預設啟用 Prometheus scrape annotations，並可選擇 render：

* `serviceMonitor.enabled=true`：讓 Prometheus Operator scrape Pilotwave app metrics。
* `istioPodMonitor.enabled=true`：當叢集尚未 scrape Istio proxy metrics 時，建立 ingressgateway PodMonitor，供 Grafana Istio traffic panels 使用。
* `grafanaDashboard.enabled=true`：產生內建 Grafana dashboard ConfigMap。
* `prometheusRule.enabled=true`：產生 Pilotwave alerts，包含 TLS certificate expiry、invalid Gateway TLS secret state、conflict rate、5xx rate 與 p95 latency。

Gateway detail page 也包含 TLS Certificates tab。它會讀取 Gateway 參照的 credential secrets，並顯示 certificate status、expiry time、remaining days、issuer、subject、SANs 與 SHA-256 fingerprint。它不會將 private keys 或 raw certificate data 回傳到瀏覽器。

原始 monitoring assets 也位於 [`manifests/grafana/`](manifests/grafana/) 與 [`manifests/prometheus/`](manifests/prometheus/)。

## 疑難排解

如果既有 checkout 還指向舊的 `git.brobridge.com` remote，請改成新的 GitHub repository：

```shell
git remote set-url origin https://github.com/BrobridgeOrg/pilotwave.git
git remote -v
```

## 測試

執行完整本機驗證流程：

```shell
make verify
```

針對執行中的 server 執行 legacy httpexpect testbench：

```shell
cd ./unit
go test unit_test.go -v
```

## Swagger API 文件

```shell
go get -u github.com/swaggo/swag/cmd/swag
swag init -g ./pkg/http_server/api/swagger.go -o pkg/http_server/api/docs
```

連線至 http://127.0.0.1:22112/swagger/v1/index.html

已提交的 OpenAPI 3 檔案是 [`doc/Pilotwave.v1.yaml`](doc/Pilotwave.v1.yaml)。請讓 route coverage 與已註冊的 Go handlers 保持同步：

```shell
make api-docs-sync
make verify-api-docs
```

`api-docs-sync` 會保留既有手寫 operation details，並為 `doc/Pilotwave.v1.yaml` 中缺少的 Go routes 新增 stubs。Request 與 response schemas 仍應在 OpenAPI 檔案中細修，或逐步遷移到 handler annotations。
