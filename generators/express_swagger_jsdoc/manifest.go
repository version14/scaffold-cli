package expressswaggerjsdoc

import "github.com/version14/dot/pkg/dotapi"

var Manifest = dotapi.Manifest{
	Name:        "express_swagger_jsdoc",
	Version:     "0.1.0",
	Description: "Classic JSDoc-driven Swagger/OpenAPI: scans source files for @openapi comments, builds the spec at boot, and mounts swagger-ui at /docs",
	DependsOn:   []string{"express_server_entrypoint"},
	Outputs: []string{
		"src/shared/swagger/swagger.config.ts",
		"src/shared/swagger/index.ts",
		"src/shared/swagger/__tests__/swagger.unit.test.ts",
	},
	Validators: []dotapi.Validator{
		{
			Name: "express-swagger-jsdoc",
			Checks: []dotapi.Check{
				{Type: dotapi.CheckFileExists, Path: "src/shared/swagger/swagger.config.ts"},
				{Type: dotapi.CheckFileExists, Path: "src/shared/swagger/index.ts"},
				{Type: dotapi.CheckJSONKeyExists, Path: "package.json", Key: "dependencies.swagger-jsdoc"},
				{Type: dotapi.CheckJSONKeyExists, Path: "package.json", Key: "dependencies.swagger-ui-express"},
			},
		},
	},
}
