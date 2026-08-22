# AO Cross-Platform Development Baseline Campaign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking. Do not use subagents unless the operator explicitly authorizes them.

**Goal:** Coordinate a manifest-bound AO development baseline that materializes and behaves equivalently on clean macOS and Windows hosts before AO Office Pool development begins.

**Architecture:** AO Mission owns the durable campaign ledger and sequential checkpoints but performs no repository mutation itself. AO Architecture owns the baseline contract, materializer, verifier, cross-platform runner, comparator, and final declaration; owning AO repositories receive only evidence-backed portability fixes discovered by the campaign. The work is split into seven independently reviewed slices so routine work can use economical models and failures cannot contaminate later qualification.

**Tech Stack:** AO Mission, AO Blueprint, AO Atlas, Python 3 standard library, JSON Schema, Git, Bash, Windows PowerShell 5.1, PowerShell 7, GitHub Actions, SHA-256, repository-native Go/Rust/Node test gates.

**Design:** `ao-architecture/docs/superpowers/specs/2026-08-22-cross-platform-development-baseline-design.md` at Architecture commit `f8c8c1d`, resolved from the common AO workspace root.

---

## Campaign Boundaries

- Do not begin AO Office Pool development.
- Do not create a universal AO installer or wrapper CLI.
- Do not publish releases unless a separately reviewed portability fix proves a patch release is unavoidable and the operator grants fresh release authority.
- Keep provider execution, deployment, promotion, compatibility activation, external beta, credentials, and RSI denied.
- Preserve completed Windows-remediation evidence unchanged. New readiness claims must use new self-contained evidence.
- Use new empty qualification roots and new Mission/Atlas identities. Never reuse consumed run roots.
- Use correlation ID `ao-cross-platform-development-baseline-20260822-r2` and a new private, empty `AO_MISSION_HOME` outside every source checkout and previous campaign root. Preserve earlier intake readbacks only as diagnostic evidence; never resume or import them.
- Mission coordinates and reconciles; it does not execute repository mutations or infer approval from readiness.
- Execute slices sequentially. A failed slice blocks later slices until its owning fix is reviewed and the baseline identity is regenerated.

## Model And Token Policy

- Default campaign model: `gpt-5.6-terra`, reasoning effort `medium`.
- Mechanical inventory, hashing, schema-format checks, and status readbacks may use `gpt-5.6-luna`, effort `low`, only when acceptance criteria and exact inputs are already frozen.
- Escalate one slice to `gpt-5.6-sol`, effort `high`, only for an unexplained cross-repository semantic mismatch, unsafe cleanup ambiguity, or final reconciliation with contradictory evidence.
- Do not use `max` or `ultra`. Do not spawn subagents without explicit operator authorization.
- Do not repeatedly reread the full repository set. Each slice reads its own `AGENTS.md`, the design, the current campaign checkpoint, and only its owning files.
- Store compact command summaries and hashes in Mission. Keep full logs in digest-addressed artifacts outside Git.

### Task 1: Start The Durable Mission And Bind The Design

**Files:**
- Read: `docs/roadmap/handoffs/active/ao-cross-platform-development-baseline-handoff-prompt.md`
- Read: `../ao-architecture/docs/superpowers/specs/2026-08-22-cross-platform-development-baseline-design.md`
- Generated only: operator-selected `AO_MISSION_HOME`

- [ ] **Step 1: Verify the design identity**

Run from AO Architecture:

```powershell
git cat-file -e f8c8c1d^{commit}
git show f8c8c1d:docs/superpowers/specs/2026-08-22-cross-platform-development-baseline-design.md |
  git hash-object --stdin
```

Expected: the commit exists and the blob digest is recorded in the Mission intake evidence.

- [ ] **Step 2: Start one correlation-bound Mission**

Run from the AO Mission checkout containing this reviewed plan. First require a
tracked-clean exact source commit and a new operator-selected private root:

```powershell
git diff --quiet
if ($LASTEXITCODE -ne 0) { throw "AO Mission tracked worktree is dirty" }
git diff --cached --quiet
if ($LASTEXITCODE -ne 0) { throw "AO Mission index is dirty" }
$missionSourceCommit = git rev-parse HEAD
if ($LASTEXITCODE -ne 0) { throw "cannot resolve AO Mission source commit" }
if (-not $env:AO_BASELINE_CAMPAIGN_ROOT) { throw "AO_BASELINE_CAMPAIGN_ROOT is required" }
$campaignRoot = [System.IO.Path]::GetFullPath($env:AO_BASELINE_CAMPAIGN_ROOT)
$sourceRoot = [System.IO.Path]::GetFullPath((Get-Location).Path)
if ($campaignRoot.StartsWith($sourceRoot + [System.IO.Path]::DirectorySeparatorChar, [System.StringComparison]::OrdinalIgnoreCase)) {
  throw "campaign root must be outside the AO Mission checkout"
}
if (Test-Path -LiteralPath $campaignRoot) { throw "campaign root already exists" }
[void](New-Item -ItemType Directory -Path $campaignRoot)
$env:AO_MISSION_HOME = Join-Path $campaignRoot "mission-state"
[void](New-Item -ItemType Directory -Path $env:AO_MISSION_HOME)
```

