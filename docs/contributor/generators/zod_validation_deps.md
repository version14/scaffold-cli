# Generator: `zod_validation_deps`

Adds the runtime dependencies and `tsconfig` flags needed by the decorator-based validation/OpenAPI stack: Zod, `@asteasolutions/zod-to-openapi`, `reflect-metadata`, `swagger-ui-express`, plus `experimentalDecorators` and `emitDecoratorMetadata`.

---

## Identity

| Field | Value |
|-------|-------|
| Name | `zod_validation_deps` |
| Version | `0.1.0` |
| Package | `generators/zod_validation_deps` |

---

## Dependencies

| Generator | Why |
|-----------|-----|
| `typescript_base` | `package.json` and `tsconfig.json` must already exist for the dependency / compiler-option merges |

---

## Answers consumed

None.

---

## Files written

None directly — only merges into existing files.

Also merges into:

| Path | Keys added / updated |
|------|---------------------|
| `package.json` | `dependencies.zod`, `dependencies.@asteasolutions/zod-to-openapi`, `dependencies.reflect-metadata`, `dependencies.swagger-ui-express`, `devDependencies.@types/swagger-ui-express` |
| `tsconfig.json` | `compilerOptions.experimentalDecorators = true`, `compilerOptions.emitDecoratorMetadata = true` |

---

## Validators

| Check | Type | Passes when |
|-------|------|-------------|
| `dependencies.zod` in `package.json` | `json_key_exists` | — |
| `dependencies.@asteasolutions/zod-to-openapi` in `package.json` | `json_key_exists` | — |
| `dependencies.reflect-metadata` in `package.json` | `json_key_exists` | — |
| `dependencies.swagger-ui-express` in `package.json` | `json_key_exists` | — |

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

- [generators/express_decorators_core.md](express_decorators_core.md)
- [generators/express_openapi_setup.md](express_openapi_setup.md)
- [docs/user/decorators.md](../../user/decorators.md)
