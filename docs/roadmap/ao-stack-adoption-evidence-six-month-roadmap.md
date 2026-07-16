# AO Stack Adoption And Evidence Maintenance Six-Month Roadmap

Status: active; Month 2 closed
Created: 2026-07-16  
Trigger: start after Month 6 no-release closure from the post-stable roadmap.

## Baseline

This roadmap starts after the completed post-stable six-month cycle.

Current public pair:

- AO2 `v0.5.1`
  - Release: https://github.com/uesugitorachiyo/ao2/releases/tag/v0.5.1
  - Tag target: `80ec5321f42d4bab17d5e64fdae6aa099ba59d4a`
- AO2 Control Plane `v0.1.15`
  - Release: https://github.com/uesugitorachiyo/ao2-control-plane/releases/tag/v0.1.15
  - Tag target: `f1702b387607566cac457458af9adb5871a5c412`

Evidence baseline:

- Month 3 compatibility matrix: 16 total edges, 16 tested, 16 canonical
  vectors, 16 consumer tests, 0 proposed edges.
- Month 4 controlled self-improvement loop: fixture-only dry-run evidence,
  rollback proof, observation, operator readback, Sentinel wording checks, and
  Promoter no-RSI/no-promotion verdict.
- Month 5 operator workflow: Architecture source of truth, Command readback,
  Foundry safe-next-work, Forge run-state, Covenant policy readback, Sentinel
  checks, and Promoter no-promotion/no-RSI status.
- Month 6 release train readiness: no release selected. The current public pair
  remains sufficient.

Standing boundaries:

- RSI remains denied.
- Live self-modification remains denied.
- External beta has not launched.
- Promotion is not requested or granted.
- Provider pilots require separate explicit authorization.
- Releases, tags, uploads, and deployments require separate exact-scope
  authorization.
- `/tt`, modules, private app pilot folders, models, frameworks, generated
  binaries, and unrelated personal files are out of scope.

## Theme

Move AO Stack from verified internal readiness to maintained adoption
readiness. The focus is evidence freshness, operator usability, compatibility
drift detection, supportability, and release discipline.

The stack should become easier to operate without weakening gates. Evidence
should stay current. Operators should know what is safe, what is denied, what
is blocked, and what proof exists.

## Month 1: Evidence Freshness And Compatibility Gate Readiness

Goal: make the current evidence base refreshable and decide what the
Architecture compatibility gate means after 16/16 edges are tested.

Closure: [Adoption/Evidence Month 1 Closure: Evidence Freshness And Gate Readiness](adoption-month1-evidence-freshness-closure.md)

Result: evidence freshness checks, gate semantics, Command readback, Sentinel
wording checks, and Promoter no-promotion/no-RSI readback are merged. The
compatibility gate state is `ready`, not active. `compatibility_gate_complete`
remains false because activation was not authorized.

Work:

- Re-read the current public AO2 and Control Plane releases.
- Re-run or replay the 16 compatibility vectors and consumer tests.
- Add drift checks for vector files, consumer tests, current-release manifest,
  and matrix counts.
- Define the compatibility gate activation criteria in AO Architecture.
- Keep the gate false unless the activation criteria are fully verified.
- Add operator readback for gate state: active, false, denied, blocked, or
  waiting for evidence.
- Record no external beta, no promotion, and RSI denied.

Success criteria:

- The 16/16 matrix can be refreshed without manual storytelling.
- Drift is detected by tests or verifiers.
- Compatibility gate semantics are clear.
- Any activation remains evidence-gated and explicit.

## Month 2: Operator Adoption Drills

Goal: prove a solo operator can use the stack from documented workflows without
falling into private evidence paths or unsupported claims.

Closure: [Adoption/Evidence Month 2 Closure: Operator Adoption Drills](adoption-month2-operator-drills-closure.md)

Result: operator adoption drills, support evidence checks, policy readback,
observation readback, Command readback, Sentinel wording checks, and Promoter
no-promotion/no-RSI readback are merged or verified from existing evidence. The
compatibility gate remains `ready`, not active.

