package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/version14/dot/pkg/dotapi"
)

// ValidationFailure describes a single failed Check.
type ValidationFailure struct {
	Validator string // dotapi.Validator.Name
	Source    string // generator name
	Check     dotapi.Check
	Reason    string
}

func (f ValidationFailure) String() string {
	return fmt.Sprintf("[%s/%s] %s %s: %s", f.Source, f.Validator, f.Check.Type, f.Check.Path, f.Reason)
}

// RunValidators executes every Check in every Validator across every manifest
// against the on-disk project at root. It returns the list of failures (empty
// = pass). The first error from disk I/O short-circuits.
func RunValidators(root string, manifests []dotapi.Manifest) ([]ValidationFailure, error) {
	var failures []ValidationFailure

	for _, m := range manifests {
		checkRoot := root
		if m.PathPrefix != "" {
			checkRoot = filepath.Join(root, filepath.FromSlash(m.PathPrefix))
		}
		for _, v := range m.Validators {
			for _, c := range v.Checks {
				ok, reason, err := runCheck(checkRoot, c)
				if err != nil {
					return failures, fmt.Errorf("validator %s/%s: %w", m.Name, v.Name, err)
				}
				if !ok {
					failures = append(failures, ValidationFailure{
						Validator: v.Name,
						Source:    m.Name,
						Check:     c,
						Reason:    reason,
					})
				}
			}
		}
	}
	return failures, nil
}

func runCheck(root string, c dotapi.Check) (bool, string, error) {
	switch c.Type {
	case dotapi.CheckFileExists:
		full := filepath.Join(root, filepath.FromSlash(c.Path))
		if _, err := os.Stat(full); err == nil {
			return true, "", nil
		} else if os.IsNotExist(err) {
			return false, "missing", nil
		} else {
			return false, "", err
		}

	case dotapi.CheckJSONKeyExists:
		full := filepath.Join(root, filepath.FromSlash(c.Path))
		data, err := os.ReadFile(full)
		if err != nil {
			if os.IsNotExist(err) {
				return false, "file missing", nil
			}
			return false, "", err
		}
		var doc map[string]interface{}
		if err := json.Unmarshal(data, &doc); err != nil {
			return false, "invalid JSON", nil
		}
		if !lookupDotted(doc, c.Key) {
			return false, "key " + c.Key + " not found", nil
		}
		return true, "", nil

	default:
		return false, "unknown check type", nil
	}
}

// lookupDotted returns true if dotted path resolves to a present (possibly
// nil) value in doc. Maps are descended; non-map mid-segments fail.
func lookupDotted(doc map[string]interface{}, key string) bool {
	parts := strings.Split(key, ".")
	var cursor interface{} = doc
	for _, p := range parts {
		m, ok := cursor.(map[string]interface{})
		if !ok {
			return false
		}
		cursor, ok = m[p]
		if !ok {
			return false
		}
	}
	return true
}
