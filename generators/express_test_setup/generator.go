package expresstestsetup

import (
	"embed"

	"github.com/version14/dot/internal/render"
	"github.com/version14/dot/internal/state"
	"github.com/version14/dot/pkg/dotapi"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string    { return Manifest.Name }
func (g *Generator) Version() string { return Manifest.Version }

//go:embed all:files
var fs embed.FS

func (g *Generator) Generate(ctx *dotapi.Context) error {
	renderer := render.NewLocalFolderRenderer(ctx.State)
	if err := renderer.Render(fs, nil); err != nil {
		return err
	}
	return ctx.State.UpdateJSON("package.json", func(d *state.JSONDoc) error {
		d.Merge(map[string]interface{}{
			"scripts": map[string]interface{}{
				"test":          "vitest run",
				"test:coverage": "vitest run --coverage",
			},
			"devDependencies": map[string]interface{}{
				"vitest":              "^2.0.0",
				"@vitest/coverage-v8": "^2.0.0",
				"supertest":           "^7.0.0",
				"@types/supertest":    "^6.0.0",
			},
		})
		return nil
	})
}
