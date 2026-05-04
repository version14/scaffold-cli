# Generator: `python_fastapi_auth`

Adds FastAPI auth routes `/register` and `/login` that always return 200.

---

## Identity

| Field | Value |
|-------|-------|
| Name | `python_fastapi_auth` |
| Version | `0.1.0` |
| Package | `generators/python_fastapi_auth` |

---

## Dependencies

| Generator | Why |
|-----------|-----|
| `python_fastapi_base` | Provides the FastAPI app that mounts the auth router |

---

## Answers consumed

None.

---

## Files written

| Path | Description |
|------|-------------|
| `routers/auth.py` | FastAPI router with `POST /auth/register` and `POST /auth/login`, always returning 200 |

---

## Validators

| Check | Type | Passes when |
|-------|------|-------------|
| `routers/auth.py` | `file_exists` | File is present after generation |

---

## Post-generation commands

No PostGenerationCommands.

## Test commands

No TestCommands.

---

## Conflicts

None.
