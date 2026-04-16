# Open Decisions

Unresolved questions that block specific features. Each decision must be made with collaborators before the affected feature is implemented.

---

## 1. Community generator loading mechanism

**Blocks:** v0.2 local custom generators

Go has no simple cross-platform plugin system. Three options:

| Option | How it works | Pros | Cons |
|--------|-------------|------|------|
| **In-process** | Community generators are Go modules imported at compile time. Users compile a custom dot binary. | Simple, no IPC | Requires recompiling dot to add generators |
| **Subprocess / RPC** | Community generators are separate binaries. dot spawns them, communicates via stdin/stdout JSON. | Flexible, any language | More complex, FileOpKind must be string-typed (already done) |
| **Embedded registry** | dot fetches generator binaries from a registry URL, runs as subprocesses. Like Terraform providers. | Best UX | Most infrastructure to build |

**Note:** The Generator interface is compatible with all three options. `FileOpKind` is already string-typed in anticipation of the subprocess option.

**Decision needed before:** v0.2 local custom generator work begins.

---

## 2. dot resolve UX and conflict marker format

**Blocks:** v0.3 `dot add module` and conflict resolution

The conflict strategy is settled (git-style markers). What is not decided:

- Exact format of the conflict markers (which metadata to include in the header lines)
- What `dot resolve` does step by step
- How to handle binary files (images, compiled assets) — cannot use text markers
- How to handle files deleted by the user after `dot init`
- How to handle renamed files — the manifest stores the original path

**Decision needed before:** v0.3 conflict resolution implementation begins.

---

## 3. dot plan diff algorithm

**Blocks:** v0.5 `dot plan` command

The output format is clear ("will add: user-service/redis") but the algorithm is not.
A complete diff must handle:

- Added apps (new key in `dot.yaml` apps)
- Removed apps (key in `.dot/config.json` but not in `dot.yaml`)
- Added modules per app
- Removed modules per app
- Changed `CoreConfig` fields
- Changed `Extensions`

**Decision needed before:** v0.5 `dot plan` implementation begins.

---

## 4. Architecture pattern: standalone generators with composition

**Blocks:** v0.2 API generators (Go, Node, Python)

Architecture pattern (MVC / Clean Architecture / Hexagonal) is selected at `dot init`. The composition model is the agreed direction: architecture generators are standalone registered generators, and API generators compose them via static composition inside `Apply()`. What still needs to be decided before implementation begins:

**Direction (agreed):**
```
GoRestAPIGenerator.Apply(spec)
  → spec.Config.Architecture == "clean"
  → GoCleanArchGenerator{}.Apply(spec) → folder structure ops
  → append REST API-specific ops (main.go, routes/, etc.)
  → return merged ops
```

**Open questions:**

| Question | Options |
|----------|---------|
| Do architecture generators register in the Registry (usable standalone), or are they internal helpers only? | A) Registered — can be invoked directly by `dot init --module go-clean-arch`. B) Internal only — not registered, only used via composition. |
| What does an architecture generator's `Modules()` return? If registered, it must claim a module name. Does `go-clean-arch` claim `["clean-arch"]`? Does it conflict with language-specific generators? | Needs naming convention decision. |
| Can an architecture generator's `Apply()` be called multiple times (once per composed API generator) without producing duplicate ops? | Determinism contract should cover this, but needs verification. |

**Decision needed before:** architecture generator implementation begins.

See [generator-interface.md](../internals/generator-interface.md#generator-composition) for the composition pattern documentation.

---

## 5. Microservices init flow

**Blocks:** v0.2 microservices project type

Two possible flows:

| Option | Flow |
|--------|------|
| **Upfront declaration** | `dot init` with `type: microservices` asks for gateway type and all services upfront. Generates everything in one shot. |
| **Incremental** | `dot init` generates the gateway only. `dot add service <name>` adds each service and links it to the gateway. |

The upfront approach is simpler to implement but inflexible. The incremental approach maps better to how real projects grow but requires `dot add service` to modify the gateway config without breaking it.

**Decision needed before:** v0.2 microservices generator design begins.

---

## 6. Microservices gateway linking mechanism

**Blocks:** v0.2 microservices project type. Depends on decision #5.

When a new service is added, it must be registered in the gateway. Three approaches:

| Option | How it works | Notes |
|--------|-------------|-------|
| **Static config** | Gateway config file (nginx.conf, Kong declarative YAML, Traefik routes.yml) is patched when a service is added. | Simple, but `dot add service` must patch complex config files. |
| **Service discovery** | Gateway is configured to use a service registry (Consul, etcd). Services self-register at startup. | Dynamic, but adds infrastructure complexity. |
| **Env-driven** | Gateway reads service URLs from env vars. dot generates the env var definitions and a `.env.example`. | Portable, but manual coordination between services. |

**Decision needed before:** v0.2 microservices gateway generator design begins.

---

## 7. Multi-language monorepo — engine iteration

**Blocks:** v0.2 monorepo and microservices project types

The current engine runs `registry.ForSpec(spec)` once for a single Spec. In a monorepo with Go + TypeScript apps, each app has its own language and modules — the engine needs to run per app, not per repo.

Questions:
- Does `dot init` for a monorepo produce one Spec per app, or a single top-level Spec with nested app specs?
- Does `pipeline.Run` write to the root, or does each app get its own run scoped to its subdirectory?
- How does `.dot/config.json` store per-app state?

**Decision needed before:** v0.2 monorepo generator design begins.
