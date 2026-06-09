#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${PILOTWAVE_BASE_URL:-http://127.0.0.1:22112}"
METRICS_URL="${BASE_URL%/}/metrics"

body="$(curl -fsS "${METRICS_URL}")"

require_metric() {
  local name="$1"

  if ! grep -q "^${name}" <<<"${body}"; then
    echo "FAIL missing ${name} from ${METRICS_URL}" >&2
    exit 1
  fi

  echo "PASS ${name}"
}

require_metric "pilotwave_build_info"
require_metric "pilotwave_http_requests_total"
require_metric "pilotwave_http_request_duration_seconds"

echo "SMOKE_OK ${METRICS_URL}"
