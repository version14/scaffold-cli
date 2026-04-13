package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"

	"github.com/version14/dot/internal/generator"
	"github.com/version14/dot/internal/pipeline"
	"github.com/version14/dot/internal/project"
	"github.com/version14/dot/internal/spec"
)

// cmdInit runs dot init: guard → survey → confirm → generate with spinner.
func cmdInit() error {
	// Guard: refuse if this is already a dot project.
	if _, err := os.Stat(".dot/config.json"); err == nil {
		return fmt.Errorf("already a dot project — use 'dot add module' to extend it")
	}

	fmt.Println(headerStyle.Render(dotBanner))
	fmt.Println(subheaderStyle.Render("  universal project companion"))
	fmt.Println()

	s, err := surveySpec()
	if err != nil {
		return err
	}

	// Print spec as JSON so the user can verify before we touch the disk.
	out, _ := json.MarshalIndent(s, "", "  ")
	fmt.Println()
	fmt.Println(mutedStyle.Render("Spec:"))
	fmt.Println(mutedStyle.Render(string(out)))
	fmt.Println()

	// Confirm before generating.
	var confirmed bool
	confirm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("Generate project?").
				Affirmative("Yes, generate").
				Negative("Cancel").
				Value(&confirmed),
		),
	).WithTheme(huh.ThemeDracula())

	if err := confirm.Run(); err != nil {
		return err
	}
	if !confirmed {
		fmt.Println(mutedStyle.Render("cancelled."))
		return nil
	}

	// Run generators inside a bubbletea spinner.
	var ops []generator.FileOp
	var genErr error

	m := newSpinnerModel("Generating project...", func() (string, error) {
		reg := buildRegistry()
		generators := reg.ForSpec(s)
		if len(generators) == 0 {
			return "", fmt.Errorf("no generators found for language=%q modules=%v",
				s.Project.Language, s.ModuleNames())
		}
		for _, g := range generators {
			genOps, err := g.Apply(s)
			if err != nil {
				return "", fmt.Errorf("generator %q: %w", g.Name(), err)
			}
			ops = append(ops, genOps...)
		}
		if err := pipeline.Run(ops); err != nil {
			return "", fmt.Errorf("pipeline: %w", err)
		}
		return fmt.Sprintf("project %q created", s.Project.Name), nil
	})

	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return err
	}
	sm := final.(spinnerModel)
	if sm.err != nil {
		genErr = sm.err
	}

	if genErr != nil {
		fmt.Println(errorStyle.Render("✗ " + genErr.Error()))
		return genErr
	}

	// Write .dot/config.json and .dot/manifest.json.
	reg := buildRegistry()
	manifest, err := project.BuildManifest(ops)
	if err != nil {
		return fmt.Errorf("build manifest: %w", err)
	}
	ctx := &project.Context{
		DotVersion:  buildVersion,
		SpecVersion: 1,
		Spec:        s,
		Commands:    project.CommandsFromDefs(reg.CommandsForSpec(s)),
	}
	if err := project.Save(".", ctx, manifest); err != nil {
		return fmt.Errorf("save context: %w", err)
	}

	fmt.Println()
	fmt.Println(successStyle.Render("✓ " + sm.result))
	fmt.Println(mutedStyle.Render("  run 'dot help' to see available commands"))
	fmt.Println()
	return nil
}

// surveySpec runs the huh form and returns a populated Spec.
func surveySpec() (spec.Spec, error) {
	var (
		name        string
		projectType string
		language    string
		modules     []string
		linter      string
		formatter   string
		ci          string
		deployment  string
	)

	accessible := os.Getenv("ACCESSIBLE") != ""

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Placeholder("my-api").
				Value(&name).
				Validate(func(s string) error {
					if len(s) == 0 {
						return fmt.Errorf("project name is required")
					}
					return nil
				}),

			huh.NewSelect[string]().
				Title("Project type").
				Options(
					huh.NewOption("REST API", "api"),
					huh.NewOption("CLI tool", "cli"),
					huh.NewOption("Library", "library"),
					huh.NewOption("Frontend", "frontend"),
					huh.NewOption("Worker", "worker"),
					huh.NewOption("Monorepo", "monorepo"),
				).
				Value(&projectType),

			huh.NewSelect[string]().
				Title("Language").
				Description("More languages coming in future releases.").
				Options(
					huh.NewOption("Go", "go"),
				).
				Value(&language),
		).Title("Project"),

		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Modules").
				Description("Select the modules to include. Space to toggle.").
				Options(
					huh.NewOption("REST API", "rest-api"),
					huh.NewOption("PostgreSQL", "postgres"),
					huh.NewOption("JWT authentication", "auth-jwt"),
					huh.NewOption("Docker", "docker"),
					huh.NewOption("GitHub Actions CI", "github-actions"),
				).
				Filterable(true).
				Value(&modules),
		).Title("Modules"),

		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Linter").
				Options(
					huh.NewOption("golangci-lint (recommended)", "golangci-lint"),
					huh.NewOption("None", "none"),
				).
				Value(&linter),

			huh.NewSelect[string]().
				Title("Formatter").
				Options(
					huh.NewOption("goimports (recommended)", "goimports"),
					huh.NewOption("gofmt", "gofmt"),
					huh.NewOption("None", "none"),
				).
				Value(&formatter),

			huh.NewSelect[string]().
				Title("CI provider").
				Options(
					huh.NewOption("GitHub Actions", "github-actions"),
					huh.NewOption("None", "none"),
				).
				Value(&ci),

			huh.NewSelect[string]().
				Title("Deployment").
				Options(
					huh.NewOption("Docker", "docker"),
					huh.NewOption("Docker Compose", "docker-compose"),
					huh.NewOption("None", "none"),
				).
				Value(&deployment),
		).Title("Config"),
	).
		WithTheme(huh.ThemeDracula()).
		WithAccessible(accessible)

	if err := form.Run(); err != nil {
		return spec.Spec{}, err
	}

	moduleSpecs := make([]spec.ModuleSpec, len(modules))
	for i, m := range modules {
		moduleSpecs[i] = spec.ModuleSpec{Name: m}
	}

	return spec.Spec{
		Project: spec.ProjectSpec{
			Name:     name,
			Language: language,
			Type:     spec.ProjectType(projectType),
		},
		Modules: moduleSpecs,
		Config: spec.CoreConfig{
			Linter:     linter,
			Formatter:  formatter,
			CI:         ci,
			Deployment: deployment,
		},
	}, nil
}

// --- Spinner model (bubbletea) ---

type spinnerDoneMsg struct {
	result string
	err    error
}

type spinnerModel struct {
	spinner spinner.Model
	label   string
	task    func() (string, error)
	result  string
	err     error
	done    bool
}

func newSpinnerModel(label string, task func() (string, error)) spinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = successStyle
	return spinnerModel{spinner: s, label: label, task: task}
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			// Give the spinner one tick to render before the task starts.
			time.Sleep(50 * time.Millisecond)
			result, err := m.task()
			return spinnerDoneMsg{result: result, err: err}
		},
	)
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case spinnerDoneMsg:
		m.done = true
		m.result = msg.result
		m.err = msg.err
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		if m.err != nil {
			return errorStyle.Render("✗ "+m.err.Error()) + "\n"
		}
		return successStyle.Render("✓ "+m.result) + "\n"
	}
	return "\n  " + m.spinner.View() + " " + m.label + "\n\n"
}
