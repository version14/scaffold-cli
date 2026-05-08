package expressserverentrypoint

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

	// Base .env.example — downstream generators append their own vars.
	// CORS_ORIGIN is consumed by src/shared/cors.ts: "*" allows any origin
	// (dev only), a comma-separated list restricts to those origins, and
	// unset falls back to http://localhost:3000.
	envExample := "PORT=3000\n" +
		"# Comma-separated list of allowed origins, or \"*\" for any.\n" +
		"CORS_ORIGIN=http://localhost:3000\n"
	ctx.State.WriteFile(".env.example", []byte(envExample), state.ContentRaw)
	return nil
}
