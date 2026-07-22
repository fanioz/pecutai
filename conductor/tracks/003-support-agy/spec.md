# Spec: Support AGY as a Provider

**Track**: 003-support-agy
**Status**: DRAFT
**Author**: AI Conductor

---

## Problem

WUPHF currently supports several LLM runtimes (Claude Code, Codex, Opencode,
Ollama, MLX-LM, Exo) but has no integration with Google Antigravity (AGY).
AGY exposes a CLI (`agy`) with `--print` (non-interactive) and `--prompt`
(alias) modes and uses OAuth from the current shell — no separate API key
management is needed.

Users who already have `agy` authenticated in their terminal should be able
to route any WUPHF agent through AGY, just as they can with `opencode` or
`codex` today.

## Requirements

### R1: New Provider Kind `agy`
- Register `KindAgy = "agy"` in `binding.go` constants.
- Add `agy` to `ValidateKind` switch and error message.
- Add `agy` to test fixtures (`TestIsGatewayKind`, `TestValidateKind`, etc.).

### R2: Provider Entry Registration
- Create `internal/provider/agy.go` with `init()` → `Register(&Entry{...})`.
- `StreamFn`: `CreateAgyCLIStreamFn(agentSlug) agent.StreamFn`.
- `OneShot`: `RunAgyOneShot(systemPrompt, prompt, cwd string) (string, error)`.
- `OneShotCtx`: `RunAgyOneShotCtx(ctx, systemPrompt, prompt, cwd) (string, error)`.
- Capabilities: `PaneEligible: false`, `SupportsOneShot: true`.

### R3: CLI Invocation Pattern
- Binary: `agy` (resolved via `runtimebin.LookPath`).
- Non-interactive flag: `--print` (equivalent to Opencode `run`).
- Model override: `--model <model>` (optional, from config).
- System prompt: prepended in `<system>...</system>` wrapper (same
  pattern as `buildOpencodePrompt`).
- Prompt: passed as trailing positional argument.
- Env: `NO_COLOR=1` to suppress ANSI in output.
- Working directory: `cmd.Dir = cwd`.
- Auth: inherits OAuth from shell environment — no API key env var needed.

### R4: Stream Reading
- AGY emits plain text on stdout (no JSONL), same as Opencode.
- Reuse `DrainStreamLines` + `readAgyStream` (same pattern as Opencode).
- Latency logging to `~/.wuphf/logs/agy-latency.log`.

### R5: Config Integration
- `config.ResolveAgyModel() string` — reads `WUPHF_AGY_MODEL` env var or
  config file field, defaults to empty (AGY uses its own default).
- Provider selectable via `--provider agy` CLI flag and `WUPHF_LLM_PROVIDER=agy`
  env var.

### R6: Tests
- `agy_test.go`: unit tests for `buildAgyArgs`, `buildAgyPrompt`,
  `describeAgyFailure`, stream reading (mock stdout), and the full
  `CreateAgyCLIStreamFn` with injected command stubs.
- Update `binding_test.go` to include `KindAgy` in validation and
  gateway-exclusion fixtures.
- Update `registry_test.go` if it has hardcoded kind counts.

## Non-Goals
- No custom tool/prompt formatting for AGY — standard OpenAI/Gemini-compatible
  text prompts only.
- No AGY SDK integration — CLI subprocess only (same as Opencode/Codex pattern).
- No interactive TUI pane for AGY (`PaneEligible: false`).

## Decision Log
| # | Decision | Rationale |
|---|----------|-----------|
| 1 | CLI subprocess, not SDK | Consistent with Opencode/Codex pattern; AGY OAuth "just works" from shell |
| 2 | `--print` flag | AGY's non-interactive mode, equivalent to `opencode run` |
| 3 | Reuse `DrainStreamLines` | AGY emits plain text stdout — identical to Opencode stream format |
| 4 | No gateway kind | AGY is a directly-runnable LLM runtime, not an imported gateway |
