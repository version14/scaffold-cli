# FAQ — dot

Frequently asked questions about developing dot.

---

## General

**Q: Where do I start?**
See [Getting Started](../getting-started/README.md) for the full setup guide.

**Q: I found a bug. How do I report it?**
Open a [Bug Report issue](../../../issues/new/choose) using the provided template.

**Q: I want to add a feature. Where do I begin?**
Open a [Feature Request issue](../../../issues/new/choose) first to discuss the idea. See [Adding a Generator](#adding-a-generator) below for implementation guidance.

**Q: How does dot work?**
It uses a **generator-based architecture**:
1. User answers TUI questions → builds a typed `Spec`
2. `Registry.ForSpec` finds generators matching the spec's language and modules
3. Each generator's `Apply(spec)` returns `[]FileOp` (create, template, append, patch)
4. The pipeline collects all ops, resolves conflicts, and writes atomically
5. `.dot/config.json` and `.dot/manifest.json` are written to the project root

See [docs/getting-started/README.md](../getting-started/README.md#project-structure) for a structural overview.

---

## Development

**Q: What's the easiest way to build and run?**
```bash
make dev     # Build and run with colored output
make run     # Run without building
make build   # Just build to bin/dot
```

**Q: How do I run a specific test?**
```bash
go test -v ./internal/generator -run TestRegistry_ForSpec
```

Or run all tests:
```bash
make test
```

**Q: How do I debug a generator?**
Add print statements or use Delve:
```bash
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug ./cmd/dot
```

**Q: Tests are failing locally but passing in CI (or vice versa).**
- Ensure your Go version matches [Prerequisites](../getting-started/README.md#prerequisites)
- Run `go mod tidy && go mod download` to sync dependencies

**Q: The build fails with module not found errors.**
```bash
go mod download
```

**Q: How do I validate all my changes before submitting a PR?**
```bash
make validate
```
This runs: formatting → vet → lint → tests.

---

## Adding a Generator

**Q: How do I add a new generator (e.g., Redis)?**

1. Create `generators/go/redis.go`
2. Implement the `Generator` interface from `internal/generator/generator.go`:
   ```go
   type GoRedisGenerator struct{}

   func (g *GoRedisGenerator) Name() string     { return "go-redis" }
   func (g *GoRedisGenerator) Language() string { return "go" }
   func (g *GoRedisGenerator) Modules() []string { return []string{"redis"} }

   func (g *GoRedisGenerator) Apply(spec generator.Spec) ([]generator.FileOp, error) {
       // Return FileOps for Redis setup
   }

   func (g *GoRedisGenerator) Commands() []generator.CommandDef {
       // Return commands this generator registers (e.g., "new cache-key")
   }

   func (g *GoRedisGenerator) RunAction(action string, args []string, spec generator.Spec) ([]generator.FileOp, error) {
       // Handle post-creation commands
   }
   ```
3. Register it in `cmd/dot/init.go` with `registry.Register(&GoRedisGenerator{})`
4. Write tests in `generators/go/redis_test.go`

See `generators/go/rest_api.go` for a complete stub example.

---

## Contributing

**Q: How large should a PR be?**
Aim for PRs reviewable in under 30 minutes. Split larger changes.

**Q: Do I need to write tests for every change?**
Yes for new features and bug fixes. Documentation-only PRs are exempt.

**Q: Who merges PRs?**
Maintainers merge PRs once they have one approving review and all CI checks are green.

**Q: What's the PR submission checklist?**
```bash
make validate
```
This checks formatting, vet, lint, and tests. Then verify documentation is updated.

---

Still stuck? Open a [Discussion](../../../discussions).
