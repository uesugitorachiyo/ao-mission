# AO Stack Month 5 Beta Operations Handoff

This prompt advances the completed Month 4 consolidation wave into Month 5.
Paste the text block into the existing AO Mission task. Do not start a duplicate
six-month-roadmap mission.

```text
Start a Codex goal and resume the existing AO Stack six-month-roadmap mission
for the Month 5 Beta Operations wave.

Do not run `ao-mission start` if the roadmap mission already exists. Recover
the mission ID, accepted Month 4 terminal state, exact next action, repository
heads, and durable evidence links from Mission and Atlas readbacks first.

Authoritative Month 4 closure:
- AO Mission merged PR: #71 at 0d3daef.
- AO Atlas merged PR: #680 at 3dc55791.
- Completed nodes: 60 total.
- Month 4 baseline: 36/36, including seven nodes tagged
  `recommendation_source=codex_additional_month4`.
- Continuation soak: 24/24.
- Ready/blocked/failed: 0/0/0.
- The 120-minute lease passed and final response was allowed.
- Strict evidence validation passed: 1,020/1,020 JSON files.
- Production-readiness score: 100/100.
- Promoter recorded no promotion.
- Command recorded the compact timeline.
- Public-safety validation passed.
- Unrestricted RSI remained denied.
- Month 4 final report:
  <workspace>/ao-atlas/docs/evidence/ao-stack-month4-consolidation-v01/final-closure/month4-final-closure-report.md
- Atlas Month 5 source handoff:
  <workspace>/ao-atlas/docs/evidence/ao-stack-month4-consolidation-v01/final-closure/next-month5-beta-operations-handoff-prompt.md

Repository safety before work:
1. Verify `origin/main` contains the merged heads above.
2. AO Atlas main must be clean and synchronized.
3. The primary AO Mission checkout is intentionally dirty and six commits
   behind because pre-existing user work was preserved. Do not stash, clean,
   reset, checkout over, stage, commit, or otherwise alter those files.
4. Create or reuse a clean isolated AO Mission worktree based on the current
   `origin/main`. Perform Month 5 Mission changes only there.
5. Use isolated branches/worktrees for every other repository with local user
   changes. Never treat a dirty primary checkout as a roadmap blocker when a
   clean verified worktree can be used safely.
6. If a required merged head is missing from `origin/main`, stop and report
   the exact repository, expected head, observed head, and safe next action.

Month 5 objective:
Harden the consolidated AO stack for bounded beta operation. Implement and
verify durable control-plane recovery, exact approval binding, Covenant policy
identity, cross-repository compatibility, failure handling, operator readback,
and a clean-room non-AO golden path. Prepare a three-user pilot without making
provider calls or claiming production readiness.

This is an implementation wave, not another architecture-document wave.
A node counts as completed only when it produces at least one of:
- a product-code or schema change with executable tests;
- a hosted-CI or cross-repository conformance result;
- a migration, restart, backup, restore, rollback, or failure-injection drill;
- a clean-room operator install/run/inspect/rollback proof; or
- a decision that closes a real compatibility or authority gap and is bound to
  repository truth by an executable check.

Status-only readbacks, duplicate fixtures, restated plans, and evidence created
only to increase node count do not satisfy the Month 5 completion budget.

Work budget and pacing:
- Generate exactly 40 bounded baseline nodes from the priorities below.
- Complete all 40 baseline nodes before Month 5 final closure.
- Keep exactly one executable mutation node active at a time.
- Non-mutating verification may run concurrently only when write scopes cannot
  overlap and outputs are independently attributable.
- Target a 2-3 hour supervised baseline wave.
- After 40/40, run a continuation soak only if it exercises distinct restart,
  recovery, cross-platform, or clean-room cases. Cap continuation at 16 nodes.
- Do not create continuation nodes that merely repeat successful fixtures.
- Stop before provider execution, live deployment, promotion, release, or a
  real-user pilot unless the operator supplies a separate explicit authority.

Baseline Month 5 priorities (one bounded node per item):
01. Generate a stack lockfile and authority manifest from current repository
    heads, contract versions, policy digests, release boundaries, and owners.
02. Regenerate Architecture truth from repository and hosted-CI evidence; fail
    on stale or manually asserted status.
03. Implement the Covenant schema-owner registry and lifecycle rules for
    draft, active, deprecated, and rejected versions.
04. Run producer-consumer fixtures across Mission, Blueprint, Atlas, Foundry,
    Forge, Command, Covenant, AO2, and the Control Plane.
05. Bind Blueprint authorization to canonical approval bytes, digest, scope,
    expiry, and base commit.
06. Prove Atlas graph and context-pack compatibility against the accepted
    Blueprint artifact and stack lockfile.
07. Prove Foundry can schedule one bounded, capability-matched node without
    obtaining approval or policy authority.
08. Prove Forge can assemble and replay a provider-free GoalRun without
    changing approval, policy, or evaluator verdicts.
09. Enforce Command as a thin read-only client over producer-owned truth.
10. Require AO2 to verify exact approval bytes, digest, scope, base commit,
    expiry, and revocation state before an execution would begin.
11. Remove or hard-deny every hard-coded autoapproval and approval-bypass path;
    add negative tests for each discovered path.
12. Bind Covenant policy decisions to policy hash, approval identity, signer,
    scope, and the exact evaluated artifact.
13. Implement Control Plane transactional write fixtures covering partial
    failure, duplicate delivery, idempotency, and torn-write recovery.
14. Exercise Control Plane migration, backup, clean-directory restore, and
    digest-verified readback.
15. Exercise Mission restart, kill, resume, lease expiry, fencing, cancellation,
    and replay without duplicate mutation.
16. Run the first complete provider-free golden-path dry run through Blueprint,
    Atlas, Foundry, Forge, Covenant, AO2 boundary checks, Control Plane,
    Sentinel, Promoter, and Command.
17. Replay the golden path against at least three representative external
    non-AO repositories using dry-run or no-mutation adapters.
18. Add or verify hosted CI for Arena, Crucible, Sentinel, and Promoter with
    read-only default permissions and responsibility-specific checks.
19. Make Sentinel consume native freshness, CI, contract, recovery, and
    provenance signals rather than fixture-authored success alone.
20. Prove Promoter remains unable to activate, deploy, publish, release, or
    convert observer evidence into approval authority.
21. Implement the Command compact timeline and approval-inbox read model while
    preserving producer ownership and denial states.
22. Carry the Month 4 provenance envelope through provider-free run records;
    make missing or silent fallback provenance invalid.
23. Extract one focused AO2 CLI module from a touched oversized file without
    changing behavior or widening authority.
24. Extract one focused Foundry or Forge module from code touched by this wave,
    with an owner, interface, focused tests, and rollback path.
25. Extract one focused Command presentation module while keeping domain truth
    outside Command.
26. Implement the evidence delta guard and content-addressed externalization
    path for new large artifacts, including retrieval and digest verification.
27. Exercise cross-platform install, upgrade, rollback, and uninstall fixtures
    for the supported beta environments without publishing a release.
28. Run the accepted failure matrix across approval, policy, scheduling,
    execution boundary, observer storage, evaluator, and presentation layers.
29. Execute a bounded soak covering restart/resume, cancellation, recovery,
    evidence retrieval, and compacted timeline readback.
30. Write and clean-room rehearse the three-user pilot runbook without provider
    calls, credentials, production mutation, or maintainer-only shortcuts.
31. Produce a real-run acceptance ledger that distinguishes implemented,
    executable, rehearsed, planned, and not-authorized capabilities.
32. Generate a beta release BOM and compatibility matrix without releasing,
    tagging, uploading, or publishing anything.
33. Add wording guards that reject unsupported beta, production, autonomous,
    promotion, or RSI claims in generated reports and operator surfaces.
34. Re-run the bounded RSI continuity gate and authority-denial corpus across
    all Month 5 changes; preserve capability without making a broader RSI claim.
35. Generate an operator beta dashboard from native producer readbacks,
    including freshness, failure, recovery, storage, and denial state.
36. Verify compact operator summary, context compaction, and resume behavior
    from a clean process.
37. Produce 40 ranked Month 6 canary candidates with owner, prerequisite,
    risk, rollback, evidence cost, and authority requirement.
38. Generate the terminal cross-repository compatibility, recovery, and
    evidence-location rollup.
39. Require Sentinel, Promoter, and Command to agree on the terminal state and
    to preserve no-promotion and RSI-denial semantics.
40. Write the Month 6 canary handoff, separating provider-free prerequisites
    from actions that require explicit provider, credential, release, pilot,
    deployment, or promotion authority.

=======================================================================
ADDITIONAL CODEX SAFETY GATES - NO EXTRA ROADMAP NODES
=======================================================================

Apply these gates across the 40 baseline nodes. They strengthen Month 5 but do
not increase the node budget or create a parallel Helix/frontier project.

GATE-A: Implementation-truth ledger
- For every claimed capability, record one of: implemented, executable-tested,
  clean-room-rehearsed, fixture-only, planned, or not-authorized.
- Never promote fixture-only or planned work to implemented beta capability.
- Bind claims to code commit, test command, result digest, and owning producer.

GATE-B: Beta reliability budget
- Establish measured baselines for successful resume, duplicate-mutation rate,
  rollback success, restore integrity, operator intervention, execution time,
  tracked evidence growth, and external artifact retrieval.
- Month 5 may establish the baseline; it must not invent an unsupported SLO.
- Record failure cases and exact repair work instead of averaging them away.

GATE-C: Approval-boundary kill switch
- Before any path reaches an execution boundary, verify approval bytes, digest,
  scope, base commit, policy hash, expiry, revocation, and capability match.
- Any mismatch must fail closed before provider or mutation setup.
- This wave tests the boundary without calling a provider.

GATE-D: Evidence-volume control
- Keep schema-required compact evidence in Git.
- Externalize bulky logs, repeated campaign data, and generated payloads through
  the content-addressed observer path.
- Every externalized artifact needs a compact digest manifest plus retrieval
  and clean-restore proof.
- Track file-count and byte deltas. Do not generate redundant node evidence to
  satisfy runtime or count targets.

GATE-E: RSI continuity without authority expansion
- Re-run the Month 4 bounded RSI continuity corpus before and after contract,
  policy, execution-boundary, and module-extraction changes.
- Block regressions in bounded recursive-improvement evidence or denial fields.
- Do not authorize unrestricted self-modification, policy-changing autonomy,
  self-promotion, provider access, or a broader RSI claim.

=======================================================================
END ADDITIONAL CODEX SAFETY GATES
=======================================================================

Per-node execution loop:
1. Read the owning implementation, tests, contracts, and latest durable
   readbacks before proposing a change.
2. Record node gate, scope, owner, authority exclusions, dependencies, accepted
   input digests, and rollback boundary.
3. Add or strengthen an executable failing test or drill before implementation
   when behavior changes.
4. Make one bounded change in one mutable scope.
5. Run targeted verification, owning-repository verification, relevant
   producer-consumer conformance, public-safety checks, and RSI continuity when
   the node touches a protected boundary.
6. Record compact candidate, rollback, verification, provenance, Sentinel,
   Promoter no-promotion, Command readback, and run-link evidence.
7. Open a reviewable PR, wait for hosted CI, merge only when green, sync the
   isolated worktree, and delete local and remote task branches.
8. Run Mission and Atlas readback. Continue automatically while exact next
   action or ready work remains and no true hard blocker exists.

Cross-repository change order:
1. Contract/schema owner and consumer inventory.
2. Failing producer-consumer or negative fixture.
3. Backward-compatible producer change.
4. Consumer updates and conformance matrix.
5. Migration/restart/rollback drill when persistence or lifecycle is touched.
6. Architecture, operator, and compact evidence readback.
7. PR, CI, merge, synchronized heads, and branch cleanup.

Month 5 pilot boundary:
- The baseline wave may prepare and clean-room rehearse a three-user pilot.
- It may not inspect credentials, call providers, mutate a real external
  repository, invite or message pilot users, publish artifacts, deploy, or
  promote.
- At closure, produce a separate exact authorization packet listing requested
  actions, repositories, providers/models, credentials needed, cost ceiling,
  mutation scope, users, rollback, stop conditions, and expected evidence.
- Absence of that separate authority is `planned_not_authorized`, not a failed
  Month 5 node and not a reason to bypass the boundary.

Safety boundaries:
- no direct mutation of any main branch;
- no destructive handling of the dirty primary AO Mission checkout;
- no credential, token, secret, or private-key inspection;
- no provider or model calls;
- no release, deploy, publish, upload, tag, or live pilot action;
- no dependency updates;
- no auth, policy, capability, sandbox, or configuration widening;
- no hidden-instruction mutation;
- no self-approval or fixture-authored authority;
- no promotion request;
- no unrestricted RSI or autonomous-operation claim.

Minimum repository verification:
- Run each touched repository's documented test, vet/lint, build, schema,
  contract, public-safety, and production-readiness gates.
- For AO Atlas, run:
  `go test ./... -count=1`
  `go vet ./...`
  `go build ./cmd/atlas`
  `scripts/atlas-foundry-roundtrip-smoke.sh`
  `scripts/production-readiness.sh`
  strict evidence validation
  `git diff --check`
  scoped public-safety scan
- For AO Mission, verify both the clean isolated worktree and the merged remote
  head without touching the dirty primary checkout.
- Require hosted CI to pass for every merged PR.

Final response is allowed only when:
- all 40 baseline nodes are completed;
- any started continuation nodes have terminal states;
- ready_nodes=0, blocked_nodes=0, and failed_nodes=0;
- final_response_allowed=true;
- the supervised lease is at least 120 minutes unless all acceptance work is
  complete and the durable scheduler explicitly records an earlier valid
  terminal condition;
- strict evidence validation passes with no unknown schemas;
- all touched repositories pass local and hosted verification;
- migration, restart, backup, restore, rollback, and failure drills pass for
  every scope that Month 5 changed;
- the implementation-truth ledger contains no inflated capability claim;
- public-safety validation passes;
- Promoter says no promotion and Command agrees;
- unrestricted RSI remains denied while bounded RSI continuity passes;
- provider, release, deployment, and live-pilot actions remain unexecuted;
- all task worktrees are clean and synchronized;
- local and remote task branches are deleted; and
- the dirty primary AO Mission checkout remains unchanged.

The final report must state:
- baseline and continuation node counts;
- ready/blocked/failed counts;
- repository heads and merged PRs;
- implementation versus fixture-only capability counts;
- compatibility, failure-injection, restart, restore, rollback, and clean-room
  results;
- tracked evidence byte/file delta and external artifact retrieval result;
- Promoter, Command, Sentinel, public-safety, promotion, and RSI state;
- whether a Month 6 canary authorization packet was produced;
- exact next action; and
- confirmation that the primary AO Mission user changes were preserved.
```
