# Support Claude Opus 4.6 & Sonnet 4.6 Natively

## TL;DR

> **Quick Summary**: Remove the temporary opus-4-6→opus-4-5 fallback routing and add complete Sonnet 4.6 model support (base + 4 thinking variants) to the Claude provider. Both models will work via the existing CLI-mimicking executor.
> 
> **Deliverables**:
> - Opus 4.6 requests go directly to Claude API (no more fallback to 4.5)
> - Sonnet 4.6 fully registered: model definitions, alias resolution, pricing
> - Sonnet 4.6 thinking variants: `-thinking`, `-thinking-low`, `-thinking-medium`, `-thinking-high`
> 
> **Estimated Effort**: Quick
> **Parallel Execution**: NO — single tightly-coupled task across 4 files
> **Critical Path**: Edit 4 files → build → test

---

## Context

### Original Request
User wants native support for Opus 4.6 and Sonnet 4.6. Opus 4.6 model definitions already exist but are temporarily routed to Opus 4.5 via fallback config. Sonnet 4.6 is completely missing. Approach: mimic the pattern used for Opus 4.5 — the Claude executor already sends requests with CLI-signature headers.

### Interview Summary
**Key Discussions**:
- Opus 4.6 fallback was a temp fix while model wasn't available — now remove it
- Sonnet 4.6: use alias `claude-sonnet-4-6` directly (no dated model ID needed)
- All 4 thinking variants for sonnet 4.6 (diverges from sonnet-4-5 which only has 1)
- Claude provider only — no Copilot/Kiro/Gemini for now

### Metis Review
**Identified Gaps** (addressed):
- Sonnet 4.5 only has 1 thinking variant; sonnet 4.6 having 4 is an intentional pattern divergence (user explicitly confirmed)
- `resolveUpstreamModel` for base `claude-sonnet-4-6` must NOT have an entry (passes through as-is); only thinking variants need resolution
- Pricing prefix matcher `claude-sonnet-4` already covers sonnet-4-6 — explicit entry is nice-to-have for clarity
- Pre-existing opus-4-6 Created timestamp bug (1739232000 = Feb 2025, should be 2026) — out of scope

---

## Work Objectives

### Core Objective
Enable native Opus 4.6 and Sonnet 4.6 support through the Claude provider by removing the temporary fallback and adding complete Sonnet 4.6 model registration.

### Concrete Deliverables
- `internal/config/config.go` — Fallback map cleared (opus-4-6 routes natively)
- `internal/registry/model_definitions.go` — 5 new sonnet-4-6 entries
- `internal/runtime/executor/providers/claude.go` — resolveUpstreamModel handles sonnet-4-6 thinking variants
- `internal/usage/pricing.go` — Sonnet 4.6 pricing entry

### Definition of Done
- [x] `go build -o /dev/null ./cmd/server` exits 0
- [x] `go test ./...` exits 0
- [x] Opus 4.6 requests no longer fall back to opus 4.5
- [x] Sonnet 4.6 appears in model registry with all 5 entries

### Must Have
- Remove ALL 5 fallback entries in config.go
- 5 model definitions for sonnet 4.6 (base + 4 thinking)
- resolveUpstreamModel entries for 4 thinking variants → `claude-sonnet-4-6`
- Pricing entry for `claude-sonnet-4-6`

### Must NOT Have (Guardrails)
- DO NOT modify any existing model entries (opus-4-6, opus-4-5, sonnet-4-5, etc.)
- DO NOT touch Kiro, Copilot, Gemini-CLI, or Antigravity model registries
- DO NOT modify `ParseClaudeThinkingFromModel`, `EnsureClaudeMaxTokens`, or `applyClaudeHeaders` — they work generically
- DO NOT add a `resolveUpstreamModel` entry for `claude-sonnet-4-6` base (it passes through as-is)
- DO NOT fix the pre-existing opus-4-6 Created timestamp bug
- DO NOT add sonnet-4-5 thinking-low/medium/high variants — separate PR
- DO NOT update documentation/README unless explicitly asked

