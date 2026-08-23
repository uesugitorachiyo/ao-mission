# Evidence-Bound Slice Checkpoints Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend AO Mission checkpoint creation so a retained, validated passing artifact can produce one ordered, idempotent S01-S07 slice checkpoint without changing lifecycle state or authority.

**Architecture:** Add a strict `SliceCheckpointOptions` path beside legacy checkpoint creation. It resolves one digest-addressed retained artifact, rehashes and validates its public-safe JSON fields, enforces slice order and replay consistency, and appends the existing v0.3 checkpoint shape through the current recoverable store transaction.

**Tech Stack:** Go standard library, existing AO Mission store and retained-artifact primitives, table-driven Go tests, PowerShell-compatible CLI verification.

---

### Task 1: Evidence Validation And S01 Append

**Files:**
- Modify: `internal/mission/checkpoint_resume.go`
- Modify: `internal/mission/terminal_reconciliation_cli_test.go`

- [ ] **Step 1: Write failing retained-evidence tests**

Add a `writeSliceCheckpointEvidence` helper that builds strict JSON containing
`schema`, `correlation_id`, `mission_ref`, `slice`, `result`, and an
all-false `authority` object; retains the bytes with `store.retainArtifact`;
and appends the resulting exact `ArtifactRef` through `Store.Update`.

Add tests proving `CreateSliceCheckpoint(store, missionID,
SliceCheckpointOptions{Slice: "S01", EvidenceDigest: digest})` appends one
checkpoint with result `slice_pass:S01:<digest>`. Capture the record before
creation and assert status, route, phase, blockers, steps, route history, exact
next action, artifact refs, and correlated imports remain unchanged after
clearing only checkpoints and update timestamp.

- [ ] **Step 2: Run the focused test and observe RED**

```powershell
go test ./internal/mission -run 'TestCreateSliceCheckpoint' -count=1
```

Expected: compilation fails because `SliceCheckpointOptions` and
`CreateSliceCheckpoint` do not exist.

- [ ] **Step 3: Implement minimal strict S01 creation**

In `checkpoint_resume.go`, add:

```go
type SliceCheckpointOptions struct {
    Slice          string
    EvidenceDigest string
}
```

Add anchored patterns for S01-S07 and prefixed lowercase SHA-256. Resolve
exactly one matching `ArtifactRef`; require its `ContentRef` to equal the
store-owned digest path; read it with retained-artifact no-follow primitives;
rehash it; and strictly decode its top-level identity/result plus required
authority flags. Require matching correlation, slice, `result == "pass"`,
and every authority flag false. Append through `Store.Update` using
`appendMissionCheckpoint`.

- [ ] **Step 4: Run focused tests and observe GREEN**

Run the Task 1 focused command. Expected: all named tests pass.

- [ ] **Step 5: Commit Task 1**

```powershell
git add internal/mission/checkpoint_resume.go internal/mission/terminal_reconciliation_cli_test.go
git commit -m "feat: add evidence-bound slice checkpoints"
```

### Task 2: Replay, Conflict, And Slice Ordering

**Files:**
- Modify: `internal/mission/checkpoint_resume.go`
- Modify: `internal/mission/terminal_reconciliation_cli_test.go`

- [ ] **Step 1: Write failing replay and ordering tests**

Cover exact S01 replay, conflicting S01 digest, S02 before S01, S02 after S01,
and S03 after S01. Exact replay must preserve count. Conflict errors must
contain `slice S01 already checkpointed with a different evidence digest`;
ordering errors must contain `slice checkpoint is out of order`.

- [ ] **Step 2: Run focused tests and observe RED**

Run the Task 1 focused command. Expected: replay, conflict, or ordering cases
fail against the initial implementation.

- [ ] **Step 3: Implement deterministic replay and order enforcement**

Scan checkpoint results matching the slice-result format. Return without
mutation for the same slice/digest; reject the same slice with another digest.
Require S01 with no prior slice checkpoint, otherwise the immediately following
numeric slice. Reject progress after S07.

- [ ] **Step 4: Run focused tests and observe GREEN**

Run the Task 1 focused command. Expected: all slice checkpoint tests pass.

- [ ] **Step 5: Commit Task 2**

```powershell
git add internal/mission/checkpoint_resume.go internal/mission/terminal_reconciliation_cli_test.go
git commit -m "fix: enforce slice checkpoint ordering"
```

### Task 3: Fail-Closed Evidence Matrix

**Files:**
- Modify: `internal/mission/checkpoint_resume.go`
- Modify: `internal/mission/terminal_reconciliation_cli_test.go`

