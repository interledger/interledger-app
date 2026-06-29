#!/usr/bin/env bash
# validate.sh — wait for deployments and health-check each via port-forward
set -euo pipefail

NAMESPACE="${1:-interledger}"
KUBE_CONTEXT="${2:-kind-interledger}"

KUBECTL="kubectl --context ${KUBE_CONTEXT}"

# Format: "deployment-name svc-name:port /health-path"
declare -a CHECKS=(
  "mock-services-mockpti-server    mock-services-mockpti-service:8080    /health"
  "mock-services-mockgatehub-server mock-services-mockgatehub-service:8080 /health"
  "mock-services-mockxago-server   mock-services-mockxago-service:8080   /health"
  "wallet-backend-server           wallet-backend-service-http:8080      /healthz"
  "wallet-frontend-server          wallet-frontend-service:3000          /healthz"
  "wallet-admin-server             wallet-admin-service:3000             /healthz"
)

# ── Wait for all deployments ───────────────────────────────────────────────────

DEPLOYMENTS=()
for entry in "${CHECKS[@]}"; do
  dep=$(echo "$entry" | awk '{print $1}')
  DEPLOYMENTS+=("$dep")
done

echo "==> Waiting for deployments in namespace '${NAMESPACE}'..."
${KUBECTL} wait deployment \
  "${DEPLOYMENTS[@]}" \
  --for=condition=available \
  --namespace "${NAMESPACE}" \
  --timeout=180s

# ── Health-check each service via port-forward ─────────────────────────────────

PF_PIDS=()

cleanup() {
  for pid in "${PF_PIDS[@]}"; do
    kill "$pid" 2>/dev/null || true
  done
}
trap cleanup EXIT

echo "==> Running health checks..."

FAILED=0
BASE_PORT=19100

for i in "${!CHECKS[@]}"; do
  entry="${CHECKS[$i]}"
  SVC=$(echo "$entry" | awk '{print $2}')
  SVC_NAME=$(echo "$SVC" | cut -d: -f1)
  SVC_PORT=$(echo "$SVC" | cut -d: -f2)
  HEALTH_PATH=$(echo "$entry" | awk '{print $3}')
  LOCAL_PORT=$((BASE_PORT + i))

  ${KUBECTL} port-forward \
    --namespace "${NAMESPACE}" \
    "svc/${SVC_NAME}" "${LOCAL_PORT}:${SVC_PORT}" \
    &>/dev/null &
  PF_PIDS+=($!)

  OK=false
  for _ in $(seq 1 30); do
    if curl -sf --max-time 1 "http://localhost:${LOCAL_PORT}${HEALTH_PATH}" >/dev/null 2>&1; then
      OK=true
      break
    fi
    sleep 0.5
  done

  if $OK; then
    echo "  ok  ${SVC_NAME}${HEALTH_PATH}"
  else
    echo "  FAIL ${SVC_NAME}${HEALTH_PATH} — did not respond within 15 s"
    FAILED=$((FAILED + 1))
  fi
done

if [ "${FAILED}" -gt 0 ]; then
  echo "==> ${FAILED} service(s) failed — see above"
  exit 1
fi

echo "==> All services healthy"
