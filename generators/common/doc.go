// Package common holds language-agnostic generators that work across all
// project types. Examples: GitHub Actions CI, Docker, docker-compose.
//
// Generators here use Language() == "*" so they match any spec language.
// Planned for v0.2:
//   - GitHubActionsGenerator  (module: "github-actions")
//   - DockerGenerator         (module: "docker")
//   - DockerComposeGenerator  (module: "docker-compose")
package common
