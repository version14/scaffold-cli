package generator

import (
	"testing"

	"github.com/version14/dot/internal/spec"
)

// stubGen is a minimal Generator for registry tests.
type stubGen struct {
	name     string
	language string
	modules  []string
}

func (s *stubGen) Name() string                                                  { return s.name }
func (s *stubGen) Language() string                                              { return s.language }
func (s *stubGen) Modules() []string                                             { return s.modules }
func (s *stubGen) Apply(_ spec.Spec) ([]FileOp, error)                           { return nil, nil }
func (s *stubGen) Commands() []CommandDef                                        { return nil }
func (s *stubGen) RunAction(_ string, _ []string, _ spec.Spec) ([]FileOp, error) { return nil, nil }

func makeSpec(lang string, modules ...string) spec.Spec {
	mods := make([]spec.ModuleSpec, len(modules))
	for i, m := range modules {
		mods[i] = spec.ModuleSpec{Name: m}
	}
	return spec.Spec{
		Project: spec.ProjectSpec{Language: lang},
		Modules: mods,
	}
}

// TestRegistryForSpec covers the 8-case table from the design doc.
func TestRegistryForSpec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		registered []Generator
		spec       spec.Spec
		wantNames  []string // generator names expected in result
	}{
		{
			name: "exact language+module match",
			registered: []Generator{
				&stubGen{name: "go-api", language: "go", modules: []string{"rest-api"}},
			},
			spec:      makeSpec("go", "rest-api"),
			wantNames: []string{"go-api"},
		},
		{
			name: "language mismatch — excluded",
			registered: []Generator{
				&stubGen{name: "py-api", language: "python", modules: []string{"rest-api"}},
			},
			spec:      makeSpec("go", "rest-api"),
			wantNames: nil,
		},
		{
			name: "language-agnostic matches any lang",
			registered: []Generator{
				&stubGen{name: "ci-gen", language: "*", modules: []string{"github-actions"}},
			},
			spec:      makeSpec("go", "github-actions"),
			wantNames: []string{"ci-gen"},
		},
		{
			name: "language-agnostic and language-specific both match",
			registered: []Generator{
				&stubGen{name: "go-api", language: "go", modules: []string{"rest-api"}},
				&stubGen{name: "ci-gen", language: "*", modules: []string{"github-actions"}},
			},
			spec:      makeSpec("go", "rest-api", "github-actions"),
			wantNames: []string{"go-api", "ci-gen"},
		},
		{
			name: "module not requested — excluded",
			registered: []Generator{
				&stubGen{name: "go-api", language: "go", modules: []string{"rest-api"}},
			},
			spec:      makeSpec("go", "postgres"),
			wantNames: nil,
		},
		{
			name: "multiple modules — partial match included",
			registered: []Generator{
				&stubGen{name: "go-api", language: "go", modules: []string{"rest-api", "postgres"}},
			},
			spec:      makeSpec("go", "rest-api"),
			wantNames: []string{"go-api"},
		},
		{
			name:       "empty registry — returns nil",
			registered: nil,
			spec:       makeSpec("go", "rest-api"),
			wantNames:  nil,
		},
		{
			name: "empty spec modules — no generators matched",
			registered: []Generator{
				&stubGen{name: "go-api", language: "go", modules: []string{"rest-api"}},
			},
			spec:      makeSpec("go"),
			wantNames: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reg := &Registry{}
			for _, g := range tt.registered {
				if err := reg.Register(g); err != nil {
					t.Fatalf("unexpected Register error: %v", err)
				}
			}

			got := reg.ForSpec(tt.spec)

			if len(got) != len(tt.wantNames) {
				t.Fatalf("ForSpec returned %d generators, want %d: got %v want %v",
					len(got), len(tt.wantNames), generatorNames(got), tt.wantNames)
			}
			for i, g := range got {
				if g.Name() != tt.wantNames[i] {
					t.Errorf("result[%d] = %q, want %q", i, g.Name(), tt.wantNames[i])
				}
			}
		})
	}
}

// TestRegistryRegisterConflict verifies that duplicate (language, module) claims
// are caught at registration time, not later.
func TestRegistryRegisterConflict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		a, b    Generator
		wantErr bool
	}{
		{
			name:    "same language same module — conflict",
			a:       &stubGen{name: "a", language: "go", modules: []string{"rest-api"}},
			b:       &stubGen{name: "b", language: "go", modules: []string{"rest-api"}},
			wantErr: true,
		},
		{
			name:    "language-agnostic vs specific — conflict",
			a:       &stubGen{name: "a", language: "*", modules: []string{"ci"}},
			b:       &stubGen{name: "b", language: "go", modules: []string{"ci"}},
			wantErr: true,
		},
		{
			name:    "different languages same module — no conflict",
			a:       &stubGen{name: "a", language: "go", modules: []string{"rest-api"}},
			b:       &stubGen{name: "b", language: "python", modules: []string{"rest-api"}},
			wantErr: false,
		},
		{
			name:    "same language different modules — no conflict",
			a:       &stubGen{name: "a", language: "go", modules: []string{"rest-api"}},
			b:       &stubGen{name: "b", language: "go", modules: []string{"postgres"}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reg := &Registry{}
			if err := reg.Register(tt.a); err != nil {
				t.Fatalf("first Register failed: %v", err)
			}
			err := reg.Register(tt.b)
			if tt.wantErr && err == nil {
				t.Error("expected conflict error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func generatorNames(gs []Generator) []string {
	names := make([]string, len(gs))
	for i, g := range gs {
		names[i] = g.Name()
	}
	return names
}
