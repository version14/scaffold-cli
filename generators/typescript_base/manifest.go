package typescriptbase

import "github.com/version14/dot/pkg/dotapi"

const packageJSONFileName = "package.json"

// Manifest declares typescript_base. It runs after base_project and writes
// the package.json + tsconfig.json that any TypeScript project needs.
var Manifest = dotapi.Manifest{
	Name:        "typescript_base",
	Version:     "0.1.0",
	Description: "TypeScript foundation: package.json + tsconfig.json",
	DependsOn:   []string{"base_project"},
	Outputs: []string{
		packageJSONFileName,
		"tsconfig.json",
	},
	PostGenerationCommands: []dotapi.Command{
		{Cmd: "pnpm install"},
	},
	TestCommands: []dotapi.Command{
		{Cmd: "pnpm install --dangerously-allow-all-builds"},
		{Cmd: "pnpm exec tsc --noEmit"},
	},
	Validators: []dotapi.Validator{
		{
			Name: "ts-foundations",
			Checks: []dotapi.Check{
				{Type: dotapi.CheckFileExists, Path: packageJSONFileName},
				{Type: dotapi.CheckFileExists, Path: "tsconfig.json"},
				{Type: dotapi.CheckJSONKeyExists, Path: packageJSONFileName, Key: "devDependencies.typescript"},
			},
		},
	},
}
