# GitHub Issues Plan

All issues follow the `feature_request.yml` template structure:
**Problem statement / Proposed solution / Alternatives considered / Area / Additional context**

Decision issues use the same structure — the "proposed solution" lists options with a recommendation.

---

## Setup first

### Milestones to create

| Name | Description |
|------|-------------|
| `v0.2` | Languages, project types, architectures, deployment, tools |
| `v0.3` | dot add module + conflict resolution |
| `v0.4` | Public community generator registry |
| `v0.5` | Project as Code (dot.yaml, dot plan, dot apply) |
| `v0.6` | GitLab CI + additional CI providers |
| `v1.0` | Full stabilization |
| `v1.1` | MCP server |

### Labels to create

| Label | Color | Description |
|-------|-------|-------------|
| `epic` | `#6B21A8` | Parent tracking issue |
| `decision` | `#DC2626` | Open design decision blocking implementation |
| `generator` | `#2563EB` | Generator implementation |
| `feature` | `#16A34A` | New CLI command or behavior |
| `polish` | `#CA8A04` | Hardening, UX, edge cases |
| `test` | `#EA580C` | Test coverage |
| `docs` | `#6B7280` | Documentation only |
| `infra` | `#0891B2` | CI, release, tooling |
| `blocked` | `#991B1B` | Waiting on a decision issue |

---

## v0.2 — Epics

---

### [Epic] v0.2: Languages, project types, architectures, deployment, tools

**Labels:** `epic`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
dot v0.1 only scaffolds a single Go REST API. It cannot generate frontend apps, Node.js or Python backends, monorepos, microservices, or any deployment setup. A developer who works in TypeScript, needs a Next.js frontend, or wants a microservices architecture cannot use dot at all.

**Proposed solution:**
Deliver the full content layer in v0.2. By the end of this milestone, dot supports:
- Project types: single project, monorepo, microservices
- Backend: Go, TypeScript (Express, NestJS), Python (FastAPI)
- Frontend: React, Next.js, Vue.js (all TypeScript)
- Architecture patterns at `dot init` time: MVC, Clean, Hexagonal (API); Feature-sliced, Atomic, Container/Presentational (frontend)
- Dev environment: Docker Compose (local services)
- Deployment: GitHub Actions deploy, Terraform (AWS + GCP), Kubernetes + Helm
- Add-on tools: Grafana, Sentry, PostHog, TanStack Router, TanStack Query, shadcn/ui, Payload CMS, gRPC, GraphQL
- CI: dynamically updated when modules are added
- Custom generators: `dot generator add/list/remove`

**Alternatives considered:**
Split v0.2 into multiple sub-versions (v0.2, v0.3, v0.4) to reduce scope. Rejected — the generator model makes these additions independent; each can be merged individually without blocking others. A single broad milestone is more readable.

**Additional context:**
Children (link issues here when created):
- [ ] [Decision] Architecture pattern (composition model, open questions on registry + naming)
- [ ] [Decision] Microservices init flow
- [ ] [Decision] Microservices gateway linking
- [ ] [Decision] Multi-language monorepo engine iteration
- [ ] CoreConfig v0.2 fields
- [ ] Multi-app engine support
- [ ] Architecture generators — Go (mvc, clean-arch, hexagonal)
- [ ] Architecture generators — TypeScript (backend + frontend patterns)
- [ ] Architecture generators — Python
- [ ] Node.js Express generator
- [ ] Node.js NestJS generator
- [ ] Python FastAPI generator
- [ ] GoRestAPIGenerator architecture composition
- [ ] React (Vite) generator
- [ ] Next.js generator
- [ ] Vue.js generator
- [ ] Docker Compose dev environment generator
- [ ] GitHub Actions deploy generator
- [ ] Terraform AWS generator
- [ ] Terraform GCP generator
- [ ] Kubernetes + Helm generator
- [ ] Grafana generator
- [ ] Sentry generator
- [ ] PostHog generator
- [ ] TanStack Router generator
- [ ] TanStack Query generator
- [ ] shadcn/ui generator
- [ ] Payload CMS generator
- [ ] gRPC generator
- [ ] GraphQL generator
- [ ] Microservices gateway generator
- [ ] dot add service command
- [ ] Dynamic CI update on dot add module
- [ ] dot generator add/list/remove

---

## v0.2 — Decisions

These must be resolved in a design session before the blocked issues can be implemented.

---

### [Decision] Architecture pattern: standalone generators with composition

**Labels:** `decision`, `blocked`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
Architecture pattern (MVC / Clean Architecture / Hexagonal) is selected at `dot init`. We need to finalize how generators implement this before writing any API generator code.

**Proposed solution:**
The agreed direction is **generator composition**: architecture generators are standalone generators, API generators compose them inside `Apply()`.

```
GoRestAPIGenerator.Apply(spec)
  → spec.Config.Architecture == "clean"
  → GoCleanArchGenerator{}.Apply(spec) → folder structure ops
  → append REST API-specific ops
  → return merged ops
```

Three open questions that need a decision before implementation:

**1. Registration model**

| Option | Description |
|--------|-------------|
| A — Registered (recommended) | Architecture generators are in the Registry, usable standalone via `dot init --module go-clean-arch`. API generators compose them by direct instantiation. |
| B — Internal helpers only | Not registered. Only used inside API generators via direct instantiation. Simpler, but not usable standalone. |

**2. Module naming**

If registered (Option A), architecture generators must claim a module name. Convention options:
- `go-clean-arch` claims `["clean-arch"]` — but this could conflict with `ts-clean-arch` also claiming `["clean-arch"]` (language disambiguates, which is fine per the Registry rules)
- `go-clean-arch` claims `["go-clean-arch"]` — no conflict risk but less ergonomic

