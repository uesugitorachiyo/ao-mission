# AO Mission Operator Next Actions

AO Mission should leave the operator with a concrete command, not only a
generated artifact path. Use this short sequence when checking or continuing a
mission.

## Current AO Stack Roadmap Action

The bounded-autonomy and product validation cycle has started.

Bounded Autonomy Month 1 is closed:
[Bounded Autonomy Month 1 Closure](roadmap/bounded-autonomy-month1-baseline-closure.md).

Month 1 verified the current public pair, created the bounded-autonomy
benchmark corpus and result schema, recorded baseline metrics, and added AO
Command benchmark readback. The compatibility gate remains `ready`, not active.

Bounded Autonomy Month 2 is closed:
[Bounded Autonomy Month 2 Closure](roadmap/bounded-autonomy-month2-end-to-end-workflows-closure.md).

Month 2 completed one documentation/support correction, one deterministic
single-repository code change, and one cross-repository producer/consumer
contract change. AO Architecture and AO Command now share a canonical
bounded-autonomy benchmark-to-operator-readback vector. The compatibility gate
remains `ready`, not active.

Next action: start Bounded Autonomy Month 3 long-running reliability and
recovery. Focus on interrupted mission resume behavior, checkpoint integrity,
failed-CI repair, partial cross-repo merge recovery, and evidence
reconciliation without duplicate work or false completion. Do not start a
release train by default.

Bounded Autonomy Month 3 is closed:
[Bounded Autonomy Month 3 Closure](roadmap/bounded-autonomy-month3-reliability-recovery-closure.md).

Month 3 added a machine-checked recovery readback covering 10 failure
injection classes and 9 recovery proof classes. The long-run drill verified
checkpoint integrity, restart proof, archive validation, event indexing,
Command status, final rollup, final reconciliation, and final-response denial
while continuation evidence remains.

Next action: start Bounded Autonomy Month 4 controlled improvement engine v1.
Evaluate bounded improvement candidates from Month 1-3 evidence, measure
baseline and candidate outcomes, require approval where existing gates require
it, verify rollback, and accept only measured improvements after green CI.
Do not start a release train, RSI, or live self-modification.

Bounded Autonomy Month 4 is closed:
[Bounded Autonomy Month 4 Closure](roadmap/bounded-autonomy-month4-controlled-improvement-closure.md).

Month 4 evaluated three controlled-improvement candidate classes, added a
machine-checked AO Mission closure fixture, and added AO Command operator
readback coverage for proposal, approval, measurement, decision, and rollback
state. The compatibility gate remains `ready`, not active.

Next action: start Bounded Autonomy Month 5 one-operator production dogfood.
Use genuine bounded AO repository maintenance where it provides evidence, and
use fixtures where a real mutation adds no proof. Do not start a release train,
external beta, promotion, provider pilot, live self-modification, or RSI work.

Bounded Autonomy Month 5 is closed:
[Bounded Autonomy Month 5 Closure](roadmap/bounded-autonomy-month5-dogfood-closure.md).

Month 5 completed the five-task one-operator dogfood portfolio and records the
result as machine-checked AO Mission production-readiness evidence. The
portfolio covered support handoff, product/reliability readiness, cross-repo
Command readback, failure recovery, and a correctly denied RSI/live
self-modification route. The compatibility gate remains `ready`, not active.

Next action: start Bounded Autonomy Month 6 autonomy qualification. Rerun the
full evidence base, compare final measurements with Month 1, and make a
release/no-release assessment. Do not publish unless Month 6 selects a release
and all release gates pass.

Bounded Autonomy Month 6 is closed:
[Bounded Autonomy Month 6 Closure](roadmap/bounded-autonomy-month6-qualification-closure.md).

Month 6 reran the bounded-autonomy evidence base. The repaired release decision
keeps AO2 at `v0.5.1` and publishes AO2 Control Plane `v0.1.16` after exact
stable-patch qualification and public-asset verification. Controlled RSI
research is not authorized by this closure; RSI remains denied.

Next action: start a new roadmap only with explicit authorization. Use the
bounded-autonomy evidence and no-release decision as the baseline.

Bounded Autonomy repair from Month 3 is closed:
[Bounded Autonomy Repair Closure](roadmap/bounded-autonomy-repair-from-month3-closure.md).

