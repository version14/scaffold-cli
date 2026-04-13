# Getting Started — dot

---

## Installing dot (end users)

### Homebrew — macOS / Linux

```bash
brew install version14/tap/dot
```

### curl — macOS / Linux (no Go required)

```bash
curl -fsSL https://raw.githubusercontent.com/version14/dot/main/install.sh | sh
```

Installs to `/usr/local/bin`. Override the destination:

```bash
INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/version14/dot/main/install.sh | sh
```

### go install

```bash
go install github.com/version14/dot/cmd/dot@latest
```

### Keeping dot up to date

```bash
dot self-update
```

Fetches the latest release from GitHub and replaces the binary in place. Works regardless of how dot was originally installed.

### Uninstalling

```bash
# Homebrew
brew uninstall dot

# curl / go install / from source
curl -fsSL https://raw.githubusercontent.com/version14/dot/main/uninstall.sh | sh
```

If you installed to a custom directory, set `INSTALL_DIR` first:

```bash
INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/version14/dot/main/uninstall.sh | sh
```

Project `.dot/` directories are left untouched — remove them manually if needed.

---

## Setting up for development

### Prerequisites

| Tool | Version | Install |
|------|---------|---------|
| go   | 1.21+   | [go.dev/doc/install](https://go.dev/doc/install) |
| git  | Latest  | [git-scm.com](https://git-scm.com/) |

### Setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/version14/dot.git
   cd dot
   ```

2. **Activate git hooks** (one-time, after cloning)

   ```bash
   make hooks
   ```

   This activates commit message linting. Your commits will be validated locally before being created.

3. **Download dependencies**

   ```bash
   go mod download
   ```

4. **Run dot**

   ```bash
   go run ./cmd/dot init
   ```

   This launches the interactive TUI to scaffold a new project.

---

## Commit Message Convention

We follow **Conventional Commits** format. Messages are validated automatically.

### Format

```
<type>(<scope>): <description>
```

### Examples

```bash
git commit -m "feat: add new generator"
git commit -m "fix(pipeline): handle empty file ops"
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

```
dot/
├── cmd/dot/                  ← CLI entry point (os.Args dispatch, no framework)
│   ├── main.go               ← thin: run(os.Args[1:]) → os.Exit
│   ├── build.go              ← buildVersion, buildRegistry()
│   ├── styles.go             ← lipgloss styles + ASCII banner
│   ├── cmd_init.go           ← dot init  (huh TUI → Spec → generators)
│   ├── cmd_new.go            ← dot new <type> <name>
│   ├── cmd_help.go           ← dot help  (reads .dot/config.json)
│   └── cmd_selfupdate.go     ← dot self-update
├── internal/
│   ├── spec/                 ← Spec, ProjectSpec, CoreConfig, ModuleSpec
│   ├── generator/            ← Generator interface, Registry, FileOp, CommandDef
│   ├── project/              ← ProjectContext, Load, Save (.dot/config.json)
│   └── pipeline/             ← FileOp collect → resolve → write atomically
├── generators/
│   ├── go/                   ← official Go generators
│   └── common/               ← language-agnostic (CI, Docker — planned v0.2)
├── templates/                ← embedded via go:embed
└── go.mod
```

---

## Common Commands

```bash
# Show available commands
make help

# Build and run dot
make dev

# Build the binary into bin/dot
make build

# Run dot directly (without building)
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
go build -o bin/dot ./cmd/dot
go run ./cmd/dot
go test ./...
go fmt ./...
```

---

## Troubleshooting

**Go version mismatch**

```bash
go version  # should be 1.26+
```

**Dependency issues**

```bash
go mod tidy
go mod download
go mod verify
```

**Tests failing**

```bash
go test -v ./...
```

**Build errors**

```bash
go mod download
go build ./...
```

For other issues, check the [FAQ](../developer-guide/faq.md) or open a [Discussion](https://github.com/version14/dot/discussions).
