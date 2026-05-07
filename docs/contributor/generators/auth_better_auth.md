# Generator: `auth_better_auth`

BetterAuth session-based authentication. Creates `src/lib/auth.ts` (auth instance with Drizzle adapter) and mounts the BetterAuth catch-all (`toNodeHandler(auth)`) at `/api/auth/*` directly inside `src/app.ts`. Also appends `cookie-parser`, `BETTER_AUTH_SECRET`, and `BETTER_AUTH_URL`.

---

## Identity

| Field | Value |
|-------|-------|
| Name | `auth_better_auth` |
| Version | `0.2.0` |
| Package | `generators/auth_better_auth` |

---

## Dependencies

| Generator | Why |
|-----------|-----|
| `drizzle_postgres_adapter` | BetterAuth uses the Drizzle adapter which requires an active `db` export |

---

## Answers consumed

None — selection is driven by `flows/init.go` (`ts-backend-auth-method = better-auth`).

---

## Files written

| Path | Description |
|------|-------------|
| `src/lib/auth.ts` | BetterAuth instance with Drizzle PG adapter and email/password enabled |
| `.env.example` | Appends `BETTER_AUTH_SECRET` and `BETTER_AUTH_URL` |

Also merges into:

| Path | Keys added / updated |
|------|---------------------|
| `package.json` | `dependencies.better-auth`, `dependencies.cookie-parser`, `devDependencies.@types/cookie-parser` |
| `src/app.ts` | Imports `cookieParser`, `toNodeHandler`, `auth`; adds `app.use(cookieParser())` and `app.all('/api/auth/*', toNodeHandler(auth))` directly (no intermediate route file) |

---

## Validators

| Check | Type | Passes when |
|-------|------|-------------|
| `src/lib/auth.ts` | `file_exists` | — |
| `dependencies.better-auth` in `package.json` | `json_key_exists` | — |

---

## Post-generation commands

No PostGenerationCommands. `pnpm install` is run by `typescript_base`.

## Test commands

No TestCommands.

---

## Decorator interaction

When `ts-backend-decorators-validation = true`, BetterAuth keeps its catch-all behaviour — `toNodeHandler` owns its routing under `/api/auth/*` and is **not** exposed through the decorator system. The decorator router and BetterAuth coexist on the same Express app; the rest of the API can use `@Controller`, `@Auth`, etc. The cookie-parser middleware injection works on the decorator-aware `app.ts` because it still contains `app.use(express.json());` as an anchor.

---

## Conflicts

None — but on the same scaffold you would not normally pair `auth_better_auth` with `auth_jwt_*` generators.

---

## See also

- [generators/auth_better_auth_schema.md](auth_better_auth_schema.md)
- [generators/auth_jwt_vanilla.md](auth_jwt_vanilla.md)
- [docs/user/decorators.md](../../user/decorators.md)
