package generator

// FileOpKind describes the type of file operation a generator requests.
// It is a string so that it serialises stably over JSON (e.g. for a future
// subprocess protocol). Never use integer constants here — adding a new kind
// in the middle of an iota would break all community generators.
type FileOpKind string

const (
	// Create writes a new file. If the file already exists the op with the
	// highest Priority wins. Two ops at the same priority → pipeline error.
	Create FileOpKind = "create"

	// Template renders a Go text/template file and writes the result.
	// Behaves like Create for conflict resolution.
	Template FileOpKind = "template"

	// Append adds content to the end of an existing file (or creates it).
	// All Append ops for the same file are applied in priority order.
	Append FileOpKind = "append"

	// Patch applies a targeted in-file insertion at a named anchor point.
	// Supported anchors are defined as AnchorXxx constants below.
	Patch FileOpKind = "patch"
)

// Anchor names for Patch ops. Generators must use these constants, not raw
// strings, so that the pipeline can match them reliably.
const (
	// AnchorImportBlock inserts an import path into the file's import block.
	// Supported forms: standard import ( ... ) and single import "pkg".
	// Unsupported: blank imports, build-tag-gated imports, multiple
	// single-line imports. The pipeline returns ErrUnsupportedImportForm
	// for those — generators must not emit this anchor for such files.
	AnchorImportBlock = "import_block"

	// AnchorMainFunc inserts content before the closing brace of func main().
	AnchorMainFunc = "main_func"

	// AnchorInitFunc inserts content before the closing brace of func init().
	AnchorInitFunc = "init_func"
)

// FileOp describes a single file operation produced by a generator.
// The pipeline collects all FileOps from all matched generators, resolves
// conflicts, then writes everything atomically (no partial writes on error).
type FileOp struct {
	Kind      FileOpKind `json:"kind"`
	Path      string     `json:"path"`      // relative to project root
	Content   string     `json:"content"`   // file content or template source
	Anchor    string     `json:"anchor"`    // used only for Patch ops
	Generator string     `json:"generator"` // name of the generator that produced this op
	Priority  int        `json:"priority"`  // higher wins on Create/Template conflicts
}
