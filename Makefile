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
		echo "golangci-lint not installed, skipping"; \
	fi

.PHONY: integration
integration:
	./scripts/run_integration.sh

.PHONY: security
security:
	@if command -v govulncheck >/dev/null 2>&1; then \
		govulncheck ./...; \
	else \
		echo "govulncheck not installed, skipping"; \
	fi
	./scripts/check_secrets.sh

.PHONY: boundary
boundary:
	./scripts/check_boundary.sh

.PHONY: contracts
contracts:
	./scripts/check_contracts.sh

.PHONY: evidence
evidence:
	./scripts/generate_manifest.sh

.PHONY: ci
ci: fmt vet lint test race boundary security contracts

.PHONY: release-check
release-check: ci integration evidence
