# Local Legacy Kubernetes Environment

This document records the standalone local Kubernetes environment used to validate Pilotwave against older Kubernetes and Istio APIs. It is adapted from the local lab under `/Users/ken/work/kind`, but the steps here are written so this repository can recreate the environment without depending on that checkout.

## Purpose

Use this environment when validating Pilotwave compatibility with:

- Kubernetes `v1.18.20+k3s1`
- Istio `1.7.5`
- Apple Silicon / `arm64` local development
- Pilotwave Istio smoke and browser E2E tests

The verified local context is:

```text
Colima profile: legacy-pilotware-1-18
kube context:   colima-legacy-1-18
Kubernetes:     v1.18.20+k3s1
Istio:          1.7.5 demo profile
Architecture:   arm64
```

Although the source lab lives in a directory named `kind`, the working legacy path is Colima plus manually installed k3s. Plain `kind` and old `k3d` attempts were tested and were not reliable for Kubernetes 1.18 on this Mac.

## Source Lab Coverage

The `/Users/ken/work/kind` lab had three related tracks. This Pilotwave document keeps the legacy-compatible track as the runnable path and records the others so future setup work does not mix them together.

Modern k3d/k3s baseline:

- Tooling: `k3d`
- Cluster name: `single-k8s`
- kube context: `k3d-single-k8s`
- Kubernetes distribution: k3s
- Image: `rancher/k3s:v1.32.12-k3s1`
- Shape: `1` server and `2` worker agents
- Traefik: disabled with `--disable=traefik`
- Load balancer ports: `localhost:19080 -> 80`, `localhost:19443 -> 443`

That modern baseline is useful for current Kubernetes testing, but it is not the default Pilotwave compatibility target because it does not exercise Kubernetes 1.18 / Istio 1.7 behavior.

Legacy attempts that were tested but not selected:

- `k3d` with `rancher/k3s:v1.18.20-k3s1` on Docker Desktop failed with `failed to find memory cgroup`.
- `kind` with `kindest/node:v1.18.20` on Docker Desktop failed during kubeadm bootstrap because kubelet did not become healthy.
- Colima built-in Kubernetes provisioning for `v1.18.20+k3s1` failed because it tried to download `k3s-airgap-images-arm64.tar.gz`, while the release provides `k3s-airgap-images-arm64.tar`.

Selected legacy path:

- Colima profile: `legacy-pilotware-1-18`
- Built-in Colima Kubernetes: disabled
- Kernel/cgroup adjustment: boot the VM with `systemd.unified_cgroup_hierarchy=0 systemd.legacy_systemd_cgroup_controller=yes` so old k3s uses cgroup v1 instead of cgroup v2
- k3s install: manual install of `v1.18.20+k3s1` inside the VM
- Host kube context: `colima-legacy-1-18`
- Istio install: `1.7.5` demo profile with the `arm64` gateway affinity patch

In short: the kind/k3d/k3s history is preserved here, but the runnable Pilotwave legacy environment is intentionally Colima + manual k3s + cgroup v1.

## Host Requirements

Install these tools on the host:

```bash
colima version
kubectl version --client
helm version
curl --version
```

Optional, but useful for Pilotwave container or chart work:

```bash
docker version
trivy --version
```

The commands below use the normal host kubeconfig at `~/.kube/config`.

## Automation Script

The preferred repo-local entrypoint is:

```bash
scripts/legacy-k8s-env.sh help
```

Useful commands:

```bash
make legacy-env-create
make legacy-env-install-istio
make install-prometheus
make status-prometheus
make legacy-env-status
make legacy-env-status-istio
make legacy-env-verify-pilotwave
```

`legacy-env-create` intentionally stops at the Kubernetes baseline. Install
Istio and Prometheus/Grafana as explicit follow-up steps so failures are easier
to isolate and retry.

Preview without changing the machine:

