package zodvalidationdeps

import (
	"github.com/version14/dot/internal/state"
	"github.com/version14/dot/pkg/dotapi"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string    { return Manifest.Name }
func (g *Generator) Version() string { return Manifest.Version }

func (g *Generator) Generate(ctx *dotapi.Context) error {
	if err := ctx.State.UpdateJSON("package.json", func(d *state.JSONDoc) error {
		d.Merge(map[string]interface{}{
			"dependencies": map[string]interface{}{
				"zod":                            "^3.23.8",
				"@asteasolutions/zod-to-openapi": "^7.3.0",
				"reflect-metadata":               "^0.2.2",
				"swagger-ui-express":             "^5.0.1",
			},
			"devDependencies": map[string]interface{}{
				"@types/swagger-ui-express": "^4.1.6",
			},
		})
		return nil
	}); err != nil {
		return err
	}

	return ctx.State.UpdateJSON("tsconfig.json", func(d *state.JSONDoc) error {
		d.Merge(map[string]interface{}{
			"compilerOptions": map[string]interface{}{
				"experimentalDecorators": true,
				"emitDecoratorMetadata":  true,
			},
		})
		return nil
	})
}