The previous Month 1-6 closure is classified as `PARTIAL_INVALID_CLOSURE`.
The repair reopens from Month 3, fixes terminal final-rollup exact-next-action
handling, reruns terminal recovery, executes rollback verification, runs a
fresh dogfood portfolio, and repeats Month 6 qualification. The repeated
release decision keeps AO2 at `v0.5.1` and publishes AO2 Control Plane
`v0.1.16`; the Control Plane `spin` lockfile change is recorded as a compiled
dependency impact. RSI remains denied.

Month 6 is closed after Control Plane stable-patch qualification. AO2 v0.5.1
remains the current public AO2 release, AO2 Control Plane v0.1.16 is the
companion release, AO
Architecture records all 16 live compatibility matrix edges with canonical
vectors and consumer tests, the controlled self-improvement loop remains
fixture-only dry-run evidence, and the operator workflow readback chain is
merged.

Closure evidence:

- [Month 3 evidence and audit compatibility closure](roadmap/month3-evidence-audit-compatibility-closure.md)
- [Month 4 controlled self-improvement dry-run closure](roadmap/month4-controlled-self-improvement-dry-run-closure.md)
- [Month 5 operator workflow hardening closure](roadmap/month5-operator-workflow-hardening-closure.md)
- [Month 6 release train readiness closure](roadmap/month6-release-train-readiness-closure.md)
- AO Architecture final matrix PR:
  https://github.com/uesugitorachiyo/ao-architecture/pull/122
- AO Architecture Month 6 no-release readiness PR:
  https://github.com/uesugitorachiyo/ao-architecture/pull/125
- AO Mission Month 6 closure records the implementation PRs and local evidence
  directory.
- The closure document records the local evidence directory and report paths.

The new adoption/evidence cycle has started:
[AO Stack adoption and evidence maintenance](roadmap/ao-stack-adoption-evidence-six-month-roadmap.md).

Month 1 is closed:
[Adoption/Evidence Month 1 Closure](roadmap/adoption-month1-evidence-freshness-closure.md).

Month 1 made the evidence base refreshable and defined compatibility gate
states. AO Architecture now records the compatibility gate state as `ready`,
not active. `compatibility_gate_complete` remains false because activation was
not authorized.

Month 2 is closed:
[Adoption/Evidence Month 2 Closure](roadmap/adoption-month2-operator-drills-closure.md).

Month 2 proved the operator adoption drill path against the current public pair
and refreshed evidence base. AO Command presents the current pair, 16/16 matrix
state, gate state `ready` but not active, safe-next-work, run-state, policy,
observation, support categories, denied states, and next safe action. Sentinel
and Promoter prevent unsupported adoption, release, external beta, promotion,
provider, or RSI claims.

Month 3 is closed:
[Adoption/Evidence Month 3 Closure](roadmap/adoption-month3-evidence-maintenance-closure.md).

Month 3 made evidence maintenance repeatable through Architecture freshness and
matrix drift checks, Atlas maintenance workgraph/readback, Command maintenance
readback, Sentinel wording checks, and Promoter no-promotion/no-RSI readback.
The current public pair still matches the manifest, the matrix remains 16/16
tested, and the compatibility gate remains `ready`, not active.

Month 4 is closed:
[Adoption/Evidence Month 4 Closure](roadmap/adoption-month4-controlled-improvement-evaluation-closure.md).

Month 4 fixed production-readiness hygiene and reverified controlled
improvement evaluation as fixture-only dry-run evidence. AO2 rollback evidence,
Control Plane observation, Command readback, Sentinel wording checks, and
Promoter no-promotion/no-RSI readback remain bounded. The compatibility gate is
`ready`, not active.

Month 5 is closed:
[Adoption/Evidence Month 5 Closure](roadmap/adoption-month5-support-readiness-closure.md).

Month 5 added the adoption support package source of truth, Command
support-readiness readback, Sentinel support wording checks, and Promoter
no-promotion/no-RSI verdict. The support package covers install, checksum,
manifest mismatch, approval/replay, rollback, Windows-safe rollback, operator
readback issues, and public-safe issue-report fields. The compatibility gate
remains `ready`, not active.

Month 6 is closed:
[Adoption/Evidence Month 6 Closure](roadmap/adoption-month6-no-release-readiness-closure.md).

