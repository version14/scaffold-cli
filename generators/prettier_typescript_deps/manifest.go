package prettiertypescriptdeps

import "github.com/version14/dot/pkg/dotapi"

var Manifest = dotapi.Manifest{
	Name:        "prettier_typescript_deps",
	Version:     "0.1.0",
	Description: "Adds prettier devDependency and format script to package.json (TypeScript projects)",
	DependsOn:   []string{"prettier_config", "*"},
	Outputs:     []string{},
	PostGenerationCommands: []dotapi.Command{
		{Cmd: "pnpm install --dangerously-allow-all-builds"},
		{Cmd: "pnpm format"},
	},
	TestCommands: []dotapi.Command{
		{Cmd: "pnpm format:check"},
	},
	Validators: []dotapi.Validator{
		{
			Name: "prettier-typescript-deps",
			Checks: []dotapi.Check{
				{Type: dotapi.CheckJSONKeyExists, Path: "package.json", Key: "devDependencies.prettier"},
				{Type: dotapi.CheckJSONKeyExists, Path: "package.json", Key: "scripts.format"},
			},
		},
	},
}
