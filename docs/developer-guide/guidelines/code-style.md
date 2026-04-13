# Go Code Style Guide — dot

This document defines code style and formatting conventions for dot. We follow Go idioms and best practices from [Effective Go](https://golang.org/doc/effective_go). Consistency matters more than any individual rule — when in doubt, follow existing patterns in the codebase.

---

## Tooling

```bash
# Run ALL checks in sequence (recommended)
make validate

# Individual checks
make fmt      # Format code
make vet      # Vet check
make lint     # Lint code
make test     # Run tests with race detector
```

**Raw Go commands:**

```bash
go fmt ./...
go vet ./...
golangci-lint run ./...
go test -race ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

All checks must pass before a PR can be merged.

---

## General Principles

- **Clarity over cleverness** — write code for the next reader
- **Explicit over implicit** — avoid magic; name things for what they do
- **Small functions** — each function should do one thing
- **No dead code** — remove commented-out code before committing

---

## Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Files | `snake_case` | `rest_api.go` |
| Packages | `lowercase` | `spec`, `generator`, `pipeline` |
| Exported types | `PascalCase` | `ProjectSpec`, `FileOp` |
| Exported functions | `PascalCase` | `Apply()`, `ForSpec()` |
| Unexported functions | `camelCase` | `resolveConflicts()`, `collectOps()` |
| Constants | `PascalCase` (exported) or `camelCase` (unexported) | `AnchorMainFunc`, `defaultPriority` |
| Interfaces | `PascalCase` ending in `er` | `Generator` |
| Errors | Start with `Err` | `ErrUnsupportedImportForm`, `ErrNotDotProject` |

---

## Formatting

`gofmt` enforces these automatically:

- **Indentation**: tabs
- **Line length**: no hard limit, ~100 chars when practical
- **Braces**: same line as declaration: `func foo() {`
- **Blank lines**: one blank line between functions; use sparingly within functions

---

## Import Order

Group and sort alphabetically within each group:

```go
import (
    // Standard library
    "encoding/json"
    "fmt"
    "os"

    // Third-party packages
    "github.com/charmbracelet/huh"
    "github.com/spf13/cobra"

    // Internal packages
    "github.com/version14/dot/internal/generator"
    "github.com/version14/dot/internal/spec"
)
```

Use `goimports` to auto-format: `goimports -w ./...`

---

## Error Handling

- Functions that can fail return `error` as the last return value
- Always check errors immediately: `if err != nil { return err }`
- Never swallow errors silently
- Use typed errors for specific cases the caller needs to handle:
  ```go
  // ErrUnsupportedImportForm is returned by the patch pipeline when an import
  // block form is not supported (e.g., build-tag-gated imports).
  var ErrUnsupportedImportForm = errors.New("unsupported import form")
  ```
- Wrap errors with context: `return fmt.Errorf("pipeline: write %s: %w", op.Path, err)`
- Validate inputs at system boundaries (CLI flags, YAML parsing); trust internal code

---

## Testing Conventions

- Test files live in the same package as the code they test, named `*_test.go`
- Test function names: `Test<FuncName>_<Scenario>`:
  ```go
  func TestRegistry_ForSpec_LanguageMismatch(t *testing.T) { ... }
  func TestPatch_AnchorImportBlock_Duplicate(t *testing.T) { ... }
  ```
- Use table-driven tests for multiple scenarios:
  ```go
  tests := []struct {
      name    string
      input   string
      content string
      want    string
      wantErr bool
  }{
      {"block import — add new pkg", `import (\n\t"fmt"\n)`, `"os"`, `import (\n\t"fmt"\n\t"os"\n)`, false},
      {"duplicate — skip", `import (\n\t"fmt"\n)`, `"fmt"`, `import (\n\t"fmt"\n)`, false},
  }
  for _, tt := range tests {
      t.Run(tt.name, func(t *testing.T) { ... })
  }
  ```
- Each test must be independent — no shared mutable state
- Use `t.Helper()` for test helper functions
- Run with race detector: `go test -race ./...`

---

## Go-Specific Best Practices

**Interfaces:**
- Keep interfaces small (1-3 methods in most cases)
- `Generator` is larger by design — it's the core extensibility point
- Accept interfaces, return concrete types

**Concurrency:**
- The FileOp pipeline is intentionally single-threaded (collect → resolve → write)
- Don't introduce goroutines without a clear need
- If you do: pass `context.Context` for cancellation

**Dependencies:**
- Keep `go.mod` minimal; only add packages you use
- Run `go mod tidy` before committing
- Current direct dependencies: `cobra`, `huh`, `lipgloss`

**Memory:**
- Use `strings.Builder` for string concatenation in hot paths
- The pipeline collects all FileOps in memory before writing — this is intentional

---

## Running the Full Validation Suite

Before pushing:

```bash
make validate
```

Executes in sequence:
1. Format code
2. Vet checks
3. Linting
4. Tests with race detector

**Check code coverage:**

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out   # Opens in browser
go tool cover -func=coverage.out   # Shows percentages
```