```bash
scripts/legacy-k8s-env.sh --dry-run create
scripts/legacy-k8s-env.sh --dry-run install-istio
scripts/legacy-k8s-env.sh --dry-run install-prometheus
```

Destructive profile deletion requires an explicit confirmation flag:

```bash
scripts/legacy-k8s-env.sh --yes destroy
```

The manual commands below mirror what the script runs. Keep them as the debugging reference when the automation fails or when adapting the flow to another machine.

## Create the Legacy Cluster

Create a dedicated Colima profile without built-in Kubernetes:

```bash
colima start legacy-pilotware-1-18 \
  --runtime docker \
  --kubernetes=false \
  --cpus 4 \
  --memory 8 \
  --disk 60 \
  --activate=false
```

Patch the VM boot arguments so old k3s can use cgroup v1:

```bash
colima ssh -p legacy-pilotware-1-18 -- sudo sh -lc '
  systemctl stop k3s || true
  if ! grep -q "systemd.unified_cgroup_hierarchy=0" /etc/default/grub; then
    cp /etc/default/grub /etc/default/grub.bak.$(date +%Y%m%d%H%M%S)
    sed -i "s/^GRUB_CMDLINE_LINUX=\"\(.*\)\"/GRUB_CMDLINE_LINUX=\"\1 systemd.unified_cgroup_hierarchy=0 systemd.legacy_systemd_cgroup_controller=yes\"/" /etc/default/grub
    update-grub
  fi
'
```

Restart the profile so the kernel arguments take effect:

```bash
colima stop legacy-pilotware-1-18
colima start legacy-pilotware-1-18 --kubernetes=false --activate=false
```

Install k3s manually inside the VM:

```bash
colima ssh -p legacy-pilotware-1-18 -- sh -lc '
  curl -sfL https://get.k3s.io |
    INSTALL_K3S_VERSION=v1.18.20+k3s1 \
    INSTALL_K3S_EXEC="server --disable traefik --write-kubeconfig-mode 644" \
    sh -
'
```

Traefik is disabled so Istio Gateway tests are not masked by another ingress controller.

## Add the Host Kubeconfig Context

Copy the k3s kubeconfig from Colima, create a profile-specific localhost tunnel,
then create a stable host context named `colima-legacy-1-18`:

```bash
tmp="$(mktemp /private/tmp/colima-legacy-kubeconfig.XXXXXX)"
cafile="$(mktemp /private/tmp/colima-ca.XXXXXX)"
trap 'rm -f "$tmp" "$cafile"' EXIT

ssh -F "$HOME/.colima/_lima/colima-legacy-pilotware-1-18/ssh.config" \
  -f -N \
  -L 127.0.0.1:16443:127.0.0.1:6443 \
  lima-colima-legacy-pilotware-1-18

colima ssh -p legacy-pilotware-1-18 -- sudo cat /etc/rancher/k3s/k3s.yaml > "$tmp"

ca="$(kubectl --kubeconfig "$tmp" config view --raw -o jsonpath='{.clusters[0].cluster.certificate-authority-data}')"
username="$(kubectl --kubeconfig "$tmp" config view --raw -o jsonpath='{.users[0].user.username}')"
password="$(kubectl --kubeconfig "$tmp" config view --raw -o jsonpath='{.users[0].user.password}')"

printf '%s' "$ca" | base64 -d > "$cafile"
cp "$HOME/.kube/config" "$HOME/.kube/config.bak.$(date +%Y%m%d%H%M%S)"

kubectl config set-cluster colima-legacy-1-18 \
  --server="https://127.0.0.1:16443" \
  --certificate-authority="$cafile" \
  --embed-certs=true
kubectl config set-credentials colima-legacy-1-18 \
  --username="$username" \
  --password="$password"
kubectl config set-context colima-legacy-1-18 \
  --cluster=colima-legacy-1-18 \
  --user=colima-legacy-1-18
```

Do not reuse `127.0.0.1:6443` when multiple Colima profiles are running. A
different profile can already own that SSH tunnel, causing TLS
`x509: certificate signed by unknown authority` errors because kubectl talks to
the wrong API server.

