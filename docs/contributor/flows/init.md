# Flow: `init`

The default project wizard. Walks the user from a project name through monorepo style, language stack, framework, architecture, decorator-based validation/OpenAPI, formatter/linter, optional database, and optional authentication. Currently focused on TypeScript + Express scaffolds.

---

## Identity

| Field | Value |
|-------|-------|
| ID | `init` |
| Title | Init / Project Wizard |
| File | `flows/init.go` |
| Root question | `project_name` |

---

## Questions

| ID | Type | Label | Options / Default |
|----|------|-------|-------------------|
| `project_name` | Text | "Project name" | Default: `"my-project"` |
| `monorepo_type` | Option | "Monorepo style" | `single` (Turborepo currently disabled) |
| `stack` | Option | "Primary language stack" | `typescript` (Go disabled) |
| `ts-backend-framework` | Option | "Library / Framework" | `express` |
| `ts-backend-architecture` | Option | "Choose your architecture." | `clean-architecture`, `mvc-architecture` (Hexagonal disabled in UI but the generator is registered) |
| `ts-backend-decorators-validation` | Confirm | "Use decorator-based validation and OpenAPI documentation?" | Default: `true`. The description in `flows/init.go` clarifies that **OpenAPI/Swagger is always served at `/docs`** — the choice is between the decorator runtime (Yes) or classic JSDoc-driven `swagger-jsdoc` (No) |
| `ts-backend-validation-lib` | Option | "Validation library" | `zod` (only option today; the constant `ValidationLibZod` keeps the door open for more) — visited only when the decorator question is `true` |
| `ts-backend-formatter` | Option | "Choose a formatter." | `biome`, `prettier` |
| `ts-backend-linter` | Option | "Choose a linter." | `biome`, `prettier` |
| `enable-db` | Confirm | "Link a database to this project?" | Default: `false` |
| `ts-backend-db-type` | Option | "Choose a database." | `postgres` (visited only when `enable-db = true`) |
| `ts-backend-orm` | Option | "Choose an ORM." | `drizzle` (visited only when `enable-db = true`) |
| `enable-auth` | Confirm | "Enable authentication?" | Default: `false` |
| `ts-backend-auth-method` | Option | "Choose an auth method." | `jwt`, `better-auth` (visited only when `enable-auth = true`) |
| `confirm-generate` | Confirm | "Generate the project now?" | Default: `true` |

---

## Question graph

```
project_name
  └── monorepo_type
        └── stack
              └── ts-backend-framework
                    └── ts-backend-architecture
                          └── ts-backend-decorators-validation
                                ├── true  → ts-backend-validation-lib → ts-backend-formatter
                                └── false → ts-backend-formatter
                                                └── ts-backend-linter
                                                      └── enable-db
                                                            ├── true  → ts-backend-db-type → ts-backend-orm → enable-auth
                                                            └── false → enable-auth
                                                                            ├── true  → ts-backend-auth-method → confirm-generate
                                                                            └── false → confirm-generate
```

---

## Generator resolution

`resolveMonorepoGenerators(spec)` in `flows/init.go` produces the ordered list of generator invocations. Order matters: dependents come after their dependencies.

| Condition | Generators added |
|-----------|------------------|
| Always | `base_project` |
| `stack = typescript` | `typescript_base` |
| `ts-backend-framework = express` | `express_server_entrypoint`, `express_server_typescript_deps`, `express_node_tsconfig`, `express_shared_errors`, `express_error_middleware`, `express_rate_limit`, `express_test_setup` |
| `ts-backend-architecture = clean-architecture` | `backend_architecture_clean_architecture` |
| `ts-backend-architecture = mvc-architecture` | `backend_architecture_mvc` |
| `ts-backend-architecture = hexagonal-architecture` | `backend_architecture_hexagonal` |
| `ts-backend-decorators-validation = true` and `ts-backend-framework = express` | `zod_validation_deps`, `express_decorators_core`, `express_openapi_setup` |
| `ts-backend-decorators-validation = true` and `ts-backend-architecture = clean-architecture` | `decorators_clean_arch_adapter` |
| `ts-backend-decorators-validation = true` and `ts-backend-architecture = mvc-architecture` | `decorators_mvc_adapter` |
| `ts-backend-decorators-validation = true` and `ts-backend-architecture = hexagonal-architecture` | `decorators_hexagonal_adapter` |
| `ts-backend-decorators-validation = false` and `ts-backend-framework = express` | `express_swagger_jsdoc` (classic Swagger that scans JSDoc `@openapi` blocks — `/docs` is always served) |
| `ts-backend-formatter = prettier` | `prettier_config`, `prettier_typescript_deps`, `prettier_express_rules` |
| `ts-backend-formatter = biome` | `biome_config` |
| `enable-db = true` and `ts-backend-db-type = postgres` | `postgres_docker_compose`, `postgres_env_example` |
| `enable-db = true` and `ts-backend-orm = drizzle` | `drizzle_config_base`, `drizzle_typescript_deps`, `drizzle_postgres_adapter` |
| `enable-auth = true` | `express_auth_validators` |
| `enable-auth = true` and `ts-backend-auth-method = better-auth` | `auth_better_auth`, `auth_better_auth_schema` (auto-adds Postgres + Drizzle if not already enabled) |
| `enable-auth = true` and `ts-backend-auth-method = jwt` | `auth_jwt_vanilla` (+ `auth_jwt_users_schema` when DB present) |
| `enable-auth = true`, JWT, MVC | `auth_jwt_mvc_route` |
| `enable-auth = true`, JWT, Clean, with DB | `auth_jwt_clean_arch_module` |

