package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/version14/dot/internal/flow"
)

// TestCase is one scripted run of a flow.
//
//	Name             — used in reports and as the testdata file basename
//	FlowID           — selects which flow root to traverse (e.g. "monorepo")
//	Answers          — questionID → recorded answer (string, bool, []string, …)
//	ExpectedIDs      — optional list of question IDs the engine MUST visit
//	SkipPostCommands — when true, do not run PostGenerationCommands
//	SkipTestCommands — when true, do not run TestCommands
//	SourcePath       — absolute path to the JSON file (set by LoadCases),
//	                  used to fingerprint the case for the test-flow cache.
type TestCase struct {
	Name             string                 `json:"name"`
	FlowID           string                 `json:"flow_id"`
	Answers          map[string]flow.Answer `json:"answers"`
	Disabled         bool                   `json:"disabled"`
	ExpectedIDs      []string               `json:"expected_visited,omitempty"`
	SkipPostCommands bool                   `json:"skip_post_commands,omitempty"`
	SkipTestCommands bool                   `json:"skip_test_commands,omitempty"`

	SourcePath string `json:"-"`
}

// LoadCases reads every *.json file under dir and parses it as a TestCase.
// File ordering is lexicographic for stable report output.
func LoadCases(dir string) ([]*TestCase, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("test-flow: read testdata: %w", err)
	}

	var cases []*TestCase
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		path := filepath.Join(dir, e.Name())
		tc, err := loadCase(path)
		if err != nil {
			return nil, err
		}
		if tc.Name == "" {
			tc.Name = e.Name()
		}
		abs, err := filepath.Abs(path)
		if err != nil {
			abs = path
		}
		tc.SourcePath = abs
		cases = append(cases, tc)
	}
	return cases, nil
}

func loadCase(path string) (*TestCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("test-flow: read %s: %w", path, err)
	}
	var tc TestCase
	if err := json.Unmarshal(data, &tc); err != nil {
		return nil, fmt.Errorf("test-flow: parse %s: %w", path, err)
	}
	return &tc, nil
}