Switch context when desired:

```bash
kubectl config use-context colima-legacy-1-18
```

## Verify Kubernetes

Host-side checks:

```bash
kubectl cluster-info --context colima-legacy-1-18
kubectl get nodes --context colima-legacy-1-18 -o wide
kubectl get pods -A --context colima-legacy-1-18
```

Expected baseline:

```text
node: colima-legacy-pilotware-1-18 Ready master v1.18.20+k3s1
pods: coredns, metrics-server, local-path-provisioner Running
```

VM-side checks:

```bash
colima ssh -p legacy-pilotware-1-18 -- sudo systemctl status k3s
colima ssh -p legacy-pilotware-1-18 -- sudo k3s kubectl get nodes -o wide
```

## Install Istio 1.7.5

Download the `linux-arm64` Istio archive inside Colima:

```bash
colima ssh -p legacy-pilotware-1-18 -- sh -lc '
  test -x /tmp/istio-1.7.5/bin/istioctl || (
    rm -rf /tmp/istio-1.7.5 /tmp/istio-1.7.5-linux-arm64.tar.gz
    curl -L https://github.com/istio/istio/releases/download/1.7.5/istio-1.7.5-linux-arm64.tar.gz \
      -o /tmp/istio-1.7.5-linux-arm64.tar.gz
    tar -xzf /tmp/istio-1.7.5-linux-arm64.tar.gz -C /tmp
  )
  /tmp/istio-1.7.5/bin/istioctl version --remote=false
'
```

Generate the demo profile manifest with the required `arm64` gateway affinity
patch, then apply it in phases:

```bash
colima ssh -p legacy-pilotware-1-18 -- sh -lc '
  KUBECONFIG=/etc/rancher/k3s/k3s.yaml \
    /tmp/istio-1.7.5/bin/istioctl manifest generate --set profile=demo |
    sed "/- s390x/a\\                - arm64" > /tmp/istio-1.7.5-demo-arm64.yaml

  sudo k3s kubectl create namespace istio-system --dry-run=client -o yaml |
    sudo k3s kubectl apply -f -

  awk '"'"'BEGIN{doc=""} /^---/{if(doc ~ /kind: CustomResourceDefinition/) print doc "\n---"; doc=""; next} {doc=doc $0 "\n"} END{if(doc ~ /kind: CustomResourceDefinition/) print doc}'"'"' \
    /tmp/istio-1.7.5-demo-arm64.yaml > /tmp/istio-1.7.5-crds.yaml

  sudo k3s kubectl apply -f /tmp/istio-1.7.5-crds.yaml
  sudo k3s kubectl wait --for=condition=Established crd --all --timeout=120s
  sudo k3s kubectl apply -f /tmp/istio-1.7.5-demo-arm64.yaml
'
```

The patch is required because Istio `1.7.5` gateway deployments include required node affinity for `amd64`, `ppc64le`, and `s390x`, but not `arm64`. Without the patch, `istio-ingressgateway` and `istio-egressgateway` stay Pending on Apple Silicon Colima.

The phased apply is also required on this legacy cluster. Applying the full
manifest in one `kubectl apply -f` can race namespace creation and CRD discovery,
producing errors such as missing `istio-system` or no matches for
`EnvoyFilter`.

Poll Istio readiness from the host:

```bash
for deploy in istiod istio-ingressgateway istio-egressgateway; do
  for i in $(seq 1 30); do
    ready="$(kubectl get deploy "$deploy" -n istio-system --context colima-legacy-1-18 -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
    desired="$(kubectl get deploy "$deploy" -n istio-system --context colima-legacy-1-18 -o jsonpath='{.spec.replicas}' 2>/dev/null || true)"
    if [ -n "$desired" ] && [ "$ready" = "$desired" ]; then
      echo "$deploy available ($ready/$desired)"
      break
    fi
    if [ "$i" = "30" ]; then
      echo "$deploy not available after 5 minutes (ready=${ready:-0}, desired=${desired:-unknown})" >&2
      exit 1
    fi
    sleep 10
  done
done
```

