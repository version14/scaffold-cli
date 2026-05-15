package monorepotsworkspaces

import "github.com/version14/dot/pkg/dotapi"

var Manifest = dotapi.Manifest{
	Name:        "monorepo_ts_workspaces",
	Version:     "0.1.0",
	Description: "pnpm workspaces root: package.json + pnpm-workspace.yaml for TypeScript monorepos",
	DependsOn:   []string{"base_project"},
	Outputs: []string{
		"package.json",
		"pnpm-workspace.yaml",
	},
	PostGenerationCommands: []dotapi.Command{
		{Cmd: "pnpm install --dangerously-allow-all-builds"},
	},
	TestCommands: []dotapi.Command{
		{Cmd: "pnpm install --dangerously-allow-all-builds"},
	},
	Validators: []dotapi.Validator{
		{
			Name: "monorepo-root",
			Checks: []dotapi.Check{
				{Type: dotapi.CheckFileExists, Path: "package.json"},
				{Type: dotapi.CheckFileExists, Path: "pnpm-workspace.yaml"},
				{Type: dotapi.CheckJSONKeyExists, Path: "package.json", Key: "workspaces"},
			},
		},
	},
}