Record `$missionSourceCommit` in intake evidence. Do not use an installed AO
Mission binary, container image, previous Mission home, or previous campaign
root. Then start the replacement Mission:

```powershell
go run ./cmd/ao-mission objective start `
  --objective "Create and independently qualify one exact AO full-source development baseline that materializes from clean inputs and produces semantically equivalent credential-free results on macOS and Windows, without beginning AO Office Pool or widening release, provider, deployment, promotion, compatibility, credential, or RSI authority. Coordinate this as one bounded implementation workgraph." `
  --correlation-id ao-cross-platform-development-baseline-20260822-r2
```

Expected: one new Mission is persisted, `routing_class=complex`,
`initial_route=ao-atlas`, AO Atlas is `required`, AO Blueprint is `omitted`, and
all execution and approval authority flags remain false. Stop before S01 if any
field differs. Blueprint remains mandatory in the independently exercised S04
fixture; accepted complex Mission intake does not itself repeat Blueprint
requirements acceptance.

- [ ] **Step 3: Record the campaign constraints**

Import or attach one public-safe intake artifact containing the design commit, design blob digest, handoff commit, model policy, denied authorities, and seven slice identifiers `S01` through `S07`.

Expected: Mission readback preserves the exact correlation ID and reports no execution or approval authority.

- [ ] **Step 4: Commit only the handoff documentation**

```powershell
git status --short
```

Expected: generated Mission state remains untracked and outside the documentation commit.

### Task 2: S01 — Freeze The Full Development Baseline Contract

**Owner:** AO Architecture

**Required slice files:**
- Create: `stack/development-baseline-manifest.json`
- Create: `docs/contracts/development-baseline-manifest-v1.schema.json`
- Create: `scripts/verify_development_baseline.py`
- Create: `scripts/test_verify_development_baseline.py`
- Create: `docs/development-baseline.md`
- Modify: `AGENTS.md`

- [ ] **Step 1: Create a slice design and TDD implementation plan**

Write and review:

```text
docs/superpowers/specs/2026-08-22-development-baseline-contract-design.md
docs/superpowers/plans/2026-08-22-development-baseline-contract.md
```

The contract must bind the 14 stable-profile repositories, exact upstream URLs and commits, the seven public releases and platform digests from `stack/current-release-manifest.json`, toolchain constraints, repository-owned gate identifiers, platform overrides, and AO Next exclusion.

- [ ] **Step 2: Implement schema and verifier test-first**

Minimum negative cases: duplicate repository, missing stable member, malformed commit, moving-branch-only identity, release digest drift, undeclared platform override, AO Next in the stable profile, unsafe relative path, and unknown property.

Run:

```powershell
python scripts/test_verify_development_baseline.py
python scripts/verify_development_baseline.py --manifest stack/development-baseline-manifest.json
python scripts/verify_architecture.py
git diff --check
```

Expected: all tests pass; verifier prints the baseline identity digest and `repositories=14`, `errors=0`.

- [ ] **Step 3: Obtain independent contract review**

Require exact inventory, release binding, source-head provenance, and authority-boundary approval. Record the reviewed commit and manifest SHA-256 in Mission.

- [ ] **Step 4: Checkpoint S01**

Mission imports the S01 manifest, test summary, review, source commit, and artifact hashes. Continue only if S01 is terminal `pass` and the worktree is tracked-clean.

### Task 3: S02 — Safe Materialization And Native Preflight

**Owner:** AO Architecture

**Required slice files:**
- Create: `scripts/bootstrap_development_baseline.py`
- Create: `scripts/test_bootstrap_development_baseline.py`
- Create: `scripts/bootstrap-development-baseline.ps1`
- Create: `scripts/bootstrap-development-baseline.sh`
- Create: `.github/workflows/development-baseline-bootstrap.yml`

- [ ] **Step 1: Write and review the S02 design and plan**

The two supported modes are exactly `materialize` for an empty root and `verify-existing` for read-only verification. No overwrite, merge, reset, cleanup, dependency update, global Git change, or credential discovery is allowed.

- [ ] **Step 2: Implement offline fixture tests first**

