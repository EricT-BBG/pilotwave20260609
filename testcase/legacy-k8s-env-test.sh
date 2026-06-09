#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPT="$ROOT_DIR/scripts/legacy-k8s-env.sh"

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"

  if [[ "$haystack" != *"$needle"* ]]; then
    fail "expected output to contain: $needle"
  fi
}

assert_not_contains() {
  local haystack="$1"
  local needle="$2"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "expected output not to contain: $needle"
  fi
}

help_output="$(bash "$SCRIPT" help)"
assert_contains "$help_output" "Usage: legacy-k8s-env.sh"
assert_contains "$help_output" "create"
assert_contains "$help_output" "refresh-kubeconfig"
assert_contains "$help_output" "install-istio"
assert_contains "$help_output" "verify-pilotwave"

dry_create="$(bash "$SCRIPT" --dry-run create)"
assert_contains "$dry_create" "[dry-run] colima start legacy-1-18"
assert_contains "$dry_create" "systemd.unified_cgroup_hierarchy=0"
assert_contains "$dry_create" "INSTALL_K3S_VERSION=v1.18.20+k3s1"
assert_contains "$dry_create" "kubectl config set-context colima-legacy-1-18"
assert_contains "$dry_create" "ssh -F"
assert_contains "$dry_create" "colima-legacy-1-18/ssh.config"
assert_contains "$dry_create" "-L 127.0.0.1:16443:127.0.0.1:6443"
assert_contains "$dry_create" "kubectl config set-cluster colima-legacy-1-18 --server=\"https://127.0.0.1:16443\""
if grep -q 'trap .*"\$tmp"' "$SCRIPT"; then
  fail "refresh_kubeconfig must not leave a RETURN trap that references local tmp/cafile under set -u"
fi

dry_istio="$(bash "$SCRIPT" --dry-run install-istio)"
assert_contains "$dry_istio" "istio-1.7.5-linux-arm64.tar.gz"
assert_contains "$dry_istio" "istioctl manifest generate --set profile=demo"
assert_contains "$dry_istio" "sed \"/- s390x/a\\\\                - arm64\""
assert_contains "$dry_istio" "kubectl create namespace istio-system"
assert_contains "$dry_istio" "kubectl apply -f /tmp/istio-1.7.5-crds.yaml"
assert_contains "$dry_istio" "kubectl wait --for=condition=Established crd --all"
assert_contains "$dry_istio" "kubectl apply -f /tmp/istio-1.7.5-demo-arm64.yaml"
assert_contains "$dry_istio" "kubectl get deploy \"\$deploy\" -n istio-system --context colima-legacy-1-18"

dry_prometheus="$(bash "$SCRIPT" --dry-run install-prometheus)"
assert_contains "$dry_prometheus" "grafana.testFramework.enabled=false"
assert_contains "$dry_prometheus" "kubectl delete pod,configmap,serviceaccount legacy-monitoring-grafana-test"

dry_stop="$(bash "$SCRIPT" --dry-run stop)"
assert_contains "$dry_stop" "[dry-run] colima ssh -p legacy-1-18 -- sudo systemctl stop k3s || true"
assert_contains "$dry_stop" "[dry-run] colima stop legacy-1-18"

set +e
destroy_output="$(bash "$SCRIPT" --dry-run destroy 2>&1)"
destroy_status=$?
set -e
if [[ "$destroy_status" -eq 0 ]]; then
  fail "destroy without --yes should fail"
fi
assert_contains "$destroy_output" "requires --yes"
assert_not_contains "$destroy_output" "colima delete"

dry_destroy="$(bash "$SCRIPT" --dry-run --yes destroy)"
assert_contains "$dry_destroy" "[dry-run] colima delete legacy-1-18"

echo "legacy-k8s-env tests passed"