## Verify Istio

Use these checks instead of `istioctl verify-install`:

```bash
kubectl get pods -n istio-system --context colima-legacy-1-18 -o wide
kubectl get deploy,svc -n istio-system --context colima-legacy-1-18
colima ssh -p legacy-pilotware-1-18 -- sh -lc '
  KUBECONFIG=/etc/rancher/k3s/k3s.yaml /tmp/istio-1.7.5/bin/istioctl version
'
```

Expected Istio state:

```text
deployment.apps/istiod                 1/1
deployment.apps/istio-ingressgateway   1/1
deployment.apps/istio-egressgateway    1/1
client version: 1.7.5
control plane version: 1.7.5
data plane version: 1.7.5
```

Confirm the gateway affinity includes `arm64`:

```bash
kubectl get deploy istio-ingressgateway \
  -n istio-system \
  --context colima-legacy-1-18 \
  -o jsonpath='{.spec.template.spec.affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].values}{"\n"}'
```

`istioctl verify-install` from Istio `1.7.5` can fail in this setup with an installed-state unmarshal error even when workloads are healthy. Also avoid relying on `kubectl rollout status`; watch/bookmark behavior can hang against this old Kubernetes version. Prefer the polling and status checks above.

## Run Pilotwave Against the Cluster

Build assets and run the backend with a flattened kubeconfig for the default
legacy context:

```bash
make run-server-cluster
```

Run the Pilotwave Istio validation suites:

```bash
make smoke-istio
make e2e-web-istio
```

These are the default local compatibility gates for the legacy Istio target.

## Deploy Pilotwave with Helm

General Helm install modes and values are documented in
[`../build/helm/pilotwave/README.md`](../build/helm/pilotwave/README.md). This
section only records the legacy local-cluster differences.

Do not use `helm --wait` on this Kubernetes 1.18 cluster. The Deployment can be
ready while newer Helm/client-go waits forever for watch bookmark behavior that
the old API server does not satisfy reliably.

If you want to deploy a local image instead of pulling from a registry, build it
into the Docker engine used by the legacy Colima profile. This is separate from
the Docker Desktop/release Buildx path:

```bash
eval "$(colima docker-env legacy-pilotware-1-18)"
IMAGE_TAG=$(make -s version | awk -F= '/IMAGE_TAG/ {print $2}')
make build-docker-legacy-local IMAGE_TAG=$IMAGE_TAG
```

Use the repository wrapper instead. It runs Helm without `--wait`, then polls
Deployment readiness with `kubectl get`:

```bash
make helm-upgrade-dev IMAGE_TAG=$IMAGE_TAG
```

For monitoring integration tests, include the Prometheus Operator resources:

```bash
make helm-upgrade-dev \
  HELM_DEV_EXTRA_ARGS="--set serviceMonitor.enabled=true --set serviceMonitor.namespace=monitoring --set serviceMonitor.honorLabels=true --set serviceMonitor.labels.release=legacy-monitoring --set istioPodMonitor.enabled=true --set istioPodMonitor.namespace=monitoring --set istioPodMonitor.labels.release=legacy-monitoring --set grafanaDashboard.enabled=true --set grafanaDashboard.namespace=monitoring --set prometheusRule.enabled=true --set prometheusRule.namespace=monitoring --set prometheusRule.labels.app=kube-prometheus-stack --set prometheusRule.labels.release=legacy-monitoring --set grafana.provider=prometheus --set grafana.host=legacy-monitoring-kube-pro-prometheus.monitoring.svc --set grafana.port=9090"
```

Check readiness without watch:

```bash
make helm-wait-ready-dev
```

Expected output:

```text
deployment/pilotwave ready (desired=1 updated=1 ready=1 available=1 ...)
```

