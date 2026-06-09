#!/usr/bin/env bash
set -euo pipefail

CONTEXT=""
NAMESPACE="default"
DEPLOYMENT=""
TIMEOUT_SECONDS=180
INTERVAL_SECONDS=5

usage() {
  cat <<EOF
Usage: k8s-wait-ready.sh --namespace <namespace> --deployment <name> [options]

Options:
  --context <context>        Kubernetes context to use.
  --timeout <seconds>        Maximum wait time. Default: ${TIMEOUT_SECONDS}.
  --interval <seconds>       Poll interval. Default: ${INTERVAL_SECONDS}.

This script avoids kubectl watch/rollout status so it works on older
Kubernetes clusters where client-go bookmark watches can timeout.
EOF
}

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --context)
      CONTEXT="${2:-}"
      shift 2
      ;;
    --namespace)
      NAMESPACE="${2:-}"
      shift 2
      ;;
    --deployment)
      DEPLOYMENT="${2:-}"
      shift 2
      ;;
    --timeout)
      TIMEOUT_SECONDS="${2:-}"
      shift 2
      ;;
    --interval)
      INTERVAL_SECONDS="${2:-}"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -z "$DEPLOYMENT" ]]; then
  echo "--deployment is required" >&2
  usage >&2
  exit 2
fi

kubectl_args=()
if [[ -n "$CONTEXT" ]]; then
  kubectl_args+=(--context "$CONTEXT")
fi
kubectl_args+=(-n "$NAMESPACE")

deadline=$((SECONDS + TIMEOUT_SECONDS))
last_status=""

while [[ "$SECONDS" -le "$deadline" ]]; do
  if ! status="$(kubectl "${kubectl_args[@]}" get deploy "$DEPLOYMENT" \
    -o jsonpath='{.spec.replicas}{" "}{.status.updatedReplicas}{" "}{.status.readyReplicas}{" "}{.status.availableReplicas}{" "}{.status.observedGeneration}{" "}{.metadata.generation}' 2>&1)"; then
    last_status="$status"
    sleep "$INTERVAL_SECONDS"
    continue
  fi

  read -r desired updated ready available observed generation <<<"$status"
  desired="${desired:-0}"
  updated="${updated:-0}"
  ready="${ready:-0}"
  available="${available:-0}"
  observed="${observed:-0}"
  generation="${generation:-0}"
  last_status="desired=${desired} updated=${updated} ready=${ready} available=${available} observedGeneration=${observed} generation=${generation}"

  if [[ "$observed" == "$generation" && "$updated" == "$desired" && "$ready" == "$desired" && "$available" == "$desired" ]]; then
    echo "deployment/${DEPLOYMENT} ready (${last_status})"
    exit 0
  fi

  sleep "$INTERVAL_SECONDS"
done

echo "deployment/${DEPLOYMENT} not ready after ${TIMEOUT_SECONDS}s (${last_status})" >&2
kubectl "${kubectl_args[@]}" get deploy,pod,svc -o wide >&2 || true
exit 1
