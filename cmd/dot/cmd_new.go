package main

import (
	"fmt"
	"strings"

	"github.com/version14/dot/internal/pipeline"
	"github.com/version14/dot/internal/project"
)

// cmdNew dispatches dot new <type> <name> to the correct generator action.
func cmdNew(artifactType, artifactName string, extraArgs []string) error {
	ctx, err := project.Load(".")
	if err != nil {
		return err
	}

	key := "new " + artifactType
	ref, ok := ctx.Commands[key]
	if !ok {
		available := commandsWithVerb(ctx.Commands, "new")
		if len(available) == 0 {
			return fmt.Errorf("no 'new' commands in this project — run 'dot init' first")
		}
		return fmt.Errorf("unknown type %q — available: %s",
			artifactType, strings.Join(available, ", "))
	}

	reg := buildRegistry()
	g, ok := reg.Get(ref.Generator)
	if !ok {
		return fmt.Errorf("generator %q not found (was it removed after dot init?)", ref.Generator)
	}

	allArgs := append([]string{artifactName}, extraArgs...)
	ops, err := g.RunAction(ref.Action, allArgs, ctx.Spec)
	if err != nil {
		return fmt.Errorf("generator %q: %w", ref.Generator, err)
	}

	if err := pipeline.Run(ops); err != nil {
		return fmt.Errorf("pipeline: %w", err)
	}

	fmt.Printf("%s  %s %q\n", successStyle.Render("✓"), artifactType, artifactName)
	return nil
}

// commandsWithVerb returns the noun part of all commands matching the given verb.
func commandsWithVerb(commands map[string]project.CommandRef, verb string) []string {
	prefix := verb + " "
	var nouns []string
	for k := range commands {
		if strings.HasPrefix(k, prefix) {
			nouns = append(nouns, strings.TrimPrefix(k, prefix))
		}
	}
	return nouns
}
