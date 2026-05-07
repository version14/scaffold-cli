package expressdecoratorscore

import "github.com/version14/dot/pkg/dotapi"

var Manifest = dotapi.Manifest{
	Name:        "express_decorators_core",
	Version:     "0.1.0",
	Description: "Framework-agnostic API decorators (Controller, route, validation, response, auth) with an Express RouterAdapter",
	DependsOn:   []string{"express_server_entrypoint", "zod_validation_deps"},
	Outputs: []string{
		"src/shared/decorators/metadata.ts",
		"src/shared/decorators/controller.decorator.ts",
		"src/shared/decorators/route.decorators.ts",
		"src/shared/decorators/validation.decorators.ts",
		"src/shared/decorators/response.decorator.ts",
		"src/shared/decorators/auth.decorator.ts",
		"src/shared/decorators/header.decorator.ts",
		"src/shared/decorators/router-adapter.ts",
		"src/shared/decorators/express-router-adapter.ts",
		"src/shared/decorators/decorator-router.ts",
		"src/shared/decorators/index.ts",
		"src/shared/middlewares/validate-request.ts",
		"src/shared/decorators/__tests__/decorators.unit.test.ts",
	},
	Validators: []dotapi.Validator{
		{
			Name: "express-decorators-core",
			Checks: []dotapi.Check{
				{Type: dotapi.CheckFileExists, Path: "src/shared/decorators/index.ts"},
				{Type: dotapi.CheckFileExists, Path: "src/shared/decorators/decorator-router.ts"},
				{Type: dotapi.CheckFileExists, Path: "src/shared/decorators/router-adapter.ts"},
				{Type: dotapi.CheckFileExists, Path: "src/shared/middlewares/validate-request.ts"},
			},
		},
	},
}