---

## Verification Strategy

> **ZERO HUMAN INTERVENTION** — ALL verification is agent-executed.

### Test Decision
- **Infrastructure exists**: YES
- **Automated tests**: Tests-after (verify existing tests still pass)
- **Framework**: `go test`
- **No new test files needed** — changes are config/registry, verified by build + existing tests

### QA Policy
Every task includes agent-executed QA scenarios.
Evidence saved to `.sisyphus/evidence/task-{N}-{scenario-slug}.{ext}`.

- **Build verification**: Use Bash (`go build`)
- **Test verification**: Use Bash (`go test ./...`)
- **Content verification**: Use Bash (`grep`)

---

## Execution Strategy

### Parallel Execution Waves

```
Wave 1 (Single task — all changes tightly coupled):
└── Task 1: Add opus 4.6 + sonnet 4.6 native support [quick]

Wave FINAL (After task 1 — independent review):
└── Task F1: Build + test verification [quick]
```

### Dependency Matrix
- **1**: None — can start immediately
- **F1**: Depends on 1

### Agent Dispatch Summary
- **Wave 1**: 1 task → `quick`
- **FINAL**: 1 task → `quick`

---

## TODOs

- [x] 1. Add native Opus 4.6 & Sonnet 4.6 support across 4 files

  **What to do**:

  **File 1: `internal/config/config.go`** — Remove fallback routing
  - Delete the 5 fallback entries at lines 309-314 (the `claude-opus-4-6-*` → `claude-opus-4-5-*` mappings)
  - The comment on line 309 should also be removed
  - Leave the `Fallbacks` map as empty: `Fallbacks: map[string][]string{}`
  - Verify `Init()` on line 318 still works with empty map (it does — `hasFallbacks = len(r.Fallbacks) > 0`)

  **File 2: `internal/registry/model_definitions.go`** — Add Sonnet 4.6 model entries
  - Insert 5 new entries in `GetClaudeModels()` after the existing sonnet-4-5 entries (after line 11)
  - Follow the exact pattern of opus-4-6 entries (lines 12-16) but for sonnet:
    ```go
    Claude("claude-sonnet-4-6").Display("Claude 4.6 Sonnet").Desc("High-performance Claude model with advanced reasoning").Created(1739232000).Canonical("claude-sonnet-4-6").Context(200000, 64000).B(),
    Claude("claude-sonnet-4-6-thinking").Display("Claude 4.6 Sonnet Thinking").Created(1739232000).Context(200000, 64000).Thinking(1024, 100000).B(),
    Claude("claude-sonnet-4-6-thinking-low").Display("Claude 4.6 Sonnet Thinking Low").Created(1739232000).Context(200000, 64000).Thinking(1024, 100000).B(),
    Claude("claude-sonnet-4-6-thinking-medium").Display("Claude 4.6 Sonnet Thinking Medium").Created(1739232000).Context(200000, 64000).Thinking(1024, 100000).B(),
    Claude("claude-sonnet-4-6-thinking-high").Display("Claude 4.6 Sonnet Thinking High").Created(1739232000).Context(200000, 64000).Thinking(1024, 100000).B(),
    ```

  **File 3: `internal/runtime/executor/providers/claude.go`** — Add resolveUpstreamModel case
  - Add a new case block in `resolveUpstreamModel()` after line 439 (after the sonnet-4-5 case):
    ```go
    case "claude-sonnet-4-6-thinking", "claude-sonnet-4-6-thinking-low", "claude-sonnet-4-6-thinking-medium", "claude-sonnet-4-6-thinking-high":
        return "claude-sonnet-4-6"
    ```
  - IMPORTANT: Do NOT add `"claude-sonnet-4-6"` to this case — the base model passes through as-is (returns `""` from the function, so the original model name is used)

  **File 4: `internal/usage/pricing.go`** — Add Sonnet 4.6 pricing
  - Add after line 18 (after sonnet-4-5 entry):
    ```go
    "claude-sonnet-4-6":            {InputPer1M: 3.00, OutputPer1M: 15.00, CachedPer1M: 0.30},
    ```
  - Note: prefix matcher already covers `claude-sonnet-4-6*` via `claude-sonnet-4` prefix, but explicit entry ensures exact match

  **Must NOT do**:
  - Do NOT modify any existing model entries
  - Do NOT touch thinking.go, applyClaudeHeaders, or ParseClaudeThinkingFromModel
  - Do NOT add Copilot/Kiro/Gemini entries
  - Do NOT add resolveUpstreamModel entry for the base `claude-sonnet-4-6`

  **Recommended Agent Profile**:
  - **Category**: `quick`
    - Reason: Well-defined changes across 4 files following established patterns — no ambiguity
  - **Skills**: []
    - No special skills needed — straightforward Go code edits
  - **Skills Evaluated but Omitted**:
    - `backend-development`: Overkill for config/registry changes

  **Parallelization**:
  - **Can Run In Parallel**: NO (single task)
  - **Parallel Group**: Wave 1 (sole task)
  - **Blocks**: F1
  - **Blocked By**: None

  **References** (CRITICAL):

  **Pattern References** (existing code to follow):
  - `internal/registry/model_definitions.go:12-16` — Opus 4.6 model entries (EXACT pattern to replicate for sonnet)
  - `internal/registry/model_definitions.go:10-11` — Sonnet 4.5 entries (shows the 1-thinking-variant pattern to diverge from)
  - `internal/runtime/executor/providers/claude.go:434-435` — Opus 4.6 resolveUpstreamModel case (pattern to follow)
  - `internal/runtime/executor/providers/claude.go:438-439` — Sonnet 4.5 resolveUpstreamModel case (has only 2 aliases — sonnet 4.6 will have 4)
  - `internal/config/config.go:307-315` — The fallback routing to remove (ALL 5 entries + comment)
  - `internal/usage/pricing.go:17-18` — Sonnet 4/4.5 pricing entries (same pricing tier for 4.6)

  **WHY Each Reference Matters**:
  - `model_definitions.go:12-16`: Copy this exact structure for sonnet entries — `.Display()`, `.Created()`, `.Canonical()`, `.Context()`, `.Thinking()` calls
  - `claude.go:434-435`: Shows the case-grouping pattern — all thinking variants in one case, returning the base dated model
  - `config.go:307-315`: This is the EXACT code to delete — verify ALL 5 entries are removed
  - `pricing.go:17-18`: Shows the sonnet pricing tier — $3/$15/$0.30

  **Acceptance Criteria**:

  **QA Scenarios (MANDATORY):**

  ```
  Scenario: Build succeeds after all changes
    Tool: Bash
    Preconditions: All 4 files edited
    Steps:
      1. Run `go build -o /dev/null ./cmd/server`
      2. Assert exit code is 0
    Expected Result: Clean build, exit code 0
    Failure Indicators: Compilation errors, non-zero exit code
    Evidence: .sisyphus/evidence/task-1-build-success.txt

  Scenario: All existing tests pass
    Tool: Bash
    Preconditions: Build succeeds
    Steps:
      1. Run `go test ./...`
      2. Assert exit code is 0
      3. Assert output contains "ok" for all packages, no "FAIL"
    Expected Result: All tests pass, zero failures
    Failure Indicators: Any FAIL lines, non-zero exit code
    Evidence: .sisyphus/evidence/task-1-test-pass.txt

  Scenario: Fallback entries fully removed
    Tool: Bash
    Preconditions: config.go edited
    Steps:
      1. Run `grep -c "claude-opus-4-6.*claude-opus-4-5" internal/config/config.go`
      2. Assert count is 0
    Expected Result: Zero matches — no fallback routing remains
    Failure Indicators: Count > 0 means fallback entries still present
    Evidence: .sisyphus/evidence/task-1-fallback-removed.txt

  Scenario: Sonnet 4.6 model entries present
    Tool: Bash
    Preconditions: model_definitions.go edited
    Steps:
      1. Run `grep -c "sonnet-4-6" internal/registry/model_definitions.go`
      2. Assert count is 5 (base + 4 thinking variants)
    Expected Result: Exactly 5 sonnet-4-6 entries in model definitions
    Failure Indicators: Count != 5
    Evidence: .sisyphus/evidence/task-1-sonnet-entries.txt

  Scenario: resolveUpstreamModel handles sonnet-4-6-thinking variants
    Tool: Bash
    Preconditions: claude.go edited
    Steps:
      1. Run `grep -c "sonnet-4-6" internal/runtime/executor/providers/claude.go`
      2. Assert count is at least 5 (4 thinking aliases + 1 return value)
      3. Run `grep "claude-sonnet-4-6-thinking" internal/runtime/executor/providers/claude.go`
      4. Assert output contains all 4 thinking variant aliases in one case block
    Expected Result: All 4 thinking variants resolve to "claude-sonnet-4-6"
    Failure Indicators: Missing variants or incorrect return value
    Evidence: .sisyphus/evidence/task-1-resolve-model.txt

  Scenario: Pricing entry exists for sonnet-4-6
    Tool: Bash
    Preconditions: pricing.go edited
    Steps:
      1. Run `grep "sonnet-4-6" internal/usage/pricing.go`
      2. Assert output contains `claude-sonnet-4-6` with pricing `3.00, 15.00, 0.30`
    Expected Result: Explicit pricing entry for claude-sonnet-4-6
    Failure Indicators: No match or wrong pricing values
    Evidence: .sisyphus/evidence/task-1-pricing.txt
  ```

  **Evidence to Capture:**
  - [ ] task-1-build-success.txt
  - [ ] task-1-test-pass.txt
  - [ ] task-1-fallback-removed.txt
  - [ ] task-1-sonnet-entries.txt
  - [ ] task-1-resolve-model.txt
  - [ ] task-1-pricing.txt

  **Commit**: YES
  - Message: `feat(claude): add native opus 4.6 & sonnet 4.6 support`
  - Files: `internal/config/config.go`, `internal/registry/model_definitions.go`, `internal/runtime/executor/providers/claude.go`, `internal/usage/pricing.go`
  - Pre-commit: `go test ./...`

