# Release Checklist

Use this checklist for Pilotwave image and Helm releases. Keep release-specific tags, registry names, and credentials outside committed files unless they are intentionally public. For Makefile target details, use `docs/MAKEFILE.md`. For Helm values and install modes, use `docs/HELM.md`.

## 1. Prepare

- Start from a clean release branch and check for unrelated local edits:

  ```sh
  git status --short
  ```

- Pick the image tag. The Makefile defaults to `VERSION` plus a UTC timestamp,
  for example `v1.3-20260522123045`:

  ```sh
  make version
  IMAGE_TAG=$(make -s version | awk -F= '/IMAGE_TAG/ {print $2}')
  ```

- Review chart changes before deployment:

  ```sh
  make helm-lint
  make helm-template
  ```

## 2. Build And Scan

- Run the local verification gate when dependencies are available:

  ```sh
  make verify
  ```

- Build the release image. Docker Buildx is the default runtime path;
  `make build-docker` is kept as a compatibility alias for
  `make build-docker-release`.

  ```sh
  make build-docker-release DOCKER_IMAGE=docker.io/kenduest/brobridge-pilotwave1 IMAGE_TAG=$IMAGE_TAG
  ```

  For Podman, set `CONTAINER_RUNTIME=podman` on image targets:

  ```sh
  make build-docker-release CONTAINER_RUNTIME=podman DOCKER_IMAGE=docker.io/kenduest/brobridge-pilotwave1 IMAGE_TAG=$IMAGE_TAG
  ```

- If the release needs an offline or handoff image artifact, save the built
  image as a compressed archive:

  ```sh
  make build-docker-archive DOCKER_IMAGE=docker.io/kenduest/brobridge-pilotwave1 IMAGE_TAG=$IMAGE_TAG
  ```

  The default archive path is
  `build/images/pilotwave-$IMAGE_TAG.tar.gz`. Load it later with:

  ```sh
  make load-docker-image IMAGE_TAG=$IMAGE_TAG
  ```

  Helm will not transfer this archive into Kubernetes. The target cluster must
  either be able to pull the image from a registry or already have the image
  imported into the node/container runtime.

- Scan the exact image that will be released:

  ```sh
  make scan-image TRIVY_IMAGE=docker.io/kenduest/brobridge-pilotwave1:$IMAGE_TAG
  ```

- For the local legacy Colima cluster only, build the `arm64` image into the
  Docker engine used by that profile instead of treating it as a release image:

  ```sh
  eval "$(colima docker-env legacy-pilotware-1-18)"
  make build-docker-legacy-local IMAGE_TAG=$IMAGE_TAG
  ```

  Deploy the local image with `image.repository=pilotwave`,
  `image.tag=$IMAGE_TAG`, and
  `image.pullPolicy=IfNotPresent`. Do not push this local-only tag as the
  release artifact.

## 3. Push

- Push only after the image scan result is acceptable for the target environment:

  ```sh
  make push-docker-image DOCKER_IMAGE=docker.io/kenduest/brobridge-pilotwave1 IMAGE_TAG=$IMAGE_TAG
  ```

  For Podman:

  ```sh
  make push-docker-image CONTAINER_RUNTIME=podman DOCKER_IMAGE=docker.io/kenduest/brobridge-pilotwave1 IMAGE_TAG=$IMAGE_TAG
  ```

- For private registries, create or verify the namespace-local Kubernetes pull secret before deploying. Prefer `imagePullSecret.existingSecret` for shared and production clusters so registry tokens are not stored in Helm release values.

## 3.5. Build User-Facing Helm Artifacts

- Build the chart package and deployment handoff files:

  ```sh
  make dist IMAGE_TAG=$IMAGE_TAG
  ```

- Release or hand off the files under `build/dist/` together with the pushed
  image tag. The directory contains the Helm chart package, production values
  template, installer notes, Helm deployment guide, chart README, and checksums.

## 4. Deploy With Helm

- Production and shared environments should use Helm directly with an
  environment-owned values file. Do not rely on Makefile dev defaults for
  production deployment. Keep detailed install-mode decisions in
  `build/helm/pilotwave/README.md`.

- For shared environments, copy and adjust the production values file, then
  deploy with the target cluster selected by your shell or CI environment:

  ```sh
  cp build/helm/pilotwave/values-production.example.yaml values-prod.yaml
  helm upgrade --install pilotwave build/helm/pilotwave \
    --namespace pilotwave --create-namespace \
    -f values-prod.yaml \
    --set image.repository=docker.io/kenduest/brobridge-pilotwave1 \
    --set image.tag=$IMAGE_TAG \
    --set imagePullSecret.existingSecret=dockerhub-regcred
  ```

- For the local legacy validation cluster, use the dev wrapper. It pins
  `HELM_KUBE_CONTEXT=colima-legacy-1-18`, applies
  `values-legacy-local.example.yaml`, and polls Deployment readiness with
  `kubectl get` to avoid legacy watch/bookmark timeouts:

  ```sh
  make helm-upgrade-dev IMAGE_TAG=$IMAGE_TAG
  ```

- For another local cluster whose runtime already has `pilotwave:$IMAGE_TAG`,
  use the generic local-image wrapper:

  ```sh
  make build-docker-local IMAGE_TAG=$IMAGE_TAG
  make helm-upgrade-local-image IMAGE_TAG=$IMAGE_TAG
  ```

  With Podman:

  ```sh
  make build-docker-local CONTAINER_RUNTIME=podman IMAGE_TAG=$IMAGE_TAG
  make helm-upgrade-local-image IMAGE_TAG=$IMAGE_TAG
  ```

- If Helm succeeds but readiness needs to be rechecked without reapplying the release:

  ```sh
  make helm-wait-ready-dev
  ```

## 5. Smoke And E2E

- Run the local Istio smoke suite against the default validation context:

  ```sh
  ISTIO_CONTEXT=colima-legacy-1-18 make smoke-istio
  ```

- Run the browser/API E2E suite when the local web test prerequisites are available:

  ```sh
  ISTIO_CONTEXT=colima-legacy-1-18 make e2e-web-istio
  ```

- Run the Helm local-image render E2E when Helm is available:

  ```sh
  make e2e-helm-local-image
  ```

  This check covers local `pilotwave:$IMAGE_TAG` rendering, `IfNotPresent`,
  existing production PVC reuse without generating a new PVC,
  ServiceMonitor/PodMonitor output, and the `istio.required=true` failure
  message when required Istio CRDs cannot be discovered.

- Record the image tag, Helm release namespace, values file, and validation commands in the release notes or PR.

## 6. Rollback Basics

- Inspect release history:

  ```sh
  helm history pilotwave --namespace pilotwave
  ```

- Roll back to the selected revision without `--wait`:

  ```sh
  helm rollback pilotwave <revision> --namespace pilotwave
  ```

- Poll readiness with the repository wrapper after rollback:

  ```sh
  make helm-wait-ready-dev
  ```

- Re-run the same smoke or E2E checks used for the release before declaring the rollback complete.
