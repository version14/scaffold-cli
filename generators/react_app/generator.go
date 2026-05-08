package reactapp

import (
	"github.com/version14/dot/internal/render"
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
		projectName = ctx.Spec.Metadata.ProjectName
	}
	if projectName == "" {
		projectName = "app"
	}

	// Merge React deps + scripts into package.json.
	if err := ctx.State.UpdateJSON("package.json", func(d *state.JSONDoc) error {
		d.Merge(map[string]interface{}{
			"scripts": map[string]interface{}{
				"dev":     "vite",
				"build":   "tsc && vite build",
				"preview": "vite preview",
			},
			"dependencies": map[string]interface{}{
				"react":     "^18.3.0",
				"react-dom": "^18.3.0",
			},
			"devDependencies": map[string]interface{}{
				"@types/react":         "^18.3.0",
				"@types/react-dom":     "^18.3.0",
				"@vitejs/plugin-react": "^4.3.0",
				"vite":                 "^5.4.0",
			},
		})
		return nil
	}); err != nil {
		return err
	}

	// Patch tsconfig for JSX.
	if err := ctx.State.UpdateJSON("tsconfig.json", func(d *state.JSONDoc) error {
		_ = d.SetNested("compilerOptions.jsx", "react-jsx")
		_ = d.SetNested("compilerOptions.lib", []interface{}{"DOM", "DOM.Iterable", "ES2022"})
		return nil
	}); err != nil {
		return err
	}

	// Source files.
	data := map[string]interface{}{"ProjectName": projectName}

	if out, err := render.Render(indexHTMLTmpl, data); err == nil {
		ctx.State.WriteFile("index.html", out, state.ContentRaw)
	} else {
		return err
	}

	ctx.State.WriteFile("vite.config.ts", []byte(viteConfig), state.ContentRaw)
	ctx.State.WriteFile("src/main.tsx", []byte(mainTSX), state.ContentRaw)

	if out, err := render.Render(appTSX, data); err == nil {
		ctx.State.WriteFile("src/App.tsx", out, state.ContentRaw)
	} else {
		return err
	}

	return nil
}

const indexHTMLTmpl = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>{{.ProjectName}}</title>
  </head>
  <body>
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
`

const viteConfig = `import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
});
`

const mainTSX = `import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
`

const appTSX = `export default function App() {
  return <h1>Hello from {{.ProjectName}}</h1>;
}
`
