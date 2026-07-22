# Development Workflow

## 1. Branching Strategy
* **Main Branch**: `main` is the primary stable default branch.
* **Feature Branches**: Create feature or bugfix branches from `main` using descriptive names (e.g., `feat/routines-filter`).
* **Pull Requests**: Merge feature branches back into `main` via squash merges to keep a clean commit history.

## 2. Commit Messages
* Follow the Conventional Commits specification:
  * `feat(scope): description` for new features
  * `fix(scope): description` for bug fixes
  * `chore(scope): description` for maintenance and dependency updates
  * `docs(scope): description` for documentation changes

## 3. Pull Request & CI Checks
* Every PR must pass:
  * Backend: `go test ./internal/...`
  * Frontend: typechecking and tests (`bun run check`, `bun run test`)
  * Linting: biome checks for TypeScript, golangci-lint for Go.
