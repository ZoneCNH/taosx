#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "Usage: scripts/docker/docker_gate.sh <build|build-check|shell|ci|release-check|release-final-check|goalcli|goalcli-image|goalcli-version|runtime-check|drift-check|contract> [extra args...]" >&2
  exit 2
fi

command="$1"
shift

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
image="${DOCKER_IMAGE:-$(basename "$repo_root")-toolchain:local}"
runtime_image="${DOCKER_RUNTIME_IMAGE:-$(basename "$repo_root")-goalcli-runtime:local}"
go_version="${GO_VERSION:-1.23}"
go_base_image="${GO_BASE_IMAGE:-golang:${go_version}-bookworm}"
go_base_image_digest="${GO_BASE_IMAGE_DIGEST:-sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855}"
golangci_lint_version="${GOLANGCI_LINT_VERSION:-v2.1.6}"
govulncheck_version="${GOVULNCHECK_VERSION:-v1.1.4}"
toolchain_image_digest="${DOCKER_TOOLCHAIN_IMAGE_DIGEST:-$go_base_image_digest}"
runtime_image_digest="${DOCKER_RUNTIME_IMAGE_DIGEST:-$go_base_image_digest}"

export DOCKER_BUILDKIT=1

require_docker() {
  if ! command -v docker >/dev/null 2>&1; then
    echo "ERROR: docker CLI not found" >&2
    exit 1
  fi
  if ! docker version >/dev/null 2>&1; then
    echo "ERROR: docker daemon unavailable" >&2
    exit 1
  fi
  if ! docker buildx version >/dev/null 2>&1; then
    echo "ERROR: docker buildx unavailable" >&2
    exit 1
  fi
}

build_target() {
  local target="$1"
  local tag="$2"
  require_docker
  local build_context_args=()
  if [[ -d "$repo_root/.cache/docker-tools" ]] \
    && [[ -f "$repo_root/.cache/docker-tools/usr/local/bin/golangci-lint" ]] \
    && [[ -f "$repo_root/.cache/docker-tools/usr/local/bin/govulncheck" ]]; then
    build_context_args+=(--build-context "tools=$repo_root/.cache/docker-tools")
  fi

  docker buildx build \
    --load \
    --target "$target" \
    --build-arg "GO_VERSION=$go_version" \
    --build-arg "GO_BASE_IMAGE=$go_base_image" \
    --build-arg "GO_BASE_IMAGE_DIGEST=$go_base_image_digest" \
    --build-arg "GOLANGCI_LINT_VERSION=$golangci_lint_version" \
    --build-arg "GOVULNCHECK_VERSION=$govulncheck_version" \
    "${build_context_args[@]}" \
    --tag "$tag" \
    "$repo_root"

  # 构建后通过 docker inspect 获取真实 digest，替换占位符默认值
  local real_digest=""
  real_digest="$(docker inspect --format='{{index .RepoDigests 0}}' "$tag" 2>/dev/null | cut -d@ -f2 || true)"
  if [[ -z "$real_digest" ]]; then
    # 本地构建的镜像可能没有 RepoDigests，使用镜像 ID 作为 fallback
    real_digest="$(docker inspect --format='{{.Id}}' "$tag" 2>/dev/null || true)"
  fi
  if [[ -n "$real_digest" ]]; then
    case "$target" in
      toolchain)
        export DOCKER_TOOLCHAIN_IMAGE_DIGEST="$real_digest"
        ;;
      goalcli-runtime)
        export DOCKER_RUNTIME_IMAGE_DIGEST="$real_digest"
        ;;
    esac
  fi
}

run_make() {
  local make_target="$1"
  shift || true
  build_target toolchain "$image"
  # 使用 build_target 导出的真实 digest（如果可用），否则回退到占位符默认值
  docker run --rm \
    --workdir /workspace \
    --volume "$repo_root:/workspace" \
    --volume go-build-cache:/root/.cache/go-build \
    --volume go-mod-cache:/go/pkg/mod \
    --env "CI=1" \
    --env "GOWORK=${GOWORK:-off}" \
    --env "XLIB_CONTEXT=${XLIB_CONTEXT:-ci_pull_request}" \
    --env "VERSION=${VERSION:-}" \
    --env "DOWNSTREAM=${DOWNSTREAM:-}" \
    --env "XLIB_ENABLE_VULNCHECK=${XLIB_ENABLE_VULNCHECK:-}" \
    --env "CHECK_STATUS=${CHECK_STATUS:-}" \
    --env "DOCKER_TOOLCHAIN_ENABLED=${DOCKER_TOOLCHAIN_ENABLED:-true}" \
    --env "DOCKER_BASE_IMAGE=${DOCKER_BASE_IMAGE:-$go_base_image}" \
    --env "DOCKER_BASE_IMAGE_DIGEST=${DOCKER_BASE_IMAGE_DIGEST:-$go_base_image_digest}" \
    --env "DOCKER_TOOLCHAIN_IMAGE=${DOCKER_TOOLCHAIN_IMAGE:-$image}" \
    --env "DOCKER_TOOLCHAIN_IMAGE_DIGEST=${DOCKER_TOOLCHAIN_IMAGE_DIGEST:-$toolchain_image_digest}" \
    --env "DOCKER_RUNTIME_IMAGE=${DOCKER_RUNTIME_IMAGE:-$runtime_image}" \
    --env "DOCKER_RUNTIME_IMAGE_DIGEST=${DOCKER_RUNTIME_IMAGE_DIGEST:-$runtime_image_digest}" \
    "$image" make "$make_target" "$@"
}

