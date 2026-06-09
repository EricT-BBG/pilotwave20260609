#!/usr/bin/env bash
set -euo pipefail

SCRIPT_NAME="$(basename "$0")"

DRY_RUN=0
ASSUME_YES=0

LEGACY_CLUSTER_NAME="${LEGACY_CLUSTER_NAME:-legacy-1-18}"
LEGACY_CONTEXT="${LEGACY_CONTEXT:-colima-${LEGACY_CLUSTER_NAME}}"
LEGACY_API_HOST="${LEGACY_API_HOST:-127.0.0.1}"
LEGACY_API_PORT="${LEGACY_API_PORT:-16443}"
LEGACY_K3S_VERSION="${LEGACY_K3S_VERSION:-v1.18.20+k3s1}"
ISTIO_VERSION="${ISTIO_VERSION:-1.7.5}"
ISTIO_ARCHIVE="${ISTIO_ARCHIVE:-istio-${ISTIO_VERSION}-linux-arm64.tar.gz}"
ISTIO_DIR="${ISTIO_DIR:-/tmp/istio-${ISTIO_VERSION}}"
ISTIO_MANIFEST="${ISTIO_MANIFEST:-/tmp/istio-${ISTIO_VERSION}-demo-arm64.yaml}"
ISTIO_CRD_MANIFEST="${ISTIO_CRD_MANIFEST:-/tmp/istio-${ISTIO_VERSION}-crds.yaml}"
PROMETHEUS_NAMESPACE="${PROMETHEUS_NAMESPACE:-monitoring}"
PROMETHEUS_RELEASE="${PROMETHEUS_RELEASE:-legacy-monitoring}"
PROMETHEUS_STACK_CHART="${PROMETHEUS_STACK_CHART:-prometheus-community/kube-prometheus-stack}"
PROMETHEUS_STACK_CHART_VERSION="${PROMETHEUS_STACK_CHART_VERSION:-10.0.1}"

usage() {
  cat <<EOF
Usage: ${SCRIPT_NAME} [--dry-run] [--yes] <command>

Commands:
  create              Create Colima profile, force cgroup v1, install k3s, refresh kubeconfig
  start               Start the Colima profile and k3s service
  stop                Stop k3s and the Colima profile
  destroy             Delete the Colima profile; requires --yes
  refresh-kubeconfig  Recreate host kube context ${LEGACY_CONTEXT}
  status              Show Kubernetes node status
  verify              Show nodes and all pods
  install-istio       Install Istio ${ISTIO_VERSION} demo profile with arm64 gateway patch
  wait-istio          Poll Istio deployments until available
  status-istio        Show Istio pods, services, and version
  install-prometheus  Install legacy kube-prometheus-stack ${PROMETHEUS_STACK_CHART_VERSION}
  status-prometheus   Show Prometheus stack status
  verify-pilotwave    Run Pilotwave Istio smoke and web E2E tests against ${LEGACY_CONTEXT}
  help                Show this help

Options:
  --dry-run           Print commands without executing them
  --yes               Confirm destructive commands such as destroy

Environment overrides:
  LEGACY_CLUSTER_NAME=${LEGACY_CLUSTER_NAME}
  LEGACY_CONTEXT=${LEGACY_CONTEXT}
  LEGACY_API_HOST=${LEGACY_API_HOST}
  LEGACY_API_PORT=${LEGACY_API_PORT}
  LEGACY_K3S_VERSION=${LEGACY_K3S_VERSION}
  ISTIO_VERSION=${ISTIO_VERSION}
EOF
}

log() {
  printf '%s\n' "$*"
}

run() {
  if [[ "$DRY_RUN" -eq 1 ]]; then
    printf '[dry-run] %s\n' "$*"
    return 0
  fi

  "$@"
}

run_shell() {
  if [[ "$DRY_RUN" -eq 1 ]]; then
    printf '[dry-run] %s\n' "$*"
    return 0
  fi

  "$@"
}

legacy_lima_name() {
  printf 'colima-%s' "$LEGACY_CLUSTER_NAME"
}

legacy_lima_ssh_config() {
  printf '%s/.colima/_lima/%s/ssh.config' "$HOME" "$(legacy_lima_name)"
}