- [ ] **Step 1: Write failing validation cases**

Use table-driven tests for empty/unknown slice, uppercase/bare/short digest,
missing artifact ref, absent or mismatched content ref, retained-byte drift,
duplicate JSON keys, wrong correlation, wrong slice, non-pass result, missing
authority key, true authority key, unknown authority key, authority-name case
variant, nested true authority, and oversized evidence. For every case assert
the error substring and unchanged record/checkpoint state.

- [ ] **Step 2: Run focused tests and observe RED**

Run the Task 1 focused command. Expected: the new negative cases fail.

- [ ] **Step 3: Complete strict validation**

Use bounded no-follow reads and the strict duplicate-key JSON decoder. Require
the common top-level schema, correlation, mission reference, slice, result, and
authority fields while allowing producer-specific evidence fields. Require the
exact authority object named in the design and recursively reject any known
authority key that is true or case-variant before `Store.Update`.

- [ ] **Step 4: Run focused and package tests**

```powershell
go test ./internal/mission -run 'TestCreateSliceCheckpoint' -count=1
go test ./internal/mission -count=1
```

Expected: both commands pass.

- [ ] **Step 5: Commit Task 3**

```powershell
git add internal/mission/checkpoint_resume.go internal/mission/terminal_reconciliation_cli_test.go
git commit -m "test: fail closed on invalid slice evidence"
```

### Task 4: CLI, Documentation, And Integration

**Files:**
- Modify: `internal/mission/cli.go`
- Modify: `internal/mission/terminal_reconciliation_cli_test.go`
- Modify: `README.md`
- Modify: `AGENTS.md`

- [ ] **Step 1: Write failing CLI tests**

Prove paired flags are required, malformed inputs return nonzero without
standard output, valid S01 JSON reports the new count/result, exact replay is
idempotent, and text output retains false-authority lines.

- [ ] **Step 2: Run CLI-focused tests and observe RED**

```powershell
go test ./internal/mission -run 'TestCheckpointCreate.*Slice' -count=1
```

Expected: flag parsing fails because the new flags are unknown.

- [ ] **Step 3: Route paired CLI flags**

Add `--slice` and `--evidence-digest` only to `checkpoint create`. Reject
one without the other. Route paired values to `CreateSliceCheckpoint`;
otherwise retain `CreateMissionCheckpoint`. Keep inspect unchanged.

- [ ] **Step 4: Update durable command documentation**

Add the paired form to `README.md`. Add an `AGENTS.md` rule that slice
checkpoints require retained passing evidence, exact ordering, and no lifecycle
or authority mutation.

- [ ] **Step 5: Run full Mission gates**

```powershell
gofmt -w internal/mission/checkpoint_resume.go internal/mission/cli.go internal/mission/terminal_reconciliation_cli_test.go
go test ./internal/mission -count=1
go test ./... -count=1
go vet ./...
go build ./cmd/ao-mission
python3 scripts/test_public_safety_scan.py
python3 ../ao-architecture/scripts/verify_agent_instruction_layout.py --workspace-root .. --repository ao-mission
git diff --check
```

Expected: every command exits zero and instruction layout is valid.

- [ ] **Step 6: Commit integration**

```powershell
git add AGENTS.md README.md internal/mission/checkpoint_resume.go internal/mission/cli.go internal/mission/terminal_reconciliation_cli_test.go
git commit -m "feat: expose evidence-bound slice checkpoints"
```

### Task 5: Apply The S01 Checkpoint

**Files:**
- Generated only: private campaign Mission state outside source checkouts

- [ ] **Step 1: Re-run committed-source verification**

Run the Task 4 full gates after the integration commit and record exact output.

- [ ] **Step 2: Create S01 from retained evidence**

```powershell
go run ./cmd/ao-mission --home $env:AO_MISSION_HOME checkpoint create --mission mission-0d10a1a990af0fdc --slice S01 --evidence-digest sha256:017d3d08c6003b78158549f712ff863fa277c47085abb4149bbb4dab0c8f0cc2 --json
```

Expected: checkpoint count 2 and latest result binds S01 to the exact retained
evidence digest.

- [ ] **Step 3: Verify idempotent Mission and Command readbacks**

Repeat the command, inspect checkpoint, Mission status, Command status, and
artifacts. Require count 2, exact correlation, unchanged active/routing/ao-atlas
state, two retained evidence artifacts, and false execution, approval, and
mutation flags.

- [ ] **Step 4: Continue to S02 only after the S01 evidence gate passes**

Record the checkpoint result privately. Do not start S02 if any digest, count,
identity, state, or authority field differs.
