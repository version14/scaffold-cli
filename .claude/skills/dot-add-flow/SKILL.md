---
name: dot-add-flow
description: Add a complete new flow to `dot` — a `flows/<name>.go` file with its full question graph, a `Generators` resolver, registration in `flows/registry.go`, every new generator the flow needs (delegating to `dot-add-generator`), every new question (delegating to `dot-add-question`), and a full set of `tools/test-flow/testdata/` fixtures. Trigger when the user says "add a flow", "new flow", "ajoute un flow", "create a scaffold for X", or describes a brand-new wizard that doesn't fit any existing flow.
source: local-git-analysis
version: 1.0.0
---

# dot — Add a Flow

A **flow** is a top-level wizard entry point (`dot scaffold <flow-id>`). It is a `FlowDef` declared in `flows/<name>.go`, registered in `flows/registry.go`, and composed of:

- A question graph (root → … → `confirmGenerate`)
- A `Generators` resolver that maps the populated `ProjectSpec` to ordered `Invocation`s
- Generators that may be reused from existing ones or freshly added
- Test fixtures in `tools/test-flow/testdata/`

This skill orchestrates the full flow lifecycle, deferring to `dot-add-question` and `dot-add-generator` for sub-tasks.

---

## HARD RULES (Iron Law — never skip)

1. **Two announcements, then act.** Before any code change:
   - **Phase A — Generator plan**: list every new or modified generator (per `dot-add-generator` rules), including version bumps and doc updates.
   - **Phase B — Question plan**: list every question in the new flow with `ID_`, type, label, options, edges, and what each answer triggers in the resolver (per `dot-add-question` rules).
   Then **ASK** for confirmation. Do not start writing files until the user says proceed.
2. **Unique, stable, kebab-case flow ID.** Used in `dot scaffold <id>` and referenced by every fixture. Never collide with existing IDs (`init`, `fullstack`, `microservices`, `plugin-template`, …) — check `flows/registry.go`.
3. **Every question ID is unique across the whole flow.** No collision with itself, no collision with existing flows if generators are shared.
4. **Reuse before you create.** Walk the existing `generators/` list; reuse anything that matches. Only add a new generator when no existing one fits. Each new generator goes through the full `dot-add-generator` workflow.
5. **The resolver must cover every reachable answer combination.** No silent gaps. If the user can answer X, some `Invocation` must result from it (or you must explicitly comment that X is intentionally a no-op).
6. **Fixtures cover every distinct branch.** At minimum: one happy-path fixture; one fixture per major branching choice (auth on/off, db on/off, each architecture, each formatter, …). Use the timestamped naming convention `YYYYMMDDhhMM_<flow>_<variant>.json`.
7. **Tests across all layers**:
   - Unit tests for non-trivial pure logic in any new generator (`<name>_test.go`)
   - Manifest `Validators` for structural checks (file presence, JSON keys)
   - `TestCommands` that build / typecheck / unit-test / lint the generated project
   - For DB generators: `drizzle-kit generate` (or equivalent) as a post-gen command
   - End-to-end via `make test-flows`
8. **Bump version + update doc** on every modified existing generator.
9. **Formatters depend on `*`** (run last). Apply to any formatter introduced by the new flow.

---

## Step-by-step workflow

### 1. Understand the target

Capture from the user (or ASK):

- **Flow ID** (kebab-case), **Title**, **Description**
- **Stack** (Go, TypeScript, Rust, multi-app monorepo, …)
- **Choice points** (architecture, ORM, auth, lint/format, …)
- **Loops** if multi-app

### 2. Inventory existing assets

- Read `generators/` directory listing — note which generators map to each choice point
- Read `flows/init.go` (and other flows) for established patterns (constants, terminal `confirmGenerate`, validation helpers like `nonEmpty`)
- Read `internal/cli/registry.go` to know what is wired

### 3. Phase A — Generator plan (MANDATORY announcement)

Output:

```
Flow `<flow-id>` — Generator plan

Reuse (no changes):
- <name> — used when <answer> == <value>
- …

Modify (existing generators that must change for this flow):
- generators/<name>/ — <what changes, why> — version <old> → <new>
- docs/contributor/generators/<name>.md — sections to update
- …

Add (brand-new generators — each via dot-add-generator):
- <name> — purpose, depends on <list>, conflicts with <list>, writes <files>, post-gen <cmds>, test cmds <cmds>
- …

Resolver outline (order matters — deps will be expanded by topo sort):
1. base_project
2. <…>
N. <formatter — DependsOn: ["*"]>
```

Stop and ASK.

### 4. Phase B — Question plan (MANDATORY announcement)

Output:

```
Flow `<flow-id>` — Question plan

Graph (root → … → confirmGenerate):
  <ID_> (<type>) — "<label>"
    └── <option/then/else> → <next ID_ or End>
  …

Per-answer impact on resolver:
- <ID_> == <value> → add <generators>
- <ID_> == <value> → add <generators>
- …

Fixtures to add (under tools/test-flow/testdata/):
- <timestamp>_<flow-id>_<variant>.json — covers <branch>
- …

Proceed?
```

Stop and ASK.

### 5. Create / modify generators

For each item in the Add list, run the `dot-add-generator` skill end-to-end (package layout, manifest, embedded `files/`, registration, doc, tests). For each item in the Modify list, apply the change with version bump + doc update.

