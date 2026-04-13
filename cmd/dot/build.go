package main

import (
	"fmt"

	gogen "github.com/version14/dot/generators/go"
	"github.com/version14/dot/internal/generator"
)

// buildVersion is set at build time via:
//
//	go build -ldflags "-X main.buildVersion=v0.1.0" ./cmd/dot
//
// Falls back to "dev" for local builds.
var buildVersion = "dev"

// buildRegistry creates and populates the generator registry with all
// official generators. This is the single place where generators are
// registered — add new ones here.
func buildRegistry() *generator.Registry {
	reg := &generator.Registry{}

	must(reg.Register(&gogen.GoRestAPIGenerator{}))

	return reg
}

// must panics if err is non-nil. Used only for generator registration errors
// at startup — a registration conflict is a programming error, not a runtime
// error, and should surface immediately.
func must(err error) {
	if err != nil {
		panic(fmt.Sprintf("generator registration failed: %v", err))
	}
}
