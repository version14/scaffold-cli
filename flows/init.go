package flows

import (
	"github.com/version14/dot/internal/flow"
	"github.com/version14/dot/internal/spec"
)

const CLEAN_ARCHITECTURE = "clean-architecture"
const MVC_ARCHITECTURE = "mvc-architecture"
const HEXAGONAL_ARCHITECTURE = "hexagonal-architecture"

// ValidationLibZod is the ID of the Zod validation library — kept as a
// constant so future libraries (yup, valibot, …) can be added without
// scattering string literals.
const ValidationLibZod = "zod"

// InitFlow is the default DOT scaffolding flow. It walks the user through
// project name → monorepo structure → language stack → linting → database → auth.
//
// Question IDs are kept stable: re-runs of `dot scaffold` reuse the persisted
// answers from .dot/spec.json keyed by these IDs.
func InitFlow() *FlowDef {
	confirmGenerate := &flow.ConfirmQuestion{
		QuestionBase: flow.QuestionBase{ID_: "confirm-generate"},
		Label:        "Generate the project now?",
		Default:      true,
		Then:         &flow.Next{End: true},
		Else:         &flow.Next{End: true},
	}

	stack := buildTSBackendChain(&flow.Next{Question: confirmGenerate})

	appsCount := &flow.LoopQuestion{
		QuestionBase: flow.QuestionBase{
			ID_: "apps_count",
		},
		Label:    "Number of apps",
		Body:     buildPerAppBody(),
		Continue: &flow.Next{Question: confirmGenerate},
	}

	monorepoType := &flow.OptionQuestion{
		QuestionBase: flow.QuestionBase{ID_: "monorepo_type"},
		Label:        "Monorepo style",
		Options: []*flow.Option{
			{Label: "Single app (no monorepo)", Value: "single", Next: &flow.Next{Question: stack}},
			{Label: "Multi  apps (monorepo)", Value: "multi", Next: &flow.Next{Question: appsCount}},
			// {Label: "Turborepo", Value: "turborepo", Next: &flow.Next{Question: stack}},
		},
	}

	projectName := &flow.TextQuestion{
		QuestionBase: flow.QuestionBase{
			ID_:   "project_name",
			Next_: &flow.Next{Question: monorepoType},
		},
		Label:       "Project name",
		Description: "Used as the package name and root directory.",
		Default:     "my-project",
		Validate:    nonEmpty,
	}

	return &FlowDef{
		ID:          "init",
		Title:       "Init / Project Wizard",
		Description: "Scaffold a new project with optional monorepo, language, and tooling.",
		Root:        projectName,
		Generators:  resolveMonorepoGenerators,
	}
}

