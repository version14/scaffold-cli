# Architecture

This document describes DOT's internal design for contributors. It explains how the pieces fit together, which package owns which responsibility, and where to look when something goes wrong.

---

## Table of Contents

- [High-level overview](#high-level-overview)
- [Package map](#package-map)
- [The scaffold pipeline](#the-scaffold-pipeline)
- [Flow engine](#flow-engine)
- [Generator pipeline](#generator-pipeline)
- [Virtual filesystem](#virtual-filesystem)
- [Plugin system](#plugin-system)
- [Command execution](#command-execution)
- [Runtime bundle](#runtime-bundle)
- [dot directory (.dot/)](#dot-directory-dot)
- [Key interfaces](#key-interfaces)

---

## High-level overview

```
User answers questions
        │
        ▼
   FlowEngine
   (internal/flow)
        │  FlowContext (answers map)
        ▼
   spec.Build
   (internal/spec)
        │  ProjectSpec
        ▼
   FlowDef.Generators(spec)  ──── plugin.Provider.ResolveExtras(spec)
   (flows/)                                    │
        │  []Invocation                        │
        └──────────────────────────────────────┘
                           │
                           ▼
           generator.ResolveInvocations
           (dep expansion + topo-sort)
                           │  []generator.Invocation (ordered)
                           ▼
             generator.Executor.Execute
                           │  writes to VirtualProjectState
                           ▼
                  state.Persist
                           │  files on disk
                           ▼
                  .dot/spec.json + .dot/manifest.json
                           │
                           ▼
         PostGenerationCommands (pnpm install, etc.)
```

---

## Package map

```
cmd/dot/                  CLI entry point (main.go → cli.Dispatch)
flows/                    Built-in flow definitions + registry
generators/               Built-in generator packages
plugins/                  Built-in (in-tree) plugins
examples/                 Example community plugin
tools/test-flow/          End-to-end test runner for flows

pkg/
  dotapi/                 Public API for generator authors (stable)
  dotplugin/              Public API for plugin authors (stable, re-exports internal)

internal/
  cli/                    Command dispatch, Scaffold(), HuhFormRunner, spinner
  flow/                   Question DSL, FlowEngine, HookRegistry, FragmentRegistry
  spec/                   ProjectSpec, spec.Build(), spec loader/saver
  generator/              Generator interface, Registry, Executor, Resolver, Sorter
  state/                  VirtualProjectState, Persist, JSON/YAML/GoMod helpers
  commands/               Command planner, runner, dedup, {token} interpolation
  dotdir/                 .dot/ directory read/write (spec.json, manifest.json)
  plugin/                 Plugin loader, installer, Provider interface
  versioning/             Semver parser, constraint checker, version cache
  file-utils/             Path helpers, safe write, directory walker
```

---

## The scaffold pipeline

`cli.Scaffold` (in `internal/cli/runner.go`) drives the entire pipeline. Every step is synchronous and returns an error — there is no goroutine fan-out in the core path.

### Step 1 — Flow execution

```go
flowCtx, err := opts.Runner.Run(opts.Flow.Root)
```

`opts.Runner` is either:
- `HuhFormRunner` — interactive TUI (default in the CLI).
- `scriptedRunner` — replays JSON answers (used by `test-flow`).

The engine traverses the question graph, collects answers into `FlowContext.Answers`, and returns.

### Step 2 — Spec building

```go
s := spec.Build(flowCtx, spec.BuildOpts{...})
```

`spec.Build` flattens `FlowContext.Answers` into a typed `ProjectSpec`. Loop answers are stored as `[]map[string]interface{}` keyed by the loop question ID.

### Step 3 — Invocation resolution

```go
flowInvs := opts.Flow.Generators(s)
// ... append plugin extras ...
invs, err := generator.ResolveInvocations(requested, opts.Registry)
```

`ResolveInvocations` (in `internal/generator/resolver.go`) does three things:
1. Transitive dependency expansion (`DependsOn`).
2. Conflict detection (`ConflictsWith`).
3. Topological sort (stable Kahn, in `internal/generator/sorter.go`).

Explicit invocations from the flow are never deduplicated — this allows the same generator to run once per loop iteration with different scoped answers. Only auto-added transitive deps are deduplicated.

### Step 4 — Execution

```go
exec := generator.NewExecutor(opts.Registry, opts.Logger)
err  := exec.Execute(invs, s, vstate)
```

The executor calls each generator's `Generate(*dotapi.Context)` method in order. Each generator writes to the shared `VirtualProjectState`.

`dotapi.Context.Answers` is a scoped view: loop frames overlay the global answers so each per-app generator invocation sees its own `app-name`, `stack`, etc.

**Per-app path scoping (monorepo):** When an invocation has a non-empty `LoopStack` and the scoped answers include a non-empty `app-name` key, the executor passes a prefixed state view to the generator:

```go
if len(inv.LoopStack) > 0 {
    if appName, ok := scoped["app-name"].(string); ok && appName != "" {
        stateForInv = vstate.WithPrefix("apps/" + appName)
    }
}
```

`VirtualProjectState.WithPrefix` returns a lightweight scoped view that shares the underlying file map but prepends `apps/<name>/` to every path the generator writes. Generators always write relative paths (`src/index.ts`, `package.json`, …) and never need to know their position in a monorepo.

The executor also records `PathPrefix` on each manifest:

```go
mans[i].PathPrefix = "apps/" + appName  // set on the resolved manifest copy
```

`PathPrefix` is used downstream by the validator (to check files in the correct app directory) and by the command planner (as a WorkDir fallback when a command has no explicit `WorkDir`).

### Step 5 — Persist

```go
count, err := state.Persist(vstate, root)
```

`state.Persist` walks the virtual tree and writes each file. Content-typed files (JSON, YAML, GoMod) are serialized before writing. Raw content is written as-is. Intermediate directories are created with `MkdirAll`.

### Step 6 — Metadata

`.dot/spec.json` and `.dot/manifest.json` are written. A `.dot/.gitignore` is created to prevent committing DOT's internal state.

### Step 7 — Post-gen commands

`PlanPostGenCommands` collects `PostGenerationCommands` from all manifests, deduplicates them, and runs them via `RunCommandsQuiet` with a live spinner.

---

## Flow engine

The flow engine lives in `internal/flow/engine.go`. It is responsible for driving question traversal and collecting answers. It has no knowledge of terminal I/O.

### Question graph

A flow is a directed acyclic graph of `Question` nodes. Each `Question` implements:

```go
type Question interface {
    ID()   string
    Next(answer Answer) *Next
}
```

`Next` is a struct that either points to the next question or marks the end of the flow:

```go
type Next struct {
    Question Question // nil = End
    End      bool
}
```

### Question types

| Type | Purpose |
|------|---------|
| `TextQuestion` | Free-text input with optional validation |
| `ConfirmQuestion` | Boolean yes/no with separate `Then`/`Else` branches |
| `OptionQuestion` | Single or multi-select from a list of options; each option may branch |
| `LoopQuestion` | Repeated collection of a body (sub-questions) until the user stops |
| `IfQuestion` | Conditional branch evaluated programmatically (no user input) |

### FlowContext

`FlowContext.Answers` holds all answers keyed by question ID. Loop answers are `[]map[string]Answer` — one map per iteration, each containing the body question answers.

### Plugin injection

Before the engine traverses a question, it checks the `HookRegistry` for injections targeting that question's ID:

- **InjectReplace** — swap the question for the plugin's replacement.
- **InjectAddOption** — append options to an `OptionQuestion`.
- **InjectInsertAfter** — splice a new question after the current one (engine pushes it onto the traversal stack).

The `FragmentRegistry` provides reusable question sub-graphs that plugins can reference without re-declaring every question type.

---

## Generator pipeline

### Generator interface

```go
// pkg/dotapi/generator.go
type Generator interface {
    Name()    string
    Version() string
    Generate(ctx *Context) error
}
```

### Registry

`generator.Registry` maps generator names to `Entry` values:

```go
type Entry struct {
    Manifest  dotapi.Manifest
    Generator dotapi.Generator
}
```

Built-in generators register themselves at program startup via `internal/cli/registry.go`. Plugins register via their `Generators()` return value, which the Runtime loads at startup.

### Manifest

`dotapi.Manifest` is the static descriptor every generator publishes:

```go
type Manifest struct {
    Name                   string
    Version                string
    Description            string
    DependsOn              []string
    ConflictsWith          []string
    Outputs                []string
    PostGenerationCommands []Command
    TestCommands           []Command
    Validators             []Validator

    // Runtime-only — not set in manifest.go files.
    PathPrefix string
}
```

`DependsOn` drives the topological sort. `Validators` are checked by `dot doctor` and by `test-flow` after generation.

`PathPrefix` is a runtime-only field. It is never declared in a generator's `manifest.go`. The executor sets it on the resolved manifest copy when a generator runs inside a loop (e.g. `"apps/api"`). It is used by:

- **`internal/generator/validator.go`** — resolves check paths relative to `<root>/<PathPrefix>` instead of `<root>`.
- **`internal/commands/runner.go`** — uses `PathPrefix` as the `WorkDir` fallback when a `PostGenerationCommand` or `TestCommand` has an empty `WorkDir`. This ensures `pnpm install`, `tsc --noEmit`, etc. run inside the app directory rather than the project root.

---

## Virtual filesystem

`state.VirtualProjectState` is an in-memory representation of a project directory. Generators write to it; `state.Persist` flushes it to disk.

### File nodes

```go
type FileNode struct {
    Content     []byte
    ContentType ContentType
    Typed       interface{} // *JSONDoc | *YAMLDoc | *GoMod | nil (raw)
}
```

Content types:

| Constant | Meaning |
|----------|---------|
| `ContentRaw` | Opaque bytes — written as-is |
| `ContentJSON` | Structured JSON; use `*JSONDoc` for merging |
| `ContentYAML` | Structured YAML; use `*YAMLDoc` for merging |
| `ContentGoMod` | `go.mod` file; use `*GoMod` to add/remove modules |

### JSON and YAML helpers

`JSONDoc` and `YAMLDoc` hold a structured representation that generators can modify by key path. Multiple generators can write to the same file; the last write wins for conflicting keys, but append-style helpers (`SetKey`, `AppendArray`) make cooperative generation possible.

`GoMod` provides `AddRequire`, `SetGoVersion`, and similar helpers for safe `go.mod` manipulation.

### Path prefixing (monorepo scoping)

`VirtualProjectState.WithPrefix(prefix string)` returns a lightweight scoped view of the state:

```go
scoped := vstate.WithPrefix("apps/api")
scoped.WriteFile("src/index.ts", content, state.ContentRaw)
// → written at "apps/api/src/index.ts" in the underlying file map
```

The scoped view shares the same underlying `Files` map as the parent. All public methods (`WriteFile`, `UpdateJSON`, `UpdateYAML`, `UpdateGoMod`, …) automatically prepend the prefix. Generators always write root-relative paths and never need to know their position in a monorepo.

---

## Plugin system

### Provider interface

```go
// internal/plugin/loader.go
type Provider interface {
    ID()             PluginID
    Generators()     []generator.Entry
    Injections()     []flow.Injection
    ResolveExtras(s *spec.ProjectSpec) []generator.Invocation
}
```

- `Generators()` — generators the plugin contributes to the Registry.
- `Injections()` — flow injections to register in the HookRegistry.
- `ResolveExtras(spec)` — called after the flow resolver; returns generator invocations that depend on the user's answers (e.g. "if the user picked Biome, add `biome_config`").

### Built-in plugins

In-tree plugins live in `plugins/`. They are blank-imported in `cmd/dot/main.go` so their `init()` functions run and call `dotplugin.RegisterBuiltin(p)`.

### Installed plugins

User-installed plugins live in `~/.dot/plugins/<id>/`. Each directory contains:

```
~/.dot/plugins/my-plugin/
├── plugin.json        ← identity: id, version, description
└── ...                ← plugin source (cloned from git)
```

`plugin.Load()` reads all subdirectories, deserializes `plugin.json`, and returns a slice of `Provider` values.

Currently DOT plugins must be vendored into the main binary (or loaded as Go plugins with `go:build plugin`). The dynamic loader is a planned feature.

### Plugin isolation

Every ID a plugin contributes — question IDs, option values, generator names — must be prefixed with `"<pluginID>."`. This is enforced at registration time by `HookRegistry.Inject`. It prevents collisions between plugins and between a plugin and the core.

---

## Command execution

`internal/commands` handles post-gen and test commands.

### Planning

`commands.Plan(invocations)` deduplicates commands across invocations (by `Cmd + WorkDir`) and returns a `[]PlannedCommand` ordered by first appearance.

**WorkDir resolution order** for each command:

1. `cmd.WorkDir` after `{key}` token interpolation, if non-empty.
2. `manifest.PathPrefix`, if non-empty (set by the executor for loop invocations).
3. Empty string → command runs at the project root.

This ensures that `pnpm install`, `tsc --noEmit`, and other per-app commands automatically run inside `apps/<name>/` for monorepo invocations without requiring each generator to hard-code `WorkDir`.

### Running

`commands.Runner` executes one `PlannedCommand` at a time:

- **Foreground commands** — run to completion; non-zero exit codes are an error.
- **Background commands** — started and then waited on for `ReadyDelay` (default 3s). If the process is still alive, it is sent `SIGTERM` and considered ready.

`cli.RunCommandsQuiet` wraps the runner with a braille spinner. Output is captured. On success, it prints `✓ <cmd> [elapsed]`. On failure, it prints `✗ <cmd>` followed by the captured output.

### Token interpolation

`{key}` tokens in `Cmd` and `WorkDir` fields are substituted from the generator's scoped answers before execution.

---

## Runtime bundle

`cli.Runtime` is assembled once at startup and threaded through every operation:

```go
type Runtime struct {
    Generators *generator.Registry
    Hooks      *flow.HookRegistry
    Fragments  *flow.FragmentRegistry
    Plugins    []plugin.Provider
}
```

`cli.DefaultRuntime()` builds it by:

1. Creating a fresh `generator.Registry` and registering built-in generators.
2. Creating `HookRegistry` and `FragmentRegistry`.
3. Loading installed plugins from `~/.dot/plugins/`.
4. Loading built-in providers from `plugin.Builtins()`.
5. For each provider: registering its generators and injections.

---

## dot directory (.dot/)

The `.dot/` directory is written at the project root after every scaffold or update:

| File | Committed | Contents |
|------|-----------|----------|
| `spec.json` | **Yes** | User answers, flow ID, tool version |
| `manifest.json` | **Yes** | Execution record: which generators ran, at what version, when |
| `.gitignore` | **Yes** | Allows only `spec.json`, `manifest.json`, `.gitignore` |

`dotdir.SaveSpec` / `dotdir.LoadSpec` and `dotdir.SaveManifest` / `dotdir.LoadManifest` are the only functions that touch these files. No other package reads or writes `.dot/` directly.

### spec.json schema

Written by `dotdir.SaveSpec`, read by `dotdir.LoadSpec`. Safe to commit; contains no secrets.

```json
{
  "flow_id": "init",
  "created_at": "2026-04-27T14:00:00Z",
  "metadata": {
    "project_name": "my-app",
    "tool_version": "0.1.0"
  },
  "answers": {
    "project_name": "my-app",
    "monorepo_type": "turborepo",
    "stack": "typescript",
    "use_react": true,
    "use_biome": true,
    "biome_extras.strict_mode": false,
    "services": [
      {"service_name": "auth", "service_port": "3001"}
    ]
  },
  "visited_nodes": ["project_name", "monorepo_type", "stack", "use_react", "use_biome", "biome_extras.strict_mode", "confirm_generate"],
  "loaded_plugins": ["biome_extras"],
  "generator_constraints": {
    "base_project": "^0.1.0",
    "typescript_base": "^0.1.0"
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `flow_id` | string | The flow used to scaffold this project |
| `created_at` | RFC3339 | Timestamp of the first scaffold run |
| `metadata.project_name` | string | Canonical project name (directory name) |
| `metadata.tool_version` | string | DOT version at scaffold time |
| `answers` | object | All collected answers keyed by question ID. Loop answers are arrays of objects. |
| `visited_nodes` | string[] | Ordered list of question IDs visited (used for `dot doctor` and debugging) |
| `loaded_plugins` | string[] | Plugin IDs active at scaffold time |
| `generator_constraints` | object | Map of generator name → semver constraint string |

**What you can safely edit:** `generator_constraints` (to pin or relax a version bound). Everything else is authoritative and changing it may break `dot update` or `dot doctor`.

### manifest.json schema

Written after every scaffold/update run. Records what actually executed, not what was planned.

```json
{
  "tool_version": "0.1.0",
  "last_executed_at": "2026-04-27T14:05:00Z",
  "execution_time_ms": 3241,
  "generators_executed": [
    {
      "name": "base_project",
      "version_constraint": "^0.1.0",
      "resolved_version": "0.1.0",
      "executed_at": "2026-04-27T14:05:00Z",
      "invocation_count": 1,
      "content_hash": ""
    }
  ]
}
```

| Field | Description |
|-------|-------------|
| `tool_version` | DOT binary version that wrote this manifest |
| `last_executed_at` | When the last run completed |
| `execution_time_ms` | Wall-clock time for the full pipeline |
| `generators_executed[].name` | Generator name |
| `generators_executed[].resolved_version` | Actual version from `Manifest.Version` |
| `generators_executed[].invocation_count` | 1 for normal generators, N for loop generators |

`manifest.json` is read by `dot doctor` to check version drift. It is overwritten on every `dot update` run.

---

## Key interfaces

| Interface | Package | Purpose |
|-----------|---------|---------|
| `flow.Question` | `internal/flow` | A node in the flow graph |
| `flow.FlowRunner` | `internal/flow` | Drives a question graph; returns `FlowContext` |
| `flow.FlowAdapter` | `internal/flow` | Provides `Ask(q, ctx)` to the engine |
| `dotapi.Generator` | `pkg/dotapi` | Produces file operations for a given spec |
| `dotapi.Logger` | `pkg/dotapi` | Log sink injected into generators |
| `plugin.Provider` | `internal/plugin` | Contributes generators + injections |

All interfaces are defined in packages that have no terminal dependencies. The Huh TUI is only imported in `internal/cli`.
