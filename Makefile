COMMIT := $(or $(COMMIT),$(shell git rev-parse --short HEAD 2>/dev/null || echo unknown))
APP_VERSION := $(or $(APP_VERSION),$(shell tr -d '[:space:]' < VERSION))
BUILD_TIMESTAMP := $(or $(BUILD_TIMESTAMP),$(shell date -u +%Y%m%d%H%M%S))
BUILD_TIME := $(or $(BUILD_TIME),$(shell date -u +%Y-%m-%dT%H:%M:%SZ))
IMAGE_TAG := $(or $(IMAGE_TAG),$(APP_VERSION)-$(BUILD_TIMESTAMP))
GO_LDFLAGS := $(or $(GO_LDFLAGS),-X git.brobridge.com/pilotwave/pilotwave/pkg/buildinfo.Version=$(APP_VERSION) -X git.brobridge.com/pilotwave/pilotwave/pkg/buildinfo.Commit=$(COMMIT) -X git.brobridge.com/pilotwave/pilotwave/pkg/buildinfo.BuildTime=$(BUILD_TIME))

# Image and container scanning
CONTAINER_RUNTIME ?= docker
CONTAINER_BUILD_COMMAND ?= $(if $(filter podman,$(CONTAINER_RUNTIME)),$(CONTAINER_RUNTIME) build,$(CONTAINER_RUNTIME) buildx build)
DOCKER_REPO = hb.k8sbridge.com/public/
DOCKER_IMAGE ?= $(DOCKER_REPO)pilotwave
DOCKER = "$(DOCKER_IMAGE):$(IMAGE_TAG)"
DOCKER_PLATFORM ?= linux/amd64
DOCKER_BUILD_FLAGS ?= $(if $(filter podman,$(CONTAINER_RUNTIME)),,--load)
DOCKER_SBOM ?= true
DOCKER_PROVENANCE ?= true
DOCKER_ATTEST_FLAGS ?= $(if $(filter podman,$(CONTAINER_RUNTIME)),,--sbom=$(DOCKER_SBOM) --provenance=$(DOCKER_PROVENANCE))
IMAGE_ARCHIVE_DIR ?= build/images
IMAGE_ARCHIVE ?= $(IMAGE_ARCHIVE_DIR)/pilotwave-$(IMAGE_TAG).tar.gz
LOCAL_DOCKER_IMAGE ?= pilotwave
LOCAL_DOCKER = "$(LOCAL_DOCKER_IMAGE):$(IMAGE_TAG)"
LEGACY_DOCKER ?= pilotwave:$(IMAGE_TAG)
LEGACY_DOCKER_PLATFORM ?= linux/arm64
LEGACY_DOCKER_BUILD_FLAGS ?= $(if $(filter podman,$(CONTAINER_RUNTIME)),,--load)
TRIVY_IMAGE ?= $(DOCKER)
TRIVY_SEVERITY ?= HIGH,CRITICAL
TRIVY_EXIT_CODE ?= 1
TRIVY_FORMAT ?= table
TRIVY_ARGS ?=

# Application paths
WEB_DIR = ./web
GO_MAIN = ./cmd/pilotwave

# Local runtime
DEV_KUBE_CONTEXT ?= colima-legacy-1-18
KUBE_CONTEXT ?= $(DEV_KUBE_CONTEXT)
ISTIO_CONTEXT ?= $(DEV_KUBE_CONTEXT)
DEV_RESET_ADMIN_PASSWORD ?= true

# Local legacy Kubernetes/Istio automation
LEGACY_ENV_SCRIPT ?= scripts/legacy-k8s-env.sh

# API docs generation
GO_CACHE_DIR ?= $(CURDIR)/tmp/go-build-cache

