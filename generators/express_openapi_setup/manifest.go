package expressopenapisetup

import "github.com/version14/dot/pkg/dotapi"

var Manifest = dotapi.Manifest{
	Name:        "express_openapi_setup",
	Version:     "0.1.0",
	Description: "OpenAPI spec generator + Swagger UI mount that consumes DecoratorRouter metadata",
	DependsOn:   []string{"express_decorators_core"},
	Outputs: []string{
		"src/shared/openapi/registry.ts",
		"src/shared/openapi/spec-generator.ts",
		"src/shared/openapi/swagger.ts",
		"src/shared/openapi/index.ts",
		"src/shared/openapi/__tests__/spec.unit.test.ts",
	},
	Validators: []dotapi.Validator{
		{
			Name: "express-openapi-setup",
			Checks: []dotapi.Check{
				{Type: dotapi.CheckFileExists, Path: "src/shared/openapi/spec-generator.ts"},
				{Type: dotapi.CheckFileExists, Path: "src/shared/openapi/swagger.ts"},
				{Type: dotapi.CheckFileExists, Path: "src/shared/openapi/registry.ts"},
			},
		},
	},
}
