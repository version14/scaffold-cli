# Generator: `auth_jwt_clean_arch_module`

Full JWT authentication module for Clean Architecture. Generates a complete `src/modules/auth/` subtree with domain entities, repository interfaces, application use-cases, controllers, an auth route, and Drizzle repository implementations.

---

## Identity

| Field | Value |
|-------|-------|
| Name | `auth_jwt_clean_arch_module` |
| Version | `0.1.0` |
| Package | `generators/auth_jwt_clean_arch_module` |

---

## Dependencies

| Generator | Why |
|-----------|-----|
| `auth_jwt_vanilla` | JWT helpers and auth middleware must exist |
| `auth_jwt_users_schema` | `users` and `refresh_tokens` Drizzle tables must be defined |

---

## Answers consumed

None directly. The generator reads `ctx.PreviousGens` at runtime:

| Probe | Effect on output |
|-------|------------------|
| Contains `express_decorators_core` (`HasDecorators`) | Controller is emitted as a `@Controller({ prefix: '/auth' })` class with `@Post`/`@Get`/`@Auth`/`@Body`/`@ApiResponse` decorators (use cases injected at module scope as before); the route file becomes a no-op `export {}`; `app.ts` is patched to chain `.registerController(new AuthController())` onto the existing `DecoratorRouter` |

---

## Files written

| Path | Description |
|------|-------------|
| `src/modules/auth/domain/entities/user.entity.ts` | `UserEntity` interface |
| `src/modules/auth/domain/interfaces/user.repository.interface.ts` | `IUserRepository` interface |
| `src/modules/auth/domain/interfaces/refresh-token.repository.interface.ts` | `IRefreshTokenRepository` interface |
| `src/modules/auth/application/use-cases/login.use-case.ts` | `LoginUseCase` |
| `src/modules/auth/application/use-cases/register.use-case.ts` | `RegisterUseCase` |
| `src/modules/auth/application/use-cases/refresh.use-case.ts` | `RefreshUseCase` |
| `src/modules/auth/application/use-cases/logout.use-case.ts` | `LogoutUseCase` |
| `src/modules/auth/application/controllers/auth.controller.ts` | Functional handlers delegating to use-cases with JSDoc `@openapi` blocks on every endpoint (decorators OFF) — picked up by `express_swagger_jsdoc` so `/docs` is fully populated. With decorators ON: a `@Controller`-decorated `AuthController` class |
| `src/modules/auth/infrastructure/database/repositories/user.repository.ts` | Drizzle `UserRepository` |
| `src/modules/auth/infrastructure/database/repositories/refresh-token.repository.ts` | Drizzle `RefreshTokenRepository` |
| `src/routes/auth.route.ts` | When decorators OFF: Express router for /register, /login, /me, /refresh, /logout. When decorators ON: empty `export {}` (routes are registered through the `DecoratorRouter` in `src/app.ts`) |

Also merges into:

| Path | Keys added / updated |
|------|---------------------|
| `package.json` | `dependencies.bcryptjs`, `devDependencies.@types/bcryptjs` |

---

## Validators

| Check | Type | Passes when |
|-------|------|-------------|
| `src/modules/auth/domain/entities/user.entity.ts` | `file_exists` | — |
| `src/modules/auth/domain/interfaces/user.repository.interface.ts` | `file_exists` | — |
| `src/modules/auth/domain/interfaces/refresh-token.repository.interface.ts` | `file_exists` | — |
| `src/modules/auth/application/use-cases/login.use-case.ts` | `file_exists` | — |
| `src/modules/auth/application/use-cases/register.use-case.ts` | `file_exists` | — |
| `src/modules/auth/application/use-cases/refresh.use-case.ts` | `file_exists` | — |
| `src/modules/auth/application/use-cases/logout.use-case.ts` | `file_exists` | — |
| `src/modules/auth/application/controllers/auth.controller.ts` | `file_exists` | — |
| `src/modules/auth/infrastructure/database/repositories/user.repository.ts` | `file_exists` | — |
| `src/modules/auth/infrastructure/database/repositories/refresh-token.repository.ts` | `file_exists` | — |
| `src/routes/auth.route.ts` | `file_exists` | — |

---

## Post-generation commands

No PostGenerationCommands.

## Test commands

No TestCommands.

---

## Conflicts

None.
