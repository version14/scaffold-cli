---
name: dot-add-generator
description: Add a new generator package to the `dot` scaffolding tool — creates the generator.go + manifest.go + files/, registers it in `internal/cli/registry.go`, writes the doc in `docs/contributor/generators/`, and updates affected sibling generators when scope overlaps. Trigger when the user says "add a generator", "create a generator", "new generator", "ajoute un générateur", or describes a chunk of project scaffolding that doesn't yet exist as a generator.
source: local-git-analysis
version: 1.0.0
---

# dot — Add a Generator

A **generator** in `dot` is a Go package under `generators/<name>/` that writes files into a `VirtualProjectState`. One generator = **one purpose**. If a new generator overlaps an existing one's output, split — never duplicate.

This skill walks adding a generator end-to-end: planning, scaffolding, registration, docs, tests, and verification.

---

## HARD RULES (Iron Law — never skip)

1. **Announce first, generate second.** Before any file write, output a short plan stating:
   - What the new generator does (its single purpose)
   - Which existing generators will be modified (and what changes — versions, manifests, files, docs)
   - Which `flows/*.go` resolvers must be updated
   - Which `tools/test-flow/testdata/*.json` fixtures must be added or extended
   Then **ASK** the user to confirm before touching code.
2. **One generator = one purpose.** If proposed output overlaps another generator's outputs, split it. Re-use via `DependsOn`.
3. **Bump version + update doc on every modification.** Any time you change a sibling generator's manifest, behaviour, outputs, or files, bump `Manifest.Version` (semver — patch for fix, minor for new behaviour) and update `docs/contributor/generators/<name>.md`.
4. **Tests live in the generator and in `test-flow`.** Every generator must have at least:
   - Unit tests for any non-trivial pure logic
   - Functional checks (manifest `Validators`)
   - Test commands that the generated project can run (build, typecheck, unit tests of generated code)
   - One or more matching `tools/test-flow/testdata/*.json` cases that exercise it end-to-end
5. **Never omit `PostGenerationCommands` or `TestCommands`.** Every generator that produces runnable code must declare both. `PostGenerationCommands` installs dependencies and runs codegen (e.g. `pnpm install --dangerously-allow-all-builds`, `drizzle-kit generate`). `TestCommands` must at minimum typecheck and run the unit test suite of the generated project (e.g. `pnpm exec tsc --noEmit`, `pnpm exec vitest run unit`, `pnpm exec biome check .`). Do not leave either field as an empty slice when the generator produces TypeScript, Go, or any other compiled/typed output.
6. **If a `drizzle` / `prisma` ORM generator is involved**, ensure post-gen commands run `drizzle-kit generate` (or equivalent) after writes. See `generators/drizzle_*/manifest.go` for the pattern.
7. **Formatters / linters depend on `*`.** Generators whose only job is to format the produced tree (Biome, Prettier, …) must set `DependsOn: []string{"*"}` so they run last. Do not list explicit deps for these.
8. **Register in `internal/cli/registry.go`.** A generator that isn't registered does not exist for the CLI.

---

## Step-by-step workflow

### 1. Scope the generator

Read and confirm with the user:

- **Name** — snake_case, matches `Manifest.Name` and directory name exactly
- **Purpose** — one sentence
- **Inputs** — which `ctx.Answers` keys it reads (if any)
- **Outputs** — exact file paths it writes
- **DependsOn** — which generators must run before
- **ConflictsWith** — mutually exclusive generators
- **PostGenerationCommands** — `pnpm install`, `drizzle-kit generate`, `go mod tidy`, etc.
- **TestCommands** — `pnpm run build`, `pnpm exec tsc --noEmit`, `pnpm exec vitest run`, optional `Background` dev-server smoke test with `NoCache: true`

### 2. Pre-generation announcement (MANDATORY)

Output text that looks like this and stop for confirmation:

