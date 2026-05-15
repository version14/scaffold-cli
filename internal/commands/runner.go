package commands

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/version14/dot/pkg/dotapi"
)

// getEnviron returns the current environment with the PWD variable removed,
// as it can interfere with command execution in unexpected ways.
func getEnviron() []string {
	out := make([]string, 0, len(os.Environ())-1)
	for _, v := range os.Environ() {
		if strings.HasPrefix(v, "PWD=") {
			continue
		}
		out = append(out, v)
	}
	return out
}

// defaultBackgroundReadyDelay is used when a background command has no
// explicit ReadyDelay — long enough for most dev servers to bind a port
// without making each test feel slow.
const defaultBackgroundReadyDelay = 3 * time.Second

// Plan resolves manifests + scoped answers into a deduplicated, ordered list
// of PlannedCommands ready for execution.
func Plan(invocations []Invocation) []PlannedCommand {
	all := make([]PlannedCommand, 0)
	for _, inv := range invocations {
		for _, cmd := range inv.Manifest.PostGenerationCommands {
			workDir := interpolate(cmd.WorkDir, inv.Answers)
			// When the manifest carries a PathPrefix (e.g. "apps/api" for a
			// per-app loop invocation) and the command has no explicit WorkDir,
			// run the command inside that app's directory.
			if workDir == "" && inv.Manifest.PathPrefix != "" {
				workDir = inv.Manifest.PathPrefix
			}
			all = append(all, PlannedCommand{
				Cmd:        interpolate(cmd.Cmd, inv.Answers),
				WorkDir:    workDir,
				Source:     inv.Manifest.Name,
				Background: cmd.Background,
				ReadyDelay: cmd.ReadyDelay,
			})
		}
	}
	return Dedup(all)
}

// Invocation pairs a manifest with the scoped answers used to interpolate
// its commands. One entry per generator invocation.
type Invocation struct {
	Manifest dotapi.Manifest
	Answers  map[string]interface{}
}

// Runner executes PlannedCommands sequentially against an on-disk project.
// Each command is run via /bin/sh -c so shell features (pipes, &&) work.
type Runner struct {
	ProjectRoot string
	Logger      dotapi.Logger
	// DryRun, when true, logs commands but does not execute them.
	DryRun bool
}

// NewRunner constructs a Runner anchored at projectRoot.
func NewRunner(projectRoot string, logger dotapi.Logger) *Runner {
	if logger == nil {
		logger = dotapi.DiscardLogger{}
	}
	return &Runner{ProjectRoot: projectRoot, Logger: logger}
}

// Run executes each command in order, streaming their stdout/stderr to the
// process's own stdout/stderr. On the first failure execution halts.
//
// For quiet UX (spinner + capture-on-failure), use RunOneCaptured.
func (r *Runner) Run(ctx context.Context, cmds []PlannedCommand) error {
	for _, c := range cmds {
		if err := r.RunOne(ctx, c); err != nil {
			return fmt.Errorf("commands: %s [%s]: %w", c.Cmd, c.Source, err)
		}
	}
	return nil
}

// RunOne dispatches to runForeground or runBackground based on the command,
// streaming output to os.Stdout/os.Stderr.
func (r *Runner) RunOne(ctx context.Context, c PlannedCommand) error {
	wd := r.workDir(c)
	if r.DryRun {
		r.Logger.Infof("[dry-run] %s (in %s)", c.Cmd, relOrDot(r.ProjectRoot, wd))
		return nil
	}
	if c.Background {
		return r.runBackground(ctx, c, wd, os.Stdout, os.Stderr)
	}
	return r.runForeground(ctx, c, wd, os.Stdout, os.Stderr)
}

// RunOneCaptured runs a command quietly: stdout+stderr are merged into a
// returned byte slice instead of streaming to the terminal. The output is
// returned regardless of success so callers can dump it on failure.
func (r *Runner) RunOneCaptured(ctx context.Context, c PlannedCommand) ([]byte, error) {
	wd := r.workDir(c)
	if r.DryRun {
		return nil, nil
	}

	var buf bytes.Buffer
	var err error
	if c.Background {
		err = r.runBackground(ctx, c, wd, &buf, &buf)
	} else {
		err = r.runForeground(ctx, c, wd, &buf, &buf)
	}
	return buf.Bytes(), err
}

func (r *Runner) workDir(c PlannedCommand) string {
	if c.WorkDir == "" {
		return r.ProjectRoot
	}
	return filepath.Join(r.ProjectRoot, c.WorkDir)
}

// fileOrBuf accepts both *os.File (for live streaming) and *bytes.Buffer (for
// capture). Both implement io.Writer; we use the interface alias to keep the
// runForeground/runBackground signatures honest about what they need.
type fileOrBuf = interface {
	Write(p []byte) (int, error)
}

func relOrDot(root, wd string) string {
	rel, err := filepath.Rel(root, wd)
	if err != nil || rel == "" {
		return "."
	}
	return rel
}

// interpolate substitutes {key}-style tokens in s using values from answers.
// Unknown tokens are left literal.
func interpolate(s string, answers map[string]interface{}) string {
	if s == "" || !strings.ContainsRune(s, '{') {
		return s
	}
	out := s
	for k, v := range answers {
		token := "{" + k + "}"
		if !strings.Contains(out, token) {
			continue
		}
		out = strings.ReplaceAll(out, token, fmt.Sprint(v))
	}
	return out
}