Do this **before** writing the flow file — the flow's resolver references generators by name, and `make build` will fail if they don't exist yet.

### 6. Write `flows/<flow-id>.go`

Mirror the structure of `flows/init.go`:

```go
package flows

import (
    "github.com/version14/dot/internal/flow"
    "github.com/version14/dot/internal/spec"
)

// constants for stable answer values reused across the resolver
const FOO_BAR = "foo-bar"

// <FlowID>Flow is the … flow. <one-paragraph description, including
// the note that question IDs are stable for dot update.>
func <FlowID>Flow() *FlowDef {
    confirmGenerate := &flow.ConfirmQuestion{
        QuestionBase: flow.QuestionBase{ID_: "confirm-generate"},
        Label:        "Generate the project now?",
        Default:      true,
        Then:         &flow.Next{End: true},
        Else:         &flow.Next{End: true},
    }

    // … declare questions bottom-up …

    return &FlowDef{
        ID:          "<flow-id>",
        Title:       "<Title>",
        Description: "<Description>",
        Root:        <rootQuestion>,
        Generators:  resolve<FlowID>Generators,
    }
}

func resolve<FlowID>Generators(s *spec.ProjectSpec) []Invocation {
    if s == nil { return nil }
    out := []Invocation{{Name: "base_project"}}
    // … read s.Answers, append Invocations …
    return out
}
```

For loops use `LoopQuestion` + emit one `Invocation` per loop frame with `LoopStack`. See `flows/microservices.go` pattern (referenced in `docs/contributor/authoring-flows.md`).

### 7. Register the flow

Edit `flows/registry.go` — add `_ = r.Register(<FlowID>Flow())` in the `Default()` builder. The flow then appears in `dot flows` and `dot scaffold`.

### 8. Fixtures (full coverage)

Under `tools/test-flow/testdata/`, create one JSON file per distinct branch. Use timestamped filenames sorted lexicographically. Required fields (see `0_template.json`):

```json
{
  "name": "<flow-id>_<variant>",
  "flow_id": "<flow-id>",
  "answers": { /* every required key, in the order the user would answer */ },
  "expected_visited": [ /* every ID_ visited, in order */ ],
  "skip_post_commands": false,
  "skip_test_commands": false
}
```

Minimum coverage matrix:

- Happy path (every default, simplest options)
- Each architecture / ORM / auth / formatter variant the flow exposes
- Loop variants (0 iterations if allowed, 1, N>1)
- Boolean confirm questions: both `true` and `false` covered somewhere

### 9. Verify

```
make build
make validate
make test-flows
```

Iterate until green. If a fixture's `expected_visited` mismatches the flow, fix the flow or the fixture deliberately — never blanket-edit fixtures to silence failures.

### 10. Documentation

- Add a row to `README.md`'s "Built-in flows" table
- Add a row to `docs/contributor/authoring-flows.md`'s "Built-in flows" table
- If you added new generators, ensure `docs/contributor/generators/<name>.md` exists and `docs/README.md` lists each one

### 11. Commit

Conventional Commits, multiple commits:

- `feat(generators): add <name>` per new generator (with its doc)
- `refactor(generators/<name>): <change>` per modified generator (with version bump)
- `feat(flow): add <flow-id> flow` for the flow file + registration
- `test(test-flow): add fixtures for <flow-id>` for the fixtures
- `docs(flows): describe <flow-id> flow` if separate doc work was needed

Do **not** auto-commit — wait for explicit user instruction.

---

## Worked example skeleton (Go backend microservice flow)

Phase A — Generator plan:

```
Reuse: base_project, postgres_docker_compose, postgres_env_example
Modify: (none)
Add:
- go_module_base — go.mod + main.go + Dockerfile + healthcheck — DependsOn: base_project
- go_chi_router — Chi router + middleware + /health — DependsOn: go_module_base
- go_postgres_repo — pgx pool + repo skeleton — DependsOn: go_module_base, postgres_docker_compose
- go_test_setup — go test scaffolding + testcontainers helper — DependsOn: go_module_base
```

Phase B — Question plan:

```
project_name (Text) → module_path (Text) → router_choice (Option: chi / stdlib)
  → enable_db (Confirm)
      Then → db_type (Option: postgres) → confirm-generate
      Else → confirm-generate
```

Then resolver pulls `go_module_base` always, `go_chi_router` when `chi`, plus postgres + `go_postgres_repo` when `enable_db && db_type == "postgres"`.

Fixtures:

- `..._go_chi_no_db.json`
- `..._go_stdlib_no_db.json`
- `..._go_chi_postgres.json`

---

## Common mistakes (do not make)

- Skipping Phase A or Phase B announcements
- Adding a flow that depends on a generator that does not yet exist (build will fail)
- Forgetting to register the flow in `flows/registry.go`
- Inadequate fixture coverage — every branching choice must be exercised by some fixture
- Renaming a question `ID_` after first commit (breaks `dot update`)
- Modifying an existing generator without bumping its version and updating its doc
- Putting a formatter ahead of architecture/feature generators in the resolver (instead set `DependsOn: ["*"]` on the formatter)

---

## Skill composition

This skill **delegates**:

- For each new generator → invoke / follow `dot-add-generator` end-to-end
- For each question → follow `dot-add-question` discipline (announcement, fixture impact, version bumps)

It only owns: the flow file, the registry entry, the resolver, and the cross-cutting Phase A / Phase B announcements that make the whole thing reviewable in one pass.
