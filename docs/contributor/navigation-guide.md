# Navigation Guide

This guide answers the question **"where do I look?"** for any change you might make to DOT. Find your task in the table, open the linked files, read the linked doc.

---

## By task

### CLI & commands

| Task | Primary file(s) | Read first |
|------|----------------|------------|
| Add a new top-level command | `internal/cli/command.go` → add a `case` in `Dispatch` | [architecture.md — scaffold pipeline](architecture.md#the-scaffold-pipeline) |
| Change a command's flags | `internal/cli/command.go` → the matching `runXxx` function | [user/cli-reference.md](../user/cli-reference.md) (update it too) |
| Fix the scaffold pipeline | `internal/cli/runner.go` — `Scaffold()` | [architecture.md — scaffold pipeline](architecture.md#the-scaffold-pipeline) |
| Fix post-gen or test command execution | `internal/cli/runner.go` — `PlanPostGenCommands`, `PlanTestCommands` + `internal/commands/` | [architecture.md — command execution](architecture.md#command-execution) |
| Fix the spinner or progress output | `internal/cli/spinner.go` | — |
| Fix `dot update` | `internal/cli/update.go` | — |
| Fix `dot doctor` | `internal/cli/doctor.go` | — |
| Fix `dot plugin install` | `internal/cli/plugin_cmd.go` + `internal/plugin/installer.go` | [architecture.md — plugin system](architecture.md#plugin-system) |

---

### Flows

| Task | Primary file(s) | Read first |
|------|----------------|------------|
| Add a new built-in flow | Create `flows/<name>.go` + register in `flows/registry.go` | [authoring-flows.md](authoring-flows.md), then copy [flows/_template.md](flows/_template.md) |
| Change a question's label or default | `flows/<flow>.go` — the relevant `&flow.XxxQuestion{}` struct | [flows/](flows/) (update the flow doc too) |
| Add a branch to an existing flow | `flows/<flow>.go` — add a new `*flow.Next` edge | [authoring-flows.md — branching](authoring-flows.md#branching) |
| Add a new question type | `internal/flow/question.go` + `internal/cli/form_walker.go` (Huh rendering) | [architecture.md — flow engine](architecture.md#flow-engine) |
| Fix back-navigation in the TUI | `internal/cli/form_walker.go` — `buildHideFunc` | [architecture.md — flow engine](architecture.md#flow-engine) |
| Fix loop question rendering | `internal/cli/prompt.go` — `runLoopSubForms` | [authoring-flows.md — loops](authoring-flows.md#loops) |
| Understand question traversal order | `internal/flow/engine.go` | [architecture.md — flow engine](architecture.md#flow-engine) |

---

### Generators

| Task | Primary file(s) | Read first |
|------|----------------|------------|
| Add a new generator | Create `generators/<name>/` + register in `internal/cli/registry.go` | [authoring-generators.md](authoring-generators.md), then copy [generators/_template.md](generators/_template.md) |
| Understand the Express generator family | `generators/express_*`, `generators/auth_jwt_*`, `flows/monorepo.go` | [express-backend-guide.md](express-backend-guide.md) — read this first |
| Understand the auth module (JWT) | `generators/auth_jwt_vanilla/`, `auth_jwt_users_schema/`, `auth_jwt_mvc_route/`, `auth_jwt_clean_arch_module/` | [auth_jwt_vanilla.md](generators/auth_jwt_vanilla.md), [auth_jwt_mvc_route.md](generators/auth_jwt_mvc_route.md), [auth_jwt_clean_arch_module.md](generators/auth_jwt_clean_arch_module.md) |
| Understand the shared infrastructure generators | `generators/express_shared_errors/`, `express_error_middleware/`, `express_rate_limit/`, `express_auth_validators/` | [express_shared_errors.md](generators/express_shared_errors.md), [express_error_middleware.md](generators/express_error_middleware.md) |
| Fix a generator's file output | `generators/<name>/generator.go` | [generators/<name>.md](generators/) |
| Add a validator to a generator | `generators/<name>/manifest.go` — `Validators` field | [authoring-generators.md — validators](authoring-generators.md#validators) |
| Add a post-gen or test command | `generators/<name>/manifest.go` — `PostGenerationCommands` / `TestCommands` | [authoring-generators.md — commands](authoring-generators.md#postgenerationcommands-and-testcommands) |
| Fix dependency ordering | `generators/<name>/manifest.go` — `DependsOn` + `internal/generator/sorter.go` | [architecture.md — generator pipeline](architecture.md#generator-pipeline) |
| Fix the topological sort | `internal/generator/sorter.go` | [architecture.md — generator pipeline](architecture.md#generator-pipeline) |
| Fix transitive dep resolution | `internal/generator/resolver.go` | — |
| Write to a JSON file cooperatively | `ctx.State.UpdateJSON(...)` in `generator.go` | [authoring-generators.md — writing files](authoring-generators.md#writing-files) |

---

### Plugins

| Task | Primary file(s) | Read first |
|------|----------------|------------|
| Write a new in-tree plugin | Create `plugins/<id>/plugin.go` + import in `cmd/dot/main.go` | [authoring-plugins.md](authoring-plugins.md), then copy [plugins/_template.md](plugins/_template.md) |
| Fix plugin injection (Replace/AddOption/InsertAfter) | `internal/flow/hook.go` + `internal/cli/form_walker.go` | [authoring-plugins.md — injection kinds](authoring-plugins.md#injection-kinds) |
| Fix `ResolveExtras` not adding generators | `internal/cli/runner.go` — the `for _, p := range opts.Plugins` loop | [architecture.md — plugin system](architecture.md#plugin-system) |
| Fix plugin loading from `~/.dot/plugins/` | `internal/plugin/loader.go` | [architecture.md — installed plugins](architecture.md#installed-plugins) |
| Fix remote install (git clone staging) | `internal/plugin/installer.go` | [user/getting-started.md — manage plugins](../user/getting-started.md#manage-plugins) |
| Understand the Fragment registry | `internal/flow/fragment.go` | [authoring-plugins.md — fragment registry](authoring-plugins.md#fragment-registry) |

---

### Skills (AI assistant prompts)

| Task | Primary file(s) | Read first |
|------|----------------|------------|
| Add a new AI skill | Create `.claude/skills/<name>/SKILL.md` + `examples.md` | [authoring-skills.md](authoring-skills.md) |
| Modify an existing skill | Edit `.claude/skills/<name>/SKILL.md` or `examples.md` | [authoring-skills.md — modifying](authoring-skills.md#modifying-an-existing-skill) |
| Sync skills to Cursor | Run `sync-skills` skill | [authoring-skills.md — sync](authoring-skills.md#the-claude--cursor-sync) |
| Wire a skill into routing | Add a line to `CLAUDE.md` Key routing rules | [authoring-skills.md — routing](authoring-skills.md#wiring-a-skill-into-claudemd) |

---

### Tests & fixtures

| Task | Primary file(s) | Read first |
|------|----------------|------------|
| Add an end-to-end test for a flow | Create `tools/test-flow/testdata/<name>.json` | [test-flow.md](test-flow.md), copy [testdata/_template.json](../../tools/test-flow/testdata/_template.json) |
| Add a unit test for a package | Create `<package>/<file>_test.go` | Go standard testing patterns |
| Fix a failing end-to-end fixture | `tools/test-flow/testdata/<name>.json` — check answer keys | [test-flow.md — fixture schema](test-flow.md#fixture-schema-reference) |
| Debug which questions a flow visits | Run `make test-flows -only <fixture>` — look at "verify visited" output | [test-flow.md — expected_visited](test-flow.md#expected_visited) |
| Add plugin answers to a fixture | Add `"plugin_id.question_id": value` to `answers` | [test-flow.md — plugin fixtures](test-flow.md#plugin-injection-fixtures) |

---

### Virtual filesystem & state

| Task | Primary file(s) | Read first |
|------|----------------|------------|
| Fix file persistence to disk | `internal/state/persist.go` | [architecture.md — virtual filesystem](architecture.md#virtual-filesystem) |
| Fix JSON merging between generators | `internal/state/json.go` | [authoring-generators.md — writing files](authoring-generators.md#json) |
| Fix YAML merging | `internal/state/yaml.go` | [authoring-generators.md — writing files](authoring-generators.md#yaml) |
| Fix `go.mod` generation | `internal/state/gomod.go` | [authoring-generators.md — writing files](authoring-generators.md#gomod) |

---

### Versioning & semver

| Task | Primary file(s) | Read first |
|------|----------------|------------|
| Fix semver parsing | `internal/versioning/semver.go` | [authoring-generators.md — versioning](authoring-generators.md#versioning-and-semver-constraints) |
| Fix `dot doctor` version drift detection | `internal/cli/doctor.go` + `internal/versioning/` | [user/cli-reference.md — doctor](../user/cli-reference.md#dot-doctor) |

---

### Documentation

| Task | Primary file(s) | Read first |
|------|----------------|------------|
| Add docs for a new generator | Create `docs/contributor/generators/<name>.md` | Copy [generators/_template.md](generators/_template.md) |
| Add docs for a new plugin | Create `docs/contributor/plugins/<name>.md` | Copy [plugins/_template.md](plugins/_template.md) |
| Add docs for a new flow | Create `docs/contributor/flows/<id>.md` | Copy [flows/_template.md](flows/_template.md) |
| Update the user-facing CLI reference | `docs/user/cli-reference.md` | — |
| Update the doc index | `docs/README.md` | [docs/README.md — documentation rules](../README.md#documentation-rules) |

---

## By area

### "I'm looking at `internal/cli/`"

This package is the integration layer. It:
- Dispatches CLI commands (`command.go`)
- Drives the full scaffold pipeline (`runner.go`)
- Renders the Huh TUI form (`prompt.go`, `form_walker.go`)
- Manages the spinner and quiet output (`spinner.go`)
- Builds the runtime bundle (`runtime.go`)

Start with `runner.go` to understand the pipeline, then `command.go` to understand command routing.

### "I'm looking at `internal/flow/`"

This package owns the question graph and traversal. No terminal I/O lives here.

- `question.go` — question types and their `Next()` logic
- `engine.go` — traversal loop and hook application
- `hook.go` — `HookRegistry`, injection validation
- `fragment.go` — `FragmentRegistry`
- `next.go` — the `Next` edge struct

### "I'm looking at `internal/generator/`"

This package resolves and executes generator invocations.

- `resolver.go` — transitive dep expansion + dedup
- `sorter.go` — stable Kahn topological sort
- `executor.go` — calls `Generate()` on each invocation
- `validator.go` — runs manifest validators against the virtual state

### "I'm looking at `pkg/`"

These are the stable public APIs. **Do not break them without a major version bump.** Changes here affect every plugin and generator author.

- `pkg/dotapi` — the generator contract (`Generator`, `Manifest`, `Context`, `Logger`)
- `pkg/dotplugin` — the plugin author API (re-exports from `internal/`)

---

## For first-time contributors

**Read in this order:**

1. [getting-started.md](getting-started.md) — get a green build (you are reading this now / should have done this already)
2. [architecture.md](architecture.md) — 15-minute read, understand the pipeline
3. Open an issue or pick a `good first issue` label on GitHub
4. Find your task in this navigation guide
5. Make the change, run `make validate && make test-flows`
6. Submit a PR following [CONTRIBUTING.md](../../CONTRIBUTING.md)

**Common beginner mistakes:**
- Editing `pkg/dotapi` without checking what all generators import — that's a breaking change.
- Forgetting to add a test fixture after changing a flow's questions.
- Forgetting to update `docs/contributor/generators/<name>.md` after changing a manifest.
- Adding a plugin injection without updating fixtures that use the target flow.
