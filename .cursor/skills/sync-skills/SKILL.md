<!-- Mirror of .claude/skills/sync-skills/SKILL.md — update both files together -->
---
name: sync-skills
description: Mirror .claude/skills/ to .cursor/skills/, adding the mirror comment to each file. Run after creating or modifying any skill. Use when the user says "sync skills", "update cursor skills", or "propagate skill changes".
---

# sync-skills

Keep `.cursor/skills/` in sync with `.claude/skills/`. The `.claude/` tree is the source of truth. This skill overwrites `.cursor/` files; it never reads them as input.

This skill never modifies `.claude/skills/`. It is write-only toward `.cursor/skills/`.

## Mode-independent rules

- `.claude/skills/` is always source. `.cursor/skills/` is always destination.
- Every file written to `.cursor/` gets a mirror comment as its first line (see format below).
- The mirror comment is the only difference between source and destination files.
- If a file already exists in `.cursor/` with different content, overwrite it unconditionally.
- If a directory exists in `.cursor/skills/` but has no counterpart in `.claude/skills/`, remove it and report.

## Mirror comment format

```
<!-- Mirror of .claude/skills/<skill-name>/<filename> — update both files together -->
```

The comment must be on line 1, before any other content including YAML frontmatter.

## Step 1 — Scope prompt (ALWAYS first)

Ask via `AskQuestion` (single-select):

```
all    — sync every skill (full mirror)
one    — sync a single named skill
check  — dry-run: list files that are out of sync without writing anything
```

## Step 2 — Structured question pass (MANDATORY, runs BEFORE any write)

`all` mode: no additional questions. Proceed directly to Step 3a.

`one` mode only:

1. `skill_name` — name of the skill directory to sync (must exist under `.claude/skills/`).

`check` mode: no additional questions. Proceed directly to Step 3c.

Validation rules:

- `skill_name` (`one` mode): must be an existing directory name under `.claude/skills/` — list the directory before asking.

## Step 3a — `all` workflow

Runs only AFTER Step 2 succeeds (or immediately for `all` mode).

1. List every file recursively under `.claude/skills/` (any depth).
2. For each file `F` at path `.claude/skills/<rel-path>`:
   a. Compute destination path: `.cursor/skills/<rel-path>`.
   b. Create parent directories under `.cursor/skills/` if they do not exist.
   c. Read the full content of `F`.
   d. Prepend the mirror comment: `<!-- Mirror of .claude/skills/<rel-path> — update both files together -->` followed by a newline.
   e. Write the result to `.cursor/skills/<rel-path>` (overwrite if exists).
3. List every file recursively under `.cursor/skills/`.
4. For each file `G` at path `.cursor/skills/<rel-path>`:
   - If `.claude/skills/<rel-path>` does NOT exist, delete `G`. If its parent directory is now empty, delete the directory too.
5. Report (see After completion).

## Step 3b — `one` workflow

1. List every file recursively under `.claude/skills/<skill_name>/`.
2. Apply steps 2a–2e from the `all` workflow for each file, scoped to `<skill_name>/`.
3. Do NOT touch other skill directories under `.cursor/skills/`.
4. Report (see After completion).

## Step 3c — `check` workflow (dry-run, no writes)

1. List all files under `.claude/skills/` and `.cursor/skills/` recursively.
2. For each file in `.claude/`:
   - Check whether the corresponding `.cursor/` file exists.
   - If it exists, compare content ignoring the mirror comment line (strip line 1 from the `.cursor/` file before comparing).
   - Classify as: `in-sync`, `missing`, or `out-of-sync`.
3. For each file in `.cursor/` with no `.claude/` counterpart: classify as `orphan`.
4. Print the full classification table. Do NOT write or delete anything.
5. If any file is not `in-sync`, suggest running `sync-skills` in `all` mode.

## After completion

Report:

```
sync-skills: <mode> completed.

Created:  <list of .cursor/ files that were created>
Updated:  <list of .cursor/ files that were overwritten>
Deleted:  <list of .cursor/ files that were removed>
No change: <count> file(s) already in sync
```

If `check` mode: print the classification table instead, no "Created / Updated / Deleted" lines.