The decorator generators run **before** the formatter/db/auth steps so later generators can detect them via `ctx.PreviousGens` and adapt their templates.

---

## Decorator-aware templating

Several existing generators read `slices.Contains(ctx.PreviousGens, "express_decorators_core")` and switch their output:

| Generator | Effect when decorators are on |
|-----------|------------------------------|
| `auth_jwt_vanilla` | Patches `app.ts` to pass `{ authMiddleware }` to `ExpressRouterAdapter` so `@Auth()` routes are gated |
| `auth_jwt_mvc_route` | Emits `AuthController` as a decorated class; route file becomes a no-op; `app.ts` chains `.registerController(new AuthController())` onto the `DecoratorRouter` |
| `auth_jwt_clean_arch_module` | Same pattern, controller in `src/modules/auth/application/controllers/` |
| `auth_better_auth` | Unchanged; BetterAuth keeps its catch-all `toNodeHandler(auth)` mounted directly in `app.ts` (the decorator router and BetterAuth coexist) |

---

## Fixture examples

Located under `tools/test-flow/testdata/`. Two illustrative ones:

**Decorators on, Clean Architecture, no DB** (`202605070101_express_clean_arch_decorators_zod.json`):

```json
{
  "name": "express_clean_arch_decorators_zod",
  "flow_id": "init",
  "answers": {
    "project_name": "my-app",
    "monorepo_type": "single",
    "stack": "typescript",
    "ts-backend-framework": "express",
    "ts-backend-architecture": "clean-architecture",
    "ts-backend-decorators-validation": true,
    "ts-backend-validation-lib": "zod",
    "ts-backend-formatter": "prettier",
    "ts-backend-linter": "prettier",
    "enable-db": false,
    "confirm-generate": true
  },
  "skip_post_commands": false,
  "skip_test_commands": false
}
```

**Decorators on, MVC + Postgres + JWT** (`202605070104_express_mvc_decorators_postgres_jwt.json`):

```json
{
  "name": "express_mvc_decorators_postgres_jwt",
  "flow_id": "init",
  "answers": {
    "project_name": "my-app",
    "monorepo_type": "single",
    "stack": "typescript",
    "ts-backend-framework": "express",
    "ts-backend-architecture": "mvc-architecture",
    "ts-backend-decorators-validation": true,
    "ts-backend-validation-lib": "zod",
    "ts-backend-formatter": "biome",
    "ts-backend-linter": "biome",
    "enable-db": true,
    "ts-backend-db-type": "postgres",
    "ts-backend-orm": "drizzle",
    "enable-auth": true,
    "ts-backend-auth-method": "jwt",
    "confirm-generate": true
  }
}
```

Existing pre-decorator scenarios were migrated to set `ts-backend-decorators-validation: false` explicitly (the question is required, so the field is mandatory in every Express scenario).

---

## Source

`flows/init.go` — `InitFlow()` builds the question graph; `resolveMonorepoGenerators()` does the conditional generator selection.

## See also

- [generators/express_decorators_core.md](../generators/express_decorators_core.md)
- [generators/express_openapi_setup.md](../generators/express_openapi_setup.md)
- [generators/zod_validation_deps.md](../generators/zod_validation_deps.md)
- [generators/decorators_clean_arch_adapter.md](../generators/decorators_clean_arch_adapter.md)
- [generators/decorators_mvc_adapter.md](../generators/decorators_mvc_adapter.md)
- [generators/decorators_hexagonal_adapter.md](../generators/decorators_hexagonal_adapter.md)
- [generators/auth_jwt_vanilla.md](../generators/auth_jwt_vanilla.md)
- [generators/auth_jwt_mvc_route.md](../generators/auth_jwt_mvc_route.md)
- [generators/auth_jwt_clean_arch_module.md](../generators/auth_jwt_clean_arch_module.md)
- [generators/auth_better_auth.md](../generators/auth_better_auth.md)
- [docs/user/decorators.md](../../user/decorators.md)
