#!/usr/bin/env bash
set -euo pipefail

CONTEXT="${ISTIO_CONTEXT:-colima-legacy-1-18}"
NAMESPACE="${ISTIO_SMOKE_NAMESPACE:-pilotwave-istio-smoke}"
INJECTION_NAMESPACE="${ISTIO_SMOKE_INJECTION_NAMESPACE:-pilotwave-istio-injection-smoke}"
CURL_IMAGE="${ISTIO_SMOKE_CURL_IMAGE:-curlimages/curl:8.8.0}"
INGRESS_HOST="${ISTIO_SMOKE_INGRESS_HOST:-istio-ingressgateway.istio-system.svc.cluster.local}"
INGRESS_URL="http://${INGRESS_HOST}/"
HTTP_HOST="smoke.pilotwave.local"
TLS_HOST="smoke-tls.pilotwave.local"
TLS_SECRET="pilotwave-istio-smoke-tls"
TLS_CONNECT="${TLS_HOST}:443:${INGRESS_HOST}:443"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FIXTURE_DIR="${ROOT_DIR}/testcase/istio-smoke/manifests"
TMP_DIR=""

kubectl_cmd() {
  kubectl --context "${CONTEXT}" "$@"
}

cleanup_tmp() {
  if [[ -n "${TMP_DIR}" ]]; then
    rm -rf "${TMP_DIR}"
  fi
}
trap cleanup_tmp EXIT INT TERM

delete_pod() {
  kubectl_cmd -n "${NAMESPACE}" delete pod "$1" --ignore-not-found=true >/dev/null
}

wait_pod_done() {
  local pod="$1"
  local phase=""

  for _ in $(seq 1 60); do
    phase="$(kubectl_cmd -n "${NAMESPACE}" get pod "${pod}" -o jsonpath='{.status.phase}' 2>/dev/null || true)"
    case "${phase}" in
      Succeeded) return 0 ;;
      Failed) return 1 ;;
    esac
    sleep 1
  done

  echo "Timed out waiting for pod ${pod}; last phase=${phase}" >&2
  return 1
}

wait_deployment_available() {
  local deployment="$1"
  local available=""

  for _ in $(seq 1 120); do
    available="$(kubectl_cmd -n "${NAMESPACE}" get deployment "${deployment}" -o jsonpath='{.status.availableReplicas}' 2>/dev/null || true)"
    if [[ "${available}" == "1" ]]; then
      return 0
    fi
    sleep 1
  done

  kubectl_cmd -n "${NAMESPACE}" get deployment "${deployment}" -o wide || true
  kubectl_cmd -n "${NAMESPACE}" get pods -l app=hello -o wide || true
  echo "Timed out waiting for deployment ${deployment} to become available" >&2
  return 1
}

run_curl_pod() {
  local pod="$1"
  shift

  delete_pod "${pod}"
  kubectl_cmd -n "${NAMESPACE}" run "${pod}" --restart=Never --image="${CURL_IMAGE}" -- "$@" >/dev/null

  if ! wait_pod_done "${pod}"; then
    kubectl_cmd -n "${NAMESPACE}" logs "${pod}" || true
    delete_pod "${pod}"
    return 1
  fi

  kubectl_cmd -n "${NAMESPACE}" logs "${pod}"
  delete_pod "${pod}"
}

expect_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"

  if [[ "${actual}" != "${expected}" ]]; then
    echo "FAIL ${label}: expected '${expected}', got '${actual}'" >&2
    exit 1
  fi

  echo "PASS ${label}: ${actual}"
}

expect_http_status_with_retry() {
  local pod_prefix="$1"
  local expected="$2"
  local label="$3"
  shift 3

  local actual=""
  for attempt in $(seq 1 30); do
    actual="$(run_curl_pod "${pod_prefix}-${attempt}" "$@")"
    if [[ "${actual}" == "${expected}" ]]; then
      echo "PASS ${label}: ${actual}"
      return 0
    fi
    sleep 1
  done

  echo "FAIL ${label}: expected '${expected}', got '${actual}'" >&2
  exit 1
}

expect_label() {
  local namespace="$1"
  local key="$2"
  local expected="$3"
  local actual=""

  actual="$(kubectl_cmd get namespace "${namespace}" -o "jsonpath={.metadata.labels['${key//./\\.}']}" 2>/dev/null || true)"
  expect_equals "${actual}" "${expected}" "namespace ${namespace} label ${key}"
}

expect_missing_label() {
  local namespace="$1"
  local key="$2"
  local actual=""

  actual="$(kubectl_cmd get namespace "${namespace}" -o "jsonpath={.metadata.labels['${key//./\\.}']}" 2>/dev/null || true)"
  if [[ -n "${actual}" ]]; then
    echo "FAIL namespace ${namespace} label ${key}: expected missing, got '${actual}'" >&2
    exit 1
  fi

  echo "PASS namespace ${namespace} label ${key}: missing"
}

cleanup_security_policies() {
  kubectl_cmd -n istio-system delete authorizationpolicy pilotwave-istio-smoke-deny-path --ignore-not-found=true >/dev/null
  kubectl_cmd -n istio-system delete authorizationpolicy pilotwave-istio-smoke-require-jwt --ignore-not-found=true >/dev/null
  kubectl_cmd -n istio-system delete requestauthentication pilotwave-istio-smoke-jwt --ignore-not-found=true >/dev/null
}

