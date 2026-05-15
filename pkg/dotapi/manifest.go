package dotapi

import "time"

// Manifest declares a generator's metadata, dependencies, side effects, and
// post-generation behaviour. Each generator package exposes a package-level
// `Manifest` value of this type and registers it with the engine at startup.
type Manifest struct {
	// Name is the unique generator ID used by flows to reference it.
	Name string

	// Version is a semver string used for compatibility checks on re-runs.
	Version string

	// Description is a short human-readable summary.
	Description string

	// DependsOn lists generator names that must run before this one. The
	// resolver enforces this with a topological sort.
	DependsOn []string

	// ConflictsWith lists generator names that may not appear in the same
	// invocation set. The resolver rejects conflicting plans.
	ConflictsWith []string

	// Outputs lists the file paths (relative to project root) this generator
	// writes. Used by the doctor command and to detect collisions.
	Outputs []string

	// PostGenerationCommands run after every generator has finished and the
	// virtual filesystem has been persisted. Tokens like {name} are
	// interpolated from the generator's scoped Answers.
	PostGenerationCommands []Command

	// TestCommands are scripts the test runner executes against generated
	// projects to verify they build/run.
	TestCommands []Command

	// Validators are structural checks run against the virtual state after
	// generation, and against the on-disk project on re-runs.
	Validators []Validator

	// PathPrefix is a runtime-only field set by the executor/runner when the
	// generator runs inside a loop (e.g. "apps/api"). Validator checks and
	// command WorkDirs are resolved relative to PathPrefix inside the project
	// root. Generators never set this themselves.
	PathPrefix string
}

// Command is a shell command run after generation. WorkDir is relative to the
// generated project root (empty = root). Cmd may contain `{name}`-style tokens
// that the runner substitutes from the invocation's scoped answers.
//
// Background = true marks long-running commands (dev servers, watch tasks).
// The runner starts them, waits ReadyDelay for them to settle, verifies the
// process did not crash, then sends SIGTERM. Foreground commands run to
// completion and their exit code is checked.
//
// NoCache = true tells the test-flow runner that this command must run on
// every invocation regardless of cache state. The DEFAULT (false) is that
// commands are cacheable — i.e. their outcome is assumed deterministic
// given the same scaffolded inputs, and the case-level cache may skip them
// on a fingerprint match. Set NoCache=true when the command depends on
// state outside the project (unpinned network calls, a dev-server probe
// whose port binding you want re-verified on every run) or when you simply
// aren't sure the outcome is deterministic.
//
// A single NoCache=true command anywhere in the resolved invocation set
// forces the entire case to re-run from scratch.
type Command struct {
	Cmd        string
	WorkDir    string
	Background bool
	// ReadyDelay is how long to wait before considering a Background command
	// "ready" (default 3s when zero).
	ReadyDelay time.Duration
	// NoCache opts the command OUT of the test-flow case-level cache.
	// Default (false) means the command is cacheable. See the docstring
	// above for the full contract.
	NoCache bool
}

// Validator is a named bundle of structural checks the engine runs to verify
// generation produced a valid project shape.
type Validator struct {
	Name   string
	Checks []Check
}

// CheckType enumerates the structural assertions Validators express.
type CheckType string

const (
	CheckFileExists    CheckType = "file_exists"
	CheckJSONKeyExists CheckType = "json_key_exists"
)

// Check is one structural assertion.
//   - file_exists:    Path must exist in the (virtual or real) filesystem.
//   - json_key_exists: Path must be a JSON document containing dotted Key.
type Check struct {
	Type CheckType
	Path string
	Key  string
}
