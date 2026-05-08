# Generator: `express_server_entrypoint`

Creates the Express TypeScript source files: `src/index.ts` (server bootstrap with `PORT`), `src/app.ts` (Express app with `/health` route), and `src/shared/cors.ts` (env-driven CORS configuration helper). Also writes the base `.env.example` that downstream generators append to — seeded with `PORT` and `CORS_ORIGIN`.

---

## Identity

| Field | Value |
|-------|-------|
| Name | `express_server_entrypoint` |
| Version | `0.1.0` |
| Package | `generators/express_server_entrypoint` |

---

## Dependencies

| Generator | Why |
|-----------|-----|
| `typescript_base` | `package.json` must exist; TypeScript compiler options applied by `express_node_tsconfig` |

---

## Answers consumed

None. Uses `Spec.Metadata.ProjectName` for documentation only.

---

## Files written

| Path | Description |
|------|-------------|
| `src/index.ts` | Bootstraps HTTP server, reads `PORT` from env |
| `src/app.ts` | Express app: CORS (via `corsOptions()` — see below), JSON body parser, `GET /health` (with JSDoc `@openapi` annotation so `swagger-jsdoc` picks it up automatically when `express_swagger_jsdoc` is part of the scaffold) |
| `src/shared/cors.ts` | `corsOptions(): CorsOptions` helper that reads `CORS_ORIGIN` from env and returns a `cors` package configuration. Reused by the decorator architecture adapters that overwrite `app.ts`. |
| `.env.example` | Seed file with `PORT=3000` and `CORS_ORIGIN=http://localhost:3000`; downstream generators append to this file |

### CORS configuration (`src/shared/cors.ts`)

`corsOptions()` resolves the runtime CORS policy from the `CORS_ORIGIN` environment variable so the generated app passes SonarQube/static-analysis rules that flag a bare `app.use(cors())`:

| `CORS_ORIGIN` value | Resolved options |
|---------------------|-------------------|
| unset | `{ origin: 'http://localhost:3000', credentials: true }` |
| `*` | `{ origin: '*' }` (no credentials — browsers reject `*` + credentials anyway) |
| `https://a.com,https://b.com` | `{ origin: ['https://a.com', 'https://b.com'], credentials: true }` |

Production deployments must set `CORS_ORIGIN` to the explicit list of trusted origins. The decorator architecture adapters (`decorators_clean_arch_adapter`, `decorators_mvc_adapter`, `decorators_hexagonal_adapter`) overwrite `src/app.ts` but reuse this helper unchanged.

---

## Validators

| Check | Type | Passes when |
|-------|------|-------------|
| `src/index.ts` | `file_exists` | — |
| `src/app.ts` | `file_exists` | — |
| `src/shared/cors.ts` | `file_exists` | — |

---

## Post-generation commands

No PostGenerationCommands.

## Test commands

No TestCommands.

---

## Conflicts

None.

---

## See also

- [`express_server_typescript_deps`](express_server_typescript_deps.md) — npm deps + run scripts
- [`express_node_tsconfig`](express_node_tsconfig.md) — tsconfig overrides for Node/CommonJS
- [`decorators_clean_arch_adapter`](decorators_clean_arch_adapter.md), [`decorators_mvc_adapter`](decorators_mvc_adapter.md), [`decorators_hexagonal_adapter`](decorators_hexagonal_adapter.md) — overwrite `src/app.ts` but keep importing `corsOptions` from `./shared/cors`
