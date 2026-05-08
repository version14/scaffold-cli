package drizzletypescriptdeps

import (
	"github.com/version14/dot/internal/state"
	"github.com/version14/dot/pkg/dotapi"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string    { return Manifest.Name }
func (g *Generator) Version() string { return Manifest.Version }

func (g *Generator) Generate(ctx *dotapi.Context) error {
	return ctx.State.UpdateJSON("package.json", func(d *state.JSONDoc) error {
		d.Merge(map[string]interface{}{
			"scripts": map[string]interface{}{
				"db:generate": "drizzle-kit generate",
				"db:migrate":  "drizzle-kit migrate",
				"db:push":     "drizzle-kit push",
				"db:studio":   "drizzle-kit studio",
			},
			"dependencies": map[string]interface{}{
				"drizzle-orm": "^0.44.0",
			},
			"devDependencies": map[string]interface{}{
				"drizzle-kit": "^0.30.0",
			},
		})
		return nil
	})
}
