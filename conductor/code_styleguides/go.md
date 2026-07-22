# Go Code Style Guide

## 1. Code Formatting & Tooling
* Always format code using `gofmt` and organize imports using `goimports`.
* Code must pass all rules in `.golangci.yml` lint configuration.

## 2. Naming Conventions
* Use short but descriptive variable names within small scopes.
* Keep package names single-word, lower-case, and non-plural (e.g., `team`, not `teams`).
* Use MixedCaps (CamelCase) with initialisms capitalized (e.g., `APIURL`, not `ApiUrl`).

## 3. Error Handling
* Check errors immediately. Avoid ignoring errors using `_` unless explicitly justified.
* Wrap errors with context using `fmt.Errorf("doing action: %w", err)` to preserve debugging trace.
* Avoid panic for control flow; use panic only for unrecoverable startup issues.

## 4. Concurrency & Performance
* Always use `context.Context` for propagation, timeouts, and cancellation.
* Avoid global state or package-level variables that require synchronization. Protect shared memory with mutexes or channels.
* Ensure Goroutine lifetimes are bounded and clean up properly using `sync.WaitGroup` or context cancellation.
