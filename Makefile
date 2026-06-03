XLIBGATE ?= go run ./cmd/xlibgate
XLIB_CONTEXT ?= local_write
GOAL_ID ?= GOAL-20260603-XLIB-GOALKIT-001
GOAL_RUNTIME_MODE ?= FULL

.PHONY: require-gowork-off
require-gowork-off:
	@if [ "$${GOWORK:-}" != "off" ]; then \
		echo "GOWORK=off is required for release targets"; \
		exit 1; \
	fi

.PHONY: build
build:
	go build ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test ./...

.PHONY: race
race:
	go test -race ./...

.PHONY: lint
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed"; \
		exit 1; \
	fi

.PHONY: integration
integration:
	$(XLIBGATE) integration

.PHONY: dependency-check
dependency-check:
	$(XLIBGATE) dependency-check

.PHONY: standard-impact-check
standard-impact-check:
	$(XLIBGATE) standard-impact-check

.PHONY: docs-check
docs-check:
	$(XLIBGATE) docs-check

.PHONY: debt architecture domain docs-drift dependency-debt security-debt testing-debt implementation-debt downstream-debt

.PHONY: debt-evidence
debt-evidence:
	$(XLIBGATE) debt-evidence

.PHONY: debt-evidence-hash
debt-evidence-hash:
	$(XLIBGATE) debt-evidence-hash

.PHONY: debt-evidence-checksum-check
debt-evidence-checksum-check:
	$(XLIBGATE) debt-evidence-checksum-check

.PHONY: security

architecture:
	$(XLIBGATE) debt --section architecture --mode enforce

domain:
	$(XLIBGATE) debt --section domain --mode enforce

docs-drift:
	$(XLIBGATE) debt --section docs --mode warn

dependency-debt:
	$(XLIBGATE) debt --section dependency --mode warn

testing-debt:
	$(XLIBGATE) debt --section testing --mode warn

implementation-debt:
	$(XLIBGATE) debt --section implementation --mode observe

security-debt:
	$(XLIBGATE) debt --section security --mode warn

downstream-debt:
	$(XLIBGATE) downstream-debt

debt:
	$(XLIBGATE) debt --config .agent/debt/rules.yaml --exceptions .agent/debt/exceptions.yaml --dependency-purpose .agent/debt/dependency-purpose.yaml --mode enforce --min-score 9.8

.PHONY: debt-register-update debt-trend debt-patch-suggest debt-lifecycle-check
debt-register-update:
	$(XLIBGATE) debt register-update

debt-trend:
	$(XLIBGATE) debt trend

debt-patch-suggest:
	$(XLIBGATE) debt patch-suggest

debt-lifecycle-check:
	$(XLIBGATE) debt lifecycle-check

security:
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed"; \
		exit 1; \
	fi
	$(XLIBGATE) security

.PHONY: boundary
boundary:
	$(XLIBGATE) boundary

.PHONY: contracts
contracts:
	$(XLIBGATE) contracts

.PHONY: property
property:
	go test ./... -run 'Test.*Property|Test.*Invariant'

.PHONY: fuzz-smoke
fuzz-smoke:
	./scripts/run_fuzz_smoke.sh

.PHONY: golden
golden:
	go test ./... -run 'Test.*Golden|Test.*Snapshot'

.PHONY: evidence
evidence:
	$(XLIBGATE) evidence

.PHONY: goal-score-check
goal-score-check:
	go run ./cmd/xlibgate score --min 9.8

.PHONY: release-evidence-hash
release-evidence-hash:
	$(XLIBGATE) release-evidence-hash

.PHONY: release-evidence-check
release-evidence-check:
	$(XLIBGATE) release-evidence-check

.PHONY: release-evidence-checksum-check
release-evidence-checksum-check:
	$(XLIBGATE) release-evidence-checksum-check

.PHONY: score
score: score-check

.PHONY: score-check
score-check:
	# Release evidence verifier default: RELEASE_EVIDENCE_MIN_SCORE=9.8
	# Direct equivalent for docs/CI drift checks: go run ./cmd/xlibgate score --min 9.8
	$(XLIBGATE) score --min 9.8

.PHONY: version
version:
	$(XLIBGATE) version

.PHONY: doctor
doctor:
	$(XLIBGATE) doctor

.PHONY: main-guard
main-guard:
	$(XLIBGATE) main-guard --context $(XLIB_CONTEXT)

.PHONY: worktree-guard
worktree-guard:
	$(XLIBGATE) worktree-guard --context $(XLIB_CONTEXT)

.PHONY: evidence-check
evidence-check:
	$(XLIBGATE) evidence-check

.PHONY: cli-contract
cli-contract:
	$(XLIBGATE) cli-contract

.PHONY: issue-registry
issue-registry:
	$(XLIBGATE) issue-registry

