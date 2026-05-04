# Authoring Skills

A **skill** is a structured prompt file that teaches an AI assistant (Claude or Cursor) how to perform a specific, multi-step task in this repository. Skills enforce a question pass before any write, carry inline templates, and produce consistent, reviewable output.

This guide covers: what a skill is, how to create one, how to modify one, and how the `.claude/` ↔ `.cursor/` sync works.

---

## Table of Contents

- [What is a skill?](#what-is-a-skill)
- [Directory layout](#directory-layout)
- [Skill file structure](#skill-file-structure)
- [Creating a new skill](#creating-a-new-skill)
- [Modifying an existing skill](#modifying-an-existing-skill)
- [The .claude / .cursor sync](#the-claude--cursor-sync)
- [Wiring a skill into CLAUDE.md](#wiring-a-skill-into-claudemd)
- [Built-in skills](#built-in-skills)
- [Design rules](#design-rules)

---

## What is a skill?

A skill is a Markdown file that an AI assistant reads before executing a task. It defines:

- A **mode selection step** — what kind of action the user wants.
- A **structured question pass** — everything the AI must ask the user before touching any file.
- **Validation rules** — what makes an answer acceptable.
- **Step-by-step workflows** — one per mode, with inline stub templates.
- An **after-completion report** — what the AI must print when done.

Skills are not code. They do not run automatically. They are loaded into an AI session when the user invokes the skill by name (e.g. `/dot-flow`) or when CLAUDE.md routing fires.

---

## Directory layout

Each skill lives under `.claude/skills/<skill-name>/` and is mirrored under `.cursor/skills/<skill-name>/`:

```
.claude/skills/
└── <skill-name>/
    ├── SKILL.md       ← the executable spec (loaded into context)
    └── examples.md    ← one fully-worked example per mode / strategy

.cursor/skills/
└── <skill-name>/
    ├── SKILL.md       ← mirror (identical content + mirror comment at top)
    └── examples.md    ← mirror
```

The `.claude/` tree is the source of truth. The `.cursor/` tree is a mirror — never edit it directly. Use the `sync-skills` skill to propagate changes.

---

## Skill file structure

### `SKILL.md` frontmatter

Every SKILL.md starts with a YAML frontmatter block:

```markdown
---
name: <skill-name>
description: <one-line description — used by the AI to decide when to invoke the skill>
---
```

The `description` must be specific enough that the AI can route to the skill without reading the full file.

### Required sections

| Section | Purpose |
|---------|---------|
| `## Mode-independent rules` | Invariants that apply regardless of mode. List them as bullets. |
| `## Step 1 — Mode prompt` | Always first. Single-select `AskQuestion`. List modes with one-line descriptions. |
| `## Step 2 — Structured question pass` | All questions, batched in a single `AskQuestion` call. Numbered list, one question per line. |
| `## Validation rules` | Regex patterns, uniqueness checks, cross-file lookups. Applied BEFORE any write. |
| `## Step 3a / 3b / 3c — <mode> workflow` | One section per mode. Numbered steps. Include inline stub templates for `init` mode. |
| `## After completion` | Bullet list of what the AI must report: created files, modified files, command outcomes. |

### `examples.md`

One fully-worked example per mode (and per writing strategy if the skill has multiple). Each example shows:

- The exact answers given in Step 2.
- Every file created or modified, with full content.
- The final summary the AI prints.

`examples.md` is the reference for what correct output looks like. It does NOT replace step-by-step instructions in `SKILL.md` — `SKILL.md` must be self-sufficient.

---

## Creating a new skill

Follow these steps. Each step references a file to read or a constraint to check.

### 1. Choose a name

Pick a `kebab-case` name that describes the task, not the domain (`sync-skills`, not `cursor-sync`). Verify it does not collide with any existing directory under `.claude/skills/`.

### 2. Create the directory and files

```
.claude/skills/<skill-name>/
├── SKILL.md
└── examples.md
```

### 3. Write `SKILL.md`

Start from this skeleton and fill in every section:

```markdown
---
name: <skill-name>
description: <one-line — specific enough to drive routing from CLAUDE.md>
---

# <skill-name>

<One-paragraph summary of what this skill does end-to-end.>

This skill never writes anything before completing the structured question pass below. No silent defaults.

## Mode-independent rules

- <invariant 1>
- <invariant 2>

## Step 1 — Mode prompt (ALWAYS first)

Ask via `AskQuestion` (single-select):

\```
init     — scaffold a minimal <thing> with TODOs
edit     — modify an existing <thing>
generate — full AI generation from a one-line description
\```

## Step 2 — Structured question pass (MANDATORY, runs BEFORE any write)

Issue a single batched `AskQuestion` call covering ALL fields below. If a value fails
validation, re-ask only the failing question.

Common to `init` and `generate`:

1. `<field>` — <type>, <constraint>.
2. ...

`edit` mode only:

N. `target` — pick the item to edit.
N+1. `change_kind` — single-select: ...

Validation rules (apply BEFORE writing):

- `<field>`: matches `<regex>`, unique in `<file>`.
- ...

## Step 3a — `init` workflow

Runs only AFTER Step 2 succeeds.

1. Create `<path>` using the stub template below.
2. ...
N. Run `<build command>` and report. If it fails, print the output and suggest the user
   run `<command> 2>&1 | head -80`. Do NOT auto-fix. Stop and wait for user input.

### Stub templates for `init`

\```go
// ... exact file content with <PLACEHOLDER> values ...
\```

## Step 3b — `edit` workflow

...

## Step 3c — `generate` workflow

...

## After completion

Report:

- Created files (...).
- Modified files (...).
- `<command>` outcome.

For end-to-end worked examples, see [examples.md](examples.md).
```

### 4. Write `examples.md`

Write one complete example per mode. Show:
- The exact Step 2 answers.
- Every file created or modified (full content, not excerpts).
- The summary the AI prints at the end.

### 5. Wire into CLAUDE.md

Add a routing rule to `CLAUDE.md` (see [Wiring a skill into CLAUDE.md](#wiring-a-skill-into-claudemd)).

### 6. Sync to Cursor

Run the `sync-skills` skill to mirror the new files under `.cursor/skills/<skill-name>/`.

---

## Modifying an existing skill

Always edit `.claude/skills/<skill-name>/SKILL.md` (or `examples.md`). Never edit the `.cursor/` mirror directly.

After editing:

1. Re-read the modified SKILL.md and verify:
   - Every step is still numbered sequentially.
   - Validation rules match the question field names.
   - Stub templates compile (if Go) or parse (if JSON/YAML).
   - `After completion` lists all files the updated workflow touches.
2. Update `examples.md` if the change affects the expected output.
3. Run the `sync-skills` skill to propagate the change to `.cursor/`.

---

## The .claude / .cursor sync

### Why two directories?

Claude Code reads from `.claude/skills/`. Cursor reads from `.cursor/skills/`. Both directories must contain the same skill logic.

### Mirror comment

Every file under `.cursor/skills/` starts with:

```markdown
<!-- Mirror of .claude/skills/<skill-name>/<file> — update both files together -->
```

This comment is the only difference between the two copies. The `sync-skills` skill adds it automatically.

### When to sync

| Trigger | Action |
|---------|--------|
| New skill created in `.claude/skills/` | Run `sync-skills` |
| Existing skill modified in `.claude/skills/` | Run `sync-skills` |
| Skill deleted from `.claude/skills/` | Run `sync-skills` (it removes the mirror) |

**Never** edit `.cursor/skills/` manually. If you notice a drift, run `sync-skills` — it overwrites the mirror.

### How sync-skills works

The `sync-skills` skill:

1. Lists all files under `.claude/skills/`.
2. For each file, computes the corresponding path under `.cursor/skills/`.
3. Creates or overwrites the Cursor file with the Claude content, prepending the mirror comment.
4. Detects files present in `.cursor/skills/` but absent from `.claude/skills/` and removes them.
5. Reports a diff summary (files created / updated / deleted).

---

## Wiring a skill into CLAUDE.md

Add a line to the `Key routing rules` section in `CLAUDE.md`:

```markdown
- <trigger phrases> → invoke <skill-name>
```

Trigger phrases should be natural-language expressions a contributor would actually type. Use 3–5 phrases separated by commas. Be specific enough to avoid false positives.

Examples:

```markdown
- Add a flow, scaffold a flow, generate a flow, work in flows/ → invoke dot-flow
- Add a generator, edit a generator, scaffold a generator, work in generators/ → invoke dot-generator
- Sync skills, update cursor skills, propagate skill changes → invoke sync-skills
```

---

## Built-in skills

| Skill | Triggers | What it does |
|-------|---------|-------------|
| [`dot-flow`](.claude/skills/dot-flow/SKILL.md) | Add / scaffold / generate a flow | Creates a flow file, registers it, writes the doc and fixture end-to-end |
| [`dot-generator`](.claude/skills/dot-generator/SKILL.md) | Add / edit / generate a generator | Creates or modifies a generator package, registers it, writes the doc |
| [`sync-skills`](.claude/skills/sync-skills/SKILL.md) | Sync skills, update cursor skills | Mirrors `.claude/skills/` to `.cursor/skills/` |

---

## Design rules

Follow these when writing or reviewing skills.

**SKILL.md must be self-sufficient.** The AI must be able to execute the skill using only SKILL.md. `examples.md` adds worked examples but is not required reading to execute the skill.

**No write before question pass.** Every skill must complete the structured question pass (Step 2) in full before creating or modifying any file.

**Inline stubs for `init` mode.** The exact file content (with `<PLACEHOLDER>` markers) must appear in SKILL.md, not just in `examples.md`.

**Validation before every write.** Regex checks, uniqueness checks, and cross-file lookups all happen in Step 2. If any validation fails, re-ask only the failing question.

**Fail loudly, never silently auto-fix.** When a build or test step fails, print the output, suggest the debug command (`<cmd> 2>&1 | head -80`), and stop. Do not attempt a fix automatically.

**Mirror comment stays.** The `<!-- Mirror of ... -->` comment at the top of `.cursor/` files must never be removed. `sync-skills` relies on it to detect manually edited mirror files.

**One skill = one domain.** Don't grow a skill into a general-purpose tool. If a skill starts handling unrelated tasks, split it.
