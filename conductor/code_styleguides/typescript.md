# TypeScript Code Style Guide

## 1. Tooling & Formatting
* Align with the project's Biome config rules (`biome.json`). Use spaces, semi-colons, and single quotes.
* Avoid raw type assertions (`as any`) unless absolutely necessary for low-level protocol parsing.

## 2. Type Safety
* Enforce strict type definitions for all function signatures and React component props.
* Use Discriminated Unions for modeling complex agent messages, states, and events.

## 3. React & Frontend Conventions
* Prefer functional components with hooks.
* Group files by feature/workspace (e.g., `web/src/components/apps/RoutinesApp.tsx`).
* Avoid inline styles; use CSS Modules or vanilla CSS classes inside dedicated stylesheet files.

## 4. Workspaces & Imports
* Use clean imports with exact paths (e.g., `@/components/...`).
* Do not import cross-workspace internals directly; go through the shared packages/SDK exports.