legacy_lima_ssh_host() {
  printf 'lima-%s' "$(legacy_lima_name)"
}

legacy_api_server() {
  printf 'https://%s:%s' "$LEGACY_API_HOST" "$LEGACY_API_PORT"
}

ensure_api_tunnel() {
  local ssh_config
  local ssh_host
  ssh_config="$(legacy_lima_ssh_config)"
  ssh_host="$(legacy_lima_ssh_host)"

  if [[ "$DRY_RUN" -eq 1 ]]; then
    log "[dry-run] ssh -F ${ssh_config} -f -N -L ${LEGACY_API_HOST}:${LEGACY_API_PORT}:127.0.0.1:6443 ${ssh_host}"
    return 0
  fi

  if command -v lsof >/dev/null 2>&1 && lsof -n -P -iTCP:"$LEGACY_API_PORT" -sTCP:LISTEN >/dev/null 2>&1; then
    return 0
  fi

  ssh -F "$ssh_config" -f -N -L "${LEGACY_API_HOST}:${LEGACY_API_PORT}:127.0.0.1:6443" "$ssh_host"
}

need_confirm_destroy() {
  if [[ "$ASSUME_YES" -ne 1 ]]; then
    echo "destroy requires --yes" >&2
    exit 2
  fi
}

create_cluster() {
  run colima start "$LEGACY_CLUSTER_NAME" \
    --runtime docker \
    --kubernetes=false \
    --cpus 4 \
    --memory 8 \
    --disk 60 \
    --activate=false

  run_shell colima ssh -p "$LEGACY_CLUSTER_NAME" -- sudo sh -lc \
    'systemctl stop k3s || true; if ! grep -q "systemd.unified_cgroup_hierarchy=0" /etc/default/grub; then cp /etc/default/grub /etc/default/grub.bak.$(date +%Y%m%d%H%M%S); sed -i "s/^GRUB_CMDLINE_LINUX=\"\(.*\)\"/GRUB_CMDLINE_LINUX=\"\1 systemd.unified_cgroup_hierarchy=0 systemd.legacy_systemd_cgroup_controller=yes\"/" /etc/default/grub; update-grub; fi'

  run colima stop "$LEGACY_CLUSTER_NAME"
  run colima start "$LEGACY_CLUSTER_NAME" --kubernetes=false --activate=false

  run_shell colima ssh -p "$LEGACY_CLUSTER_NAME" -- sh -lc \
    "curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION=${LEGACY_K3S_VERSION} INSTALL_K3S_EXEC=\"server --disable traefik --write-kubeconfig-mode 644\" sh -"

  refresh_kubeconfig
}

start_cluster() {
  run colima start "$LEGACY_CLUSTER_NAME" --kubernetes=false --activate=false
  run colima ssh -p "$LEGACY_CLUSTER_NAME" -- sudo systemctl start k3s
}

stop_cluster() {
  if [[ "$DRY_RUN" -eq 1 ]]; then
    log "[dry-run] colima ssh -p ${LEGACY_CLUSTER_NAME} -- sudo systemctl stop k3s || true"
  else
    colima ssh -p "$LEGACY_CLUSTER_NAME" -- sudo systemctl stop k3s || true
  fi
  run colima stop "$LEGACY_CLUSTER_NAME"
}

destroy_cluster() {
  need_confirm_destroy
  run colima delete "$LEGACY_CLUSTER_NAME"
}

