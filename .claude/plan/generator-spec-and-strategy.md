# Implementation Plan: Generator Spec and Strategy (Issues #66 + #67)

## Context

The `docs/developer-guide/generators/` directory already has solid implementation-level docs
(`authoring-guide.md`, `fileop-reference.md`, `official-generators.md`, `patch-strategies.md`)
and the internal interface docs are complete. What's missing is:

1. **#66** — A clear panoramic doc that answers "what can generators do, what can't they do,
   and how does one work?" as a *specification* (not a how-to guide).
2. **#67** — A strategy document for the phased *implementation* of generators in dot.

---

## Issue #66 — Specify how generator will work

### Gap analysis

The existing `authoring-guide.md` has a brief "Before you start" section with 4 bullet points
per column. That's not enough for someone trying to understand the system from scratch.
No existing doc answers all three checklist items:
- What generators **can** do
- What generators **cannot** do
- How a generator **works** (lifecycle, not just interface)

### What to create

**New file:** `docs/developer-guide/generators/generator-spec.md`

Audience: contributors building generators and developers trying to understand the system.
Structure: What is a generator → What it can do → What its limits are → How it works.
This is the canonical "understand generators" reference. The authoring guide links here
for the "why"; it stays focused on "how".

#### Section 0 — What is a generator

