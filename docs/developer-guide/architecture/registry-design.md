# Registry Design

> **See also:** [Generator spec](../generators/generator-spec.md) — what generators are and what they do. See the "Init phase" and "Post-init phase" diagrams in that document for a detailed walkthrough of the two-phase survey that the registry drives.

---

## The problem this solves

The original registry was a flat list of generators. Adding a new language meant editing
`cmd_init.go` — hardcoding new language options, new module lists, new linter choices.
There was no way for a community contributor to extend the flow without forking core code.

The registry design makes the survey entirely driven by pluggable question trees. The base
flow is minimal and stable. Everything else is a plugin.

---

## Two plugin types

There are exactly two ways to extend the flow. They differ only in where they attach.

### Type 1 — New flow (attaches at project type)

Introduces a brand new project type option. The plugin owns the entire flow from that
point — its own questions, its own modules, its own generators. Nothing from the base
language/architecture flow applies.

Use this when you are adding a fundamentally different kind of project that doesn't fit
the existing type options.

```
base: Project type?
        ├── api        → base flow (language → architecture → ...)
        ├── cli        → base flow
        ├── frontend   → base flow
        └── game       → GameDevPlugin  ← Type 1 plugin, owns everything from here
                          ├── "Engine?" → [UnityGenerator | GodotGenerator | BevyGenerator]
                          └── "Platform?" → [DesktopGenerator | MobileGenerator | WebGenerator]
```

Examples: game development, data science notebooks, embedded systems, mobile apps.

### Type 2 — Subflow (attaches at any specific question)

Adds a new option to any existing question anywhere in the flow. When the user picks that
option, the plugin's subflow runs from that branch point. The rest of the base flow
continues normally after the subflow completes.

Use this when you want to extend an existing path — adding a new language, a new
architecture pattern, a new database, a new deployment target — without replacing the
whole flow.

```
base: Language?
        ├── go         → GoRegistry
        ├── typescript → TsRegistry
        └── rust       → RustPlugin  ← Type 2 plugin, attached at "language" question
                          ├── "Runtime?" → [ActixGenerator | AxumGenerator | TokioGenerator]
                          └── "ORM?"     → [SQLxGenerator | DieselGenerator | none]

base (within GoRegistry): Architecture?
        ├── mvc        → GoRestAPIGenerator
        ├── clean      → GoCleanArchGenerator
        └── event-driven → EventDrivenPlugin  ← Type 2 plugin, attached at "architecture"
                           ├── "Event bus?" → [NATSGenerator | gRPCGenerator]
                           └── "Gateway?"   → [LocalGatewayGenerator | CloudGatewayGenerator]
```

The attachment point is any question template ID. A plugin declares: "I want to add an
option to question X." The survey engine discovers all plugins attached to a given question
and merges their options into that question's list.

---

## Question templates

A question template is a shared definition — ID, title, description, and base options —
stored once and referenced by ID across all plugins and registries.

```
QuestionTemplate{
    ID:      "architecture",
    Title:   "Architecture pattern",
    Options: [mvc, clean, hexagonal],
}

QuestionTemplate{
    ID:      "language",
    Title:   "Language",
    Options: [go, typescript],   ← base options; plugins extend this list
}

QuestionTemplate{
    ID:      "database",
    Title:   "Database",
    Options: [postgres, mysql, mongodb, redis, none],
}
```

A RustPlugin and a GoRegistry both use `template:"architecture"` — the user sees the same
question text and the same base options. Each plugin may also add its own options to the
question. The answer activates whichever generators are wired to that option in the
current plugin's branch map.

**The question is defined once. Plugins wire answers to their own generators.**

**Language wildcard resolution:** When a generator uses `Language() = "*"` to match any project language, it is activated alongside language-specific generators. If both a wildcard generator and a language-specific generator claim the same file at the same priority, the pipeline returns a conflict error — use the `Priority` field to resolve (higher priority wins). See [Generator spec — Priority](../generators/generator-spec.md) for details.

---

## How the survey engine works

The survey engine resolves the full question tree at runtime by merging the base questions
with all registered plugins, then runs it in two phases.

### Phase 1 — base questions

Always the same. Hardcoded in `cmd_init.go`. Three questions only:

```
Project name?
Project type?   [api / cli / frontend / monorepo / ...]  ← Type 1 plugins add options here
Language?       [go / typescript / ...]                   ← Type 2 plugins add options here
```

After Phase 1, the engine knows the project type and language. It looks up which registry
(built-in or plugin) handles that combination.

### Phase 2 — plugin-driven question tree

The selected registry's question tree is traversed. At each node, the engine checks
whether any Type 2 plugins are attached to that question template and merges their options
in. When the user picks an option owned by a plugin, that plugin's subflow takes over from
that branch point.

```
Phase 2 traversal at "architecture" node:
  base options:  mvc, clean, hexagonal       (from GoRegistry)
  plugin options: event-driven               (from EventDrivenPlugin, attached here)
  ───────────────────────────────────────────
  rendered options: mvc / clean / hexagonal / event-driven

user picks "event-driven"
  → EventDrivenPlugin.subflow runs
  → asks: "Event bus?" then "Gateway?"
  → activates NATSGenerator + LocalGatewayGenerator
  → returns to parent flow
```

The `huh` form for Phase 2 is built programmatically — one `huh.Group` per question node.
This is a second `form.Run()` call after Phase 1 completes.

---

## Plugin declaration

A plugin declares its type, its attachment point, and its subflow:

