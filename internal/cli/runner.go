package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/version14/dot/flows"
	"github.com/version14/dot/internal/commands"
	"github.com/version14/dot/internal/dotdir"
	"github.com/version14/dot/internal/flow"
	"github.com/version14/dot/internal/generator"
	"github.com/version14/dot/internal/plugin"
	"github.com/version14/dot/internal/spec"
	"github.com/version14/dot/internal/state"
	"github.com/version14/dot/pkg/dotapi"
)

// ScaffoldOptions configures one end-to-end scaffold run.
type ScaffoldOptions struct {
	Flow        *flows.FlowDef
	Registry    *generator.Registry
	OutputDir   string                 // parent directory; the project goes inside
	ToolVersion string                 // recorded in spec.Metadata.ToolVersion
	Logger      dotapi.Logger          // step-by-step log sink
	Runner      flow.FlowRunner        // form runner (defaults to HuhFormRunner)
	PreAnswers  map[string]flow.Answer // optional pre-filled answers (re-run path)

	// Hooks lets the caller inject plugin contributions (Replace/AddOption/
	// InsertAfter) into the form runner. When non-nil and Runner is nil,
	// Scaffold builds a HuhFormRunner with these hooks attached.
	Hooks *flow.HookRegistry

	// Fragments mirrors Hooks for fragment resolvers (also injected into the
	// default HuhFormRunner when Runner is nil).
	Fragments *flow.FragmentRegistry

	// Plugins are the active providers; Scaffold calls ResolveExtras on each
	// after the flow's resolver runs and appends the results to the
	// invocation set before topo-sorting.
	Plugins []plugin.Provider
}

// ScaffoldResult is what callers (CLI / test-flow) get back after Scaffold.
type ScaffoldResult struct {
	Spec        *spec.ProjectSpec
	State       *state.VirtualProjectState
	ProjectRoot string
	Invocations []generator.Invocation
	Manifests   []dotapi.Manifest // ordered, one per invocation
	Duration    time.Duration
}

// Scaffold runs the full pipeline:
//
//  1. Drive the flow runner against opts.Flow.Root → FlowContext
//  2. Build a ProjectSpec from the FlowContext
//  3. Resolve Invocations via opts.Flow.Generators
//  4. Execute generators against a fresh VirtualProjectState
//  5. Persist the virtual state to opts.OutputDir/<project>/
//  6. Write .dot/spec.json + manifest + .gitignore
//
// Post-generation commands are NOT executed here — callers run them via
// commands.Plan + commands.Runner so test-flow can interleave per-step logs.
func Scaffold(ctx context.Context, opts ScaffoldOptions) (*ScaffoldResult, error) {
	_ = ctx
	if opts.Flow == nil {
		return nil, fmt.Errorf("cli: scaffold: nil flow")
	}
	if opts.Registry == nil {
		return nil, fmt.Errorf("cli: scaffold: nil registry")
	}
	if opts.Logger == nil {
		opts.Logger = dotapi.DiscardLogger{}
	}
	if opts.Runner == nil {
		fr := NewHuhFormRunner()
		if opts.Hooks != nil {
			fr.Hooks = opts.Hooks
		}
		if opts.Fragments != nil {
			fr.Fragments = opts.Fragments
		}
		opts.Runner = fr
	}

	start := time.Now()

	opts.Logger.Infof("→ flow: %s", opts.Flow.Title)
	flowCtx, err := opts.Runner.Run(opts.Flow.Root)
	if err != nil {
		return nil, fmt.Errorf("cli: flow: %w", err)
	}

	for k, v := range opts.PreAnswers {
		if _, exists := flowCtx.Answers[k]; !exists {
			flowCtx.Answers[k] = v
		}
	}

	projectName, _ := flowCtx.Answers["project_name"].(string)
	if projectName == "" {
		projectName = "project"
	}
	s := spec.Build(flowCtx, spec.BuildOpts{
		FlowID:      opts.Flow.ID,
		ProjectName: projectName,
		ToolVersion: opts.ToolVersion,
	})

	if opts.Flow.Generators == nil {
		return nil, fmt.Errorf("cli: flow %q has no Generators resolver", opts.Flow.ID)
	}
	flowInvs := opts.Flow.Generators(s)

	// Convert to generator.Invocation, then append plugin extras based on
	// plugin-contributed answers, then expand transitive deps + topo-sort.
	requested := make([]generator.Invocation, 0, len(flowInvs))
	for _, fi := range flowInvs {
		requested = append(requested, generator.Invocation{Name: fi.Name, LoopStack: fi.LoopStack})
	}
	for _, p := range opts.Plugins {
		requested = append(requested, p.ResolveExtras(s)...)
	}
	invs, err := generator.ResolveInvocations(requested, opts.Registry)
	if err != nil {
		return nil, fmt.Errorf("cli: resolve generators: %w", err)
	}

	mans := make([]dotapi.Manifest, len(invs))
	for i, inv := range invs {
		entry, ok := opts.Registry.Get(inv.Name)
		if !ok {
			return nil, fmt.Errorf("cli: unknown generator %q after resolve", inv.Name)
		}
		mans[i] = entry.Manifest
		if len(inv.LoopStack) > 0 {
			scoped := flow.FlattenScope(s.Answers, inv.LoopStack)
			if appName, ok2 := scoped["app-name"].(string); ok2 && appName != "" {
				mans[i].PathPrefix = "apps/" + appName
			}
		}
	}

	vstate := state.NewVirtualProjectState(s.Metadata)
	exec := generator.NewExecutor(opts.Registry, opts.Logger)

	opts.Logger.Infof("→ executing %d generators", len(invs))
	if err := exec.Execute(invs, s, vstate); err != nil {
		return nil, fmt.Errorf("cli: execute: %w", err)
	}

	root := filepath.Join(opts.OutputDir, projectName)
	count, err := state.Persist(vstate, root)
	if err != nil {
		return nil, fmt.Errorf("cli: persist: %w", err)
	}
	opts.Logger.Infof("→ wrote %d files to %s", count, root)

	if err := dotdir.SaveSpec(root, s); err != nil {
		return nil, fmt.Errorf("cli: save spec: %w", err)
	}
	if err := dotdir.SaveManifest(root, manifestSummary(invs, mans, opts.ToolVersion, time.Since(start))); err != nil {
		return nil, fmt.Errorf("cli: save manifest: %w", err)
	}
	if err := dotdir.WriteIgnore(root); err != nil {
		return nil, fmt.Errorf("cli: write .dot/.gitignore: %w", err)
	}

	return &ScaffoldResult{
		Spec:        s,
		State:       vstate,
		ProjectRoot: root,
		Invocations: invs,
		Manifests:   mans,
		Duration:    time.Since(start),
	}, nil
}