## Optional Prometheus Operator Stack

Pilotwave exposes metrics at `/metrics`. For Prometheus/Grafana testing on the same legacy cluster, use a Kubernetes 1.18-compatible chart instead of the latest `kube-prometheus-stack`.

Pinned local-lab choice:

```text
Helm chart: prometheus-community/kube-prometheus-stack
Chart:      10.0.1
Operator:   0.42.1
Namespace:  monitoring
Release:    legacy-monitoring
```

Install path:

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts || true
helm repo update prometheus-community

helm show crds prometheus-community/kube-prometheus-stack --version 10.0.1 |
  kubectl apply --context colima-legacy-1-18 -f -

kubectl create namespace monitoring --context colima-legacy-1-18 --dry-run=client -o yaml |
  kubectl apply --context colima-legacy-1-18 -f -

helm template legacy-monitoring prometheus-community/kube-prometheus-stack \
  --version 10.0.1 \
  --namespace monitoring \
  --skip-crds \
  --kube-version 1.18.20 \
  --set kubeTargetVersionOverride=1.18.20 \
  --set kubeControllerManager.enabled=false \
  --set kubeScheduler.enabled=false \
  --set kubeProxy.enabled=false \
  --set kubeEtcd.enabled=false \
  --set prometheusOperator.admissionWebhooks.enabled=false \
  --set prometheusOperator.tlsProxy.enabled=false \
  --set grafana.testFramework.enabled=false \
  --set kube-state-metrics.image.repository=registry.k8s.io/kube-state-metrics/kube-state-metrics \
  --set kube-state-metrics.image.tag=v1.9.8 |
  kubectl apply --context colima-legacy-1-18 -f -

kubectl delete pod,configmap,serviceaccount legacy-monitoring-grafana-test \
  -n monitoring \
  --context colima-legacy-1-18 \
  --ignore-not-found=true
```

Verify:

```bash
kubectl get pods,svc,deploy,sts -n monitoring --context colima-legacy-1-18
kubectl get prometheus,alertmanager,servicemonitor,podmonitor -n monitoring --context colima-legacy-1-18
kubectl top pods -A --context colima-legacy-1-18
```

Grafana default credentials from this chart are `admin:prom-operator`.

The Grafana chart normally renders a `legacy-monitoring-grafana-test` Helm test
hook. Because this runbook uses `helm template | kubectl apply` for Kubernetes
1.18 compatibility, that hook would be created as an ordinary Pod and can remain
in `Error` even when Grafana and Prometheus are healthy. The automation disables
`grafana.testFramework` and deletes any old hook remnants.

## Start, Stop, and Cleanup

Start an existing profile:

```bash
colima start legacy-pilotware-1-18 --kubernetes=false --activate=false
colima ssh -p legacy-pilotware-1-18 -- sudo systemctl start k3s
```

Stop without deleting:

```bash
colima ssh -p legacy-pilotware-1-18 -- sudo systemctl stop k3s || true
colima stop legacy-pilotware-1-18
```

Remove Istio:

```bash
colima ssh -p legacy-pilotware-1-18 -- sh -lc '
  sudo k3s kubectl delete -f /tmp/istio-1.7.5-demo-arm64.yaml
'
```

Delete the whole profile:

```bash
colima delete legacy-pilotware-1-18
```

## Known Failed Paths

These were already tested in the source local lab and should not be retried as the first option for this legacy target:

- `k3d` with `rancher/k3s:v1.18.20-k3s1` on Docker Desktop failed with `failed to find memory cgroup`.
- `kind` with `kindest/node:v1.18.20` on Docker Desktop failed during kubeadm bootstrap because kubelet did not become healthy.
- Colima built-in Kubernetes provisioning for `v1.18.20+k3s1` failed because it tried to download `k3s-airgap-images-arm64.tar.gz`, while the release provides `k3s-airgap-images-arm64.tar`.

The durable local path is the Colima profile with cgroup v1 kernel args and manual k3s install.