Month 6 previously selected `release_decision=no_release`; the later repaired
stable-patch qualification selected and published AO2 Control Plane `v0.1.16`.
AO2 remains `v0.5.1`, and the compatibility gate remains `ready`, not active.

The GitHub issue-to-draft-PR cycle has started:
[GitHub Issue To Draft PR Month 1 Closure](roadmap/github-issue-to-draft-pr-month1-closure.md).

Month 1 established the supervised contract boundary for GitHub issue URL
intake, immutable issue evidence, policy classification, draft-PR authority,
Command readback, Sentinel wording checks, and Promoter no-promotion/no-RSI
readback. Feature-generated PRs remain draft and unmerged by default.

Month 2 is closed:
[GitHub Issue To Draft PR Month 2 Closure](roadmap/github-issue-to-draft-pr-month2-closure.md).

Month 2 added the authenticity and reproduction gate. It records the issue
truth set, isolated acquisition planning, command and network policy,
deterministic/flaky reproduction metrics, non-bug and security-sensitive stop
states, Control Plane observation, Command readback, Sentinel wording checks,
and AO Mission supervision. Feature-generated PRs remain draft and unmerged by
default.

Next action: continue with GitHub issue-to-draft-PR Month 3 isolated repair,
verification, rollback, and replay. Prove the smallest safe repair only after
failing pre-patch reproduction evidence. Preserve fork and draft-PR boundaries,
and do not approve, merge, mark ready, or otherwise advance feature-generated
PRs.

External beta has not launched, promotion is not requested or granted, and RSI
remains denied. This is not a release train by default.

For doubled 2-3 hour waves, use the dedicated
[Long-Run Operator Runbook](long-run-operator-runbook.md). It defines the
30-node request shape, role routing, stop gate, per-node evidence, and Atlas
continuation prompt template.

For local private pilots on real local codebases, use the
[Local Private Pilot Workflow](local-private-pilot-workflow.md). For iOS app or
device smoke work, use the
[iOS/Xcode Local App Smoke Checklist](templates/ios-xcode-app-smoke-checklist.md).
These guides keep evidence local, preserve provided-library boundaries, and
avoid upload or deploy actions.

## Start And Inspect

```sh
ao-mission start "<objective>"
ao-mission doctor
ao-mission mission list --json
ao-mission next --mission <mission-id>
ao-mission status --mission <mission-id>
```

## Continue Locally

```sh
ao-mission continue --mission <mission-id> --until-done --max-iterations 20 --min-nodes 15 --min-minutes 120 --max-minutes 180
ao-mission mission history --mission <mission-id>
ao-mission mission events index --out tmp/mission-event-index.json
ao-mission mission events query-index --index tmp/mission-event-index.json --out tmp/mission-timeline-query-index.json
ao-mission mission events search --mission <mission-id> --query "AO Atlas" --index tmp/mission-event-index.json --json
ao-mission mission events search --mission <mission-id> --query "checkpoint" --index tmp/mission-event-index.json --json
ao-mission mission events search --mission <mission-id> --query "return_gate" --index tmp/mission-event-index.json --json
ao-mission mission events resume-prompt --mission <mission-id> --out tmp/<mission-id>-compaction-resume-prompt.json --json
ao-mission mission dashboard --mission <mission-id> --compact --out tmp/<mission-id>-dashboard.json
ao-mission mission verification-bundle --mission <mission-id> --readiness-bundle tmp/ao-mission-readiness-bundle.json --gateway-replay-bundle tmp/gateway-replay-bundle.json --out tmp/<mission-id>-verification-bundle.json
ao-mission mission compact --mission <mission-id> --keep-route-history 25 --keep-steps 25
ao-mission mission compact --mission <mission-id> --keep-route-history 25 --keep-steps 25 --timeline
ao-mission mission archive --mission <mission-id> --out tmp/<mission-id>-archive.json
ao-mission mission validate-archive --path tmp/<mission-id>-archive.json --out tmp/<mission-id>-archive-validation.json
ao-mission artifacts manifest --mission <mission-id> --out tmp/<mission-id>-artifact-manifest.json
ao-mission import atlas-recommendation-readback --mission <mission-id> --path tmp/recommendation-readback.json
ao-mission final rollup --mission <mission-id>
ao-mission final reconcile --mission <mission-id>
```

