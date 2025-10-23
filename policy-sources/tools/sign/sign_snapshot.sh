#!/usr/bin/env bash
set -euo pipefail
IN="$1"
OUT="${2:-${1%.json}.signed.json}"
HASH=$(jq -c 'del(.signature)' "$IN" | openssl dgst -sha256 -binary | xxd -p -c 256)
jq --arg h "sha256:$HASH" '.signature = $h' "$IN" > "$OUT"
echo "$OUT"
