> COMPLETED HISTORICAL HANDOFF - DO NOT REEXECUTE.
> Preserved for audit history only; use the active handoff directory for current execution.

# AO Stack Month 3 Golden-Path Implementation Handoff

Paste this prompt into the existing AO Mission task. Do not use `ao-mission start` to create a duplicate mission.

```text
Start a Codex goal and resume the AO Mission supervised AO Atlas Month 3 Golden-Path Implementation wave.

Context:
- The Month 2 AO Atlas contract-convergence wave is complete.
- Evidence root: <workspace>/ao-atlas/docs/evidence/ao-stack-contract-convergence-month2-wave-v01
- Terminal state: 40 completed, 0 ready, 0 blocked, 0 failed.
- Final response was allowed after final Promoter, Command, and Sentinel binding.
- Final Atlas main head: 525727e31562282e92006b33da9ab879ebc64d43.
- Month 2 found real gaps instead of proving a product: Sentinel hosted CI is missing, Promoter signed assurance is unproven, Architecture compatibility has zero canonical vectors and zero consumer tests, Mission remains a readback/router handoff, and the golden path is not implemented.

Objective:
Build one real, safe, reproducible Mission -> Blueprint -> Atlas -> Foundry -> Forge -> Covenant -> AO2 -> Control Plane -> Command golden path.

Work budget:
- Generate at least 40 bounded implementation, contract, integration, and operator-readiness nodes.
- Complete at least 30 nodes before final response.
- If the first 30 finish quickly and no blocker remains, continue toward all 40.
- Target a 2-3 hour supervised run when ready work remains.
- Do not stop after one audit, one repository, one Foundry import, one PR, one CI pass, or one short batch.

Priority order:
1. Current source-of-truth and authority statements.
2. Canonical contracts, schemas, vectors, and compatibility gates.
3. One bounded golden path with no provider execution by default.
4. Approval, digest, rollback, lease, and restart correctness.
5. Operator readback and evidence usability.
6. Consolidation and evidence externalization.
7. Beta readiness and real workload measurement.
8. RSI remains paused and denied.

Required themes, split into implementation and regression/evidence nodes as needed:
1. Implement AO2 approval binding to exact proposed bytes and base commit.
2. Remove or permanently gate AO2 hardcoded auto-approval.
3. Include Covenant policy fields in event hashes and add tamper tests.
4. Make Covenant the canonical schema registry owner.
5. Add stable/experimental/deprecated contract classification.
6. Add shared Go/Rust canonical JSON vectors.
7. Add producer/consumer compatibility tests for every gate-critical edge.
8. Add generated type smoke tests without adding a dependency unless separately authorized.
9. Add the missing Sentinel hosted CI workflow with read-only permissions.
10. Make Sentinel consume real CI, runtime, policy, and evidence-freshness signals where safe.
11. Add Promoter signed-assurance input verification and signer/key binding design tests.
12. Keep Promoter activation outside the evaluator and dry-run by default.
13. Implement Foundry workspace-root golden-path preflight on a tiny non-AO repository.
14. Require isolated worktree, exact digest approval, verified diff, PR review, and rollback receipt.
15. Keep Mission responsible for durable state, routing, recovery, migrations, and lifecycle metrics.
16. Add Mission restart, lease, migration, event-index, and handoff-accounting invariants.
17. Add control-plane indexed storage migrations, leases, backup/restore, and crash-recovery tests.
18. Make Command a thin readback client over Mission and the control plane.
19. Add AO2 model/provider provenance fields without calling providers.
20. Add Forge GoalRun lifecycle, stop-gate, rollback, and no-provider boundary tests.
21. Add Arena independent repeated-task benchmark harness design.
22. Add Crucible failure-injection, fuzzing, and controlled-chaos harness design without live mutation.
23. Externalize bulk Atlas and Foundry evidence while retaining small replayable fixtures.
24. Generate Architecture status from source behavior and current lockfile data.
25. Add cross-platform golden-path replay and install/rollback checks.
26. Produce a beta readiness matrix with exact unresolved blockers.

Routing:
- AO Mission owns supervision, continuation, routing, recovery, and final return gates.
- AO Atlas owns workgraph state, context packs, one-node selection, and durable evidence.
- AO Foundry owns bounded implementation imports, readiness, and serialized PR sequencing.
- Route implementation to its owning repository; do not hide implementation in Atlas evidence.
- Use Blueprint only for genuinely new requirements, authorization, or a governed plan.
- AO Covenant owns policy and contract authority.
- AO2 owns execution-runtime design and provider boundaries.
- AO2 Control Plane owns observer storage and evidence-service behavior.
- AO Command owns compact operator readback only.
- AO Sentinel owns public-safety and freshness verdicts.
- AO Promoter owns promotion/no-promotion decisions and signed assurance inputs.

Execution loop:
1. Re-verify relevant repository status and record pre-existing dirty state without reverting it.
2. Create the Month 3 Atlas workgraph with exactly one executable node active.
3. Emit Foundry import/readiness evidence for exactly that node.
4. Implement bounded code in the owning repository when the node requires implementation.
5. Run targeted tests, full native verification, public-safety scans, and contract validation.
6. Commit on a codex/* branch, push, open a PR, wait for CI, merge, sync main, and delete branches.
7. Record node gate, candidate, rollback, tests, verification, Sentinel, Promoter, Command, run-link, checkpoint, and readback evidence.
8. Re-evaluate the next stop gate and continue automatically while ready work or an exact next action remains.

Implementation rules:
- Do not call providers or inspect credentials in this wave.
- Provider execution may only be designed as an explicit opt-in boundary; do not run it.
- Do not claim a feature is implemented from a fixture-only audit.
- Do not convert proposed compatibility matrices into passed gates without vectors and consumer tests.
- Do not call a dry-run promotion a production promotion.
- Do not claim a golden path until a replayable tiny-repository run produces reviewed PR and observer readback evidence.

Safety boundaries:
- no direct main mutation
- no credential, token, secret, or environment inspection
- no provider calls
- no release, deploy, publish, upload, or tag
- no dependency updates unless separately authorized
- no auth, policy, or config widening
- no hidden instruction mutation
- no concurrent mutation
- no broad public claim
- no broad RSI claim; RSI remains denied

Final response is allowed only when:
- all generated nodes are complete, or a true hard blocker remains after safe repair attempts
- completed_nodes >= 30 and preferably 40
- ready_nodes=0, blocked_nodes=0, failed_nodes=0
- final_response_allowed=true
- all required evidence validates and every run-link digest matches
- local verification and GitHub CI pass for touched repositories
- public-safety and stale-wording scans pass
- Promoter says no_promotion_requested or the exact no-promotion equivalent
- Command agrees with the final readback
- RSI remains denied
- touched repositories are clean and synced with origin/main
- local and remote codex/* branches are deleted

Final report must include:
- completed nodes / total nodes and all node statuses
- merged PRs by repository and final merge heads
- evidence roots and run-link digest status
- exact implementation capabilities proven versus gaps only audited
- final Promoter and Command rollups
- public-safety result
- verification commands and results
- clean/synced repository status
- at least 10 ranked Feature Depth Recommendations
- exact next action or exact hard blocker
```
