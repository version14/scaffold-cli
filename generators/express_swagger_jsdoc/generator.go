package expressswaggerjsdoc

import (
	"embed"
	"strings"

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

const swaggerImports = "import { mountSwagger } from './shared/swagger';\n"

const swaggerMount = "\nmountSwagger(app);\n"

func (g *Generator) Generate(ctx *dotapi.Context) error {
	if err := render.NewLocalFolderRenderer(ctx.State).Render(fs, nil); err != nil {
		return err
	}

	if err := ctx.State.UpdateJSON("package.json", func(d *state.JSONDoc) error {
		d.Merge(map[string]interface{}{
			"dependencies": map[string]interface{}{
				"swagger-jsdoc":      "^6.2.8",
				"swagger-ui-express": "^5.0.1",
			},
			"devDependencies": map[string]interface{}{
				"@types/swagger-jsdoc":      "^6.0.4",
				"@types/swagger-ui-express": "^4.1.6",
			},
		})
		// swagger-jsdoc bundles @scarf/scarf for install-time analytics.
		return d.AppendStringSet("pnpm.onlyBuiltDependencies", "@scarf/scarf")
	}); err != nil {
		return err
	}

	if f, ok := ctx.State.GetFile("src/app.ts"); ok {
		content := string(f.Content)
		if !strings.Contains(content, "mountSwagger") {
			content = swaggerImports + content
			if strings.Contains(content, "export default app;") {
				content = strings.Replace(content, "export default app;", swaggerMount+"\nexport default app;", 1)
			} else {
				content += swaggerMount
			}
			ctx.State.WriteFile("src/app.ts", []byte(content), state.ContentRaw)
		}
	}

	return nil
}
