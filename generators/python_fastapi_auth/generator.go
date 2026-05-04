// Package pythonfastapiauth adds FastAPI auth routes /register and /login (always return 200).
package pythonfastapiauth

import (
	"github.com/version14/dot/internal/state"
	"github.com/version14/dot/pkg/dotapi"
)

type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string    { return Manifest.Name }
func (g *Generator) Version() string { return Manifest.Version }

func (g *Generator) Generate(ctx *dotapi.Context) error {
	ctx.State.WriteFile("routers/auth.py", []byte(`from fastapi import APIRouter

router = APIRouter(prefix="/auth", tags=["auth"])


@router.post("/register", status_code=200)
def register():
    return {"message": "registered"}


@router.post("/login", status_code=200)
def login():
    return {"message": "logged in"}
`), state.ContentRaw)
	return nil
}
