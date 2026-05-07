# Generator: `auth_jwt_mvc_route`

JWT auth route and controller for MVC architecture. Generates `src/routes/auth.route.ts` (register/login/me/refresh/logout) and `src/controllers/auth.controller.ts`. The controller is fully implemented when a Drizzle adapter has been generated; otherwise it returns 501 stubs.

---

## Identity

| Field | Value |
|-------|-------|
| Name | `auth_jwt_mvc_route` |
| Version | `0.1.0` |
| Package | `generators/auth_jwt_mvc_route` |

---

## Dependencies

| Generator | Why |
|-----------|-----|
| `auth_jwt_vanilla` | `src/shared/services/jwt.ts` and `src/shared/middlewares/auth.middleware.ts` must exist |

---

## Answers consumed

None directly. The generator reads `ctx.PreviousGens` at runtime:

| Probe | Effect on output |
|-------|------------------|
| Contains `drizzle_postgres_adapter` (`HasDB`) | Controller emits a real bcrypt + Drizzle implementation; otherwise emits 501 stubs |
| Contains `express_decorators_core` (`HasDecorators`) | Controller is emitted as a `@Controller({ prefix: '/auth' })` class with `@Post`/`@Get`/`@Auth`/`@Body`/`@ApiResponse` decorators; the route file becomes a no-op `export {}`; `app.ts` is patched to chain `.registerController(new AuthController())` onto the existing `DecoratorRouter` |

---

## Files written

| Path | Description |
|------|-------------|
| `src/routes/auth.route.ts` | When decorators OFF: Express router wiring POST /register, /login, GET /me, POST /refresh, /logout. When decorators ON: empty `export {}` (routes are registered via `DecoratorRouter` in `src/app.ts`) |
| `src/controllers/auth.controller.ts` | Functional handlers with JSDoc `@openapi` blocks on every endpoint (decorators OFF) — picked up by `express_swagger_jsdoc` so `/docs` is fully populated. With decorators ON: a `@Controller`-decorated `AuthController` class. When `drizzle_postgres_adapter` is present: real implementation; otherwise: 501 stubs |
| `src/__tests__/auth.db.test.ts` | Supertest DB-tests covering register/login/me/refresh/logout (only emitted when Drizzle is present) |

Also merges into (when Drizzle is present):

| Path | Keys added / updated |
|------|---------------------|
| `package.json` | `dependencies.bcryptjs`, `devDependencies.@types/bcryptjs` |

---

## Validators

| Check | Type | Passes when |
|-------|------|-------------|
| `src/routes/auth.route.ts` | `file_exists` | — |
| `src/controllers/auth.controller.ts` | `file_exists` | — |

---

## Post-generation commands

| Command | WorkDir | Notes |
|---------|---------|-------|
| `pnpm install` | project root | Deduped |

## Test commands

No TestCommands.

---

## Conflicts

None.
