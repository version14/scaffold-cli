// Package spec defines the core types that flow through the dot pipeline.
// Every input layer (CLI TUI, dot.yaml PaC, future MCP) produces a Spec.
// The generator engine consumes it. Neither side knows about the other.
package spec

// ProjectType identifies the broad shape of a project.
type ProjectType string

const (
	ProjectTypeAPI      ProjectType = "api"
	ProjectTypeCLI      ProjectType = "cli"
	ProjectTypeMonorepo ProjectType = "monorepo"
	ProjectTypeLibrary  ProjectType = "library"
	ProjectTypeFrontend ProjectType = "frontend"
	ProjectTypeWorker   ProjectType = "worker"
)

// ProjectSpec holds the top-level identity of a project.
type ProjectSpec struct {
	Name     string      `json:"name"`
	Language string      `json:"language"` // open string: "go", "python", "typescript", ...
	Type     ProjectType `json:"type"`
}

// CoreConfig holds configuration fields used by official generators.
// Official generators read typed fields here (e.g. spec.Config.Linter).
// Community generators use Extensions instead.
type CoreConfig struct {
	Linter     string `json:"linter"`     // "golangci-lint" | "eslint" | "none"
	Formatter  string `json:"formatter"`  // "gofmt" | "goimports" | "prettier" | "none"
	CI         string `json:"ci"`         // "github-actions" | "gitlab-ci" | "none"
	Deployment string `json:"deployment"` // "docker" | "docker-compose" | "vercel" | "none"
	Monitoring string `json:"monitoring"` // "grafana" | "datadog" | "none"
	Tracking   string `json:"tracking"`   // "posthog" | "sentry" | "none"
}

// ModuleSpec represents one module requested in the project.
// Config holds module-specific options (e.g. postgres pool size).
type ModuleSpec struct {
	Name   string         `json:"name"`
	Config map[string]any `json:"config,omitempty"`
}

// Spec is the authoritative description of a project. It is produced by
// input layers (TUI, YAML) and consumed by the generator engine.
type Spec struct {
	Project    ProjectSpec    `json:"project"`
	Modules    []ModuleSpec   `json:"modules"`
	Config     CoreConfig     `json:"config"`
	Extensions map[string]any `json:"extensions,omitempty"` // community generators use this
}

// ModuleNames returns the name of every module in the spec.
func (s Spec) ModuleNames() []string {
	names := make([]string, len(s.Modules))
	for i, m := range s.Modules {
		names[i] = m.Name
	}
	return names
}

// HasModule reports whether the spec includes a module with the given name.
func (s Spec) HasModule(name string) bool {
	for _, m := range s.Modules {
		if m.Name == name {
			return true
		}
	}
	return false
}
