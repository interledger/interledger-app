#!/usr/bin/env bash
set -euo pipefail

SOURCE_DIR="${1:-e2e/debug}"
OUTPUT_DIR="${2:-e2e/baseline-export}"
# TEST_NAME is a metadata label only. It names the GCS path and manifest entry.
# It does NOT filter which tests run. Test scope is controlled by the -tags flag
# passed to `go test` before this script is invoked.
TEST_NAME="${3:-gatehub-p2p-payment}"
COMMIT_SHA="${4:-unknown}"
RUN_ID="${5:-unknown}"
RUN_URL="${6:-unknown}"
CAPTURED_AT="${7:-$(date -u +%Y-%m-%dT%H:%M:%SZ)}"

if [[ ! -d "$SOURCE_DIR" ]]; then
  echo "source directory does not exist: $SOURCE_DIR" >&2
  exit 1
fi

rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR/images"

manifest_rows_file="$(mktemp)"

# Collect and normalize screenshot files from debug/<feature>__<scenario>/NN-name.png.
# For backward compatibility, older <testid>-NN-name.png files are also normalized.
while IFS= read -r file; do
  rel_path="${file#${SOURCE_DIR}/}"
  scenario_dir="${rel_path%%/*}"
  filename="${rel_path##*/}"

  # Remove random test identifier prefix. Expected format: <digits>-NN-name.png
  canonical_filename="$filename"
  if [[ "$filename" =~ ^[0-9]+-(.+)$ ]]; then
    canonical_filename="${BASH_REMATCH[1]}"
  fi

  logical_key="${scenario_dir}/${canonical_filename}"
  output_rel="images/${logical_key}"
  output_abs="${OUTPUT_DIR}/${output_rel}"

  mkdir -p "$(dirname "$output_abs")"
  cp "$file" "$output_abs"

  checksum="$(sha256sum "$output_abs" | awk '{print $1}')"
  printf '%s\t%s\t%s\n' "$logical_key" "$output_rel" "$checksum" >> "$manifest_rows_file"
done < <(find "$SOURCE_DIR" -type f -name '*.png' | sort)

if [[ ! -s "$manifest_rows_file" ]]; then
  echo "no PNG screenshots found under $SOURCE_DIR" >&2
  exit 1
fi

sorted_rows_file="$(mktemp)"
sort "$manifest_rows_file" > "$sorted_rows_file"
image_count="$(wc -l < "$sorted_rows_file" | tr -d ' ')"

manifest_file="${OUTPUT_DIR}/manifest.json"
{
  printf '{\n'
  printf '  "testName": "%s",\n' "$TEST_NAME"
  printf '  "sourceCommit": "%s",\n' "$COMMIT_SHA"
  printf '  "runId": "%s",\n' "$RUN_ID"
  printf '  "runUrl": "%s",\n' "$RUN_URL"
  printf '  "capturedAt": "%s",\n' "$CAPTURED_AT"
  printf '  "imageCount": %s,\n' "$image_count"
  printf '  "images": [\n'

  first="true"
  while IFS=$'\t' read -r logical_key output_rel checksum; do
    if [[ "$first" == "true" ]]; then
      first="false"
    else
      printf ',\n'
    fi
    printf '    {"logicalKey":"%s","path":"%s","sha256":"%s"}' "$logical_key" "$output_rel" "$checksum"
  done < "$sorted_rows_file"

  printf '\n  ]\n'
  printf '}\n'
} > "$manifest_file"

echo "exported baseline assets to $OUTPUT_DIR"
echo "manifest: $manifest_file"

echo "image_count=$image_count"

rm -f "$manifest_rows_file" "$sorted_rows_file"