**3. Composition scope**

When `GoRestAPIGenerator` composes `GoCleanArchGenerator` directly (not via the Registry), does `GoCleanArchGenerator` need to be registered at all? If it is registered, the Registry might try to match it on its own for specs that don't include a separate architecture module. Does that cause double application?

**Alternatives considered:**
Flag within each API generator (no separate architecture generators). Rejected — duplicates folder structure logic across all API generators per language. Adding a new pattern requires updating every generator.

**Additional context:**
Blocks all architecture generator implementation and all API generator work.
See `docs/developer-guide/roadmap/open-decisions.md #4` and `generator-interface.md#generator-composition`.

---

### [Decision] Microservices init flow

**Labels:** `decision`, `blocked`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
A microservices project in dot consists of a gateway and one or more services, each auto-linked to the gateway. We need to decide how a developer initializes this structure before building the microservices generator.

**Proposed solution:**

**Option A — Incremental (recommended)**
`dot init` with `type: microservices` generates the gateway only. `dot add service <name>` adds each service and links it to the gateway.

*Pros:* Maps to how real microservices projects grow — you don't know all services upfront. Each `dot add service` run is atomic and reversible. Easier to implement first.
*Cons:* Requires `dot add service` to patch the gateway config without breaking it. Depends on v0.3 conflict detection infrastructure being available.

**Option B — Upfront declaration**
`dot init` asks for gateway type and all service names/languages upfront. Generates everything in one shot.

*Pros:* Simpler to implement. No patching needed.
*Cons:* Forces the developer to know all services at project creation time. Inflexible. Doesn't scale to real workflows.

**Alternatives considered:**
A hybrid: `dot init` generates the gateway and one initial service. Additional services via `dot add service`. Similar to Option A but with an initial service included.

**Additional context:**
Blocks the microservices gateway generator and `dot add service` command.
See `docs/developer-guide/roadmap/open-decisions.md #5`.

---

### [Decision] Microservices gateway linking mechanism

**Labels:** `decision`, `blocked`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
When `dot add service <name>` runs, the new service must be registered in the gateway. We need to decide how this registration works before building the gateway generator and the `dot add service` command.

**Proposed solution:**

**Option A — Static config patch (recommended)**
The gateway config file (nginx.conf / Kong declarative YAML / Traefik routes.yml) is patched by `dot add service` using a new anchor (e.g. `AnchorGatewayRoutes`). The route for the new service is appended to the upstream/route list.

*Pros:* Consistent with dot's existing Patch model. No extra infrastructure. The gateway config is readable and reviewable in git.
*Cons:* Requires adding a new anchor type per gateway (nginx, Kong, Traefik). The patch logic must handle each config format.

**Option B — Env-driven**
The gateway reads service URLs from environment variables. dot generates `.env.example` with the service URL variables. No gateway config patching needed.

*Pros:* Simple. No complex patch logic.
*Cons:* Not automatic — the developer must wire env vars manually. Doesn't fulfill "automatically linked" promise.

**Option C — Service discovery**
Gateway uses Consul or etcd. Services self-register at startup.

*Pros:* Fully dynamic. Production-grade.
*Cons:* Adds heavy infrastructure. Out of scope for a scaffolding tool.

**Alternatives considered:**
Option C is out of scope for v0.2. Option B is a fallback if Option A proves too complex per gateway type.

**Additional context:**
Depends on: Microservices init flow decision.
Blocks: Microservices gateway generator.
See `docs/developer-guide/roadmap/open-decisions.md #6`.

---

### [Decision] Multi-language monorepo — engine iteration model

**Labels:** `decision`, `blocked`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
The current engine runs `registry.ForSpec(spec)` once for a single `Spec`. In a monorepo with a Go API and a Next.js frontend, each app has its own language and modules. The engine must run per app. We need to design this before building any monorepo or microservices generator.

**Proposed solution:**

**Option A — Per-app Spec array**
`dot init` for a monorepo produces one `Spec` per app. The engine iterates over the array, running `ForSpec` + `Apply` for each, scoping the pipeline's write path to the app's subdirectory.

`.dot/config.json` stores per-app state as a map keyed by app name:
```json
{
  "apps": {
    "api": { "spec": {...}, "commands": {...} },
    "frontend": { "spec": {...}, "commands": {...} }
  }
}
```

*Pros:* Clean separation. Each app is independently manageable. Backward compatible (single-app projects remain unchanged — they just have one entry in the map).
*Cons:* Requires a `Context` struct refactor and changes to `Load`/`Save`.

**Option B — Nested Spec**
A single top-level `Spec` with a `Children []Spec` field. The engine recurses.

*Pros:* Single Spec type.
*Cons:* More complex recursion. Harder to scope the write path per child.

**Alternatives considered:**
Option A is the clear choice. Option B adds complexity with no benefit.

**Additional context:**
This is a Spec and Context refactor. Resolve early — it affects `internal/spec/spec.go`, `internal/project/context.go`, and `cmd/dot/cmd_init.go`.
Blocks all monorepo and microservices generator work.
See `docs/developer-guide/roadmap/open-decisions.md #7`.

---

## v0.2 — Engine and Spec

---

### CoreConfig v0.2 fields

**Labels:** `feature`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
`CoreConfig` in `internal/spec/spec.go` only has fields for the v0.1 Go generator (linter, formatter, CI, deployment, monitoring, tracking). The new v0.2 generators need additional fields that `dot init` will ask for and that generators will read.

**Proposed solution:**
Add the following fields to `CoreConfig`:

