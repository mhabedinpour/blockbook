.EXPORT_ALL_VARIABLES:
ENV=val
################################
# Dependency related commands
################################

.PHONY: dependency
dependency: install-golangci-lint install-ganache-cli

.PHONY: install-golangci-lint
install-golangci-lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.59.0

.PHONY: install-ganache-cli
install-ganache-cli:
	npm install -g ganache-cli

################################
# CI related commands
################################

.PHONY: fix-imports
fix-imports:
	@goimports -w ./cmd
	@goimports -w ./internal
	@goimports -w ./pkg

.PHONY: lint
lint:
	./scripts/lint.sh

.PHONY: test
test:
	./scripts/test.sh

.PHONY: install-hooks
install-hooks:
	cp .hooks/pre-commit .git/hooks/pre-commit
	sudo chmod +x .git/hooks/pre-commit

.PHONY: uninstall-hooks
uninstall-hooks:
	rm .git/hooks/pre-commit