A generator is a Go struct that implements the `Generator` interface. It knows about one
specific combination of language + module (e.g. `go` + `rest-api`). Given a `Spec`
(the structured description of the user's project), it produces a list of `FileOp`s —
instructions for which files to create, append to, or patch.

A generator does not write files. It does not call APIs. It does not make decisions based
on anything outside the Spec. It is a pure function of its input.

#### Section 1 — What a generator can do (technically)

- **Create files** — emit `Create` or `Template` FileOps to write new files from scratch
- **Append to files** — emit `Append` ops; multiple generators can each contribute lines
  to the same file (e.g. a shared `Makefile` or `docker-compose.yml`)
- **Patch existing files** — emit `Patch` ops using named anchors (import block, main func,
  init func) to inject code into specific locations in an already-created file
- **Emit an AI prompt scaffold** — append to `.dot/prompts.md` via an `Append` op; this file
  is the handoff point between dot (structural scaffolding) and an LLM (business logic).
  Example: a REST API generator appends a prompt describing the generated routes so the
  user can paste it into an LLM and ask it to fill in the handler bodies.
- **Register post-init commands** — return `CommandDef` entries from `Commands()`; these become
  `dot new <noun>` subcommands, persisted to `.dot/config.json` after init
- **Compose other generators** — call another generator's `Apply()` and merge the FileOps;
  this is how architecture pattern generators (clean, hexagonal) plug into REST API generators
- **Vary output by Spec** — read any field from `spec.Spec` to branch output (language,
  architecture, modules, config, etc.)
- **Be language-agnostic** — use `Language() = "*"` to match any project language
- **Claim multiple modules** — return multiple strings from `Modules()` to handle related modules

#### Section 2 — What a generator cannot do (limits)

- **Cannot generate business logic** — dot is not an AI. Generators produce deterministic
  structural scaffolding only. They do not infer domain intent, write handler bodies, or
  produce code that depends on understanding the user's specific problem. That is the user's
  job (aided by the `.dot/prompts.md` handoff).
- **Cannot write files directly** — generators return FileOps; the pipeline performs all I/O
- **Cannot read existing files** — `Apply()` receives only the Spec; it has no filesystem access
- **Cannot be non-deterministic** — no `rand`, no `time.Now()`, no map iteration order in output.
  Same Spec must always produce identical FileOps.
- **Cannot perform side effects** — no network calls, no shell exec, no DB access inside `Apply()`
  or `RunAction()`
- **Cannot override another generator's `Create` at the same priority** — the pipeline aborts;
  use higher priority or switch to `Append`/`Patch`
- **Cannot claim the same (Language, Module) pair as an existing generator** — `Register()` rejects it
- **Cannot write to `.dot/config.json` or `.dot/manifest.json`** — these are pipeline-managed.
  Generators may write to other `.dot/` paths (e.g. `.dot/prompts.md`) via Append ops.

#### Section 3 — How a generator works (lifecycle)

**The binding model: survey questions are interfaces, generators are implementations.**

The `dot init` survey questions act as a typed interface. Each question maps to a field in
`spec.Spec`. The user's answer fills that field. `Registry.ForSpec(spec)` then selects all
generators whose `Language()` and `Modules()` match the Spec. The generator is plugged in
by the registry based entirely on what the user answered — not by any generator-internal logic.

```
survey question          Spec field              generator activated
────────────────────     ─────────────────────   ───────────────────
"Language?" → "go"    →  spec.Project.Language   GoRestAPIGenerator (Language()="go")
"Modules?" → "redis"  →  spec.Modules[].Name     GoRedisGenerator   (Modules()=["redis"])
```

This is the core contract: **questions define the interface; generators implement it.**

**Init phase (`dot init`):**

```
user answers survey
        ↓
spec.Spec is built from answers
        ↓
Registry.ForSpec(spec) → matched generators
        ↓
for each matched generator: Apply(spec) → []FileOp
        ↓
pipeline: resolve conflicts, sort by priority
        ↓
pipeline: write files to disk
        ↓
.dot/config.json written with CommandDefs from matched generators
.dot/prompts.md written if any generator emitted Append ops targeting it
```

**Post-init phase (`dot new <noun> <args...>`):**

```
dot new route UserController
        ↓
.dot/config.json consulted → CommandDef{Action: "rest-api.new-route", Generator: "go-rest-api"}
        ↓
Registry.Get("go-rest-api") → generator
        ↓
generator.RunAction("rest-api.new-route", ["UserController"], spec) → []FileOp
        ↓
pipeline applies ops to existing project files
```

Add mermaid diagrams for both phases.

---

## Issue #67 — Create strategy about the implementation of generators

### What to create

**New file:** `docs/developer-guide/generators/implementation-strategy.md`

This is a strategy document: which generators to build, in what order, and why.

#### Section 1 — Principles

- **Generators are output-only.** No filesystem reads inside Apply/RunAction.
- **One generator per (Language, Module) pair.** The registry enforces this.
- **Composition over duplication.** Architecture pattern generators are their own thing;
  REST API generators compose them rather than duplicating folder structure logic.
- **Stability first.** A generator's `Name()` is a permanent key in users' `.dot/config.json`.
  Once shipped, names must never change.
- **Test coverage required.** Each generator needs table-driven tests before it can ship.

#### Section 2 — Phased rollout

**v0.1 (ships today):**
- `GoRestAPIGenerator` — `go` + `rest-api` (MVC only, no composition yet)

**v0.2 (next milestone):**
Priority order based on usage frequency and composition dependencies:
1. `GoPostgresGenerator` — `go` + `postgres`
   - Why first: highest demand, needed for CRUD generators downstream
   - Deps: none
2. `GoRedisGenerator` — `go` + `redis`
   - Deps: none
3. `GoAuthJWTGenerator` — `go` + `auth-jwt`
   - Deps: ideally after postgres (auth usually stores tokens)
4. `GitHubActionsGenerator` — `*` + `github-actions`
   - Language-agnostic, high value for CI
5. `DockerGenerator` + `DockerComposeGenerator` — `*` + `docker` / `docker-compose`
   - Compose these: DockerCompose can compose Docker

**v0.3+:**
- Architecture pattern generators (`GoCleanArchGenerator`, `GoHexagonalGenerator`)
- These enable REST API generator to compose them instead of embedding MVC-only structure
- Community generator loading (discover from `~/.dot/generators/`)

#### Section 3 — Monorepo strategy

When `spec.Monorepo = true` and services have different languages:
- Each service's generators run independently against that service's sub-Spec
- `MicroservicesGatewayGenerator` (dynamic composition) assembles them
- File paths in FileOps are relative to each service root, not the monorepo root
- The pipeline receives a service root prefix per run

Open question (ref `roadmap/open-decisions.md`): do we run one pipeline per service, or
one pipeline with path-prefixed ops? TBD before v0.3.

#### Section 4 — Community generators

**v0.4 target:** public registry + local loading.

- Local: `~/.dot/generators/` — Go plugins (`.so`) or interpreted (if we add a scripting layer)
- Registry: `dot generator add <name>` fetches and caches
- Validation: community generators must pass the same `Register()` conflict check
- No trust escalation: community generators cannot write outside the project root

#### Section 5 — Testing strategy

Every generator must have:
1. `TestApply_*` — one case per module/arch combination, assert exact file paths and kinds
2. `TestRunAction_*` — one case per action, assert FileOp paths and content
3. `TestRegister_*` — assert no conflict with all other registered generators

Shared test helpers go in `generators/testutil/testutil.go`.

---

## Implementation Steps

### Step 1 — Create `generator-spec.md` (closes #66)

File: `docs/developer-guide/generators/generator-spec.md`

Content: sections 1–3 from the #66 plan above, with a mermaid lifecycle diagram.
Link from `authoring-guide.md` "Before you start" → "See generator-spec.md".
Link from `docs/developer-guide/README.md` generators section.

### Step 2 — Create `implementation-strategy.md` (closes #67)

File: `docs/developer-guide/generators/implementation-strategy.md`

Content: sections 1–5 from the #67 plan above.
Link from `docs/developer-guide/README.md` generators section.

### Step 3 — Update README index

`docs/developer-guide/README.md` generators section currently lists 4 files.
Add the two new files.

### Step 4 — Update authoring guide cross-reference

`docs/developer-guide/generators/authoring-guide.md` "Before you start" section:
Add a line: "For the complete specification of what generators can and cannot do,
see [generator-spec.md](generator-spec.md)."

---

## Key Files

| File | Operation | Description |
|------|-----------|-------------|
| `docs/developer-guide/generators/generator-spec.md` | Create | Closes #66 |
| `docs/developer-guide/generators/implementation-strategy.md` | Create | Closes #67 |
| `docs/developer-guide/README.md` | Modify | Add links to new files |
| `docs/developer-guide/generators/authoring-guide.md` | Modify | Cross-ref to spec |

---

## Risks and Mitigation

| Risk | Mitigation |
|------|------------|
| Monorepo pipeline strategy is an open question | Document as open decision in strategy, link to `open-decisions.md` |
| v0.2 generator priority order may shift based on user demand | Strategy doc frames priority as guidance, not contract |
| Community generator format (plugins vs scripting) undecided | Note as open decision; don't design the API in this doc |
