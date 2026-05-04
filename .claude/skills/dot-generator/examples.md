# dot-generator examples

One fully-worked example per mode, plus one example per writing strategy.

---

## Example 1 — `init`

Step 2 answers:

```
name              = hello_world
version           = 0.1.0
description       = Writes a hello.txt at the project root.
depends_on        = []
answers           = [{ key: "project_name", type: "string" }]
outputs           = [{ path: "hello.txt", format: "raw" }]
validators        = []   (CheckFileExists hello.txt is auto-added)
post_gen_commands = []
```

### File created: `generators/hello_world/manifest.go`

```go
package helloworld

import "github.com/version14/dot/pkg/dotapi"

var Manifest = dotapi.Manifest{
	Name:        "hello_world",
	Version:     "0.1.0",
	Description: "Writes a hello.txt at the project root.",
	Outputs:     []string{"hello.txt"},
	Validators: []dotapi.Validator{
		{
			Name: "hello-world",
			Checks: []dotapi.Check{
				{Type: dotapi.CheckFileExists, Path: "hello.txt"},
			},
		},
	},
}
```

### File created: `generators/hello_world/generator.go`

```go
// Package helloworld writes a hello.txt at the project root.
package helloworld

import "github.com/version14/dot/pkg/dotapi"

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string    { return Manifest.Name }
func (g *Generator) Version() string { return Manifest.Version }

func (g *Generator) Generate(ctx *dotapi.Context) error {
	// TODO: implement — writes hello.txt
	_ = ctx
	return nil
}
```

### File modified: `internal/cli/registry.go`

```go
import (
	// ... existing imports ...
	helloworld "github.com/version14/dot/generators/hello_world"
)

func builtinGeneratorEntries() []generator.Entry {
	return []generator.Entry{
		// ... existing entries ...
		{Manifest: helloworld.Manifest, Generator: helloworld.New()},
	}
}
```

### File created: `docs/contributor/generators/hello_world.md`

Copy of `_template.md` with Identity, Files written, Validators, Post-gen prefilled.

### Files modified

- `docs/contributor/authoring-generators.md` — add a row for `hello_world` in "Built-in generators".
- `docs/README.md` — add a link under the generators index.

---

## Example 2 — `edit` (add a validator)

Step 2 answers:

```
target      = base_project
change_kind = add validator
validator   = { kind: "CheckJSONKeyExists", path: "package.json", key: "scripts.dev" }
```

### Diff to `generators/base_project/manifest.go`

```go
Validators: []dotapi.Validator{
	{
		Name: "base-project",
		Checks: []dotapi.Check{
			{Type: dotapi.CheckFileExists, Path: "README.md"},
			{Type: dotapi.CheckJSONKeyExists, Path: "package.json", Key: "scripts.dev"},
		},
	},
},
```

### Diff to `docs/contributor/generators/base_project.md` (Validators table)

```markdown
| `package.json` / `scripts.dev` | `json_key_exists` | The `scripts.dev` key is present |
```

Then the skill runs `go build ./...` and `make test`.

---

## Example 3 — `generate` (raw strategy, two outputs)

Step 2 answers:

```
name        = readme_writer
version     = 0.1.0
description = Writes README.md and CONTRIBUTING.md from the project name.
outputs     = [
  { path: "README.md",       format: "raw" },
  { path: "CONTRIBUTING.md", format: "raw" }
]
answers     = [{ key: "project_name", type: "string" }]
```

### `generators/readme_writer/generator.go`

```go
package readmewriter

import (
	"fmt"

	"github.com/version14/dot/pkg/dotapi"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string    { return Manifest.Name }
func (g *Generator) Version() string { return Manifest.Version }

func (g *Generator) Generate(ctx *dotapi.Context) error {
	name, _ := ctx.Answers["project_name"].(string)
	if name == "" {
		return fmt.Errorf("readme_writer: missing project_name")
	}
	ctx.State.WriteRaw("README.md", []byte("# "+name+"\n"))
	ctx.State.WriteRaw("CONTRIBUTING.md", []byte("# Contributing to "+name+"\n"))
	return nil
}
```

---

## Example 4 — `generate` (JSON cooperative strategy)

Step 2 answers (excerpt):

```
outputs    = [{ path: "package.json", format: "json" }]
validators = [{ kind: "CheckJSONKeyExists", path: "package.json", key: "scripts.dev" }]
```

### `Generate` body

```go
func (g *Generator) Generate(ctx *dotapi.Context) error {
	name, _ := ctx.Answers["project_name"].(string)
	doc := ctx.State.OpenJSON("package.json")
	doc.Set("name", name)
	doc.Set("version", "0.1.0")
	doc.Set("private", true)
	doc.SetPath([]string{"scripts", "dev"}, "vite")
	return nil
}
```

`OpenJSON` is cooperative — other generators may also write to `package.json` and the writes merge.

---

## Example 5 — `generate` (embed strategy, multi-file template tree)

Mirror of `[generators/plugin_repo_skeleton/](../../../generators/plugin_repo_skeleton)`.

### File created: `generators/saas_starter/generator.go`

```go
package saasstarter

import (
	"embed"

	"github.com/version14/dot/internal/render"
	"github.com/version14/dot/pkg/dotapi"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string    { return Manifest.Name }
func (g *Generator) Version() string { return Manifest.Version }

//go:embed all:files
var fs embed.FS

func (g *Generator) Generate(ctx *dotapi.Context) error {
	name, _ := ctx.Answers["project_name"].(string)
	data := map[string]interface{}{
		"ProjectName": name,
	}
	renderer := render.NewLocalFolderRenderer(ctx.State)
	return renderer.Render(fs, data)
}
```

### File tree created: `generators/saas_starter/files/`

```
generators/saas_starter/files/
├── README.md.tmpl
├── package.json.tmpl
└── src/
    └── index.ts.tmpl
```

`README.md.tmpl`:

```markdown
# {{ .ProjectName }}

Generated by the saas_starter generator.
```

The `.tmpl` suffix is stripped at render time; non-`.tmpl` files are copied verbatim. Every file under `files/` becomes a project output, so `Manifest.Outputs` MUST list each one and have a matching `CheckFileExists`.

---

## Example 6 — `generate` (go.mod strategy)

```go
func (g *Generator) Generate(ctx *dotapi.Context) error {
	modulePath, _ := ctx.Answers["module_path"].(string)
	gomod := ctx.State.OpenGoMod("go.mod")
	gomod.SetModule(modulePath)
	gomod.SetGoVersion("1.26")
	gomod.AddRequire("github.com/charmbracelet/huh", "v1.0.0")
	return nil
}
```

`OpenGoMod` is cooperative the same way `OpenJSON` and `OpenYAML` are — last write wins per key.

---

## After-completion summary the skill prints

```
Created generator: generators/<name>/
  Strategy: <raw | json | yaml | gomod | embed>
  Outputs: <list>
  Doc: docs/contributor/generators/<name>.md
  Registered in: internal/cli/registry.go
  go build ./...: PASS
```
