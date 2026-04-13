package generator

import (
	"fmt"

	"github.com/version14/dot/internal/spec"
)

// Registry holds registered generators and resolves which ones apply to a
// given Spec. It is created once at startup and passed through the CLI.
type Registry struct {
	generators []Generator
}

// Register adds a generator to the registry.
// Returns an error if another already-registered generator claims the same
// (language, module) pair. Language "*" conflicts with any language-specific
// generator claiming the same module name. This catches conflicts at startup
// rather than mid-write when a FileOp conflict surfaces.
func (r *Registry) Register(g Generator) error {
	for _, existing := range r.generators {
		for _, newMod := range g.Modules() {
			for _, existMod := range existing.Modules() {
				if newMod != existMod {
					continue
				}
				newLang := g.Language()
				existLang := existing.Language()
				// Conflict if: same language, or either is language-agnostic ("*")
				if newLang == existLang || newLang == "*" || existLang == "*" {
					return fmt.Errorf(
						"generator %q conflicts with %q: both claim module %q for language %q",
						g.Name(), existing.Name(), newMod, newLang,
					)
				}
			}
		}
	}
	r.generators = append(r.generators, g)
	return nil
}

// ForSpec returns all generators that match the spec's language and have at
// least one module in common with the requested modules.
//
// Matching rule:
//   - generator.Language() == spec.Project.Language OR generator.Language() == "*"
//   - AND at least one of generator.Modules() appears in spec.Modules[].Name
func (r *Registry) ForSpec(s spec.Spec) []Generator {
	requested := make(map[string]bool, len(s.Modules))
	for _, m := range s.Modules {
		requested[m.Name] = true
	}

	var matched []Generator
	for _, g := range r.generators {
		if g.Language() != s.Project.Language && g.Language() != "*" {
			continue
		}
		for _, mod := range g.Modules() {
			if requested[mod] {
				matched = append(matched, g)
				break
			}
		}
	}
	return matched
}

// CommandsForSpec returns all CommandDefs from generators matched by the spec.
// These are written to .dot/config.json as available_commands after dot init.
func (r *Registry) CommandsForSpec(s spec.Spec) []CommandDef {
	var cmds []CommandDef
	for _, g := range r.ForSpec(s) {
		cmds = append(cmds, g.Commands()...)
	}
	return cmds
}

// Get returns a generator by Name. Used during dot new dispatch.
func (r *Registry) Get(name string) (Generator, bool) {
	for _, g := range r.generators {
		if g.Name() == name {
			return g, true
		}
	}
	return nil, false
}
