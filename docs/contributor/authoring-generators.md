# Authoring Generators

A **generator** is a Go struct that receives the user's answers and writes files into a `VirtualProjectState`. This guide covers the generator interface, how to write files (raw, JSON, YAML, GoMod), how to declare a Manifest, and how to write validators and post-gen commands.

---

## Table of Contents

- [Generator interface](#generator-interface)
- [Context](#context)
- [VirtualProjectState](#virtualprojectstate)
- [Writing files](#writing-files)
  - [From github repository](#from-github-repository)
  - [From external URL](#from-external-url)
  - [From local folder](#from-local-folder)
  - [Raw content](#raw-content)
  - [JSON](#json)
  - [YAML](#yaml)
  - [go.mod](#go-mod)
- [The Manifest](#the-manifest)
- [Dependencies and conflicts](#dependencies-and-conflicts)
- [Validators](#validators)
- [PostGenerationCommands and TestCommands](#postgenerationcommands-and-testcommands)
- [Registering a generator](#registering-a-generator)
- [Loop generators](#loop-generators)
- [Built-in generators](#built-in-generators)

---

## Generator interface

```go
// pkg/dotapi/generator.go
type Generator interface {
    Name()    string
    Version() string
    Generate(ctx *Context) error
}
```

`Name()` must match `Manifest.Name`. `Version()` must match `Manifest.Version`. The engine checks these at registration time.

A minimal generator:

```go
package mygenerator

import (
    "github.com/version14/dot/internal/state"
    "github.com/version14/dot/pkg/dotapi"
)

type MyGenerator struct{}

func (g *MyGenerator) Name()    string { return "my_generator" }
func (g *MyGenerator) Version() string { return "0.1.0" }

func (g *MyGenerator) Generate(ctx *dotapi.Context) error {
    name, _ := ctx.Answers["project_name"].(string)
    ctx.State.WriteFile("hello.txt", []byte("Hello, "+name+"!\n"), state.ContentRaw)
    return nil
}
```

---

## Context

`dotapi.Context` is the per-invocation handle:

```go
type Context struct {
    Spec         *spec.ProjectSpec          // read-only full spec
    Answers      map[string]interface{}     // scoped answers (globals + loop frame)
    State        *state.VirtualProjectState // target filesystem
    PreviousGens []string                   // names of generators already run
    Logger       Logger                     // log sink
}
```

### Answers

`ctx.Answers` is a flat map. For loop-aware generators, loop frame answers overlay the global answers — the generator does not need to know it is inside a loop:

```go
// Works for both non-loop and loop invocations:
serviceName, _ := ctx.Answers["service_name"].(string)
```

### Checking previous generators

Use `ctx.PreviousGens` to guard conditional writes:

```go
import "slices"

if slices.Contains(ctx.PreviousGens, "typescript_base") {
    // typescript is in the project — write tsconfig extension
}
```

---

## VirtualProjectState

`ctx.State` holds the entire in-memory project. All generator writes go here; nothing touches the disk until `state.Persist` is called.

### Path conventions

All paths are relative to the generator's root. For single-app generators this is the project root; for per-app loop generators the executor automatically scopes writes under `apps/<name>/` via `VirtualProjectState.WithPrefix` — generators always write relative paths and never need to know their position in a monorepo:

```go
ctx.State.WriteFile("src/index.ts", content, state.ContentRaw)   // → apps/api/src/index.ts in multi-app
ctx.State.UpdateJSON("package.json", func(d *state.JSONDoc) error { ... }) // → apps/api/package.json
```

Do not use absolute paths or `../` segments. Never construct paths with the app name manually — the executor handles scoping.

---

## Writing files

### Static files via embedded FS (preferred for multi-file generators)

Embed a `files/` directory and render it via `render.NewLocalFolderRenderer`. Files ending in `.tmpl` are executed as Go templates; all others are copied as-is.

```go
import (
    "embed"
    "github.com/version14/dot/internal/render"
    "github.com/version14/dot/pkg/dotapi"
)

//go:embed all:files
var fs embed.FS

func (g *Generator) Generate(ctx *dotapi.Context) error {
    return render.NewLocalFolderRenderer(ctx.State).Render(fs, ctx.Answers)
}
```

See `generators/express_openapi_setup` for the canonical example. The `all:` prefix on `//go:embed` is required to include hidden files and directories.

### Raw content

```go
ctx.State.WriteFile("README.md", []byte("# My Project\n"), state.ContentRaw)
```

Use for Markdown, plain text, shell scripts, or any file format without a structured helper. `WriteFile` overwrites; `CreateFile` returns an error if the path already exists.

### JSON (cooperative merging)

Multiple generators can contribute to the same JSON file — each call merges its keys:

```go
if err := ctx.State.UpdateJSON("package.json", func(d *state.JSONDoc) error {
    d.Merge(map[string]interface{}{
        "name":    projectName,
        "version": "0.1.0",
        "private": true,
        "scripts": map[string]interface{}{
            "build": "tsc",
        },
        "devDependencies": map[string]interface{}{
            "typescript": "^5.4.0",
        },
    })
    return nil
}); err != nil {
    return err
}
```

`UpdateJSON` loads the file if it exists, calls the callback with a `*JSONDoc`, then serializes and writes back. `d.Merge` does a shallow merge; use `d.Set(key, value)` or `d.SetNested([]string{"scripts","dev"}, value)` for targeted writes.

### YAML

```go
if err := ctx.State.UpdateYAML("docker-compose.yml", func(d *state.YAMLDoc) error {
    d.Set("version", "3.9")
    d.SetNested([]string{"services", serviceName, "image"}, image)
    return nil
}); err != nil {
    return err
}
```

Same cooperative pattern as JSON — safe to call from multiple generators targeting the same file.

### go.mod

```go
if err := ctx.State.UpdateGoMod(func(m *state.GoMod) error {
    m.SetModule("github.com/myorg/myapp")
    m.SetGoVersion("1.22")
    m.AddRequire("github.com/charmbracelet/huh", "v1.0.0")
    return nil
}); err != nil {
    return err
}
```

### Fetching from a GitHub archive (base_project pattern)

When a generator needs to seed files from a remote GitHub repository (e.g. a template repo), use `render.PopulateStateFromSnapshot`:

```go
import "github.com/version14/dot/internal/render"

func (g *Generator) Generate(ctx *dotapi.Context) error {
    fetcher := render.NewGitHubArchiveFetcher()
    snapshot, err := fetcher.FetchRepo(ctx, "https://github.com/owner/repo.git", render.FetchOptions{})
    if err != nil {
        return err
    }
    return render.PopulateStateFromSnapshot(ctx.State, snapshot)
}
```

See `generators/base_project/generator.go` for the canonical example. This is only needed for template repos — prefer embedded `files/` for everything else.

---

## The Manifest

Every generator package exports a `Manifest` variable at package scope:

```go
// generators/my_generator/manifest.go
package mygenerator

import "github.com/version14/dot/pkg/dotapi"

var Manifest = dotapi.Manifest{
    Name:        "my_generator",
    Version:     "0.1.0",
    Description: "Scaffolds my stack",
    DependsOn:   []string{"base_project"},
    Outputs:     []string{"src/index.ts", "tsconfig.json"},
    Validators: []dotapi.Validator{
        {
            Name: "my-files",
            Checks: []dotapi.Check{
                {Type: dotapi.CheckFileExists, Path: "src/index.ts"},
            },
        },
    },
    PostGenerationCommands: []dotapi.Command{
        {Cmd: "pnpm install --dangerously-allow-all-builds", WorkDir: ""},
    },
}
```

---

## Versioning and semver constraints

`Manifest.Version` is a semver string (`"0.1.0"`, `"1.2.3-beta"`). `dot doctor` compares recorded constraints against the installed version using the constraint syntax below.

### Constraint syntax

| Expression | Meaning | Example passes |
|------------|---------|---------------|
| `1.2.3` or `=1.2.3` | Exact match | `1.2.3` |
| `>=1.2.3` | At least | `1.2.3`, `1.3.0`, `2.0.0` |
| `>1.2.3` | Strictly greater | `1.2.4`, `2.0.0` |
| `<=1.2.3` | At most | `1.2.3`, `1.0.0` |
| `<1.2.3` | Strictly less | `1.2.2`, `0.9.0` |
| `~1.2.3` | Same major + minor, patch ≥ 3 | `1.2.3`, `1.2.9` — **not** `1.3.0` |
| `^1.2.3` | Same major, version ≥ 1.2.3 | `1.2.3`, `1.9.0` — **not** `2.0.0` |
| `^0.2.3` | Same major=0 + minor, patch ≥ 3 | `0.2.3`, `0.2.9` — **not** `0.3.0` |
| _(empty)_ | Accept any version | always passes |

Use `^` for stable packages (allows minor bumps). Use `~` to lock to a patch range. Use exact match only when a specific API version is required.

Constraints are parsed by `internal/versioning` and stored in `.dot/spec.json` under `generator_constraints`. You normally do not set them manually — `dot doctor` reads the version from `Manifest.Version` at generation time.

---

## Dependencies and conflicts

### DependsOn

List generator names that must run before yours. The resolver does a topological sort and places your generator after all its dependencies.

```go
DependsOn: []string{"base_project", "typescript_base"},
```

If a listed dependency is not in the invocation set, the resolver adds it automatically (transitive dep expansion).

### ConflictsWith

List generator names that may not coexist with yours. The resolver returns an error if both are requested.

```go
ConflictsWith: []string{"webpack_config"},
```

---

## Validators

Validators are structural checks the engine runs after generation and the `dot doctor` command runs on subsequent runs against the on-disk project.

```go
Validators: []dotapi.Validator{
    {
        Name: "structure",
        Checks: []dotapi.Check{
            {Type: dotapi.CheckFileExists, Path: "src/index.ts"},
            {Type: dotapi.CheckFileExists, Path: "tsconfig.json"},
            {Type: dotapi.CheckJSONKeyExists, Path: "package.json", Key: "scripts.dev"},
        },
    },
},
```

**Check types**

| Type | Fields | Passes when |
|------|--------|-------------|
| `CheckFileExists` | `Path` | The file exists in the virtual state (or on disk for `dot doctor`) |
| `CheckJSONKeyExists` | `Path`, `Key` | The JSON file at `Path` contains the dotted `Key` (e.g. `"scripts.dev"`) |

More check types can be added in `pkg/dotapi/manifest.go` and implemented in `internal/generator/validator.go`.

---

## PostGenerationCommands and TestCommands

### PostGenerationCommands

Run after the entire generator pipeline has finished and files have been persisted:

```go
PostGenerationCommands: []dotapi.Command{
    {Cmd: "pnpm install --dangerously-allow-all-builds"},
    {Cmd: "go mod tidy", WorkDir: "api"},
},
```

Commands from all generators are deduplicated (by `Cmd + WorkDir`) and run in declaration order. Use `{key}` tokens for answer substitution:

```go
{Cmd: "go mod init {module_path}", WorkDir: ""},
```

### TestCommands

Run by `test-flow` to verify the generated project works. They are not run during normal scaffolding.

```go
TestCommands: []dotapi.Command{
    {Cmd: "pnpm run build"},
    {
        Cmd:        "pnpm run dev",
        Background: true,
        ReadyDelay: 3 * time.Second,
    },
},
```

Background commands are started, waited on for `ReadyDelay`, checked for crash, and then sent `SIGTERM`. This lets you test that a dev server starts without having to stop it manually.

### NoCache (caching is opt-out)

`Command` has a `NoCache bool` field used by `test-flow`'s case-level cache. **The default is cacheable** — on a fingerprint match the test-flow runner can skip the command. Set `NoCache: true` only when the command must run every invocation regardless of cache state.

```go
PostGenerationCommands: []dotapi.Command{
    // Cacheable by default — no extra field needed.
    {Cmd: "pnpm install --dangerously-allow-all-builds"},
},
TestCommands: []dotapi.Command{
    {Cmd: "pnpm exec tsc --noEmit"},
    {Cmd: "pnpm exec vitest run unit"},
    {Cmd: "pnpm exec biome check ."},

    // Opt out: smoke-start the real dev server on every run.
    {Cmd: "pnpm exec vite", Background: true, ReadyDelay: 4 * time.Second, NoCache: true},
},
```

The case-level cache only short-circuits when **no** PostGen/Test command across the involved manifests has `NoCache: true`. A single opt-out anywhere in the resolved set forces the case to re-run from scratch.

Set `NoCache: true` when:

- the command starts a real Background process whose actual boot you want re-verified on every run (`pnpm exec vite` with `Background: true` in `react_app` is the canonical example),
- the command depends on remote state with no pinned snapshot (an unpinned network call against a live API, etc.),
- you simply aren't confident the outcome is deterministic.

See [test-flow.md — Case-level cache](test-flow.md#case-level-cache) for invalidation rules and the `-no-cache` flag.

---

## Registering a generator

Add the generator to `internal/cli/registry.go`:

```go
func DefaultGeneratorRegistry() (*generator.Registry, error) {
    r := generator.NewRegistry()

    entries := []generator.Entry{
        {Manifest: baseproject.Manifest, Generator: &baseproject.Generator{}},
        {Manifest: mygenerator.Manifest, Generator: &mygenerator.MyGenerator{}},
        // ...
    }

    for _, e := range entries {
        if err := r.Register(e); err != nil {
            return nil, err
        }
    }
    return r, nil
}
```

The generator immediately appears in `dot generators`.

---

## Loop generators

A generator that participates in a loop receives each iteration's answers via scoped `ctx.Answers`. The `LoopStack` in the invocation tells the executor which frame to overlay.

**Write the generator as if it always receives a single set of answers and always writes to the project root.** Do not compute subdirectory paths using loop-frame values — the executor does that automatically.

When a generator invocation has a non-empty `LoopStack`, the executor calls `VirtualProjectState.WithPrefix("apps/<name>")` and passes the prefixed state as `ctx.State`. Every relative path the generator writes (`src/index.ts`, `package.json`, …) is automatically rooted under `apps/<name>/` without the generator knowing its position in the monorepo.

```go
func (g *AppGenerator) Generate(ctx *dotapi.Context) error {
    // Works identically for single-app and loop (multi-app) invocations.
    // ctx.State is already scoped to apps/<app-name>/ when inside a loop.
    return render.NewLocalFolderRenderer(ctx.State).Render(fs, ctx.Answers)
}
```

To read the loop-frame key that identifies the app (e.g. `app-name`), access it from `ctx.Answers` — the executor merges global and loop-frame answers before passing the context:

```go
appName, _ := ctx.Answers["app-name"].(string)
```

Each loop iteration is a separate `generator.Invocation`. The same generator function is called once per iteration, each time with different `ctx.Answers` and a different prefixed `ctx.State`.

---

## Built-in generators

| Name | Package | Purpose |
|------|---------|---------|
| `base_project` | `generators/base_project` | README, .gitignore, LICENSE — always runs first |
| `typescript_base` | `generators/typescript_base` | tsconfig.json, package.json, tooling |
| `react_app` | `generators/react_app` | Vite, React Router, Tailwind; depends on `typescript_base` |
| `biome_config` | `generators/biome_config` | biome.json formatter/linter config; `DependsOn: ["*"]` — runs last |
| `monorepo_ts_workspaces` | `generators/monorepo_ts_workspaces` | Root package.json with workspaces + pnpm-workspace.yaml for TS monorepos |
| `plugin_repo_skeleton` | `generators/plugin_repo_skeleton` | Full DOT plugin repository scaffold |
| `backend_architecture_clean_architecture` | `generators/backend_architecture_clean_architecture` | Clean Architecture folder structure for Express |
| `backend_architecture_mvc_architecture` | `generators/backend_architecture_mvc_architecture` | MVC folder structure for Express |
| `backend_architecture_hexagonal_architecture` | `generators/backend_architecture_hexagonal_architecture` | Hexagonal Architecture folder structure for Express |
| `express_server_entrypoint` | `generators/express_server_entrypoint` | Express app entrypoint (src/app.ts, src/server.ts) |
| `express_server_typescript_deps` | `generators/express_server_typescript_deps` | Express + TS npm dependencies in package.json |
| `express_node_tsconfig` | `generators/express_node_tsconfig` | tsconfig.json tuned for Node/Express |
| `express_shared_errors` | `generators/express_shared_errors` | Shared error classes |
| `express_error_middleware` | `generators/express_error_middleware` | Global error-handling middleware |
| `express_rate_limit` | `generators/express_rate_limit` | express-rate-limit middleware |
| `express_test_setup` | `generators/express_test_setup` | Vitest + supertest setup |
| `express_auth_validators` | `generators/express_auth_validators` | Auth input validators |
| `express_swagger_jsdoc` | `generators/express_swagger_jsdoc` | Swagger JSDoc annotation setup |
| `zod_validation_deps` | `generators/zod_validation_deps` | Zod + reflect-metadata npm deps |
| `express_decorators_core` | `generators/express_decorators_core` | routing-controllers + class-transformer bootstrap |
| `express_openapi_setup` | `generators/express_openapi_setup` | routing-controllers-openapi + swagger-ui-express |
| `decorators_clean_arch_adapter` | `generators/decorators_clean_arch_adapter` | Decorator-compatible Clean Architecture adapter |
| `decorators_mvc_adapter` | `generators/decorators_mvc_adapter` | Decorator-compatible MVC adapter |
| `decorators_hexagonal_adapter` | `generators/decorators_hexagonal_adapter` | Decorator-compatible Hexagonal adapter |
| `prettier_config` | `generators/prettier_config` | .prettierrc; `DependsOn: ["*"]` — runs last |
| `prettier_typescript_deps` | `generators/prettier_typescript_deps` | Prettier npm deps |
| `prettier_express_rules` | `generators/prettier_express_rules` | Express-specific Prettier rules |
| `postgres_docker_compose` | `generators/postgres_docker_compose` | docker-compose.yml with Postgres service |
| `postgres_env_example` | `generators/postgres_env_example` | .env.example with DATABASE_URL |
| `drizzle_config_base` | `generators/drizzle_config_base` | drizzle.config.ts |
| `drizzle_typescript_deps` | `generators/drizzle_typescript_deps` | Drizzle ORM npm deps |
| `drizzle_postgres_adapter` | `generators/drizzle_postgres_adapter` | Drizzle schema + postgres adapter; runs `drizzle-kit generate` |
| `auth_better_auth` | `generators/auth_better_auth` | better-auth setup wired into Express |
| `auth_jwt_vanilla` | `generators/auth_jwt_vanilla` | Vanilla JWT auth (no framework) |
| `auth_better_auth_schema` | `generators/auth_better_auth_schema` | Drizzle schema for better-auth |
| `auth_jwt_users_schema` | `generators/auth_jwt_users_schema` | Drizzle users schema for JWT auth |
| `auth_jwt_mvc_route` | `generators/auth_jwt_mvc_route` | JWT auth route for MVC architecture |
| `auth_jwt_clean_arch_module` | `generators/auth_jwt_clean_arch_module` | JWT auth module for Clean Architecture |
