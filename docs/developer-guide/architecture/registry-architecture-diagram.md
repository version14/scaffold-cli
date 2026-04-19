# Registry Architecture — Mermaid Diagrams

Here are several mermaid charts visualizing the registry architecture from different angles.

---

## 1. Plugin Attachment System Overview

```mermaid
graph TD
    Registry["<b>Registry</b><br/>Question Trees + Plugins"]
    
    Base["<b>Base Questions</b><br/>name | type | language"]
    
    Type1["<b>Type 1 Plugin</b><br/>New Flow<br/>Attaches at: project_type"]
    Type2["<b>Type 2 Plugin</b><br/>Subflow<br/>Attaches at: any question"]
    
    QTemplate["<b>Question Templates</b><br/>shared definitions<br/>language | architecture | database"]
    
    Generators["<b>Generators</b><br/>Language × Module pairs"]
    
    Registry -->|manages| Base
    Registry -->|discovers| Type1
    Registry -->|discovers| Type2
    Registry -->|references| QTemplate
    Type1 -->|owns entire flow| Generators
    Type2 -->|extends question<br/>with new option| QTemplate
    QTemplate -->|maps answers to| Generators
```

---

## 2. Type 1 Plugin Example — Game Development

```mermaid
graph TD
    Base["<b>Phase 1</b><br/>name | type | language"]
    
    TypeQ["Project type?"]
    
    API["api → base flow<br/>language → architecture"]
    CLI["cli → base flow<br/>language → architecture"]
    Game["game → <b>GameDevPlugin</b>"]
    
    Engine["Engine?<br/>unity | godot | bevy"]
    Platform["Platform?<br/>desktop | mobile | web"]
    
    Gens["UnityGenerator<br/>GodotGenerator<br/>BevyGenerator<br/>DesktopGenerator<br/>etc."]
    
    Base --> TypeQ
    TypeQ -->|api| API
    TypeQ -->|cli| CLI
    TypeQ -->|game| Game
    
    Game --> Engine
    Engine --> Platform
    Platform --> Gens
    
    style Game fill:#ffcccc
    style Engine fill:#ffcccc
    style Platform fill:#ffcccc
    style Gens fill:#ffcccc
```

---

## 3. Type 2 Plugin Example — Multi-Language Support

```mermaid
graph TD
    Base["<b>Phase 1</b><br/>name | type | language"]
    Lang["Language?"]
    
    Go["go → <b>GoRegistry</b>"]
    TS["typescript → <b>TsRegistry</b>"]
    Rust["rust → <b>RustPlugin</b><br/>(Type 2, attached here)"]
    
    GoArch["<b>GoRegistry Phase 2</b><br/>Architecture?<br/>mvc | clean | hexagonal"]
    RustRuntime["<b>RustPlugin Phase 2</b><br/>Runtime?<br/>actix | axum | tokio"]
    
    GoGen["GoRestAPIGenerator<br/>GoCleanArchGenerator"]
    RustGen["RustActixGenerator<br/>RustAxumGenerator"]
    
    Base --> Lang
    Lang -->|go| Go
    Lang -->|typescript| TS
    Lang -->|rust| Rust
    
    Go --> GoArch
    GoArch --> GoGen
    Rust --> RustRuntime
    RustRuntime --> RustGen
    
    style Rust fill:#ccddff
    style RustRuntime fill:#ccddff
    style RustGen fill:#ccddff
```

---

## 4. Question Tree with Plugin Option Merging

```mermaid
graph TD
    Q["<b>Architecture Question</b><br/>(Official Template)"]
    
    Official["Official Options<br/>(from GoRegistry)<br/>mvc | clean | hexagonal"]
    Attached["Attached Plugin Options<br/>(EventDrivenPlugin)<br/>event-driven"]
    
    Merged["<b>Merged Options</b><br/>shown to user<br/>mvc | clean | hexagonal<br/>event-driven"]
    
    User["User selects<br/>event-driven"]
    
    EventPlugin["EventDrivenPlugin<br/>subflow takes over<br/>asks: Event bus?<br/>Gateway?"]
    
    Q -->|base| Official
    Q -->|discovered plugins| Attached
    Official --> Merged
    Attached --> Merged
    Merged --> User
    User --> EventPlugin
    
    style Attached fill:#ccddff
    style EventPlugin fill:#ccddff
    style Merged fill:#ffffcc
```

---

## 5. Full Two-Phase Survey with Plugin Decision Tree