---

## Final Verification Wave

- [x] F1. **Build + Test + Content Verification** — `quick`
  Run `go build -o /dev/null ./cmd/server` + `go test ./...`. Verify grep counts: 5 sonnet-4-6 in model_definitions.go, 0 opus-4-6→opus-4-5 fallbacks in config.go, 5+ sonnet-4-6 in claude.go, 1+ sonnet-4-6 in pricing.go.
  Output: `Build [PASS/FAIL] | Tests [N pass/N fail] | Content [N/N checks pass] | VERDICT`

---

## Commit Strategy

- **1**: `feat(claude): add native opus 4.6 & sonnet 4.6 support` — config.go, model_definitions.go, claude.go, pricing.go — `go test ./...`

---

## Success Criteria

### Verification Commands
```bash
go build -o /dev/null ./cmd/server  # Expected: exit 0
go test ./...                        # Expected: all PASS
grep -c "sonnet-4-6" internal/registry/model_definitions.go  # Expected: 5
grep -c "claude-opus-4-6.*claude-opus-4-5" internal/config/config.go  # Expected: 0
grep -c "sonnet-4-6" internal/runtime/executor/providers/claude.go  # Expected: >= 5
grep "sonnet-4-6" internal/usage/pricing.go  # Expected: 1 line with 3.00/15.00/0.30
```

### Final Checklist
- [x] All "Must Have" present
- [x] All "Must NOT Have" absent
- [x] All tests pass
- [x] Fallback routing completely removed
- [x] Sonnet 4.6 fully registered with 5 model entries
