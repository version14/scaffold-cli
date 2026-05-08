# Generator: `decorators_hexagonal_adapter`

Wires the decorator runtime into a Hexagonal project. Emits a sample `ExampleController` in `src/adapters/primary/http/controllers/`, schemas in `src/adapters/primary/http/schemas/`, and overwrites `src/app.ts` to bootstrap the `DecoratorRouter` + Swagger UI. Ships an end-to-end Vitest test.

---

## Identity

| Field | Value |
|-------|-------|
| Name | `decorators_hexagonal_adapter` |
| Version | `0.1.0` |
| Package | `generators/decorators_hexagonal_adapter` |

---

## Dependencies

| Generator | Why |
|-----------|-----|
| `backend_architecture_hexagonal` | Hexagonal skeleton must exist (`src/adapters/primary/http/...`, `src/core/...`) |
| `express_decorators_core` | Provides the decorators and the Express adapter |
| `express_openapi_setup` | Provides the registry / spec generator / Swagger mount |

---

## Answers consumed

None directly — `flows/init.go` selects this generator when both `ts-backend-architecture = hexagonal-architecture` and `ts-backend-decorators-validation = true`. (Hexagonal is currently disabled in the user-facing flow but the adapter is registered for future activation.)

---

## Files written

| Path | Description |
|------|-------------|
| `src/app.ts` | Decorator-aware bootstrap (imports `corsOptions` from `./shared/cors`, mounts `DecoratorRouter` at root, Swagger at `/docs`). The `src/shared/cors.ts` helper is provided by `express_server_entrypoint`; this generator reuses it as-is. |
| `src/adapters/primary/http/controllers/example.controller.ts` | `@Controller({ prefix: '/api/example' })` sample |
| `src/adapters/primary/http/schemas/example.schemas.ts` | Zod schemas |
| `src/__tests__/decorators-hexagonal.e2e.test.ts` | Supertest E2E |

---

## Validators

| Check | Type | Passes when |
|-------|------|-------------|
| `src/adapters/primary/http/controllers/example.controller.ts` | `file_exists` | — |
| `src/adapters/primary/http/schemas/example.schemas.ts` | `file_exists` | — |
| `src/__tests__/decorators-hexagonal.e2e.test.ts` | `file_exists` | — |

---

## Post-generation commands

No PostGenerationCommands.

## Test commands

The embedded E2E test runs via `pnpm exec vitest run e2e`.

---

## Conflicts

None.

---

## See also

- [generators/express_decorators_core.md](express_decorators_core.md)
- [generators/express_openapi_setup.md](express_openapi_setup.md)
- [generators/backend_architecture_hexagonal_architecture.md](backend_architecture_hexagonal_architecture.md)
- [generators/express_server_entrypoint.md](express_server_entrypoint.md) — owner of `src/shared/cors.ts`
- [docs/user/decorators.md](../../user/decorators.md)
