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
| Add a `Description` to a Confirm question (multi-line context shown under the title) | `internal/flow/question.go` — `ConfirmQuestion.Description` is rendered by `internal/cli/prompt.go` via `huh.NewConfirm().Description(...)` | [flows/init.md](flows/init.md) — see `ts-backend-decorators-validation` for an example |
| Wire the decorator/validation/OpenAPI questions into the init flow | `flows/init.go` — `ts-backend-decorators-validation` (Confirm) and `ts-backend-validation-lib` (Option), plus the conditional generator selection in `resolveMonorepoGenerators` | [flows/init.md](flows/init.md) |
| Fix back-navigation in the TUI | `internal/cli/form_walker.go` — `buildHideFunc` | [architecture.md — flow engine](architecture.md#flow-engine) |
| Fix loop question rendering | `internal/cli/prompt.go` — `runLoopSubForms` | [authoring-flows.md — loops](authoring-flows.md#loops) |
| Understand question traversal order | `internal/flow/engine.go` | [architecture.md — flow engine](architecture.md#flow-engine) |

---

### Generators

| Task | Primary file(s) | Read first |
|------|----------------|------------|
| Add a new generator | Create `generators/<name>/` + register in `internal/cli/registry.go` | [authoring-generators.md](authoring-generators.md), then copy [generators/_template.md](generators/_template.md) |
| Understand the Express generator family | `generators/express_*`, `generators/auth_jwt_*`, `flows/init.go` | [express-backend-guide.md](express-backend-guide.md) — read this first |
| Understand the decorator-based validation / OpenAPI stack | `generators/zod_validation_deps/`, `express_decorators_core/`, `express_openapi_setup/`, `decorators_clean_arch_adapter/`, `decorators_mvc_adapter/`, `decorators_hexagonal_adapter/` | [user/decorators.md](../user/decorators.md), [express_decorators_core.md](generators/express_decorators_core.md), [express_openapi_setup.md](generators/express_openapi_setup.md), [flows/init.md — decorator-aware templating](flows/init.md#decorator-aware-templating) |
| Understand the classic JSDoc-driven Swagger | `generators/express_swagger_jsdoc/` — selected by `flows/init.go` when `ts-backend-decorators-validation = false`. Existing controllers (`express_server_entrypoint`, `auth_jwt_mvc_route`, `auth_jwt_clean_arch_module`) ship `@openapi` JSDoc blocks that swagger-jsdoc scans at boot | [express_swagger_jsdoc.md](generators/express_swagger_jsdoc.md), [user/decorators.md — Classic JSDoc path](../user/decorators.md#classic-jsdoc-path-writing-new-docs) |
| Add `@openapi` JSDoc to a new route handler | Copy the format from `auth_jwt_mvc_route/files/src/controllers/auth.controller.ts.tmpl` (decorators-OFF branch) | [user/decorators.md](../user/decorators.md), [express_swagger_jsdoc.md](generators/express_swagger_jsdoc.md) |
| Make an existing generator decorator-aware | `generators/<name>/generator.go` — read `slices.Contains(ctx.PreviousGens, "express_decorators_core")`, render templates with a `HasDecorators` flag (see `auth_jwt_mvc_route` and `auth_jwt_clean_arch_module` for the pattern) | [flows/init.md — decorator-aware templating](flows/init.md#decorator-aware-templating) |
| Understand the auth module (JWT) | `generators/auth_jwt_vanilla/`, `auth_jwt_users_schema/`, `auth_jwt_mvc_route/`, `auth_jwt_clean_arch_module/` | [auth_jwt_vanilla.md](generators/auth_jwt_vanilla.md), [auth_jwt_mvc_route.md](generators/auth_jwt_mvc_route.md), [auth_jwt_clean_arch_module.md](generators/auth_jwt_clean_arch_module.md) |
| Understand BetterAuth wiring | `generators/auth_better_auth/` — `lib/auth.ts` plus a direct `app.ts` mount of `toNodeHandler(auth)` (no intermediate route file) | [auth_better_auth.md](generators/auth_better_auth.md) |
| Understand the shared infrastructure generators | `generators/express_shared_errors/`, `express_error_middleware/`, `express_rate_limit/`, `express_auth_validators/` | [express_shared_errors.md](generators/express_shared_errors.md), [express_error_middleware.md](generators/express_error_middleware.md) |
| Fix a generator's file output | `generators/<name>/generator.go` | [generators/<name>.md](generators/) |
| Add a validator to a generator | `generators/<name>/manifest.go` — `Validators` field | [authoring-generators.md — validators](authoring-generators.md#validators) |
| Add a post-gen or test command | `generators/<name>/manifest.go` — `PostGenerationCommands` / `TestCommands` | [authoring-generators.md — commands](authoring-generators.md#postgenerationcommands-and-testcommands) |
| Opt a command **out** of the test-flow cache | `generators/<name>/manifest.go` — set `NoCache: true` on the `dotapi.Command{}` (commands are cacheable by default) | [authoring-generators.md — NoCache](authoring-generators.md#nocache-caching-is-opt-out), [test-flow.md — Case-level cache](test-flow.md#case-level-cache) |
| Force a full test-flow re-run | `go run ./tools/test-flow -no-cache` (or `rm -rf .test-flow-cache`) | [test-flow.md — Forcing a re-run](test-flow.md#forcing-a-re-run) |
| Run every test-flow case even after a failure | Pass `-keep-going` (the runner is fail-fast by default) | [test-flow.md — Fail-fast](test-flow.md#fail-fast-default) |
| Bump the test-flow cache schema (invalidate every entry) | `tools/test-flow/cache_persist.go` — increment `cacheSchemaVersion` | [test-flow.md — Case-level cache](test-flow.md#case-level-cache) |
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
