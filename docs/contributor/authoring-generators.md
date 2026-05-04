# Authoring Generators

A **generator** is a Go struct that receives the user's answers and writes files into a `VirtualProjectState`. This guide covers the generator interface, how to write files (raw, JSON, YAML, GoMod), how to declare a Manifest, and how to write validators and post-gen commands.

---

## Table of Contents

- [Generator interface](#generator-interface)
- [Context](#context)
- [VirtualProjectState](#virtualprojectstate)
- [Writing files](#writing-files)
  - [From github repository](#from-github-repository)
  - [From external URL](#from-external-url)
  - [From local folder](#from-local-folder)
  - [Raw content](#raw-content)
  - [JSON](#json)
  - [YAML](#yaml)
  - [go.mod](#go-mod)
- [The Manifest](#the-manifest)
- [Dependencies and conflicts](#dependencies-and-conflicts)
- [Validators](#validators)
- [PostGenerationCommands and TestCommands](#postgenerationcommands-and-testcommands)
- [Registering a generator](#registering-a-generator)
- [Loop generators](#loop-generators)
- [Built-in generators](#built-in-generators)

---

## Generator interface

```go
// pkg/dotapi/generator.go
type Generator interface {
    Name()    string
    Version() string
    Generate(ctx *Context) error
}
```

`Name()` must match `Manifest.Name`. `Version()` must match `Manifest.Version`. The engine checks these at registration time.

A minimal generator:

```go
package mygenerator

import "github.com/version14/dot/pkg/dotapi"

type MyGenerator struct{}

func (g *MyGenerator) Name()    string { return "my_generator" }
func (g *MyGenerator) Version() string { return "0.1.0" }

func (g *MyGenerator) Generate(ctx *dotapi.Context) error {
    name, _ := ctx.Answers["project_name"].(string)
    ctx.State.WriteRaw("hello.txt", []byte("Hello, "+name+"!\n"))
    return nil
}
```

---

## Context

`dotapi.Context` is the per-invocation handle:

```go
type Context struct {
    Spec         *spec.ProjectSpec          // read-only full spec
    Answers      map[string]interface{}     // scoped answers (globals + loop frame)
    State        *state.VirtualProjectState // target filesystem
    PreviousGens []string                   // names of generators already run
    Logger       Logger                     // log sink
}
```

### Answers

`ctx.Answers` is a flat map. For loop-aware generators, loop frame answers overlay the global answers — the generator does not need to know it is inside a loop:

```go
// Works for both non-loop and loop invocations:
serviceName, _ := ctx.Answers["service_name"].(string)
```

### Checking previous generators

Use `ctx.PreviousGens` to guard conditional writes:

```go
import "slices"

if slices.Contains(ctx.PreviousGens, "typescript_base") {
    // typescript is in the project — write tsconfig extension
}
```

---

## VirtualProjectState

`ctx.State` holds the entire in-memory project. All generator writes go here; nothing touches the disk until `state.Persist` is called.

### Path conventions

All paths are relative to the project root:

```go
ctx.State.WriteRaw("src/index.ts", content)
ctx.State.WriteRaw("packages/ui/package.json", content)
```

Do not use absolute paths or `../` segments.

---

## Writing files


### From github repository

When creating a generator, you may want to retrieve files from an external GitHub repository. The `VirtualProjectState` provides a `WriteFilesFromGitHub` method to simplify this process. This method allows you to fetch and write files from a specified GitHub repository, enabling you to reuse templates, configurations, or other assets across different projects.

#### Usage

To use `WriteFilesFromGitHub`, you need to provide the repository owner, name, and a list of `FileMapping` objects. Each `FileMapping` specifies the source path in the repository and the destination path in your project.

```go
// From a generator
func (g *MyGenerator) Generate(ctx *dotapi.Context) error {
    mappings := []state.FileMapping{
        {Source: "templates/README.md", Destination: "README.md"},
        {Source: "config/.env.example", Destination: ".env"},
    }

    err := ctx.State.WriteFilesFromGitHub("owner", "repo-name", mappings)
    if err != nil {
        return fmt.Errorf("failed to write files from github: %w", err)
    }

    return nil
}
```
This will fetch `templates/README.md` and `config/.env.example` from the `owner/repo-name` repository and write them to `README.md` and `.env` in your project, respectively.

### From external URL

In addition to fetching files from a Git repository, you can also retrieve content from any external URL. The `VirtualProjectState` provides a `WriteFileFromExternal` method that fetches content from a given URL and writes it to a specified file path in your project. This is useful for retrieving configuration files, scripts, or any other assets hosted on the web.

#### Usage

To use `WriteFileFromExternal`, provide the destination file path and the URL of the external content. The method will fetch the content and write it to the specified path.

```go
// From a generator
func (g *MyGenerator) Generate(ctx *dotapi.Context) error {
    err := ctx.State.WriteFileFromExternal("config/config.json", "http://example.com/config.json")
    if err != nil {
        return fmt.Errorf("failed to write file from external: %w", err)
    }

    return nil
}
```
This will fetch the content from `http://example.com/config.json` and write it to `config/config.json` in your project.


### From local folder

To scaffold a project from a local template directory, you can use the `local.Renderer`. This is particularly useful for complex generators with many template files, as it keeps the templates separate from the Go code.

The renderer takes a source `fs.FS` (like from `embed.FS`), a source directory within the filesystem, a target directory on disk, and data for template execution.

It walks the source directory and performs one of two actions for each file:
- If the file name ends in `.tmpl`, it's treated as a Go template and executed with the provided data. The `.tmpl` suffix is removed from the destination file name.
- Otherwise, the file is copied directly to the destination.

#### Usage

First, embed your template directory in your generator file:

```go
import "embed"

//go:embed all:templates/my_skeleton
var mySkeletonFS embed.FS

const skeletonDir = "templates/my_skeleton"
```
**Note:** The `all:` prefix is required for `go:embed` to embed a directory.

Then, in your generator's `Generate` method, create and run the renderer:

```go
import "github.com/version14/dot/pkg/plugins/scaffolder/local"

func (g *MyGenerator) Generate(ctx *dotapi.Context) error {
    // This can be any struct or map
    templateData := struct {
        ProjectName string
        Author      string
    }{
        ProjectName: ctx.Answers["project_name"].(string),
        Author:      "The DOT team",
    }

    // Assume TargetDir is the root of the new project
    renderer := local.NewRenderer(mySkeletonFS, skeletonDir, ctx.State.GetTargetDir(), templateData)
    if err := renderer.Render(); err != nil {
        return fmt.Errorf("failed to render local templates: %w", err)
    }

    return nil
}
```

This will process all files from the embedded `templates/my_skeleton` directory and write them to the project's target directory.

### Raw content

```go
ctx.State.WriteRaw("README.md", []byte("# My Project\n"))
```

Use for Markdown, plain text, shell scripts, or any file format without a structured helper.

### JSON

```go
doc := ctx.State.OpenJSON("package.json")
doc.Set("name", projectName)
doc.Set("version", "0.1.0")
doc.Set("private", true)
doc.SetPath([]string{"scripts", "dev"}, "vite")
```

`OpenJSON` returns an existing `*JSONDoc` if the file already exists, or creates a new one. This allows multiple generators to contribute to the same `package.json`.

`doc.Set(key, value)` sets a top-level key. `doc.SetPath(keys, value)` sets a nested key. Both overwrite.

### YAML

```go
doc := ctx.State.OpenYAML("docker-compose.yml")
doc.Set("version", "3.9")
doc.SetPath([]string{"services", serviceName, "image"}, image)
```

Same cooperative pattern as JSON.

### go.mod

```go
gomod := ctx.State.OpenGoMod("go.mod")
gomod.SetModule(modulePath)
gomod.SetGoVersion("1.26")
gomod.AddRequire("github.com/charmbracelet/huh", "v1.0.0")
```

`OpenGoMod` parses an existing `go.mod` if present. Safe to call from multiple generators; the last write for a given key wins.

---

## The Manifest

Every generator package exports a `Manifest` variable at package scope:

```go
// generators/my_generator/manifest.go
package mygenerator

import "github.com/version14/dot/pkg/dotapi"

var Manifest = dotapi.Manifest{
    Name:        "my_generator",
    Version:     "0.1.0",
    Description: "Scaffolds my stack",
    DependsOn:   []string{"base_project"},
    Outputs:     []string{"src/index.ts", "tsconfig.json"},
    Validators: []dotapi.Validator{
        {
            Name: "my-files",
            Checks: []dotapi.Check{
                {Type: dotapi.CheckFileExists, Path: "src/index.ts"},
            },
        },
    },
    PostGenerationCommands: []dotapi.Command{
        {Cmd: "pnpm install", WorkDir: ""},
    },
}
```

---

## Versioning and semver constraints

`Manifest.Version` is a semver string (`"0.1.0"`, `"1.2.3-beta"`). `dot doctor` compares recorded constraints against the installed version using the constraint syntax below.

### Constraint syntax

| Expression | Meaning | Example passes |
|------------|---------|---------------|
| `1.2.3` or `=1.2.3` | Exact match | `1.2.3` |
| `>=1.2.3` | At least | `1.2.3`, `1.3.0`, `2.0.0` |
| `>1.2.3` | Strictly greater | `1.2.4`, `2.0.0` |
| `<=1.2.3` | At most | `1.2.3`, `1.0.0` |
| `<1.2.3` | Strictly less | `1.2.2`, `0.9.0` |
| `~1.2.3` | Same major + minor, patch ≥ 3 | `1.2.3`, `1.2.9` — **not** `1.3.0` |
| `^1.2.3` | Same major, version ≥ 1.2.3 | `1.2.3`, `1.9.0` — **not** `2.0.0` |
| `^0.2.3` | Same major=0 + minor, patch ≥ 3 | `0.2.3`, `0.2.9` — **not** `0.3.0` |
| _(empty)_ | Accept any version | always passes |

Use `^` for stable packages (allows minor bumps). Use `~` to lock to a patch range. Use exact match only when a specific API version is required.

Constraints are parsed by `internal/versioning` and stored in `.dot/spec.json` under `generator_constraints`. You normally do not set them manually — `dot doctor` reads the version from `Manifest.Version` at generation time.

---

## Dependencies and conflicts

### DependsOn

List generator names that must run before yours. The resolver does a topological sort and places your generator after all its dependencies.

```go
DependsOn: []string{"base_project", "typescript_base"},
```

If a listed dependency is not in the invocation set, the resolver adds it automatically (transitive dep expansion).

### ConflictsWith

List generator names that may not coexist with yours. The resolver returns an error if both are requested.

```go
ConflictsWith: []string{"webpack_config"},
```

---

## Validators

Validators are structural checks the engine runs after generation and the `dot doctor` command runs on subsequent runs against the on-disk project.

```go
Validators: []dotapi.Validator{
    {
        Name: "structure",
        Checks: []dotapi.Check{
            {Type: dotapi.CheckFileExists, Path: "src/index.ts"},
            {Type: dotapi.CheckFileExists, Path: "tsconfig.json"},
            {Type: dotapi.CheckJSONKeyExists, Path: "package.json", Key: "scripts.dev"},
        },
    },
},
```

**Check types**

| Type | Fields | Passes when |
|------|--------|-------------|
| `CheckFileExists` | `Path` | The file exists in the virtual state (or on disk for `dot doctor`) |
| `CheckJSONKeyExists` | `Path`, `Key` | The JSON file at `Path` contains the dotted `Key` (e.g. `"scripts.dev"`) |

More check types can be added in `pkg/dotapi/manifest.go` and implemented in `internal/generator/validator.go`.

---

## PostGenerationCommands and TestCommands

### PostGenerationCommands

Run after the entire generator pipeline has finished and files have been persisted:

```go
PostGenerationCommands: []dotapi.Command{
    {Cmd: "pnpm install"},
    {Cmd: "go mod tidy", WorkDir: "api"},
},
```

Commands from all generators are deduplicated (by `Cmd + WorkDir`) and run in declaration order. Use `{key}` tokens for answer substitution:

```go
{Cmd: "go mod init {module_path}", WorkDir: ""},
```

### TestCommands

Run by `test-flow` to verify the generated project works. They are not run during normal scaffolding.

```go
TestCommands: []dotapi.Command{
    {Cmd: "pnpm run build"},
    {
        Cmd:        "pnpm run dev",
        Background: true,
        ReadyDelay: 3 * time.Second,
    },
},
```

Background commands are started, waited on for `ReadyDelay`, checked for crash, and then sent `SIGTERM`. This lets you test that a dev server starts without having to stop it manually.

---

## Registering a generator

Add the generator to `internal/cli/registry.go`:

```go
func DefaultGeneratorRegistry() (*generator.Registry, error) {
    r := generator.NewRegistry()

    entries := []generator.Entry{
        {Manifest: baseproject.Manifest, Generator: &baseproject.Generator{}},
        {Manifest: mygenerator.Manifest, Generator: &mygenerator.MyGenerator{}},
        // ...
    }

    for _, e := range entries {
        if err := r.Register(e); err != nil {
            return nil, err
        }
    }
    return r, nil
}
```

The generator immediately appears in `dot generators`.

---

## Loop generators

A generator that participates in a loop receives each iteration's answers via scoped `ctx.Answers`. The `LoopStack` in the invocation tells the executor which frame to overlay.

Write the generator as if it always receives a single set of answers:

```go
func (g *ServiceWriter) Generate(ctx *dotapi.Context) error {
    name, _ := ctx.Answers["service_name"].(string)
    port, _ := ctx.Answers["service_port"].(string)

    // Write to a directory named after the service:
    ctx.State.WriteRaw(
        filepath.Join("services", name, "main.go"),
        renderMain(name, port),
    )
    return nil
}
```

Each loop iteration is a separate `generator.Invocation` — the same generator function is called multiple times, once per iteration, each time with different `ctx.Answers`. Files from different iterations go into different paths (controlled by the generator itself, typically using the service name as a subdirectory).

---

## Built-in generators

| Name | Package | Purpose |
|------|---------|---------|
| `base_project` | `generators/base_project` | README, .gitignore, LICENSE — always runs first |
| `typescript_base` | `generators/typescript_base` | tsconfig.json, package.json, tooling |
| `react_app` | `generators/react_app` | Vite, React Router, Tailwind; depends on `typescript_base` |
| `biome_config` | `generators/biome_config` | biome.json formatter/linter config |
| `service_writer` | `generators/service_writer` | Go microservice (HTTP server, Dockerfile, healthcheck) |
| `plugin_repo_skeleton` | `generators/plugin_repo_skeleton` | Full DOT plugin repository scaffold |
| `python_fastapi_base` | `generators/python_fastapi_base` | Base FastAPI app with `/health` endpoint |
| `python_fastapi_auth` | `generators/python_fastapi_auth` | FastAPI auth routes `/register` and `/login` |