# Helm packaging
HELM_CHART_DIR ?= build/helm/pilotwave
HELM_RELEASE ?= pilotwave
HELM_NAMESPACE ?= pilotwave
HELM_PACKAGE_DIR ?= build/package
DIST_DIR ?= build/dist
HELM_ARGS ?=
HELM_LOCAL_IMAGE_ARGS ?= --set image.repository=$(LOCAL_DOCKER_IMAGE) --set image.tag=$(IMAGE_TAG) --set image.pullPolicy=IfNotPresent
HELM_DEV_ARGS ?= -f build/helm/pilotwave/values-legacy-local.example.yaml --set image.repository=pilotwave --set image.tag=$(IMAGE_TAG) --set image.pullPolicy=IfNotPresent
HELM_DEV_EXTRA_ARGS ?=
HELM_KUBE_CONTEXT ?=
HELM_READY_DEPLOYMENT ?= $(HELM_RELEASE)
HELM_READY_TIMEOUT ?= 180

.PHONY: all help

.PHONY: web-install lint-web test-web build-web
.PHONY: generate-go build-go run-server run-server-cluster run-server-mock run-server-k8s run-server-offline
.PHONY: smoke-istio e2e-web-istio e2e-helm-local-image

.PHONY: legacy-env-create legacy-env-start legacy-env-stop
.PHONY: legacy-env-status legacy-env-verify legacy-env-refresh-kubeconfig
.PHONY: legacy-env-install-istio legacy-env-status-istio
.PHONY: legacy-env-install-prometheus legacy-env-status-prometheus
.PHONY: install-prometheus status-prometheus
.PHONY: legacy-env-verify-pilotwave test-legacy-env-script

.PHONY: api-docs-sync verify-api-docs

.PHONY: build verify build-all
.PHONY: version
.PHONY: build-docker build-docker-release build-docker-local build-docker-save build-docker-archive save-docker-image save-local-docker-image load-docker-image push-docker-image build-docker-legacy-local scan-image
.PHONY: helm-lint helm-template helm-upgrade helm-upgrade-local-image helm-upgrade-dev helm-wait-ready helm-wait-ready-dev deploy-yaml-check package-helm dist

all: help

