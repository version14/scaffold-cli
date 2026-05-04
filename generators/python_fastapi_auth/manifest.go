package pythonfastapiauth

import "github.com/version14/dot/pkg/dotapi"

var Manifest = dotapi.Manifest{
	Name:        "python_fastapi_auth",
	Version:     "0.1.0",
	Description: "Adds FastAPI auth routes /register and /login (always return 200)",
	DependsOn:   []string{"python_fastapi_base"},
	Outputs:     []string{"routers/auth.py"},
	Validators: []dotapi.Validator{
		{
			Name: "python_fastapi_auth",
			Checks: []dotapi.Check{
				{Type: dotapi.CheckFileExists, Path: "routers/auth.py"},
			},
		},
	},
}
