#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUT_FILE="$(mktemp /tmp/pilotwave-openapi.XXXXXX.yaml)"
trap 'rm -f "$OUT_FILE"' EXIT

cd "$ROOT_DIR"
export GOCACHE="${GOCACHE:-$ROOT_DIR/tmp/go-build-cache}"
mkdir -p "$GOCACHE"

go run ./scripts/api-doc-sync.go \
  --input doc/Pilotwave.v1.yaml \
  --output "$OUT_FILE"

assert_contains() {
  local needle="$1"

  if ! grep -Fq "$needle" "$OUT_FILE"; then
    echo "FAIL: expected generated OpenAPI to contain: $needle" >&2
    exit 1
  fi
}

assert_contains "/namespace/{name}/istio-injection:"
assert_contains "/cluster/capabilities:"
assert_contains "/security/authpolicies:"
assert_contains "/security/requestauth/{namespace}/{name}:"
assert_contains "/gateway/{namespace}/{name}:"
assert_contains "/gateways/{namespace}/{name}/routers:"
assert_contains "/router/{namespace}/{name}/rules:"
assert_contains "operationId: patch-namespace-name-istio-injection"
assert_contains "operationId: get-security-requestauth-namespace-name"

echo "api doc route sync tests passed"
