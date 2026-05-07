package authjwtvanilla

import (
	"embed"
	"fmt"
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

func (g *Generator) Generate(ctx *dotapi.Context) error {
	renderer := render.NewLocalFolderRenderer(ctx.State)
	if err := renderer.Render(fs, nil); err != nil {
		return err
	}

	if err := ctx.State.UpdateJSON("package.json", func(d *state.JSONDoc) error {
		d.Merge(map[string]interface{}{
			"dependencies": map[string]interface{}{
				"jsonwebtoken":  "^9.0.2",
				"cookie-parser": "^1.4.7",
			},
			"devDependencies": map[string]interface{}{
				"@types/jsonwebtoken":  "^9.0.7",
				"@types/cookie-parser": "^1.4.8",
			},
		})
		return nil
	}); err != nil {
		return err
	}

	// Append JWT env vars to .env.example
	existing := ""
	if f, ok := ctx.State.GetFile(".env.example"); ok {
		existing = string(f.Content)
	}
	updated := existing + fmt.Sprintf("\n# Auth (JWT)\nJWT_SECRET=%s\nJWT_EXPIRES_IN=7d\nJWT_REFRESH_EXPIRES_IN=30d\n", "change-me-to-a-random-secret")
	ctx.State.WriteFile(".env.example", []byte(updated), state.ContentRaw)

	// Inject cookie-parser middleware into app.ts and, if decorators are
	// active, plug the JWT auth middleware into ExpressRouterAdapter so that
	// every @Auth()-decorated route gets gated automatically.
	hasDecorators := slices.Contains(ctx.PreviousGens, "express_decorators_core")

	if f, ok := ctx.State.GetFile("src/app.ts"); ok {
		content := string(f.Content)
		if !strings.Contains(content, "cookieParser") {
			content = "import cookieParser from 'cookie-parser';\n" + content
			content = strings.Replace(content, "app.use(express.json());", "app.use(express.json());\napp.use(cookieParser());", 1)
		}
		if hasDecorators && !strings.Contains(content, "authMiddleware") {
			content = "import { authMiddleware } from './shared/middlewares/auth.middleware';\n" + content
			content = strings.Replace(content, "new ExpressRouterAdapter()", "new ExpressRouterAdapter({ authMiddleware })", 1)
		}
		ctx.State.WriteFile("src/app.ts", []byte(content), state.ContentRaw)
	}

	return nil
}
