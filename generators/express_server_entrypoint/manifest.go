package expressserverentrypoint

import "github.com/version14/dot/pkg/dotapi"

var Manifest = dotapi.Manifest{
	Name:        "express_server_entrypoint",
	Version:     "0.1.0",
	Description: "Express server TypeScript source files: src/index.ts (bootstrap) and src/app.ts (/health route)",
	DependsOn:   []string{"typescript_base"},
	Outputs: []string{
		"src/index.ts",
		"src/app.ts",
		"src/shared/cors.ts",
		".env.example",
	},
	Validators: []dotapi.Validator{
		{
			Name: "express-entrypoint",
			Checks: []dotapi.Check{
				{Type: dotapi.CheckFileExists, Path: "src/index.ts"},
				{Type: dotapi.CheckFileExists, Path: "src/app.ts"},
				{Type: dotapi.CheckFileExists, Path: "src/shared/cors.ts"},
			},
		},
	},
}
