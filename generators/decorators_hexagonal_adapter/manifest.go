package decoratorshexagonaladapter

import "github.com/version14/dot/pkg/dotapi"

var Manifest = dotapi.Manifest{
	Name:        "decorators_hexagonal_adapter",
	Version:     "0.1.0",
	Description: "Wires the decorator router into a Hexagonal project: sample primary HTTP adapter controller, schemas, OpenAPI mount in app.ts",
	DependsOn: []string{
		"backend_architecture_hexagonal",
		"express_decorators_core",
		"express_openapi_setup",
	},
	Outputs: []string{
		"src/app.ts",
		"src/adapters/primary/http/controllers/example.controller.ts",
		"src/adapters/primary/http/schemas/example.schemas.ts",
		"src/__tests__/decorators-hexagonal.e2e.test.ts",
	},
	Validators: []dotapi.Validator{
		{
			Name: "decorators-hexagonal-adapter",
			Checks: []dotapi.Check{
				{Type: dotapi.CheckFileExists, Path: "src/adapters/primary/http/controllers/example.controller.ts"},
				{Type: dotapi.CheckFileExists, Path: "src/adapters/primary/http/schemas/example.schemas.ts"},
				{Type: dotapi.CheckFileExists, Path: "src/__tests__/decorators-hexagonal.e2e.test.ts"},
			},
		},
	},
}
