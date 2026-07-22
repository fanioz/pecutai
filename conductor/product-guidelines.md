# Product Guidelines

## 1. Prose & Communication Style
* **Direct & Objective**: Write like high-quality developer documentation. Avoid marketing adjectives ("revolutionary", "magical", "powerful").
* **Playful yet Serious**: The tone can lean into the "office" and "WUPHF" themes (e.g., using "desk", "brain", "employees") but never compromises technical correctness.
* **Command & Output First**: Explain options by showing code examples, exact CLI flags, or mock outputs.

## 2. Brand & Visual Identity
* **Local-First Retro/Terminal Aesthetic**: Lean into terminal-style outputs, simple pixel layouts, clean grid designs, and monospace fonts.
* **Minimalist Gradients**: Reject standard SaaS gradients and glowing glassmorphism in favor of functional status indicators, sharp borders, and clean layouts.
* **Visual Workflows**: Emphasize visible actions (agents typing, calling tools, resolving conflicts) to keep the "shared brain" tangible.

## 3. User Experience (UX) & Design Principles
* **Onboarding Friction is Zero**: Installing the workspace should be one command: `npx wuphf`. No database configurations, API keys, or backend selections required by default.
* **Agent Transparency**: Every agent's thoughts, tools called, and outputs should be fully visible, inspectable, and logged in real-time.
* **Offline-Ready / Local-First**: The system should run fully on local SQLite, local text index (Bleve), and git-based wikis.
* **Collaborative Control**: Users must be able to interrupt, redirect (`/focus`), or chat 1:1 with any agent in the workspace.