// PlanPostGenCommands turns the manifest+answers into a deduplicated, ordered
// command plan ready for the commands.Runner.
func PlanPostGenCommands(s *spec.ProjectSpec, manifests []dotapi.Manifest) []commands.PlannedCommand {
	invs := make([]commands.Invocation, len(manifests))
	for i, m := range manifests {
		invs[i] = commands.Invocation{Manifest: m, Answers: s.Answers}
	}
	return commands.Plan(invs)
}

// PlanTestCommands builds a command plan from manifests' TestCommands rather
// than PostGenerationCommands. Used by test-flow.
func PlanTestCommands(s *spec.ProjectSpec, manifests []dotapi.Manifest) []commands.PlannedCommand {
	all := make([]commands.PlannedCommand, 0)
	for _, m := range manifests {
		for _, c := range m.TestCommands {
			workDir := interpolateAnswers(c.WorkDir, s.Answers)
			if workDir == "" && m.PathPrefix != "" {
				workDir = m.PathPrefix
			}
			all = append(all, commands.PlannedCommand{
				Cmd:        interpolateAnswers(c.Cmd, s.Answers),
				WorkDir:    workDir,
				Source:     m.Name,
				Background: c.Background,
				ReadyDelay: c.ReadyDelay,
			})
		}
	}
	return commands.Dedup(all)
}

// manifestSummary builds a dotdir.Manifest from the invocations.
func manifestSummary(
	invs []generator.Invocation,
	mans []dotapi.Manifest,
	toolVersion string,
	elapsed time.Duration,
) *dotdir.Manifest {
	now := time.Now().UTC()
	gens := make([]dotdir.ExecutedGenerator, len(invs))
	for i, inv := range invs {
		gens[i] = dotdir.ExecutedGenerator{
			Name:            inv.Name,
			ResolvedVersion: mans[i].Version,
			ExecutedAt:      now,
			InvocationCount: 1,
		}
	}
	return &dotdir.Manifest{
		ToolVersion:        toolVersion,
		LastExecutedAt:     now,
		ExecutionTimeMs:    elapsed.Milliseconds(),
		GeneratorsExecuted: gens,
	}
}

func interpolateAnswers(s string, answers map[string]interface{}) string {
	if s == "" || !strings.ContainsRune(s, '{') {
		return s
	}
	out := s
	for k, v := range answers {
		out = strings.ReplaceAll(out, "{"+k+"}", fmt.Sprint(v))
	}
	return out
}