refresh_kubeconfig() {
  if [[ "$DRY_RUN" -eq 1 ]]; then
    log "[dry-run] tmp=\"\$(mktemp /private/tmp/colima-legacy-kubeconfig.XXXXXX)\""
    log "[dry-run] cafile=\"\$(mktemp /private/tmp/colima-ca.XXXXXX)\""
    ensure_api_tunnel
    log "[dry-run] colima ssh -p ${LEGACY_CLUSTER_NAME} -- sudo cat /etc/rancher/k3s/k3s.yaml > \"\$tmp\""
    log "[dry-run] kubectl config set-cluster ${LEGACY_CONTEXT} --server=\"$(legacy_api_server)\" --certificate-authority=\"\$cafile\" --embed-certs=true"
    log "[dry-run] kubectl config set-credentials ${LEGACY_CONTEXT} --username=\"\$username\" --password=\"\$password\""
    log "[dry-run] kubectl config set-context ${LEGACY_CONTEXT} --cluster=${LEGACY_CONTEXT} --user=${LEGACY_CONTEXT}"
    return 0
  fi

  ensure_api_tunnel

  local tmp
  local cafile
  tmp="$(mktemp /private/tmp/colima-legacy-kubeconfig.XXXXXX)"
  cafile="$(mktemp /private/tmp/colima-ca.XXXXXX)"

  colima ssh -p "$LEGACY_CLUSTER_NAME" -- sudo cat /etc/rancher/k3s/k3s.yaml > "$tmp"

  local ca
  local username
  local password
  ca="$(kubectl --kubeconfig "$tmp" config view --raw -o jsonpath='{.clusters[0].cluster.certificate-authority-data}')"
  username="$(kubectl --kubeconfig "$tmp" config view --raw -o jsonpath='{.users[0].user.username}')"
  password="$(kubectl --kubeconfig "$tmp" config view --raw -o jsonpath='{.users[0].user.password}')"

  printf '%s' "$ca" | base64 -d > "$cafile"
  cp "$HOME/.kube/config" "$HOME/.kube/config.bak.$(date +%Y%m%d%H%M%S)"
  kubectl config set-cluster "$LEGACY_CONTEXT" --server="$(legacy_api_server)" --certificate-authority="$cafile" --embed-certs=true >/dev/null
  kubectl config set-credentials "$LEGACY_CONTEXT" --username="$username" --password="$password" >/dev/null
  kubectl config set-context "$LEGACY_CONTEXT" --cluster="$LEGACY_CONTEXT" --user="$LEGACY_CONTEXT" >/dev/null
  rm -f "$tmp" "$cafile"
}

status_cluster() {
  run kubectl cluster-info --context "$LEGACY_CONTEXT"
  run kubectl get nodes --context "$LEGACY_CONTEXT" -o wide
}

verify_cluster() {
  run kubectl get nodes --context "$LEGACY_CONTEXT"
  run kubectl get pods -A --context "$LEGACY_CONTEXT"
}

download_istio() {
  run_shell colima ssh -p "$LEGACY_CLUSTER_NAME" -- sh -lc \
    "test -x ${ISTIO_DIR}/bin/istioctl || (rm -rf ${ISTIO_DIR} /tmp/${ISTIO_ARCHIVE} && curl -L https://github.com/istio/istio/releases/download/${ISTIO_VERSION}/${ISTIO_ARCHIVE} -o /tmp/${ISTIO_ARCHIVE} && tar -xzf /tmp/${ISTIO_ARCHIVE} -C /tmp)"
}

