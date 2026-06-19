#!/usr/bin/env bash
# validate.sh — wait for mock-service deployments and health-check each via port-forward
set -euo pipefail

NAMESPACE="${1:-mock-services}"
KUBE_CONTEXT="${2:-kind-interledger-ci}"
RELEASE="mock-services"

KUBECTL="kubectl --context ${KUBE_CONTEXT}"

SERVICES=(
  mockpti-service
  mockgatehub-service
  mockxago-service
)

DEPLOYMENTS=(
  mockpti-server
  mockgatehub-server
  mockxago-server
)

# ── Wait for all deployments ───────────────────────────────────────────────────

echo "==> Waiting for deployments in namespace '${NAMESPACE}'..."
${KUBECTL} wait deployment \
  $(printf "${RELEASE}-%s " "${DEPLOYMENTS[@]}") \
  --for=condition=available \
  --namespace "${NAMESPACE}" \
  --timeout=120s

# ── Health-check each service via port-forward ─────────────────────────────────

PF_PIDS=()

cleanup() {
  for pid in "${PF_PIDS[@]:-}"; do
    kill "$pid" 2>/dev/null || true
  done
}
trap cleanup EXIT

echo "==> Running health checks..."

FAILED=0
BASE_PORT=19100

for i in "${!SERVICES[@]}"; do
  SVC="${RELEASE}-${SERVICES[$i]}"
  LOCAL_PORT=$((BASE_PORT + i))

  ${KUBECTL} port-forward \
    --namespace "${NAMESPACE}" \
    "svc/${SVC}" "${LOCAL_PORT}:8080" \
    &>/dev/null &
  PF_PIDS+=($!)

  # Retry until the port-forward is ready (up to 10 s)
  OK=false
  for _ in $(seq 1 20); do
    if curl -sf --max-time 1 "http://localhost:${LOCAL_PORT}/health" >/dev/null 2>&1; then
      OK=true
      break
    fi
    sleep 0.5
  done

  if $OK; then
    echo "  ✓ ${SVC}"
  else
    echo "  ✗ ${SVC} — /health did not respond within 10 s"
    FAILED=$((FAILED + 1))
  fi
done

if [ "${FAILED}" -gt 0 ]; then
  echo "==> ${FAILED} service(s) failed — see above"
  exit 1
fi

echo "==> All services healthy"
