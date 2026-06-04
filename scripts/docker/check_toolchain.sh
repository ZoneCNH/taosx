#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
report_dir="$repo_root/release/docker"
summary_dir="$repo_root/release/evidence"
report="$report_dir/toolchain-check.md"
summary="$summary_dir/docker-toolchain-summary.json"
mode="static"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --drift)
      mode="drift"
      shift
      ;;
    -h|--help)
      echo "usage: scripts/docker/check_toolchain.sh [--drift]"
      exit 0
      ;;
    *)
      echo "ERROR: unknown docker toolchain check argument $1" >&2
      exit 2
      ;;
  esac
done

if [[ "$mode" == "drift" ]]; then
  report="$report_dir/toolchain-drift.md"
fi

mkdir -p "$report_dir" "$summary_dir"

declare -a check_names=()
declare -a check_statuses=()
declare -a check_details=()
failures=0

add_check() {
  local name="$1"
  local status="$2"
  local detail="$3"
  check_names+=("$name")
  check_statuses+=("$status")
  check_details+=("$detail")
  if [[ "$status" != "passed" ]]; then
    failures=$((failures + 1))
  fi
}

json_escape() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

file_exists() {
  [[ -f "$repo_root/$1" ]]
}

file_contains() {
  local file="$1"
  local text="$2"
  [[ -f "$repo_root/$file" ]] && grep -Fq -- "$text" "$repo_root/$file"
}

file_not_contains() {
  local file="$1"
  local text="$2"
  [[ -f "$repo_root/$file" ]] && ! grep -Fq -- "$text" "$repo_root/$file"
}

file_matches() {
  local file="$1"
  local pattern="$2"
  [[ -f "$repo_root/$file" ]] && grep -Eq -- "$pattern" "$repo_root/$file"
}

target_exists() {
  local target="$1"
  grep -Eq "^${target}:" "$repo_root/Makefile"
}

extract_go_mod_version() {
  awk '$1 == "go" {print $2; exit}' "$repo_root/go.mod"
}

extract_tool_versions_go() {
  awk '$1 == "golang" {print $2; exit}' "$repo_root/.tool-versions"
}

dockerfile_arg_value() {
  local name="$1"
  sed -n "s/^ARG ${name}=//p" "$repo_root/Dockerfile" | head -n1
}

normalize_go_minor() {
  local version="${1#go}"
  local major minor patch
  IFS=. read -r major minor patch <<< "$version"
  if [[ -z "$major" || -z "$minor" ]]; then
    printf '%s\n' "$version"
  else
    printf '%s.%s\n' "$major" "$minor"
  fi
}

check_file() {
  local file="$1"
  if file_exists "$file"; then
    add_check "file:$file" "passed" "$file exists"
  else
    add_check "file:$file" "failed" "$file is missing"
  fi
}

check_contains() {
  local file="$1"
  local text="$2"
  local label="$3"
  if file_contains "$file" "$text"; then
    add_check "$label" "passed" "$file contains $text"
  else
    add_check "$label" "failed" "$file must contain $text"
  fi
}

check_not_contains() {
  local file="$1"
  local text="$2"
  local label="$3"
  if file_not_contains "$file" "$text"; then
    add_check "$label" "passed" "$file does not contain $text"
  else
    add_check "$label" "failed" "$file must not contain $text"
  fi
}

check_regex() {
  local file="$1"
  local pattern="$2"
  local label="$3"
  if file_matches "$file" "$pattern"; then
    add_check "$label" "passed" "$file matches $pattern"
  else
    add_check "$label" "failed" "$file must match $pattern"
  fi
}

check_equal() {
  local name="$1"
  local left="$2"
  local right="$3"
  local detail="$4"
  if [[ "$left" == "$right" ]]; then
    add_check "$name" "passed" "$detail: $left"
  else
    add_check "$name" "failed" "$detail drifted: $left != $right"
  fi
}

check_target() {
  local target="$1"
  if target_exists "$target"; then
    add_check "make:$target" "passed" "Makefile exposes $target"
  else
    add_check "make:$target" "failed" "Makefile must expose $target"
  fi
}

docker_cli="unavailable"
docker_daemon="unavailable"
docker_buildx="unavailable"
docker_compose="unavailable"
if command -v docker >/dev/null 2>&1; then
  docker_cli="$(docker --version 2>/dev/null || true)"
  if docker version >/dev/null 2>&1; then
    docker_daemon="available"
  fi
  docker_buildx="$(docker buildx version 2>/dev/null || true)"
  if [[ -z "$docker_buildx" ]]; then
    docker_buildx="unavailable"
  fi
  docker_compose="$(docker compose version 2>/dev/null || true)"
  if [[ -z "$docker_compose" ]]; then
    docker_compose="unavailable"
  fi
fi
add_check "docker:cli-static-metadata" "passed" "docker CLI: $docker_cli; daemon: $docker_daemon; buildx: $docker_buildx; compose: $docker_compose"