install_istio() {
  download_istio

  if [[ "$DRY_RUN" -eq 1 ]]; then
    log "[dry-run] colima ssh -p ${LEGACY_CLUSTER_NAME} -- sh -lc KUBECONFIG=/etc/rancher/k3s/k3s.yaml ${ISTIO_DIR}/bin/istioctl manifest generate --set profile=demo | sed \"/- s390x/a\\\\                - arm64\" > ${ISTIO_MANIFEST}"
    log "[dry-run] colima ssh -p ${LEGACY_CLUSTER_NAME} -- sh -lc sudo k3s kubectl create namespace istio-system --dry-run=client -o yaml | sudo k3s kubectl apply -f -"
    log "[dry-run] colima ssh -p ${LEGACY_CLUSTER_NAME} -- sh -lc awk ... ${ISTIO_MANIFEST} > ${ISTIO_CRD_MANIFEST}"
    log "[dry-run] colima ssh -p ${LEGACY_CLUSTER_NAME} -- sh -lc sudo k3s kubectl apply -f ${ISTIO_CRD_MANIFEST}"
    log "[dry-run] colima ssh -p ${LEGACY_CLUSTER_NAME} -- sh -lc sudo k3s kubectl wait --for=condition=Established crd --all --timeout=120s"
    log "[dry-run] colima ssh -p ${LEGACY_CLUSTER_NAME} -- sh -lc sudo k3s kubectl apply -f ${ISTIO_MANIFEST}"
    wait_istio
    return 0
  fi

  run_shell colima ssh -p "$LEGACY_CLUSTER_NAME" -- sh -lc \
    "KUBECONFIG=/etc/rancher/k3s/k3s.yaml ${ISTIO_DIR}/bin/istioctl manifest generate --set profile=demo | sed \"/- s390x/a\\\\                - arm64\" > ${ISTIO_MANIFEST}"
  run_shell colima ssh -p "$LEGACY_CLUSTER_NAME" -- sh -lc \
    "sudo k3s kubectl create namespace istio-system --dry-run=client -o yaml | sudo k3s kubectl apply -f -"
  run_shell colima ssh -p "$LEGACY_CLUSTER_NAME" -- sh -lc \
    "awk 'BEGIN{doc=\"\"} /^---/{if(doc ~ /kind: CustomResourceDefinition/) print doc \"\\n---\"; doc=\"\"; next} {doc=doc \$0 \"\\n\"} END{if(doc ~ /kind: CustomResourceDefinition/) print doc}' ${ISTIO_MANIFEST} > ${ISTIO_CRD_MANIFEST}"
  run_shell colima ssh -p "$LEGACY_CLUSTER_NAME" -- sh -lc \
    "sudo k3s kubectl apply -f ${ISTIO_CRD_MANIFEST}"
  run_shell colima ssh -p "$LEGACY_CLUSTER_NAME" -- sh -lc \
    "sudo k3s kubectl wait --for=condition=Established crd --all --timeout=120s"
  run_shell colima ssh -p "$LEGACY_CLUSTER_NAME" -- sh -lc \
    "sudo k3s kubectl apply -f ${ISTIO_MANIFEST}"
  wait_istio
}

wait_istio() {
  if [[ "$DRY_RUN" -eq 1 ]]; then
    cat <<EOF
[dry-run] for deploy in istiod istio-ingressgateway istio-egressgateway; do
[dry-run]   ready="\$(kubectl get deploy "\$deploy" -n istio-system --context ${LEGACY_CONTEXT} -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
[dry-run]   desired="\$(kubectl get deploy "\$deploy" -n istio-system --context ${LEGACY_CONTEXT} -o jsonpath='{.spec.replicas}' 2>/dev/null || true)"
[dry-run] done
EOF
    return 0
  fi

  local deploy
  local i
  local ready
  local desired

  for deploy in istiod istio-ingressgateway istio-egressgateway; do
    for i in $(seq 1 30); do
      ready="$(kubectl get deploy "$deploy" -n istio-system --context "$LEGACY_CONTEXT" -o jsonpath='{.status.readyReplicas}' 2>/dev/null || true)"
      desired="$(kubectl get deploy "$deploy" -n istio-system --context "$LEGACY_CONTEXT" -o jsonpath='{.spec.replicas}' 2>/dev/null || true)"
      if [[ -n "$desired" && "$ready" == "$desired" ]]; then
        echo "$deploy available ($ready/$desired)"
        break
      fi
      if [[ "$i" == "30" ]]; then
        echo "$deploy not available after 5 minutes (ready=${ready:-0}, desired=${desired:-unknown})" >&2
        exit 1
      fi
      sleep 10
    done
  done
}

status_istio() {
  run kubectl get pods -n istio-system --context "$LEGACY_CONTEXT" -o wide
  run kubectl get deploy,svc -n istio-system --context "$LEGACY_CONTEXT"
  run_shell colima ssh -p "$LEGACY_CLUSTER_NAME" -- sh -lc \
    "KUBECONFIG=/etc/rancher/k3s/k3s.yaml ${ISTIO_DIR}/bin/istioctl version"
}

