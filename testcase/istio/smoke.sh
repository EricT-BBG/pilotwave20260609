#!/usr/bin/env bash
set -euo pipefail

CONTEXT="${ISTIO_CONTEXT:-colima-legacy-1-18}"
NAMESPACE="${ISTIO_SMOKE_NAMESPACE:-pilotwave-istio-demo}"
CURL_IMAGE="${ISTIO_SMOKE_CURL_IMAGE:-curlimages/curl:8.8.0}"
INGRESS_URL="http://istio-ingressgateway.istio-system.svc.cluster.local/"
TLS_CONNECT="hello-tls.pilotwave.local:443:istio-ingressgateway.istio-system.svc.cluster.local:443"

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
FIXTURE_DIR="${ROOT_DIR}/testcase/istio"

kubectl_cmd() {
  kubectl --context "${CONTEXT}" "$@"
}

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

cleanup_security_policies() {
  kubectl_cmd -n istio-system delete authorizationpolicy pilotwave-deny-hello-path --ignore-not-found=true >/dev/null
  kubectl_cmd -n istio-system delete authorizationpolicy pilotwave-require-jwt-path --ignore-not-found=true >/dev/null
  kubectl_cmd -n istio-system delete requestauthentication pilotwave-hello-jwt --ignore-not-found=true >/dev/null
}

cleanup_security_policies

echo "Applying base routing fixture on context ${CONTEXT}"
kubectl_cmd apply -f "${FIXTURE_DIR}/hello-routing/base/manifest.yaml" >/dev/null
wait_deployment_available hello-v1
wait_deployment_available hello-v2

v1_output="$(run_curl_pod curl-v1-smoke curl -sS -H "Host: hello.pilotwave.local" "${INGRESS_URL}")"
expect_equals "${v1_output}" "hello from v1" "base route v1"

echo "Applying v2 route fixture"
kubectl_cmd apply -f "${FIXTURE_DIR}/hello-routing/route-v2/virtualservice.yaml" >/dev/null
v2_output="$(run_curl_pod curl-v2-smoke curl -sS -H "Host: hello.pilotwave.local" "${INGRESS_URL}")"
expect_equals "${v2_output}" "hello from v2" "route v2"

echo "Applying weighted route fixture"
kubectl_cmd apply -f "${FIXTURE_DIR}/hello-routing/weighted-v2-75-v1-25/virtualservice.yaml" >/dev/null
weighted_output="$(run_curl_pod curl-weighted-smoke sh -c "for i in \$(seq 1 40); do curl -s -H 'Host: hello.pilotwave.local' ${INGRESS_URL}; done | sort | uniq -c")"
echo "${weighted_output}"

if ! grep -q "hello from v1" <<<"${weighted_output}" || ! grep -q "hello from v2" <<<"${weighted_output}"; then
  echo "FAIL weighted route: expected both v1 and v2 responses" >&2
  exit 1
fi
echo "PASS weighted route: saw both v1 and v2 responses"

echo "Applying TLS gateway fixture"
kubectl_cmd apply -f "${FIXTURE_DIR}/gateway-tls/manifest.yaml" >/dev/null
tls_output="$(run_curl_pod curl-tls-smoke curl -k -sS --connect-to "${TLS_CONNECT}" https://hello-tls.pilotwave.local/)"
expect_equals "${tls_output}" "hello from v1" "gateway tls"

echo "Applying AuthorizationPolicy deny fixture"
kubectl_cmd apply -f "${FIXTURE_DIR}/security/authz-deny-path/authorizationpolicy.yaml" >/dev/null
deny_status="$(run_curl_pod curl-deny-smoke curl -sS -o /tmp/body -w "%{http_code}" -H "Host: hello.pilotwave.local" "${INGRESS_URL}deny")"
expect_equals "${deny_status}" "403" "authorization policy deny"
kubectl_cmd -n istio-system delete authorizationpolicy pilotwave-deny-hello-path --ignore-not-found=true >/dev/null

echo "Applying RequestAuthentication missing-JWT fixture"
kubectl_cmd apply -f "${FIXTURE_DIR}/security/requestauth-require-jwt/manifest.yaml" >/dev/null
jwt_status="$(run_curl_pod curl-jwt-smoke curl -sS -o /tmp/body -w "%{http_code}" -H "Host: hello.pilotwave.local" "${INGRESS_URL}jwt")"
expect_equals "${jwt_status}" "403" "request authentication missing jwt"
cleanup_security_policies

echo "Resetting route to base fixture"
kubectl_cmd apply -f "${FIXTURE_DIR}/hello-routing/base/manifest.yaml" >/dev/null

echo "SMOKE_OK context=${CONTEXT} namespace=${NAMESPACE}"