ensure_tls_secret() {
  if ! command -v openssl >/dev/null 2>&1; then
    echo "openssl is required to generate the temporary TLS certificate for the ClusterIP Gateway check" >&2
    exit 1
  fi

  TMP_DIR="$(mktemp -d)"
  openssl req -x509 -nodes -newkey rsa:2048 -days 7 \
    -subj "/CN=${TLS_HOST}" \
    -addext "subjectAltName=DNS:${TLS_HOST}" \
    -keyout "${TMP_DIR}/tls.key" \
    -out "${TMP_DIR}/tls.crt" >/dev/null 2>&1

  kubectl_cmd -n istio-system create secret tls "${TLS_SECRET}" \
    --cert="${TMP_DIR}/tls.crt" \
    --key="${TMP_DIR}/tls.key" \
    --dry-run=client -o yaml | kubectl_cmd apply -f - >/dev/null
}

validate_namespace_injection_labels() {
  echo "Applying namespace injection fixture"
  kubectl_cmd apply -f "${FIXTURE_DIR}/injection-namespace.yaml" >/dev/null
  kubectl_cmd label namespace "${INJECTION_NAMESPACE}" istio-injection=enabled istio.io/rev- --overwrite >/dev/null
  expect_label "${INJECTION_NAMESPACE}" "istio-injection" "enabled"
  expect_missing_label "${INJECTION_NAMESPACE}" "istio.io/rev"

  echo "Switching namespace injection to disabled"
  kubectl_cmd label namespace "${INJECTION_NAMESPACE}" istio-injection=disabled istio.io/rev- --overwrite >/dev/null
  expect_label "${INJECTION_NAMESPACE}" "istio-injection" "disabled"
  expect_missing_label "${INJECTION_NAMESPACE}" "istio.io/rev"

  echo "Switching namespace injection to revision label"
  kubectl_cmd label namespace "${INJECTION_NAMESPACE}" istio-injection- istio.io/rev=pilotwave-smoke-rev --overwrite >/dev/null
  expect_missing_label "${INJECTION_NAMESPACE}" "istio-injection"
  expect_label "${INJECTION_NAMESPACE}" "istio.io/rev" "pilotwave-smoke-rev"
}

cleanup_security_policies

echo "Running ClusterIP-only Istio smoke suite on context ${CONTEXT}"
validate_namespace_injection_labels

echo "Applying base Gateway, VirtualService, DestinationRule, Service, and workloads"
kubectl_cmd apply -f "${FIXTURE_DIR}/base.yaml" >/dev/null
wait_deployment_available hello-v1
wait_deployment_available hello-v2

v1_output="$(run_curl_pod curl-v1-smoke curl -sS -H "Host: ${HTTP_HOST}" "${INGRESS_URL}")"
expect_equals "${v1_output}" "hello from v1" "Gateway plus VirtualService route to v1 subset"

echo "Applying 100% v2 VirtualService"
kubectl_cmd apply -f "${FIXTURE_DIR}/route-v2.yaml" >/dev/null
v2_output="$(run_curl_pod curl-v2-smoke curl -sS -H "Host: ${HTTP_HOST}" "${INGRESS_URL}")"
expect_equals "${v2_output}" "hello from v2" "VirtualService route to v2 subset"

echo "Applying 75/25 weighted VirtualService"
kubectl_cmd apply -f "${FIXTURE_DIR}/weighted-v2-75-v1-25.yaml" >/dev/null
weighted_output="$(run_curl_pod curl-weighted-smoke sh -c "for i in \$(seq 1 40); do curl -s -H 'Host: ${HTTP_HOST}' ${INGRESS_URL}; done | sort | uniq -c")"
echo "${weighted_output}"

if ! grep -q "hello from v1" <<<"${weighted_output}" || ! grep -q "hello from v2" <<<"${weighted_output}"; then
  echo "FAIL weighted VirtualService: expected both DestinationRule subsets in the sample" >&2
  exit 1
fi
echo "PASS weighted VirtualService: saw both v1 and v2 subsets"

echo "Applying HTTPS Gateway over ingress ClusterIP"
ensure_tls_secret
kubectl_cmd apply -f "${FIXTURE_DIR}/gateway-tls.yaml" >/dev/null
tls_output="$(run_curl_pod curl-tls-smoke curl -k -sS --connect-to "${TLS_CONNECT}" "https://${TLS_HOST}/")"
expect_equals "${tls_output}" "hello from v1" "TLS Gateway route through ClusterIP"

echo "Applying AuthorizationPolicy deny fixture"
kubectl_cmd apply -f "${FIXTURE_DIR}/authz-deny-path.yaml" >/dev/null
deny_status="$(run_curl_pod curl-deny-smoke curl -sS -o /tmp/body -w "%{http_code}" -H "Host: ${HTTP_HOST}" "${INGRESS_URL}deny")"
expect_equals "${deny_status}" "403" "AuthorizationPolicy denies /deny"
kubectl_cmd -n istio-system delete authorizationpolicy pilotwave-istio-smoke-deny-path --ignore-not-found=true >/dev/null

echo "Applying RequestAuthentication plus JWT-required AuthorizationPolicy"
kubectl_cmd apply -f "${FIXTURE_DIR}/requestauth-require-jwt.yaml" >/dev/null
expect_http_status_with_retry curl-jwt-smoke 403 "RequestAuthentication path rejects missing JWT" \
  curl -sS -o /tmp/body -w "%{http_code}" -H "Host: ${HTTP_HOST}" "${INGRESS_URL}jwt"
cleanup_security_policies

echo "Resetting default route to v1"
kubectl_cmd apply -f "${FIXTURE_DIR}/base.yaml" >/dev/null

echo "SMOKE_OK context=${CONTEXT} namespace=${NAMESPACE} ingress=${INGRESS_HOST}"
