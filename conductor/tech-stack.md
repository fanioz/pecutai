# Technology Stack

## 1. Programming Languages & Runtimes
* **Go (Golang)**: Version `1.25.11+`. Used for the root-level core backend, broker execution, agent terminal-based loop, and the main `wuphf` binary CLI command.
* **TypeScript & JavaScript (Node.js/Bun)**: Used for all web and client-side modules.
* **Bun**: The primary runtime, package manager, and test runner for frontend and workspace packages.

## 2. Core Frameworks & Applications
* **React**: Frontend UI framework for the web dashboard (`web/`).
* **Vite**: Build tool and dev server for React/TypeScript and Electron processes.
* **Electron**: Desktop container shell under `apps/desktop/` to run WUPHF locally as a desktop app.
* **Wails/Tmux (Legacy)**: Backward-compatible support for terminal multiplexing in older TUI modes.

## 3. Database, Cache & Search Indexing
* **SQLite**: Embedded database used for local relational storage, configurations, and fact catalogs (`modernc.org/sqlite`).
* **Bleve Search**: Go-based local text search index used for search queries and wiki indexing (`github.com/blevesearch/bleve/v2`).
* **Git**: System version control used to track changes, back up notebooks, and maintain the shared team wiki.

## 4. Developer Tooling & Integrations
* **Lefthook**: Git githooks manager.
* **Secretlint**: Secret scanning tool.
* **Commitlint**: Linting commit messages.