install_prometheus() {
  if [[ "$DRY_RUN" -eq 1 ]]; then
    log "[dry-run] helm repo add prometheus-community https://prometheus-community.github.io/helm-charts || true"
  else
    helm repo add prometheus-community https://prometheus-community.github.io/helm-charts || true
  fi
  run helm repo update prometheus-community

  if [[ "$DRY_RUN" -eq 1 ]]; then
    log "[dry-run] helm show crds ${PROMETHEUS_STACK_CHART} --version ${PROMETHEUS_STACK_CHART_VERSION} | kubectl apply --context ${LEGACY_CONTEXT} -f -"
    log "[dry-run] kubectl create namespace ${PROMETHEUS_NAMESPACE} --context ${LEGACY_CONTEXT} --dry-run=client -o yaml | kubectl apply --context ${LEGACY_CONTEXT} -f -"
    log "[dry-run] helm template ${PROMETHEUS_RELEASE} ${PROMETHEUS_STACK_CHART} --version ${PROMETHEUS_STACK_CHART_VERSION} --namespace ${PROMETHEUS_NAMESPACE} --skip-crds --kube-version 1.18.20 --set grafana.testFramework.enabled=false ... | kubectl apply --context ${LEGACY_CONTEXT} -f -"
    log "[dry-run] kubectl delete pod,configmap,serviceaccount ${PROMETHEUS_RELEASE}-grafana-test -n ${PROMETHEUS_NAMESPACE} --context ${LEGACY_CONTEXT} --ignore-not-found=true"
    return 0
  fi

  helm show crds "$PROMETHEUS_STACK_CHART" --version "$PROMETHEUS_STACK_CHART_VERSION" |
    kubectl apply --context "$LEGACY_CONTEXT" -f -

  kubectl create namespace "$PROMETHEUS_NAMESPACE" --context "$LEGACY_CONTEXT" --dry-run=client -o yaml |
    kubectl apply --context "$LEGACY_CONTEXT" -f -

  helm template "$PROMETHEUS_RELEASE" "$PROMETHEUS_STACK_CHART" \
    --version "$PROMETHEUS_STACK_CHART_VERSION" \
    --namespace "$PROMETHEUS_NAMESPACE" \
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
    kubectl apply --context "$LEGACY_CONTEXT" -f -

  kubectl delete pod,configmap,serviceaccount "${PROMETHEUS_RELEASE}-grafana-test" \
    -n "$PROMETHEUS_NAMESPACE" \
    --context "$LEGACY_CONTEXT" \
    --ignore-not-found=true
}

status_prometheus() {
  run kubectl get pods,svc,deploy,sts -n "$PROMETHEUS_NAMESPACE" --context "$LEGACY_CONTEXT" -o wide
  run kubectl get prometheus,alertmanager,servicemonitor,podmonitor -n "$PROMETHEUS_NAMESPACE" --context "$LEGACY_CONTEXT"
  run kubectl top pods -A --context "$LEGACY_CONTEXT"
}

verify_pilotwave() {
  if [[ "$DRY_RUN" -eq 1 ]]; then
    log "[dry-run] ISTIO_CONTEXT=${LEGACY_CONTEXT} make smoke-istio"
    log "[dry-run] ISTIO_CONTEXT=${LEGACY_CONTEXT} make e2e-web-istio"
    return 0
  fi

  ISTIO_CONTEXT="$LEGACY_CONTEXT" make smoke-istio
  ISTIO_CONTEXT="$LEGACY_CONTEXT" make e2e-web-istio
}

if [[ "$#" -eq 0 ]]; then
  usage
  exit 0
fi

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --dry-run)
      DRY_RUN=1
      shift
      ;;
    --yes|-y)
      ASSUME_YES=1
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      break
      ;;
  esac
done

command="${1:-help}"

case "$command" in
  create) create_cluster ;;
  start) start_cluster ;;
  stop) stop_cluster ;;
  destroy) destroy_cluster ;;
  refresh-kubeconfig) refresh_kubeconfig ;;
  status) status_cluster ;;
  verify) verify_cluster ;;
  install-istio) install_istio ;;
  wait-istio) wait_istio ;;
  status-istio) status_istio ;;
  install-prometheus) install_prometheus ;;
  status-prometheus) status_prometheus ;;
  verify-pilotwave) verify_pilotwave ;;
  help) usage ;;
  *)
    echo "unknown command: $command" >&2
    usage >&2
    exit 2
    ;;
esac
