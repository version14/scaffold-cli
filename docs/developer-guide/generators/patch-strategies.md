# Patch Strategies

Implementation: `internal/pipeline/patch.go`

Patch ops insert content at a named anchor inside an existing file.
This document defines the supported anchors, their constraints, and the known-unsupported edge cases.

These three anchors (`AnchorImportBlock`, `AnchorMainFunc`, `AnchorInitFunc`) are designed for the most common code injection patterns:
adding imports, startup logic, and initialization hooks. They are intentionally limited to patterns that generators created —
for example, if a generator creates a Go file from scratch, it can safely assume there is exactly one `func main()` with clear boundaries.
New anchors should only be added for patterns that cannot be expressed via `Append` ops (appending to the end of a file).

---

## AnchorImportBlock

Constant: `generator.AnchorImportBlock` (`"import_block"`)

Inserts a new import path into a Go file's import block.

**Supported forms:**

| Input form | Behavior |
|---|---|
| `import ( ... )` block | Appends inside the block before the closing `)` |
| Single `import "pkg"` | Expands to a parenthesized block, then appends |
| No import at all | Inserts `import "pkg"` after the `package` line |
| Duplicate import | Skipped silently (not an error) |

**Unsupported forms (returns `ErrUnsupportedImportForm`):**

| Input form | Why unsupported |
|---|---|
| `import _ "pkg"` blank import | Blank imports have semantic meaning; reordering them is unsafe |
| `//go:build` tag above the import | Build-tag-gated files are structurally special |
| Multiple single-line imports (`import "a"\nimport "b"`) | Ambiguous insertion point |

Generators must only emit `AnchorImportBlock` ops for files where the import form is known to be safe. When the import form is uncertain (e.g. a file the generator did not create), check the content before emitting the op or return `ErrUnsupportedImportForm`.

**Content format:** pass the bare import path, with or without quotes.

```go
// Both of these are equivalent:
Content: `"database/sql"`,
Content: `database/sql`,
```

---

## AnchorMainFunc

Constant: `generator.AnchorMainFunc` (`"main_func"`)

Inserts content before the closing `}` of `func main()`.

The pipeline uses brace-depth tracking: depth increments on `{`, decrements on `}`. When depth reaches `0` after entering `func main(`, that line is the closing brace. Content is inserted with one tab of indentation per line.

**Constraint:** The file must contain exactly one `func main(`. If not found, returns an error.

**Content format:** multi-line content is supported. Each line gets one tab prepended.

```go
generator.FileOp{
    Kind:      generator.Patch,
    Path:      "main.go",
    Anchor:    generator.AnchorMainFunc,
    Generator: g.Name(),
    Content:   "db := setupDB()\ndefer db.Close()",
}
```

Result in `main.go`:
```go
func main() {
    // ... existing code ...
    db := setupDB()
    defer db.Close()
}
```

---

## AnchorInitFunc

Constant: `generator.AnchorInitFunc` (`"init_func"`)

Same as `AnchorMainFunc` but targets `func init()`.

**Constraint:** The file must contain `func init(`. If not found, returns an error.

---

## Adding new anchors

If you need an anchor that does not exist:

1. Add the constant to `internal/generator/fileop.go`
2. Implement the handler function in `internal/pipeline/patch.go`
3. Add the `case` to `applyPatch()` in `patch.go`
4. Write table-driven tests in `internal/pipeline/patch_test.go` covering:
   - All supported input forms
   - All unsupported forms (must return the right error)
   - Duplicate/idempotent cases
   - Edge cases (empty file, minimal file with just `package` declaration)

Tests for new anchors are not optional — the patch logic is subtle and easy to break.