```go
Architecture         string // "mvc" | "clean" | "hexagonal"
FrontendArchitecture string // "feature-sliced" | "atomic" | "container-presentational"
DeploymentTarget     string // "aws" | "gcp" | "none"
DeploymentType       string // "terraform" | "kubernetes" | "github-actions" | "all"
GatewayType          string // "nginx" | "kong" | "traefik" | "none"
ServiceComm          string // "rest" | "grpc" | "both"
```

Add new `ProjectType` constant: `microservices`.

Update the `dot init` TUI survey in `cmd/dot/cmd_init.go` to ask for these fields contextually (e.g. `Architecture` only appears for API project types; `GatewayType` only for microservices).

**Alternatives considered:**
Store these in `Extensions map[string]any` (the community generator escape hatch). Rejected — these are official generator fields; they belong in the typed struct.

**Additional context:**
Depends on: Architecture pattern decision (to know if `Architecture` is a top-level field or goes elsewhere).

---

### Multi-app engine support (monorepo and microservices)

**Labels:** `feature`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
The engine currently runs once for a single-app project. Monorepos and microservices have multiple apps of different languages. Without per-app engine iteration, these project types cannot be built.

**Proposed solution:**
Implement Option A from the engine iteration decision:
- Refactor `Context` in `internal/project/context.go` to store per-app state in an `Apps map[string]AppContext`
- `AppContext` contains the app's `Spec` and `Commands`, scoped to its subdirectory
- The engine iterates over `Apps`, running `ForSpec` + `Apply` per app, with `pipeline.Run` scoped to `<root>/<app-name>/`
- Single-app projects remain unchanged — they have one entry in `Apps`

**Alternatives considered:**
Nested `Spec` (Option B from the decision issue). Rejected for complexity reasons.

**Additional context:**
Depends on: Multi-language monorepo decision.
This is the foundation for all monorepo and microservices generators. Must ship before those generators can be written.

---

## v0.2 — Architecture generators

---

### Architecture generators — Go

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
There are no architecture pattern generators for Go. `GoRestAPIGenerator` can only produce a flat structure. Developers who want Clean Architecture or Hexagonal for their Go project cannot use dot.

**Proposed solution:**
Implement three standalone architecture generators in `generators/go/`:

- `GoMVCGenerator` (`go-mvc`) — `handlers/`, `models/`, `routes/`
- `GoCleanArchGenerator` (`go-clean-arch`) — `domain/`, `usecases/`, `interfaces/`, `infrastructure/`
- `GoHexagonalGenerator` (`go-hexagonal`) — `core/` (domain + ports), `adapters/` (primary + secondary)

Each generator:
- Implements the full `Generator` interface
- `Apply()` returns `Create` ops for the directory structure and base files (e.g. `domain/entity.go` stub)
- Is registered independently in `buildRegistry()`
- Is tested independently in `generators/go/<pattern>_test.go`

API generators compose these via static composition based on `spec.Config.Architecture`.

**Alternatives considered:**
Fold architecture logic into each API generator. Rejected — duplication across Go, Node, Python generators. The composition model means architecture logic is written once per language.

**Additional context:**
These generators are also usable standalone — a developer who wants just the folder structure without a specific framework can use `dot init --module go-clean-arch`.

---

### Architecture generators — TypeScript

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
No architecture pattern generators exist for TypeScript. All TS framework generators (Express, NestJS, React, Next.js, Vue) need to compose architecture generators.

**Proposed solution:**
Implement in `generators/typescript/`:

**Backend:**
- `TSMVCGenerator` (`ts-mvc`) — `controllers/`, `models/`, `routes/`
- `TSCleanArchGenerator` (`ts-clean-arch`) — `domain/`, `usecases/`, `interfaces/`, `infrastructure/`
- `TSHexagonalGenerator` (`ts-hexagonal`) — `core/`, `adapters/`

**Frontend:**
- `TSFeatureSlicedGenerator` (`ts-feature-sliced`) — `features/`, `entities/`, `shared/`, `pages/`, `app/`
- `TSAtomicDesignGenerator` (`ts-atomic-design`) — `atoms/`, `molecules/`, `organisms/`, `templates/`, `pages/`
- `TSContainerPresentationalGenerator` (`ts-container-presentational`) — `containers/`, `components/`, `pages/`

**Alternatives considered:**
Share architecture generators across languages (e.g. a single `CleanArchGenerator` that emits language-agnostic directory ops). Rejected — folder structures and base file content differ meaningfully by language (Go interfaces vs TypeScript types vs Python abstract classes).

---

### Architecture generators — Python

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
No architecture pattern generators for Python. FastAPI generator needs to compose these.

**Proposed solution:**
Implement in `generators/python/`:

- `PythonMVCGenerator` (`python-mvc`) — `routers/`, `models/`, `schemas/`
- `PythonCleanArchGenerator` (`python-clean-arch`) — `domain/`, `usecases/`, `interfaces/`, `infrastructure/`
- `PythonHexagonalGenerator` (`python-hexagonal`) — `core/`, `adapters/`

---

## v0.2 — Backend generators

---

### Node.js Express generator (TypeScript)

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
Developers building Node.js REST APIs with Express cannot use dot. They have to set up TypeScript config, project structure, linting, and the server boilerplate by hand every time.

**Proposed solution:**
Implement `NodeExpressGenerator` in `generators/typescript/express.go`.

- `Name()`: `"node-express"`
- `Language()`: `"typescript"`
- `Modules()`: `["express"]`

`Apply()` generates (file structure varies by `spec.Config.Architecture`):

