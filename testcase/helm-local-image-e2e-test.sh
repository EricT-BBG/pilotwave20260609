#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CHART_DIR="${ROOT_DIR}/build/helm/pilotwave"
OUT_FILE="$(mktemp /tmp/pilotwave-helm-e2e.XXXXXX.yaml)"
trap 'rm -f "$OUT_FILE"' EXIT

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

assert_contains() {
  local needle="$1"

  if ! grep -Fq "$needle" "$OUT_FILE"; then
    fail "expected rendered Helm output to contain: $needle"
  fi
}

assert_not_contains() {
  local needle="$1"

  if grep -Fq "$needle" "$OUT_FILE"; then
    fail "expected rendered Helm output not to contain: $needle"
  fi
}

if ! command -v helm >/dev/null 2>&1; then
  echo "SKIP: helm is required for helm-local-image E2E render checks"
  exit 0
fi

set +e
required_output="$(
  helm template pilotwave "$CHART_DIR" \
    --namespace pilotwave \
    --set istio.required=true \
    2>&1 >/dev/null
)"
required_status=$?
set -e

if [[ "$required_status" -eq 0 ]]; then
  fail "istio.required=true should fail without live Istio CRD lookup"
fi

if [[ "$required_output" != *"istio.required=true but required Istio CRDs are not available"* ]]; then
  fail "expected istio.required=true failure to explain missing Istio CRDs"
fi

helm template pilotwave "$CHART_DIR" \
  --namespace pilotwave \
  --set image.repository=pilotwave \
  --set image.tag=e2e-local \
  --set image.pullPolicy=IfNotPresent \
  --set persistence.enabled=true \
  --set persistence.existingClaim=pilotwave-prod-data \
  --set istio.required=false \
  --set serviceMonitor.enabled=true \
  --set serviceMonitor.namespace=monitoring \
  --set serviceMonitor.labels.release=legacy-monitoring \
  --set istioPodMonitor.enabled=true \
  --set istioPodMonitor.namespace=monitoring \
  --set istioPodMonitor.labels.release=legacy-monitoring \
  > "$OUT_FILE"

assert_contains "image: \"pilotwave:e2e-local\""
assert_contains 'imagePullPolicy: "IfNotPresent"'
assert_contains "claimName: pilotwave-prod-data"
assert_contains "kind: ServiceMonitor"
assert_contains "kind: PodMonitor"
assert_not_contains "kind: PersistentVolumeClaim"

echo "helm local image E2E render checks passed"
