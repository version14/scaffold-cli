<!-- Mirror of .claude/skills/dot-generator/SKILL.md — update both files together -->
---
name: dot-generator
description: Create, edit, or AI-generate a dot generator (Manifest plus Generator plus docs plus registry entry). Always asks the user a structured set of questions first, then writes the package, registers it, and updates the docs. Use when the user wants to add a generator, edit one, scaffold a generator from a description, or otherwise work in generators/.
---

# dot-generator

Create or modify a dot generator end-to-end: package under `[generators/](../../../generators)`, registration in `[internal/cli/registry.go](../../../internal/cli/registry.go)`, doc in `[docs/contributor/generators/](../../../docs/contributor/generators)`.

This skill never writes anything before completing the structured question pass below. No silent defaults.

## Mode-independent rules

- Generators are pure functions of `ctx`. No `time.Now()`, no `rand`, no map iteration order in output, no network calls outside the documented helpers (`WriteFilesFromGitHub`, `WriteFileFromExternal`).
- `Manifest.Name` MUST equal `Generator.Name()` AND the directory name.
- `Manifest.Version` MUST equal `Generator.Version()`.
- Every output declared in `Manifest.Outputs` MUST have a matching `CheckFileExists` validator.
- Read but do NOT duplicate `[docs/contributor/authoring-generators.md](../../../docs/contributor/authoring-generators.md)`.
- Mirror the style of `[generators/plugin_repo_skeleton/](../../../generators/plugin_repo_skeleton)` (multi-file template embed) or `[generators/base_project/](../../../generators/base_project)` (raw writes).

## Step 1 — Mode prompt (ALWAYS first)

Ask via `AskQuestion` (single-select):

```
init     — scaffold a minimal generator package with TODOs
edit     — modify an existing generator (Manifest tweak, add output, add validator, ...)
generate — full AI generation from a one-line description
```

## Step 2 — Structured question pass (MANDATORY, runs BEFORE any write)

Issue a single batched `AskQuestion` call covering ALL fields below. If a value fails validation, re-ask only the failing question.

Common to `init` and `generate`:

1. `name` — snake_case, used as both the directory under `[generators/](../../../generators)` and `Manifest.Name`. Must NOT collide with any entry in `[internal/cli/registry.go](../../../internal/cli/registry.go)` `builtinGeneratorEntries()` (read it first).
2. `version` — semver string, default `0.1.0`.
3. `description` — one line, used in `Manifest.Description`.
4. `depends_on` — multi-select from existing generator names (or empty).
5. `answers` — list of `{key, type}` entries for the `ctx.Answers["..."]` keys the generator will consume.
6. `outputs` — list of `{path, format}` entries. `format` is one of:
   - `raw` — `ctx.State.WriteRaw`
   - `json` — `ctx.State.OpenJSON` cooperative pattern
   - `yaml` — `ctx.State.OpenYAML`
   - `gomod` — `ctx.State.OpenGoMod`
   - `embed` — multi-file template tree using `local.NewRenderer` + `//go:embed all:files`
7. `validators` — auto-derived: every output gets a `CheckFileExists`. Ask the user for any extra `CheckJSONKeyExists` entries (path + dotted key).
8. `post_gen_commands` — optional list of `{cmd, work_dir}`. Defaults to none.

`edit` mode only:

9. `target` — pick the generator to edit from a list of `[generators/](../../../generators)` subdirs.
10. `change_kind` — single-select: `add file output` / `change manifest field` / `add depends_on` / `add validator` / `add post-gen command` / `change generate behavior`.
11. Follow-ups depending on `change_kind` (path + format for new outputs, manifest field + new value, etc.).

Validation rules (apply BEFORE writing):

- `name`: matches `^[a-z][a-z0-9_]*$`, unique across (1) `builtinGeneratorEntries()` in `internal/cli/registry.go` AND (2) existing directories under `generators/` — check both before validating.
- `version`: parses cleanly through `internal/versioning` (semver `MAJOR.MINOR.PATCH`).
- Each `outputs[*].path`: relative, no leading `/`, no `..` segments.
- Each `depends_on[*]`: must be an existing generator name in the same registry list.

## Step 3a — `init` workflow

Runs only AFTER Step 2 succeeds.

1. Create `generators/<name>/manifest.go` with `var Manifest = dotapi.Manifest{...}` populated from the structured answers (Name, Version, Description, DependsOn, Outputs, Validators, PostGenerationCommands). Every output gets a `CheckFileExists` validator entry automatically.
2. Create `generators/<name>/generator.go` with the struct, `Name()`, `Version()`, `New()`, and a `Generate(ctx)` body that returns `nil` plus a `// TODO: implement` for each declared output.
3. Register in `[internal/cli/registry.go](../../../internal/cli/registry.go)`:
   - Add the import (alphabetical with the existing `<pkg> "github.com/version14/dot/generators/<name>"` lines).
   - Append `{Manifest: <pkg>.Manifest, Generator: <pkg>.New()}` to `builtinGeneratorEntries()`.