**MVC:**
```
src/
  index.ts          ← Express server entry
  controllers/
  models/
  routes/
  middlewares/
package.json
tsconfig.json
.eslintrc.json
```

**Clean Architecture:**
```
src/
  domain/
  usecases/
  interfaces/controllers/
  infrastructure/
```

**Hexagonal:**
```
src/
  core/             ← domain + ports
  adapters/         ← primary (HTTP) + secondary (DB, external)
```

Commands registered: `new route <name>`, `new controller <name>`, `new middleware <name>`

**Alternatives considered:**
Generate a framework-agnostic Node.js setup and let the developer add Express. Rejected — dot's value is the complete, wired-up scaffold, not a blank slate.

**Additional context:**
Uses static composition with `TSMVCGenerator`, `TSCleanArchGenerator`, or `TSHexagonalGenerator` based on `spec.Config.Architecture`. Depends on: TypeScript architecture generators.

---

### Node.js NestJS generator (TypeScript)

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
NestJS is the most widely used opinionated Node.js framework. Developers starting a NestJS project still need to manually configure TypeScript, linting, architecture, and the module/controller/service structure.

**Proposed solution:**
Implement `NodeNestJSGenerator` in `generators/typescript/nestjs.go`.

- `Name()`: `"node-nestjs"`
- `Language()`: `"typescript"`
- `Modules()`: `["nestjs"]`

`Apply()` generates a NestJS application with:
- `package.json` (NestJS + TypeScript dependencies)
- `tsconfig.json`, `nest-cli.json`
- `src/app.module.ts`, `src/main.ts`
- File structure adapted to `spec.Config.Architecture`

Commands: `new module <name>`, `new controller <name>`, `new service <name>`

**Alternatives considered:**
Use the NestJS CLI under the hood (`nest new`). Rejected — dot generates deterministic files from templates; executing external CLIs adds a runtime dependency.

**Additional context:**
Uses static composition with TypeScript backend architecture generators. Depends on: TypeScript architecture generators.

---

### Python FastAPI generator

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
Python developers using FastAPI have no dot support. Setting up a FastAPI project with the right structure, typing, linting (ruff), and architecture pattern is repetitive work.

**Proposed solution:**
Implement `PythonFastAPIGenerator` in `generators/python/fastapi.go`.

- `Name()`: `"python-fastapi"`
- `Language()`: `"python"`
- `Modules()`: `["fastapi"]`

`Apply()` generates:
- `pyproject.toml` (or `requirements.txt`) with FastAPI + uvicorn
- `main.py` — FastAPI app entry point
- File structure varies by `spec.Config.Architecture`
- `.python-version` for pyenv users

Commands: `new route <name>`, `new schema <name>`, `new dependency <name>`

**Alternatives considered:**
Use `cookiecutter` templates. Rejected — same reason as NestJS: dot generates from code, not external tools.

**Additional context:**
Uses static composition with Python architecture generators. Depends on: Python architecture generators.
Note: this is dot's first Python generator. It establishes the `generators/python/` package.

---

### GoRestAPIGenerator — add architecture patterns

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
The existing `GoRestAPIGenerator` only generates a flat `main.go` + `routes/` structure (implicit MVC). Developers who want Clean Architecture or Hexagonal for their Go API cannot use dot.

**Proposed solution:**
Extend `GoRestAPIGenerator.Apply()` to branch on `spec.Config.Architecture`:

- `"mvc"` (default): current behavior (`main.go`, `routes/`, `handlers/`)
- `"clean"`: `domain/`, `usecases/`, `interfaces/`, `infrastructure/`
- `"hexagonal"`: `core/` (domain + ports), `adapters/` (primary: HTTP, secondary: DB)

Update the TUI survey to ask for architecture pattern when `language = go` and `type = api`.

**Alternatives considered:**
Create separate generators per architecture (e.g. `GoCleanAPIGenerator`). Rejected — they share 80% of the code and would create confusing duplication in `official-generators.md`.

**Additional context:**
Uses static composition with Go architecture generators based on `spec.Config.Architecture`. Depends on: Go architecture generators.

---

## v0.2 — Frontend generators

---

### React (Vite + TypeScript) generator

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
Developers starting a React SPA still wire Vite, TypeScript, ESLint, and folder structure by hand. There is no dot support for any frontend project type.

**Proposed solution:**
Implement `ReactTSGenerator` in `generators/typescript/react.go`.

- `Name()`: `"ts-react"`
- `Language()`: `"typescript"`
- `Modules()`: `["react"]`

`Apply()` generates:
- `package.json`, `tsconfig.json`, `vite.config.ts`
- `index.html`, `src/main.tsx`, `src/App.tsx`
- File structure under `src/` varies by `spec.Config.FrontendArchitecture`:
  - **Feature-sliced**: `features/`, `entities/`, `shared/`, `pages/`, `app/`
  - **Atomic Design**: `atoms/`, `molecules/`, `organisms/`, `templates/`, `pages/`
  - **Container/Presentational**: `containers/`, `components/`, `pages/`

Commands: `new component <name>`, `new page <name>`, `new hook <name>`

**Alternatives considered:**
Use Vite's `create-vite` under the hood. Rejected — external tools add runtime dependencies and break determinism.

**Additional context:**
Uses static composition with TypeScript frontend architecture generators (`TSFeatureSlicedGenerator`, `TSAtomicDesignGenerator`, `TSContainerPresentationalGenerator`) based on `spec.Config.FrontendArchitecture`.
This is the first TypeScript generator. Establishes `generators/typescript/` package. Depends on: TypeScript architecture generators.

---

### Next.js generator (TypeScript)

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
Next.js is the most widely used React framework. Starting a Next.js project with a clean structure, proper TypeScript config, and a chosen architecture pattern requires significant manual setup.