Cover empty-root enforcement, sibling layout, detached exact-commit checkout, upstream mismatch, dirty repository, symlink/reparse traversal, unsafe archive member, digest mismatch, bounded downloads, paths containing spaces, CRLF checkout, PowerShell 5.1 parsing, and capability-dependent symlink classification.

Run:

```powershell
python scripts/test_bootstrap_development_baseline.py
powershell.exe -NoProfile -NonInteractive -Command "[void][scriptblock]::Create((Get-Content -Raw scripts/bootstrap-development-baseline.ps1))"
python scripts/verify_architecture.py
git diff --check
```

Expected: offline tests and PowerShell parsing pass with zero repository mutation outside fixture roots.

- [ ] **Step 3: Prove fresh materialization on both hosted platforms**

Dispatch the reviewed workflow on `macos-26` and `windows-2025`. Require the same baseline identity digest, 14 exact commits, clean detached checkouts, successful native preflight, and zero undeclared skips.

- [ ] **Step 4: Checkpoint S02**

Mission retains the two host artifacts and independent rehash report. Continue only if both artifact manifests close and cleanup reports zero run-owned residue.

### Task 4: S03 — Repository-Owned Development Gates

**Owner:** AO Architecture, with fixes owned by the failing repository

**Required slice files:**
- Create: `scripts/run_development_baseline_gates.py`
- Create: `scripts/test_run_development_baseline_gates.py`
- Modify: `stack/development-baseline-manifest.json`
- Modify: `.github/workflows/development-baseline-bootstrap.yml`

- [ ] **Step 1: Freeze the gate inventory**

For each included repository, read its exact-head `AGENTS.md` and quality manifest. Record stable command IDs and native argv arrays in the baseline manifest. Do not translate a repository command unless its own Windows contract requires Git Bash or PowerShell.

- [ ] **Step 2: Implement the sequential gate runner test-first**

Cover unknown command ID, shell mismatch, timeout, nonzero exit, undeclared skip, Git-state drift, bounded stdout/stderr hashing, partial-result preservation, and deterministic ordering.

Run:

```powershell
python scripts/test_run_development_baseline_gates.py
python scripts/verify_development_baseline.py --manifest stack/development-baseline-manifest.json
python scripts/verify_architecture.py
git diff --check
```

Expected: tests pass and fixture results remain deterministic.

- [ ] **Step 3: Run all gates on clean macOS and Windows roots**

Require every mandatory gate to pass. A repository failure creates a separate owning-repository defect slice with its own design, TDD fix, native CI, and re-frozen baseline commit. Never patch around the failure in Architecture.

- [ ] **Step 4: Checkpoint S03**

Mission imports per-repository summaries and exact log digests. Continue only when all 14 repositories are terminal and the regenerated baseline identity has independent approval.

### Task 5: S04 — Complete Credential-Free AO Workflow

**Owners:** AO Architecture contract runner plus existing component-owned fixture commands

**Required slice files:**
- Create: `scripts/run_development_baseline_workflow.py`
- Create: `scripts/test_run_development_baseline_workflow.py`
- Create: `stack/fixtures/development-baseline-v1/fixture-manifest.json`
- Create: `stack/fixtures/development-baseline-v1/README.md`
- Modify: `.github/workflows/development-baseline-bootstrap.yml`

- [ ] **Step 1: Write and independently review the S04 fixture contract**

The exact path is Mission intake → Blueprint authorization → Atlas workgraph → Foundry/Forge coordination → Covenant decision → AO2 bounded scripted fixture → Control Plane observation → Command/Mission readback → Arena/Crucible/Sentinel assurance → Promoter no-promotion result.

- [ ] **Step 2: Add contract-level failing tests**

Reject missing producer artifacts, wrong repository/source identity, correlation drift, digest mismatch, over-authority fields, provider requests, publication intent, promotion, RSI, nonterminal stages, duplicate stages, and incomplete cleanup ownership.

- [ ] **Step 3: Implement the smallest orchestration layer**

Invoke only component-owned CLIs with validated argv arrays and fixture paths. Architecture records transitions and verifies artifacts; it does not synthesize component approval or rewrite component output.

Run:

```powershell
python scripts/test_run_development_baseline_workflow.py
python scripts/run_development_baseline_workflow.py --fixture stack/fixtures/development-baseline-v1/fixture-manifest.json --output .ao-baseline/workflow-result.json
python scripts/verify_architecture.py
git diff --check
```

Expected: the local credential-free fixture reaches terminal `pass`, Promoter denies promotion, provider/credential/deployment/publication/RSI counts are zero, and cleanup succeeds.

- [ ] **Step 4: Checkpoint S04**

