# Plan: Support AGY as a Provider

**Track**: 003-support-agy
**Status**: DRAFT
**Spec**: [spec.md](spec.md)

---

## Overview

Add Google Antigravity (`agy`) as a selectable LLM provider. This follows the
exact same CLI-subprocess pattern as `opencode.go` and `codex.go` — spawn
`agy --print <prompt>`, stream stdout lines, log latency.

**Estimated effort**: ~3 hours (small track, single provider file + tests).

---

## Steps

### Step 1: Add Kind constant and update binding validation
**Files**: `internal/provider/binding.go`

- [ ] Add `KindAgy = "agy"` to the Kind constants block (after `KindExo`).
- [ ] Add `KindAgy` to the `ValidateKind` switch cases and error message string.
- [ ] Add `KindAgy` to the LLM list in `TestIsGatewayKind` (non-gateway side)
      and `TestValidateKind` fixtures in `binding_test.go`.

**Verification**: `go test ./internal/provider/ -run TestValidateKind -v`

---

### Step 2: Add config helper for AGY model
**Files**: `internal/config/config.go`

- [ ] Add `ResolveAgyModel() string` — reads `WUPHF_AGY_MODEL` env or config
      file value, returns empty string if unset.

**Verification**: `go test ./internal/config/ -run TestResolveAgyModel -v`

---

### Step 3: Create the AGY provider implementation
**Files**: `internal/provider/agy.go` (new)

- [ ] `init()` → `Register(&Entry{...})` with Kind `"agy"`.
- [ ] `CreateAgyCLIStreamFn(agentSlug string) agent.StreamFn`
  - LookPath `"agy"` → error chunk if missing.
  - Build prompt via `buildAgyPrompt(systemPrompt, prompt)`.
  - Build args via `buildAgyArgs(model, prompt)`:
    `["--print", "--model", model, prompt]` (omit `--model` if empty).
  - Run subprocess with `cmd.Dir = cwd`, `NO_COLOR=1` env.
  - Stream stdout via `DrainStreamLines` (reuse existing infra).
  - Latency log to `~/.wuphf/logs/agy-latency.log`.
- [ ] `RunAgyOneShot(systemPrompt, prompt, cwd string) (string, error)`
- [ ] `RunAgyOneShotCtx(ctx, systemPrompt, prompt, cwd) (string, error)`
- [ ] `buildAgyPrompt(systemPrompt, prompt string) string` — `<system>` wrapper,
      same zero-width-space escaping as Opencode.
- [ ] `buildAgyArgs(model, prompt string) []string`
- [ ] `describeAgyFailure(err error) string`
- [ ] `appendAgyLatencyLog(agentSlug, line string)`

**Verification**: `go build ./...` (compilation check)

---

### Step 4: Write provider tests
**Files**: `internal/provider/agy_test.go` (new)

- [ ] `TestBuildAgyArgs` — with/without model.
- [ ] `TestBuildAgyPrompt` — system+user, user-only, escaping.
- [ ] `TestDescribeAgyFailure` — auth error, generic error.
- [ ] `TestCreateAgyCLIStreamFn` — success path (mock stdout), error path
      (agy not found), exit-code error path.
- [ ] `TestRunAgyOneShot` — basic roundtrip with mock command.
- [ ] `TestRunAgyOneShotCtxCancel` — context cancellation.

**Verification**: `go test ./internal/provider/ -run TestAgy -v`

---

### Step 5: Update binding tests for new Kind
**Files**: `internal/provider/binding_test.go`

- [ ] Add `"agy"` to `TestValidateKind` valid-kind inputs.
- [ ] Add `KindAgy` to the LLM slice in `TestIsGatewayKind`.
- [ ] Add `KindAgy` to `TestLLMProviderKindsExcludesGateways` if it
      has a hardcoded expected-kind list.

**Verification**: `go test ./internal/provider/ -run TestValidateKind -v`

---

### Step 6: Update registry tests
**Files**: `internal/provider/registry_test.go`

- [ ] Bump any hardcoded kind-count assertions to include the new `agy` entry.

**Verification**: `go test ./internal/provider/ -v`

---

### Step 7: Full test suite verification
- [ ] `go test ./internal/...` — all backend tests pass.
- [ ] `pnpm run broker:test` — broker integration tests.

---

### Step 8: Update documentation
**Files**: `README.md`, `conductor/tech-stack.md`

- [ ] Add `agy` to the supported providers list in README.
- [ ] Add AGY mention in tech-stack.md provider section (if one exists).

---

## Risk Assessment

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| `agy` CLI not installed on user machine | Medium | Graceful error message with install instructions |
| OAuth token expired/missing | Low | `describeAgyFailure` detects auth errors and gives clear guidance |
| AGY CLI output format changes | Low | Plain-text stdout is stable; `DrainStreamLines` handles any line length |

## Commit Plan

Single conventional commit:
```
feat(provider): add AGY (Google Antigravity) as a selectable LLM provider

Register "agy" Kind, implement CLI-subprocess streaming (same pattern as
opencode), add one-shot support, config.ResolveAgyModel(), and full test
coverage.
```
