#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SELECTOR_SCRIPT="${SCRIPT_DIR}/argocd_select_applications.sh"

usage() {
  cat <<'USAGE'
Usage: argocd_refresh_sync.sh [--env-file <path>]

Options:
  --env-file <path>  Load environment variables from file before validation.
                     Expected keys: ARGOCD_ENDPOINT, ARGOCD_AUTH_TOKEN,
                     CF_ACCESS_CLIENT_ID, CF_ACCESS_CLIENT_SECRET,
                     APPLICATION_SELECTOR.
USAGE
}

env_file=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    --env-file)
      env_file="${2:-}"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [[ -n "${env_file}" ]]; then
  if [[ ! -f "${env_file}" ]]; then
    echo "::error::Environment file not found: ${env_file}"
    exit 2
  fi

  # shellcheck disable=SC1090
  set -a
  source "${env_file}"
  set +a
fi

HTTP_CLIENT="${ARGOCD_HTTP_CLIENT:-curl}"

require_command() {
  local command_name="$1"
  if ! command -v "${command_name}" >/dev/null 2>&1; then
    echo "::error::Missing required command: ${command_name}"
    exit 1
  fi
}

require_command "${HTTP_CLIENT}"
require_command jq

if [[ ! -x "${SELECTOR_SCRIPT}" ]]; then
  echo "::error::Selector script is missing or not executable: ${SELECTOR_SCRIPT}"
  exit 1
fi

for required_name in ARGOCD_ENDPOINT ARGOCD_AUTH_TOKEN CF_ACCESS_CLIENT_ID CF_ACCESS_CLIENT_SECRET APPLICATION_SELECTOR; do
  if [[ -z "${!required_name:-}" ]]; then
    echo "::error::${required_name} is required. Check the selected GitHub Environment variables and secrets."
    exit 1
  fi
done

TIMEOUT_SECONDS="${TIMEOUT_SECONDS:-1800}"
POLL_INTERVAL_SECONDS="${POLL_INTERVAL_SECONDS:-20}"
APPLICATION_SELECTOR="${APPLICATION_SELECTOR//[[:space:]]/}"

ARGOCD_BASE="${ARGOCD_ENDPOINT%/}"
if [[ "${ARGOCD_BASE}" != http://* && "${ARGOCD_BASE}" != https://* ]]; then
  ARGOCD_BASE="https://${ARGOCD_BASE}"
fi

if [[ -n "${ARGOCD_AUTH_TOKEN:-}" ]]; then
  echo "::add-mask::${ARGOCD_AUTH_TOKEN}"
fi
if [[ -n "${CF_ACCESS_CLIENT_SECRET:-}" ]]; then
  echo "::add-mask::${CF_ACCESS_CLIENT_SECRET}"
fi

curl_argocd() {
  "${HTTP_CLIENT}" -fsSL \
    -H "CF-Access-Client-Id: ${CF_ACCESS_CLIENT_ID}" \
    -H "CF-Access-Client-Secret: ${CF_ACCESS_CLIENT_SECRET}" \
    -H "Authorization: Bearer ${ARGOCD_AUTH_TOKEN}" \
    "$@"
}

refresh_application() {
  local app_name="$1"
  echo "Refreshing Argo CD application '${app_name}'..."
  curl_argocd \
    -X GET \
    "${ARGOCD_BASE}/api/v1/applications/${app_name}?refresh=normal" \
    >/dev/null
}

list_selected_applications() {
  local selector="$1"
  local response_file
  response_file="$(mktemp)"

  curl_argocd \
    -X GET \
    "${ARGOCD_BASE}/api/v1/applications" \
    > "${response_file}"

  if ! jq -e '.' "${response_file}" >/dev/null 2>&1; then
    echo "::error::Argo CD returned non-JSON response while listing applications." >&2
    head -c 400 "${response_file}" || true
    rm -f "${response_file}"
    return 1
  fi

  "${SELECTOR_SCRIPT}" --selector "${selector}" --input "${response_file}"

  rm -f "${response_file}"
}

sync_application() {
  local app_name="$1"
  echo "Syncing Argo CD application '${app_name}'..."
  curl_argocd \
    -X POST \
    -H 'Content-Type: application/json' \
    --data '{"prune":false,"dryRun":false}' \
    "${ARGOCD_BASE}/api/v1/applications/${app_name}/sync" \
    >/dev/null
}

wait_for_green() {
  local app_name="$1"
  local deadline=$((SECONDS + TIMEOUT_SECONDS))
  local status_file
  status_file="$(mktemp)"

  echo "Waiting for '${app_name}' to become Healthy and Synced..."
  while (( SECONDS < deadline )); do
    curl_argocd \
      -X GET \
      "${ARGOCD_BASE}/api/v1/applications/${app_name}" \
      > "${status_file}"

    if ! jq -e '.' "${status_file}" >/dev/null 2>&1; then
      echo "::error::Argo CD returned non-JSON response while checking '${app_name}'."
      head -c 400 "${status_file}" || true
      rm -f "${status_file}"
      return 1
    fi

    local health_status
    local sync_status
    local operation_phase
    health_status="$(jq -r '.status.health.status // "Unknown"' "${status_file}")"
    sync_status="$(jq -r '.status.sync.status // "Unknown"' "${status_file}")"
    operation_phase="$(jq -r '.status.operationState.phase // ""' "${status_file}")"

    echo "${app_name}: health=${health_status}, sync=${sync_status}, operation=${operation_phase:-none}"

    if [[ "${health_status}" == "Healthy" && "${sync_status}" == "Synced" ]]; then
      echo "${app_name} is Healthy and Synced."
      rm -f "${status_file}"
      return 0
    fi

    sleep "${POLL_INTERVAL_SECONDS}"
  done

  echo "::error::Timed out waiting for '${app_name}' to become Healthy and Synced."
  jq -r '.status.conditions // []' "${status_file}" || true
  rm -f "${status_file}"
  return 1
}

declare -a applications_to_sync
applications_output="$(list_selected_applications "${APPLICATION_SELECTOR}")" || exit 1
if [[ -z "${applications_output}" ]]; then
  applications_to_sync=()
else
  mapfile -t applications_to_sync <<< "${applications_output}"
fi
if [[ ${#applications_to_sync[@]} -eq 0 ]]; then
  echo "::error::No Argo CD applications matched selector '${APPLICATION_SELECTOR}'."
  exit 1
fi
echo "Selected Argo CD applications: ${applications_to_sync[*]}"

for app_name in "${applications_to_sync[@]}"; do
  refresh_application "${app_name}"
done

for app_name in "${applications_to_sync[@]}"; do
  sync_application "${app_name}"
done

for app_name in "${applications_to_sync[@]}"; do
  wait_for_green "${app_name}"
done

if [[ -n "${GITHUB_STEP_SUMMARY:-}" ]]; then
  {
    echo "### Argo CD deployment"
    echo ""
    echo "- Refreshed and synced applications matching: \`${APPLICATION_SELECTOR}\`"
    echo "- Matched applications: \`${applications_to_sync[*]}\`"
    echo "- Endpoint: \`${ARGOCD_BASE}\`"
  } >> "${GITHUB_STEP_SUMMARY}"
fi