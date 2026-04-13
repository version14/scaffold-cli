package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/version14/dot/internal/generator"
	"github.com/version14/dot/internal/project"
)

// cmdHelp reads .dot/config.json and prints all available commands for the
// current project, styled with lipgloss.
func cmdHelp() error {
	ctx, err := project.Load(".")
	if err != nil {
		return err
	}

	if len(ctx.Commands) == 0 {
		fmt.Println(mutedStyle.Render("no commands registered in this project"))
		return nil
	}

	// Sort commands alphabetically for stable output.
	keys := make([]string, 0, len(ctx.Commands))
	for k := range ctx.Commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build a table: two columns, command name and description.
	// We look up the description from the live registry (it's not persisted).
	reg := buildRegistry()
	descMap := buildDescMap(reg, ctx)

	// Measure the longest command name for alignment.
	maxLen := 0
	for _, k := range keys {
		if len(k) > maxLen {
			maxLen = len(k)
		}
	}

	fmt.Println()
	fmt.Println(headerStyle.Render("  " + ctx.Spec.Project.Name + " — available commands"))

	rows := make([]string, 0, len(keys)+2)
	for _, k := range keys {
		padded := k + strings.Repeat(" ", maxLen-len(k)+2)
		desc := descMap[k]
		row := "  dot " +
			commandNameStyle.Render(padded) +
			commandDescStyle.Render(desc)
		rows = append(rows, row)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, rows...)
	fmt.Println(boxStyle.Render(content))
	fmt.Println()
	return nil
}

// buildDescMap looks up CommandDef.Description for each persisted command key.
func buildDescMap(reg *generator.Registry, ctx *project.Context) map[string]string {
	m := make(map[string]string, len(ctx.Commands))
	for k, ref := range ctx.Commands {
		g, ok := reg.Get(ref.Generator)
		if !ok {
			continue
		}
		for _, cmd := range g.Commands() {
			if cmd.Name == k {
				m[k] = cmd.Description
				break
			}
		}
	}
	return m
}
