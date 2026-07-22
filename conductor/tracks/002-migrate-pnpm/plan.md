# Implementation Plan: Migrate Project from Bun to pnpm

## Phase 1: Workspace & Lockfile Conversion
- [ ] Task: Remove old lockfiles and configure pnpm workspace
    - [ ] Delete `bun.lock` at workspace root.
    - [ ] Create `pnpm-workspace.yaml` defining workspaces (`packages/*`, `apps/*`).
- [ ] Task: Install dependencies with pnpm
    - [ ] Run `pnpm install` at workspace root.
    - [ ] Address any package resolution or hoisting issues.
- [ ] Task: Conductor - User Manual Verification 'Phase 1: Workspace & Lockfile Conversion' (Protocol in workflow.md)

## Phase 2: Configuration & Scripts Migration
- [ ] Task: Update package.json scripts
    - [ ] Replace `bun run --cwd` scripts with `pnpm --filter` scripts in root `package.json`.
    - [ ] Verify script compatibility inside `packages/*` and `apps/*`.
- [ ] Task: Verify Vitest tests
    - [ ] Execute `pnpm run broker:test` synchronously.
    - [ ] Execute `pnpm run desktop:test` synchronously.
- [ ] Task: Conductor - User Manual Verification 'Phase 2: Configuration & Scripts Migration' (Protocol in workflow.md)

## Phase 3: CI/CD Pipeline Migration
- [ ] Task: Update GitHub Actions workflows
    - [ ] Replace `setup-bun` actions with `pnpm/action-setup` and `actions/setup-node`.
    - [ ] Update testing commands to use `pnpm run` / `pnpm test`.
- [ ] Task: Conductor - User Manual Verification 'Phase 3: CI/CD Pipeline Migration' (Protocol in workflow.md)
