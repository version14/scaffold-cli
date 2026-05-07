package decoratorscleanarchadapter

import "github.com/version14/dot/pkg/dotapi"

var Manifest = dotapi.Manifest{
	Name:        "decorators_clean_arch_adapter",
	Version:     "0.1.0",
	Description: "Wires the decorator router into a Clean Architecture project: sample controller in application layer, schemas, OpenAPI mount in app.ts",
	DependsOn: []string{
		"backend_architecture_clean_architecture",
		"express_decorators_core",
		"express_openapi_setup",
	},
	Outputs: []string{
		"src/app.ts",
		"src/modules/example/application/controllers/example.controller.ts",
		"src/modules/example/application/validators/example.schemas.ts",
		"src/__tests__/decorators-clean.e2e.test.ts",
	},
	Validators: []dotapi.Validator{
		{
			Name: "decorators-clean-arch-adapter",
			Checks: []dotapi.Check{
				{Type: dotapi.CheckFileExists, Path: "src/modules/example/application/controllers/example.controller.ts"},
				{Type: dotapi.CheckFileExists, Path: "src/modules/example/application/validators/example.schemas.ts"},
				{Type: dotapi.CheckFileExists, Path: "src/__tests__/decorators-clean.e2e.test.ts"},
			},
		},
	},
}
