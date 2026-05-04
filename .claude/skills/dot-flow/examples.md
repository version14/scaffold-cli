# dot-flow examples

One fully-worked example per mode. Use these as the source of truth for diffs the skill produces.

---

## Example 1 — `init`

User invokes the skill. After Step 1 (`init` selected) and Step 2 the answers are:

```
flow_id          = my-stack
title            = My Stack
description      = Scaffolds a barebones My Stack project.
root_question_id = project_name
```

### File created: `flows/my_stack.go`

```go
// Package flows registers the My Stack flow.
package flows

import (
	"github.com/version14/dot/internal/flow"
	"github.com/version14/dot/internal/spec"
)

// MyStackFlow scaffolds a barebones My Stack project.
func MyStackFlow() *FlowDef {
	root := &flow.TextQuestion{
		QuestionBase: flow.QuestionBase{
			ID_:   "project_name",
			Next_: &flow.Next{End: true},
		},
		Label:    "Project name",
		Validate: nonEmpty,
	}

	return &FlowDef{
		ID:          "my-stack",
		Title:       "My Stack",
		Description: "Scaffolds a barebones My Stack project.",
		Root:        root,
		Generators:  resolveMyStackFlowGenerators,
	}
}

func resolveMyStackFlowGenerators(_ *spec.ProjectSpec) []Invocation {
	// TODO: emit one Invocation per generator the flow should run.
	return []Invocation{}
}
```

### File modified: `flows/registry.go`

```go
func Default() *Registry {
	r := NewRegistry()
	_ = r.Register(InitFlow())
	_ = r.Register(MyStackFlow())
	_ = r.Register(PluginTemplateFlow())
	return r
}
```

### File created: `docs/contributor/flows/my-stack.md`

Copy of `_template.md` with the Identity table prefilled:

```markdown
# Flow: `my-stack`

Scaffolds a barebones My Stack project.

## Identity

| Field | Value |
|-------|-------|
| ID | `my-stack` |
| Title | My Stack |
| File | `flows/my_stack.go` |
| Root question | `project_name` |
```

(Other `<!-- TODO -->` blocks remain as placeholders for the contributor to fill.)

### File created: `tools/test-flow/testdata/my_stack_full.json`

```json
{
  "name": "my_stack_full",
  "flow_id": "my-stack",
  "answers": {
    "project_name": "test-value"
  },
  "expected_visited": [
    "project_name"
  ],
  "skip_post_commands": true,
  "skip_test_commands": true
}
```

### Files modified (one row each)

- `docs/contributor/authoring-flows.md` — add `| my-stack | flows/my_stack.go | project_name |` to the "Built-in flows" table.
- `docs/README.md` — add a link to `docs/contributor/flows/my-stack.md` under the flows index.

---

## Example 2 — `generate`

Step 2 answers:

```
flow_id          = next-stack
title            = Next.js Stack
description      = Next.js app with optional Prisma and Tailwind.
root_question_id = project_name
scaffold_summary = Next.js app with optional Prisma and Tailwind
generators       = [
  { name: "base_project",    status: "existing" },
  { name: "next_app",        status: "new" },
  { name: "prisma_db",       status: "new" }
]
```

### Sub-step: invoke `dot-generator` for each `new` entry

The skill spawns the `dot-generator` skill twice — once for `next_app`, once for `prisma_db` — and waits. Both must succeed before continuing.

### File created: `flows/next_stack.go`

```go
// Package flows registers the Next.js Stack flow.
package flows

import (
	"github.com/version14/dot/internal/flow"
	"github.com/version14/dot/internal/spec"
)

func NextStackFlow() *FlowDef {
	includePrisma := &flow.ConfirmQuestion{
		QuestionBase: flow.QuestionBase{ID_: "include_prisma"},
		Label:        "Include Prisma ORM?",
		Default:      false,
		Then:         &flow.Next{End: true},
		Else:         &flow.Next{End: true},
	}

	includeTailwind := &flow.ConfirmQuestion{
		QuestionBase: flow.QuestionBase{ID_: "include_tailwind"},
		Label:        "Include Tailwind?",
		Default:      true,
		Then:         &flow.Next{Question: includePrisma},
		Else:         &flow.Next{Question: includePrisma},
	}

	projectName := &flow.TextQuestion{
		QuestionBase: flow.QuestionBase{
			ID_:   "project_name",
			Next_: &flow.Next{Question: includeTailwind},
		},
		Label:    "Project name",
		Validate: nonEmpty,
	}

	return &FlowDef{
		ID:          "next-stack",
		Title:       "Next.js Stack",
		Description: "Next.js app with optional Prisma and Tailwind.",
		Root:        projectName,
		Generators:  resolveNextStackGenerators,
	}
}

func resolveNextStackGenerators(s *spec.ProjectSpec) []Invocation {
	invs := []Invocation{
		{Name: "base_project"},
		{Name: "next_app"},
	}
	if v, _ := s.Answers["include_prisma"].(bool); v {
		invs = append(invs, Invocation{Name: "prisma_db"})
	}
	return invs
}
```

### File created: `tools/test-flow/testdata/next_stack_full.json`

```json
{
  "name": "next_stack_full",
  "flow_id": "next-stack",
  "answers": {
    "project_name": "demo",
    "include_tailwind": true,
    "include_prisma": true
  },
  "expected_visited": [
    "project_name",
    "include_tailwind",
    "include_prisma"
  ],
  "skip_post_commands": true,
  "skip_test_commands": true
}
```

### Same registry / doc / index updates as Example 1

- `flows/registry.go` — add `_ = r.Register(NextStackFlow())`.
- `docs/contributor/flows/next-stack.md` — copy from `_template.md`, prefill Identity.
- `docs/contributor/authoring-flows.md` and `docs/README.md` — add the row / link.

### Final summary the skill prints

```
Created flow: flows/next_stack.go
  Generators invoked: base_project, next_app, prisma_db (prisma_db conditional on include_prisma)
  Doc: docs/contributor/flows/next-stack.md
  Fixture: tools/test-flow/testdata/next_stack_full.json
  Registered in: flows/registry.go
  make test: PASS
  make test-flows: PASS
```