4. Copy `[docs/contributor/generators/_template.md](../../../docs/contributor/generators/_template.md)` to `docs/contributor/generators/<name>.md`. Prefill the Identity table (Name, Version, Package), the Files written table (one row per output), the Validators table, and Post-generation commands.
5. Add a row to the "Built-in generators" table in `[docs/contributor/authoring-generators.md](../../../docs/contributor/authoring-generators.md)` and to the generators index in `[docs/README.md](../../../docs/README.md)`.
6. Run `go build ./...` and report. If it fails, print the compiler output and suggest the user run `go build ./... 2>&1 | head -80` to investigate. Do NOT auto-fix. Stop and wait for user input.

### Stub templates for `init`

Use these exact files for mode `init` (replace `<…>` with values from Step 2):

**`generators/<name>/manifest.go`**

```go
package <pkg>

import "github.com/version14/dot/pkg/dotapi"

var Manifest = dotapi.Manifest{
	Name:        "<name>",
	Version:     "<version>",
	Description: "<description>",
	DependsOn:   []string{<depends_on>},
	Outputs:     []string{<outputs_paths>},
	Validators: []dotapi.Validator{
		{
			Name: "<name>",
			Checks: []dotapi.Check{
				// TODO: one CheckFileExists per declared output is auto-added here.
			},
		},
	},
}
```

**`generators/<name>/generator.go`**

```go
// Package <pkg> <description>.
package <pkg>

import "github.com/version14/dot/pkg/dotapi"

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string    { return Manifest.Name }
func (g *Generator) Version() string { return Manifest.Version }

func (g *Generator) Generate(ctx *dotapi.Context) error {
	// TODO: implement — write <outputs_paths>.
	_ = ctx
	return nil
}
```

`<pkg>` is `<name>` with underscores removed and lowercased (e.g. `hello_world` → `helloworld`).

## Step 3b — `edit` workflow

Runs only AFTER Step 2 (including target + change_kind + follow-ups) succeeds.

1. Read `generators/<target>/generator.go`, `generators/<target>/manifest.go`, and `docs/contributor/generators/<target>.md`.
2. Apply the minimal diff for the chosen `change_kind`:
   - `add file output` — extend `Manifest.Outputs`, add the matching `CheckFileExists`, add the write call in `Generate`.
   - `change manifest field` — replace exactly one field; bump `Version` by patch only when the change is observable to the project on disk.
   - `add depends_on` — append to `Manifest.DependsOn`.
   - `add validator` — append a `Check` to the existing `Validator` (or add a new named `Validator` if none).
   - `add post-gen command` — append to `Manifest.PostGenerationCommands`.
   - `change generate behavior` — apply the user-described change inside `Generate` only; do NOT touch `Name()`, `Version()`, or `Manifest`.
3. Update the matching tables in `docs/contributor/generators/<target>.md` (Files written / Validators / Post-generation commands / Dependencies). Bump the Version row in Identity if `Manifest.Version` changed.
4. Run `go build ./...` then `make test` and report. If either fails, print the output and suggest the user run the failing command with `2>&1 | head -80`. Do NOT auto-fix. Stop and wait for user input.

## Step 3c — `generate` workflow

Runs only AFTER Step 2 succeeds.

1. Decide the writing strategy from the answers:

   | Outputs shape | Strategy |
   |---|---|
   | A single text/binary file | `raw` -> `ctx.State.WriteRaw(path, []byte(...))` |
   | A `package.json`-style structured file | `json` -> `doc := ctx.State.OpenJSON(path); doc.Set(...)` |
   | A `docker-compose.yml`-style structured file | `yaml` -> `doc := ctx.State.OpenYAML(path); doc.SetPath(...)` |
   | A `go.mod` | `gomod` -> `gomod := ctx.State.OpenGoMod(path); gomod.SetModule(...); gomod.AddRequire(...)` |
   | More than two files OR template substitution | `embed` -> embed the `files/` directory, render with `local.NewRenderer` (mirror `[generators/plugin_repo_skeleton/generator.go](../../../generators/plugin_repo_skeleton/generator.go)`) |

2. Produce both `generator.go` and `manifest.go` end-to-end with the chosen strategy. For each declared output, emit:
   - The corresponding write call in `Generate`.
   - A `CheckFileExists` (always) plus any `CheckJSONKeyExists` from the question pass.
3. If the strategy is `embed`, ALSO create `generators/<name>/files/` with one `.tmpl` file per declared output; populate each with a sensible starting body using `{{ .ProjectName }}` etc., wired to the data map built from `ctx.Answers`.
4. Run all of `init` steps 3–6 (register, doc, indexes, `go build`).

## After completion

Report:

- Created files (`generators/<name>/manifest.go`, `generators/<name>/generator.go`, optional `generators/<name>/files/...`, `docs/contributor/generators/<name>.md`).
- Modified files (`internal/cli/registry.go`, `docs/contributor/authoring-generators.md`, `docs/README.md`).
- `go build ./...` (and, in edit mode, `make test`) outcome.
- `make test-flows` outcome.
- Strategy chosen (raw / json / yaml / gomod / embed).

For end-to-end worked examples per mode and per strategy, see [examples.md](examples.md).
