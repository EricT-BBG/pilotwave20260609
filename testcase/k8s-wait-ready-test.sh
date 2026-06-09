#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPT="$ROOT_DIR/scripts/k8s-wait-ready.sh"

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

help_output="$(bash "$SCRIPT" --help)"
assert_contains "$help_output" "Usage: k8s-wait-ready.sh"
assert_contains "$help_output" "--deployment"
assert_contains "$help_output" "avoids kubectl watch/rollout status"

set +e
missing_output="$(bash "$SCRIPT" --namespace pilotwave 2>&1)"
missing_status=$?
set -e

if [[ "$missing_status" -eq 0 ]]; then
  fail "missing --deployment should fail"
fi
assert_contains "$missing_output" "--deployment is required"

echo "k8s-wait-ready tests passed"