**Proposed solution:**
Implement `NextJSGenerator` in `generators/typescript/nextjs.go`.

- `Name()`: `"ts-nextjs"`
- `Language()`: `"typescript"`
- `Modules()`: `["nextjs"]`

`Apply()` generates:
- `package.json`, `tsconfig.json`, `next.config.ts`
- App Router structure (`app/layout.tsx`, `app/page.tsx`, `app/loading.tsx`)
- `public/` directory
- File structure under `src/` varies by `spec.Config.FrontendArchitecture`

Router choice (App Router vs Pages Router) asked at `dot init`.

Commands: `new page <name>`, `new component <name>`, `new api-route <name>`

**Alternatives considered:**
Use `create-next-app` under the hood. Rejected — same reason as other generators.

**Additional context:**
Uses static composition with TypeScript frontend architecture generators based on `spec.Config.FrontendArchitecture`. Depends on: TypeScript architecture generators.

---

### Vue.js (Vite + TypeScript) generator

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
Vue.js developers have no dot support. Setting up a Vite + Vue 3 + TypeScript project with the right architecture is manual work.

**Proposed solution:**
Implement `VueTSGenerator` in `generators/typescript/vue.go`.

- `Name()`: `"ts-vue"`
- `Language()`: `"typescript"`
- `Modules()`: `["vue"]`

`Apply()` generates:
- `package.json`, `tsconfig.json`, `vite.config.ts`
- `src/App.vue`, `src/main.ts`
- `src/components/`, `src/views/`, `src/router/`, `src/stores/` (Pinia)
- File structure varies by `spec.Config.FrontendArchitecture`

Commands: `new component <name>`, `new view <name>`, `new store <name>`

**Alternatives considered:**
Use `create-vue` under the hood. Rejected — same reason as other generators.

**Additional context:**
Uses static composition with TypeScript frontend architecture generators based on `spec.Config.FrontendArchitecture`. Depends on: TypeScript architecture generators.

---

## v0.2 — Dev environment

---

### Docker Compose dev environment generator

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
Developers using dot to set up a project with Postgres and Redis still have to write the `docker-compose.yml` for local development by hand. There is no way to spin up all services with a single command after `dot init`.

**Proposed solution:**
Implement `DockerComposeDevGenerator` in `generators/common/docker_compose_dev.go`.

- `Name()`: `"common-docker-compose-dev"`
- `Language()`: `"*"`
- `Modules()`: `["docker-compose-dev"]`

`Apply()` generates a `docker-compose.yml` that includes all declared modules as services:
- `postgres` → Postgres service on port 5432 with env vars
- `redis` → Redis service on port 6379
- etc.

The app itself is **not** in the compose file — developers run the app locally.

Also generates `docker-compose.override.yml.example` for local customization.

When `dot add module <service>` is run later (v0.3), the compose file is updated to add the new service.

**Alternatives considered:**
Include the app itself in Docker Compose. Rejected — mixing local dev (hot reload) with Docker-managed containers adds complexity and slows iteration.

**Additional context:**
This is dev environment only. Production deployment is handled by the deployment generators.
Language `"*"` — works alongside any backend language.

---

## v0.2 — Deployment generators

---

### GitHub Actions deploy workflow generator

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
After scaffolding a project, developers still have to write the deployment CI workflow manually. Each cloud provider (AWS, GCP) has different steps, secrets, and service configurations.

**Proposed solution:**
Implement `GitHubActionsDeployGenerator` in `generators/common/github_actions_deploy.go`.

- `Name()`: `"common-github-actions-deploy"`
- `Language()`: `"*"`
- `Modules()`: `["github-actions-deploy"]`

`Apply()` generates `.github/workflows/deploy.yml` based on `spec.Config.DeploymentTarget`:

- **AWS**: build Docker image → push to ECR → deploy to ECS (or App Runner)
- **GCP**: build Docker image → push to Artifact Registry → deploy to Cloud Run

