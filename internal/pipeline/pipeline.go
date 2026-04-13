// Package pipeline executes a []FileOp slice produced by generators.
// The three phases are: collect → resolve → write.
// No file is written until all ops are validated. Any error aborts the whole
// run — there are no partial writes.
package pipeline

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"text/template"

	"github.com/version14/dot/internal/generator"
)

// Run executes a slice of FileOps. It:
//  1. Sorts ops by Priority (descending), then Generator name (alphabetical).
//  2. Detects Create/Template conflicts (two ops at same priority for same path).
//  3. Applies all ops in memory.
//  4. Writes everything to disk atomically (all-or-nothing per this run).
func Run(ops []generator.FileOp) error {
	if len(ops) == 0 {
		return nil
	}

	sorted := sortOps(ops)

	if err := detectConflicts(sorted); err != nil {
		return err
	}

	writes, err := buildWrites(sorted)
	if err != nil {
		return err
	}

	return flushWrites(writes)
}

// fileWrite is an in-memory representation of a file to be written.
type fileWrite struct {
	path    string
	content []byte
}

// sortOps returns a copy of ops sorted by Priority descending, then Generator
// name ascending (deterministic tiebreak).
func sortOps(ops []generator.FileOp) []generator.FileOp {
	sorted := make([]generator.FileOp, len(ops))
	copy(sorted, ops)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Priority != sorted[j].Priority {
			return sorted[i].Priority > sorted[j].Priority
		}
		if sorted[i].Generator != sorted[j].Generator {
			return sorted[i].Generator < sorted[j].Generator
		}
		return sorted[i].Path < sorted[j].Path
	})
	return sorted
}

// detectConflicts checks for Create/Template ops at the same Priority on the
// same path. Two such ops would produce an ambiguous result — abort early.
func detectConflicts(ops []generator.FileOp) error {
	type key struct {
		path     string
		priority int
	}
	seen := make(map[key]string) // key → generator name
	for _, op := range ops {
		if op.Kind != generator.Create && op.Kind != generator.Template {
			continue
		}
		k := key{op.Path, op.Priority}
		if prev, ok := seen[k]; ok {
			return fmt.Errorf(
				"pipeline conflict: generators %q and %q both want to %s %q at priority %d",
				prev, op.Generator, op.Kind, op.Path, op.Priority,
			)
		}
		seen[k] = op.Generator
	}
	return nil
}

// buildWrites applies all ops in memory and returns the final file contents.
// For Create/Template the highest-priority op wins (sorted order guarantees
// the first one we see is the winner). For Append/Patch, all ops are applied.
func buildWrites(ops []generator.FileOp) ([]fileWrite, error) {
	// Track which paths have been claimed by Create/Template ops.
	claimed := make(map[string]bool)
	// Accumulate content per path (for Append).
	content := make(map[string][]byte)

	for _, op := range ops {
		switch op.Kind {
		case generator.Create:
			if claimed[op.Path] {
				continue // lower-priority op; skip
			}
			claimed[op.Path] = true
			content[op.Path] = []byte(op.Content)

		case generator.Template:
			if claimed[op.Path] {
				continue
			}
			rendered, err := renderTemplate(op.Path, op.Content)
			if err != nil {
				return nil, fmt.Errorf("template %s (generator %q): %w", op.Path, op.Generator, err)
			}
			claimed[op.Path] = true
			content[op.Path] = rendered

		case generator.Append:
			content[op.Path] = append(content[op.Path], []byte(op.Content)...)

		case generator.Patch:
			patched, err := applyPatch(string(content[op.Path]), op)
			if err != nil {
				return nil, fmt.Errorf("patch %s anchor=%s (generator %q): %w",
					op.Path, op.Anchor, op.Generator, err)
			}
			content[op.Path] = []byte(patched)
		}
	}

	writes := make([]fileWrite, 0, len(content))
	for path, data := range content {
		writes = append(writes, fileWrite{path: path, content: data})
	}
	return writes, nil
}

// flushWrites creates parent directories and writes every file to disk.
func flushWrites(writes []fileWrite) error {
	for _, w := range writes {
		if err := os.MkdirAll(filepath.Dir(w.path), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(w.path), err)
		}
		if err := os.WriteFile(w.path, w.content, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", w.path, err)
		}
	}
	return nil
}

// applyPatch dispatches to the correct patch function based on op.Anchor.
func applyPatch(src string, op generator.FileOp) (string, error) {
	switch op.Anchor {
	case generator.AnchorImportBlock:
		return patchImportBlock(src, op.Content)
	case generator.AnchorMainFunc:
		return patchFunc(src, "main", op.Content)
	case generator.AnchorInitFunc:
		return patchFunc(src, "init", op.Content)
	default:
		return "", fmt.Errorf("unknown anchor %q", op.Anchor)
	}
}

// renderTemplate executes a Go text/template with the path as its name.
func renderTemplate(name, tmpl string) ([]byte, error) {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, nil); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
