# Scaffold CLI

[![CI](https://github.com/version14/scaffold-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/version14/scaffold-cli/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Scaffold CLI is a modular code generator that builds complete, production-ready project structures from interactive questionnaires. Instead of merging templates, it composes independent generators that create files from specifications, eliminating conflicts and scaling effortlessly from single apps to enterprise monorepos.

---

## Table of Contents

- [Overview](#overview)
- [Getting Started](#getting-started)
- [Development](#development)
- [Architecture](#architecture)
- [Contributing](#contributing)
- [License](#license)

---

## Overview

**The Problem:** Building project scaffolding tools typically involves merging incompatible templates, managing conflicting dependencies, and maintaining exponential combinations of configurations.

**The Solution:** Scaffold CLI uses a **generator-based architecture** where each feature (API layer, database, CI/CD, etc.) is an independent generator. Generators compose together, eliminating conflicts and making the system trivially extensible.

**Key Features:**
- **Interactive CLI**: Answer questions to build a project specification (JSON)
- **Modular Generators**: Independent, composable generators for different layers (base, API, database, auth, CI/CD, testing)
- **Smart Merging**: Conflict-safe file generation with merge strategies for multi-generator files
- **Extensible**: Add new generators in ~1 hour with a clear interface
- **Production-Ready**: Generates complete, deployable monorepos and single-app projects

**Design Philosophy:**
- Generate from specs, don't merge templates
- One generator = one concern
- Minimal complexity, maximum flexibility

---

## Getting Started

### Prerequisites

| Tool | Version | Install                               |
|------|---------|---------------------------------------|
| go   | 1.26+  | [Install](https://go.dev/doc/install) |

### Installation

```bash
git clone https://github.com/version14/scaffold-cli.git
cd scaffold-cli

# Activate the commit-msg hook (optional but recommended)
git config core.hooksPath .githooks

# Install dependencies
go mod download
```

See [docs/getting-started](docs/getting-started/README.md) for the full setup guide.

---

## Development

We use a **Makefile** for convenient command execution with clean, colored output:

```bash
# See all available commands
make help

# Build and run the CLI
make scaffold

# Run validation suite (fmt → vet → lint → test)
make validate

# Individual commands
make build       # Build to bin/scaffold
make test        # Run tests
make fmt         # Format code
make lint        # Lint code
make clean       # Remove build artifacts
```

**Or use raw Go commands:**

```bash
go build -o scaffold ./cmd/scaffold          # Build
go run ./cmd/scaffold                         # Run
go test ./...                                 # Test
go fmt ./...                                  # Format
golangci-lint run ./...                       # Lint
```

---

## Architecture

Scaffold CLI follows a **modular generator pattern** where specifications drive code generation. See [Architecture.md](.claude/ressources/Architecture.md) for detailed design decisions.

```
scaffold-cli/
├── cmd/
│   └── scaffold/
│       └── main.go              # CLI entrypoint
├── internal/
│   ├── survey/                  # Interactive questionnaire
│   │   └── questions.go
│   ├── spec/                    # Project specification (JSON)
│   │   └── spec.go
│   ├── generators/              # Composable generators
│   │   ├── base.go              # Base project structure
│   │   ├── api.go               # API layer (REST/gRPC)
│   │   ├── database.go          # Database setup
│   │   ├── ci_cd.go             # GitHub Actions, Docker
│   │   ├── auth.go              # Authentication scaffolding
│   │   └── testing.go           # Test setup
│   ├── template/                # Template rendering
│   │   └── render.go
│   └── merge/                   # Smart file merging
│       └── merge.go             # Conflict-safe appending
├── templates/                   # Reusable template files
│   ├── rest_handler.go.tpl
│   ├── grpc_handler.proto.tpl
│   ├── github_actions.yml.tpl
│   └── ...
├── docs/                        # Documentation
├── .github/                     # GitHub Actions workflows
└── go.mod
```

### Workflow

1. **Survey** → User answers questions via CLI
2. **Spec** → Answers converted to a project specification (JSON)
3. **Generators** → Independent generators read the spec and produce files
4. **Merge** → Multiple generators can safely modify the same file
5. **Write** → All files written to disk, project complete

---

## Contributing

Contributions are welcome. Please read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a pull request.

---

## License

Distributed under the [MIT License](LICENSE).
