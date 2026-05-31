#!/usr/bin/env bash
set -euo pipefail

echo "checking secrets..."

PATTERNS=(
  "password="
  "passwd="
  "secret="
  "token="
  "access_key="
  "secret_key="
  "AKIA[0-9A-Z]{16}"
  "BEGIN RSA PRIVATE KEY"
  "BEGIN OPENSSH PRIVATE KEY"
)

for pattern in "${PATTERNS[@]}"; do
  if grep -R -E "$pattern" . \
    --exclude-dir=.git \
    --exclude-dir=.omx \
    --exclude-dir=vendor \
    --exclude="*.sum" \
    --exclude="check_secrets.sh" \
    --exclude="goal.md"; then
    echo "ERROR: possible secret found: $pattern"
    exit 1
  fi
done

echo "secret check passed"
