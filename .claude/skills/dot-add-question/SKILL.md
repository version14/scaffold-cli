---
name: dot-add-question
description: Add a question (TextQuestion, ConfirmQuestion, OptionQuestion, LoopQuestion, IfQuestion) to a `dot` flow — wires the node into the graph, updates the flow's `Generators` resolver, asks whether any existing generators need to react to the new branch, and extends the matching `tools/test-flow/testdata/` fixtures. Trigger when the user says "add a question", "ajoute une question", "new flow question", or describes a new choice the wizard should offer.
source: local-git-analysis
version: 1.0.0
---

# dot — Add a Question

Questions live in `flows/<flow>.go`. They form a directed-acyclic graph traversed by `FlowEngine`. Each question has a stable `ID_` — the same ID is read later by the `Generators` resolver and by the generators themselves through `ctx.Answers`.

This skill walks adding a question safely: classify its impact on existing generators, announce, wire, then test.

---

## HARD RULES (Iron Law — never skip)

1. **Announce first, generate second.** Before any code change, output a plan stating:
   - The question's purpose, type, label, options/default, and stable `ID_`
   - Where it slots into the existing graph (which question now points to it, where it points next)
   - Which generators consume the new answer key (via `ctx.Answers["<ID_>"]`)
   - Whether existing generators need their `manifest.Outputs` / behaviour modified to react to the new branch — and what those modifications are
   - Which `tools/test-flow/testdata/*.json` fixtures need new `answers` keys and updated `expected_visited`
   Then **ASK** the user to confirm before touching code.
2. **Stable IDs are forever.** `ID_` is the key persisted to `.dot/spec.json` and read on `dot update`. Never rename an existing ID. New IDs must be `kebab-case` (existing convention: `ts-backend-framework`, `enable-auth`, `monorepo_type`, …).
3. **Every reachable branch must converge.** Every new edge (`Then`, `Else`, `Option.Next`, `Next_`) must eventually reach `&flow.Next{End: true}` (typically via the shared `confirmGenerate` terminal). Dangling edges hang the TUI.
4. **Update the resolver.** If the new answer should pull in or drop a generator, edit the `Generators` function (e.g. `resolveMonorepoGenerators` in `flows/init.go`).
5. **Update or add fixtures.** Every `tools/test-flow/testdata/*.json` that traverses the new branch must include the new key in `answers` and the new ID in `expected_visited`. Add new fixtures to cover every distinct branch.
6. **Bump versions on impacted generators.** If you change a generator's behaviour because of a new answer (new conditional output, new dep), bump `Manifest.Version` and update `docs/contributor/generators/<name>.md`.
7. **Update documentation — always, not only for Case C.** After every flow change, update the affected contributor docs:
   - Flow graph changes (new question, new branch, LoopQuestion body expansion): update `docs/contributor/authoring-flows.md` and any loop/resolver examples that are now stale.
   - New generator wired via the resolver: add/update a row in the Built-in generators table in `docs/contributor/authoring-generators.md`.
   - New architectural pattern established (new question type, new resolver helper): update the relevant section of `docs/contributor/architecture.md`.
   Skip only when the change is truly invisible at the doc level (e.g. a fixture-only tweak).

---

## Step-by-step workflow

### 1. Classify the impact

Determine which case applies:

| Case | What you must do |
|------|------------------|
| **A. Pure-routing question** (only redirects flow, no generator consumes it) | Add the question node, wire edges, extend fixtures. No generator changes. |
| **B. New answer selects existing generators** (e.g. choose between two ORMs already implemented) | Add the question, then update the `Generators` resolver branches. No generator code changes. |
| **C. New answer requires generator changes** (a sibling generator must behave differently on this choice) | Add the question + update the resolver + modify each impacted generator (manifest, files, version bump, doc). |
| **D. New answer needs a generator that does not exist** | **Stop**: use `dot-add-generator` first, then return to this skill. |

If unsure, ASK the user explicitly which case applies.

### 2. Pre-generation announcement (MANDATORY)

Output text like this and stop for confirmation:

```
I will add a <type> question `<ID_>` to flow `<flow-id>` (file: flows/<flow>.go).

Label: "<label>"
Default / Options: <…>
Position: inserted between `<prev>` and `<next>` (Then=…, Else=…)

Resolver impact (flows/<flow>.go → <ResolverFn>):
- When `<ID_>` == "<value>", include generators: <list>
- When `<ID_>` == "<other>", drop generators: <list>

Generator impact:
- generators/<name>/ — <what changes, why> — version <old> → <new>
- docs/contributor/generators/<name>.md — sections to update

Fixture impact:
- tools/test-flow/testdata/<file>.json — add answers["<ID_>"] = …, append "<ID_>" to expected_visited
- New fixtures to add: <list with reasons>

Proceed?
```

### 3. Pick the right Question type

| Type | When to use |
|------|-------------|
| `TextQuestion` | Free-text input (project name, port, path) — supports `Default` and `Validate` |
| `ConfirmQuestion` | Yes/no with **separate** branches (`Then` / `Else`) — use when each branch leads somewhere different |
| `OptionQuestion` | Single or multi-select; each `Option.Next` controls its own destination |
| `LoopQuestion` | Repeat a sub-body N times; `Body` questions terminate iterations with `Next{End: true}`; the loop itself advances via `Continue` |
| `IfQuestion` | **Not asked** — silent branch on a computed `Condition(ctx)` over earlier answers |

Cross-check against existing patterns in `flows/init.go`.

### 4. Edit the flow

Open the flow file (most commonly `flows/init.go`). Build nodes **bottom-up** — the terminal `confirmGenerate` exists first; declare the new node above any predecessor that will reference it; then update the predecessor's edge.

Worked example — inserting a new `OptionQuestion` between `ts-backend-architecture` and `ts-backend-decorators-validation`:

```go
// new question
cacheChoice := &flow.OptionQuestion{
    QuestionBase: flow.QuestionBase{ID_: "ts-backend-cache"},
    Label:        "Cache layer",
    Description:  "Optional in-memory or Redis cache for the API.",
    Options: []*flow.Option{
        {Label: "None",   Value: "none",   Next: &flow.Next{Question: decorators}},
        {Label: "Memory", Value: "memory", Next: &flow.Next{Question: decorators}},
        {Label: "Redis",  Value: "redis",  Next: &flow.Next{Question: decorators}},
    },
}

// update the predecessor's edge to point at the new node instead of `decorators`
architecture := &flow.OptionQuestion{
    QuestionBase: flow.QuestionBase{ID_: "ts-backend-architecture"},
    Label:        "Choose your architecture.",
    Options: []*flow.Option{
        {Label: "Clean Architecture", Value: CLEAN_ARCHITECTURE, Next: &flow.Next{Question: cacheChoice}},
        {Label: "MVC",                Value: MVC_ARCHITECTURE,   Next: &flow.Next{Question: cacheChoice}},
    },
}
```

For values reused across the codebase (architectures, validation libs), declare a top-of-file constant — see `CLEAN_ARCHITECTURE`, `ValidationLibZod` in `flows/init.go`.

### 5. Update the resolver

In the same file, the `Generators` function reads `s.Answers["<ID_>"]`. Add the new conditional. Example:

```go
cache, _ := s.Answers["ts-backend-cache"].(string)
switch cache {
case "memory":
    out = append(out, Invocation{Name: "express_memory_cache"})
case "redis":
    out = append(out,
        Invocation{Name: "redis_docker_compose"},
        Invocation{Name: "express_redis_cache"},
    )
}
```

Keep the order topologically sane — `DependsOn` will fix detailed ordering, but related groups stay together.

### 6. Update affected generators

For Case C only: each generator that reads the new answer needs:

- Code change in `Generate()` (or template `.tmpl` change in `files/`)
- `Manifest.Version` bump (semver: patch for fix, minor for new behaviour)
- `docs/contributor/generators/<name>.md` updated — Identity (version) and any of Answers consumed / Files written / Validators / PostGenerationCommands / TestCommands sections that changed

### 7. Update fixtures

In `tools/test-flow/testdata/`:

- For every existing fixture whose `expected_visited` passes through the modified branch, add the new key to `answers` and the new `ID_` to `expected_visited` (in the correct order).
- Add a new fixture for each new branch value that nothing else covers. Naming convention: `YYYYMMDDhhMM_<flow>_<short_variant_description>.json`.
- Cross-reference `tools/test-flow/testdata/0_template.json` for required fields (`name`, `flow_id`, `answers`, `expected_visited`, `skip_post_commands`, `skip_test_commands`).

### 8. Update documentation

For every flow change, update docs even if no generator code changed:

- **`docs/contributor/authoring-flows.md`** — update examples, tables, or the LoopQuestion section if body structure, `Continue` wiring, or `buildPerAppBody` patterns changed.
- **`docs/contributor/authoring-generators.md`** — update the Built-in generators table when a new generator is wired via the resolver; update `PostGenerationCommands`/`WorkDir` notes when command scoping changes.
- **`docs/contributor/architecture.md`** — update only when the pipeline itself changes (executor scoping, state prefix, command planning).

Prefer targeted edits to existing sections over writing new files. Do not write doc unless it is currently missing or factually wrong.

### 9. Verify

```
make build
make test         # unit tests
make test-flows   # exercises all fixtures end-to-end
```

If `make test-flows` fails on a fixture whose new branch you added, fix that branch (generator code or fixture answers) — do not edit `expected_visited` to silence the failure.

### 10. Commit

Conventional Commits format:

- `feat(flow): add <ID_> question to <flow-id>` for the flow edit
- `feat(generators): wire <ID_> into <generator>` plus version bump
- `test(test-flow): add fixture for <variant>` for new fixtures

Do **not** auto-commit — wait for explicit user instruction.

---

## Question authoring checklist

- [ ] `ID_` is kebab-case, unique across the entire flow, and never collides with an existing key
- [ ] Every outgoing edge eventually reaches `Next{End: true}`
- [ ] `Default` is set when the user can sensibly skip (TextQuestion, ConfirmQuestion)
- [ ] `Validate` is set for TextQuestion when empty input is invalid — reuse `nonEmpty` from `flows/init.go`
- [ ] `Description` is set when the label alone is ambiguous (decorator example in `flows/init.go` is a good model)
- [ ] All `Option.Value` strings are stable (these end up in `.dot/spec.json`) — extract constants when reused
- [ ] LoopQuestion body terminates each iteration with `Next{End: true}`, not the flow terminal
- [ ] IfQuestion has no Label/Default (it is never rendered)
- [ ] Resolver covers every reachable value of the new answer
- [ ] Any generator newly wired by this question declares non-empty `PostGenerationCommands` **and** `TestCommands` (use `dot-add-generator` Iron Law #5 as the standard)
- [ ] Affected `docs/contributor/` files updated (authoring-flows, authoring-generators Built-in table, architecture if pipeline changed)

---

## Reference patterns from this repo

| Pattern | Example in `flows/init.go` |
|---------|----------------------------|
| ConfirmQuestion with separate then/else paths | `enable-auth` → authMethod vs confirmGenerate |
| OptionQuestion converging to one next | `ts-backend-formatter` (biome/prettier) → linter |
| Branch off / converge back | `enable-db` Then → dbType … Else → confirmGenerate |
| Loop + body | `apps_count` LoopQuestion wrapping `stack` |
| Stable value constants | `CLEAN_ARCHITECTURE`, `ValidationLibZod` |
| Default-true ConfirmQuestion | `ts-backend-decorators-validation` |

---

## Common mistakes (do not make)

- Renaming an existing `ID_` — breaks `dot update` for projects scaffolded earlier
- Forgetting to update `expected_visited` in fixtures — `make test-flows` will fail
- Adding a question without updating the resolver when a generator should be pulled in / dropped
- Wiring `Then`/`Else` to the same edge on a ConfirmQuestion (use OptionQuestion if both branches converge — or use IfQuestion when the choice is computable)
- Skipping the pre-generation announcement (Iron Law #1)
- Modifying a generator's behaviour without bumping `Manifest.Version` and updating its doc
- Wiring a new generator without verifying it has non-empty `PostGenerationCommands` and `TestCommands`