help:
	@echo "Available targets:"
	@echo ""
	@echo "Frontend:"
	@echo "  make web-install       - install frontend dependencies"
	@echo "  make lint-web          - lint the Vue 3 frontend"
	@echo "  make test-web          - run frontend unit tests"
	@echo "  make build-web         - build frontend assets with Vite"
	@echo ""
	@echo "Backend and local run:"
	@echo "  make generate-go       - regenerate embedded static assets"
	@echo "  make build-go          - compile the Go backend"
	@echo "  make run-server        - alias for run-server-cluster"
	@echo "  make run-server-cluster - real Kubernetes/Istio dev server against $(DEV_KUBE_CONTEXT)"
	@echo "  make run-server-mock   - mock/no-cluster UI server with Kubernetes/Istio disabled"
	@echo "  make run-server-k8s    - legacy alias for run-server-cluster"
	@echo "  make run-server-offline - legacy alias for run-server-mock"
	@echo "  make smoke-istio       - run ClusterIP-only Istio smoke tests against ISTIO_CONTEXT"
	@echo "  make e2e-web-istio     - run browser E2E tests against ISTIO_CONTEXT"
	@echo "  make e2e-helm-local-image - render-check Helm local image and production PVC settings"
	@echo ""
	@echo "Legacy Kubernetes/Istio:"
	@echo "  make legacy-env-create        - create Colima Kubernetes 1.18 + Istio-ready local env"
	@echo "  make legacy-env-install-istio - install Istio 1.7.5 into the legacy env"
	@echo "  make install-prometheus       - install Prometheus/Grafana into the legacy env"
	@echo "  make status-prometheus        - show Prometheus/Grafana status in the legacy env"
	@echo "  make legacy-env-verify-pilotwave - run Pilotwave Istio smoke and E2E tests"
	@echo "  make test-legacy-env-script      - test the legacy env automation wrapper"
	@echo ""
	@echo "API docs:"
	@echo "  make api-docs-sync   - sync doc/Pilotwave.v1.yaml route coverage from Go handlers"
	@echo "  make verify-api-docs - verify generated API route coverage is current"
	@echo ""
	@echo "Build and image:"
	@echo "  make build        - build frontend and backend locally"
	@echo "  make verify       - lint, test, and build frontend plus backend build"
	@echo "  make version      - show version metadata and default image tag"
	@echo "  Container runtime: CONTAINER_RUNTIME='$(CONTAINER_RUNTIME)' CONTAINER_BUILD_COMMAND='$(CONTAINER_BUILD_COMMAND)'"
	@echo "  make build-docker - alias for build-docker-release"
	@echo "  make build-docker-release - build the release $(DOCKER_PLATFORM) image with the configured container runtime"
	@echo "  make build-docker-local - build LOCAL_DOCKER_IMAGE:IMAGE_TAG for local cluster runtimes"
	@echo "  Build attestations: DOCKER_ATTEST_FLAGS='$(DOCKER_ATTEST_FLAGS)'"
	@echo "  make build-docker-save - build local image and save compressed archive to IMAGE_ARCHIVE"
	@echo "  make build-docker-archive - build image and save compressed archive to IMAGE_ARCHIVE"
	@echo "  make save-docker-image - save current DOCKER image to compressed IMAGE_ARCHIVE"
	@echo "  make save-local-docker-image - save current LOCAL_DOCKER image to compressed IMAGE_ARCHIVE"
	@echo "  make load-docker-image - load compressed IMAGE_ARCHIVE into the configured container runtime"
	@echo "  make push-docker-image - push DOCKER_IMAGE:IMAGE_TAG to its registry"
	@echo "  make build-docker-legacy-local - build the local legacy $(LEGACY_DOCKER_PLATFORM) image"
	@echo "  make scan-image   - scan TRIVY_IMAGE with Trivy"
	@echo "  make build-all    - run verify and then build the container image"
	@echo ""
	@echo "Helm:"
	@echo "  make helm-lint         - lint the Helm chart"
	@echo "  make helm-template     - render Helm manifests without applying them"
	@echo "  make helm-upgrade      - generic Helm install/upgrade; uses current kube context unless HELM_KUBE_CONTEXT is set"
	@echo "  make helm-upgrade-local-image - Helm install/upgrade using LOCAL_DOCKER_IMAGE:IMAGE_TAG"
	@echo "  make helm-upgrade-dev  - dev install/upgrade into $(DEV_KUBE_CONTEXT) with legacy-local values"
	@echo "  make helm-wait-ready   - generic readiness poll without kubectl watch"
	@echo "  make helm-wait-ready-dev - dev readiness poll in $(DEV_KUBE_CONTEXT)"
	@echo "  make deploy-yaml-check - render and parse deployment YAML without applying it"
	@echo "  make package-helm      - package the Helm chart into HELM_PACKAGE_DIR"
	@echo "  make dist              - build user-facing Helm release artifacts into DIST_DIR"

# Frontend
web-install:
	cd $(WEB_DIR) && npm ci --include=dev

lint-web: web-install
	cd $(WEB_DIR) && npm run lint

test-web: web-install
	cd $(WEB_DIR) && npm test

build-web: web-install
	rm -rf $(WEB_DIR)/dist
	cd $(WEB_DIR) && VITE_APP_VERSION="$(APP_VERSION)" VITE_BUILD_TIMESTAMP="$(BUILD_TIME)" npm run build

# Backend and local run
generate-go:
	mkdir -p $(GO_CACHE_DIR)
	GOCACHE=$(GO_CACHE_DIR) go generate $(GO_MAIN)

build-go: generate-go
	mkdir -p $(GO_CACHE_DIR)
	GOCACHE=$(GO_CACHE_DIR) go build -ldflags "$(GO_LDFLAGS)" $(GO_MAIN)

