# Generator: `decorators_mvc_adapter`

Wires the decorator runtime into an MVC project. Emits a sample `ExampleController` in `src/controllers/`, schemas in `src/shared/validators/`, and overwrites `src/app.ts` to bootstrap the `DecoratorRouter` + Swagger UI. Ships an end-to-end Vitest test.

---

## Identity

| Field | Value |
|-------|-------|
| Name | `decorators_mvc_adapter` |
| Version | `0.1.0` |
| Package | `generators/decorators_mvc_adapter` |

---

## Dependencies

| Generator | Why |
|-----------|-----|
| `backend_architecture_mvc` | MVC skeleton must exist (`src/controllers/`, `src/shared/validators/`) |
| `express_decorators_core` | Provides the decorators and the Express adapter |
| `express_openapi_setup` | Provides the registry / spec generator / Swagger mount |

---

## Answers consumed

None directly — `flows/init.go` selects this generator when both `ts-backend-architecture = mvc-architecture` and `ts-backend-decorators-validation = true`.

---

## Files written

| Path | Description |
|------|-------------|
| `src/app.ts` | Decorator-aware bootstrap (mounts `DecoratorRouter` at root, Swagger at `/docs`) |
| `src/controllers/example.controller.ts` | `@Controller({ prefix: '/api/example' })` sample |
| `src/shared/validators/example.schemas.ts` | Zod schemas |
| `src/__tests__/decorators-mvc.e2e.test.ts` | Supertest E2E |

---

## Validators

| Check | Type | Passes when |
|-------|------|-------------|
| `src/controllers/example.controller.ts` | `file_exists` | — |
| `src/shared/validators/example.schemas.ts` | `file_exists` | — |
| `src/__tests__/decorators-mvc.e2e.test.ts` | `file_exists` | — |

---

## Post-generation commands

No PostGenerationCommands.

## Test commands

The embedded E2E test runs via `pnpm exec vitest run e2e`.

---

## Conflicts

None — only one architecture-specific decorator adapter runs per scaffold.

---

## See also

- [generators/express_decorators_core.md](express_decorators_core.md)
- [generators/express_openapi_setup.md](express_openapi_setup.md)
- [generators/backend_architecture_mvc_architecture.md](backend_architecture_mvc_architecture.md)
- [docs/user/decorators.md](../../user/decorators.md)