```
I will create the generator `<name>` that <one-line purpose>.

It will:
- Write: <file paths>
- Depend on: <list>
- Conflict with: <list or none>
- Run post-gen: <commands>
- Run in tests: <commands>

It will also modify:
- generators/<sibling>/manifest.go — <why> — version bump <old> → <new>
- docs/contributor/generators/<sibling>.md — <what is updated>
- flows/<flow>.go — <which resolver branch + which question consumes this>
- tools/test-flow/testdata/<id>_<name>.json — <new case OR which existing case is extended>

Proceed?
```

### 3. Create the package

Layout:

```
generators/<name>/
├── generator.go    # Go struct + Generate(ctx)
├── manifest.go     # dotapi.Manifest var
└── files/          # //go:embed all:files — static or .tmpl templates
```

**`generator.go` minimal template** (mirror `generators/express_openapi_setup/generator.go`):

```go
package <name>

import (
    "embed"

    "github.com/version14/dot/internal/render"
    "github.com/version14/dot/pkg/dotapi"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string    { return Manifest.Name }
func (g *Generator) Version() string { return Manifest.Version }

//go:embed all:files
var fs embed.FS

func (g *Generator) Generate(ctx *dotapi.Context) error {
    return render.NewLocalFolderRenderer(ctx.State).Render(fs, nil)
}
```

For generators that merge into existing files (e.g. `package.json`, `src/app.ts`), prefer `ctx.State.OpenJSON` / `OpenYAML` / `OpenGoMod` / direct `WriteRaw` with anchor-replacement rather than full file rewrites — see `generators/auth_better_auth` and `generators/express_swagger_jsdoc`.

**`manifest.go` template**:

```go
package <name>

import "github.com/version14/dot/pkg/dotapi"

var Manifest = dotapi.Manifest{
    Name:        "<name>",
    Version:     "0.1.0",
    Description: "<one-line>",
    DependsOn:   []string{"<dep>"},
    Outputs:     []string{"<path>"},
    Validators: []dotapi.Validator{
        {
            Name: "<name>-structure",
            Checks: []dotapi.Check{
                {Type: dotapi.CheckFileExists, Path: "<path>"},
            },
        },
    },
    PostGenerationCommands: []dotapi.Command{
        // {Cmd: "pnpm install --dangerously-allow-all-builds"},
    },
    TestCommands: []dotapi.Command{
        // {Cmd: "pnpm exec tsc --noEmit"},
    },
}
```

### 4. Register

Edit `internal/cli/registry.go`:

```go
import <name>pkg "github.com/version14/dot/generators/<name>"

// inside builtinGeneratorEntries():
{Manifest: <name>pkg.Manifest, Generator: <name>pkg.New()},
```

Group it with related generators (Foundation / Backend architecture / Express server / OpenAPI / Prettier / PostgreSQL / Drizzle / Auth …) — preserve the comment-banner sections.

### 5. Wire into flows

If the user's answer to a question should pull this generator in, edit the relevant flow resolver in `flows/*.go` (typically `flows/init.go → resolveMonorepoGenerators`). Append `Invocation{Name: "<name>"}` inside the right conditional branch.

If no existing question selects this generator, **stop** and use the `dot-add-question` skill first.

### 6. Documentation (MANDATORY — all generators, no exceptions)

Create `docs/contributor/generators/<name>.md` from `docs/contributor/generators/_template.md`. Replace every `<!-- TODO -->` and `_placeholder_`. Then:

1. **`docs/contributor/authoring-generators.md`** — add a row to the Built-in generators table at the bottom. If the new generator establishes a new pattern (loop scoping, path prefix, monorepo root), add an example to the relevant section.
2. **`docs/README.md`** — add a row to the generators table if one exists.
3. **Sibling generator docs** — when you modify a sibling generator (version bump, manifest change, new command), update its `docs/contributor/generators/<sibling>.md` — Identity (version), Files written, Validators, PostGenerationCommands, TestCommands sections.
4. **`docs/contributor/authoring-flows.md`** — update the loop resolver example if the new generator is loop-scoped (e.g. per-app generators).
5. **`docs/contributor/architecture.md`** — update only if the generator changes the pipeline contract (new `PathPrefix` usage, new executor scoping rule, new command WorkDir behaviour).