run-server: run-server-cluster

run-server-cluster: build-web generate-go
	@if [ -n "$(KUBE_CONTEXT)" ]; then \
		tmp_kubeconfig="$${TMPDIR:-/tmp}/pilotwave-$(KUBE_CONTEXT).kubeconfig"; \
		kubectl config view --raw --flatten --minify --context "$(KUBE_CONTEXT)" > "$$tmp_kubeconfig"; \
		echo "Running Pilotwave real-cluster dev server with Kubernetes context $(KUBE_CONTEXT)"; \
		mkdir -p $(GO_CACHE_DIR); \
		GOCACHE=$(GO_CACHE_DIR) KUBECONFIG="$$tmp_kubeconfig" PILOTWAVE_DEV_RESET_ADMIN_PASSWORD=$(DEV_RESET_ADMIN_PASSWORD) go run -ldflags "$(GO_LDFLAGS)" $(GO_MAIN); \
	else \
		echo "Running Pilotwave real-cluster dev server with current KUBECONFIG context"; \
		mkdir -p $(GO_CACHE_DIR); \
		GOCACHE=$(GO_CACHE_DIR) PILOTWAVE_DEV_RESET_ADMIN_PASSWORD=$(DEV_RESET_ADMIN_PASSWORD) go run -ldflags "$(GO_LDFLAGS)" $(GO_MAIN); \
	fi

run-server-mock: build-web generate-go
	@echo "Running Pilotwave mock/no-cluster dev server; Kubernetes and Istio APIs are disabled"
	mkdir -p $(GO_CACHE_DIR)
	GOCACHE=$(GO_CACHE_DIR) PILOTWAVE_CLUSTER_DISABLED=true PILOTWAVE_DEV_RESET_ADMIN_PASSWORD=true go run -ldflags "$(GO_LDFLAGS)" $(GO_MAIN)

run-server-k8s: run-server-cluster

run-server-offline: run-server-mock

smoke-istio:
	ISTIO_CONTEXT="$(ISTIO_CONTEXT)" bash testcase/istio-smoke/smoke.sh

e2e-web-istio:
	ISTIO_CONTEXT="$(ISTIO_CONTEXT)" bash testcase/istio/web-e2e.sh

e2e-helm-local-image:
	bash testcase/helm-local-image-e2e-test.sh

# Legacy Kubernetes/Istio environment
legacy-env-create:
	bash $(LEGACY_ENV_SCRIPT) create

legacy-env-start:
	bash $(LEGACY_ENV_SCRIPT) start

legacy-env-stop:
	bash $(LEGACY_ENV_SCRIPT) stop

legacy-env-status:
	bash $(LEGACY_ENV_SCRIPT) status

legacy-env-verify:
	bash $(LEGACY_ENV_SCRIPT) verify

legacy-env-refresh-kubeconfig:
	bash $(LEGACY_ENV_SCRIPT) refresh-kubeconfig

legacy-env-install-istio:
	bash $(LEGACY_ENV_SCRIPT) install-istio

legacy-env-status-istio:
	bash $(LEGACY_ENV_SCRIPT) status-istio

legacy-env-install-prometheus:
	bash $(LEGACY_ENV_SCRIPT) install-prometheus

legacy-env-status-prometheus:
	bash $(LEGACY_ENV_SCRIPT) status-prometheus

install-prometheus: legacy-env-install-prometheus

status-prometheus: legacy-env-status-prometheus

legacy-env-verify-pilotwave:
	bash $(LEGACY_ENV_SCRIPT) verify-pilotwave

test-legacy-env-script:
	bash testcase/legacy-k8s-env-test.sh

# API docs
api-docs-sync:
	mkdir -p $(GO_CACHE_DIR)
	GOCACHE=$(GO_CACHE_DIR) go run ./scripts/api-doc-sync.go --input doc/Pilotwave.v1.yaml --output doc/Pilotwave.v1.yaml

