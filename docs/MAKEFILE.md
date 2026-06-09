# Makefile Guide

The repository `Makefile` is the preferred entrypoint for local development, verification, image handling, Helm deployment, and local Kubernetes/Istio validation.

Run the built-in summary with:

```sh
make help
```

## Common Flows

### Full local verification

```sh
make verify
```

Runs frontend install, lint, tests, Vite build, generated static assets, and Go backend build.

### Local backend against the default validation cluster

```sh
make run-server
```

`run-server` is an alias for `run-server-cluster`. The default cluster context is `colima-legacy-1-18`.

### UI-only backend without Kubernetes/Istio

```sh
make run-server-mock
```

Starts the backend with `PILOTWAVE_CLUSTER_DISABLED=true` for UI/API smoke work that does not need a live cluster.

### Build and push a release image

```sh
IMAGE_TAG=$(make -s version | awk -F= '/IMAGE_TAG/ {print $2}')
make build-docker-release DOCKER_IMAGE=docker.io/kenduest/brobridge-pilotwave1 IMAGE_TAG=$IMAGE_TAG
make scan-image TRIVY_IMAGE=docker.io/kenduest/brobridge-pilotwave1:$IMAGE_TAG
make push-docker-image DOCKER_IMAGE=docker.io/kenduest/brobridge-pilotwave1 IMAGE_TAG=$IMAGE_TAG
```

Docker is the default container runtime. To run the same image targets with
Podman, set `CONTAINER_RUNTIME=podman`:

```sh
make build-docker-release CONTAINER_RUNTIME=podman IMAGE_TAG=$IMAGE_TAG
make push-docker-image CONTAINER_RUNTIME=podman DOCKER_IMAGE=docker.io/kenduest/brobridge-pilotwave1 IMAGE_TAG=$IMAGE_TAG
```

### Build a compressed image archive

```sh
IMAGE_TAG=$(make -s version | awk -F= '/IMAGE_TAG/ {print $2}')
make build-docker-archive IMAGE_TAG=$IMAGE_TAG
```

Default output:

```text
build/images/pilotwave-$IMAGE_TAG.tar.gz
```

Load it later with:

```sh
make load-docker-image IMAGE_TAG=$IMAGE_TAG
```

For a local-cluster image archive that preserves `pilotwave:$IMAGE_TAG` inside
the tarball, use:

```sh
make build-docker-save IMAGE_TAG=$IMAGE_TAG
```

### Deploy with Helm

Registry image:

```sh
make helm-upgrade HELM_ARGS="--set image.repository=docker.io/kenduest/brobridge-pilotwave1 --set image.tag=$IMAGE_TAG"
```

Local image already visible to the cluster runtime:

```sh
make build-docker-local IMAGE_TAG=$IMAGE_TAG
make helm-upgrade-local-image IMAGE_TAG=$IMAGE_TAG
```

Local legacy validation cluster:

```sh
eval "$(colima docker-env legacy-pilotware-1-18)"
make build-docker-legacy-local IMAGE_TAG=$IMAGE_TAG
make helm-upgrade-dev IMAGE_TAG=$IMAGE_TAG
```

See `docs/HELM.md` for complete Helm values and deployment details.

### Local Istio and Helm E2E checks

```sh
ISTIO_CONTEXT=colima-legacy-1-18 make smoke-istio
ISTIO_CONTEXT=colima-legacy-1-18 make e2e-web-istio
make e2e-helm-local-image
```

`smoke-istio` validates ClusterIP-only Istio routing and policy fixtures.
`e2e-web-istio` runs the browser/API suite against the selected
`ISTIO_CONTEXT`. `e2e-helm-local-image` render-checks Helm local-image values,
existing production PVC reuse, ServiceMonitor/PodMonitor output, and the
`istio.required=true` failure message when live Istio CRD lookup is unavailable.

## Variables

### Version and build metadata

