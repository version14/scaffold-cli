# dot — Project Overview

**One-line:** dot is a CLI that scaffolds production-ready projects and extends them safely over time.

---

## The problem

Starting a project has three broken paths:

1. **Opinionated starters** — fast, but you spend hours removing what you didn't ask for.
2. **Template repos** — someone else's decisions, 30 minutes of untangling.
3. **From scratch** — full control, 200 lines of boilerplate before a single line of real code.

Extending an existing project is worse. Add Postgres, the CI config breaks. Add auth, the import graph is a mess. Every addition is manual surgery.

dot's answer: describe what you want once. dot generates a clean base. Later, add modules safely — dot knows your structure and extends without breaking what's there.

---

## What it does today (v0.1)

| Command | What it does |
|---------|-------------|
| `dot init` | TUI survey → generates a Go REST API project |
| `dot new route <name>` | Adds a route file to an existing project |
| `dot new handler <name>` | Adds a handler stub |
| `dot help` | Lists available commands from `.dot/config.json` |
| `dot version` | Prints the binary version |
| `dot self-update` | Updates the binary to the latest release |

One official generator exists: `GoRestAPIGenerator`. It scaffolds `main.go`, `go.mod`, and `routes/routes.go`.

---

## Architecture

Three layers. Every action flows through all three.

```
Input layer      →    Generator engine    →    FileOp pipeline
(CLI TUI / yaml)      (Spec → Registry        (collect → resolve
                       → Apply → FileOps)       → write atomically)
```

**The central invariant:** dot either leaves the project better than it found it, or leaves it exactly as it found it. Never worse. All ops are assembled in memory before anything touches disk. Any failure aborts the whole run.

### Key packages

| Package | Purpose |
|---------|---------|
| `cmd/dot/` | CLI entry point — thin: parse args, call internal, print |
| `internal/spec/` | `Spec` type — the single contract between input and engine |
| `internal/generator/` | `Generator` interface, `Registry`, `FileOp`, `CommandDef` |
| `internal/project/` | `.dot/config.json` and `.dot/manifest.json` read/write |
| `internal/pipeline/` | FileOp collect → conflict resolve → write |
| `generators/go/` | Official Go generators (package `gogen`) |

### How dot init works

```
dot init
  → TUI survey (huh library) → Spec
  → registry.ForSpec(spec) → []Generator
  → generator.Apply(spec) → []FileOp per generator
  → pipeline.Run(ops) → files on disk
  → project.Save(root, ctx, manifest) → .dot/config.json + .dot/manifest.json
```

### How dot new works

```
dot new route UserController
  → project.Load(".") → .dot/config.json
  → ctx.Commands["new route"] → {Generator: "go-rest-api", Action: "rest-api.new-route"}
  → registry.Get("go-rest-api").RunAction("rest-api.new-route", ["UserController"], spec)
  → []FileOp → pipeline.Run → routes/UserController.go
```

### Generators

A generator is a Go struct implementing:

```go
type Generator interface {
    Name() string                                         // unique stable ID
    Language() string                                     // "go", "python", "*"
    Modules() []string                                    // ["rest-api"]
    Apply(s spec.Spec) ([]FileOp, error)                  // dot init
    Commands() []CommandDef                               // post-creation commands
    RunAction(action string, args []string, s spec.Spec) ([]FileOp, error)  // dot new
}
```

Adding support for a new language or framework = writing one struct. No engine changes.

### FileOp kinds

| Kind | What it does |
|---|---|
| `Create` | Write a new file (priority wins on conflict) |
| `Template` | Render a Go text/template then write |
| `Append` | Add to the end of an existing file |
| `Patch` | Insert at a named anchor (`import_block`, `main_func`, `init_func`) |

### .dot/ directory (committed to git)

```
.dot/
├── config.json    ← spec + available commands (required for dot new)
└── manifest.json  ← SHA-256 hash of every generated file (required for conflict detection)
```

---

## Roadmap

| Version | Theme | Status |
|---------|-------|--------|
| v0.1 | CLI loop, one Go REST API generator | Done |
| v0.2 | All languages, all project types, architectures, deployment, tools | Not started |
| v0.3 | `dot add module` + conflict resolution | Not started |
| v0.4 | Public community generator registry | Not started |
| v0.5 | Project as Code (`dot.yaml`, `dot plan`, `dot apply`) | Not started |
| v0.6 | GitLab CI and additional CI providers | Not started |
| v1.0 | Full stabilization — everything works together | Not started |
| v1.1 | MCP server — AI agents can scaffold via MCP protocol | Not started |
| v1.x | Database table definitions, web dashboard, more | Future |

