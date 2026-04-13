// Package gogen contains official Go generators for dot.
package gogen

import (
	"fmt"

	"github.com/version14/dot/internal/generator"
	"github.com/version14/dot/internal/spec"
)

// GoRestAPIGenerator scaffolds a minimal Go REST API project.
// It handles the "rest-api" module for language "go".
type GoRestAPIGenerator struct{}

func (g *GoRestAPIGenerator) Name() string     { return "go-rest-api" }
func (g *GoRestAPIGenerator) Language() string { return "go" }
func (g *GoRestAPIGenerator) Modules() []string {
	return []string{"rest-api"}
}

func (g *GoRestAPIGenerator) Apply(s spec.Spec) ([]generator.FileOp, error) {
	return []generator.FileOp{
		{
			Kind:      generator.Create,
			Path:      "main.go",
			Generator: g.Name(),
			Priority:  0,
			Content: fmt.Sprintf(`package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Println("starting %s on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
`, s.Project.Name),
		},
		{
			Kind:      generator.Create,
			Path:      "go.mod",
			Generator: g.Name(),
			Priority:  0,
			Content:   fmt.Sprintf("module github.com/%s\n\ngo 1.26\n", s.Project.Name),
		},
		{
			Kind:      generator.Create,
			Path:      "routes/routes.go",
			Generator: g.Name(),
			Priority:  0,
			Content: `package routes

import "net/http"

// Register mounts all application routes onto mux.
func Register(mux *http.ServeMux) {
	// TODO: register your routes here
}
`,
		},
	}, nil
}

func (g *GoRestAPIGenerator) Commands() []generator.CommandDef {
	return []generator.CommandDef{
		{
			Name:        "new route",
			Args:        []string{"<name>"},
			Description: "generate a new HTTP route handler",
			Action:      "rest-api.new-route",
			Generator:   g.Name(),
		},
		{
			Name:        "new handler",
			Args:        []string{"<name>"},
			Description: "generate a new handler stub",
			Action:      "rest-api.new-handler",
			Generator:   g.Name(),
		},
	}
}

func (g *GoRestAPIGenerator) RunAction(action string, args []string, s spec.Spec) ([]generator.FileOp, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("name argument required")
	}
	name := args[0]

	switch action {
	case "rest-api.new-route":
		return g.newRoute(name, s)
	case "rest-api.new-handler":
		return g.newHandler(name, s)
	default:
		return nil, fmt.Errorf("unknown action %q", action)
	}
}

func (g *GoRestAPIGenerator) newRoute(name string, _ spec.Spec) ([]generator.FileOp, error) {
	return []generator.FileOp{
		{
			Kind:      generator.Create,
			Path:      fmt.Sprintf("routes/%s.go", name),
			Generator: g.Name(),
			Priority:  0,
			Content: fmt.Sprintf(`package routes

import "net/http"

// %sHandler handles requests to /%s.
func %sHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: implement %s handler
	w.WriteHeader(http.StatusOK)
}
`, name, name, name, name),
		},
	}, nil
}

func (g *GoRestAPIGenerator) newHandler(name string, _ spec.Spec) ([]generator.FileOp, error) {
	return []generator.FileOp{
		{
			Kind:      generator.Create,
			Path:      fmt.Sprintf("handlers/%s.go", name),
			Generator: g.Name(),
			Priority:  0,
			Content: fmt.Sprintf(`package handlers

import "net/http"

// %s handles its domain logic.
type %s struct{}

// ServeHTTP implements http.Handler.
func (h *%s) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// TODO: implement %s
	w.WriteHeader(http.StatusOK)
}
`, name, name, name, name),
		},
	}, nil
}
