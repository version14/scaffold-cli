# Generator: `decorators_clean_arch_adapter`

Wires the decorator runtime into a Clean Architecture project. Emits a sample `ExampleController` in the application layer, a Zod schema module in `application/validators/`, and overwrites `src/app.ts` to bootstrap the `DecoratorRouter` + Swagger UI. Ships an end-to-end Vitest test that boots the real app.

---

## Identity

| Field | Value |
|-------|-------|
| Name | `decorators_clean_arch_adapter` |
| Version | `0.1.0` |
| Package | `generators/decorators_clean_arch_adapter` |

---

## Dependencies

| Generator | Why |
|-----------|-----|
| `backend_architecture_clean_architecture` | Module skeleton must exist (`src/modules/example/application/...`) |
| `express_decorators_core` | Provides the decorators and the Express adapter |
| `express_openapi_setup` | Provides the registry / spec generator / Swagger mount |

---

## Answers consumed

None directly — `flows/init.go` selects this generator when both `ts-backend-architecture = clean-architecture` and `ts-backend-decorators-validation = true`.

---

## Files written

| Path | Description |
|------|-------------|
| `src/app.ts` | Overwritten with a decorator-aware bootstrap (mounts `DecoratorRouter` at the root, builds the OpenAPI spec, mounts Swagger at `/docs`) |
| `src/modules/example/application/controllers/example.controller.ts` | `@Controller({ prefix: '/api/example' })` sample controller demonstrating `@Get`, `@Post`, `@Body`, `@Params`, `@ApiResponse` |
| `src/modules/example/application/validators/example.schemas.ts` | Zod request/response schemas for the example controller |
| `src/__tests__/decorators-clean.e2e.test.ts` | Supertest E2E covering 200/400 paths and `/docs/openapi.json` |

---

## Validators

| Check | Type | Passes when |
|-------|------|-------------|
| `src/modules/example/application/controllers/example.controller.ts` | `file_exists` | — |
| `src/modules/example/application/validators/example.schemas.ts` | `file_exists` | — |
| `src/__tests__/decorators-clean.e2e.test.ts` | `file_exists` | — |

---

## Post-generation commands

No PostGenerationCommands. Tooling is set up by upstream generators.

## Test commands

The embedded E2E test is matched by `pnpm exec vitest run e2e` (declared by `express_test_setup`).

---

## Conflicts

None — but the init flow only includes one of the three architecture-specific decorator adapters at a time.

---

## See also

- [generators/express_decorators_core.md](express_decorators_core.md)
- [generators/express_openapi_setup.md](express_openapi_setup.md)
- [generators/backend_architecture_clean_architecture.md](backend_architecture_clean_architecture.md)
- [docs/user/decorators.md](../../user/decorators.md)
