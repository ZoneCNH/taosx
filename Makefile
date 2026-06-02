XLIBGATE ?= go run ./cmd/xlibgate
XLIB_CONTEXT ?= local_write

.PHONY: require-gowork-off
require-gowork-off:
	@if [ "$${GOWORK:-}" != "off" ]; then \
		echo "GOWORK=off is required for release targets"; \
		exit 1; \
	fi

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

.PHONY: security
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
score:
	$(XLIBGATE) score --min 9.8

.PHONY: score-check
score-check:
	# Release evidence verifier default: RELEASE_EVIDENCE_MIN_SCORE=9.5
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
context-lite: require-gowork-off main-guard worktree-guard evidence-check cli-contract command-registry issue-registry makefile-baseline context-profile-check

.PHONY: context-standard
context-standard: context-lite p1-governance-check docs-check

.PHONY: context-full
context-full: context-standard p2-runtime-check

.PHONY: context-release
context-release: require-gowork-off context-standard standard-impact-check score-check
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

.PHONY: agent-team-contract scope-lock pr-template acceptance-matrix runtime-health upgrade-standard conformance-profile downstream-registry self-healing-skeleton goal-runtime github-governance supply-chain changelog governance-fixture-test autoresearch policy-schema github-settings toolchain evidence-artifacts naming
agent-team-contract scope-lock pr-template acceptance-matrix runtime-health upgrade-standard conformance-profile downstream-registry self-healing-skeleton goal-runtime github-governance supply-chain changelog governance-fixture-test autoresearch policy-schema github-settings toolchain evidence-artifacts naming:
	$(XLIBGATE) $@ --dry-run --verify

.PHONY: install-runtime upgrade-runtime release-ready evidence-replay attest-conformance pack-standard pack-gate pack-evidence downstream-baseline downstream-adoption runtime-file-ownership
install-runtime upgrade-runtime release-ready evidence-replay attest-conformance pack-standard pack-gate pack-evidence downstream-baseline downstream-adoption runtime-file-ownership:
	$(XLIBGATE) $@ --dry-run --verify

.PHONY: execution-context
execution-context:
	$(XLIBGATE) $@ --dry-run --verify

.PHONY: governance-check
governance-check: require-gowork-off main-guard worktree-guard evidence-check boundary security contracts docs-check cli-contract issue-registry command-registry makefile-baseline

.PHONY: p1-governance-check
p1-governance-check: agent-team-contract scope-lock pr-template acceptance-matrix runtime-health upgrade-standard conformance-profile downstream-registry self-healing-skeleton goal-runtime github-governance supply-chain changelog governance-fixture-test autoresearch policy-schema github-settings toolchain evidence-artifacts naming

.PHONY: p2-runtime-check
p2-runtime-check: install-runtime upgrade-runtime release-ready evidence-replay attest-conformance pack-standard pack-gate pack-evidence downstream-baseline downstream-adoption runtime-file-ownership execution-context

.PHONY: context-profile
context-profile:
	$(XLIBGATE) context-profile

.PHONY: context-profile-check
context-profile-check:
	$(XLIBGATE) context-profile-check

.PHONY: context-schema-check
context-schema-check:
	$(XLIBGATE) context-schema-check

.PHONY: context-lite
context-lite: require-gowork-off governance-check context-profile-check

.PHONY: context-standard
context-standard: require-gowork-off governance-check p1-governance-check context-profile-check

.PHONY: context-full
context-full: require-gowork-off governance-check p1-governance-check p2-runtime-check context-profile-check

.PHONY: context-release
context-release: require-gowork-off context-full integration dependency-check standard-impact-check score-check
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
ci: fmt vet lint test race boundary security contracts governance-check score

.PHONY: ci-extended
ci-extended: ci property golden fuzz-smoke

.PHONY: release-check
release-check: require-gowork-off ci integration dependency-check standard-impact-check docs-check score-check governance-check p1-governance-check p2-runtime-check
	CHECK_STATUS=passed $(MAKE) evidence
	$(MAKE) release-evidence-hash
	$(MAKE) release-evidence-check
	$(MAKE) release-evidence-checksum-check

.PHONY: release-check-extended
release-check-extended: require-gowork-off ci-extended integration dependency-check standard-impact-check docs-check score-check governance-check p1-governance-check p2-runtime-check
	CHECK_STATUS=passed $(MAKE) evidence
	$(MAKE) release-evidence-hash
	$(MAKE) release-evidence-check
	$(MAKE) release-evidence-checksum-check

.PHONY: release-final-check
release-final-check:
	XLIB_CONTEXT=release_verify GOWORK=off $(MAKE) context-release
	$(XLIBGATE) score --min 9.8
	RELEASE_EVIDENCE_REQUIRE_PASSED=1 RELEASE_EVIDENCE_REQUIRE_CLEAN=1 RELEASE_EVIDENCE_MIN_SCORE=9.5 ./scripts/check_release_evidence.sh
	$(MAKE) release-evidence-checksum-check

.PHONY: release-preflight
release-preflight:
	./scripts/check_release_preflight.sh "$(VERSION)"
	GOWORK=off XLIB_CONTEXT=release_verify VERSION="$(VERSION)" $(MAKE) release-final-check
