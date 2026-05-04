// Package pythonfastapibase scaffolds a base FastAPI Python project with a health endpoint.
package pythonfastapibase

import (
	"fmt"

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
		return fmt.Errorf("python_fastapi_base: missing project_name")
	}

	mainPy := `from fastapi import FastAPI

app = FastAPI(title="` + projectName + `")


@app.get("/health")
def health():
    return {"status": "ok"}
`
	ctx.State.WriteFile("main.py", []byte(mainPy), state.ContentRaw)
	ctx.State.WriteFile("requirements.txt", []byte("fastapi>=0.110.0\nuvicorn[standard]>=0.29.0\n"), state.ContentRaw)

	return nil
}
