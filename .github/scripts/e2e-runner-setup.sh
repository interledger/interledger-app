#!/bin/bash
# =============================================================================
# E2E Runner Startup Script
# =============================================================================
#
# This script runs on boot for spot instances created from the e2e-tester
# machine image. It reads configuration from GCP instance metadata, registers
# a GitHub Actions self-hosted runner, and starts it.
#
# INSTALLATION ON THE IMAGE:
#   1. Save this script to /opt/runner-setup.sh and make it executable
#   2. Install the systemd service (see e2e-runner.service below)
#   3. Remove existing runner registration:
#        cd /home/runner/actions-runner
#        sudo ./svc.sh stop
#        sudo ./svc.sh uninstall
#        ./config.sh remove --token <REMOVAL_TOKEN>
#   4. Save a new machine image
#
# To generate a removal token:
#   curl -X POST \
#     -H "Authorization: token <PAT>" \
#     -H "Accept: application/vnd.github+json" \
#     https://api.github.com/repos/interledger/interledger-app/actions/runners/remove-token \
#     | jq -r '.token'
#
# METADATA KEYS (set by the scaler workflow at instance creation):
#   runner-token   - GitHub runner registration token
#   runner-name    - Unique runner name (e.g., e2e-spot-1710072000-5432)
#   runner-labels  - Comma-separated labels (e.g., e2e-tester-dynamic)
# =============================================================================
set -euo pipefail

METADATA_URL="http://metadata.google.internal/computeMetadata/v1/instance/attributes"
METADATA_HEADER="Metadata-Flavor: Google"
RUNNER_DIR="/home/runner/actions-runner"
RUNNER_USER="runner"
REPO_URL="https://github.com/interledger/interledger-app"

log() { echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"; }

# Wait for metadata service to be available
for i in {1..30}; do
  if curl -sf -H "$METADATA_HEADER" "$METADATA_URL/" &>/dev/null; then
    break
  fi
  log "Waiting for metadata service (attempt $i/30)..."
  sleep 2
done

# Read metadata
RUNNER_TOKEN=$(curl -sf -H "$METADATA_HEADER" "$METADATA_URL/runner-token")
RUNNER_NAME=$(curl -sf -H "$METADATA_HEADER" "$METADATA_URL/runner-name")
RUNNER_LABELS=$(curl -sf -H "$METADATA_HEADER" "$METADATA_URL/runner-labels" 2>/dev/null || echo "")

if [ -z "$RUNNER_TOKEN" ] || [ -z "$RUNNER_NAME" ]; then
  log "ERROR: Missing required metadata (runner-token or runner-name). Exiting."
  exit 1
fi

log "Configuring runner: name=$RUNNER_NAME, labels=$RUNNER_LABELS"

cd "$RUNNER_DIR"

# Remove any existing configuration
if [ -f ".runner" ]; then
  log "Removing existing runner configuration..."
  sudo -u "$RUNNER_USER" ./config.sh remove --token "$RUNNER_TOKEN" 2>/dev/null || true
fi

# Build config command
CONFIG_ARGS=(
  --url "$REPO_URL"
  --token "$RUNNER_TOKEN"
  --name "$RUNNER_NAME"
  --work _work
  --unattended
  --replace
)

if [ -n "$RUNNER_LABELS" ]; then
  CONFIG_ARGS+=(--labels "$RUNNER_LABELS")
fi

# Register runner
log "Registering runner..."
sudo -u "$RUNNER_USER" ./config.sh "${CONFIG_ARGS[@]}"

# Install and start as a systemd service
log "Installing and starting runner service..."
./svc.sh install "$RUNNER_USER"
./svc.sh start

log "Runner $RUNNER_NAME is ready."
