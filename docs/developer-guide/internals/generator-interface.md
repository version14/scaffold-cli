# Generator Interface

Implementation: `internal/generator/generator.go`

---

## The interface

```go
type Generator interface {
    Name() string
    Language() string
    Modules() []string
    Apply(s spec.Spec) ([]FileOp, error)
    Commands() []CommandDef
    RunAction(action string, args []string, s spec.Spec) ([]FileOp, error)
}
```

**`Name()`** — A unique, stable identifier. Used as the key in `.dot/config.json` commands. Once a generator ships, its name must never change. Example: `"go-rest-api"`.

**`Language()`** — The language this generator targets. Use `"*"` for language-agnostic generators (e.g. GitHub Actions CI). The Registry uses this to filter generators against `spec.Project.Language`.

**`Modules()`** — The module names this generator handles. Example: `[]string{"rest-api"}`. The Registry uses this to match generators against `spec.Modules[].Name`. Two generators must not claim the same `(Language, Module)` pair.

**`Apply(spec)`** — Called once during `dot init`. Receives the full project Spec. Returns the `[]FileOp` needed to scaffold the generator's files. Must be deterministic: same Spec, same FileOps, every time.

**`Commands()`** — Returns `[]CommandDef` describing the post-creation commands this generator registers. These are written to `.dot/config.json` after `dot init` so that `dot new` can dispatch them later.

**`RunAction(action, args, spec)`** — Called by `dot new`. `action` matches `CommandDef.Action`. `args` are the positional arguments (e.g. `["UserController"]`). Returns `[]FileOp` to inject into the existing project.

---

## Generator composition

Generators can call other generators. A generator's `Apply()` can invoke another generator's `Apply()` and merge the returned `[]FileOp` into its own output. The pipeline sees a flat list — it does not know or care that some ops came from a composed generator.

This is how architecture patterns work in practice: `GoRestAPIGenerator` composes `GoCleanArchGenerator` when `spec.Config.Architecture == "clean"`. The architecture generator owns the folder structure ops; the REST API generator owns the HTTP-specific files. Neither knows about the other's internals.

### Two composition patterns

**Static composition** — the composing generator directly instantiates the dependency:

```go
func (g *GoRestAPIGenerator) Apply(s spec.Spec) ([]generator.FileOp, error) {
    var ops []generator.FileOp

    // Compose with the architecture generator
    switch s.Config.Architecture {
    case "clean":
        arch := &GoCleanArchGenerator{}
        archOps, err := arch.Apply(s)
        if err != nil {
            return nil, fmt.Errorf("clean-arch: %w", err)
        }
        ops = append(ops, archOps...)
    case "hexagonal":
        arch := &GoHexagonalGenerator{}
        archOps, err := arch.Apply(s)
        if err != nil {
            return nil, fmt.Errorf("hexagonal: %w", err)
        }
        ops = append(ops, archOps...)
    }

    // Then append REST API-specific ops
    ops = append(ops, generator.FileOp{
        Kind: generator.Create, Path: "main.go", ...
    })
    return ops, nil
}
```

Use static composition when the dependency is known at compile time — same package or a stable internal package. Architecture pattern generators are the canonical example.

**Dynamic composition via registry injection** — the composing generator receives the registry at construction time and looks up generators by name at `Apply()` time:

```go
type MicroservicesGatewayGenerator struct {
    Registry *generator.Registry // injected at construction
}

func (g *MicroservicesGatewayGenerator) Apply(s spec.Spec) ([]generator.FileOp, error) {
    var ops []generator.FileOp

    // Compose all declared service generators
    for _, service := range s.Services {
        gen, ok := g.Registry.Get(service.GeneratorName)
        if !ok {
            return nil, fmt.Errorf("service generator %q not registered", service.GeneratorName)
        }
        serviceOps, err := gen.Apply(service.Spec)
        if err != nil {
            return nil, fmt.Errorf("service %s: %w", service.Name, err)
        }
        ops = append(ops, serviceOps...)
    }

    // Add gateway-specific ops
    ops = append(ops, g.gatewayOps(s)...)
    return ops, nil
}
```

In `cmd/dot/build.go`, inject the registry at registration time:

```go
reg := &generator.Registry{}
// Register all service generators first
must(reg.Register(&gogen.GoRestAPIGenerator{}))
must(reg.Register(&tsgen.NodeNestJSGenerator{}))
// Then register the gateway with registry access
must(reg.Register(&commgen.MicroservicesGatewayGenerator{Registry: reg}))
```

By the time `Apply()` is called, all generators are registered — the pointer is valid.

Use dynamic composition when the set of composed generators is not known until the Spec is read (e.g. a microservices project with services of different languages).

### Composition rules

**Priority matters on overlapping paths.** If both the composed generator and the composing generator emit a `Create` op for the same file, the higher-priority op wins. The composing generator uses a higher `Priority` when it needs to override the composed generator.

**Errors must propagate.** If a composed generator returns an error, wrap it with context and return. Never swallow errors from composed generators.

**Determinism is inherited.** If a composed generator is non-deterministic (e.g. uses map iteration), the composing generator inherits that non-determinism. Keep all generators deterministic.

**Composition is invisible to the pipeline.** The pipeline receives a flat `[]FileOp`. It does not know which ops came from composed generators. Conflict detection, priority resolution, and writing work the same way.

---

## Rules every generator must follow

**Apply() must be deterministic.** Same Spec always produces the same FileOps. No random IDs, no timestamps, no map iteration in templates that could vary between runs.

**RunAction() must be safe on an existing project.** It returns FileOps, and the pipeline will apply them to files that already exist. Prefer `Append` and `Patch` over `Create` for RunAction — don't overwrite files the user may have modified.

**Stay inside your module's concern.** A Redis generator should not write to `main.go` except via `Patch`. It should not claim paths that belong to another generator. If you need to modify a file another generator owns, use `Append` or `Patch` with a clearly documented anchor.

**Return errors, don't panic.** If the file's import block is in an unsupported form, return `ErrUnsupportedImportForm`. If a required argument is missing, return a descriptive error. The pipeline handles errors gracefully; a panic takes down the whole process.

---

## CommandDef format

```go
type CommandDef struct {
    Name        string   // "new route"
    Args        []string // ["<name>"] — for dot help display only
    Description string   // shown in dot help
    Action      string   // passed to RunAction, e.g. "rest-api.new-route"
    Generator   string   // matches Generator.Name()
}
```

**`Name` format: `"verb noun"`** where noun is a single hyphenated word with no spaces.

Correct: `"new route"`, `"new handler"`, `"add migration"`

Wrong: `"new rest api"` (space in noun), `"generate route"` (non-standard verb)

Why: `dot new` splits on the first space after `"new"`. `dot new route UserController` maps to key `"new route"` with args `["UserController"]`. A space in the noun breaks this lookup.

Use hyphens for multi-word nouns: `"new rest-endpoint"` not `"new rest endpoint"`.