| Variable | Default | Purpose |
| --- | --- | --- |
| `APP_VERSION` | contents of `VERSION` | Application version injected into builds |
| `COMMIT` | `git rev-parse --short HEAD` | Git commit injected into builds |
| `BUILD_TIMESTAMP` | UTC timestamp | Used by default image tags |
| `BUILD_TIME` | UTC RFC3339 time | Build metadata injected into Go binary |
| `IMAGE_TAG` | `$(APP_VERSION)-$(BUILD_TIMESTAMP)` | Default image tag |
| `GO_LDFLAGS` | buildinfo values | Go linker flags |

### Container runtime and image archive

| Variable | Default | Purpose |
| --- | --- | --- |
| `CONTAINER_RUNTIME` | `docker` | Container runtime used for inspect, save, load, and push; set to `podman` for Podman |
| `CONTAINER_BUILD_COMMAND` | `docker buildx build` or `podman build` | Build command selected from `CONTAINER_RUNTIME` unless overridden |
| `DOCKER_REPO` | `hb.k8sbridge.com/public/` | Default release image repository prefix |
| `DOCKER_IMAGE` | `$(DOCKER_REPO)pilotwave` | Release image repository |
| `DOCKER_PLATFORM` | `linux/amd64` | Release image platform |
| `DOCKER_BUILD_FLAGS` | `--load` for Docker, empty for Podman | Extra container build flags |
| `DOCKER_ATTEST_FLAGS` | `--sbom=true --provenance=true` for Docker, empty for Podman | Build attestation flags |
| `LOCAL_DOCKER_IMAGE` | `pilotwave` | Local cluster image repository |
| `IMAGE_ARCHIVE_DIR` | `build/images` | Image archive output directory |
| `IMAGE_ARCHIVE` | `build/images/pilotwave-$(IMAGE_TAG).tar.gz` | Compressed image archive path |
| `LEGACY_DOCKER` | `pilotwave:$(IMAGE_TAG)` | Legacy local cluster image tag |
| `LEGACY_DOCKER_PLATFORM` | `linux/arm64` | Legacy local cluster image platform |
| `TRIVY_IMAGE` | `$(DOCKER_IMAGE):$(IMAGE_TAG)` | Image scanned by Trivy |

### Local cluster and runtime

| Variable | Default | Purpose |
| --- | --- | --- |
| `DEV_KUBE_CONTEXT` | `colima-legacy-1-18` | Default local validation context |
| `KUBE_CONTEXT` | `$(DEV_KUBE_CONTEXT)` | Context used by local server targets |
| `ISTIO_CONTEXT` | `$(DEV_KUBE_CONTEXT)` | Context used by smoke/E2E Istio checks |
| `DEV_RESET_ADMIN_PASSWORD` | `true` | Reset local built-in admin password for dev targets |
| `LEGACY_ENV_SCRIPT` | `scripts/legacy-k8s-env.sh` | Local legacy cluster wrapper script |

### Helm

| Variable | Default | Purpose |
| --- | --- | --- |
| `HELM_CHART_DIR` | `build/helm/pilotwave` | Chart directory |
| `HELM_RELEASE` | `pilotwave` | Release name |
| `HELM_NAMESPACE` | `pilotwave` | Release namespace |
| `HELM_PACKAGE_DIR` | `build/package` | `helm package` output directory |
| `HELM_ARGS` | empty | Extra args for `helm template` and `helm upgrade` |
| `HELM_LOCAL_IMAGE_ARGS` | local `pilotwave:$IMAGE_TAG` values | Image args used by `helm-upgrade-local-image` |
| `HELM_DEV_ARGS` | legacy local values and local image overrides | Args used by `helm-upgrade-dev` |
| `HELM_DEV_EXTRA_ARGS` | empty | Extra args appended to `helm-upgrade-dev` |
| `HELM_KUBE_CONTEXT` | empty | Optional kube context passed to Helm/readiness checks |
| `HELM_READY_DEPLOYMENT` | `$(HELM_RELEASE)` | Deployment name polled for readiness |
| `HELM_READY_TIMEOUT` | `180` | Readiness timeout in seconds |
| `DIST_DIR` | `build/dist` | `make dist` output directory |

## Target Reference

### Frontend

| Target | Purpose |
| --- | --- |
| `web-install` | Install frontend dependencies with `npm ci --include=dev` |
| `lint-web` | Run frontend ESLint |
| `test-web` | Run frontend unit tests |
| `build-web` | Build frontend assets with Vite |

