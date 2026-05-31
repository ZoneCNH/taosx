#!/usr/bin/env bash
set -euo pipefail

mkdir -p release/manifest

MODULE="$(go list -m)"
VERSION="${VERSION:-v0.1.0}"
COMMIT="$(git rev-parse HEAD 2>/dev/null || echo unknown)"
GO_VERSION="$(go version | awk '{print $3}')"
GENERATED_AT="$(date -u +"%Y-%m-%dT%H:%M:%SZ")"

cat > release/manifest/latest.json <<JSON
{
  "module": "${MODULE}",
  "version": "${VERSION}",
  "commit": "${COMMIT}",
  "go_version": "${GO_VERSION}",
  "generated_at": "${GENERATED_AT}",
  "checks": {
    "fmt": "manual-or-ci",
    "vet": "manual-or-ci",
    "unit_test": "manual-or-ci",
    "race_test": "manual-or-ci",
    "boundary": "manual-or-ci",
    "secret_scan": "manual-or-ci",
    "contract": "manual-or-ci"
  },
  "artifacts": [
    "release/manifest/latest.json"
  ],
  "notes": {
    "breaking_changes": "none",
    "known_risks": []
  }
}
JSON

echo "generated release/manifest/latest.json"