verify-api-docs:
	bash testcase/api-doc-routes-test.sh
	mkdir -p $(GO_CACHE_DIR)
	@tmp="$$(mktemp /tmp/pilotwave-openapi.XXXXXX.yaml)"; \
	trap 'rm -f "$$tmp"' EXIT; \
	GOCACHE=$(GO_CACHE_DIR) go run ./scripts/api-doc-sync.go --input doc/Pilotwave.v1.yaml --output "$$tmp"; \
	diff -u doc/Pilotwave.v1.yaml "$$tmp"

# Aggregate build and verification
version:
	@echo "APP_VERSION=$(APP_VERSION)"
	@echo "COMMIT=$(COMMIT)"
	@echo "BUILD_TIME=$(BUILD_TIME)"
	@echo "IMAGE_TAG=$(IMAGE_TAG)"

build: build-web build-go

verify: lint-web test-web build-web build-go

build-all: verify build-docker

# Container image
build-docker: build-docker-release

build-docker-release:
	$(CONTAINER_BUILD_COMMAND) --platform $(DOCKER_PLATFORM) $(DOCKER_BUILD_FLAGS) $(DOCKER_ATTEST_FLAGS) --build-arg APP_VERSION="$(APP_VERSION)" --build-arg COMMIT="$(COMMIT)" --build-arg BUILD_TIME="$(BUILD_TIME)" -t $(DOCKER) -f build/docker/Dockerfile .

build-docker-local:
	$(CONTAINER_BUILD_COMMAND) --platform $(DOCKER_PLATFORM) $(DOCKER_BUILD_FLAGS) $(DOCKER_ATTEST_FLAGS) --build-arg APP_VERSION="$(APP_VERSION)" --build-arg COMMIT="$(COMMIT)" --build-arg BUILD_TIME="$(BUILD_TIME)" -t $(LOCAL_DOCKER) -f build/docker/Dockerfile .

build-docker-save: build-docker-local save-local-docker-image

build-docker-archive: build-docker-release save-docker-image

save-docker-image:
	mkdir -p $(IMAGE_ARCHIVE_DIR)
	$(CONTAINER_RUNTIME) image inspect $(DOCKER) >/dev/null
	$(CONTAINER_RUNTIME) save $(DOCKER) | gzip -c > $(IMAGE_ARCHIVE)
	@echo "Saved compressed image archive: $(IMAGE_ARCHIVE)"
	@echo "Load later with: $(CONTAINER_RUNTIME) load -i $(IMAGE_ARCHIVE)"

save-local-docker-image:
	mkdir -p $(IMAGE_ARCHIVE_DIR)
	$(CONTAINER_RUNTIME) image inspect $(LOCAL_DOCKER) >/dev/null
	$(CONTAINER_RUNTIME) save $(LOCAL_DOCKER) | gzip -c > $(IMAGE_ARCHIVE)
	@echo "Saved compressed local image archive: $(IMAGE_ARCHIVE)"
	@echo "Image inside archive: $(LOCAL_DOCKER)"
	@echo "Load later with: $(CONTAINER_RUNTIME) load -i $(IMAGE_ARCHIVE)"

load-docker-image:
	$(CONTAINER_RUNTIME) load -i $(IMAGE_ARCHIVE)

push-docker-image:
	$(CONTAINER_RUNTIME) image inspect $(DOCKER) >/dev/null
	$(CONTAINER_RUNTIME) push $(DOCKER)

build-docker-legacy-local:
	$(CONTAINER_BUILD_COMMAND) --platform $(LEGACY_DOCKER_PLATFORM) $(LEGACY_DOCKER_BUILD_FLAGS) $(DOCKER_ATTEST_FLAGS) --build-arg APP_VERSION="$(APP_VERSION)" --build-arg COMMIT="$(COMMIT)" --build-arg BUILD_TIME="$(BUILD_TIME)" -t $(LEGACY_DOCKER) -f build/docker/Dockerfile .