> **Note:** This is pseudocode showing the conceptual structure. See `internal/registry/plugin.go` for actual struct field names and types.

```go
// Type 1 — entirely new project type
registry.RegisterPlugin(registry.Plugin{
    Kind:        registry.NewFlow,
    AttachTo:    "project_type",       // always "project_type" for Type 1
    OptionLabel: "Game",
    OptionValue: "game",
    Flow: &registry.QuestionNode{
        Template: registry.NewTemplate("engine", "Game engine", []string{"unity", "godot", "bevy"}),
        Branches: map[string]registry.Branch{
            "unity": {Generators: []generator.Generator{&UnityGenerator{}}},
            "godot": {Generators: []generator.Generator{&GodotGenerator{}}},
            "bevy":  {Generators: []generator.Generator{&BevyGenerator{}}},
        },
    },
})

// Type 2 — subflow at a specific question
registry.RegisterPlugin(registry.Plugin{
    Kind:        registry.Subflow,
    AttachTo:    "language",           // any question template ID
    OptionLabel: "Rust",
    OptionValue: "rust",
    Flow: &registry.QuestionNode{
        Template: registry.TemplateRef("architecture"),  // reuse shared template
        Branches: map[string]registry.Branch{
            "mvc": {
                Generators: []generator.Generator{&RustActixGenerator{}},
                Next: &registry.QuestionNode{
                    Template: registry.TemplateRef("database"),
                    Branches: map[string]registry.Branch{
                        "postgres": {Generators: []generator.Generator{&RustSQLxGenerator{}}},
                        "none":     {},
                    },
                },
            },
            "clean": {Generators: []generator.Generator{&RustAxumCleanGenerator{}}},
        },
    },
})
```

---

## Spec assembly

Phase 1 answers go into typed `spec.Spec` fields:
- `spec.Project.Name`, `spec.Project.Type`, `spec.Project.Language`

Phase 2 answers from official question templates go into `spec.CoreConfig` typed fields:
- `template:"architecture"` → `spec.Config.Architecture`
- `template:"linter"`       → `spec.Config.Linter`
- `template:"database"`     → `spec.Modules[].Name`

Phase 2 answers from plugin-specific templates (not in the core set) go into
`spec.Extensions`:
- `template:"engine"`       → `spec.Extensions["engine"]`
- `template:"rust-runtime"` → `spec.Extensions["rust-runtime"]`

Community generators read from `spec.Extensions` for their plugin-specific answers. They
read from `spec.CoreConfig` for cross-language answers (architecture pattern, linter) that
official templates already define.

---

## Module list derivation

Module options shown during Phase 2 are not hardcoded. They are derived from the generator
leaves reachable in the current registry's question tree plus any attached plugins.

```
GoRegistry reachable generators:
  GoRestAPIGenerator  (module: rest-api)
  GoPostgresGenerator (module: postgres)
  GoRedisGenerator    (module: redis)
  GoAuthJWTGenerator  (module: auth-jwt)

→ Module question shows: [REST API, PostgreSQL, Redis, JWT auth]

RustPlugin (attached at language=rust) reachable generators:
  RustActixGenerator  (module: rest-api)
  RustSQLxGenerator   (module: postgres)

→ Module question shows: [REST API, PostgreSQL]
```

Each plugin exposes only the modules its generators handle. The survey engine builds the
module list per-registry, never globally.

---

## ASCII — full flow with both plugin types

```
dot init
═══════════════════════════════════════════════════════════════════
PHASE 1 — base questions
───────────────────────────────────────────────────────────────────
  name?

  project type?
    ├── api              → base flow continues
    ├── cli              → base flow continues
    ├── frontend         → base flow continues
    └── game             ← Type 1 plugin (GameDevPlugin)
           │
           ▼
    [GameDevPlugin subflow — completely owns flow from here]
    engine? → unity/godot/bevy
    platform? → desktop/mobile/web
    → generators activated, Spec assembled, done

  language?   (shown only for non-Type-1 paths)
    ├── go               → GoRegistry (built-in)
    ├── typescript       → TsRegistry (built-in)
    └── rust             ← Type 2 plugin (RustPlugin, attached at "language")

═══════════════════════════════════════════════════════════════════
PHASE 2 — registry question tree (dynamic, per selected language)
───────────────────────────────────────────────────────────────────
  [GoRegistry selected — user picked "go"]

  architecture?               ← official template "architecture"
    ├── mvc                   → GoRestAPIGenerator
    ├── clean                 → GoCleanArchGenerator
    ├── hexagonal             → GoHexagonalGenerator
    └── event-driven          ← Type 2 plugin (EventDrivenPlugin, attached at "architecture")
           │
           ▼
    [EventDrivenPlugin subflow]
    event bus? → nats/grpc/none
    gateway?   → local/cloud
    → NATSGenerator + GatewayGenerator activated

  modules? (options auto-derived from reachable generators)
    ├── postgres  → GoPostgresGenerator
    ├── redis     → GoRedisGenerator
    └── auth-jwt  → GoAuthJWTGenerator

  linter?    ← official template "linter"
  formatter? ← official template "formatter"

═══════════════════════════════════════════════════════════════════
ASSEMBLY
───────────────────────────────────────────────────────────────────
  spec.Spec assembled from Phase 1 + Phase 2 answers
  All activated generators: Apply(spec) → []FileOp
  Pipeline → files on disk
           → .dot/config.json (CommandDefs)
           → .dot/prompts.md  (if any generator emitted it)
```
