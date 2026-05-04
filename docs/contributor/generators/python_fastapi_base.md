# Generator: `python_fastapi_base`

Scaffolds a base FastAPI Python project with a `/health` endpoint.

---

## Identity

| Field | Value |
|-------|-------|
| Name | `python_fastapi_base` |
| Version | `0.1.0` |
| Package | `generators/python_fastapi_base` |

---

## Dependencies

| Generator | Why |
|-----------|-----|
| `base_project` | Provides README, .gitignore, LICENSE |

---

## Answers consumed

| Key | Type | Required | Notes |
|-----|------|----------|-------|
| `project_name` | string | Yes | Used as the FastAPI app title |

---

## Files written

| Path | Description |
|------|-------------|
| `main.py` | FastAPI app with `/health` route returning `{"status": "ok"}` |
| `requirements.txt` | `fastapi` + `uvicorn[standard]` pinned to recent versions |

---

## Validators

| Check | Type | Passes when |
|-------|------|-------------|
| `main.py` | `file_exists` | File is present after generation |
| `requirements.txt` | `file_exists` | File is present after generation |

---

## Post-generation commands

No PostGenerationCommands.

## Test commands

No TestCommands.

---

## Conflicts

None.
