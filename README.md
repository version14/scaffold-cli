# dot

[![CI](https://github.com/version14/dot/actions/workflows/ci.yml/badge.svg)](https://github.com/version14/dot/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

**dot** is a generative project scaffolding tool. Answer a few questions — dot generates a production-ready project.

- Interactive TUI survey → production-ready project on disk
- Works for TypeScript apps, monorepos, microservices, Go backends, and more
- Extensible: publish generators and plugins for any language or pattern

---

## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [Built-in flows](#built-in-flows)
- [Development](#development)
- [Architecture](#architecture)
- [CI/CD](#cicd)
- [Contributing](#contributing)
- [License](#license)

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

Requires Go 1.26+. Binary lands in `$GOPATH/bin`.

### From source

```bash
git clone https://github.com/version14/dot.git
cd dot
make build        # → bin/dot
./bin/dot version
```

### Keep it up to date

```bash
dot update
```

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
dot scaffold [flow-id] [-out DIR]   # Run an interactive scaffold flow
dot update [PATH]                   # Re-run generators against an existing project
dot doctor [PATH]                   # Diagnose drift between spec and installed tools
dot plugin <list|install|uninstall> # Manage installable plugins
dot flows                           # List available flows
dot generators                      # List registered generators
dot version                         # Print version
dot help                            # Show help
```

**Quick start:**

```bash
dot scaffold                  # pick a flow interactively
dot scaffold monorepo         # use a specific flow by ID
dot scaffold fullstack -out ~/projects
```

After scaffolding, a `.dot/` directory is written alongside the project. It stores the full spec (`spec.json`) and generator manifest (`manifest.json`) so `dot update` and `dot doctor` can work later.

See [docs/user/getting-started.md](docs/user/getting-started.md) for the full walkthrough.
See [docs/README.md](docs/README.md) for the complete documentation index.

---

## Built-in flows

| Flow ID | What it builds |
|---------|---------------|
| `monorepo` | General-purpose project — TypeScript, optional React, optional Biome |
| `fullstack` | TypeScript frontend + optional Go backend |
| `microservices` | N independent services, each with its own name and port |
| `plugin-template` | A publishable dot plugin repository |

The `init` flow also offers a decorator-based API option for Express backends:
class decorators (`@Controller`, `@Get`, `@Body`, `@Response`, `@Auth`),
request/response validation via Zod, and an OpenAPI v3 spec served at
`/docs`. See [docs/user/decorators.md](docs/user/decorators.md).

Run `dot flows` to see the up-to-date list with descriptions.

---

## Development

```bash
make help        # See all available targets

make build       # Compile → bin/dot
make dev         # Build and run

make validate    # fmt → vet → lint → test  (run before every PR)
make test        # Unit tests with race detector
make test-flows  # End-to-end fixture tests (requires pnpm)
make fmt         # Format code
make lint        # Lint with golangci-lint
make clean       # Remove build artifacts

make hooks       # Activate git hooks (commit message validation)
```

**Or raw Go commands:**

```bash
go build -o bin/dot ./cmd/dot
go test ./...
golangci-lint run ./...
```

First time setting up? See [docs/contributor/getting-started.md](docs/contributor/getting-started.md) — includes a one-command setup script for macOS, Linux, and Windows.

---

## Architecture

dot uses a **flow → spec → generator pipeline** architecture.

```
dot scaffold
 └── TUI survey (flow graph)
       └── Spec (typed answers)
             └── Generator resolver (topological sort)
                   └── VirtualProjectState (in-memory file tree)
                         └── Persist → project on disk + .dot/
```

**Package layout:**

```
dot/
├── cmd/dot/          ← main() — thin entry point, imports plugins
├── flows/            ← built-in flow definitions + registry
├── generators/       ← built-in generator packages (one per generator)
├── plugins/          ← in-tree plugins (biome_extras, ...)
├── examples/         ← reference plugin implementations
├── tools/test-flow/  ← end-to-end test runner + fixtures
│
├── internal/
│   ├── cli/          ← command dispatch, Scaffold(), TUI form runner, spinner
│   ├── flow/         ← question DSL, FlowEngine, HookRegistry, FragmentRegistry
│   ├── spec/         ← ProjectSpec, builder, loader
│   ├── generator/    ← registry, executor, resolver, topological sorter, validator
│   ├── state/        ← VirtualProjectState, Persist, JSON/YAML/GoMod helpers
│   ├── commands/     ← post-gen + test command planner and runner
│   ├── dotdir/       ← .dot/ read/write (spec.json, manifest.json)
│   ├── plugin/       ← provider interface, loader, installer
│   └── versioning/   ← semver parser and constraint checker
│
└── pkg/
    ├── dotapi/       ← public Generator interface, Manifest, Context (stable API)
    └── dotplugin/    ← public plugin author API — re-exports from internal
```

See [docs/contributor/architecture.md](docs/contributor/architecture.md) for a deep-dive into each subsystem.

---

## CI/CD

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| **CI** | Push / PR | Vet, lint, test, build |
| **Commitlint** | Push / PR | Validate commit messages |
| **Release** | `v*.*.*` tag | Build multi-platform binaries, create GitHub Release |

**Local checks before pushing:**

```bash
make validate     # fmt → vet → lint → test
make test-flows   # end-to-end fixture tests
make hooks        # activate git hooks for commit validation
```

---

## Contributing

Contributions are welcome. Read [CONTRIBUTING.md](CONTRIBUTING.md) before opening a PR.

Key steps:

1. Set up your environment: `bash scripts/setup-dev.sh` (or `scripts/setup-dev.ps1` on Windows)
2. Make changes and write tests
3. Run `make validate && make test-flows`
4. Commit with [Conventional Commits](https://www.conventionalcommits.org/) format
5. Open a PR

For contributor orientation (where to look, what to read, how the pipeline works), see [docs/contributor/navigation-guide.md](docs/contributor/navigation-guide.md).

---

## License

Distributed under the [MIT License](LICENSE).
