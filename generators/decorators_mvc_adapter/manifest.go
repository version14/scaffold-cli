package decoratorsmvcadapter

import "github.com/version14/dot/pkg/dotapi"

var Manifest = dotapi.Manifest{
	Name:        "decorators_mvc_adapter",
	Version:     "0.1.0",
	Description: "Wires the decorator router into an MVC project: sample controller in src/controllers, schemas in src/shared/validators, OpenAPI mount in app.ts",
	DependsOn: []string{
		"backend_architecture_mvc",
		"express_decorators_core",
		"express_openapi_setup",
	},
	Outputs: []string{
		"src/app.ts",
		"src/controllers/example.controller.ts",
		"src/shared/validators/example.schemas.ts",
		"src/__tests__/decorators-mvc.e2e.test.ts",
	},
	Validators: []dotapi.Validator{
		{
			Name: "decorators-mvc-adapter",
			Checks: []dotapi.Check{
				{Type: dotapi.CheckFileExists, Path: "src/controllers/example.controller.ts"},
				{Type: dotapi.CheckFileExists, Path: "src/shared/validators/example.schemas.ts"},
				{Type: dotapi.CheckFileExists, Path: "src/__tests__/decorators-mvc.e2e.test.ts"},
			},
		},
	},
}
