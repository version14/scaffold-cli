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
- Case-level cache for `tools/test-flow`. A SHA-256 fingerprint over the
  fixture, every involved generator's source tree, the entire `flows/`
  directory, `pkg/dotapi/`, and `tools/test-flow/` itself is computed once
  scaffolding has resolved. On a hit, post-gen + test commands are skipped
  with a `cache: HIT` line in the report. Typical full warm runs go from
  ~7 min to ~4 s. Failed runs intentionally leave no cache entry. Disable
  with `-no-cache` or by removing `.test-flow-cache/`.
- `dotapi.Command.NoCache` field. Commands are **cacheable by default**;
  generator authors opt out by setting `NoCache: true` on the relevant
  `dotapi.Command{}`. The case-level cache only short-circuits a case when
  no PostGen/Test command involved has `NoCache: true`. The Background
  dev-server probe in `react_app` (`pnpm exec vite`) ships with
  `NoCache: true` so a real boot is verified on every run. Cache schema
  bumped to v2 — existing `.test-flow-cache/` entries are invalidated
  automatically.
- `tools/test-flow` is now **fail-fast** by default — it stops at the first
  failing case so the failure surfaces immediately. Pass `-keep-going` to
  run every case (useful for triaging multiple unrelated failures or for
  CI runs that want a complete report). The summary line distinguishes
  total / failed / not run, e.g. `✗ 1/18 cases failed (10 not run)`.

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
