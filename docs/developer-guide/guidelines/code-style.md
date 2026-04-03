# Go Code Style Guide

This document defines the code style and formatting conventions for Scaffold CLI. We follow Go idioms and best practices from [Effective Go](https://golang.org/doc/effective_go). Consistency matters more than any individual rule — when in doubt, follow existing patterns in the codebase.

---

## Tooling

We use standard Go tooling to enforce code quality. **Use the Makefile for easy access:**

```bash
# Run ALL checks in sequence (recommended)
make validate

# Individual checks:
make fmt      # Format code
make vet      # Vet check
make lint     # Lint code
make test     # Run tests with race detector
```

**Or use raw Go commands:**

```bash
go fmt ./...              # Format code
go vet ./...              # Vet for suspicious constructs
golangci-lint run ./...   # Lint code
go test -race ./...       # Run tests with race detector
go test -coverprofile=coverage.out ./...  # Check coverage
go tool cover -html=coverage.out          # View coverage report
```

All checks must pass before a PR can be merged. The CI pipeline enforces this automatically with:
- `go fmt` (no unformatted code)
- `golangci-lint` (no style violations)
- `go test -race` (no race conditions)

---

## General Principles

- **Clarity over cleverness** — write code for the next reader, not the compiler
- **Explicit over implicit** — avoid magic; name things for what they do
- **Small functions** — each function should do one thing
- **No dead code** — remove commented-out code before committing

---

## Naming Conventions

Go uses specific conventions for identifiers. Follow these strictly:

| Element | Convention | Example |
|---------|------------|---------|
| Files | `snake_case` | `api_generator.go` |
| Packages | `lowercase` | `survey`, `generators` |
| Exported types | `PascalCase` | `ProjectSpec`, `APIGenerator` |
| Exported functions | `PascalCase` | `Generate()`, `AskQuestions()` |
| Unexported functions | `camelCase` | `renderTemplate()`, `mergeFiles()` |
| Constants | `PascalCase` (exported) or `camelCase` (unexported) | `MaxRetries`, `defaultTimeout` |
| Interfaces | `PascalCase` ending in `er` | `Generator`, `Reader`, `Writer` |
| Errors | Start with `Err` or end with `Error` | `ErrNotFound`, `InvalidSpecError` |

**Rule:** Exported identifiers (visible outside the package) start with an uppercase letter. Unexported identifiers start with a lowercase letter.

---

## Formatting

Go's `gofmt` enforces these conventions automatically:

- **Indentation**: Use tabs (enforced by `gofmt`)
- **Max line length**: No hard limit, but keep readable (~100 chars when practical)
- **Spaces**: Go uses specific spacing rules; `gofmt` enforces them
- **Braces**: Always on the same line as the declaration: `func foo() {` not `func foo()\n{`
- **Blank lines**: Use sparingly; one blank line between functions and logical sections

---

## Import Order

Imports must be grouped and sorted alphabetically within each group:

```go
import (
	// Standard library
	"encoding/json"
	"fmt"
	"log"

	// Third-party packages
	"github.com/AlecAivazis/survey/v2"

	// Internal packages
	"scaffold-cli/internal/generators"
	"scaffold-cli/internal/spec"
)
```

**Rules:**
1. Standard library imports first
2. Third-party imports next (sorted alphabetically)
3. Internal package imports last (sorted alphabetically)
4. Separate each group with a blank line
5. Use `goimports` to auto-format: `go install golang.org/x/tools/cmd/goimports@latest && goimports -w ./...`

---

## Error Handling

**Go's approach to errors:**

- Functions that can fail return `error` as the last return value
- Always check errors immediately: `if err != nil { return err }`
- Never swallow errors silently — always log or propagate
- Use typed errors for specific cases:
  ```go
  type InvalidSpecError struct {
      Field string
      Reason string
  }

  func (e InvalidSpecError) Error() string {
      return fmt.Sprintf("invalid spec: %s - %s", e.Field, e.Reason)
  }
  ```
- Validate inputs at system boundaries (API, CLI); trust internal code
- Wrap errors with context using `fmt.Errorf`: `return fmt.Errorf("failed to generate: %w", err)`

---

## Testing Conventions

**Go testing patterns:**

- Test files live in the same package as the code they test, named `*_test.go`
  - Code: `generators/api.go`
  - Tests: `generators/api_test.go`
- Test function names start with `Test`, followed by the function being tested:
  ```go
  func TestAPIGenerator_Generate(t *testing.T) { ... }
  func TestAPIGenerator_GenerateWithoutDatabase(t *testing.T) { ... }
  ```
- Use table-driven tests for multiple scenarios:
  ```go
  tests := []struct {
      name    string
      input   ProjectSpec
      wantErr bool
  }{
      {"valid spec", spec1, false},
      {"invalid service", spec2, true},
  }
  for _, tt := range tests {
      t.Run(tt.name, func(t *testing.T) { ... })
  }
  ```
- Each test should be independent and not rely on shared mutable state
- Use `t.Helper()` for test helper functions
- Run tests with race detector: `go test -race ./...`

---

## Go-Specific Best Practices

**Interfaces & Composition:**
- Keep interfaces small (1-3 methods). See the `Generator` interface.
- Use interface{} sparingly; prefer concrete types
- Embed interfaces for composition:
  ```go
  type APIGenerator struct {
      BaseGenerator // embed to reuse
  }
  ```

**Concurrency:**
- Use goroutines and channels for independent tasks
- Always handle context cancellation
- Avoid global state; pass dependencies as arguments

**Dependencies:**
- Keep `go.mod` minimal; only add packages you use
- Run `go mod tidy` before committing
- Prefer `context.Context` for cancellation and deadlines

**Memory & Performance:**
- Avoid unnecessary allocations; reuse buffers when possible
- Use `strings.Builder` for string concatenation
- Profile with `pprof` for hot paths

---

## Running the Full Validation Suite

**Recommended: Use the Makefile**

Before pushing or submitting a PR, run:

```bash
make validate
```

This executes in sequence with pretty colored output:
1. ✓ Formats code
2. ✓ Vet checks
3. ✓ Linting
4. ✓ Tests with race detector

**Or run individual checks:**

```bash
make fmt       # Step 1
make vet       # Step 2
make lint      # Step 3
make test      # Step 4
make build     # Step 5
```

**Raw Go commands (if needed):**

```bash
# Complete validation
go fmt ./... && \
go vet ./... && \
golangci-lint run ./... && \
go test -race -coverprofile=coverage.out ./... && \
go build -o bin/scaffold ./cmd/scaffold && \
echo "✓ All checks passed!"
```

**Check code coverage:**

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out   # Opens in browser
go tool cover -func=coverage.out   # Shows percentages
```