### v0.2 scope (large milestone)

v0.2 is the "content" release. What ships:

**Project types:** single project, monorepo, microservices (with gateway + auto-linked services)

**Languages/frameworks:**
- Backend: Go `net/http`, Node.js Express, Node.js NestJS, Python FastAPI
- Frontend: React (Vite), Next.js, Vue.js (Vite) — all TypeScript

**Architecture patterns** (chosen at `dot init`):
- APIs: MVC, Clean Architecture, Hexagonal
- Frontend: Feature-sliced, Atomic Design, Container/Presentational

**Dev environment:** Docker Compose (dev only — all declared modules as local services)

**Deployment modules:** GitHub Actions deploy workflow, Terraform (AWS + GCP), Kubernetes manifests + Helm

**Add-on tools:** Grafana, Sentry, PostHog, TanStack Router, TanStack Query, shadcn/ui, Payload CMS, gRPC, GraphQL

**CI:** GitHub Actions only. Dynamic — updated when modules are added (adds service containers, deploy jobs).

**Custom generators:** `dot generator add/list/remove` for local generators.

### Open decisions that block future work

1. **Community generator loading** — in-process, subprocess, or embedded registry. Blocks v0.2 custom generators.
2. **Conflict marker format + `dot resolve` UX** — Blocks v0.3.
3. **`dot plan` diff algorithm** — Blocks v0.5.
4. **Architecture pattern as generator modifier** — flag within each generator vs separate modifier generator. Blocks v0.2 API generators.
5. **Microservices init flow** — upfront declaration vs incremental `dot add service`. Blocks v0.2 microservices.
6. **Microservices gateway linking** — static config patch vs service discovery vs env-driven. Blocks v0.2 microservices.
7. **Multi-language monorepo engine iteration** — how does ForSpec run per-app in a monorepo. Blocks v0.2 monorepo/microservices.

---

## Tech stack

| Concern | Tool |
|---------|------|
| Language | Go 1.26 |
| CLI arg dispatch | `os.Args` (no framework) |
| Interactive TUI | `charmbracelet/huh` (forms), `charmbracelet/bubbletea` (progress) |
| Styling | `charmbracelet/lipgloss` |
| Release | GoReleaser — 5 targets (linux/darwin amd64/arm64, windows amd64) |
| Distribution | Homebrew (`version14/homebrew-tap`), curl, `go install` |
| Self-update | GitHub Releases API + atomic binary replacement |
| CI | GitHub Actions — vet, lint (`golangci-lint`), test, build |
| Commit convention | Conventional Commits (validated via hook + CI) |

---

## What dot does NOT do

- Business logic. A CRUD scaffold is the limit.
- Domain-specific patterns — dot does not know what a "UserController" should do in your app.
- Runtime operations — no deploy, run, or monitor.
- AI generation — everything is deterministic, rule-based.

---

## File structure

```
dot/
├── cmd/dot/                  ← CLI (thin layer)
│   ├── main.go               ← run(os.Args[1:]) → os.Exit
│   ├── build.go              ← buildVersion(), buildRegistry()
│   ├── cmd_init.go           ← dot init
│   ├── cmd_new.go            ← dot new
│   ├── cmd_help.go           ← dot help
│   └── cmd_selfupdate.go     ← dot self-update
├── internal/
│   ├── spec/                 ← Spec types
│   ├── generator/            ← Generator interface, Registry, FileOp
│   ├── project/              ← .dot/ read/write
│   └── pipeline/             ← file write execution
├── generators/
│   ├── go/                   ← Go generators
│   └── common/               ← language-agnostic (v0.2+)
├── docs/
│   ├── getting-started/      ← install + dev setup
│   ├── product-documentation/← product brief
│   └── developer-guide/      ← architecture, internals, generator authoring, roadmap
├── install.sh / uninstall.sh ← curl installers
└── .goreleaser.yaml          ← release config
```

---

## Key rules for contributors

1. `cmd/` never imports from other `cmd/` packages. `internal/` never imports from `cmd/`.
2. `Apply()` must be deterministic — same Spec, same FileOps, every time.
3. Never write to disk except through `pipeline.Run`.
4. A registration conflict (two generators claiming the same language+module) is a programming error. `must()` panics.
5. Add table-driven tests for every new patch anchor and registry matching edge case.

---

For the full developer guide: [`docs/developer-guide/`](docs/developer-guide/README.md)
For the product brief: [`docs/product-documentation/README.md`](docs/product-documentation/README.md)
