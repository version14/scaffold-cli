package authjwtcleanarchmodule

import (
	"embed"
	"slices"
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

const authRouteImport = "import authRouter from './routes/auth.route';\n"
const authRouteUse = "app.use('/auth', authRouter);\n"

const authControllerImport = "import { AuthController } from './modules/auth/application/controllers/auth.controller';\n"

func (g *Generator) Generate(ctx *dotapi.Context) error {
	hasDecorators := slices.Contains(ctx.PreviousGens, "express_decorators_core")

	renderer := render.NewLocalFolderRenderer(ctx.State)
	if err := renderer.Render(fs, struct{ HasDecorators bool }{HasDecorators: hasDecorators}); err != nil {
		return err
	}

	appPath := "src/app.ts"

	if f, ok := ctx.State.GetFile(appPath); ok {
		content := string(f.Content)
		if hasDecorators {
			if !strings.Contains(content, "AuthController") {
				content = authControllerImport + content
				content = strings.Replace(
					content,
					".registerController(new ExampleController());",
					".registerController(new ExampleController())\n  .registerController(new AuthController());",
					1,
				)
				ctx.State.WriteFile(appPath, []byte(content), state.ContentRaw)
			}
		} else if !strings.Contains(content, "authRouter") {
			content = authRouteImport + content
			if strings.Contains(content, "app.use(errorMiddleware)") {
				content = strings.Replace(content, "app.use(errorMiddleware)", authRouteUse+"\napp.use(errorMiddleware)", 1)
			} else {
				content = strings.Replace(content, "export default app;", authRouteUse+"\nexport default app;", 1)
			}
			ctx.State.WriteFile(appPath, []byte(content), state.ContentRaw)
		}
	}

	return ctx.State.UpdateJSON("package.json", func(d *state.JSONDoc) error {
		d.Merge(map[string]interface{}{
			"dependencies": map[string]interface{}{
				"bcryptjs": "^2.4.3",
			},
			"devDependencies": map[string]interface{}{
				"@types/bcryptjs": "^2.4.6",
			},
		})
		return nil
	})
}
