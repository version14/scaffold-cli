# CI/CD Workflows — dot

This document explains the automated CI/CD pipelines for dot.

---

## Overview

We use GitHub Actions to automate code quality checks, testing, and releases.

### Workflows

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| **CI** | Push to main, PR | Format check, linting, tests, build |
| **Commitlint** | Push to main, PR | Validate commit messages |
| **Release** | Tag push (`v*.*.*`) | Build binaries, create release |

---

## CI Workflow

**File:** `.github/workflows/ci.yml`

Runs on every push to `main` and every PR.

### Jobs

#### 1. Vet — Go static analysis
- Checks for suspicious constructs
- **Command:** `go vet ./...`

#### 2. Lint — Code quality
- Runs golangci-lint
- Checks code formatting with `go fmt`
- **Commands:** `golangci-lint run ./...`, `go fmt ./...`

#### 3. Test — Unit tests
- Runs all tests with race detector
- Generates coverage report, uploads to Codecov
- **Command:** `go test -race -v -coverprofile=coverage.out ./...`

#### 4. Build — Binary compilation
- Builds the `dot` binary
- Depends on vet, lint, test (all must pass)
- **Command:** `go build -v -o dot ./cmd/dot`

### Key Features

- Concurrent execution — vet, lint, and test run in parallel
- Race detection — catches concurrency bugs early
- Required checks — PR can't be merged if any job fails

---

## Commitlint Workflow

**File:** `.github/workflows/commitlint.yml`

Validates every commit message against Conventional Commits format.

```
<type>(<scope>): <description>
```

**Allowed types:** feat, fix, docs, style, refactor, perf, test, chore, ci, revert

See [CONTRIBUTING.md](../CONTRIBUTING.md#commit-conventions) for details.

---

## Release Workflow

**File:** `.github/workflows/release.yml`

Runs when a tag matching `v*.*.*` is pushed.

### Builds

| Platform | Artifact |
|----------|----------|
| Linux x86_64 | `dot-linux-amd64` |
| Linux ARM64 | `dot-linux-arm64` |
| macOS x86_64 | `dot-darwin-amd64` |
| macOS ARM64 | `dot-darwin-arm64` |
| Windows x86_64 | `dot-windows-amd64.exe` |

### How to Release

```bash
git tag v0.1.0
git push origin v0.1.0
```

The workflow will build binaries for all platforms, create a GitHub Release, and attach them.

---

## Local vs CI

### Local Checks (Before Push)

```bash
make validate
```

Runs: format → vet → lint → tests.

### Commit Message Check

**Local:** `.githooks/commit-msg` validates on every commit (`make hooks` to activate)

**CI:** `commitlint` validates on every PR and push

---

## Troubleshooting

**CI passed locally but failed on GitHub**

Check Go version:
```bash
go version  # should match ci.yml (currently Go 1.26)
go mod download
```

**Commitlint failed on PR**

```bash
make commit-lint          # view rules
git commit --amend -m "feat(scope): correct message"
git push origin your-branch
```

**Build job fails**

```bash
go build ./cmd/dot        # run locally, check errors
go vet ./...
```

---

## Configuration Files

| File | Purpose |
|------|---------|
| `.github/workflows/ci.yml` | Main CI pipeline |
| `.github/workflows/commitlint.yml` | Commit validation |
| `.github/workflows/release.yml` | Release automation |
| `.commitlintrc.json` | Commitlint rules |
| `.githooks/commit-msg` | Local commit hook |

---

## Questions?

See [CONTRIBUTING.md](../CONTRIBUTING.md) or open a [Discussion](https://github.com/version14/dot/discussions).