// buildTSBackendChain builds the reusable question sub-graph from `stack` down
// to terminal. Used by both the single-app flow (terminal → confirmGenerate)
// and the per-app loop body (terminal → End:true) so question definitions
// are never duplicated.
func buildTSBackendChain(terminal *flow.Next) flow.Question {
	authMethod := &flow.OptionQuestion{
		QuestionBase: flow.QuestionBase{ID_: "ts-backend-auth-method"},
		Label:        "Choose an auth method.",
		Options: []*flow.Option{
			{Label: "BetterAuth (sessions + Drizzle adapter)", Value: "better-auth", Next: terminal},
			{Label: "Vanilla JWT", Value: "jwt", Next: terminal},
		},
	}

	enableAuth := &flow.ConfirmQuestion{
		QuestionBase: flow.QuestionBase{ID_: "enable-auth"},
		Label:        "Enable authentication?",
		Default:      false,
		Then:         &flow.Next{Question: authMethod},
		Else:         terminal,
	}

	orm := &flow.OptionQuestion{
		QuestionBase: flow.QuestionBase{ID_: "ts-backend-orm"},
		Label:        "Choose an ORM.",
		Options: []*flow.Option{
			{Label: "Drizzle", Value: "drizzle", Next: &flow.Next{Question: enableAuth}},
		},
	}

	dbType := &flow.OptionQuestion{
		QuestionBase: flow.QuestionBase{ID_: "ts-backend-db-type"},
		Label:        "Choose a database.",
		Options: []*flow.Option{
			{Label: "PostgreSQL", Value: "postgres", Next: &flow.Next{Question: orm}},
		},
	}

	enableDb := &flow.ConfirmQuestion{
		QuestionBase: flow.QuestionBase{ID_: "enable-db"},
		Label:        "Link a database?",
		Default:      false,
		Then:         &flow.Next{Question: dbType},
		Else:         &flow.Next{Question: enableAuth},
	}

	linter := &flow.OptionQuestion{
		QuestionBase: flow.QuestionBase{ID_: "ts-backend-linter"},
		Label:        "Choose a linter.",
		Options: []*flow.Option{
			{Label: "Biome", Value: "biome", Next: &flow.Next{Question: enableDb}},
			{Label: "Prettier", Value: "prettier", Next: &flow.Next{Question: enableDb}},
		},
	}

	formatter := &flow.OptionQuestion{
		QuestionBase: flow.QuestionBase{ID_: "ts-backend-formatter"},
		Label:        "Choose a formatter.",
		Options: []*flow.Option{
			{Label: "Biome", Value: "biome", Next: &flow.Next{Question: linter}},
			{Label: "Prettier", Value: "prettier", Next: &flow.Next{Question: linter}},
		},
	}

	validationLib := &flow.OptionQuestion{
		QuestionBase: flow.QuestionBase{ID_: "ts-backend-validation-lib"},
		Label:        "Validation library",
		Description:  "Schema library used to validate request inputs and document the OpenAPI spec.",
		Options: []*flow.Option{
			{Label: "Zod", Value: ValidationLibZod, Next: &flow.Next{Question: formatter}},
		},
	}

	decorators := &flow.ConfirmQuestion{
		QuestionBase: flow.QuestionBase{ID_: "ts-backend-decorators-validation"},
		Label:        "Use decorator-based validation and OpenAPI documentation?",
		Description: "OpenAPI/Swagger is always available at /docs. Choose Yes for an end-to-end " +
			"decorator API (@Controller, @Get, @Body, @Response) with automatic Zod request/response " +
			"validation and a spec built from runtime metadata. Choose No to keep plain Express " +
			"handlers — the generated code includes JSDoc @openapi comments that swagger-jsdoc scans " +
			"to build the spec.",
		Default: true,
		Then:    &flow.Next{Question: validationLib},
		Else:    &flow.Next{Question: formatter},
	}

	architecture := &flow.OptionQuestion{
		QuestionBase: flow.QuestionBase{ID_: "ts-backend-architecture"},
		Label:        "Choose your architecture.",
		Options: []*flow.Option{
			{Label: "Clean Architecture", Value: CLEAN_ARCHITECTURE, Next: &flow.Next{Question: decorators}},
			{Label: "MVC", Value: MVC_ARCHITECTURE, Next: &flow.Next{Question: decorators}},
			// {Label: "Hexagonal", Value: HEXAGONAL_ARCHITECTURE, Next: &flow.Next{Question: decorators}},
		},
	}

	framework := &flow.OptionQuestion{
		QuestionBase: flow.QuestionBase{ID_: "ts-backend-framework"},
		Label:        "Library / Framework",
		Description:  "Choose a library or framework to scaffold your backend.",
		Options: []*flow.Option{
			{Label: "Express", Value: "express", Next: &flow.Next{Question: architecture}},
		},
	}

	return &flow.OptionQuestion{
		QuestionBase: flow.QuestionBase{ID_: "stack"},
		Label:        "Primary language stack",
		Description:  "DOT will scaffold the matching toolchain.",
		Options: []*flow.Option{
			{Label: "TypeScript", Value: "typescript", Next: &flow.Next{Question: framework}},
			// {Label: "Go", Value: "go", Next: terminal},
		},
	}
}

// buildPerAppBody returns a fresh question sub-graph for one loop iteration.
// It reuses buildTSBackendChain with End:true as the terminal so each body
// question terminates the iteration instead of routing into confirmGenerate.
func buildPerAppBody() []flow.Question {
	chain := buildTSBackendChain(&flow.Next{End: true})

	appName := &flow.TextQuestion{
		QuestionBase: flow.QuestionBase{
			ID_:   "app-name",
			Next_: &flow.Next{Question: chain},
		},
		Label:       "App name",
		Description: "Used as the app's directory name (apps/<name>/).",
		Validate:    nonEmpty,
	}

	return []flow.Question{appName}
}

// resolveMonorepoGenerators maps the populated spec to the ordered generator
// invocations. Order is significant: dependents come after their deps.
func resolveMonorepoGenerators(s *spec.ProjectSpec) []Invocation {
	if s == nil {
		return nil
	}

	out := []Invocation{{Name: "base_project"}}

	if monorepoType, _ := s.Answers["monorepo_type"].(string); monorepoType == "multi" {
		out = append(out, Invocation{Name: "monorepo_ts_workspaces"})
		for i, appAnswers := range extractAppAnswers(s.Answers["apps_count"]) {
			frame := flow.LoopFrame{
				QuestionID: "apps_count",
				Index:      i,
				Answers:    appAnswers,
			}
			out = append(out, resolveAppGenerators(appAnswers, []flow.LoopFrame{frame})...)
		}
		return out
	}

	// Single-app: read directly from top-level answers.
	flat := make(map[string]interface{}, len(s.Answers))
	for k, v := range s.Answers {
		flat[k] = v
	}
	out = append(out, resolveAppGenerators(flat, nil)...)
	return out
}