### 7. Tests

- **Unit tests** for non-trivial logic (template helpers, file-merge functions) — `<name>_test.go` next to `generator.go`.
- **Test fixture(s)**: extend an existing case in `tools/test-flow/testdata/` that already touches this scaffolding path, OR add a new case with a timestamped filename `YYYYMMDDhhMM_<flow>_<variant>.json`. Make sure `expected_visited` reflects the new path through the flow if questions changed.
- For ORM generators, set `PostGenerationCommands` with `drizzle-kit generate` (cacheable by default).
- For generators that start a dev server in tests, use `Background: true`, `ReadyDelay: 3-5 * time.Second`, and `NoCache: true`.

### 8. Verify

Run, in order:

```
make build       # or go build -o bin/dot ./cmd/dot
make validate    # fmt + vet + lint + test
make test-flows  # end-to-end fixture tests
```

Fix any failures before declaring done. Generated code must pass its own `tsc --noEmit` / `biome check` / `vitest run` (the project's `TestCommands`).

### 9. Commit

One commit per logical unit. Conventional Commits format (see `.commitlintrc.json`):

- `feat(generators): add <name> generator` for the new package
- `refactor(generators/<sibling>): <what changed>` plus version bump for each modified sibling
- `docs(generators): add doc for <name>` (often folded into the feat commit)
- `test(test-flow): add fixture for <name>` if a new fixture was added

Do **not** auto-commit — wait for explicit user instruction (per project memory).

---

## Reference patterns from this repo

| Pattern | Example generator |
|---------|-------------------|
| Static file embedding via `embed.FS` + `render.NewLocalFolderRenderer` | `express_openapi_setup` |
| Merging into `package.json` via `OpenJSON` | `drizzle_typescript_deps`, `express_server_typescript_deps` |
| Anchor-replacement in existing TS file | `auth_better_auth` (mounts onto `src/app.ts`) |
| Architecture branching (`clean` / `mvc` / `hexagonal`) | `decorators_<arch>_adapter` triplet |
| Wildcard `DependsOn: ["*"]` for last-run formatters | `prettier_config`, `biome_config` |
| ORM + `drizzle-kit generate` post-gen | `drizzle_postgres_adapter` |
| Background dev-server `TestCommand` with `NoCache: true` | `react_app` |

---

## Splitting rule (decision tree)

Before adding files to an existing generator, ask:

1. Does another generator already write any of the proposed paths? → **split**, depend on the existing one.
2. Could a future flow want one subset but not another? → **split**.
3. Are the files governed by different `ConflictsWith` rules? → **split**.
4. Otherwise → extend in place, bump version, update doc.

Examples of splits already in the repo: `auth_better_auth` (code) + `auth_better_auth_schema` (Drizzle schema); `express_server_entrypoint` + `express_server_typescript_deps` + `express_node_tsconfig`.

---

## Common mistakes (do not make)

- Forgetting to register in `internal/cli/registry.go` (generator silently absent)
- Forgetting to bump version on a modified sibling
- Forgetting to update `docs/contributor/generators/<name>.md` after a manifest change
- Forgetting to add the new generator to the Built-in generators table in `docs/contributor/authoring-generators.md`
- Forgetting to update `docs/contributor/authoring-flows.md` resolver examples when a loop-scoped generator is added
- Writing a generator that produces files another generator already owns
- Using `DependsOn: []string{}` for a formatter — must be `["*"]`
- Adding a `PostGenerationCommand` that mutates global state without `NoCache: true` where appropriate
- Leaving `PostGenerationCommands` or `TestCommands` empty for a generator that produces typed/compiled output — always declare both (Iron Law #5)
- Skipping the pre-generation announcement (Iron Law #1)