go_mod_version="$(extract_go_mod_version)"
tool_versions_go="$(extract_tool_versions_go)"
dockerfile_go_version="$(dockerfile_arg_value GO_VERSION)"
dockerfile_go_base_digest="$(dockerfile_arg_value GO_BASE_IMAGE_DIGEST)"
check_equal "drift:go-mod-tool-versions" "$(normalize_go_minor "$go_mod_version")" "$(normalize_go_minor "$tool_versions_go")" "go.mod and .tool-versions Go minor"
check_equal "drift:docker-go-version" "$(normalize_go_minor "$go_mod_version")" "$(normalize_go_minor "$dockerfile_go_version")" "go.mod and Dockerfile Go minor"

for file in \
  "Dockerfile" \
  "docker-compose.yml" \
  ".dockerignore" \
  ".devcontainer/devcontainer.json" \
  "scripts/docker/check_toolchain.sh" \
  "scripts/docker/docker_gate.sh" \
  "scripts/docker/prefetch_tools.sh" \
  ".github/workflows/docker-contract.yml" \
  "contracts/docker-toolchain.schema.json" \
  "docs/standard/docker-toolchain-standard.md" \
  "docs/self-improving/docker-toolchain.md" \
  ".agent/retrospective/docker-toolchain.md" \
  ".agent/evidence/docker-toolchain.jsonl" \
  ".agent/metrics/docker-toolchain.json"; do
  check_file "$file"
done

for text in \
  "ARG GO_VERSION=1.23" \
  "ARG GO_BASE_IMAGE=golang:\${GO_VERSION}-bookworm" \
  "ARG GO_BASE_IMAGE_DIGEST=sha256:" \
  "ARG GOLANGCI_LINT_VERSION=v2.1.6" \
  "ARG GOVULNCHECK_VERSION=v1.1.4" \
  "AS toolchain" \
  "AS dev" \
  "AS gate" \
  "AS build-goalcli" \
  "AS goalcli-runtime" \
  "--mount=type=cache,target=/go/pkg/mod" \
  "--mount=type=cache,target=/root/.cache/go-build" \
  "golangci-lint" \
  "govulncheck" \
  "COPY --from=tools" \
  "FROM scratch AS goalcli-runtime" \
  "ENTRYPOINT [\"/goalcli\"]"; do
  check_contains "Dockerfile" "$text" "dockerfile:$text"
done
check_not_contains "Dockerfile" "gcr.io/distroless" "dockerfile:no-distroless-runtime"
check_regex "Dockerfile" "^ARG GO_BASE_IMAGE_DIGEST=sha256:[0-9a-f]{64}$" "dockerfile:base-digest-complete"

for text in \
  "target: dev" \
  "target: gate" \
  "target: goalcli-runtime" \
  "release-check:" \
  "goalcli-runtime:"; do
  check_contains "docker-compose.yml" "$text" "compose:$text"
done

check_contains ".devcontainer/devcontainer.json" "\"service\": \"dev\"" "devcontainer:service-dev"
check_contains ".devcontainer/devcontainer.json" "\"remoteUser\": \"xlib\"" "devcontainer:remote-user"
for text in ".git" ".worktree" ".omx" ".env" "release/manifest/latest.json" "release/evidence/" "release/docker/" "secrets" "bin" "dist"; do
  check_contains ".dockerignore" "$text" "dockerignore:$text"
done

for target in \
  docker-toolchain-check \
  docker-build \
  docker-build-check \
  docker-shell \
  docker-ci \
  docker-release-check \
  docker-release-final-check \
  docker-goalcli \
  docker-goalcli-image \
  docker-goalcli-version \
  docker-runtime-check \
  docker-drift-check \
  docker-contract; do
  check_target "$target"
  check_contains "cmd/goalcli/main.go" "\"$target\"" "goalcli:$target"
done

for text in \
  "build|docker-build" \
  "build-check" \
  "shell|docker-shell" \
  "goalcli|docker-goalcli" \
  "goalcli-image" \
  "goalcli-version" \
  "runtime-check" \
  "contract" \
  "DOCKER_BASE_IMAGE_DIGEST" \
  "DOCKER_TOOLCHAIN_IMAGE_DIGEST" \
  "DOCKER_RUNTIME_IMAGE" \
  "DOCKER_RUNTIME_IMAGE_DIGEST" \
  "release-final-check" \
  "build-context" \
  "docker inspect"; do
  check_contains "scripts/docker/docker_gate.sh" "$text" "docker-gate:$text"
done

for text in \
  "docker/setup-buildx-action" \
  "make docker-build-check" \
  "make docker-drift-check" \
  "make docker-runtime-check" \
  "make docker-goalcli-version" \
  "make docker-release-final-check" \
  "release/evidence/docker-toolchain-summary.json" \
  "go-version-file: go.mod" \
  "golangci-lint v2.1.6" \
  "govulncheck v1.1.4" \
  "XLIB_ENABLE_VULNCHECK" \
  "actions/upload-artifact"; do
  check_contains ".github/workflows/docker-contract.yml" "$text" "workflow:$text"
