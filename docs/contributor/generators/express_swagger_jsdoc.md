# Generator: `express_swagger_jsdoc`

Classic JSDoc-driven Swagger / OpenAPI for Express scaffolds that opted **out** of the decorator-based stack. Adds `swagger-jsdoc`, builds a spec at boot by scanning `src/**/*.ts` for `@openapi` JSDoc blocks, and mounts `swagger-ui-express` at `/docs`.

OpenAPI is therefore **always** available on a generated Express app: either via the decorator runtime (`express_openapi_setup`) or via this generator — never both at the same time.

---

## Identity

| Field | Value |
|-------|-------|
| Name | `express_swagger_jsdoc` |
| Version | `0.1.0` |
| Package | `generators/express_swagger_jsdoc` |

---

## Dependencies

| Generator | Why |
|-----------|-----|
| `express_server_entrypoint` | `src/app.ts` must exist so the generator can inject the `mountSwagger(app)` call |

---

## Answers consumed

None directly. `flows/init.go` selects this generator when `ts-backend-framework = express` and `ts-backend-decorators-validation = false`.

---

## Files written

| Path | Description |
|------|-------------|
| `src/shared/swagger/swagger.config.ts` | `swagger-jsdoc` options and `swaggerSpec` (calls `swaggerJSDoc(...)` once at module load) |
| `src/shared/swagger/index.ts` | `mountSwagger(app, opts?)` — serves `/docs` (UI) and `/docs/openapi.json` (raw) |
| `src/shared/swagger/__tests__/swagger.unit.test.ts` | Vitest test that hits `/docs/openapi.json` and asserts the `/health` path is present (proves JSDoc scanning works end to end) |

Also merges into:

| Path | Keys added / updated |
|------|---------------------|
| `package.json` | `dependencies.swagger-jsdoc`, `dependencies.swagger-ui-express`, `devDependencies.@types/swagger-jsdoc`, `devDependencies.@types/swagger-ui-express` |
| `src/app.ts` | Imports `mountSwagger` and calls `mountSwagger(app)` before `export default app;` |

---

## How the spec is populated

`swagger-jsdoc` reads every `.ts`/`.js` file under `src/` looking for `/** @openapi … */` blocks. The dot generators that produce route handlers ship JSDoc OpenAPI annotations on every endpoint:

| Generator | Endpoint(s) documented |
|-----------|-----------------------|
| `express_server_entrypoint` | `GET /health` |
| `auth_jwt_mvc_route` (DB branch, decorators OFF) | `POST /auth/register`, `POST /auth/login`, `POST /auth/refresh`, `POST /auth/logout`, `GET /auth/me` |
| `auth_jwt_clean_arch_module` (decorators OFF) | Same five endpoints |

When you add new routes, drop a `@openapi` JSDoc block above the handler and `/docs` picks it up at the next boot — no codegen step.

---

## Validators

| Check | Type | Passes when |
|-------|------|-------------|
| `src/shared/swagger/swagger.config.ts` | `file_exists` | — |
| `src/shared/swagger/index.ts` | `file_exists` | — |
| `dependencies.swagger-jsdoc` in `package.json` | `json_key_exists` | — |
| `dependencies.swagger-ui-express` in `package.json` | `json_key_exists` | — |

---

## Post-generation commands

No PostGenerationCommands.

## Test commands

The embedded `swagger.unit.test.ts` runs as part of `pnpm exec vitest run unit` (declared by `express_test_setup`).

---

## Conflicts

None at the dependency-resolution level, but `flows/init.go` ensures only **one** of `express_swagger_jsdoc` and `express_openapi_setup` runs per scaffold (decorator choice is mutually exclusive).

---

## See also

- [generators/express_openapi_setup.md](express_openapi_setup.md) — decorator-driven counterpart
- [generators/express_decorators_core.md](express_decorators_core.md)
- [flows/init.md](../flows/init.md)
- [docs/user/decorators.md](../../user/decorators.md)
