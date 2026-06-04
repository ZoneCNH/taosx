# syntax=docker/dockerfile:1.7

ARG GO_VERSION=1.23
ARG GO_BASE_IMAGE=golang:${GO_VERSION}-bookworm
ARG GO_BASE_IMAGE_DIGEST=sha256:e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855
ARG GOLANGCI_LINT_VERSION=v2.1.6
ARG GOVULNCHECK_VERSION=v1.1.4

# ── 预编译工具 stage：下载 golangci-lint 和 govulncheck ─────
# 若通过 --build-context tools=... 提供外部缓存则跳过下载
FROM ${GO_BASE_IMAGE} AS tools
ARG GOLANGCI_LINT_VERSION
ARG GOVULNCHECK_VERSION
ENV GOTOOLCHAIN=auto
RUN GOLANGCI_LINT_SEMVER="${GOLANGCI_LINT_VERSION#v}" \
    && curl -sSfL "https://github.com/golangci/golangci-lint/releases/download/${GOLANGCI_LINT_VERSION}/golangci-lint-${GOLANGCI_LINT_SEMVER}-linux-amd64.tar.gz" \
      | tar -xz --strip-components=1 -C /usr/local/bin "golangci-lint-${GOLANGCI_LINT_SEMVER}-linux-amd64/golangci-lint" \
    && go install golang.org/x/vuln/cmd/govulncheck@${GOVULNCHECK_VERSION} \
    && cp "$(go env GOPATH)/bin/govulncheck" /usr/local/bin/govulncheck

FROM ${GO_BASE_IMAGE} AS toolchain

ARG GO_BASE_IMAGE
ARG GO_BASE_IMAGE_DIGEST
ARG GOLANGCI_LINT_VERSION
ARG GOVULNCHECK_VERSION

LABEL org.opencontainers.image.base.name="${GO_BASE_IMAGE}"
LABEL org.opencontainers.image.base.digest="${GO_BASE_IMAGE_DIGEST}"

ENV CGO_ENABLED=1 \
    GOWORK=off \
    XLIB_CONTEXT=docker_toolchain \
    GOPATH=/go \
    GOCACHE=/root/.cache/go-build \
    GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct \
    GOLANGCI_LINT_VERSION=${GOLANGCI_LINT_VERSION} \
    GOVULNCHECK_VERSION=${GOVULNCHECK_VERSION}

RUN --mount=type=cache,target=/var/cache/apt \
    --mount=type=cache,target=/var/lib/apt/lists \
    apt-get update \
    && apt-get install -y --no-install-recommends \
      bash \
      ca-certificates \
      curl \
      git \
      jq \
      make \
      openssh-client \
      python3 \
      python3-yaml \
      tar \
      xz-utils \
    && git config --system --add safe.directory '*'

# ── 预编译工具：从本地缓存目录复制（零下载）─────────────────
# 构建前先运行: make docker-prefetch
# 然后: docker buildx build --build-context tools=.cache/docker-tools .
COPY --from=tools /usr/local/bin/golangci-lint /usr/local/bin/golangci-lint
COPY --from=tools /usr/local/bin/govulncheck  /usr/local/bin/govulncheck

WORKDIR /workspace

CMD ["bash"]

FROM toolchain AS dev

ARG USERNAME=xlib
ARG USER_UID=1000
ARG USER_GID=1000

RUN groupadd --gid ${USER_GID} ${USERNAME} \
    && useradd --uid ${USER_UID} --gid ${USER_GID} -m ${USERNAME} \
    && mkdir -p /workspace /home/${USERNAME}/.cache/go-build /go/pkg/mod \
    && chown -R ${USERNAME}:${USERNAME} /workspace /home/${USERNAME} /go

USER ${USERNAME}
ENV XLIB_CONTEXT=docker_toolchain \
    GOCACHE=/home/${USERNAME}/.cache/go-build

CMD ["bash"]

FROM toolchain AS gate

COPY . /workspace

CMD ["make", "ci"]

FROM toolchain AS build-goalcli

COPY . /workspace
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/goalcli ./cmd/goalcli

FROM scratch AS goalcli-runtime

COPY --from=build-goalcli /out/goalcli /goalcli

ENTRYPOINT ["/goalcli"]
CMD ["version"]
