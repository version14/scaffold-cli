# Authoring Flows

A **flow** is the question graph that DOT presents to the user before generating a project. This guide explains how to write one, how branching works, how to use loops, and how to register the flow so it appears in `dot scaffold`.

---

## Table of Contents

- [What is a flow?](#what-is-a-flow)
- [FlowDef structure](#flowdef-structure)
- [Question types](#question-types)
- [Wiring questions together (Next)](#wiring-questions-together-next)
- [Branching](#branching)
- [Loops](#loops)
- [Conditional nodes (IfQuestion)](#conditional-nodes-ifquestion)
- [The Generators resolver](#the-generators-resolver)
- [Registering a flow](#registering-a-flow)
- [Built-in flows](#built-in-flows)
- [Testing a flow](#testing-a-flow)

---

## What is a flow?

A flow is a directed acyclic graph of `Question` nodes. The `FlowEngine` traverses it starting from the root question. At each node the engine calls the `FlowAdapter` (the interactive TUI or a scripted adapter in tests) to collect an answer, then follows the `Next` edge the question returns.

When the engine reaches a `Next{End: true}` edge, traversal stops and the collected answers are returned as a `FlowContext`.

---

## FlowDef structure

```go
// flows/registry.go
type FlowDef struct {
    ID          string
    Title       string
    Description string
    Root        flow.Question
    Generators  func(*spec.ProjectSpec) []Invocation
}
```

| Field | Required | Notes |
|-------|----------|-------|
| `ID` | Yes | Unique kebab-case identifier (e.g. `"monorepo"`, `"my-stack"`). Used in `dot scaffold <id>`. |
| `Title` | Yes | Short human-readable name shown in `dot flows`. |
| `Description` | No | Longer description shown below the title. |
| `Root` | Yes | The first question the engine will ask. |
| `Generators` | Yes | Resolver function that maps answers to generator invocations. |

---

## Question types

All question types live in `internal/flow/question.go`. Import them via `pkg/dotplugin` if you are writing a plugin; use `internal/flow` directly if you are writing a built-in flow in the `flows/` package.

### TextQuestion

Free-text input.

```go
&flow.TextQuestion{
    QuestionBase: flow.QuestionBase{
        ID_:   "project_name",
        Next_: &flow.Next{Question: nextQuestion},
    },
    Label:       "Project name",
    Description: "Used as the module name and directory.",
    Default:     "my-project",
    Validate:    func(s string) error { /* return nil or error */ },
}
```

| Field | Notes |
|-------|-------|
| `ID_` | Must be unique across the entire flow. Used as the key in `FlowContext.Answers`. |
| `Next_` | Edge to follow after the user answers. |
| `Default` | Pre-filled value in the TUI. |
| `Validate` | Optional. Called before advancing; return `nil` to accept, non-nil to show an error. |

### ConfirmQuestion

Boolean yes/no question with separate branches.

```go
&flow.ConfirmQuestion{
    QuestionBase: flow.QuestionBase{ID_: "include_docker"},
    Label:        "Include a Dockerfile?",
    Default:      true,
    Then:         &flow.Next{Question: dockerPortQuestion},
    Else:         &flow.Next{Question: confirmGenerate},
}
```

| Field | Notes |
|-------|-------|
| `Then` | Edge when the user answers **yes**. |
| `Else` | Edge when the user answers **no**. |

### OptionQuestion

Single or multi-select from a list of options.

```go
&flow.OptionQuestion{
    QuestionBase: flow.QuestionBase{ID_: "css_framework"},
    Label:        "CSS framework",
    Multiple:     false, // true = checkbox list
    Options: []*flow.Option{
        {Label: "Tailwind", Value: "tailwind", Next: &flow.Next{Question: confirmGenerate}},
        {Label: "CSS Modules", Value: "css_modules", Next: &flow.Next{Question: confirmGenerate}},
        {Label: "None", Value: "none", Next: &flow.Next{Question: confirmGenerate}},
    },
}
```

For `Multiple: false` (default), each `Option.Next` controls where selecting that option leads. For `Multiple: true`, all options share the `QuestionBase.Next_` edge.

### LoopQuestion

Collect a fixed body of sub-questions N times, where N comes from a preceding count input. Returns `[]map[string]Answer` — one map per iteration.

```go
&flow.LoopQuestion{
    QuestionBase: flow.QuestionBase{ID_: "apps_count"},
    Label:        "Configure app",
    Body:         buildPerAppBody(confirmGenerate), // []flow.Question
    Continue:     &flow.Next{Question: confirmGenerate},
}
```

**How the TUI runs it:**

1. A preceding `TextQuestion` asks how many times the body should repeat (e.g. "Number of apps"). Its answer is stored as a string at the loop question's ID.
2. The `HuhFormRunner` renders the `Body` as a separate sub-form for each iteration and stores the collected answers.
3. After all iterations complete, the `Continue` edge is rendered as a post-loop sub-form (not part of the body). This is how post-loop questions (e.g. `confirmGenerate`) appear after per-app questions without appearing between iterations.

**Iteration count** is resolved by the engine from the answer at the loop question's own ID: it parses the string as an integer. If the answer is missing or non-numeric, the loop runs zero times.

`Body` questions use `Next{End: true}` to signal the end of one iteration (not the end of the flow).

The answer stored at `FlowContext.Answers["apps_count"]` is `[]map[string]interface{}` — one element per iteration.

**Important:** Do not walk `Continue` in the form pre-walker. `Continue` is deferred to a post-loop sub-form by `HuhFormRunner`. Walking it would cause `confirmGenerate` to appear before the per-app body questions.

### IfQuestion

A conditional branch that does not ask the user anything. The engine evaluates `Condition(ctx)` and follows `Then` or `Else`.

```go
&flow.IfQuestion{
    QuestionBase: flow.QuestionBase{ID_: "needs_biome"},
    Condition: func(ctx *flow.FlowContext) bool {
        fw, _ := ctx.Answers["css_framework"].(string)
        return fw == "tailwind"
    },
    Then: &flow.Next{Question: biomeQuestion},
    Else: &flow.Next{Question: confirmGenerate},
}
```

`IfQuestion` has no label or default because it is never shown to the user.

---

## Wiring questions together (Next)

The `*flow.Next` struct is the edge between questions:

```go
type Next struct {
    Question Question // destination question; nil when End is true
    End      bool     // true = stop the flow
}
```

Build the graph bottom-up — declare the last question first so that earlier questions can reference it:

```go
// 3. Confirmation (terminal)
confirm := &flow.ConfirmQuestion{
    QuestionBase: flow.QuestionBase{ID_: "confirm_generate"},
    Label:        "Generate now?",
    Then:         &flow.Next{End: true},
    Else:         &flow.Next{End: true},
}

// 2. Middle question
description := &flow.TextQuestion{
    QuestionBase: flow.QuestionBase{
        ID_:   "description",
        Next_: &flow.Next{Question: confirm},
    },
    Label: "Project description",
}

// 1. Root (first question asked)
projectName := &flow.TextQuestion{
    QuestionBase: flow.QuestionBase{
        ID_:   "project_name",
        Next_: &flow.Next{Question: description},
    },
    Label: "Project name",
}
```

---

## Branching

Branches are created by pointing `Then`/`Else` (for `ConfirmQuestion`) or `Option.Next` (for `OptionQuestion`) to different questions. Branches can converge back to a shared question.

```go
confirmGenerate := &flow.ConfirmQuestion{
    QuestionBase: flow.QuestionBase{ID_: "confirm_generate"},
    Label: "Generate?",
    Then:  &flow.Next{End: true},
    Else:  &flow.Next{End: true},
}

dockerPort := &flow.TextQuestion{
    QuestionBase: flow.QuestionBase{
        ID_:   "docker_port",
        Next_: &flow.Next{Question: confirmGenerate}, // converge
    },
    Label: "Docker port",
}

includeDocker := &flow.ConfirmQuestion{
    QuestionBase: flow.QuestionBase{ID_: "include_docker"},
    Label:        "Include Docker?",
    Then:         &flow.Next{Question: dockerPort},    // branch
    Else:         &flow.Next{Question: confirmGenerate}, // skip Docker
}
```

The Huh form runner pre-walks the entire graph and hides branches that are not currently reachable given live answers. This is what makes back-navigation work — all fields are rendered once and revealed/hidden dynamically.

---

## Loops

Loop answers are stored under the loop question's ID. Depending on how they were collected (interactive TUI vs. JSON fixture replay), the raw type may be `[]map[string]interface{}` or `[]interface{}`. Always normalise before iterating:

```go
// Normalise helper — handle both native and JSON-unmarshalled shapes.
func extractIterations(raw interface{}) []map[string]interface{} {
    switch v := raw.(type) {
    case []map[string]interface{}:
        return v
    case []interface{}:
        out := make([]map[string]interface{}, 0, len(v))
        for _, item := range v {
            if m, ok := item.(map[string]interface{}); ok {
                out = append(out, m)
            }
        }
        return out
    }
    return nil
}

// In your Generators resolver:
func myGenerators(s *spec.ProjectSpec) []flows.Invocation {
    invs := []flows.Invocation{{Name: "base_project"}}

    iterations := extractIterations(s.Answers["apps_count"])
    for i, iter := range iterations {
        invs = append(invs, flows.Invocation{
            Name: "typescript_base",
            LoopStack: []flow.LoopFrame{
                {QuestionID: "apps_count", Index: i, Answers: iter},
            },
        })
    }
    return invs
}
```

In the generator, loop frames are exposed through the scoped `ctx.Answers`. The executor merges global answers with the loop-frame answers, so the generator reads them identically regardless of whether it is inside a loop:

```go
func (g *AppGenerator) Generate(ctx *dotapi.Context) error {
    appName, _ := ctx.Answers["app-name"].(string)
    // ctx.State is already scoped to apps/<appName>/ — write relative paths only.
    return render.NewLocalFolderRenderer(ctx.State).Render(fs, ctx.Answers)
}
```

Each loop iteration becomes one `generator.Invocation` with a `LoopStack` entry. The executor scopes `ctx.State` via `VirtualProjectState.WithPrefix("apps/<app-name>")` so generators never need to construct app-prefixed paths themselves. See [authoring-generators.md — Loop generators](authoring-generators.md#loop-generators) for the full picture.

---

## Conditional nodes (IfQuestion)

`IfQuestion` is useful when you want to skip a section based on a combination of earlier answers that the user cannot directly express with a single yes/no question:

```go
&flow.IfQuestion{
    QuestionBase: flow.QuestionBase{ID_: "check_needs_lint"},
    Condition: func(ctx *flow.FlowContext) bool {
        lang, _ := ctx.Answers["language"].(string)
        framework, _ := ctx.Answers["framework"].(string)
        return lang == "typescript" && framework != "none"
    },
    Then: &flow.Next{Question: lintConfigQuestion},
    Else: &flow.Next{Question: confirmGenerate},
}
```

The engine calls `Condition` with the `FlowContext` built up to that point. The condition has access to all answers collected so far.

---

## The Generators resolver

`FlowDef.Generators` is a function that maps the populated `ProjectSpec` to an ordered list of generator invocations:

```go
func myGenerators(s *spec.ProjectSpec) []flows.Invocation {
    invs := []flows.Invocation{
        {Name: "base_project"},
    }
    if lang, _ := s.Answers["language"].(string); lang == "typescript" {
        invs = append(invs, flows.Invocation{Name: "typescript_base"})
    }
    return invs
}
```

The resolver does not need to be exhaustive about transitive dependencies — `DependsOn` in each manifest handles that. Focus on the **explicit** generators that reflect the user's choices.

### Loop invocations

For loop-based flows, emit one `Invocation` per loop frame. Use the normalisation helper (see [Loops](#loops)) to handle both interactive and scripted-runner shapes:

```go
func resolveAppGenerators(answers map[string]interface{}, loopStack []flow.LoopFrame) []flows.Invocation {
    var invs []flows.Invocation
    // Read per-iteration answers (falls back to global answers when not in a loop).
    stack, _ := answers["stack"].(string)
    if stack == "typescript" {
        invs = append(invs, flows.Invocation{Name: "typescript_base", LoopStack: loopStack})
    }
    return invs
}

func multiAppGenerators(s *spec.ProjectSpec) []flows.Invocation {
    invs := []flows.Invocation{{Name: "base_project"}, {Name: "monorepo_ts_workspaces"}}

    for i, iter := range extractIterations(s.Answers["apps_count"]) {
        loopStack := []flow.LoopFrame{{QuestionID: "apps_count", Index: i, Answers: iter}}
        invs = append(invs, resolveAppGenerators(iter, loopStack)...)
    }
    return invs
}
```

`LoopFrame.Answers` is merged with global answers by the executor before the generator runs. The executor also calls `VirtualProjectState.WithPrefix("apps/<app-name>")` so each generator invocation writes into the correct app directory automatically.

---

## Registering a flow

Add your flow to `flows/registry.go`:

```go
func Default() *Registry {
    r := NewRegistry()
    _ = r.Register(InitFlow())
    _ = r.Register(FullstackFlow())
    _ = r.Register(MicroservicesFlow())
    _ = r.Register(PluginTemplateFlow())
    _ = r.Register(MyNewFlow())   // ← add here
    return r
}
```

The flow immediately appears in `dot flows` and `dot scaffold`.

---

## Built-in flows

| ID | File | Key questions |
|----|------|---------------|
| `monorepo` | `flows/monorepo.go` | `project_name`, `package_manager`, `include_biome` |
| `fullstack` | `flows/fullstack.go` | `project_name`, `api_language`, `frontend_framework` |
| `microservices` | `flows/microservices.go` | `project_name`, `services` (loop: `service_name`, `service_port`) |
| `plugin-template` | `flows/plugin_template.go` | `project_name`, `module_path`, `plugin_description`, `plugin_author`, `plugin_year`, `plugin_include_injection`, `plugin_include_generator` |

---

## Testing a flow

Every flow should have a fixture in `tools/test-flow/testdata/`. See [test-flow.md](test-flow.md) for the full guide. At minimum, create one fixture that exercises the happy path:

```json
{
  "name": "my_flow_full",
  "flow_id": "my-flow",
  "answers": {
    "project_name": "test-project",
    "language": "typescript",
    "confirm_generate": true
  },
  "expected_visited": [
    "project_name",
    "language",
    "confirm_generate"
  ],
  "skip_post_commands": true,
  "skip_test_commands": false
}
```

Run it:

```bash
go run ./tools/test-flow -only my_flow_full
```
