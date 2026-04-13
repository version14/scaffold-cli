package pipeline

import (
	"errors"
	"testing"
)

// TestPatchImportBlock covers the full case table from the design doc.
// Every documented supported and unsupported form must be tested here before
// AnchorImportBlock ships in v0.1.
func TestPatchImportBlock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		src       string
		newImport string
		want      string
		wantErr   error
	}{
		{
			name:      "block import — add new pkg",
			src:       "package main\n\nimport (\n\t\"fmt\"\n)\n",
			newImport: "os",
			want:      "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n)\n",
		},
		{
			name:      "block import — duplicate skipped",
			src:       "package main\n\nimport (\n\t\"fmt\"\n)\n",
			newImport: "fmt",
			want:      "package main\n\nimport (\n\t\"fmt\"\n)\n",
		},
		{
			name:      "single-line import — expand to block",
			src:       "package main\n\nimport \"fmt\"\n",
			newImport: "os",
			want:      "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n)\n",
		},
		{
			name:      "single-line import — duplicate skipped",
			src:       "package main\n\nimport \"fmt\"\n",
			newImport: "fmt",
			want:      "package main\n\nimport \"fmt\"\n",
		},
		{
			name:      "no import — insert after package line",
			src:       "package main\n\nfunc main() {}\n",
			newImport: "fmt",
			want:      "package main\n\nimport \"fmt\"\n\nfunc main() {}\n",
		},
		{
			name:      "package only no blank line — insert after package",
			src:       "package main",
			newImport: "fmt",
			want:      "package main\n\nimport \"fmt\"\n",
		},
		{
			name:      "blank import — unsupported",
			src:       "package main\n\nimport _ \"embed\"\n",
			newImport: "fmt",
			wantErr:   ErrUnsupportedImportForm,
		},
		{
			name:      "build tag — unsupported",
			src:       "//go:build linux\n\npackage main\n\nimport \"fmt\"\n",
			newImport: "os",
			wantErr:   ErrUnsupportedImportForm,
		},
		{
			name:      "multiple single-line imports — unsupported",
			src:       "package main\n\nimport \"fmt\"\nimport \"os\"\n",
			newImport: "io",
			wantErr:   ErrUnsupportedImportForm,
		},
		{
			name:      "block import — already has multiple imports, add new",
			src:       "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n)\n",
			newImport: "io",
			want:      "package main\n\nimport (\n\t\"fmt\"\n\t\"os\"\n\t\"io\"\n)\n",
		},
		{
			name:      "third-party import path",
			src:       "package main\n\nimport (\n\t\"fmt\"\n)\n",
			newImport: "github.com/charmbracelet/huh",
			want:      "package main\n\nimport (\n\t\"fmt\"\n\t\"github.com/charmbracelet/huh\"\n)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := patchImportBlock(tt.src, tt.newImport)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

// TestPatchFunc covers the patchFunc helper used by AnchorMainFunc and AnchorInitFunc.
func TestPatchFunc(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		src      string
		funcName string
		content  string
		want     string
		wantErr  bool
	}{
		{
			name:     "insert before closing brace of main",
			src:      "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}\n",
			funcName: "main",
			content:  "println(\"world\")",
			want:     "package main\n\nfunc main() {\n\tprintln(\"hello\")\n\tprintln(\"world\")\n}\n",
		},
		{
			name:     "func not found — error",
			src:      "package main\n\nfunc other() {}\n",
			funcName: "main",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := patchFunc(tt.src, tt.funcName, tt.content)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}
