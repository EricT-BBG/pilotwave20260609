#!/usr/bin/env bash
set -euo pipefail

CONTEXT="${ISTIO_CONTEXT:-colima-legacy-1-18}"
BASE_URL="${E2E_BASE_URL:-http://127.0.0.1:22112}"
START_SERVER="${E2E_START_SERVER:-true}"
TEST_REVISION="${ISTIO_E2E_TEST_REVISION:-test-revision}"
TEST_REVISION_DEPLOYMENT="pilotwave-e2e-istiod-revision"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SERVER_PID=""
KUBECONFIG_FILE=""
SERVER_BIN=""
SERVER_LOG=""
CREATED_TEST_REVISION=""

cleanup() {
  if [[ -n "${SERVER_PID}" ]]; then
    kill "${SERVER_PID}" >/dev/null 2>&1 || true
    wait "${SERVER_PID}" >/dev/null 2>&1 || true
  fi
  if [[ -n "${SERVER_LOG}" ]]; then
    rm -f "${SERVER_LOG}"
  fi
  if [[ -n "${SERVER_BIN}" ]]; then
    rm -f "${SERVER_BIN}"
  fi
  if [[ -n "${KUBECONFIG_FILE}" ]]; then
    rm -f "${KUBECONFIG_FILE}"
  fi
  if [[ -n "${CREATED_TEST_REVISION}" ]]; then
    kubectl --context "${CONTEXT}" -n istio-system delete deployment "${TEST_REVISION_DEPLOYMENT}" --ignore-not-found=true >/dev/null 2>&1 || true
  fi
}
trap cleanup EXIT INT TERM

wait_for_url() {
  local url="$1"
  for _ in $(seq 1 120); do
    if curl -fsS "${url}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done
  echo "Timed out waiting for ${url}" >&2
  return 1
}

if [[ "${START_SERVER}" == "true" ]]; then
  kubectl --context "${CONTEXT}" -n istio-system apply -f - >/dev/null <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ${TEST_REVISION_DEPLOYMENT}
  labels:
    app: istiod
    istio.io/rev: ${TEST_REVISION}
spec:
  replicas: 0
  selector:
    matchLabels:
      app: pilotwave-e2e-istiod-revision
  template:
    metadata:
      labels:
        app: pilotwave-e2e-istiod-revision
        istio.io/rev: ${TEST_REVISION}
    spec:
      containers:
        - name: pause
          image: registry.k8s.io/pause:3.2
EOF
  CREATED_TEST_REVISION="true"

  KUBECONFIG_FILE="$(mktemp)"
  kubectl config view --raw --minify --context "${CONTEXT}" > "${KUBECONFIG_FILE}"

  (
    cd "${ROOT_DIR}"
    make build-web generate-go
  )

  SERVER_BIN="$(mktemp -t pilotwave-e2e-server)"
  SERVER_LOG="$(mktemp -t pilotwave-e2e-server-log)"
  (
    cd "${ROOT_DIR}"
    go build -o "${SERVER_BIN}" ./cmd/pilotwave
  )

  (
    cd "${ROOT_DIR}"
    KUBECONFIG="${KUBECONFIG_FILE}" "${SERVER_BIN}"
  ) >"${SERVER_LOG}" 2>&1 &
  SERVER_PID="$!"

  wait_for_url "${BASE_URL}/"
fi

(
  cd "${ROOT_DIR}/web"
  ISTIO_CONTEXT="${CONTEXT}" E2E_BASE_URL="${BASE_URL}" npx playwright test --reporter=line
)
