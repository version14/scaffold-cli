# Generator: `express_decorators_core`

Framework-agnostic API decorators (`@Controller`, `@Get`/`@Post`/…, `@Body`, `@Query`, `@Params`, `@ApiResponse`, `@Auth`, `@RequiredHeaders`) plus the `RouterAdapter` interface and an Express implementation. Ships unit tests covering route registration, validation, auth, and required headers.

---

## Identity

| Field | Value |
|-------|-------|
| Name | `express_decorators_core` |
| Version | `0.1.0` |
| Package | `generators/express_decorators_core` |

---

## Dependencies

| Generator | Why |
|-----------|-----|
| `express_server_entrypoint` | `src/app.ts` must exist; the Express adapter relies on `express` being installed |
| `zod_validation_deps` | Decorators import Zod schemas and need `experimentalDecorators` + `emitDecoratorMetadata` |

---

## Answers consumed

None — selection is driven by `flows/init.go` based on `ts-backend-decorators-validation`.

---

## Files written

| Path | Description |
|------|-------------|
| `src/shared/decorators/metadata.ts` | Reflect-metadata helpers (`getRoutes`, `setController`, `setProtected`, …) |
| `src/shared/decorators/controller.decorator.ts` | `@Controller({ tag, prefix?, description? })` class decorator |
| `src/shared/decorators/route.decorators.ts` | `@Get`/`@Post`/`@Put`/`@Patch`/`@Delete`, plus `@Summary` and `@Description` overrides |
| `src/shared/decorators/validation.decorators.ts` | `@Body`/`@Query`/`@Params` method decorators |
| `src/shared/decorators/response.decorator.ts` | Stackable `@ApiResponse(status, description, schema?)` |
| `src/shared/decorators/auth.decorator.ts` | `@Auth()` marker (gates the route via `ExpressRouterAdapter`'s `authMiddleware` and adds `BearerAuth` to OpenAPI) |
| `src/shared/decorators/header.decorator.ts` | `@RequiredHeaders([...])` |
| `src/shared/decorators/router-adapter.ts` | `RouterAdapter<TNative>` interface + `RouteRegistration` type |
| `src/shared/decorators/express-router-adapter.ts` | Express implementation; wraps async handlers via `next(err)` |
| `src/shared/decorators/decorator-router.ts` | Reads metadata, calls `adapter.register(...)`, exposes `routes()` for the OpenAPI generator |
| `src/shared/decorators/index.ts` | Barrel re-exporting the public API |
| `src/shared/middlewares/validate-request.ts` | Express middleware that runs a Zod schema against `req.body` / `params` / `query` |
| `src/shared/decorators/__tests__/decorators.unit.test.ts` | Vitest unit tests (route registration, validation, auth, headers) |

---

## Validators

| Check | Type | Passes when |
|-------|------|-------------|
| `src/shared/decorators/index.ts` | `file_exists` | — |
| `src/shared/decorators/decorator-router.ts` | `file_exists` | — |
| `src/shared/decorators/router-adapter.ts` | `file_exists` | — |
| `src/shared/middlewares/validate-request.ts` | `file_exists` | — |

---

## Post-generation commands

No PostGenerationCommands. Installation is handled by `typescript_base` / `express_test_setup`.

## Test commands

The embedded `decorators.unit.test.ts` runs as part of the `pnpm exec vitest run unit` command from `express_test_setup`.

---

## Conflicts

None — but the decorator system only works when `zod_validation_deps` and `express_openapi_setup` are also generated. The init flow ensures they are paired.

---

## See also

- [generators/zod_validation_deps.md](zod_validation_deps.md)
- [generators/express_openapi_setup.md](express_openapi_setup.md)
- [generators/decorators_clean_arch_adapter.md](decorators_clean_arch_adapter.md)
- [generators/decorators_mvc_adapter.md](decorators_mvc_adapter.md)
- [generators/decorators_hexagonal_adapter.md](decorators_hexagonal_adapter.md)
- [docs/user/decorators.md](../../user/decorators.md)
