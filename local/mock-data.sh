#!/usr/bin/env bash
# Export or import mock service data snapshots by talking directly to Valkey.
#
# Usage:
#   ./mock-data.sh build                  — build all four CLI tools
#   ./mock-data.sh export [output-dir]    — export all four services to JSON files
#   ./mock-data.sh import [input-dir]     — import all four services from JSON files
#
# The output-dir defaults to ./mock-data-snapshots.
#
# Each CLI tool reads its Valkey connection from the same env vars the services use:
#   MOCKGATEHUB_REDIS_URL / MOCKGATEHUB_REDIS_DB
#   MOCKPTI_REDIS_URL     / MOCKPTI_REDIS_DB
#   MOCKXAGO_REDIS_URL    / MOCKXAGO_REDIS_DB
#   MOCKCHIMONEY_REDIS_URL / MOCKCHIMONEY_REDIS_DB
#
# Examples:
#   ./mock-data.sh build
#   ./mock-data.sh export ./snapshots
#   ./mock-data.sh import ./snapshots

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GO_DIR="$REPO_ROOT/go"
BIN_DIR="$REPO_ROOT/local/.mock-data-bins"

build_tools() {
    mkdir -p "$BIN_DIR"
    echo "Building CLI tools..."
    (cd "$GO_DIR" && go build -o "$BIN_DIR/mockgatehub-data"  ./mock/mockgatehub/cmd/mockgatehub-data/...)
    (cd "$GO_DIR" && go build -o "$BIN_DIR/mockpti-data"      ./mock/mockpti/cmd/mockpti-data/...)
    (cd "$GO_DIR" && go build -o "$BIN_DIR/mockxago-data"     ./mock/mockxago/cmd/mockxago-data/...)
    (cd "$GO_DIR" && go build -o "$BIN_DIR/mockchimoney-data" ./mock/mockchimoney/cmd/mockchimoney-data/...)
    echo "Build complete → $BIN_DIR"
}

require_bins() {
    for bin in mockgatehub-data mockpti-data mockxago-data mockchimoney-data; do
        if [[ ! -x "$BIN_DIR/$bin" ]]; then
            echo "ERROR: $BIN_DIR/$bin not found — run './mock-data.sh build' first" >&2
            exit 1
        fi
    done
}

export_all() {
    local dir="${1:-./mock-data-snapshots}"
    require_bins
    mkdir -p "$dir"

    echo "Exporting mockgatehub..."
    "$BIN_DIR/mockgatehub-data" export --output "$dir/mockgatehub.json"

    echo "Exporting mockpti..."
    "$BIN_DIR/mockpti-data" export --output "$dir/mockpti.json"

    echo "Exporting mockxago..."
    "$BIN_DIR/mockxago-data" export --output "$dir/mockxago.json"

    echo "Exporting mockchimoney..."
    "$BIN_DIR/mockchimoney-data" export --output "$dir/mockchimoney.json"

    echo "Export complete → $dir"
    ls -lh "$dir"/*.json
}

import_all() {
    local dir="${1:-./mock-data-snapshots}"
    require_bins

    for f in mockgatehub.json mockpti.json mockxago.json mockchimoney.json; do
        if [[ ! -f "$dir/$f" ]]; then
            echo "ERROR: $dir/$f not found — run export first" >&2
            exit 1
        fi
    done

    echo "Importing mockgatehub..."
    "$BIN_DIR/mockgatehub-data" import --input "$dir/mockgatehub.json"

    echo "Importing mockpti..."
    "$BIN_DIR/mockpti-data" import --input "$dir/mockpti.json"

    echo "Importing mockxago..."
    "$BIN_DIR/mockxago-data" import --input "$dir/mockxago.json"

    echo "Importing mockchimoney..."
    "$BIN_DIR/mockchimoney-data" import --input "$dir/mockchimoney.json"

    echo "Import complete ← $dir"
}

case "${1:-}" in
    build)  build_tools ;;
    export) export_all "${2:-./mock-data-snapshots}" ;;
    import) import_all "${2:-./mock-data-snapshots}" ;;
    *)
        echo "Usage: $0 build|export|import [dir]"
        echo ""
        echo "  build         Build all four CLI tools"
        echo "  export [dir]  Export all mock service data to JSON files in dir"
        echo "  import [dir]  Import all mock service data from JSON files in dir"
        echo ""
        echo "  dir defaults to ./mock-data-snapshots"
        exit 1
        ;;
esac
