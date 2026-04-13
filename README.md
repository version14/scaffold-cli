# dot

[![CI](https://github.com/version14/dot/actions/workflows/ci.yml/badge.svg)](https://github.com/version14/dot/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

dot is a **universal project companion**. Describe what you want — dot builds it.

- You answer a few questions. dot generates a production-ready project.
- After creation, dot knows your project's architecture and gives you commands to manage it.
- Works for REST APIs, CLIs, frontend apps, monorepos, and more.
- Extensible: anyone can publish generators for new languages, frameworks, and patterns.

---

## Table of Contents

- [Overview](#overview)
- [Install](#install)
- [Usage](#usage)
- [Development](#development)
- [Architecture](#architecture)
- [CI/CD](#cicd)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

**The Problem:** Starting a project today means one of three broken paths:

1. **Opinionated starters** — fast, but you spend hours removing what you don't need.
2. **GitHub template repos** — someone did the work, but you spend 30+ minutes filtering out their decisions.
3. **From scratch** — full control, but 200 lines of boilerplate before you write a single line of business logic.

**dot's answer:** describe exactly what you want (stack, modules, config) and get a working project. Not just for new projects — also for adding features to existing ones.

**How it works:**

```
dot init
 └── TUI survey → Spec
                   └── Generator engine
                          └── FileOp pipeline → project on disk + .dot/config.json
```

After `dot init`, the project has a `.dot/config.json` that knows which generators were used and what commands they registered. `dot new route UserController` works from anywhere in the project.

---

## Install

### Homebrew (macOS / Linux)

```bash
brew install version14/tap/dot
```

### curl (macOS / Linux — no Go required)

```bash
curl -fsSL https://raw.githubusercontent.com/version14/dot/main/install.sh | sh
```

Installs to `/usr/local/bin/dot` by default. Override with `INSTALL_DIR=~/bin sh install.sh`.

### go install

```bash
go install github.com/version14/dot/cmd/dot@latest
```

Requires Go 1.21+. Binary lands in `$GOPATH/bin` (usually already on `$PATH`).

### From source

```bash
git clone https://github.com/version14/dot.git
cd dot
make build        # → bin/dot
./bin/dot version
```

### Keep it up to date

```bash
dot self-update
```

Fetches the latest release from GitHub and replaces the binary in place. Works regardless of how you installed it.

### Uninstall

```bash
# Homebrew
brew uninstall dot

# curl / go install / from source
curl -fsSL https://raw.githubusercontent.com/version14/dot/main/uninstall.sh | sh
```

Project `.dot/` directories are left untouched — remove them manually if needed.

---

## Usage

```bash
dot init                  # Launch TUI → generate project
dot new route <name>      # Generate a new artifact in the current project
dot help                  # List available commands for the current project
dot version               # Print version
dot self-update           # Update to the latest release
```

All commands except `dot init` look for `.dot/config.json` by traversing up from `$PWD` to the git root.

See [docs/getting-started](docs/getting-started/README.md) for the full walkthrough.

---

## Development

We use a **Makefile** for convenient command execution with clean, colored output:

```bash
make help        # See all available commands

make dev         # Build and run dot
make build       # Build to bin/dot
make run         # Run directly (no build step)

make validate    # Full check suite: fmt → vet → lint → test
make test        # Run tests with race detector
make fmt         # Format code
make lint        # Lint code
make clean       # Remove build artifacts
```

**Or raw Go commands:**

```bash
go build -o bin/dot ./cmd/dot
go run ./cmd/dot
go test ./...
go fmt ./...
golangci-lint run ./...
```

---

## Architecture

dot uses a **generator-based architecture** — specifications drive file generation.

```
dot/
├── cmd/dot/                  ← CLI entry point (thin: parse → call internal → print)
├── internal/
│   ├── spec/                 ← Spec, ProjectSpec, CoreConfig, ModuleSpec
│   ├── generator/            ← Generator interface, Registry, FileOp, CommandDef
│   ├── project/              ← ProjectContext, Load, Save (.dot/config.json)
│   └── pipeline/             ← FileOp collect → resolve → write
├── generators/
│   ├── go/                   ← official Go generators
│   └── common/               ← language-agnostic (CI, Docker, etc.)
└── templates/                ← embedded via go:embed
```

**Workflow:**

1. **Survey** → TUI collects user choices
2. **Spec** → choices become a typed `Spec` struct
3. **Registry** → finds generators matching the spec's language + modules
4. **Apply** → generators return `[]FileOp` (create, template, append, patch)
5. **Pipeline** → ops collected in memory, conflicts resolved, then written atomically
6. **Context** → `.dot/config.json` written with spec + available commands

See [docs/developer-guide](docs/developer-guide) for deep-dives.

---

## CI/CD

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| **CI** | Push / PR | Vet, lint, test, build |
| **Commitlint** | Push / PR | Validate commit messages |
| **Release** | `v*.*.*` tag | Build multi-platform binaries, create GitHub Release |

See [docs/CI_CD.md](docs/CI_CD.md) for details.

**Local checks before pushing:**

```bash
make validate    # fmt → vet → lint → test
make hooks       # Activate git hooks for local commit validation
```

---

## Contributing

Contributions are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a PR.

Key steps:
1. Activate git hooks: `make hooks`
2. Make changes and write tests
3. Run: `make validate`
4. Commit with Conventional Commits format
5. Open a PR

---

## License

Distributed under the [MIT License](LICENSE).
