// Package generator defines the Generator interface, Registry, and CommandDef.
// Every official and community generator implements Generator and registers
// with a Registry. The engine calls ForSpec to find the right generators for
// a given project, then calls Apply to collect FileOps.
package generator

import "github.com/version14/dot/internal/spec"

// Generator is the core extension point for dot. Implement this interface to
// add support for a new language, framework, or pattern.
//
// Name and Language together form the generator's identity. Two generators
// must not claim the same (Language, Module) pair — Register will reject it.
type Generator interface {
	// Name returns a unique, stable identifier for this generator.
	// Used as the key in .dot/config.json commands. Example: "go-rest-api".
	Name() string

	// Language returns the language this generator targets.
	// Use "*" for language-agnostic generators (e.g. GitHub Actions CI).
	Language() string

	// Modules returns the module names this generator handles.
	// Example: []string{"rest-api", "postgres"}.
	Modules() []string

	// Apply is called during dot init. It receives the full project Spec and
	// returns the FileOps needed to scaffold the generator's files.
	Apply(s spec.Spec) ([]FileOp, error)

	// Commands returns the post-creation commands this generator registers.
	// These are written to .dot/config.json after dot init so that dot new
	// can dispatch them later.
	Commands() []CommandDef

	// RunAction is called by dot new. action is the CommandDef.Action value,
	// args are the positional arguments (e.g. ["UserController"]).
	RunAction(action string, args []string, s spec.Spec) ([]FileOp, error)
}

// CommandDef describes a post-creation command registered by a generator.
//
// Name format: "verb noun" where noun is a single hyphenated word with no
// spaces. Examples: "new route", "new handler", "add migration".
// The CLI looks up commands with key = "new " + os.Args[2].
// Do not use spaces in the noun — use hyphens: "new rest-api" not "new rest api".
type CommandDef struct {
	Name        string   `json:"name"`        // "new route" — key in .dot/config.json
	Args        []string `json:"args"`        // ["<name>"] — for dot help display only
	Description string   `json:"description"` // shown in dot help
	Action      string   `json:"action"`      // passed to RunAction, e.g. "rest-api.new-route"
	Generator   string   `json:"generator"`   // matches Generator.Name()
}