Run independently on clean macOS and Windows roots. Mission retains both raw terminal results without claiming parity yet.

### Task 6: S05 — Independent Semantic Parity Comparator

**Owner:** AO Architecture

**Required slice files:**
- Create: `scripts/compare_development_baseline_results.py`
- Create: `scripts/test_compare_development_baseline_results.py`
- Create: `docs/contracts/development-baseline-result-v1.schema.json`
- Create: `docs/contracts/development-baseline-parity-v1.schema.json`
- Modify: `stack/development-baseline-manifest.json`

- [ ] **Step 1: Freeze the normalization allowlist**

Allow only absolute root, path separator, executable suffix, shell name, timestamp, duration, process ID, and archive-format differences. Require exact equality for identities, transitions, policy outcomes, evidence content, readbacks, assurance results, denied authorities, and cleanup disposition.

- [ ] **Step 2: Implement the comparator test-first**

Each allowed normalization receives a positive and negative test. Add explicit failures for undeclared field drift, missing/extra evidence, manifest mismatch, self-declared pass with invalid content, cleanup disagreement, and authority widening.

Run:

```powershell
python scripts/test_compare_development_baseline_results.py
python scripts/compare_development_baseline_results.py --macos .ao-baseline/macos/result.json --windows .ao-baseline/windows/result.json --output .ao-baseline/parity.json
python scripts/verify_architecture.py
git diff --check
```

Expected: fixture parity passes; every injected semantic mismatch fails closed.

- [ ] **Step 3: Compare the real S04 artifacts**

The comparator independently rehashes both artifact manifests before comparison. Escalate to `gpt-5.6-sol` high only if the mismatch cannot be classified by the frozen rules.

- [ ] **Step 4: Checkpoint S05**

Mission records the parity verdict, exact input/output digests, and any owning-repository defect. Continue only on `parity=pass`.

### Task 7: S06 — Hosted Clean Qualification And Evidence Closure

**Owner:** AO Architecture

**Required slice files:**
- Create: `.github/workflows/development-baseline-qualification.yml`
- Create: `scripts/verify_development_baseline_evidence.py`
- Create: `scripts/test_verify_development_baseline_evidence.py`
- Modify: `docs/development-baseline.md`

- [ ] **Step 1: Add the final workflow with least authority**

Use read-only repository permissions, clean `macos-26` and `windows-2025` jobs, paths containing spaces, no AO/provider credentials, and one final comparison job. Upload bounded host results, manifests, and the parity verdict.

- [ ] **Step 2: Test evidence closure offline**

Reject duplicate paths, missing/extra files, size drift, digest drift, absolute/private paths, unsafe authority, wrong runner, wrong baseline identity, nonzero residue, and mismatched host result sets.

- [ ] **Step 3: Merge through required review and run on merged main**

Require Architecture verification, focused tests, `git diff --check`, pull-request CI, and the final workflow on merged `main`. Independently download and rehash all uploaded evidence.

- [ ] **Step 4: Checkpoint S06**

Mission imports only the independently verified evidence package and preserves the workflow run, jobs, runners, source commit, artifact names, sizes, and SHA-256 values.

### Task 8: S07 — Final Mission Reconciliation And Baseline Declaration

**Owners:** AO Mission and AO Architecture

**Required slice files:**
- Create: `../ao-architecture/docs/development-baseline-qualification.md`
- Modify: `../ao-architecture/docs/development-baseline.md`
- Move after completion: `docs/roadmap/handoffs/active/ao-cross-platform-development-baseline-handoff-prompt.md` to `docs/roadmap/handoffs/completed/`
- Modify: `docs/roadmap/handoffs/README.md`

- [ ] **Step 1: Audit every completion criterion**

Require one exact manifest, 14 exact repositories, two clean materializations, all mandatory development gates, two terminal workflow results, parity pass, zero residue, closed evidence manifests, green merged-main CI, and no authority widening.

- [ ] **Step 2: Reconcile Mission state**

Import the final Architecture terminal index and comparison result. Verify mission/correlation identities, source status, projected terminal status, artifact retention, and idempotent readback.

- [ ] **Step 3: Publish the bounded baseline declaration**

State the exact baseline digest, repository commits, runtime releases, supported host profiles, declared skips, qualification workflow, evidence digests, and limitations. Say only that the AO dependency foundation is ready for later AO Office Pool development; do not say Office Pool has started or passed.

- [ ] **Step 4: Close and archive the handoff**

Move the handoff to completed history, update the handoff index, verify Mission and Architecture documentation, and commit those documentation-only changes through their normal review paths.

- [ ] **Step 5: Stop**

Do not start AO Office Pool. Return the exact next action as a separate future objective that consumes the frozen baseline digest.
