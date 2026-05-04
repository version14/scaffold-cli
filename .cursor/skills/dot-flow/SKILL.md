<!-- Mirror of .claude/skills/dot-flow/SKILL.md — update both files together -->
---
name: dot-flow
description: Create or AI-generate a dot scaffolding flow (question graph plus generator resolver). Always asks the user a structured set of questions first, then generates the flow file, registers it, writes the doc, and adds a fixture. Use when the user wants to add a flow, scaffold a flow from a description, or otherwise work in flows/.
---

# dot-flow

Create a new dot flow end-to-end: file in `[flows/](../../../flows)`, registration in `[flows/registry.go](../../../flows/registry.go)`, doc in `[docs/contributor/flows/](../../../docs/contributor/flows)`, and a happy-path fixture in `[tools/test-flow/testdata/](../../../tools/test-flow/testdata)`.

This skill never writes anything before completing the structured question pass below. No silent defaults.

## Mode-independent rules

- Build the question graph bottom-up (declare terminal questions first; root last).
- Question IDs MUST be unique within a flow.
- Loop body questions use `&flow.Next{End: true}` to end the iteration, NOT the flow.
- Read but do NOT duplicate `[docs/contributor/authoring-flows.md](../../../docs/contributor/authoring-flows.md)`.
- Mirror the style of `[flows/plugin_template.go](../../../flows/plugin_template.go)`.

## Step 1 — Mode prompt (ALWAYS first)

Ask via `AskQuestion` (single-select):

```
init     — scaffold a minimal flow file with TODOs
generate — full AI generation from a one-line description
edit     — modify an existing flow
```

## Step 2 — Structured question pass (MANDATORY, runs BEFORE any write)

Issue a single batched `AskQuestion` call covering ALL of these. If a value fails validation, re-ask only the failing question.

Common (both modes):

1. `flow_id` — kebab-case, must NOT collide with any `FlowDef.ID` already registered in `[flows/registry.go](../../../flows/registry.go)` (read the file first).
2. `title` — human-readable, shown in `dot flows`.
3. `description` — one line, shown in `dot flows`.
4. `root_question_id` — snake_case, used as the first question's `ID_`.

Generate-mode only:

5. `scaffold_summary` — one-line description of what the flow scaffolds (e.g. "Next.js app with optional Prisma and Tailwind").
6. `generators` — list of generator names the resolver should invoke. For each, mark `existing` (already in `[generators/](../../../generators)`) or `new`. New ones MUST be created via the `dot-generator` skill BEFORE this skill writes the resolver.

Validation rules (apply BEFORE writing):

- `flow_id`: matches `^[a-z][a-z0-9-]*$`, unique in `[flows/registry.go](../../../flows/registry.go)`.
- `root_question_id`: matches `^[a-z][a-z0-9_]*$`.
- `generators[*].name`: each `existing` entry must have a matching directory under `[generators/](../../../generators)`.

## Step 3a — `init` workflow

Runs only AFTER Step 2 succeeds.

1. Write `flows/<flow_id_underscored>.go` (file name uses underscores, ID stays kebab-case). Use the stub template below. Then edit `[flows/init.go](../../../flows/init.go)` to add a branch point to this new flow.
2. Register in `[flows/registry.go](../../../flows/registry.go)` `Default()`. Insert next to the existing `_ = r.Register(...)` calls, alphabetically by flow ID.
3. Copy `[docs/contributor/flows/_template.md](../../../docs/contributor/flows/_template.md)` to `docs/contributor/flows/<flow_id>.md`. Prefill the Identity table (ID, Title, File, Root question), and replace every other `<!-- TODO -->` with a placeholder pointing back at the generated `flows/<flow_id_underscored>.go`.
4. Add a row to the "Built-in flows" table in `[docs/contributor/authoring-flows.md](../../../docs/contributor/authoring-flows.md)` and to the flows index in `[docs/README.md](../../../docs/README.md)`.
5. Create `tools/test-flow/testdata/<flow_id_underscored>_full.json` with `skip_post_commands: true` and `skip_test_commands: true`. Answers cover every question in the stub. `expected_visited` lists every question id in the stub.
6. Run `make test` and `make test-flows` and report the result. If either fails, print the failing output and suggest the user run the failing command with `2>&1 | head -80` to investigate. Do NOT auto-fix. Stop and wait for user input.

