package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/version14/dot/internal/generator"
	"github.com/version14/dot/pkg/dotapi"
)

// cacheSchemaVersion is bumped whenever the on-disk cache format or the
// fingerprint algorithm changes. Older entries become invalid automatically.
//   - v1: initial Cacheable opt-in (default false)
//   - v2: flipped polarity to NoCache opt-out (default cacheable)
const cacheSchemaVersion = 2

// cacheRoot is where successful case fingerprints are persisted. Kept under
// the repository root so it follows the working copy (and is gitignored).
const cacheRoot = ".test-flow-cache"

// CacheEntry records the outcome of one successful case run. We only persist
// passing runs — failed runs intentionally leave no trace so the next
// invocation re-tries them.
type CacheEntry struct {
	SchemaVersion int    `json:"schema_version"`
	Fingerprint   string `json:"fingerprint"`
	CaseName      string `json:"case_name"`
	FlowID        string `json:"flow_id"`
	LastSuccessAt string `json:"last_success_at"`
	// Generators that contributed to this fingerprint. Stored for human
	// inspection only; not used during equality checks.
	Generators []string `json:"generators"`
}

// CacheKeyInputs aggregates everything that must be hashed to produce the
// case fingerprint. The runner fills it as soon as scaffolding has resolved
// the invocation list.
type CacheKeyInputs struct {
	CaseFile      string                 // absolute path to the testdata JSON
	FlowsDir      string                 // absolute path to the flows/ directory (whole dir is hashed)
	Invocations   []generator.Invocation // resolved generator list
	Manifests     []dotapi.Manifest      // matches Invocations
	SkipPostFlag  bool                   // -skip-post CLI flag
	SkipTestFlag  bool                   // -skip-test CLI flag
	GeneratorsDir string                 // absolute path to repo's generators/ dir
	RepoRoot      string                 // absolute path to the repo root
}

// ComputeFingerprint hashes everything that can plausibly change a case's
// behaviour: the testdata file, every involved generator's source tree, the
// flow definition file, the Manifest schema (pkg/dotapi), and the test-flow
// runner itself. CLI flags that change command execution (`-skip-post`,
// `-skip-test`) are folded in too so different modes get different cache
// slots.
func ComputeFingerprint(in CacheKeyInputs) (string, error) {
	h := sha256.New()
	fmt.Fprintf(h, "schema:%d\n", cacheSchemaVersion)

	caseBytes, err := os.ReadFile(in.CaseFile)
	if err != nil {
		return "", fmt.Errorf("hash case file: %w", err)
	}
	fmt.Fprintf(h, "case:%s\n", sha256Bytes(caseBytes))

	if in.FlowsDir != "" {
		if flowsHash, err := hashDir(in.FlowsDir); err == nil {
			fmt.Fprintf(h, "flows-dir:%s\n", flowsHash)
		}
	}

	// Hash every involved generator's source tree. Order by name so the
	// fingerprint is stable regardless of resolver output order.
	names := make([]string, 0, len(in.Invocations))
	for _, inv := range in.Invocations {
		names = append(names, inv.Name)
	}
	sort.Strings(names)
	for _, name := range names {
		dir := filepath.Join(in.GeneratorsDir, name)
		genHash, err := hashDir(dir)
		if err != nil {
			// A missing dir means the generator is registered out-of-tree
			// (a plugin). Hash an empty marker so different sets stay
			// distinguishable.
			fmt.Fprintf(h, "gen:%s:absent\n", name)
			continue
		}
		fmt.Fprintf(h, "gen:%s:%s\n", name, genHash)
	}

	// pkg/dotapi controls the Manifest schema; touching it reasonably
	// invalidates every case.
	if dotapiHash, err := hashDir(filepath.Join(in.RepoRoot, "pkg", "dotapi")); err == nil {
		fmt.Fprintf(h, "pkg-dotapi:%s\n", dotapiHash)
	}

	// The test-flow tool itself can change semantics (cache logic included);
	// hashing its source guarantees the cache invalidates on tool edits.
	if toolHash, err := hashDir(filepath.Join(in.RepoRoot, "tools", "test-flow")); err == nil {
		fmt.Fprintf(h, "tool-test-flow:%s\n", toolHash)
	}

	fmt.Fprintf(h, "skip-post:%t\n", in.SkipPostFlag)
	fmt.Fprintf(h, "skip-test:%t\n", in.SkipTestFlag)

	return hex.EncodeToString(h.Sum(nil)), nil
}

// AllCommandsCacheable returns true when no PostGenerationCommand or
// TestCommand across the supplied manifests opted out of caching via
// NoCache. The case-level cache only fires when this is true — a single
// NoCache command (docker compose up, dev-server probe, network call) is
// enough to force the case to re-run.
func AllCommandsCacheable(manifests []dotapi.Manifest) bool {
	for _, m := range manifests {
		for _, c := range m.PostGenerationCommands {
			if c.NoCache {
				return false
			}
		}
		for _, c := range m.TestCommands {
			if c.NoCache {
				return false
			}
		}
	}
	return true
}

// NonCacheableCommands lists the (generator, command) pairs that block the
// cache from short-circuiting the case. Useful for the reporter so the user
// understands why caching did not apply.
func NonCacheableCommands(manifests []dotapi.Manifest) []string {
	var blocking []string
	for _, m := range manifests {
		for _, c := range m.PostGenerationCommands {
			if c.NoCache {
				blocking = append(blocking, m.Name+" • post • "+c.Cmd)
			}
		}
		for _, c := range m.TestCommands {
			if c.NoCache {
				blocking = append(blocking, m.Name+" • test • "+c.Cmd)
			}
		}
	}
	return blocking
}

// LoadCacheEntry reads the per-case JSON record. Returns (nil, nil) when the
// cache file is missing — that is not an error.
func LoadCacheEntry(repoRoot, caseName string) (*CacheEntry, error) {
	path := cacheFilePath(repoRoot, caseName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read cache %s: %w", path, err)
	}
	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("decode cache %s: %w", path, err)
	}
	if entry.SchemaVersion != cacheSchemaVersion {
		return nil, nil
	}
	return &entry, nil
}

// SaveCacheEntry persists a successful run. Failures intentionally never
// write — leaving no entry forces the next invocation to retry.
func SaveCacheEntry(repoRoot string, entry CacheEntry) error {
	path := cacheFilePath(repoRoot, entry.CaseName)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func cacheFilePath(repoRoot, caseName string) string {
	return filepath.Join(repoRoot, cacheRoot, sanitizeName(caseName)+".json")
}

func sanitizeName(name string) string {
	out := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-', c == '_':
			out = append(out, c)
		default:
			out = append(out, '_')
		}
	}
	return string(out)
}

func sha256Bytes(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// hashDir produces a content hash that depends on every file under root —
// sorted by path so ordering is deterministic. Symlinks and hidden files
// are included; that's deliberate (test-flow templates can hide anywhere).
func hashDir(root string) (string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		// Treat a single file as a one-element directory.
		b, err := os.ReadFile(root)
		if err != nil {
			return "", err
		}
		return sha256Bytes(b), nil
	}

	type fileEntry struct {
		path string
		hash string
	}
	var files []fileEntry

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		b, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		rel, _ := filepath.Rel(root, path)
		files = append(files, fileEntry{path: rel, hash: sha256Bytes(b)})
		return nil
	})
	if err != nil {
		return "", err
	}

	sort.Slice(files, func(i, j int) bool { return files[i].path < files[j].path })

	h := sha256.New()
	for _, f := range files {
		fmt.Fprintf(h, "%s\x00%s\n", f.path, f.hash)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