// extractAppAnswers normalises the apps_count answer into a typed slice.
// The engine stores []map[string]interface{}; JSON round-trip produces []interface{}.
func extractAppAnswers(raw interface{}) []map[string]interface{} {
	switch v := raw.(type) {
	case []map[string]interface{}:
		return v
	case []interface{}:
		out := make([]map[string]interface{}, 0, len(v))
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				out = append(out, m)
			}
		}
		return out
	}
	return nil
}

// resolveAppGenerators maps one app's answers to an ordered generator list.
// loopStack is nil for single-app; for multi-app it carries the iteration frame
// so the executor can scope writes to apps/<name>/.
func resolveAppGenerators(answers map[string]interface{}, loopStack []flow.LoopFrame) []Invocation {
	inv := func(name string) Invocation { return Invocation{Name: name, LoopStack: loopStack} }

	stack, _ := answers["stack"].(string)
	framework, _ := answers["ts-backend-framework"].(string)
	architecture, _ := answers["ts-backend-architecture"].(string)
	formatter, _ := answers["ts-backend-formatter"].(string)
	dbEnabled, _ := answers["enable-db"].(bool)
	dbType, _ := answers["ts-backend-db-type"].(string)
	orm, _ := answers["ts-backend-orm"].(string)
	authEnabled, _ := answers["enable-auth"].(bool)
	authMethod, _ := answers["ts-backend-auth-method"].(string)
	decoratorsEnabled, _ := answers["ts-backend-decorators-validation"].(bool)

	var out []Invocation

	if stack == "typescript" {
		out = append(out, inv("typescript_base"))
	}

	if framework == "express" {
		out = append(out,
			inv("express_server_entrypoint"),
			inv("express_server_typescript_deps"),
			inv("express_node_tsconfig"),
			inv("express_shared_errors"),
			inv("express_error_middleware"),
			inv("express_rate_limit"),
			inv("express_test_setup"),
		)
	}

	switch architecture {
	case CLEAN_ARCHITECTURE:
		out = append(out, inv("backend_architecture_clean_architecture"))
	case MVC_ARCHITECTURE:
		out = append(out, inv("backend_architecture_mvc"))
	case HEXAGONAL_ARCHITECTURE:
		out = append(out, inv("backend_architecture_hexagonal_architecture"))
	}

	if framework == "express" {
		if decoratorsEnabled {
			out = append(out,
				inv("zod_validation_deps"),
				inv("express_decorators_core"),
				inv("express_openapi_setup"),
			)
			switch architecture {
			case CLEAN_ARCHITECTURE:
				out = append(out, inv("decorators_clean_arch_adapter"))
			case MVC_ARCHITECTURE:
				out = append(out, inv("decorators_mvc_adapter"))
			case HEXAGONAL_ARCHITECTURE:
				out = append(out, inv("decorators_hexagonal_adapter"))
			}
		} else {
			// Always wire the JSDoc-based Swagger so /docs works regardless of
			// the decorator choice — generated controllers ship with @openapi
			// comments that swagger-jsdoc picks up at boot.
			out = append(out, inv("express_swagger_jsdoc"))
		}
	}

	if formatter == "prettier" {
		out = append(out,
			inv("prettier_config"),
			inv("prettier_typescript_deps"),
			inv("prettier_express_rules"),
		)
	} else if formatter == "biome" {
		out = append(out, inv("biome_config"))
	}

	if dbEnabled {
		if dbType == "postgres" {
			out = append(out,
				inv("postgres_docker_compose"),
				inv("postgres_env_example"),
			)
		}
		if orm == "drizzle" {
			out = append(out,
				inv("drizzle_config_base"),
				inv("drizzle_typescript_deps"),
				inv("drizzle_postgres_adapter"),
			)
		}
	}

	if authEnabled {
		out = append(out, inv("express_auth_validators"))
		switch authMethod {
		case "better-auth":
			// BetterAuth needs Drizzle + Postgres; add them if not already included.
			if !dbEnabled {
				out = append(out,
					inv("postgres_docker_compose"),
					inv("postgres_env_example"),
					inv("drizzle_config_base"),
					inv("drizzle_typescript_deps"),
					inv("drizzle_postgres_adapter"),
				)
			}
			out = append(out, inv("auth_better_auth"))
			out = append(out, inv("auth_better_auth_schema"))
		case "jwt":
			out = append(out, inv("auth_jwt_vanilla"))
			if dbEnabled && orm == "drizzle" {
				out = append(out, inv("auth_jwt_users_schema"))
			}
			switch architecture {
			case MVC_ARCHITECTURE:
				out = append(out, inv("auth_jwt_mvc_route"))
			case CLEAN_ARCHITECTURE:
				if dbEnabled && orm == "drizzle" {
					out = append(out, inv("auth_jwt_clean_arch_module"))
				}
			}
		}
	}

	return out
}

func nonEmpty(s string) error {
	if s == "" {
		return errEmpty
	}
	return nil
}

// errEmpty is reused so we don't allocate per validate call.
var errEmpty = errString("required")

type errString string

func (e errString) Error() string { return string(e) }
