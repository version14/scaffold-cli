# Contributing

Thank you for your interest in contributing! This document explains how to get involved, what we expect, and how to get your changes merged.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
  - [Reporting Bugs](#reporting-bugs)
  - [Suggesting Features](#suggesting-features)
  - [Activating the commit-msg hook](#activating-the-commit-msg-hook)
  - [Submitting Code Changes](#submitting-code-changes)
- [Commit Conventions](#commit-conventions)
- [Pull Request Process](#pull-request-process)
- [Code Style](#code-style)
- [Testing](#testing)

---

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to uphold these standards. Please report unacceptable behavior to the maintainers.

---

## Getting Started

1. **Fork** the repository on GitHub
2. **Clone** your fork locally:
   ```bash
   git clone https://github.com/your-username/scaffold-cli.git
   cd scaffold-cli
   ```
3. **Add the upstream remote:**
   ```bash
   git remote add upstream https://github.com/version14/scaffold-cli.git
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

When opening a bug report, use the **Bug Report** template and include:
- Steps to reproduce
- Expected vs actual behavior
- Your environment (OS, runtime version, relevant tool versions)
- Relevant logs or screenshots

### Suggesting Features

Open a **Feature Request** issue with:
- A clear description of the problem the feature solves
- Your proposed solution
- Alternatives you considered

Features that align with the project's scope and architecture are more likely to be accepted.

### Activating git hooks

Git hooks validate commit messages locally and prevent commits that don't follow our conventions. Activate them once after cloning:

**Using Make (recommended):**
```bash
make hooks
```

**Or manually:**
```bash
git config core.hooksPath .githooks
chmod +x .githooks/commit-msg
```

Commit messages are validated:
- **Locally** — before commit (via `.githooks/commit-msg` hook)
- **In CI** — on every PR (via GitHub Actions with commitlint)

**View commit rules anytime:**
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

3. **Write or update tests** as needed

4. **Run validation locally**:
   ```bash
   make validate
   ```
   This runs: format → vet → lint → tests (with nice colored output)

5. **Commit following [commit conventions](#commit-conventions)**:
   ```bash
   git commit -m "feat(scope): your message"
   ```
   Your commit message will be validated automatically by the local hook.

6. **Push and open a Pull Request**:
   ```bash
   git push origin feat/your-feature-name
   ```

**Before submitting the PR, verify:**
- [ ] All validations pass (`make validate`)
- [ ] Commits follow Conventional Commits (`make commit-lint` to review rules)
- [ ] Tests pass (`make test`)
- [ ] Code is formatted (`make fmt`)
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

**Scope** (optional): the module or area affected, e.g. `auth`, `api`, `generators`, `docker`.

### Examples

✅ **Good examples:**
```
feat: add user authentication
feat(api): add rate limiting middleware
fix(generators): handle empty project spec
docs(readme): update installation steps
refactor(survey): extract validation logic
test: add test for spec validation
chore: update dependencies
ci: add commitlint to GitHub Actions
```

❌ **Bad examples (will be rejected):**
```
Add user auth              # Missing type
FEAT: add auth             # Type not lowercase
feat: Add auth.            # Description starts with uppercase, ends with period
feat(api): this is a very long commit message that exceeds the 100 character limit  # Too long
```

### Rules

- **Type is required** and must be lowercase
- **Scope is optional** (lowercase) and indicates what part changed
- **Description is required**, starts with lowercase, no period at end
- **Max 100 characters** for the full header (subject line)
- Use the **imperative mood** ("add" not "adds" or "added")
- Reference issues in the footer: `Closes #42`, `Fixes #17`

### Local Validation

Your commits are validated automatically before creation. If the message is invalid, the commit is rejected with a helpful error message.

**To see validation rules:**
```bash
make commit-lint
```

**If a commit is rejected, fix and try again:**
```bash
git commit --amend -m "feat(scope): corrected message"
```

### CI Validation

Commits are also validated in GitHub Actions using commitlint with the `.commitlintrc.json` configuration. This ensures consistency across all contributions.

---

## Pull Request Process

1. **One PR per concern** — don't mix unrelated changes
2. **Fill the PR template** — describe what changed and why
3. **Keep diffs small** — large PRs are hard to review; split if needed
4. **All CI checks must pass** before merging
5. **Address review comments** — don't push force-merges; iterate on feedback
6. **Squash or rebase** before merge if history is messy

PRs are merged by maintainers once they have one approving review and all checks are green.

---

## Code Style

See [Code Style Guide](docs/developer-guide/guidelines/code-style.md) for detailed conventions.

---

## Testing

Every PR should maintain or improve the existing test coverage.

---

## Questions?

Open a [Discussion](../../discussions) or check the [FAQ](docs/developer-guide/faq.md).