For 2-3 hour work, AO Mission owns the long-run supervisor lease and checkpoint
ledger. Atlas owns workgraph and context-heavy sequencing. Foundry owns exactly
one bounded implementation node at a time. Blueprint is used only when Mission
or Atlas lacks requirements, authorization, or a safe class boundary. Do not
route ready Atlas nodes through Blueprint for batching.

The final rollup is not a final response unless `final_response_allowed=true`.
If it is false, use `exact_next_action`, the latest checkpoint bundle, and the
Feature Depth Recommendations to continue. A true terminal hard blocker is a
blocked or denied rollup, stopped mission, explicit blocker list, or safety
boundary mismatch after repair/repack/support work has already been attempted.

Use `atlas-recommendation-readback` when AO Atlas has already executed a
recommendation wave and emitted a terminal readback. Mission closes only when
the Atlas readback reports all generated nodes complete, zero ready nodes,
checkpoint evidence for the wave, the long-run lease minimum met, and
`final_response_allowed=true`. If any of those are missing, Mission keeps the
route on AO Atlas and preserves the exact next action for continuation.

For the 25-node Atlas recommendation import wave, use Mission as the supervisor
and keep importing Atlas readbacks after each bounded node. Do not stop at a
green PR, a single rollup, or one completed import. Continue until at least 25
nodes complete, all ready nodes are gone, the 120-minute lease minimum is met,
and Mission, Atlas, Foundry, Command, and Promoter readbacks agree. This is the
25-node Atlas recommendation import wave closure gate.

Use `examples/valid/final-reconciliation-packet.json` as the shape reference for
the final Mission reconciliation packet. A valid packet keeps
`promotion_claimed=false`, `rsi_remains_denied=true`,
`claims_authority_advance=false`, and all execution/approval flags false.

## Command And Final Reconciliation Closure Check

Before any final response for an Atlas recommendation import wave, run the
Command and final reconciliation closure check:

```sh
ao-mission command status --mission <mission-id>
ao-mission command status --mission <mission-id> --json
ao-mission mission events index --out tmp/mission-event-index.json
ao-mission mission events search --mission <mission-id> --kind atlas_recommendation --index tmp/mission-event-index.json --json
ao-mission mission events search --mission <mission-id> --kind final_reconciliation --index tmp/mission-event-index.json --json
ao-mission final reconcile --mission <mission-id>
```

Treat the final response as denied if Command status, event search, or final
reconciliation still shows ready nodes, a stale route, `final_response_allowed`
false, a missing Promoter no-promotion summary, a missing Foundry terminal
rollup binding, a missing Command compact timeline, or an exact next action.
The next action stays with Atlas unless the imported readback is terminal and
all no-authority closure evidence agrees.

## Gateway Fixture Checks

