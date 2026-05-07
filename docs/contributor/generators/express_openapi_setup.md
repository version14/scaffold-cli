# Generator: `express_openapi_setup`

OpenAPI v3 spec generation plus a Swagger UI mount that consume the route metadata produced by `DecoratorRouter`. Provides a registry helper, the spec aggregator, the Swagger mount, and unit tests.

---

## Identity

| Field | Value |
|-------|-------|
| Name | `express_openapi_setup` |
| Version | `0.1.0` |
| Package | `generators/express_openapi_setup` |

---

## Dependencies

| Generator | Why |
|-----------|-----|
| `express_decorators_core` | The spec aggregator imports `RegisteredRoute` from `../decorators` |

---

## Answers consumed

None.

---

## Files written

| Path | Description |
|------|-------------|
| `src/shared/openapi/registry.ts` | `createRegistry()` (fresh, isolated for tests) and `getSharedRegistry()` (lazy singleton); also re-exports `z` extended with `@asteasolutions/zod-to-openapi` |
| `src/shared/openapi/spec-generator.ts` | `buildOpenApiSpec({ info, servers?, routes, registry? })` — converts decorator metadata to OpenAPI v3 |
| `src/shared/openapi/swagger.ts` | `mountSwagger(app, spec, opts?)` — serves `/docs` (UI) and `/docs/openapi.json` (raw) |
| `src/shared/openapi/index.ts` | Barrel re-exporting the public API |
| `src/shared/openapi/__tests__/spec.unit.test.ts` | Vitest tests for paths, tags, BearerAuth, default 200 fallback, served `/docs/openapi.json` |

---

## Validators

| Check | Type | Passes when |
|-------|------|-------------|
| `src/shared/openapi/spec-generator.ts` | `file_exists` | — |
| `src/shared/openapi/swagger.ts` | `file_exists` | — |
| `src/shared/openapi/registry.ts` | `file_exists` | — |

---

## Post-generation commands

No PostGenerationCommands.

## Test commands

The embedded `spec.unit.test.ts` runs as part of `pnpm exec vitest run unit` from `express_test_setup`.

---

## Conflicts

None.

---

## See also

- [generators/express_decorators_core.md](express_decorators_core.md)
- [docs/user/decorators.md](../../user/decorators.md)
