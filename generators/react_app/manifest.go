package reactapp

import (
	"time"

	"github.com/version14/dot/pkg/dotapi"
)

// Manifest declares react_app. It depends on typescript_base — every React
// project needs the TypeScript foundation first.
var Manifest = dotapi.Manifest{
	Name:        "react_app",
	Version:     "0.1.0",
	Description: "React + Vite application setup",
	DependsOn:   []string{"typescript_base"},
	Outputs: []string{
		"src/main.tsx",
		"src/App.tsx",
		"index.html",
		"vite.config.ts",
	},
	PostGenerationCommands: []dotapi.Command{
		{Cmd: "pnpm install --dangerously-allow-all-builds"},
	},
	TestCommands: []dotapi.Command{
		{Cmd: "pnpm exec tsc --noEmit"},
		{Cmd: "pnpm exec vite build"},
		// Smoke-start the dev server in background to confirm it boots.
		// NoCache: true — we want a real boot every run to catch
		// port-binding / runtime regressions, not skip on a cache hit.
		{Cmd: "pnpm exec vite", Background: true, ReadyDelay: 4 * time.Second, NoCache: true},
	},
	Validators: []dotapi.Validator{
		{
			Name: "react-app-files",
			Checks: []dotapi.Check{
				{Type: dotapi.CheckFileExists, Path: "src/main.tsx"},
				{Type: dotapi.CheckFileExists, Path: "src/App.tsx"},
				{Type: dotapi.CheckFileExists, Path: "index.html"},
				{Type: dotapi.CheckFileExists, Path: "vite.config.ts"},
				{Type: dotapi.CheckJSONKeyExists, Path: "package.json", Key: "dependencies.react"},
			},
		},
	},
}
