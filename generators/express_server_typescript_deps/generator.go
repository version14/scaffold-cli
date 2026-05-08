package expressservertypescriptdeps

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
				"dev":   "nodemon --exec tsx src/index.ts",
				"build": "tsc",
				"start": "node dist/index.js",
			},
			"dependencies": map[string]interface{}{
				"cors":    "^2.8.5",
				"dotenv":  "^16.4.0",
				"express": "^4.21.0",
			},
			"devDependencies": map[string]interface{}{
				"@types/cors":    "^2.8.17",
				"@types/express": "^5.0.0",
				"@types/node":    "^22.0.0",
				"nodemon":        "^3.1.0",
				"tsx":            "^4.19.0",
			},
		})
		return nil
	})
}
