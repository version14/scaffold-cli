package monorepotsworkspaces

import (
	"github.com/version14/dot/internal/state"
	"github.com/version14/dot/pkg/dotapi"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string    { return Manifest.Name }
func (g *Generator) Version() string { return Manifest.Version }

func (g *Generator) Generate(ctx *dotapi.Context) error {
	projectName, _ := ctx.Answers["project_name"].(string)
	if projectName == "" {
		projectName = ctx.Spec.Metadata.ProjectName
	}
	if projectName == "" {
		projectName = "monorepo"
	}

	if err := ctx.State.UpdateJSON("package.json", func(d *state.JSONDoc) error {
		d.Merge(map[string]interface{}{
			"name":       projectName,
			"version":    "0.1.0",
			"private":    true,
			"workspaces": []interface{}{"apps/*"},
		})
		return nil
	}); err != nil {
		return err
	}

	ctx.State.WriteFile(
		"pnpm-workspace.yaml",
		[]byte("packages:\n  - \"apps/*\"\n"),
		state.ContentRaw,
	)
	return nil
}
