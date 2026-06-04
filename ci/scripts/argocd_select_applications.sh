#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'USAGE'
Usage: argocd_select_applications.sh --selector <selector> --input <applications.json>

Select Argo CD application names from an applications list JSON payload.

Arguments:
  --selector  Comma-separated selector expression.
              Supported formats per part: key=value or key!=value
  --input     Path to Argo CD applications JSON payload (for example API /applications response)
USAGE
}

selector=""
input_file=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --selector)
      selector="${2:-}"
      shift 2
      ;;
    --input)
      input_file="${2:-}"
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

if [[ -z "${selector}" || -z "${input_file}" ]]; then
  echo "Both --selector and --input are required." >&2
  usage >&2
  exit 2
fi

if [[ ! -f "${input_file}" ]]; then
  echo "Input file not found: ${input_file}" >&2
  exit 2
fi

jq -r --arg selector "${selector}" '
  def parse_selector($selector):
    ($selector | split(",") | map(gsub("^\\s+|\\s+$"; "") | select(length > 0)));
  def selector_matches($selector):
    (.metadata.labels // {}) as $labels
    | all(parse_selector($selector)[];
        . as $selector_part
        | if ($selector_part | test("^[^!=]+=[^=]+$")) then
            ($selector_part | capture("^(?<key>[^=]+)=(?<value>.+)$")) as $m
            | (($labels[$m.key] // "") == $m.value)
          elif ($selector_part | test("^[^!=]+!=.+$")) then
            ($selector_part | capture("^(?<key>[^!=]+)!=(?<value>.+)$")) as $m
            | (($labels[$m.key] // "") != $m.value)
          else
            false
          end
      );
  .items[]
  | select(selector_matches($selector))
  | .metadata.name
' "${input_file}"