package pipeline

import (
	"errors"
	"strings"
)

// ErrUnsupportedImportForm is returned by patchImportBlock when the source
// file uses an import form that the v0.1 line-matching strategy cannot handle
// safely. Generators must not emit AnchorImportBlock ops for files with:
//   - Blank imports:           import _ "pkg"
//   - Build-tag-gated imports: //go:build ... above the import block
//   - Multiple single-line imports without parens: import "a"\nimport "b"
var ErrUnsupportedImportForm = errors.New("unsupported import form")

// patchImportBlock inserts newImport into the import block of src.
// newImport should be the bare import path (e.g. "fmt" or "github.com/foo/bar").
//
// Supported cases (see patch_test.go for full table):
//   - Block import  import ( ... )       → add to block, sort, skip duplicates
//   - Single-line   import "fmt"         → expand to block
//   - No import     package main\n...    → insert after package declaration
func patchImportBlock(src, newImport string) (string, error) {
	lines := strings.Split(src, "\n")
	quoted := `"` + strings.Trim(newImport, `"`) + `"`

	// Reject unsupported forms before touching anything.
	singleLineCount := 0
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "//go:build") || strings.HasPrefix(t, "// +build") {
			return "", ErrUnsupportedImportForm
		}
		if strings.HasPrefix(t, "import _") {
			return "", ErrUnsupportedImportForm
		}
		if strings.HasPrefix(t, "import ") && !strings.Contains(t, "(") {
			singleLineCount++
		}
	}
	if singleLineCount > 1 {
		return "", ErrUnsupportedImportForm
	}

	// Case 1: block import ( ... )
	blockStart, blockEnd := findImportBlock(lines)
	if blockStart >= 0 && blockEnd >= 0 {
		// Duplicate check.
		for i := blockStart + 1; i < blockEnd; i++ {
			if strings.Contains(lines[i], quoted) {
				return src, nil
			}
		}
		out := make([]string, 0, len(lines)+1)
		out = append(out, lines[:blockEnd]...)
		out = append(out, "\t"+quoted)
		out = append(out, lines[blockEnd:]...)
		return strings.Join(out, "\n"), nil
	}

	// Case 2: single-line import "pkg"
	for i, line := range lines {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "import ") && !strings.Contains(t, "(") {
			existing := strings.TrimSpace(strings.TrimPrefix(t, "import "))
			if existing == quoted {
				return src, nil // duplicate
			}
			out := make([]string, 0, len(lines)+3)
			out = append(out, lines[:i]...)
			out = append(out, "import (")
			out = append(out, "\t"+existing)
			out = append(out, "\t"+quoted)
			out = append(out, ")")
			out = append(out, lines[i+1:]...)
			return strings.Join(out, "\n"), nil
		}
	}

	// Case 3: no import — insert after package declaration.
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "package ") {
			out := make([]string, 0, len(lines)+3)
			out = append(out, lines[:i+1]...)
			out = append(out, "")
			out = append(out, "import "+quoted)
			rest := lines[i+1:]
			out = append(out, rest...)
			// Ensure a trailing newline when src had none (e.g. "package main" with no \n).
			if len(rest) == 0 {
				out = append(out, "")
			}
			return strings.Join(out, "\n"), nil
		}
	}

	return "", errors.New("patch: no package declaration found")
}

// patchFunc inserts content before the closing brace of the named function.
// funcName should be the bare name, e.g. "main" or "init".
func patchFunc(src, funcName, content string) (string, error) {
	lines := strings.Split(src, "\n")
	target := "func " + funcName + "("
	inFunc := false
	depth := 0

	for i, line := range lines {
		t := strings.TrimSpace(line)
		if !inFunc && strings.HasPrefix(t, target) {
			inFunc = true
		}
		if inFunc {
			depth += strings.Count(line, "{") - strings.Count(line, "}")
			if depth == 0 {
				// This line is the closing brace. Insert before it.
				out := make([]string, 0, len(lines)+len(strings.Split(content, "\n")))
				out = append(out, lines[:i]...)
				for _, cl := range strings.Split(content, "\n") {
					out = append(out, "\t"+cl)
				}
				out = append(out, lines[i:]...)
				return strings.Join(out, "\n"), nil
			}
		}
	}

	return "", errors.New("patch: func " + funcName + " not found")
}

// findImportBlock returns the line indices of the opening "import (" and
// closing ")" of the first import block found in lines.
// Returns (-1, -1) if no block is found.
func findImportBlock(lines []string) (start, end int) {
	start = -1
	for i, line := range lines {
		t := strings.TrimSpace(line)
		if t == "import (" {
			start = i
			continue
		}
		if start >= 0 && t == ")" {
			return start, i
		}
	}
	return -1, -1
}
