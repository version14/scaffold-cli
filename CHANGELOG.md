# Changelog — dot

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

### Added
- Decorator-based API validation and OpenAPI documentation flow (#91). When
  scaffolding an Express backend, dot now offers a `@Controller`/`@Get`/`@Body`
  decorator API with strongly-typed Zod schemas, request/response validation
  middleware, and an OpenAPI v3 spec served at `/docs`. Adapters ship for Clean
  Architecture, MVC, and Hexagonal projects, and a `RouterAdapter` interface
  keeps the system extensible to non-Express frameworks. See
  [docs/user/decorators.md](docs/user/decorators.md) for the quickstart.
- Classic JSDoc-driven Swagger fallback (`express_swagger_jsdoc`). When the
  decorator option is declined, dot still wires `swagger-ui-express` at
  `/docs` and ships `@openapi` JSDoc blocks on every generated handler
  (`/health`, `/auth/*`) so the spec is fully populated out of the box.
  The Swagger UI is therefore always available on a generated Express app —
  decorators only change *how* the spec is built.
- `auth_better_auth` no longer emits an unused `src/routes/auth.route.ts`;
  the BetterAuth catch-all (`toNodeHandler(auth)`) is mounted directly in
  `src/app.ts`.

### Changed
### Deprecated
### Removed
### Fixed
### Security

---

<!-- Copy the block above for each new release:

## [1.0.0] - YYYY-MM-DD

### Added
- Initial release

-->
