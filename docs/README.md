# DOT Documentation

The docs are split into two audiences. If you are **using** DOT to scaffold projects, start in `docs/user/`. If you are **contributing** to DOT or writing a plugin, start in `docs/contributor/`.

---

## For users

| File | What it covers |
|------|----------------|
| [user/getting-started.md](user/getting-started.md) | Install, first scaffold, plugin management |
| [user/cli-reference.md](user/cli-reference.md) | Every command, flag, and exit code |
| [user/decorators.md](user/decorators.md) | Decorator-based validation and OpenAPI documentation for Express |

---

## For contributors

### Start here

| File | What it covers |
|------|----------------|
| [contributor/getting-started.md](contributor/getting-started.md) | Install tools, first build, IDE setup, repo structure |
| [contributor/navigation-guide.md](contributor/navigation-guide.md) | **"For this task, look in this file"** — task → file mapping |

### Deep-dive guides

| File | What it covers |
|------|----------------|
| [contributor/architecture.md](contributor/architecture.md) | Pipeline, flow engine, generator system, plugin system, `.dot/` schemas |
| [contributor/authoring-flows.md](contributor/authoring-flows.md) | Writing flow graphs: questions, branching, loops |
| [contributor/authoring-generators.md](contributor/authoring-generators.md) | Writing generators: VirtualProjectState, Manifest, validators, semver |
| [contributor/authoring-plugins.md](contributor/authoring-plugins.md) | Writing plugins: injections, fragments, publishing |
| [contributor/test-flow.md](contributor/test-flow.md) | End-to-end fixture testing with test-flow |

### Flow reference (`docs/contributor/flows/`)

One file per built-in flow. Each covers: question IDs, branching diagram, generator resolution, fixture examples.

| File | Flow |
|------|------|
| [contributor/flows/init.md](contributor/flows/init.md) | `init` — default project wizard (TypeScript / Express / decorators / DB / auth) |
| [contributor/flows/monorepo.md](contributor/flows/monorepo.md) | `monorepo` — general-purpose project wizard |
| [contributor/flows/fullstack.md](contributor/flows/fullstack.md) | `fullstack` — TypeScript + optional React + optional Go backend |
| [contributor/flows/microservices.md](contributor/flows/microservices.md) | `microservices` — N services via LoopQuestion |
| [contributor/flows/plugin-template.md](contributor/flows/plugin-template.md) | `plugin-template` — publishable plugin repository |

### Generator reference (`docs/contributor/generators/`)

One file per built-in generator. Each covers: answers consumed, files written, validators, commands.

| File | Generator |
|------|-----------|
| [contributor/generators/base_project.md](contributor/generators/base_project.md) | `base_project` — README, .gitignore, LICENSE |
| [contributor/generators/typescript_base.md](contributor/generators/typescript_base.md) | `typescript_base` — package.json, tsconfig.json |
| [contributor/generators/react_app.md](contributor/generators/react_app.md) | `react_app` — React + Vite application |
| [contributor/generators/biome_config.md](contributor/generators/biome_config.md) | `biome_config` — Biome formatter + linter |
| [contributor/generators/service_writer.md](contributor/generators/service_writer.md) | `service_writer` — one service per loop iteration |
| [contributor/generators/plugin_repo_skeleton.md](contributor/generators/plugin_repo_skeleton.md) | `plugin_repo_skeleton` — publishable plugin repo |
| [contributor/generators/backend_architecture_mvc_architecture.md](contributor/generators/backend_architecture_mvc_architecture.md) | `backend_architecture_mvc` — MVC backend structure |
| [contributor/generators/backend_architecture_clean_architecture.md](contributor/generators/backend_architecture_clean_architecture.md) | `backend_architecture_clean_architecture` — Clean Architecture backend structure |
| [contributor/generators/backend_architecture_hexagonal_architecture.md](contributor/generators/backend_architecture_hexagonal_architecture.md) | `backend_architecture_hexagonal` — Hexagonal Architecture backend structure |
| [contributor/generators/zod_validation_deps.md](contributor/generators/zod_validation_deps.md) | `zod_validation_deps` — Zod / zod-to-openapi / reflect-metadata deps + decorator tsconfig flags |
| [contributor/generators/express_decorators_core.md](contributor/generators/express_decorators_core.md) | `express_decorators_core` — `@Controller`/`@Get`/`@Body`/… decorators + `RouterAdapter` (Express impl) |
| [contributor/generators/express_openapi_setup.md](contributor/generators/express_openapi_setup.md) | `express_openapi_setup` — OpenAPI v3 spec generator + Swagger UI mount (decorator path) |
| [contributor/generators/express_swagger_jsdoc.md](contributor/generators/express_swagger_jsdoc.md) | `express_swagger_jsdoc` — classic JSDoc-driven Swagger; scans `src/**/*.ts` for `@openapi` blocks |
| [contributor/generators/decorators_clean_arch_adapter.md](contributor/generators/decorators_clean_arch_adapter.md) | `decorators_clean_arch_adapter` — wires `DecoratorRouter` into a Clean Architecture project |
| [contributor/generators/decorators_mvc_adapter.md](contributor/generators/decorators_mvc_adapter.md) | `decorators_mvc_adapter` — wires `DecoratorRouter` into an MVC project |
| [contributor/generators/decorators_hexagonal_adapter.md](contributor/generators/decorators_hexagonal_adapter.md) | `decorators_hexagonal_adapter` — wires `DecoratorRouter` into a Hexagonal project |

