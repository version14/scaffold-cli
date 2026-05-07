package decoratorshexagonaladapter

import (
	"embed"

	"github.com/version14/dot/internal/render"
	"github.com/version14/dot/pkg/dotapi"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string    { return Manifest.Name }
func (g *Generator) Version() string { return Manifest.Version }

//go:embed all:files
var fs embed.FS

func (g *Generator) Generate(ctx *dotapi.Context) error {
	return render.NewLocalFolderRenderer(ctx.State).Render(fs, nil)
}
