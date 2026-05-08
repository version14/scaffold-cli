package state

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// JSONDoc is a thin wrapper over a generic JSON value with helpers for
// nested key access. Nested keys use dotted notation: "scripts.build".
type JSONDoc struct {
	root map[string]interface{}
}

func NewJSONDoc() *JSONDoc {
	return &JSONDoc{root: map[string]interface{}{}}
}

func (d *JSONDoc) Load(data []byte) error {
	if len(data) == 0 {
		d.root = map[string]interface{}{}
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return fmt.Errorf("json: unmarshal: %w", err)
	}
	if m == nil {
		m = map[string]interface{}{}
	}
	d.root = m
	return nil
}

func (d *JSONDoc) Marshal() ([]byte, error) {
	out, err := json.MarshalIndent(d.root, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("json: marshal: %w", err)
	}
	return append(out, '\n'), nil
}

// Root exposes the underlying map for advanced use.
func (d *JSONDoc) Root() map[string]interface{} { return d.root }

// SetNested sets a value at a dotted key, creating intermediate maps.
func (d *JSONDoc) SetNested(path string, value interface{}) error {
	keys := splitPath(path)
	if len(keys) == 0 {
		return fmt.Errorf("json: empty path")
	}
	cursor := d.root
	for i, k := range keys[:len(keys)-1] {
		next, ok := cursor[k]
		if !ok {
			child := map[string]interface{}{}
			cursor[k] = child
			cursor = child
			continue
		}
		child, ok := next.(map[string]interface{})
		if !ok {
			return fmt.Errorf("json: %q is not an object at segment %d", path, i)
		}
		cursor = child
	}
	cursor[keys[len(keys)-1]] = value
	return nil
}

// GetNested returns the value at a dotted path. Second return is false when
// any segment is missing.
func (d *JSONDoc) GetNested(path string) (interface{}, bool) {
	keys := splitPath(path)
	if len(keys) == 0 {
		return nil, false
	}
	var cursor interface{} = d.root
	for _, k := range keys {
		m, ok := cursor.(map[string]interface{})
		if !ok {
			return nil, false
		}
		cursor, ok = m[k]
		if !ok {
			return nil, false
		}
	}
	return cursor, true
}

// DeleteKey removes a top-level or nested key. Missing keys are a no-op.
func (d *JSONDoc) DeleteKey(path string) {
	keys := splitPath(path)
	if len(keys) == 0 {
		return
	}
	cursor := d.root
	for _, k := range keys[:len(keys)-1] {
		next, ok := cursor[k].(map[string]interface{})
		if !ok {
			return
		}
		cursor = next
	}
	delete(cursor, keys[len(keys)-1])
}

// AppendStringSet appends string values into the array at the dotted path,
// deduplicating and sorting the result for deterministic output. Intermediate
// objects and the array itself are created if missing — this is the right
// helper when several generators each contribute entries to a shared list
// (e.g. `pnpm.onlyBuiltDependencies`). Returns an error if a non-array value
// already lives at path or any intermediate segment is not an object.
func (d *JSONDoc) AppendStringSet(path string, values ...string) error {
	keys := splitPath(path)
	if len(keys) == 0 {
		return fmt.Errorf("json: empty path")
	}
	cursor := d.root
	for i, k := range keys[:len(keys)-1] {
		next, ok := cursor[k]
		if !ok {
			child := map[string]interface{}{}
			cursor[k] = child
			cursor = child
			continue
		}
		child, ok := next.(map[string]interface{})
		if !ok {
			return fmt.Errorf("json: %q is not an object at segment %d", path, i)
		}
		cursor = child
	}
	leaf := keys[len(keys)-1]
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	switch existing := cursor[leaf].(type) {
	case nil:
		// fresh array
	case []interface{}:
		for _, v := range existing {
			s, ok := v.(string)
			if !ok {
				return fmt.Errorf("json: %q contains non-string element", path)
			}
			if _, dup := seen[s]; !dup {
				seen[s] = struct{}{}
				out = append(out, s)
			}
		}
	default:
		return fmt.Errorf("json: %q is not an array", path)
	}
	for _, v := range values {
		if _, dup := seen[v]; !dup {
			seen[v] = struct{}{}
			out = append(out, v)
		}
	}
	sort.Strings(out)
	arr := make([]interface{}, len(out))
	for i, s := range out {
		arr[i] = s
	}
	cursor[leaf] = arr
	return nil
}

// AddDep is a convenience for the common case of adding a key under a
// "dependencies" or "devDependencies" object.
func (d *JSONDoc) AddDep(section, name, version string) error {
	return d.SetNested(section+"."+name, version)
}

// Merge deep-merges src into the document. Values in src override existing
// scalars; maps are merged recursively.
func (d *JSONDoc) Merge(src map[string]interface{}) {
	mergeMap(d.root, src)
}

func mergeMap(dst, src map[string]interface{}) {
	for k, v := range src {
		if srcMap, ok := v.(map[string]interface{}); ok {
			if dstMap, ok := dst[k].(map[string]interface{}); ok {
				mergeMap(dstMap, srcMap)
				continue
			}
		}
		dst[k] = v
	}
}

func splitPath(path string) []string {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	return strings.Split(path, ".")
}