```sh
ao-mission telegram replay --matrix examples/valid/telegram-command-matrix.json --config examples/valid/telegram-config.json
ao-mission telegram replay-updates --fixture examples/valid/telegram-update-replay.json --config examples/valid/telegram-config.json
ao-mission telegram webhook-replay --fixture examples/valid/telegram-webhook-replay.json --config examples/valid/telegram-config.json
ao-mission telegram role-matrix --config examples/valid/telegram-config.json --out tmp/telegram-role-matrix.json
ao-mission a2a serve --http --once
ao-mission a2a replay --fixture examples/valid/a2a-http-integration.json
ao-mission a2a lifecycle --fixture examples/valid/a2a-task-lifecycle-edges.json
ao-mission a2a compatibility --agent-card examples/valid/a2a-agent-card.json --http examples/valid/a2a-http-integration.json --lifecycle examples/valid/a2a-task-lifecycle-artifacts.json --out tmp/a2a-compatibility.json
ao-mission a2a streaming-denial --agent-card examples/invalid/a2a-agent-card-streaming.json --out tmp/a2a-streaming-denial.json
ao-mission a2a streaming-denial --agent-card examples/invalid/a2a-agent-card-streaming-sse.json --out tmp/a2a-streaming-sse-denial.json
ao-mission gateway replay-bundle --telegram-config examples/valid/telegram-config.json --telegram-matrix examples/valid/telegram-command-matrix.json --telegram-webhook examples/valid/telegram-webhook-replay.json --telegram-updates examples/valid/telegram-update-replay.json --a2a-http examples/valid/a2a-http-integration.json --a2a-lifecycle examples/valid/a2a-task-lifecycle-artifacts.json --scheduler examples/valid/scheduler-readback-replay.json --out tmp/gateway-replay-bundle.json
ao-mission gateway replay-suite --telegram-config examples/valid/telegram-config.json --telegram-webhook examples/valid/telegram-webhook-replay.json --telegram-updates examples/valid/telegram-update-replay.json --a2a-http examples/valid/a2a-http-integration.json --a2a-lifecycle examples/valid/a2a-task-lifecycle-artifacts.json --out tmp/gateway-replay-suite.json
ao-mission gateway readiness-rollup --mission <mission-id> --suite tmp/gateway-replay-suite.json --a2a-compatibility tmp/a2a-compatibility.json --archive-validation tmp/<mission-id>-archive-validation.json --snapshot-diff <snapshot-diff.json> --out tmp/gateway-readiness-rollup.json
ao-mission schedule replay --fixture examples/valid/scheduler-readback-replay.json
ao-mission schedule alerts --fixture examples/valid/scheduler-readback-replay.json
ao-mission schedule recover --mission <mission-id> --fixture examples/valid/scheduler-readback-replay.json
ao-mission import scheduler-recovery-readback --mission <mission-id> --path examples/valid/scheduler-recovery-readback.json
ao-mission import ledger-compaction-readback --mission <mission-id> --path examples/valid/ledger-compaction-readback.json
ao-mission validate contract --path examples/valid/timeline-compaction-readback.json
ao-mission mission compact --mission <mission-id> --keep-route-history 25 --keep-steps 25 --dry-run
ao-mission artifacts repair-manifest --path <artifact-manifest.json> --out <artifact-manifest.repaired.json>
ao-mission governance diff --before <snapshot-before.json> --after <snapshot-after.json>
ao-mission mission archive --mission <mission-id> --out tmp/<mission-id>-archive.json
```

## Cross-Repo Readiness Bundle

```sh
ao-mission mission readiness-bundle --repo ao-mission=tmp/ao-mission-readiness.txt --repo ao-atlas=tmp/ao-atlas-readiness.txt --out tmp/ao-mission-readiness-bundle.json
```

Use local readiness summaries only. The bundle records status and SHA-256
digests for operator review; it does not push branches, open PRs, wait for
hosted CI, merge, sync main, or delete branches.

The verification bundle is the final local handoff packet for this loop. It
binds the event index, dashboard, artifact manifest, readiness bundle, and
gateway replay bundle with SHA-256 digests, but it still does not perform
credentialed remote lifecycle work.

Telegram and A2A fixture checks record intent/readback only. They do not grant
execution authority, approval authority, repository mutation, provider calls,
credential use, release/deploy/publish/upload/tag authority, dependency update
authority, direct-main mutation, concurrent mutation, hidden instruction
mutation, unrestricted self-modification, unrestricted RSI, or broad_RSI.

## Route Handoff

If `ao-mission next --mission <mission-id>` returns `ao-blueprint`, Move to AO
Blueprint and create or inspect the Blueprint pack.

If it returns `ao-atlas`, Move to AO Atlas and compile/import the authorized
Mission context into an Atlas workgraph.

If it returns `ao-foundry`, Move to AO Foundry and consume only the first safe
Atlas node through Foundry gates.

If it returns `complete`, read the final rollup and recommended next tasks. Do
not treat a generated handoff file alone as completion.

## Long-Run Role Routing

Use AO Mission when the operator asks for a 2-3 hour run, minimum node budget,
checkpoint/resume behavior, final-response gating, or cross-repo readback
reconciliation.

Use AO Atlas when the work needs an ordered workgraph, context pack, exact next
ready node, repair plan, or Feature Depth Recommendation wave.

Use AO Foundry when Atlas has produced a ready bounded node and the next action
is implementation evidence, tests, rollback evidence, or a run-link/final
rollup.

Use AO Blueprint only for missing requirements, missing authorization, or an
underspecified objective. Blueprint should not be used to split ordinary ready
implementation batches.

For doubled waves, Blueprint is not a batching queue. Use the long-run runbook
and route ready Atlas nodes directly to Foundry one bounded implementation node
at a time.