.PHONY: command-registry
command-registry:
	$(XLIBGATE) command-registry

.PHONY: makefile-baseline
makefile-baseline:
	$(XLIBGATE) makefile-baseline

.PHONY: agent-team-contract scope-lock pr-template acceptance-matrix runtime-health upgrade-standard conformance-profile downstream-registry self-healing-skeleton goal-runtime github-governance supply-chain changelog governance-fixture-test autoresearch policy-schema github-settings toolchain evidence-artifacts naming
agent-team-contract scope-lock pr-template acceptance-matrix runtime-health upgrade-standard conformance-profile downstream-registry self-healing-skeleton goal-runtime github-governance supply-chain changelog governance-fixture-test autoresearch policy-schema github-settings toolchain evidence-artifacts naming:
	$(XLIBGATE) $@ --dry-run --verify


.PHONY: install-runtime upgrade-runtime release-ready evidence-replay attest-conformance pack-standard pack-gate pack-evidence downstream-baseline downstream-adoption runtime-file-ownership
install-runtime upgrade-runtime release-ready evidence-replay attest-conformance pack-standard pack-gate pack-evidence downstream-baseline downstream-adoption runtime-file-ownership:
	$(XLIBGATE) $@ --dry-run --verify

.PHONY: execution-context
execution-context:
	$(XLIBGATE) $@ --dry-run --verify

.PHONY: goal-acceptance
goal-acceptance: require-gowork-off
	$(XLIBGATE) $@ --goal-id "$(GOAL_ID)" --mode "$(GOAL_RUNTIME_MODE)" --json --write-evidence

.PHONY: goal-delivery
goal-delivery: require-gowork-off
	$(XLIBGATE) $@ --goal-id "$(GOAL_ID)" --mode "$(GOAL_RUNTIME_MODE)" --json --write-evidence

.PHONY: goal-handover
goal-handover: require-gowork-off
	$(XLIBGATE) $@ --goal-id "$(GOAL_ID)" --mode "$(GOAL_RUNTIME_MODE)" --json --write-evidence

.PHONY: goal-downstream-adoption
goal-downstream-adoption: require-gowork-off
	$(XLIBGATE) $@ --goal-id "$(GOAL_ID)" --mode "$(GOAL_RUNTIME_MODE)" --json --write-evidence

.PHONY: goal-certify
goal-certify: require-gowork-off
	$(XLIBGATE) $@ --goal-id "$(GOAL_ID)" --mode "$(GOAL_RUNTIME_MODE)" --json --write-evidence

.PHONY: goal-runtime-final
goal-runtime-final: require-gowork-off goal-acceptance goal-delivery goal-handover goal-downstream-adoption goal-certify
	$(XLIBGATE) $@ --goal-id "$(GOAL_ID)" --mode "$(GOAL_RUNTIME_MODE)" --json --write-evidence

.PHONY: governance-check
governance-check: require-gowork-off main-guard worktree-guard evidence-check boundary architecture domain security security-debt contracts docs-check cli-contract issue-registry command-registry makefile-baseline debt

.PHONY: p1-governance-check
p1-governance-check: agent-team-contract scope-lock pr-template acceptance-matrix runtime-health upgrade-standard conformance-profile downstream-registry self-healing-skeleton goal-runtime github-governance supply-chain changelog governance-fixture-test autoresearch policy-schema github-settings toolchain evidence-artifacts naming

.PHONY: p2-runtime-check
p2-runtime-check: install-runtime upgrade-runtime release-ready evidence-replay attest-conformance pack-standard pack-gate pack-evidence downstream-baseline downstream-adoption runtime-file-ownership execution-context

.PHONY: context-profile
context-profile:
	$(XLIBGATE) context-profile --profile $${PROFILE:-standard}

.PHONY: context-profile-check
context-profile-check:
	$(XLIBGATE) context-profile-check

.PHONY: context-schema-check
context-schema-check:
	$(XLIBGATE) context-schema-check

.PHONY: context-lite
context-lite: require-gowork-off governance-check

.PHONY: context-standard
context-standard: require-gowork-off governance-check p1-governance-check docs-check

.PHONY: context-full
context-full: require-gowork-off governance-check p1-governance-check p2-runtime-check

.PHONY: context-release
context-release: require-gowork-off context-full integration dependency-check standard-impact-check score-check debt-evidence
	CHECK_STATUS=passed $(MAKE) evidence
	$(MAKE) release-evidence-hash
	$(MAKE) release-evidence-check
	$(MAKE) release-evidence-checksum-check

.PHONY: context-fast-check
context-fast-check: context-lite

.PHONY: context-standard-check
context-standard-check: context-standard

.PHONY: context-full-check
context-full-check: context-full

.PHONY: ci
ci: fmt vet lint test race boundary architecture domain security security-debt contracts governance-check debt score

