.PHONY: help fmt lint build scaffold test clean vet run install-tools validate commit-lint hooks

# Colors
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
CYAN := \033[0;36m
BOLD := \033[1m
RESET := \033[0m

# Variables
BINARY_NAME=scaffold
BIN_DIR=bin
GO=go
GOFLAGS=-v

# Helper functions
define print_header
	@echo "$(CYAN)╔════════════════════════════════════════════════════════════╗$(RESET)"
	@echo "$(CYAN)║ $(BOLD)$(1)$(RESET)$(CYAN)$(2)║$(RESET)"
	@echo "$(CYAN)╚════════════════════════════════════════════════════════════╝$(RESET)"
endef

define print_success
	@echo "$(GREEN)✓ $(1)$(RESET)"
endef

define print_info
	@echo "$(BLUE)→ $(1)$(RESET)"
endef

define print_warning
	@echo "$(YELLOW)⚠ $(1)$(RESET)"
endef

help: ## Display this help screen
	@echo ""
	@echo "$(BOLD)$(CYAN)Scaffold CLI - Make Targets$(RESET)"
	@echo "$(CYAN)════════════════════════════════════════════$(RESET)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(CYAN)%-18s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(YELLOW)Examples:$(RESET)"
	@echo "  make scaffold    # Build and run the CLI"
	@echo "  make validate    # Run all checks (fmt → vet → lint → test)"
	@echo "  make clean       # Remove build artifacts"
	@echo ""

fmt: ## Format Go code
	$(call print_header,"FMT","                              ")
	$(call print_info,"Formatting code...")
	@$(GO) fmt ./...
	$(call print_success,"Code formatted")
	@echo ""

lint: ## Lint Go code with golangci-lint
	$(call print_header,"LINT","                             ")
	$(call print_info,"Running linter...")
	@golangci-lint run ./... || (echo "$(RED)✗ Linting failed$(RESET)"; exit 1)
	$(call print_success,"Linting passed")
	@echo ""

vet: ## Run go vet to check for suspicious constructs
	$(call print_header,"VET","                              ")
	$(call print_info,"Checking for suspicious constructs...")
	@$(GO) vet ./... || (echo "$(RED)✗ Vet check failed$(RESET)"; exit 1)
	$(call print_success,"Vet check passed")
	@echo ""

test: ## Run all tests with race detector
	$(call print_header,"TEST","                             ")
	$(call print_info,"Running tests...")
	@$(GO) test -race -v ./... || (echo "$(RED)✗ Tests failed$(RESET)"; exit 1)
	$(call print_success,"All tests passed")
	@echo ""

build: fmt vet ## Build the scaffold binary into bin/ directory
	$(call print_header,"BUILD","                            ")
	$(call print_info,"Building $(BINARY_NAME)...")
	@mkdir -p $(BIN_DIR)
	@$(GO) build $(GOFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./cmd/scaffold 2>&1 | grep -v "^$(BIN_DIR)" || true
	$(call print_success,"Binary built: $(BIN_DIR)/$(BINARY_NAME)")
	@echo ""

scaffold: build ## Build and run the scaffold CLI
	$(call print_header,"SCAFFOLD","                        ")
	$(call print_info,"Starting Scaffold CLI...")
	@echo ""
	@$(BIN_DIR)/$(BINARY_NAME)
	@echo ""

run: ## Run scaffold directly (without building)
	$(call print_header,"RUN","                              ")
	$(call print_info,"Running Scaffold CLI...")
	@echo ""
	@$(GO) run ./cmd/scaffold
	@echo ""

clean: ## Remove build artifacts
	$(call print_header,"CLEAN","                            ")
	$(call print_info,"Removing build artifacts...")
	@rm -rf $(BIN_DIR)
	@$(GO) clean
	$(call print_success,"Clean complete")
	@echo ""

hooks: ## Activate git hooks for commit linting
	$(call print_header,"SETUP HOOKS","                       ")
	$(call print_info,"Activating git hooks...")
	@git config core.hooksPath .githooks
	@chmod +x .githooks/commit-msg .githooks/pre-push
	$(call print_success,"Git hooks activated")
	@echo "$(BLUE)→ Commit messages will now be validated locally$(RESET)"
	@echo ""

install-tools: ## Install required development tools
	$(call print_header,"INSTALL-TOOLS","                   ")
	$(call print_info,"Installing development tools...")
	@echo "  • Installing golangci-lint..."
	@$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest > /dev/null 2>&1
	@echo "  • Installing goimports..."
	@$(GO) install golang.org/x/tools/cmd/goimports@latest > /dev/null 2>&1
	$(call print_success,"Tools installed")
	@echo ""

commit-lint: ## Validate commit messages (shows format rules)
	$(call print_header,"COMMIT LINT","                      ")
	@echo ""
	@echo "$(BOLD)Conventional Commits Format$(RESET)"
	@echo "$(BLUE)────────────────────────────────────────────────$(RESET)"
	@echo ""
	@echo "  $(CYAN)<type>(<scope>): <description>$(RESET)"
	@echo ""
	@echo "$(YELLOW)Allowed types:$(RESET)"
	@echo "  • feat     — new feature"
	@echo "  • fix      — bug fix"
	@echo "  • docs     — documentation"
	@echo "  • style    — code style (formatting, semicolons, etc)"
	@echo "  • refactor — code refactoring without feature change"
	@echo "  • perf     — performance improvement"
	@echo "  • test     — test changes"
	@echo "  • chore    — dependency or tooling change"
	@echo "  • ci       — CI/CD changes"
	@echo "  • revert   — revert a previous commit"
	@echo ""
	@echo "$(YELLOW)Rules:$(RESET)"
	@echo "  • Type and scope are lowercase"
	@echo "  • Scope is optional"
	@echo "  • Description starts with lowercase"
	@echo "  • No period at end"
	@echo "  • Max 100 characters"
	@echo ""
	@echo "$(YELLOW)Examples:$(RESET)"
	@echo "  feat: add user authentication"
	@echo "  fix(api): handle null responses"
	@echo "  docs(readme): update setup instructions"
	@echo "  refactor(generators): extract common logic"
	@echo ""

validate: fmt vet lint test ## Run full validation suite
	@echo ""
	$(call print_header,"VALIDATION PASSED","            ")
	@echo "$(GREEN)"
	@echo "  ✓ Code formatted"
	@echo "  ✓ Vet checks passed"
	@echo "  ✓ Linting passed"
	@echo "  ✓ Tests passed"
	@echo "$(RESET)"
	@echo ""

.DEFAULT_GOAL := help
