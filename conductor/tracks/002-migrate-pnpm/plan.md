# Implementation Plan: Migrate Project from Bun to pnpm

## Phase 1: Workspace & Lockfile Conversion
- [x] Task: Remove old lockfiles and configure pnpm workspace
    - [x] Delete `bun.lock` at workspace root.
    - [x] Create `pnpm-workspace.yaml` defining workspaces (`packages/*`, `apps/*`).
- [x] Task: Install dependencies with pnpm
    - [x] Run `pnpm install` at workspace root.
    - [x] Address any package resolution or hoisting issues.
- [x] Task: Conductor - User Manual Verification 'Phase 1: Workspace & Lockfile Conversion' (Protocol in workflow.md)

## Phase 2: Configuration & Scripts Migration
- [x] Task: Update package.json scripts
    - [x] Replace `bun run --cwd` scripts with `pnpm --filter` scripts in root `package.json`.
    - [x] Verify script compatibility inside `packages/*` and `apps/*`.
- [x] Task: Verify Vitest tests
    - [x] Execute `pnpm run broker:test` synchronously.
    - [x] Execute `pnpm run desktop:test` synchronously.
- [x] Task: Conductor - User Manual Verification 'Phase 2: Configuration & Scripts Migration' (Protocol in workflow.md)

## Phase 3: CI/CD Pipeline Migration
- [x] Task: Update GitHub Actions workflows
    - [x] Replace `setup-bun` actions with `pnpm/action-setup` and `actions/setup-node`.
    - [x] Update testing commands to use `pnpm run` / `pnpm test`.
- [x] Task: Conductor - User Manual Verification 'Phase 3: CI/CD Pipeline Migration' (Protocol in workflow.md)
