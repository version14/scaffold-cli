# Developer Guide — dot

This guide is for contributors and people building on top of dot.
It covers how dot works internally, how to build generators, and where the project is going.

---

## Contents

### Architecture
How dot is structured and why.

- [Overview](architecture/overview.md) — the big picture: input layers, engine, pipeline
- [Repository structure](architecture/repository-structure.md) — package layout and the progressive split plan
- [Input layers](architecture/input-layers.md) — CLI TUI, dot.yaml (PaC), future layers
- [Registry design](architecture/registry-design.md) — how registries are question trees, how community registries plug in
- [Registry architecture diagrams](architecture/registry-architecture-diagram.md) — visual diagrams of plugins, question trees, and two-phase survey

### Internals
How the core engine works, piece by piece.

- [Spec](internals/spec.md) — the Spec type: what it contains, how it is produced
- [FileOp pipeline](internals/fileop-pipeline.md) — how file operations are collected, resolved, and written
- [Generator interface](internals/generator-interface.md) — the Generator contract: methods, rules, constraints
- [Registry](internals/registry.md) — how generators are registered and resolved against a Spec
- [ProjectContext](internals/project-context.md) — .dot/config.json and .dot/manifest.json
- [Conflict resolution](internals/conflict-resolution.md) — how dot handles user-modified files

### Generators
How generators work and how to build one.

- [Generator spec](generators/generator-spec.md) — what a generator is, what it can and cannot do, how it works
- [Authoring guide](generators/authoring-guide.md) — step-by-step: implement the interface, register, test
- [FileOp reference](generators/fileop-reference.md) — all FileOp kinds and their constraints
- [Patch strategies](generators/patch-strategies.md) — AnchorImportBlock, AnchorMainFunc, AnchorInitFunc
- [Official generators](generators/official-generators.md) — what ships with dot today

### Roadmap
Where the project is going, version by version.

- [v0.1](roadmap/v0.1.md) — CLI loop, done
- [v0.2](roadmap/v0.2.md) — more official generators + local custom generators
- [v0.3](roadmap/v0.3.md) — `dot add module` + conflict resolution
- [v0.4](roadmap/v0.4.md) — public generator registry
- [v0.5](roadmap/v0.5.md) — Project as Code (`dot.yaml`, `dot plan`, `dot apply`)
- [v1.0](roadmap/v1.0.md) — full custom project creation, everything works together
- [v1.1](roadmap/v1.1.md) — MCP server
- [v1.x](roadmap/v1.x.md) — future directions (database definitions, dashboard, more)
- [Open decisions](roadmap/open-decisions.md) — unresolved questions that block future work

---

## Contributing

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for setup, commit conventions, and PR process.
See [code-style.md](guidelines/code-style.md) for Go style rules.