Work:

- Run clean operator drills using the current public pair.
- Exercise Command readback for stack state, policy gates, safe-next-work, and
  no-RSI status.
- Exercise Foundry and Forge handoff from safe-next-work to run-state.
- Exercise Covenant policy readback for approval-required and denied states.
- Exercise Control Plane observation readback for dry-run evidence.
- Capture gaps as small docs/tests PRs.

Success criteria:

- The operator can answer: what is current, what is safe, what is denied, and
  what to do next.
- Support evidence can be collected without credentials, private logs, or
  provider execution.
- No release is selected unless a shipped-artifact issue is found.

## Month 3: Evidence Maintenance Automation

Goal: reduce manual evidence upkeep by adding scheduled or repeatable
maintenance checks.

Work:

- Add or harden repeatable checks for current-release metadata.
- Add matrix/vector freshness checks.
- Add operator workflow freshness checks.
- Add Sentinel wording profiles for adoption docs and operator readbacks.
- Add Promoter no-promotion/no-RSI checks to readiness bundles.
- Create evidence maintenance reports that can be compared across runs.

Success criteria:

- Evidence drift is easy to detect.
- Current release state, compatibility state, operator workflow state, and
  denied authority state are checked together.
- Maintenance reports are machine-readable and human-readable.

## Month 4: Controlled Improvement Evaluation

Goal: improve the fixture-only controlled self-improvement loop without
granting RSI or live self-modification.

Work:

- Expand dry-run improvement fixtures.
- Add before/after measurement criteria for docs, workflows, prompts, and
  templates.
- Add rollback expectations for every improvement class.
- Add Command readback for proposed improvement, policy decision, dry-run
  result, rollback result, and denied authority.
- Add Sentinel and Promoter checks that keep self-improvement language bounded.

Success criteria:

- AO can evaluate proposed improvements in fixture-only dry-run mode.
- Every improvement has measurement and rollback evidence.
- RSI remains denied.
- No live self-modification or provider pilot occurs.

## Month 5: Adoption Package And Support Readiness

Goal: prepare a durable adoption package without launching external beta by
default.

Work:

- Assemble a current operator package:
  - current public pair
  - compatibility matrix state
  - operator workflow
  - policy gate readback
  - dry-run/no-RSI boundary
  - support reproduction checklist
- Refresh README and docs links only where needed.
- Run support drills for install, rollback, checksum, manifest mismatch,
  approval/replay, and operator readback issues.
- Decide whether a private/internal adopter workflow is ready.
- Keep external user contact and external beta launch unauthorized unless the
  operator explicitly grants that scope later.

Success criteria:

- A maintainer can hand off the adoption package without explaining internal
  history.
- Support cases have reproducible evidence paths.
- No unsupported external beta, promotion, or RSI claim appears.

## Month 6: Release Or No-Release Decision

Goal: run the next release-train readiness assessment and choose release or
no-release based on shipped-artifact impact.

Work:

- Inventory AO2 and Control Plane commits since the current public pair.
- Classify changes as runtime, packaging, install/rollback, docs, tests,
  fixtures, evidence, or support.
- Select no-release if changes do not require replacing public artifacts.
- Select patch/minor release only if shipped binary or package behavior changed
  enough to require publication.
- If release is selected, run qualification, publication, public verification,
  and post-public readback.
- If no release is selected, close with no-release evidence.

Success criteria:

- Release/no-release decision is evidence-backed.
- Public artifacts are replaced only when needed.
- Current release pair, compatibility state, operator workflow, Sentinel, and
  Promoter readbacks are recorded.
- RSI remains denied.

## Recommendation

Start Month 3 with evidence maintenance automation using the refreshed evidence
base and Month 2 operator drill results. Do not begin with a release train. The
current public pair remains sufficient until a readiness assessment finds
shipped-artifact impact.

Use AO Mission as supervisor and AO Atlas for the full-month workgraph. The
final response for each month should be allowed only when the full month is
closed or a true hard blocker remains.