scan-image:
	trivy image --severity $(TRIVY_SEVERITY) --exit-code $(TRIVY_EXIT_CODE) --format $(TRIVY_FORMAT) $(TRIVY_ARGS) $(TRIVY_IMAGE)

# Helm
helm-lint:
	helm lint $(HELM_CHART_DIR) $(HELM_ARGS)

helm-template:
	helm template $(HELM_RELEASE) $(HELM_CHART_DIR) --namespace $(HELM_NAMESPACE) $(HELM_ARGS)

helm-upgrade:
	helm upgrade --install $(HELM_RELEASE) $(HELM_CHART_DIR) --namespace $(HELM_NAMESPACE) --create-namespace $(if $(HELM_KUBE_CONTEXT),--kube-context $(HELM_KUBE_CONTEXT),) $(HELM_ARGS)
	$(MAKE) helm-wait-ready HELM_RELEASE=$(HELM_RELEASE) HELM_NAMESPACE=$(HELM_NAMESPACE) HELM_KUBE_CONTEXT="$(HELM_KUBE_CONTEXT)" HELM_READY_DEPLOYMENT=$(HELM_READY_DEPLOYMENT) HELM_READY_TIMEOUT=$(HELM_READY_TIMEOUT)

helm-upgrade-local-image:
	$(MAKE) helm-upgrade HELM_RELEASE=$(HELM_RELEASE) HELM_NAMESPACE=$(HELM_NAMESPACE) HELM_KUBE_CONTEXT="$(HELM_KUBE_CONTEXT)" HELM_ARGS="$(HELM_ARGS) $(HELM_LOCAL_IMAGE_ARGS)" HELM_READY_DEPLOYMENT=$(HELM_READY_DEPLOYMENT) HELM_READY_TIMEOUT=$(HELM_READY_TIMEOUT)

helm-upgrade-dev:
	$(MAKE) helm-upgrade HELM_RELEASE=$(HELM_RELEASE) HELM_NAMESPACE=$(HELM_NAMESPACE) HELM_KUBE_CONTEXT="$(DEV_KUBE_CONTEXT)" HELM_ARGS="$(HELM_DEV_ARGS) $(HELM_DEV_EXTRA_ARGS)" HELM_READY_DEPLOYMENT=$(HELM_READY_DEPLOYMENT) HELM_READY_TIMEOUT=$(HELM_READY_TIMEOUT)

helm-wait-ready:
	bash scripts/k8s-wait-ready.sh $(if $(HELM_KUBE_CONTEXT),--context "$(HELM_KUBE_CONTEXT)",) --namespace "$(HELM_NAMESPACE)" --deployment "$(HELM_READY_DEPLOYMENT)" --timeout "$(HELM_READY_TIMEOUT)"

helm-wait-ready-dev:
	$(MAKE) helm-wait-ready HELM_NAMESPACE=$(HELM_NAMESPACE) HELM_KUBE_CONTEXT="$(DEV_KUBE_CONTEXT)" HELM_READY_DEPLOYMENT=$(HELM_READY_DEPLOYMENT) HELM_READY_TIMEOUT=$(HELM_READY_TIMEOUT)

deploy-yaml-check:
	helm template $(HELM_RELEASE) $(HELM_CHART_DIR) --namespace $(HELM_NAMESPACE) $(HELM_ARGS) >/dev/null

package-helm:
	mkdir -p $(HELM_PACKAGE_DIR)
	helm package $(HELM_CHART_DIR) --destination $(HELM_PACKAGE_DIR)

dist:
	HELM_CHART_DIR="$(HELM_CHART_DIR)" DIST_DIR="$(DIST_DIR)" IMAGE_TAG="$(IMAGE_TAG)" bash scripts/build-helm-dist.sh
