# Getting Started

This guide walks you through setting up Scaffold CLI for local development.

---

## Prerequisites

| Tool | Version | Install                               |
|------|---------|---------------------------------------|
| go   | 1.26+  | [Install](https://go.dev/doc/install) |
| git  | Latest  | [Install](https://git-scm.com/)       |

---

## Setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/version14/scaffold-cli.git
   cd scaffold-cli
   ```

2. **Activate git hooks** (one-time, after cloning)

   ```bash
   make hooks
   ```

   This activates commit message linting. Your commits will now be validated locally before being created.

3. **Download dependencies**

   ```bash
   go mod download
   ```

4. **Run the CLI**

   ```bash
   go run ./cmd/scaffold new
   ```

   This starts an interactive questionnaire that will scaffold a new project.

---

## Commit Message Convention

We follow **Conventional Commits** format. Commit messages are validated automatically.

### Format

```
<type>(<scope>): <description>
```

### Examples

```bash
git commit -m "feat: add new generator"
git commit -m "fix(api): handle empty responses"
git commit -m "docs(readme): update installation steps"
git commit -m "refactor(generators): extract common logic"
```

### Types

- `feat` — new feature
- `fix` — bug fix
- `docs` — documentation
- `style` — code style (formatting, etc)
- `refactor` — refactoring
- `perf` — performance
- `test` — tests
- `chore` — dependencies/tooling
- `ci` — CI/CD
- `revert` — revert commit

**Rules:**
- Type is required (lowercase)
- Scope is optional (lowercase)
- Description starts with lowercase
- No period at end
- Max 100 characters

View commit rules anytime:

```bash
make commit-lint
```

For details, see [CONTRIBUTING.md](../../CONTRIBUTING.md).

---

## Project Structure

Here's what you'll work with:

```
scaffold-cli/
├── cmd/scaffold/           # CLI entrypoint
├── internal/
│   ├── survey/            # Interactive questionnaire
│   ├── spec/              # Project specification
│   ├── generators/        # Composable generators
│   ├── template/          # Template rendering
│   └── merge/             # Smart file merging
├── templates/             # Template files
└── go.mod                 # Module definition
```

For details, see the [Architecture Documentation](../../.claude/ressources/Architecture.md).

---

## Common Commands

We use a `Makefile` for convenient command execution. All commands produce clean, colored output:

```bash
# Show available commands
make help

# Build and run the interactive CLI
make scaffold

# Build the binary into bin/scaffold
make build

# Run CLI directly (without building)
make run

# Format code
make fmt

# Lint code
make lint

# Run tests with race detector
make test

# Run full validation (fmt → vet → lint → test)
make validate

# Clean up build artifacts
make clean

# Install development tools (golangci-lint, goimports)
make install-tools
```

**Or use raw Go commands:**

```bash
go build -o scaffold ./cmd/scaffold
go run ./cmd/scaffold
go test ./...
go fmt ./...
```

---

## Troubleshooting

**Go version mismatch**

Make sure your Go version matches the one listed in [Prerequisites](#prerequisites):

```bash
go version
```

**Dependency issues**

If you encounter dependency problems, try:

```bash
go mod tidy
go mod download
go mod verify
```

**Tests failing**

Run tests with verbose output to see what's failing:

```bash
go test -v ./...
```

**Build errors**

Ensure all dependencies are installed:

```bash
go mod download
go build ./...
```

For other issues, check the [FAQ](../developer-guide/faq.md) or open a [Discussion](../../../discussions).
