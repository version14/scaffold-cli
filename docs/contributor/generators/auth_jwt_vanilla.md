# Generator: `auth_jwt_vanilla`

Vanilla JWT authentication. Creates `src/shared/services/jwt.ts` (sign/verify helpers) and `src/shared/middlewares/auth.middleware.ts` (Bearer token guard). Adds `jsonwebtoken` + `@types/jsonwebtoken` to `package.json` and appends `JWT_SECRET`/`JWT_EXPIRES_IN` to `.env.example`.

---

## Identity

| Field | Value |
|-------|-------|
| Name | `auth_jwt_vanilla` |
| Version | `0.1.0` |
| Package | `generators/auth_jwt_vanilla` |

---

## Dependencies

| Generator | Why |
|-----------|-----|
| `express_server_entrypoint` | `.env.example` and `src/` directory must exist |

---

## Answers consumed

None.

---

## Files written

| Path | Description |
|------|-------------|
| `src/shared/services/jwt.ts` | `signToken`, `signRefreshToken`, and `verifyToken<T>` helpers backed by `process.env.JWT_SECRET` |
| `src/shared/middlewares/auth.middleware.ts` | Express middleware that validates `Authorization: Bearer <token>` headers |
| `.env.example` | Appends `JWT_SECRET` and `JWT_EXPIRES_IN` |

Also merges into:

| Path | Keys added / updated |
|------|---------------------|
| `package.json` | `dependencies.jsonwebtoken`, `devDependencies.@types/jsonwebtoken` |

---

## Validators

| Check | Type | Passes when |
|-------|------|-------------|
| `src/shared/services/jwt.ts` | `file_exists` | — |
| `src/shared/middlewares/auth.middleware.ts` | `file_exists` | — |
| `dependencies.jsonwebtoken` in `package.json` | `json_key_exists` | — |

---

## Post-generation commands

| Command | WorkDir | Notes |
|---------|---------|-------|
| `pnpm install` | project root | Deduped |

## Test commands

No TestCommands.

---

## Decorator interaction

When `express_decorators_core` ran earlier in the pipeline, this generator additionally patches `src/app.ts` to wire the JWT middleware into `ExpressRouterAdapter`:

```ts
import { authMiddleware } from './shared/middlewares/auth.middleware';
// ...
new ExpressRouterAdapter({ authMiddleware })
```

That makes every `@Auth()`-decorated route gated by JWT verification automatically. Detection is done via `slices.Contains(ctx.PreviousGens, "express_decorators_core")` — there is no extra answer to set.

---

## Conflicts

None.
