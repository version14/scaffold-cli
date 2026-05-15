package typescriptbase

import (
	"github.com/version14/dot/internal/state"
	"github.com/version14/dot/pkg/dotapi"
)

// Generator writes a minimal package.json and tsconfig.json. Other generators
// (react_app, biome_config) merge into the same files via UpdateJSON.
type Generator struct{}

func New() *Generator { return &Generator{} }

func (g *Generator) Name() string    { return Manifest.Name }
func (g *Generator) Version() string { return Manifest.Version }

func (g *Generator) Generate(ctx *dotapi.Context) error {
	// In multi-app context app-name scopes the package; fall back to project_name for single-app.
	projectName := stringAnswer(ctx.Answers, "app-name", "")
	if projectName == "" {
		projectName = stringAnswer(ctx.Answers, "project_name", ctx.Spec.Metadata.ProjectName)
	}
	if projectName == "" {
		projectName = "app"
	}

	if err := ctx.State.UpdateJSON("package.json", func(d *state.JSONDoc) error {
		d.Merge(map[string]interface{}{
			"name":    projectName,
			"version": "0.1.0",
			"private": true,
			"type":    "module",
			"scripts": map[string]interface{}{
				"build": "tsc",
			},
			"devDependencies": map[string]interface{}{
				"typescript": "^5.4.0",
			},
		})
		return nil
	}); err != nil {
		return err
	}

	return ctx.State.UpdateJSON("tsconfig.json", func(d *state.JSONDoc) error {
		if _, ok := d.GetNested("compilerOptions"); !ok {
			d.Merge(map[string]interface{}{
				"compilerOptions": map[string]interface{}{
					"target":           "ES2022",
					"module":           "ESNext",
					"moduleResolution": "Bundler",
					"strict":           true,
					"esModuleInterop":  true,
					"skipLibCheck":     true,
					"outDir":           "dist",
				},
				"include": []interface{}{"src"},
			})
		}
		return nil
	})
}

// stringAnswer returns answers[key] as a string, falling back to fallback if
// missing or non-string.
func stringAnswer(answers map[string]interface{}, key, fallback string) string {
	if v, ok := answers[key]; ok {
		if s, ok := v.(string); ok && s != "" {
			return s
		}
	}
	return fallback
}
