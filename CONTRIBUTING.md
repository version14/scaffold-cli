# Contributing to dot

Thank you for your interest in contributing. This document explains how to get involved, what we expect, and how to get your changes merged.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Features](#suggesting-features)
  - [Activating git hooks](#activating-git-hooks)
  - [Submitting Code Changes](#submitting-code-changes)
- [Commit Conventions](#commit-conventions)
- [Pull Request Process](#pull-request-process)
- [Code Style](#code-style)
- [Testing](#testing)

---

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to uphold these standards.

---

## Getting Started

1. **Fork** the repository on GitHub
2. **Clone** your fork locally:
   ```bash
   git clone https://github.com/your-username/dot.git
   cd dot
   ```
3. **Add the upstream remote:**
   ```bash
   git remote add upstream https://github.com/version14/dot.git
   ```
4. Follow the [development setup guide](docs/getting-started/README.md)

---

## Development Setup

See [Getting Started](docs/getting-started/README.md) for the full setup guide.

---

## How to Contribute

### Reporting Bugs

Before opening an issue:
- Search [existing issues](../../issues) to avoid duplicates
- Make sure you are on the latest version (`git pull upstream main`)

When opening a bug report, include:
- Steps to reproduce
- Expected vs actual behavior
- Your environment (OS, Go version)
- Relevant logs or error output

### Suggesting Features

Open a **Feature Request** issue with:
- A clear description of the problem the feature solves
- Your proposed solution
- Alternatives you considered

Features that align with the project architecture and roadmap are more likely to be accepted.

### Activating git hooks

Git hooks validate commit messages locally before they are created. Activate them once after cloning:

```bash
make hooks
```

Or manually:
```bash
git config core.hooksPath .githooks
chmod +x .githooks/commit-msg
```

View commit rules anytime:
```bash
make commit-lint
```

### Submitting Code Changes

1. **Create a branch** from `main`:
   ```bash
   git checkout main
   git pull upstream main
   git checkout -b feat/your-feature-name
   ```

2. **Make your changes** following the [code style](#code-style) guidelines

3. **Write or update tests** — every new behavior needs a test

4. **Run validation locally**:
   ```bash
   make validate
   ```
   This runs: format → vet → lint → tests

5. **Commit following [commit conventions](#commit-conventions)**

6. **Push and open a Pull Request**:
   ```bash
   git push origin feat/your-feature-name
   ```

**Before submitting the PR, verify:**
- [ ] All validations pass (`make validate`)
- [ ] Commits follow Conventional Commits
- [ ] Tests pass (`make test`)
- [ ] Documentation is updated if needed

---

## Commit Conventions

We follow **Conventional Commits** format. Messages are validated both locally (via git hook) and in CI.

### Format

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type       | When to use                           |
|------------|---------------------------------------|
| `feat`     | New feature or behavior               |
| `fix`      | Bug fix                               |
| `docs`     | Documentation only                    |
| `style`    | Code style (formatting, semicolons)   |
| `refactor` | Code change with no behavior change   |
| `perf`     | Performance improvement               |
| `test`     | Adding or updating tests              |
| `chore`    | Tooling, dependencies, config         |
| `ci`       | CI/CD changes                         |
| `revert`   | Revert a previous commit              |

**Scope** (optional): the area affected, e.g. `spec`, `pipeline`, `generators`, `registry`.

### Examples

```
feat: add user authentication
feat(pipeline): add conflict marker support
fix(registry): error on duplicate module claim
docs(readme): update installation steps
refactor(spec): extract validation logic
test(pipeline): add AnchorImportBlock edge cases
chore: update dependencies
ci: add Go 1.26 matrix
```

### Rules

- Type is required (lowercase)
- Scope is optional (lowercase)
- Description starts with lowercase, no period at end
- Max 100 characters for the subject line
- Use imperative mood ("add" not "adds")
- Reference issues in the footer: `Closes #42`

---

## Pull Request Process

1. **One PR per concern** — don't mix unrelated changes
2. **Fill the PR template** — describe what changed and why
3. **Keep diffs small** — large PRs are hard to review; split if needed
4. **All CI checks must pass** before merging
5. **Address review comments** — iterate on feedback

PRs are merged by maintainers once they have one approving review and all checks are green.

---

## Code Style

See [Code Style Guide](docs/developer-guide/guidelines/code-style.md) for detailed conventions.

---

## Testing

Every PR should maintain or improve existing test coverage.

Critical areas that require table-driven tests:
- `internal/pipeline/patch.go` — import block patching
- `internal/generator/registry.go` — `ForSpec` matching
- `internal/spec/` — spec serialization round-trips

---

## Questions?

Open a [Discussion](../../discussions) or check the [FAQ](docs/developer-guide/faq.md).
