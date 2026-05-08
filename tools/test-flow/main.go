package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/version14/dot/flows"
	"github.com/version14/dot/internal/cli"

	// Built-in plugins (imported for side-effect — register at init()).
	_ "github.com/version14/dot/plugins/biome_extras"
)

// test-flow drives the full DOT scaffold pipeline against scripted JSON
// fixtures so authors can verify a flow + its generators + their commands
// without running the interactive Huh form.
//
// For each fixture it:
//
//  1. Picks the named flow from the flows registry.
//  2. Replays the recorded answers via a scriptedRunner.
//  3. Scaffolds into a fresh temp dir (full generator pipeline).
//  4. Runs validators against the generated tree.
//  5. Runs PostGenerationCommands (unless skipped).
//  6. Runs TestCommands incl. background dev servers (unless skipped).
//
// Each step is logged hierarchically (case → step → sub-step). Exit code 0
// when every case passes, 1 if any failed, 2 for usage / I/O errors.
//
// Flags:
//
//	-dir         testdata directory (default tools/test-flow/testdata)
//	-tmp         parent dir for per-case scratch (default os.TempDir())
//	-skip-post   skip every PostGenerationCommand
//	-skip-test   skip every TestCommand
//	-only        comma-separated case names to run (matches Name field)
func main() {
	dir := flag.String("dir", "tools/test-flow/testdata", "directory containing TestCase fixtures")
	tmpRoot := flag.String("tmp", "", "parent directory for per-case scratch dirs (default: os temp)")
	skipPost := flag.Bool("skip-post", false, "skip every PostGenerationCommand (e.g. for offline runs)")
	skipTest := flag.Bool("skip-test", false, "skip every TestCommand (faster iteration; default: run them)")
	only := flag.String("only", "", "comma-separated subset of case names to run")
	keep := flag.Bool("keep", false, "do not delete per-case scratch dirs (so you can inspect outputs)")
	noCache := flag.Bool("no-cache", false, "ignore cache hits and re-run every case from scratch (cache entries are still refreshed on success)")
	keepGoing := flag.Bool("keep-going", false, "continue running remaining cases after a failure (default: stop at the first failure)")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	registry := flows.Default()

	cases, err := LoadCases(*dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "test-flow:", err)
		os.Exit(2)
	}
	if len(cases) == 0 {
		fmt.Fprintf(os.Stderr, "test-flow: no .json fixtures found in %s\n", *dir)
		os.Exit(2)
	}

	cases = filterCases(cases, *only)
	if len(cases) == 0 {
		fmt.Fprintln(os.Stderr, "test-flow: no cases match -only filter")
		os.Exit(2)
	}

	rt, err := cli.DefaultRuntime()
	if err != nil {
		fmt.Fprintln(os.Stderr, "test-flow:", err)
		os.Exit(2)
	}

	rep := NewReporter(len(cases))
	results := make([]*Result, 0, len(cases))

	repoRoot, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "test-flow:", err)
		os.Exit(2)
	}
	repoRoot, _ = filepath.Abs(repoRoot)

	opts := caseOptions{
		tempDirRoot:      *tmpRoot,
		skipPostCommands: *skipPost,
		skipTestCommands: *skipTest,
		keepScratch:      *keep,
		noCache:          *noCache,
		repoRoot:         repoRoot,
	}

	// Fail-fast by default — stop the loop on the first failing case so
	// developers see the failure immediately instead of waiting for the
	// remaining cases to finish. Pass -keep-going to run every case.
	stopped := false
	for _, tc := range cases {
		if tc.Disabled {
			continue
		}

		def, ok := registry.Get(tc.FlowID)
		if !ok {
			r := &Result{Case: tc, Err: fmt.Errorf("unknown flow_id %q", tc.FlowID)}
			rep.CaseStart(tc.Name, tc.FlowID)
			rep.Step("flow lookup", false, "", r.Err)
			rep.CaseEnd(false)
			results = append(results, r)
			if !*keepGoing {
				stopped = true
				break
			}
			continue
		}

		caseOpts := opts
		caseOpts.caseFile = tc.SourcePath
		caseOpts.flowsDir = flowsDir(repoRoot)

		r := runOne(ctx, tc, def, rt, rep, caseOpts)
		results = append(results, r)
		if !r.Pass() && !*keepGoing {
			stopped = true
			break
		}
	}

	if stopped {
		fmt.Fprintln(os.Stdout)
		fmt.Fprintln(os.Stdout, "Stopped at first failure (pass -keep-going to run every case).")
	}

	totalCases := 0
	for _, c := range cases {
		if !c.Disabled {
			totalCases++
		}
	}
	if Summarize(os.Stdout, results, totalCases) > 0 {
		os.Exit(1)
	}
}

// flowsDir returns the absolute path to the flows/ directory. The cache
// fingerprint hashes the whole directory so any edit to a flow definition
// invalidates every case (that's the desired behaviour: it's hard to tell
// from a flow ID alone which Go file produced it, and over-invalidation is
// safer than missing a relevant change).
func flowsDir(repoRoot string) string {
	return filepath.Join(repoRoot, "flows")
}

// filterCases narrows cases to those whose Name appears in the comma-separated
// only string. Empty only returns cases unchanged.
func filterCases(cases []*TestCase, only string) []*TestCase {
	if only == "" {
		return cases
	}
	wanted := map[string]bool{}
	start := 0
	for i := 0; i <= len(only); i++ {
		if i == len(only) || only[i] == ',' {
			name := only[start:i]
			if name != "" {
				wanted[name] = true
			}
			start = i + 1
		}
	}
	out := cases[:0:0]
	for _, c := range cases {
		if wanted[c.Name] {
			out = append(out, c)
		}
	}
	return out
}
