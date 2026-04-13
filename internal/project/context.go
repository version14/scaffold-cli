// Package project manages the .dot/ directory written to every generated project.
// Named "project" (not "context") to avoid shadowing the stdlib context package.
package project

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/version14/dot/internal/generator"
	"github.com/version14/dot/internal/spec"
)

const dotDir = ".dot"

// Context is written to .dot/config.json after dot init.
// It stores the project's full spec and all available post-creation commands.
// This file must be committed to git — dot new and dot add module require it.
type Context struct {
	DotVersion  string                `json:"dot_version"`        // semver of dot that created this
	SpecVersion int                   `json:"spec_version"`       // schema version; v0.1 = 1
	Spec        spec.Spec             `json:"spec"`               // authoritative project state
	Commands    map[string]CommandRef `json:"available_commands"` // keyed by CommandDef.Name
}

// CommandRef is the persisted form of a CommandDef (without display-only fields).
type CommandRef struct {
	Generator string `json:"generator"` // matches Generator.Name()
	Action    string `json:"action"`    // passed to generator.RunAction()
}

// Manifest is written to .dot/manifest.json after dot init.
// It stores the SHA-256 hash of every generated file at creation time.
// On dot add module, the pipeline compares current hashes to detect
// user modifications and write git-style conflict markers where needed.
type Manifest struct {
	Files map[string]FileRecord `json:"files"`
}

// FileRecord tracks a single generated file.
type FileRecord struct {
	Hash      string `json:"hash"`      // "sha256:<hex>"
	Generator string `json:"generator"` // which generator produced this file
}

// Load reads .dot/config.json by traversing up from startDir to the git root.
// Returns ErrNotDotProject if no .dot/config.json is found.
func Load(startDir string) (*Context, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return nil, fmt.Errorf("load context: %w", err)
	}

	for {
		candidate := filepath.Join(dir, dotDir, "config.json")
		if _, err := os.Stat(candidate); err == nil {
			return readContext(candidate)
		}
		// Check for git root to stop traversal.
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached filesystem root
		}
		dir = parent
	}

	return nil, ErrNotDotProject
}

// ErrNotDotProject is returned when no .dot/config.json is found in the
// directory tree up to the nearest git root.
var ErrNotDotProject = fmt.Errorf("not a dot project (no .dot/config.json found)")

func readContext(path string) (*Context, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read context: %w", err)
	}
	var ctx Context
	if err := json.Unmarshal(data, &ctx); err != nil {
		return nil, fmt.Errorf("parse context: %w", err)
	}
	return &ctx, nil
}

// Save writes the Context to <root>/.dot/config.json and the Manifest to
// <root>/.dot/manifest.json. Creates .dot/ if it does not exist.
func Save(root string, ctx *Context, manifest *Manifest) error {
	dir := filepath.Join(root, dotDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create .dot dir: %w", err)
	}

	if err := writeJSON(filepath.Join(dir, "config.json"), ctx); err != nil {
		return err
	}
	if manifest != nil {
		if err := writeJSON(filepath.Join(dir, "manifest.json"), manifest); err != nil {
			return err
		}
	}
	return nil
}

// BuildManifest creates a Manifest from a slice of FileOps by hashing the
// files that were written to disk. Call after pipeline.Run succeeds.
func BuildManifest(ops []generator.FileOp) (*Manifest, error) {
	m := &Manifest{Files: make(map[string]FileRecord, len(ops))}
	for _, op := range ops {
		if op.Kind != generator.Create && op.Kind != generator.Template {
			continue // only track files we fully own
		}
		hash, err := hashFile(op.Path)
		if err != nil {
			return nil, fmt.Errorf("manifest hash %s: %w", op.Path, err)
		}
		m.Files[op.Path] = FileRecord{Hash: hash, Generator: op.Generator}
	}
	return m, nil
}

// CommandsFromDefs converts a []CommandDef slice into the map format stored
// in .dot/config.json. The key is CommandDef.Name (e.g. "new route").
func CommandsFromDefs(defs []generator.CommandDef) map[string]CommandRef {
	m := make(map[string]CommandRef, len(defs))
	for _, d := range defs {
		m[d.Name] = CommandRef{Generator: d.Generator, Action: d.Action}
	}
	return m
}

func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", filepath.Base(path), err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil)), nil
}
