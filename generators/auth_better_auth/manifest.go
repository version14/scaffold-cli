package authbetterauth

import "github.com/version14/dot/pkg/dotapi"

var Manifest = dotapi.Manifest{
	Name:        "auth_better_auth",
	Version:     "0.2.0",
	Description: "BetterAuth setup with Drizzle adapter: src/lib/auth.ts plus a direct toNodeHandler mount in src/app.ts (no intermediate route file)",
	DependsOn:   []string{"drizzle_postgres_adapter"},
	Outputs: []string{
		"src/lib/auth.ts",
	},
	Validators: []dotapi.Validator{
		{
			Name: "auth-better-auth",
			Checks: []dotapi.Check{
				{Type: dotapi.CheckFileExists, Path: "src/lib/auth.ts"},
				{Type: dotapi.CheckJSONKeyExists, Path: "package.json", Key: "dependencies.better-auth"},
			},
		},
	},
}
