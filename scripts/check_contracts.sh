#!/usr/bin/env bash
set -euo pipefail

echo "checking contracts..."

REQUIRED_FILES=(
  "contracts/config.schema.json"
  "contracts/health.schema.json"
  "contracts/error.schema.json"
  "contracts/metrics.md"
)

for file in "${REQUIRED_FILES[@]}"; do
  if [[ ! -f "$file" ]]; then
    echo "ERROR: missing contract file: $file"
    exit 1
  fi
done

echo "contract check passed"
