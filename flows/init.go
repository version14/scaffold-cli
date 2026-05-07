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

	authMethod := &flow.OptionQuestion{
		QuestionBase: flow.QuestionBase{ID_: "ts-backend-auth-method"},
		Label:        "Choose an auth method.",
		Options: []*flow.Option{
			{Label: "BetterAuth (sessions + Drizzle adapter)", Value: "better-auth", Next: &flow.Next{Question: confirmGenerate}},
			{Label: "Vanilla JWT", Value: "jwt", Next: &flow.Next{Question: confirmGenerate}},
		},
	}

	enableAuth := &flow.ConfirmQuestion{
		QuestionBase: flow.QuestionBase{ID_: "enable-auth"},
		Label:        "Enable authentication?",
		Default:      false,
		Then:         &flow.Next{Question: authMethod},
		Else:         &flow.Next{Question: confirmGenerate},
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
		Label:        "Link a database to this project?",
		Default:      false,
		Then:         &flow.Next{Question: dbType},
		Else:         &flow.Next{Question: confirmGenerate},
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

	stack := &flow.OptionQuestion{
		QuestionBase: flow.QuestionBase{ID_: "stack"},
		Label:        "Primary language stack",
		Description:  "DOT will scaffold the matching toolchain.",
		Options: []*flow.Option{
			{Label: "TypeScript", Value: "typescript", Next: &flow.Next{Question: framework}},
			// {Label: "Go", Value: "go", Next: &flow.Next{Question: confirmGenerate}},
		},
	}

	monorepoType := &flow.OptionQuestion{
		QuestionBase: flow.QuestionBase{ID_: "monorepo_type"},
		Label:        "Monorepo style",
		Options: []*flow.Option{
			{Label: "Single app (no monorepo)", Value: "single", Next: &flow.Next{Question: stack}},
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

// resolveMonorepoGenerators maps the populated spec to the ordered generator
// invocations. Order is significant: dependents come after their deps.
func resolveMonorepoGenerators(s *spec.ProjectSpec) []Invocation {
	if s == nil {
		return nil
	}

	out := []Invocation{
		{Name: "base_project"},
	}

	stack, _ := s.Answers["stack"].(string)
	framework, _ := s.Answers["ts-backend-framework"].(string)
	architecture, _ := s.Answers["ts-backend-architecture"].(string)
	formatter, _ := s.Answers["ts-backend-formatter"].(string)
	dbEnabled, _ := s.Answers["enable-db"].(bool)
	dbType, _ := s.Answers["ts-backend-db-type"].(string)
	orm, _ := s.Answers["ts-backend-orm"].(string)
	authEnabled, _ := s.Answers["enable-auth"].(bool)
	authMethod, _ := s.Answers["ts-backend-auth-method"].(string)
	decoratorsEnabled, _ := s.Answers["ts-backend-decorators-validation"].(bool)

	if stack == "typescript" {
		out = append(out, Invocation{Name: "typescript_base"})
	}

	if framework == "express" {
		out = append(out,
			Invocation{Name: "express_server_entrypoint"},
			Invocation{Name: "express_server_typescript_deps"},
			Invocation{Name: "express_node_tsconfig"},
			Invocation{Name: "express_shared_errors"},
			Invocation{Name: "express_error_middleware"},
			Invocation{Name: "express_rate_limit"},
			Invocation{Name: "express_test_setup"},
		)
	}

	switch architecture {
	case CLEAN_ARCHITECTURE:
		out = append(out, Invocation{Name: "backend_architecture_clean_architecture"})
	case MVC_ARCHITECTURE:
		out = append(out, Invocation{Name: "backend_architecture_mvc"})
	case HEXAGONAL_ARCHITECTURE:
		out = append(out, Invocation{Name: "backend_architecture_hexagonal_architecture"})
	}

	if framework == "express" {
		if decoratorsEnabled {
			out = append(out,
				Invocation{Name: "zod_validation_deps"},
				Invocation{Name: "express_decorators_core"},
				Invocation{Name: "express_openapi_setup"},
			)
			switch architecture {
			case CLEAN_ARCHITECTURE:
				out = append(out, Invocation{Name: "decorators_clean_arch_adapter"})
			case MVC_ARCHITECTURE:
				out = append(out, Invocation{Name: "decorators_mvc_adapter"})
			case HEXAGONAL_ARCHITECTURE:
				out = append(out, Invocation{Name: "decorators_hexagonal_adapter"})
			}
		} else {
			// Always wire the JSDoc-based Swagger so /docs works regardless of
			// the decorator choice — generated controllers ship with @openapi
			// comments that swagger-jsdoc picks up at boot.
			out = append(out, Invocation{Name: "express_swagger_jsdoc"})
		}
	}

	if formatter == "prettier" {
		out = append(out,
			Invocation{Name: "prettier_config"},
			Invocation{Name: "prettier_typescript_deps"},
			Invocation{Name: "prettier_express_rules"},
		)
	} else if formatter == "biome" {
		out = append(out, Invocation{Name: "biome_config"})
	}

	if dbEnabled {
		if dbType == "postgres" {
			out = append(out,
				Invocation{Name: "postgres_docker_compose"},
				Invocation{Name: "postgres_env_example"},
			)
		}
		if orm == "drizzle" {
			out = append(out,
				Invocation{Name: "drizzle_config_base"},
				Invocation{Name: "drizzle_typescript_deps"},
				Invocation{Name: "drizzle_postgres_adapter"},
			)
		}
	}

	if authEnabled {
		out = append(out, Invocation{Name: "express_auth_validators"})
		switch authMethod {
		case "better-auth":
			// BetterAuth needs Drizzle + Postgres; add them if not already included
			if !dbEnabled {
				out = append(out,
					Invocation{Name: "postgres_docker_compose"},
					Invocation{Name: "postgres_env_example"},
					Invocation{Name: "drizzle_config_base"},
					Invocation{Name: "drizzle_typescript_deps"},
					Invocation{Name: "drizzle_postgres_adapter"},
				)
			}
			out = append(out, Invocation{Name: "auth_better_auth"})
			out = append(out, Invocation{Name: "auth_better_auth_schema"})
		case "jwt":
			out = append(out, Invocation{Name: "auth_jwt_vanilla"})
			if dbEnabled && orm == "drizzle" {
				out = append(out, Invocation{Name: "auth_jwt_users_schema"})
			}
			switch architecture {
			case MVC_ARCHITECTURE:
				out = append(out, Invocation{Name: "auth_jwt_mvc_route"})
			case CLEAN_ARCHITECTURE:
				if dbEnabled && orm == "drizzle" {
					out = append(out, Invocation{Name: "auth_jwt_clean_arch_module"})
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