Also generates `docs/deployment.md` listing required GitHub secrets (e.g. `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `ECR_REPOSITORY`).

Triggered on push to `main`. Configurable (branch name).

**Alternatives considered:**
Use GitHub Actions reusable workflows. Good for DRY, but adds a dependency on an external workflow repo. Start with inline workflows for simplicity.

---

### Terraform AWS generator

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
Developers deploying to AWS need Terraform infrastructure setup. Writing VPC, ECS, RDS, IAM from scratch is complex and error-prone. There is no dot support for infrastructure-as-code.

**Proposed solution:**
Implement `TerraformAWSGenerator` in `generators/common/terraform_aws.go`.

- `Name()`: `"common-terraform-aws"`
- `Language()`: `"*"`
- `Modules()`: `["terraform-aws"]`

`Apply()` generates `infrastructure/` with Terraform modules:
- `vpc.tf` — VPC, subnets, security groups
- `ecs.tf` — ECS cluster + task definition
- `rds.tf` — only if `postgres` module is present (`spec.HasModule("postgres")`)
- `elasticache.tf` — only if `redis` module is present
- `ecr.tf` — ECR repository
- `iam.tf` — IAM roles and policies
- `variables.tf`, `outputs.tf`, `main.tf`
- `README.md` — setup and apply instructions

**Alternatives considered:**
Use CDK instead of Terraform. Terraform is more language-agnostic and more widely adopted for infrastructure work across teams.

---

### Terraform GCP generator

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
Same as Terraform AWS, but for GCP (Cloud Run, Cloud SQL, Memorystore).

**Proposed solution:**
Implement `TerraformGCPGenerator` in `generators/common/terraform_gcp.go`.

- `Name()`: `"common-terraform-gcp"`
- `Language()`: `"*"`
- `Modules()`: `["terraform-gcp"]`

`Apply()` generates `infrastructure/` with:
- `cloud_run.tf` — Cloud Run service
- `cloud_sql.tf` — only if `postgres` module is present
- `memorystore.tf` — only if `redis` module is present
- `artifact_registry.tf`
- `iam.tf`
- `variables.tf`, `outputs.tf`, `main.tf`
- `README.md`

---

### Kubernetes + Helm generator

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
Teams deploying to Kubernetes need base manifests and a Helm chart to parameterize them. Writing these by hand is repetitive, and the structure varies enough that a generator provides real value.

**Proposed solution:**
Implement `KubernetesGenerator` in `generators/common/kubernetes.go`.

- `Name()`: `"common-kubernetes"`
- `Language()`: `"*"`
- `Modules()`: `["kubernetes"]`

`Apply()` generates:
- `k8s/deployment.yaml`, `k8s/service.yaml`, `k8s/ingress.yaml`
- `k8s/configmap.yaml`, `k8s/secret.yaml` (stubs with placeholder keys)
- `helm/Chart.yaml`, `helm/values.yaml`, `helm/templates/` mirroring `k8s/`
- `README.md` — kubectl and helm usage

Namespace, image name, and replica count are taken from the project spec. All values are parameterized in `helm/values.yaml`.

---

## v0.2 — Add-on tool generators

---

### Grafana generator

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
Developers who want Grafana monitoring for their project have to manually configure dashboards, data sources, and provisioning. There is no automated setup.

**Proposed solution:**
Implement `GrafanaGenerator` in `generators/common/grafana.go`.

- `Name()`: `"common-grafana"`
- `Language()`: `"*"`
- `Modules()`: `["grafana"]`

`Apply()` generates:
- `grafana/provisioning/datasources/datasource.yml` — Prometheus data source
- `grafana/provisioning/dashboards/dashboard.yml` — dashboard provisioning config
- `grafana/dashboards/app.json` — base application dashboard
- Grafana service added to `docker-compose.yml` (via Patch on the compose file)
- Backend: Prometheus metrics endpoint setup (varies by language)

---

### Sentry generator

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
Integrating Sentry requires installing the SDK, initializing it in the entry point, adding an error boundary (frontend), or error middleware (backend). This is identical across projects and a good candidate for automation.

**Proposed solution:**
Implement `SentryGenerator` in `generators/common/sentry.go`.

- `Name()`: `"common-sentry"`
- `Language()`: `"*"`
- `Modules()`: `["sentry"]`

`Apply()` generates (varies by language):
- **Go**: Sentry middleware for HTTP handler, Sentry init in `main.go` (via `AnchorMainFunc`)
- **TypeScript/Node**: Sentry init at top of entry point, Express error handler
- **TypeScript/React + Next.js**: `sentry.client.config.ts`, `sentry.server.config.ts`, error boundary component
- **Python**: Sentry middleware for FastAPI, `sentry_sdk.init()` in `main.py`
- `.env.example` patched to add `SENTRY_DSN=`

---

### PostHog generator

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
PostHog analytics setup requires SDK installation, provider wrapping (frontend), and event capture helpers (backend). Same boilerplate every time.

**Proposed solution:**
Implement `PostHogGenerator` in `generators/common/posthog.go`.

- `Name()`: `"common-posthog"`
- `Language()`: `"*"`
- `Modules()`: `["posthog"]`

`Apply()` generates:
- **Frontend (React/Next/Vue)**: PostHog provider wrapper in app entry, `usePostHog` hook
- **Backend**: server-side event capture helper function
- `.env.example` patched to add `NEXT_PUBLIC_POSTHOG_KEY=` (frontend) or `POSTHOG_API_KEY=` (backend)

---

### TanStack Router generator

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
TanStack Router requires specific setup (file-based routing config, route tree generation, provider wrapping) that is the same across every React or Vue project.

**Proposed solution:**
Implement `TanstackRouterGenerator` in `generators/typescript/tanstack_router.go`.

- `Name()`: `"ts-tanstack-router"`
- `Language()`: `"typescript"`
- `Modules()`: `["tanstack-router"]`

`Apply()` generates:
- `vite.config.ts` patched to add the TanStack Router Vite plugin
- `src/routes/__root.tsx` — root route with layout
- `src/routes/index.tsx` — index route
- `src/main.tsx` patched to wrap the app in `RouterProvider`

Commands: `new route <name>` — creates a new route file under `src/routes/`

---

### TanStack Query generator

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
TanStack Query (React Query) requires QueryClient setup, provider wrapping, and devtools configuration. Same three files every project.

**Proposed solution:**
Implement `TanstackQueryGenerator` in `generators/typescript/tanstack_query.go`.

- `Name()`: `"ts-tanstack-query"`
- `Language()`: `"typescript"`
- `Modules()`: `["tanstack-query"]`

`Apply()` generates:
- `src/lib/query-client.ts` — QueryClient instance
- `src/main.tsx` patched to wrap in `QueryClientProvider` and add `ReactQueryDevtools`

---

### shadcn/ui generator

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
shadcn/ui requires Tailwind CSS, `components.json`, a `components/ui/` directory, and specific `tsconfig.json` path aliases. The setup is documented but tedious and error-prone.

**Proposed solution:**
Implement `ShadcnUIGenerator` in `generators/typescript/shadcn_ui.go`.

- `Name()`: `"ts-shadcn-ui"`
- `Language()`: `"typescript"`
- `Modules()`: `["shadcn-ui"]`
- Compatible with `ts-react` and `ts-nextjs` only.

`Apply()` generates:
- `components.json` — shadcn/ui config
- `tailwind.config.ts` — Tailwind + shadcn preset
- `src/components/ui/` — empty directory with `.gitkeep`
- `src/lib/utils.ts` — `cn()` utility
- `tsconfig.json` patched to add `@/` path alias

---

### Payload CMS generator

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
Payload CMS requires specific Next.js configuration, a dedicated admin route, database config, and a collection definition. Setting this up alongside a new Next.js project takes significant time.

**Proposed solution:**
Implement `PayloadCMSGenerator` in `generators/typescript/payload.go`.

- `Name()`: `"ts-payload"`
- `Language()`: `"typescript"`
- `Modules()`: `["payload"]`
- Compatible with `ts-nextjs` only.

`Apply()` generates:
- `payload.config.ts` — Payload config with DB adapter (MongoDB or Postgres based on spec)
- `app/(payload)/admin/[[...segments]]/page.tsx` — Payload admin route
- `app/(payload)/admin/[[...segments]]/not-found.tsx`
- `collections/Users.ts` — sample collection
- `next.config.ts` patched to add Payload plugin

---

### gRPC generator

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
gRPC setup (protobuf files, generated stubs, build scripts, server/client wiring) is complex and varies by language. It is required for microservices using gRPC for service communication.

**Proposed solution:**
Implement `GRPCGenerator` in `generators/common/grpc.go`.

- `Name()`: `"common-grpc"`
- `Language()`: `"*"`
- `Modules()`: `["grpc"]`

`Apply()` generates (varies by language):
- `proto/` directory with a sample `service.proto` file
- **Go**: `Makefile` target for `protoc`, generated stub output dir, gRPC server setup in `main.go`
- **TypeScript**: `proto/` + `@grpc/proto-loader` setup, server and client stubs
- **Python**: `proto/` + `grpcio-tools` setup, generated stub import

Commands: `new service <name>` — adds a new proto service definition

---

### GraphQL generator

**Labels:** `generator`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
GraphQL setup differs meaningfully by language and framework, but the core structure (schema, resolvers, server setup) is the same every time.

**Proposed solution:**
Implement `GraphQLGenerator` in `generators/common/graphql.go`.

- `Name()`: `"common-graphql"`
- `Language()`: `"*"`
- `Modules()`: `["graphql"]`

`Apply()` generates (varies by language):
- **Go**: `gqlgen.yml`, `graph/schema.graphqls`, `graph/resolver.go`, `server.go` patched to mount GraphQL handler
- **TypeScript/Node**: Apollo Server setup, `src/schema.ts`, `src/resolvers/index.ts`
- **Python**: Strawberry GraphQL setup in FastAPI

Commands: `new type <name>` — adds a new GraphQL type and resolver

---

## v0.2 — Microservices

---

### Microservices gateway generator

**Labels:** `generator`, `blocked`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
A microservices project needs a gateway that routes requests to the correct service. Setting up nginx, Kong, or Traefik as a gateway with the right routing rules is complex and must be done consistently with dot's Patch model.

**Proposed solution:**
Implement `MicroservicesGatewayGenerator` in `generators/common/microservices_gateway.go`.

- `Name()`: `"common-microservices-gateway"`
- `Language()`: `"*"`
- `Modules()`: `["microservices-gateway"]`

`Apply()` generates based on `spec.Config.GatewayType`:
- **nginx**: `gateway/nginx.conf` with upstream + location blocks for initial services
- **Kong**: `gateway/kong.yml` (declarative) with service and route definitions
- **Traefik**: `gateway/traefik.yml` + `gateway/dynamic/routes.yml`
- `docker-compose.yml` entry for the gateway service

Defines a new anchor `AnchorGatewayRoute` per gateway type for use by `dot add service`.

**Alternatives considered:**
See gateway linking decision. Static config patch chosen over service discovery and env-driven.

**Additional context:**
Depends on: Microservices init flow decision, Microservices gateway linking decision.

---

### dot add service command

**Labels:** `feature`, `blocked`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
After creating a microservices project, there is no way to add a new service and have it automatically linked to the gateway. Developers have to manually create the service directory and update the gateway config.

**Proposed solution:**
Implement `dot add service <name>` in `cmd/dot/cmd_add.go`.

```bash
dot add service user-service --lang go --type rest-api
dot add service payment-service --lang typescript --framework nestjs
```

1. Resolves the generator for the specified language + framework
2. Runs `generator.Apply(spec)` scoped to `<root>/services/<name>/`
3. Patches the gateway config to add the new service route (using `AnchorGatewayRoute`)
4. Updates `.dot/config.json` with the new app entry

**Alternatives considered:**
Re-use `dot add module`. Rejected — a service is a full app with its own spec and directory, not a module added to an existing app.

**Additional context:**
Depends on: Microservices gateway generator, Multi-app engine support.

---

## v0.2 — CI and custom generators

---

### Dynamic CI update on dot add module

**Labels:** `feature`
**Milestone:** v0.2
**Area:** Core

**Problem statement:**
When a developer adds a Postgres module to an existing project (v0.3 command), the GitHub Actions CI workflow is not updated. Tests will fail in CI because there is no Postgres service container. The developer has to update the workflow manually.

**Proposed solution:**
Extend `GitHubActionsGenerator.RunAction()` to handle a `"ci.add-service"` action. When `dot add module postgres` runs, the CI generator is invoked to patch `.github/workflows/ci.yml` and add the Postgres service container.

Each service module (postgres, redis, etc.) declares its required CI service definition. The GitHub Actions generator maps module names to service container snippets and patches them in.

**Alternatives considered:**
Regenerate the whole CI file on `dot add module`. Rejected — the file may have been user-modified. Patching with conflict detection (v0.3 infrastructure) is the correct approach.

**Additional context:**
Depends on: `dot add module` (v0.3). This issue is in v0.2 because the generator code must be written here; it is activated by v0.3 work.

---

### dot generator add/list/remove

**Labels:** `feature`, `blocked`
**Milestone:** v0.2
**Area:** Developer experience

**Problem statement:**
Developers who want to use a custom generator with dot have no way to register it without modifying the dot source code. There is no CLI to manage local generators.

**Proposed solution:**
Implement `dot generator add/list/remove` in `cmd/dot/cmd_generator.go`.

```bash
dot generator add ./my-generator    # register a local generator path
dot generator list                  # list registered generators: name, language, modules, source
dot generator remove my-generator   # unregister by name
```

Local generators are stored in `.dot/generators.json`:
```json
{
  "generators": [
    { "name": "my-generator", "source": "./my-generator", "version": "local" }
  ]
}
```

At startup, `buildRegistry()` reads `.dot/generators.json` and loads registered generators in addition to the official ones.

**Alternatives considered:**
Require users to fork dot and add generators to `cmd/dot/build.go`. Current workaround but not scalable.
Subprocess-based loading. Better long-term but requires the loading mechanism decision first.

**Additional context:**
Depends on: Community generator loading mechanism decision (open-decisions.md #1).

---

## v0.3 — Epic (light)

---

### [Epic] v0.3: dot add module + conflict resolution

**Labels:** `epic`
**Milestone:** v0.3
**Area:** Core

**Problem statement:**
There is no way to safely add a module (Postgres, Redis, Docker) to a project after `dot init`. The conflict detection infrastructure (manifest hash comparison, conflict markers) exists in design but is not implemented.

**Proposed solution:**
Implement the full `dot add module` flow with conflict detection:
1. `dot add module <name>` — run generator Apply(), pipe through conflict-aware pipeline
2. Conflict detection — compare current file hash to `.dot/manifest.json`
3. Conflict marker writing — git-style markers for modified files
4. `dot status` — list files with unresolved conflicts
5. `dot resolve` — mark resolved, update manifest hashes

**Additional context:**
Children:
- [ ] [Decision] Conflict marker format + dot resolve UX (open-decisions.md #2)
- [ ] dot add module command
- [ ] Conflict detection (manifest hash comparison)
- [ ] Conflict marker writing
- [ ] dot status command
- [ ] dot resolve command
- [ ] Edge case tests (binary files, deleted files, renamed files)

---

## v0.4 — Epic (light)

---

### [Epic] v0.4: Public community generator registry

**Labels:** `epic`
**Milestone:** v0.4
**Area:** Core

**Problem statement:**
Community generators written for dot cannot be shared or discovered. There is no registry.

**Proposed solution:**
Build a public registry for community generators with `dot generator publish`, `dot generator install`, and `dot generator search/info/update`.

**Additional context:**
Children:
- [ ] [Decision] Registry infrastructure (GitHub-based index vs dedicated service)
- [ ] dot generator publish
- [ ] dot generator install (with checksum verification + version pinning)
- [ ] dot generator search/info/update

---

## v0.5 — Epic (light)

---

### [Epic] v0.5: Project as Code

**Labels:** `epic`
**Milestone:** v0.5
**Area:** Core

**Problem statement:**
Teams cannot describe a full project declaratively in a file, version it in git, and reproduce it with a command. The TUI is the only input layer.

**Proposed solution:**
Add a `dot.yaml` input layer with `dot plan` (diff) and `dot apply` (delta execution).

**Additional context:**
Children:
- [ ] [Decision] dot plan diff algorithm (open-decisions.md #3)
- [ ] dot.yaml parser → Spec
- [ ] dot plan command
- [ ] dot apply command
- [ ] dot.yaml schema reference documentation

---

## Cross-cutting (no milestone)

---

### Improve test coverage — internal/pipeline patch anchors

**Labels:** `test`
**Area:** Core

**Problem statement:**
`AnchorMainFunc` and `AnchorInitFunc` have no dedicated test file. Edge cases (function not found, nested braces, empty function body) are untested.

**Proposed solution:**
Add table-driven tests to `internal/pipeline/patch_test.go` covering:
- Function found, single statement inserted
- Function not found → error
- Nested braces inside the function (depth tracking)
- Empty function body
- Multiple functions in file (only the target is modified)

---

### Improve test coverage — GoRestAPIGenerator

**Labels:** `test`
**Area:** Core

**Problem statement:**
`generators/go/rest_api.go` has no test file. `Apply()` and `RunAction()` are untested.

**Proposed solution:**
Create `generators/go/rest_api_test.go` with table-driven tests:
- `Apply()`: given a known Spec, assert exact file paths, kinds, and content snippets
- `RunAction("rest-api.new-route", ["UserController"], spec)`: assert `routes/UserController.go` is created with correct content
- `RunAction("rest-api.new-handler", ["UserController"], spec)`: same for handlers
- `RunAction("unknown-action", ...)`: assert error returned

---

### Go version audit

**Labels:** `infra`
**Area:** Core

**Problem statement:**
`go.mod` declares `go 1.26`. The troubleshooting section in `docs/getting-started/README.md` says "should be 1.21+". These are inconsistent.

**Proposed solution:**
Decide on the actual minimum Go version, update `go.mod` and all documentation to match. If 1.26 features are used, document why. If not, lower to the actual minimum.