done

for text in \
  "docker-contract:" \
  "docker/setup-buildx-action" \
  "make docker-toolchain-check" \
  "make docker-build-check" \
  "make docker-ci" \
  "make docker-release-check" \
  "release/evidence/docker-toolchain-summary.json"; do
  check_contains ".github/workflows/ci.yml" "$text" "ci-workflow:$text"
done

for text in \
  "\"docker\"" \
  "\"base_image_digest\"" \
  "\"runtime_image_digest\"" \
  "\"validated_by\"" \
  "docker-toolchain-check" \
  "docker-build-check" \
  "docker-ci" \
  "docker-release-check" \
  "docker-release-final-check" \
  "docker-goalcli-image" \
  "docker-goalcli-version" \
  "docker-runtime-check" \
  "docker-drift-check" \
  "docker-contract"; do
  check_contains "release/manifest/template.json" "$text" "manifest:$text"
done

for text in \
  docker_toolchain_check \
  docker_build_check \
  docker_ci \
  docker_release_check \
  docker_release_final_check \
  docker_goalcli_image \
  docker_goalcli_version \
  docker_runtime_check \
  docker_drift_check \
  docker_contract; do
  check_contains ".agent/harness/harness.yaml" "$text" "harness:$text"
done

for text in \
  docker-toolchain-check \
  docker-build-check \
  docker-ci \
  docker-release-check \
  docker-release-final-check \
  docker-goalcli-image \
  docker-goalcli-version \
  docker-runtime-check \
  docker-drift-check \
  docker-contract; do
  check_contains ".agent/registries/command-registry.yaml" "$text" "command-registry:$text"
  check_contains ".agent/registries/makefile-target-registry.yaml" "$text" "target-registry:$text"
  check_contains ".agent/registries/makefile-baseline.yaml" "$text" "makefile-baseline:$text"
done

for text in docker-runtime-check goalcli-runtime docker-drift-check docker-contract; do
  check_contains "docs/standard/docker-toolchain-standard.md" "$text" "docker-standard:$text"
done

check_contains "renovate.json" "dockerfile" "renovate:dockerfile"
check_contains ".github/dependabot.yml" "docker" "dependabot:docker"
check_contains "scripts/check_rendered_template.sh" "docker-runtime-check" "render-check:docker-runtime-check"

overall_status="passed"
if [[ "$failures" -gt 0 ]]; then
  overall_status="failed"
fi

{
  echo "# Docker Toolchain Check"
  echo
  echo "- status: $overall_status"
  echo "- mode: $mode"
  echo "- context: ${XLIB_CONTEXT:-unset}"
  echo "- gowork: ${GOWORK:-unset}"
  echo "- docker_cli: $docker_cli"
  echo "- docker_daemon: $docker_daemon"
  echo "- docker_buildx: $docker_buildx"
  echo "- docker_compose: $docker_compose"
  echo
  echo "## Checks"
  echo
  for i in "${!check_names[@]}"; do
    echo "- ${check_statuses[$i]} ${check_names[$i]}: ${check_details[$i]}"
  done
} > "$report"

{
  echo "{"
  printf '  "status": "%s",\n' "$overall_status"
  printf '  "mode": "%s",\n' "$mode"
  printf '  "context": "%s",\n' "$(json_escape "${XLIB_CONTEXT:-unset}")"
  printf '  "gowork": "%s",\n' "$(json_escape "${GOWORK:-unset}")"
  printf '  "docker_cli": "%s",\n' "$(json_escape "$docker_cli")"
  printf '  "docker_daemon": "%s",\n' "$(json_escape "$docker_daemon")"
  printf '  "docker_buildx": "%s",\n' "$(json_escape "$docker_buildx")"
  printf '  "docker_compose": "%s",\n' "$(json_escape "$docker_compose")"
  printf '  "checks": [\n'
  for i in "${!check_names[@]}"; do
    comma=","
    if [[ "$i" == "$((${#check_names[@]} - 1))" ]]; then
      comma=""
    fi
    printf '    {"name": "%s", "status": "%s", "detail": "%s"}%s\n' \
      "$(json_escape "${check_names[$i]}")" \
      "$(json_escape "${check_statuses[$i]}")" \
      "$(json_escape "${check_details[$i]}")" \
      "$comma"
  done
  printf '  ]\n'
  echo "}"
} > "$summary"

if [[ "$overall_status" != "passed" ]]; then
  echo "ERROR: docker toolchain contract has $failures failure(s); wrote $report and $summary" >&2
  exit 1
fi

echo "docker toolchain contract passed; wrote $report and $summary"
