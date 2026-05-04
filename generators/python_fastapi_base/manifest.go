package pythonfastapibase

import "github.com/version14/dot/pkg/dotapi"

var Manifest = dotapi.Manifest{
	Name:        "python_fastapi_base",
	Version:     "0.1.0",
	Description: "Scaffolds a base FastAPI Python project with a health endpoint",
	DependsOn:   []string{"base_project"},
	Outputs:     []string{"main.py", "requirements.txt"},
	Validators: []dotapi.Validator{
		{
			Name: "python_fastapi_base",
			Checks: []dotapi.Check{
				{Type: dotapi.CheckFileExists, Path: "main.py"},
				{Type: dotapi.CheckFileExists, Path: "requirements.txt"},
			},
		},
	},
}