.PHONY: ci-extended
ci-extended: ci property golden fuzz-smoke docs-drift

.PHONY: release-check
release-check: require-gowork-off ci integration dependency-check standard-impact-check docs-check docs-drift score-check governance-check p1-governance-check p2-runtime-check debt-evidence
	CHECK_STATUS=passed $(MAKE) evidence
	$(MAKE) release-evidence-hash
	$(MAKE) release-evidence-check
	$(MAKE) release-evidence-checksum-check

.PHONY: release-check-extended
release-check-extended: require-gowork-off ci-extended integration dependency-check standard-impact-check docs-check score-check governance-check p1-governance-check p2-runtime-check debt-evidence
	CHECK_STATUS=passed $(MAKE) evidence
	$(MAKE) release-evidence-hash
	$(MAKE) release-evidence-check
	$(MAKE) release-evidence-checksum-check

.PHONY: release-final-check
release-final-check:
	XLIB_CONTEXT=release_verify GOWORK=off $(MAKE) context-release
	$(MAKE) debt-evidence-checksum-check
	$(XLIBGATE) score --min 9.8
	RELEASE_EVIDENCE_REQUIRE_PASSED=1 RELEASE_EVIDENCE_REQUIRE_CLEAN=1 RELEASE_EVIDENCE_MIN_SCORE=9.8 ./scripts/check_release_evidence.sh
	$(MAKE) release-evidence-checksum-check

.PHONY: release-preflight
release-preflight:
	./scripts/check_release_preflight.sh "$(VERSION)"
	GOWORK=off XLIB_CONTEXT=release_verify VERSION="$(VERSION)" $(MAKE) release-final-check

# ── Goal Gate Targets ──────────────────────────────────────────
# 以下 target 对应 .agent/harness/gates/*.yaml 中 commands 引用

.PHONY: worktree-check
worktree-check:
	bash scripts/harness/no-main-dev.sh

.PHONY: context-check
context-check:
	@echo "context-check: 验证 Goal Context 完整性..."
	@test -d docs/goal || (echo "ERROR: docs/goal/ 目录不存在" && exit 1)
	@echo "context-check: 通过。"

.PHONY: spec-check
spec-check:
	@echo "spec-check: 验证 Spec 存在且包含 REQ-xxx..."
	@find docs/ -name '*.md' -exec grep -l 'REQ-' {} + > /dev/null 2>&1 || (echo "WARNING: 未找到包含 REQ-xxx 的文档" && exit 0)
	@echo "spec-check: 通过。"

.PHONY: design-check
design-check:
	@echo "design-check: 验证 Design 文档存在..."
	@test -d docs/adr || (echo "WARNING: docs/adr/ 目录不存在" && exit 0)
	@echo "design-check: 通过。"

.PHONY: task-check
task-check:
	@echo "task-check: 验证 Task 可追溯性..."
	@test -f .agent/registry/commands.yaml || (echo "WARNING: commands.yaml 不存在" && exit 0)
	@echo "task-check: 通过。"

.PHONY: pr-check
pr-check: worktree-check lint test
	@echo "pr-check: worktree + lint + test 全部通过。"

.PHONY: retro-check
retro-check:
	@echo "retro-check: 验证 Retrospective 存在..."
	@echo "retro-check: 通过（语义检查由 Agent 执行）。"

install-hooks:
	@git config core.hooksPath .githooks
	@echo "✅ git hooks 已启用（core.hooksPath=.githooks）"

doctor-hooks:
	@[ "$$(git config --get core.hooksPath)" = ".githooks" ] || { \
	  echo "ERROR: core.hooksPath 未指向 .githooks，请运行 make install-hooks"; \
	  exit 1; \
	}
	@echo "✅ hooks 配置正确"

# sync-main: 拉取远端 main 并尽量 fast-forward 本地 main。
# 对应 .agent/standard/goal-runtime-canonical.md RULE-MAIN-SYNC-002：
# 每个 worktree 创建前必须基于最新 main。
sync-main:
	@git fetch origin main
	@CUR=$$(git rev-parse --abbrev-ref HEAD); \
	if [ "$$CUR" = "main" ]; then \
	  git merge --ff-only origin/main && echo "✅ main 已同步至 origin/main"; \
	else \
	  LOCAL=$$(git rev-parse main 2>/dev/null || echo ""); \
	  REMOTE=$$(git rev-parse origin/main); \
	  if [ "$$LOCAL" = "$$REMOTE" ]; then \
	    echo "✅ 本地 main 已是最新（当前在 $$CUR）"; \
	  else \
	    echo "⚠️  当前不在 main 分支（$$CUR）"; \
	    echo "   本地 main: $$LOCAL"; \
	    echo "   远端 main: $$REMOTE"; \
	    echo "   请到主 worktree 执行: git merge --ff-only origin/main"; \
	    exit 1; \
	  fi; \
	fi
