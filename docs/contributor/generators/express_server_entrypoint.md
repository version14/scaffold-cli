# Generator: `express_server_entrypoint`

Creates the Express TypeScript source files: `src/index.ts` (server bootstrap with `PORT`) and `src/app.ts` (Express app with `/health` route). Also writes the base `.env.example` that downstream generators append to.

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
| `src/app.ts` | Express app: CORS, JSON body parser, `GET /health` (with JSDoc `@openapi` annotation so `swagger-jsdoc` picks it up automatically when `express_swagger_jsdoc` is part of the scaffold) |
| `.env.example` | Seed file with `PORT=3000`; downstream generators append to this file |

---

## Validators

| Check | Type | Passes when |
|-------|------|-------------|
| `src/index.ts` | `file_exists` | — |
| `src/app.ts` | `file_exists` | — |

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