### Backend and local run

| Target | Purpose |
| --- | --- |
| `generate-go` | Regenerate embedded frontend/static assets |
| `build-go` | Compile the Go backend |
| `run-server` | Alias for `run-server-cluster` |
| `run-server-cluster` | Start backend against a real Kubernetes/Istio cluster |
| `run-server-mock` | Start backend with cluster access disabled |
| `run-server-k8s` | Legacy alias for `run-server-cluster` |
| `run-server-offline` | Legacy alias for `run-server-mock` |
| `smoke-istio` | Run ClusterIP-only Istio smoke tests |
| `e2e-web-istio` | Run browser/API Istio E2E tests |
| `e2e-helm-local-image` | Render-check Helm local-image, PVC, monitoring, and Istio-required behavior |

### Local legacy Kubernetes/Istio

| Target | Purpose |
| --- | --- |
| `legacy-env-create` | Create the Colima Kubernetes 1.18 local environment |
| `legacy-env-start` | Start the local legacy environment |
| `legacy-env-stop` | Stop the local legacy environment |
| `legacy-env-status` | Show environment status |
| `legacy-env-verify` | Verify environment prerequisites/status |
| `legacy-env-refresh-kubeconfig` | Refresh local kubeconfig for the legacy cluster |
| `legacy-env-install-istio` | Install Istio 1.7.5 |
| `legacy-env-status-istio` | Show Istio status |
| `legacy-env-install-prometheus` | Install Prometheus/Grafana |
| `legacy-env-status-prometheus` | Show Prometheus/Grafana status |
| `install-prometheus` | Alias for Prometheus/Grafana install |
| `status-prometheus` | Alias for Prometheus/Grafana status |
| `legacy-env-verify-pilotwave` | Run Pilotwave smoke and E2E checks |
| `test-legacy-env-script` | Test the legacy env automation wrapper |

### API docs

| Target | Purpose |
| --- | --- |
| `api-docs-sync` | Sync OpenAPI route coverage from Go handlers |
| `verify-api-docs` | Verify generated API docs are current |

### Build and image

| Target | Purpose |
| --- | --- |
| `build` | Build frontend and backend locally |
| `verify` | Run lint, tests, frontend build, and backend build |
| `build-all` | Run `verify`, then build the release image |
| `version` | Print build metadata and default image tag |
| `build-docker` | Alias for `build-docker-release` |
| `build-docker-release` | Build the registry-oriented release image |
| `build-docker-local` | Build `LOCAL_DOCKER_IMAGE:IMAGE_TAG` for local cluster runtimes |
| `build-docker-save` | Build the local image and save it to `IMAGE_ARCHIVE` |
| `build-docker-archive` | Build release image and save compressed archive |
| `save-docker-image` | Save the existing release image to `IMAGE_ARCHIVE` |
| `save-local-docker-image` | Save the existing local image to `IMAGE_ARCHIVE` |
| `load-docker-image` | Load `IMAGE_ARCHIVE` into the configured container runtime |
| `push-docker-image` | Push `DOCKER_IMAGE:IMAGE_TAG` to its registry |
| `build-docker-legacy-local` | Build the local legacy `arm64` image |
| `scan-image` | Scan `TRIVY_IMAGE` with Trivy |

### Helm and distribution

| Target | Purpose |
| --- | --- |
| `helm-lint` | Lint the Helm chart |
| `helm-template` | Render manifests without applying them |
| `helm-upgrade` | Generic Helm install/upgrade |
| `helm-upgrade-local-image` | Helm install/upgrade using local `pilotwave:$IMAGE_TAG` image values |
| `helm-upgrade-dev` | Deploy to the local legacy validation cluster |
| `helm-wait-ready` | Poll Deployment readiness without `kubectl rollout status` |
| `helm-wait-ready-dev` | Poll readiness in the local legacy validation cluster |
| `deploy-yaml-check` | Render chart YAML for a lightweight manifest check |
| `package-helm` | Package chart into `HELM_PACKAGE_DIR` |
| `dist` | Build handoff artifacts into `DIST_DIR` |