```mermaid
graph TD
    P1["<b>PHASE 1: Base Questions</b>"]
    
    Name["name?"]
    Type["type?"]
    Lang["language?"]
    
    TypeDecision{"type =<br/>game?"}
    LangDecision{"language =<br/>rust?"}
    
    GamePath["✓ Type 1 Plugin<br/>GameDevPlugin<br/>owns entire flow"]
    RustPath["✓ Type 2 Plugin<br/>RustPlugin<br/>subflow"]
    BasePath["✓ Base Flow<br/>GoRegistry<br/>continues"]
    
    P2["<b>PHASE 2: Registry Question Tree</b><br/>(Type 2 plugins merge options)"]
    
    Spec["spec.Spec Assembled<br/>from Phase 1 + Phase 2 answers"]
    Resolve["Registry.ForSpec<br/>match by Language × Module"]
    Gens["Generators Matched<br/>Apply each → []FileOp"]
    
    P1 --> Name --> Type --> Lang
    
    Type --> TypeDecision
    TypeDecision -->|yes| GamePath
    TypeDecision -->|no| LangDecision
    
    Lang --> LangDecision
    LangDecision -->|yes| RustPath
    LangDecision -->|no| BasePath
    
    GamePath --> P2
    RustPath --> P2
    BasePath --> P2
    
    P2 --> Spec --> Resolve --> Gens
    
    style GamePath fill:#ffcccc
    style RustPath fill:#ccddff
    style BasePath fill:#ccffcc
```

---

## 6. Generator Registration — Conflict Detection Matrix

```mermaid
graph TD
    Reg["<b>Registry.Register</b><br/>Conflict Detection"]
    
    Gen1["GoRestAPIGenerator<br/>Language: go<br/>Module: rest-api"]
    Gen2["GoRedisGenerator<br/>Language: go<br/>Module: redis"]
    Gen3["DockerGenerator<br/>Language: *<br/>Module: docker"]
    Gen4["BadGenerator<br/>Language: go<br/>Module: rest-api"]
    Gen5["PythonDockerGenerator<br/>Language: python<br/>Module: docker"]
    
    OK1["✓ Different modules<br/>No conflict"]
    OK2["✓ Language-agnostic<br/>No conflict with specifics"]
    OK3["✓ Different languages<br/>No conflict"]
    CONFLICT["✗ CONFLICT!<br/>Same Language ×<br/>Module pair"]
    
    Reg --> Gen1 --> OK1
    Reg --> Gen2 --> OK1
    Reg --> Gen3 --> OK2
    Reg --> Gen4 --> CONFLICT
    Reg --> Gen5 --> OK3
    
    OK1 --> Success["✓ All registered<br/>startup continues"]
    OK2 --> Success
    OK3 --> Success
    CONFLICT --> Error["✗ Startup panic<br/>conflict error"]
    
    style Gen1 fill:#ccffcc
    style Gen2 fill:#ccffcc
    style Gen3 fill:#ccffcc
    style Gen4 fill:#ffcccc
    style Gen5 fill:#ccffcc
    style Success fill:#ccffcc
    style Error fill:#ff9999
```

---

## 7. Module List Derivation — Auto-Generated from Reachable Generators

```mermaid
graph TD
    GoReg["<b>GoRegistry</b><br/>in current branch"]
    
    Gen1["GoRestAPIGenerator<br/>module: rest-api"]
    Gen2["GoPostgresGenerator<br/>module: postgres"]
    Gen3["GoRedisGenerator<br/>module: redis"]
    Gen4["GoAuthJWTGenerator<br/>module: auth-jwt"]
    
    Reachable["<b>Reachable Generators</b><br/>from current position<br/>in question tree"]
    
    ModuleQ["<b>Modules question</b><br/>auto-generated<br/>options list"]
    
    Options["Rendered options:<br/>[REST API<br/>PostgreSQL<br/>Redis<br/>JWT Auth]"]
    
    User["User selects:<br/>postgres + redis"]
    
    Match["Registry.ForSpec<br/>matches<br/>GoPostgresGenerator<br/>GoRedisGenerator"]
    
    GoReg -->|contains| Gen1
    GoReg -->|contains| Gen2
    GoReg -->|contains| Gen3
    GoReg -->|contains| Gen4
    
    Gen1 -->|traversable| Reachable
    Gen2 -->|traversable| Reachable
    Gen3 -->|traversable| Reachable
    Gen4 -->|traversable| Reachable
    
    Reachable -->|derive module list| ModuleQ
    ModuleQ -->|display| Options
    Options --> User
    User --> Match
    
    style Reachable fill:#ffffcc
    style Options fill:#ffffcc
    style Match fill:#ccffcc
```

---

## 8. Survey Engine at Runtime — Registry Resolution

```mermaid
graph TD
    P1Answer["<b>Phase 1 Complete</b><br/>User answered:<br/>type=api<br/>language=go"]
    
    Lookup["<b>Registry Lookup</b><br/>find registry for 'go'"]
    
    Registries["Available Registries:<br/>GoRegistry<br/>TsRegistry<br/>RustPlugin<br/>PythonRegistry<br/>..."]
    
    Selected["Selected: GoRegistry"]
    
    P2Start["<b>Phase 2 Begins</b><br/>render GoRegistry.QuestionTree"]
    
    ArchQ["Architecture question<br/>(template: architecture)"]
    
    PluginCheck["Check for Type 2 plugins<br/>attached to 'architecture'"]
    
    PluginsFound["Plugins found:<br/>EventDrivenPlugin"]
    
    Merge["Merge options:<br/>base: mvc, clean, hexagonal<br/>plugin: event-driven"]
    
    Display["Render merged question<br/>to user terminal"]
    
    UserPick["User picks:<br/>clean"]
    
    P1Answer --> Lookup
    Lookup --> Registries
    Registries --> Selected
    Selected --> P2Start
    P2Start --> ArchQ
    ArchQ --> PluginCheck
    PluginCheck --> PluginsFound
    PluginsFound --> Merge
    Merge --> Display
    Display --> UserPick
    
    style Selected fill:#ccffcc
    style PluginsFound fill:#ccddff
    style Merge fill:#ffffcc
    style Display fill:#ffffcc
```

