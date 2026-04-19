# Spec

Implementation: `internal/spec/spec.go`

---

## What the Spec is

The Spec is the single contract between input layers and the generator engine.

Every input layer (TUI, dot.yaml, future MCP) produces a Spec. Every generator consumes a Spec. Neither side knows about the other. A generator test can construct a Spec directly without touching the CLI. The CLI can be rewritten without touching generators.

---

## Types

```go
// ProjectSpec holds the top-level identity of a project.
type ProjectSpec struct {
    Name     string      // "my-api"
    Language string      // "go", "python", "typescript", ...
    Type     ProjectType // api | cli | monorepo | library | frontend | worker
}

// CoreConfig holds configuration used by official generators.
type CoreConfig struct {
    Linter     string // "golangci-lint" | "eslint" | "none"
    Formatter  string // "gofmt" | "goimports" | "prettier" | "none"
    CI         string // "github-actions" | "gitlab-ci" | "none"
    Deployment string // "docker" | "docker-compose" | "vercel" | "none"
    Monitoring string // "grafana" | "datadog" | "none"
    Tracking   string // "posthog" | "sentry" | "none"
}

// ModuleSpec represents one module requested in the project.
type ModuleSpec struct {
    Name   string
    Config map[string]any // module-specific options, e.g. postgres pool size
}

// Spec is the authoritative description of a project.
type Spec struct {
    Project    ProjectSpec
    Modules    []ModuleSpec
    Config     CoreConfig
    Extensions map[string]any // community generators use this
}
```

**`CoreConfig` vs `Extensions`**: Official generators read typed fields from `CoreConfig` (e.g. `spec.Config.Linter == "golangci-lint"`). Community generators that need options outside `CoreConfig` should read from `Extensions`, which is a free-form map. This keeps the official interface stable while allowing community generators full flexibility.

---

## How Spec flows through the system

```
TUI survey → surveySpec() → Spec
                             │
                             ├── generators activated by registry question tree
                             │   (current v0.1: Registry.ForSpec(spec) → []Generator)
                             │
                             └── generator.Apply(spec) → []FileOp
                                                          │
                                                          └── stored in .dot/config.json
                                                              as ProjectContext.Spec
```

> The current dispatch uses `Registry.ForSpec(spec)` — a flat match on language and modules. The planned redesign (see [architecture/registry-design.md](../architecture/registry-design.md)) activates generators via question-tree traversal instead, so the set of active generators is determined during the survey rather than after it. The `spec.Spec` contract is identical in both models.

The Spec is immutable once produced. Generators may read it but must not modify it.

---

## Helper methods

```go
// ModuleNames returns all module names as a string slice.
spec.ModuleNames() // ["rest-api", "postgres"]

// HasModule checks whether a specific module is in the spec.
spec.HasModule("postgres") // true or false
```

Use `HasModule` in generators to conditionally emit FileOps based on which modules the project requested.