## Step 3b — `generate` workflow

Runs only AFTER Step 2 succeeds.

1. For every generator in `generators[*]` that is marked `new`, execute the full `dot-generator` skill workflow inline (Steps 1–3c of `dot-generator/SKILL.md`) before continuing. Do not proceed to the next generator until the current one is created and `go build ./...` passes. Only continue to step 2 below once ALL generators are successfully created.
2. Produce `flows/<flow_id_underscored>.go` end-to-end:
   - Bottom-up question graph using the patterns documented in `[docs/contributor/authoring-flows.md](../../../docs/contributor/authoring-flows.md)` (`TextQuestion`, `ConfirmQuestion`, `OptionQuestion`, `LoopQuestion`, `IfQuestion`).
   - Validators on text inputs: reuse `nonEmpty`, `validateModulePath`, `validatePluginID` from `[flows/plugin_template.go](../../../flows/plugin_template.go)` when applicable; otherwise add new validators in the same file.
   - `Generators` resolver returns `[]Invocation` derived from the user's answers — one entry per declared generator, with loop frames where the graph has a `LoopQuestion`.
3. Run steps 2–6 of the `init` workflow (register, doc row, fixture, `make test`).
4. Print a summary block: file path, generators invoked, fixture path, registry entry.

## Step 4 — `edit` workflow

Runs only when the user selects `edit` in Step 1.

1. Confirm the user intends to edit `[flows/init.go](../../../flows/init.go)`. No other flow file is editable through this skill — standalone flows are append-only here.
2. Read `flows/init.go` plus its doc plus its fixtures.
3. Ask, via `AskQuestion`, the change kind: `add question` / `remove question` / `rename id` / `change branch` / `change resolver`.
4. Apply the minimal diff. Update the matching row in `docs/contributor/flows/<flow_id>.md` (Questions table + ASCII graph). Update fixtures if any answer key changed. Run `make test` and `make test-flows`.

## Stub flow template

Use this as the file body for `init` (replace the placeholders the structured question pass collected):

```go
// Package flows registers the <FLOW_TITLE> flow.
package flows

import (
	"github.com/version14/dot/internal/flow"
	"github.com/version14/dot/internal/spec"
)

// <FlowFuncName> scaffolds <SCAFFOLD_SUMMARY_OR_TODO>.
func <FlowFuncName>() *FlowDef {
	root := &flow.TextQuestion{
		QuestionBase: flow.QuestionBase{
			ID_:   "<root_question_id>",
			Next_: &flow.Next{End: true},
		},
		Label:    "<ROOT_QUESTION_LABEL>",
		Validate: nonEmpty,
	}

	return &FlowDef{
		ID:          "<flow_id>",
		Title:       "<TITLE>",
		Description: "<DESCRIPTION>",
		Root:        root,
		Generators:  resolve<FlowFuncName>Generators,
	}
}

func resolve<FlowFuncName>Generators(_ *spec.ProjectSpec) []Invocation {
	// TODO: emit one Invocation per generator the flow should run.
	return []Invocation{}
}
```

`FlowFuncName` is `<flow_id>` upper-camel-cased + `Flow` (e.g. `my-stack` -> `MyStackFlow`). The file name uses underscores (`my_stack.go`).

## Stub fixture template

```json
{
  "name": "<flow_id_underscored>_full",
  "flow_id": "<flow_id>",
  "answers": {
    "<root_question_id>": "test-value"
  },
  "expected_visited": [
    "<root_question_id>"
  ],
  "skip_post_commands": true,
  "skip_test_commands": true
}
```

## After completion

Report:

- Created files (`flows/<file>.go`, `docs/contributor/flows/<id>.md`, `tools/test-flow/testdata/<id>_full.json`).
- Modified files (`flows/registry.go`, `flows/init.go`, `docs/contributor/authoring-flows.md`, `docs/README.md`).
- `make test` and `make test-flows` outcome.
- Any new generators that were created via the `dot-generator` skill (generate mode).

For end-to-end worked examples, see [examples.md](examples.md).