---

## 9. Architecture Evolution: Current (v0.1) vs Planned (v0.2+)

```mermaid
graph TD
    subgraph v01["<b>v0.1 — Current Implementation</b>"]
        A["Registry = flat list<br/>of generators"]
        B["ForSpec matching<br/>Language × Module<br/>simple equality check"]
        C["Hard-coded options<br/>in cmd_init.go<br/>no plugins"]
        D["No dynamic<br/>option merging"]
    end
    
    subgraph v02["<b>v0.2+ — Planned Redesign</b>"]
        E["Registry = question tree<br/>with plugin attachment points"]
        F["Plugin attachment<br/>at question templates<br/>Type 1 &amp; Type 2"]
        G["Dynamic option merging<br/>base + all attached plugins"]
        H["Community contributors<br/>can extend survey<br/>without forking"]
    end
    
    A -->|evolves| E
    B -->|evolves| F
    C -->|evolves| G
    D -->|evolves| H
    
    style v01 fill:#ffffcc
    style v02 fill:#ccffcc
```

---

## 10. Complete System Diagram — init → spec.Spec → Generators → FileOps

```mermaid
graph TD
    Survey["<b>Survey Engine</b><br/>(Phase 1 + Phase 2<br/>with plugins)"]
    
    Answers["User Answers<br/>name | type | language<br/>modules | architecture<br/>linter | database"]
    
    Spec["<b>spec.Spec</b><br/>Project { name, type, language }<br/>Modules [ { name, config } ]<br/>Config { linter, architecture, ... }<br/>Extensions { ... }"]
    
    Match["<b>Registry.ForSpec</b><br/>match generators by<br/>Language × Module"]
    
    Gens["Matched Generators<br/>GoRestAPIGenerator<br/>GoPostgresGenerator<br/>DockerGenerator"]
    
    Apply["<b>Apply Phase</b><br/>for each generator:<br/>Apply(spec) → []FileOp"]
    
    Ops["<b>Collected FileOps</b><br/>Create: main.go<br/>Create: internal/postgres.go<br/>Create: Dockerfile<br/>Patch: go.mod imports<br/>Append: docker-compose.yml"]
    
    Pipeline["<b>Pipeline</b><br/>resolve conflicts<br/>sort by priority<br/>write to disk"]
    
    Files["<b>Project Files</b><br/>main.go<br/>internal/postgres.go<br/>Dockerfile<br/>docker-compose.yml<br/>go.mod"]
    
    Config[".dot/config.json<br/>CommandDefs from<br/>each generator"]
    
    Prompts[".dot/prompts.md<br/>Append ops<br/>for LLM handoff"]
    
    Survey --> Answers
    Answers --> Spec
    Spec --> Match
    Match --> Gens
    Gens --> Apply
    Apply --> Ops
    Ops --> Pipeline
    Pipeline --> Files
    Pipeline --> Config
    Pipeline --> Prompts
    
    style Spec fill:#ffffcc
    style Match fill:#ccffcc
    style Gens fill:#ccffcc
    style Ops fill:#ffffcc
    style Pipeline fill:#ffffcc
    style Files fill:#ccffcc
```

---

## Key Concepts Summary Table

| Concept | Definition | Example |
|---------|-----------|---------|
| **Question Template** | Reusable survey question definition with ID and base options | `template:"architecture"` with options `[mvc, clean, hexagonal]` |
| **Type 1 Plugin** | Entire new flow branch attached at `project_type` | `GameDevPlugin` replaces language/architecture flow |
| **Type 2 Plugin** | Extension of existing question, adds new options | `RustPlugin` adds `rust` option to `language` question |
| **Attachment Point** | Question template ID where a plugin adds options | `"language"`, `"architecture"`, `"database"` |
| **Module Derivation** | Auto-generate module list from reachable generators | Show only modules this registry can handle |
| **Conflict Matrix** | Rules for which generators can coexist | Same `(Language, Module)` pair = registration error |
| **Plugin Discovery** | At runtime, survey engine finds all plugins for a question | Merge base + plugin options before displaying |
| **Spec Assembly** | Build `spec.Spec` from Phase 1 + Phase 2 answers | Each official template maps to typed `spec.CoreConfig` field |
| **Registry.ForSpec** | Match generators to a Spec by Language × Module | Called after Spec assembly to activate generators |