### Plugin reference (`docs/contributor/plugins/`)

One file per plugin. Each covers: plugin ID, injections (target, kind, question IDs), generators contributed, ResolveExtras logic.

| File | Plugin |
|------|--------|
| [contributor/plugins/biome_extras.md](contributor/plugins/biome_extras.md) | `biome_extras` — strict-mode Biome overlay |
| [contributor/plugins/example_plugin.md](contributor/plugins/example_plugin.md) | `example` — reference implementation |

### Templates

Copy these when adding new items. Each contains inline instructions (`<!-- HTML comments -->`) and `_placeholder_` values. Delete the instruction header and replace all placeholders before committing.

| Template | Use when |
|----------|---------|
| [contributor/generators/_template.md](contributor/generators/_template.md) | Adding a generator to `generators/` |
| [contributor/plugins/_template.md](contributor/plugins/_template.md) | Adding a plugin to `plugins/` or `examples/` |
| [contributor/flows/_template.md](contributor/flows/_template.md) | Adding a flow to `flows/` |
| [../tools/test-flow/testdata/_template.json](../tools/test-flow/testdata/_template.json) | Adding a test-flow fixture |

---

## Documentation rules

These rules keep the docs accurate as the codebase evolves.

### When to create a new file

| Trigger | Action |
|---------|--------|
| New generator in `generators/` | Create `docs/contributor/generators/<name>.md` (from template) + update table above |
| New plugin in `plugins/` or `examples/` | Create `docs/contributor/plugins/<name>.md` (from template) + update table above |
| New flow in `flows/` | Create `docs/contributor/flows/<id>.md` (from template) + update table above |
| New CLI command | Add to `docs/user/cli-reference.md` |
| New major subsystem | Add a section to `docs/contributor/architecture.md` |

### Which code changes require a doc update

| Change | Required update |
|--------|----------------|
| New CLI command or flag | `docs/user/cli-reference.md` |
| New flow | `docs/contributor/flows/<id>.md` + test fixture |
| Flow question IDs change | `docs/contributor/flows/<id>.md` + affected fixtures |
| New question type | `docs/contributor/authoring-flows.md` + `architecture.md` |
| New injection kind | `docs/contributor/authoring-plugins.md` |
| New exported type in `pkg/dotapi` or `pkg/dotplugin` | `authoring-generators.md` or `authoring-plugins.md` |
| Generator manifest fields change | `docs/contributor/generators/<name>.md` |
| Plugin injection IDs change | `docs/contributor/plugins/<name>.md` + affected fixtures |
| Pipeline step added/removed | `docs/contributor/architecture.md` |
| `.dot/` schema changes | `docs/contributor/architecture.md` (spec/manifest sections) |
| New `test-flow` flag | `docs/contributor/test-flow.md` |
| Install/uninstall mechanism changes | `docs/user/getting-started.md` |
| Setup script changes | `docs/contributor/getting-started.md` |

### Rules

- **One PR = one unit of documentation.** If a PR adds a flow, the flow doc and fixture are part of the same PR.
- **Keep docs close to the code they describe.** A manifest change and its generator doc update go in the same commit.
- **Prefer examples over prose.** A code snippet is worth 10 sentences.
- **No placeholder "TODO" paragraphs.** Either document it or leave the section out.
- **Who updates docs:** whoever writes the code. There is no separate documentation pass.
