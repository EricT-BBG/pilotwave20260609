# 一鍵部署 Pilotwave（單台 VM）

在已經有 Kubernetes 叢集的 VM 上，把這個 repo 拉下來後執行：

```sh
./deploy
```

腳本會依序完成：

1. **前置檢查** — 確認 `kubectl` 能連到叢集；缺少 `helm` / `istioctl` 時自動安裝。
2. **安裝 Istio** — 安裝 `default` profile（istiod + ingress gateway）。若叢集已經有 Istio 控制平面會自動略過。
3. **從原始碼建置映像** — 用 `build/docker/Dockerfile`（多階段：Vite 前端 + Go 後端 → distroless）在 VM 上建出 `pilotwave:local`。**因為新版 UI 只存在於原始碼，一定要在 VM 上重新 build 才會有新介面**，不能直接用 registry 上的舊映像。
4. **匯入映像** — 自動偵測叢集 runtime（k3s / RKE2 / containerd / microk8s / minikube / kind）並把映像匯入，讓叢集不需要 registry 也能用本地映像。
5. **Helm 部署** — `helm upgrade --install`，以 **NodePort** 對外。
6. **驗證** — 等待 rollout 完成並印出存取網址。

完成後：

```
URL:   http://10.10.7.209:30080
登入：  admin / admin   （第一次登入後請立即修改）
```

## 需求

- 一個可用的 Kubernetes 叢集（`kubectl` 已指向它）。
- 一個映像建置工具：`docker`（建議）、`nerdctl` 或 `podman`。
- 對外網路（用來下載 istioctl 與安裝 Helm）。

## 常用環境變數（執行前 export 即可覆寫）

| 變數 | 預設值 | 說明 |
|------|--------|------|
| `NODE_PORT` | `30080` | 對外 NodePort（範圍 30000–32767）。 |
| `NODE_IP` | `10.10.7.209` | 只用於印出最終存取網址。 |
| `ISTIO_VERSION` | `1.22.6` | 要安裝的 Istio 版本。Pilotwave 使用 `networking.istio.io/v1alpha3` 與 `security.istio.io/v1beta1`，1.20–1.24 都仍支援；若叢集 Kubernetes 版本較舊，請改用對應的舊版 Istio。 |
| `INSTALL_ISTIO` | `auto` | `auto` / `yes` / `no`。`auto` 會在偵測到既有 Istio 時略過安裝。 |
| `NAMESPACE` | `pilotwave` | 部署的 namespace。 |
| `IMAGE_TAG` | `local` | 本地映像 tag。 |

範例：

```sh
NODE_PORT=31000 ISTIO_VERSION=1.20.8 ./deploy
```

## 重新部署 / 升級

腳本是 idempotent 的，再次執行 `./deploy` 會原地升級（會重新 build 映像並 `helm upgrade`）。Auth secret 會快取在 `.deploy/auth-secret`，重跑不會讓既有登入失效。

## 移除

```sh
helm -n pilotwave uninstall pilotwave
```

## 疑難排解

- **Pod 卡在 `ImagePullBackOff`**：表示映像沒有成功匯入叢集 runtime。腳本最後若印出警告，依提示手動匯入，例如：
  ```sh
  docker save pilotwave:local | sudo ctr -n k8s.io images import -
  ```
- **Istio CRD 缺失 / Pilotwave 操作 Istio 資源失敗**：確認 `ISTIO_VERSION` 與叢集 Kubernetes 版本相容，並確認 `kubectl get crd | grep istio` 有出現 gateways / virtualservices / authorizationpolicies / requestauthentications。
- **看 log**：`kubectl -n pilotwave logs deploy/pilotwave -f`
