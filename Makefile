XLIBGATE ?= go run ./cmd/xlibgate

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
	$(XLIBGATE) secrets

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

.PHONY: score-check
score-check:
	go run ./cmd/xlibgate score --min 9.8

.PHONY: release-evidence-hash
release-evidence-hash:
	$(XLIBGATE) release-evidence-hash >/dev/null

.PHONY: release-evidence-check
release-evidence-check:
	RELEASE_EVIDENCE_REQUIRE_PASSED=1 $(XLIBGATE) release-evidence-check

.PHONY: release-evidence-checksum-check
release-evidence-checksum-check:
	$(XLIBGATE) release-evidence-checksum-check

.PHONY: require-gowork-off
require-gowork-off:
	@if [ "$(GOWORK)" != "off" ]; then \
		echo "GOWORK=off is required for release targets"; \
		exit 1; \
	fi

.PHONY: score
score:
	# Direct equivalent for docs/CI drift checks: go run ./cmd/xlibgate score --min 9.8
	$(XLIBGATE) score --min 9.8

.PHONY: score-check
score-check:
	# Release evidence verifier default: RELEASE_EVIDENCE_MIN_SCORE=9.5
	$(XLIBGATE) score --min 9.5

.PHONY: ci
ci: fmt vet lint test race boundary security contracts score

.PHONY: ci-extended
ci-extended: ci property golden fuzz-smoke

.PHONY: release-check
release-check: require-gowork-off ci integration dependency-check standard-impact-check docs-check score-check
	CHECK_STATUS=passed $(MAKE) evidence
	$(MAKE) release-evidence-hash
	$(MAKE) release-evidence-check
	$(MAKE) release-evidence-checksum-check

.PHONY: release-check-extended
release-check-extended: require-gowork-off ci-extended integration dependency-check standard-impact-check docs-check score-check
	CHECK_STATUS=passed $(MAKE) evidence
	$(MAKE) release-evidence-hash
	$(MAKE) release-evidence-check
	$(MAKE) release-evidence-checksum-check

.PHONY: release-final-check
release-final-check: release-check
	go run ./cmd/xlibgate score --min 9.5
	RELEASE_EVIDENCE_REQUIRE_PASSED=1 RELEASE_EVIDENCE_REQUIRE_CLEAN=1 RELEASE_EVIDENCE_MIN_SCORE=9.5 ./scripts/check_release_evidence.sh
	$(MAKE) release-evidence-checksum-check

.PHONY: release-preflight
release-preflight:
	./scripts/check_release_preflight.sh "$(VERSION)"
	GOWORK=off VERSION="$(VERSION)" $(MAKE) release-final-check
