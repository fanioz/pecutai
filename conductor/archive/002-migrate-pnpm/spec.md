# Spec: Migrate Project from Bun to pnpm

## 1. Overview
This track covers the complete migration of the WUPHF project from using Bun as the package manager and runtime to using `pnpm` for package installation and resolution, and Node/Vitest for runtime execution and testing.

## 2. Functional Requirements
* **Lockfile Migration**: Remove `bun.lock` and generate a `pnpm-lock.yaml` file at the workspace root.
* **Workspace Configuration**: Configure `pnpm-workspace.yaml` containing the workspaces defined in `package.json` (`packages/*`, `apps/*`).
* **Dependency Resolution**: Ensure all dependencies install cleanly across all monorepo packages.
* **Scripts Update**: Rewrite all scripts in the root `package.json` and package-level `package.json` files to use `pnpm` instead of `bun` (e.g., replace `bun run` with `pnpm run`, `bun --cwd` with `pnpm --filter`).
* **CI/CD Workflows**: Update all GitHub Actions workflows in `.github/workflows/` (e.g. `ci.yml`, `nightly.yml`, `ratchet-check.yml`) to use actions for setting up Node and `pnpm`, caching dependencies, and running tests.
* **Test Runners**: Ensure test runs use `vitest` / `node` directly without dependency on the `bun` binary.

## 3. Non-Functional Requirements
* **Zero Downtime**: Existing local executables and Go code compilation must continue to build and function correctly.
* **Maintain Performance**: Ensure CI build times and local dependency installation remain fast and reliable.

## 4. Acceptance Criteria
* Running `pnpm install` installs all workspace packages successfully.
* All test commands (`pnpm run broker:test`, `pnpm run desktop:test`, etc.) compile and run to completion using Node/Vitest.
* GitHub action runs pass with clean status using `pnpm`.