run_shell() {
  build_target dev "$image"
  local docker_flags=(--rm)
  if [[ -t 0 && -t 1 ]]; then
    docker_flags+=(-it)
  fi
  docker run "${docker_flags[@]}" \
    --workdir /workspace \
    --volume "$repo_root:/workspace" \
    --volume go-build-cache:/home/xlib/.cache/go-build \
    --volume go-mod-cache:/go/pkg/mod \
    --env "GOWORK=${GOWORK:-off}" \
    --env "XLIB_CONTEXT=${XLIB_CONTEXT:-ci_pull_request}" \
    --env "VERSION=${VERSION:-}" \
    --env "DOWNSTREAM=${DOWNSTREAM:-}" \
    --env "XLIB_ENABLE_VULNCHECK=${XLIB_ENABLE_VULNCHECK:-}" \
    --env "DOCKER_TOOLCHAIN_ENABLED=${DOCKER_TOOLCHAIN_ENABLED:-true}" \
    --env "DOCKER_BASE_IMAGE=${DOCKER_BASE_IMAGE:-$go_base_image}" \
    --env "DOCKER_BASE_IMAGE_DIGEST=${DOCKER_BASE_IMAGE_DIGEST:-$go_base_image_digest}" \
    --env "DOCKER_TOOLCHAIN_IMAGE=${DOCKER_TOOLCHAIN_IMAGE:-$image}" \
    --env "DOCKER_TOOLCHAIN_IMAGE_DIGEST=${DOCKER_TOOLCHAIN_IMAGE_DIGEST:-$toolchain_image_digest}" \
    --env "DOCKER_RUNTIME_IMAGE=${DOCKER_RUNTIME_IMAGE:-$runtime_image}" \
    --env "DOCKER_RUNTIME_IMAGE_DIGEST=${DOCKER_RUNTIME_IMAGE_DIGEST:-$runtime_image_digest}" \
    "$image" bash "$@"
}

run_goalcli_runtime() {
  build_target goalcli-runtime "$runtime_image"
  docker run --rm \
    --workdir /workspace \
    --volume "$repo_root:/workspace" \
    --env "GOWORK=${GOWORK:-off}" \
    --env "XLIB_CONTEXT=${XLIB_CONTEXT:-docker_runtime}" \
    --env "DOCKER_BASE_IMAGE=${DOCKER_BASE_IMAGE:-$go_base_image}" \
    --env "DOCKER_BASE_IMAGE_DIGEST=${DOCKER_BASE_IMAGE_DIGEST:-$go_base_image_digest}" \
    --env "DOCKER_TOOLCHAIN_IMAGE=${DOCKER_TOOLCHAIN_IMAGE:-$image}" \
    --env "DOCKER_TOOLCHAIN_IMAGE_DIGEST=${DOCKER_TOOLCHAIN_IMAGE_DIGEST:-$toolchain_image_digest}" \
    --env "DOCKER_RUNTIME_IMAGE=${DOCKER_RUNTIME_IMAGE:-$runtime_image}" \
    --env "DOCKER_RUNTIME_IMAGE_DIGEST=${DOCKER_RUNTIME_IMAGE_DIGEST:-$runtime_image_digest}" \
    "$runtime_image" "$@"
}

"$repo_root/scripts/docker/check_toolchain.sh"

case "$command" in
  build|docker-build)
    build_target toolchain "$image"
    ;;
  build-check|docker-build-check)
    build_target toolchain "$image"
    build_target gate "$image-gate"
    build_target goalcli-runtime "$runtime_image"
    ;;
  shell|docker-shell)
    run_shell "$@"
    ;;
  ci|docker-ci)
    run_make ci "$@"
    ;;
  release-check|docker-release-check)
    XLIB_CONTEXT="${XLIB_CONTEXT:-release_verify}" run_make release-check "$@"
    ;;
  release-final-check|docker-release-final-check)
    XLIB_CONTEXT="${XLIB_CONTEXT:-release_verify}" run_make release-final-check "$@"
    ;;
  goalcli|docker-goalcli)
    if [[ "$#" -eq 0 ]]; then
      run_goalcli_runtime version
    else
      run_goalcli_runtime "$@"
    fi
    ;;
  goalcli-image|docker-goalcli-image)
    build_target goalcli-runtime "$runtime_image"
    ;;
  goalcli-version|docker-goalcli-version)
    run_goalcli_runtime version "$@"
    ;;
  runtime-check|docker-runtime-check)
    run_goalcli_runtime doctor --json
    ;;
  drift-check|docker-drift-check)
    "$repo_root/scripts/docker/check_toolchain.sh" --drift
    ;;
  contract|docker-contract)
    "$repo_root/scripts/docker/check_toolchain.sh" --drift
    build_target toolchain "$image"
    build_target gate "$image-gate"
    run_goalcli_runtime version
    ;;
  *)
    echo "ERROR: unknown docker gate command $command" >&2
    exit 2
    ;;
esac
