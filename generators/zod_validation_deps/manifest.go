package zodvalidationdeps

import "github.com/version14/dot/pkg/dotapi"

var packageFileName = "package.json"

var Manifest = dotapi.Manifest{
	Name:        "zod_validation_deps",
	Version:     "0.1.0",
	Description: "Adds Zod, zod-to-openapi, swagger-ui-express, and reflect-metadata dependencies plus tsconfig flags for decorator metadata",
	DependsOn:   []string{"typescript_base"},
	Outputs:     []string{},
	Validators: []dotapi.Validator{
		{
			Name: "zod-validation-deps",
			Checks: []dotapi.Check{
				{Type: dotapi.CheckJSONKeyExists, Path: packageFileName, Key: "dependencies.zod"},
				{Type: dotapi.CheckJSONKeyExists, Path: packageFileName, Key: "dependencies.@asteasolutions/zod-to-openapi"},
				{Type: dotapi.CheckJSONKeyExists, Path: packageFileName, Key: "dependencies.reflect-metadata"},
				{Type: dotapi.CheckJSONKeyExists, Path: packageFileName, Key: "dependencies.swagger-ui-express"},
			},
		},
	},
}
