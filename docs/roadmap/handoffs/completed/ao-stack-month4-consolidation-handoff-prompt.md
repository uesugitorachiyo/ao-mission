> COMPLETED HISTORICAL HANDOFF - DO NOT REEXECUTE.
> Preserved for audit history only; use the active handoff directory for current execution.

# AO Stack Month 4 Consolidation Handoff

Use this prompt only after the current AO Atlas Month 3 real-run preparation
wave reaches a valid terminal handoff. Paste it into the existing AO Mission
task. Do not create a duplicate mission.

The section labeled `ADDITIONAL CODEX RECOMMENDATIONS` contains work added to
the original six-month roadmap. Keep that label in the workgraph, node IDs,
status readbacks, and final report so the operator can distinguish added work
from the baseline Month 4 scope.

```text
Start a Codex goal and resume the existing AO Mission for the AO Stack Month 4
Consolidation wave.

Do not use `ao-mission start` when the six-month roadmap mission already
exists. Reconstruct the mission ID, phase, accepted evidence, exact next action,
and current repository heads from durable Mission and Atlas readbacks before
creating the Month 4 workgraph.

Context:
- AO Atlas Month 3 final closure completed 30 of 30 nodes.
- Month 3 final closure main head:
  a458f01fedc9a6056d44f654e9aab134b84e9e76.
- Month 3 final closure reported 0 ready, 0 blocked, and 0 failed nodes.
- Month 3 final response was allowed.
- Month 3 final Promoter state was `no_promotion_requested`.
- Month 3 final public-safety validation passed.
- Promotion remained false and RSI remained denied.
- Month 3 final readiness source:
  <workspace>/ao-atlas/docs/evidence/ao-m3-final-closure-v01/nodes/mission-recommendation-month3-final-closure-30-final-report/month3-final-readiness-report.json
- Month 3 recommendation readback:
  <workspace>/ao-atlas/docs/evidence/ao-m3-final-closure-v01/nodes/mission-recommendation-month3-final-closure-30-final-report/recommendation-readback-after.json
- Month 3 evidence validation:
  <workspace>/ao-atlas/docs/evidence/ao-m3-final-closure-v01/evidence-validation-after-node-30.json

Month 4 entry gate:
1. The Month 3 real-run preparation wave must have a terminal validated
   readback. The 30-node final-closure result alone does not satisfy this gate.
2. Its minimum node budget must be satisfied.
3. It must report `ready_nodes=0`, `blocked_nodes=0`, `failed_nodes=0`, and
   `final_response_allowed=true`.
4. Every touched repository must be clean and synchronized with `origin/main`.
5. Required PRs and CI must be merged and green.
6. Promoter and Command must agree that no promotion occurred.
7. Public-safety validation must pass.
8. The terminal handoff must preserve current bounded RSI evidence and the
   denial of unrestricted RSI, policy-changing autonomy, and unrestricted
   self-modification.
9. If any entry condition is missing, remain in Month 3 and return the exact
   missing gate. Do not label preparation evidence as Month 4 execution.

Month 4 objective:
Consolidate the AO control and assurance surfaces behind stable compatibility
wrappers, externalize bulk evidence, preserve current CLI behavior, and prove
that migration can roll back without changing authority or losing bounded RSI
evidence.

Baseline Month 4 outcomes from the six-month roadmap:
1. Begin migration of Mission, Blueprint, Atlas, Foundry, Forge, and Command
   toward the accepted `ao-control` product boundary.
2. Preserve current repository implementations and CLIs until each component
   has a migration ADR, compatibility gate, and rollback proof.
3. Consolidate Arena, Crucible, Sentinel, and Promoter behind one assurance
   workspace and release boundary while keeping their evaluator, monitor, and
   promotion-decision responsibilities separate.
4. Move bulk campaign evidence from Git into signed CI artifacts or the
   content-addressed AO2 Control Plane store.
5. Retain compact signed manifests, canonical fixtures, operator readbacks,
   and public-safe reproduction instructions in Git.
6. Preserve existing CLI examples through compatibility wrappers.
7. Reduce Atlas and Foundry tracked generated JSON by at least 80 percent from
   the accepted pre-migration baseline without deleting required evidence
   before externalization, digest verification, and restore proof.
8. Produce one assurance release boundary that can replace four independent
   fixture release trains after parity gates pass.

Work budget:
- Generate 36 bounded Month 4 nodes.
- Complete at least 24 nodes before final response.
- If 24 nodes finish and ready work remains without a true blocker, continue
  toward all 36.
- Target a 2-3 hour supervised wave while ready work remains.
- Keep exactly one executable mutation node active at a time.
- Planning, readback, verification, and CI-wait nodes may proceed only when
  they cannot overlap a mutable write scope.
- Do not stop after one audit, one package, one repository, one PR, one CI
  pass, one evidence export, or one short batch.

Baseline priority order:
1. Re-verify Month 3 exit evidence and current repository truth.
2. Accept component-specific migration ADRs and rollback boundaries.
3. Freeze gate-critical contract behavior during moves.
4. Add compatibility wrappers before moving implementation.
5. Migrate one package or one bounded authority-neutral slice at a time.
6. Run producer-consumer conformance after every contract-affecting move.
7. Consolidate assurance packaging only after hosted CI and parity pass.
8. Externalize evidence only after content-addressed ingest, retrieval, and
   restore verification pass.
9. Update generated Architecture status and operator documentation.
10. Produce rollback-tested Month 4 closure evidence.

Baseline required work:
1. Record the accepted five-boundary topology and current source repositories
   in a digest-bound migration manifest.
2. Create a migration ADR for each control component before moving code.
3. Define package interfaces for Mission lifecycle, Blueprint authorization,
   Atlas workgraph/context compilation, Foundry scheduling, Forge GoalRun, and
   Command presentation.
4. Keep policy and approval authority in AO Covenant.
5. Keep provider execution and evaluator closure in AO2.
6. Keep observer storage and readback in AO2 Control Plane.
7. Keep Command read-only and prevent migrated presentation code from owning
   domain truth.
8. Add compatibility wrappers that preserve existing CLI commands, exit codes,
   JSON field names, contract versions, and public-safe denial fields.
9. Add a deprecation schedule only after a wrapper has consumer evidence and a
   rollback path.
10. Move no more than one bounded package or cohesive responsibility per PR.
11. Re-run the relevant cross-repository contract matrix after every merged
    migration PR.
12. Create the assurance workspace package boundaries for Arena, Crucible,
    Sentinel, and Promoter.
13. Preserve independent benchmark, adversarial, monitoring, and promotion
    verdict artifacts inside the consolidated assurance boundary.
14. Prevent Promoter from activating changes or treating observer evidence as
    approval authority.
15. Inventory tracked Atlas and Foundry evidence by class, size, digest, owner,
    retention requirement, and restore requirement.
16. Externalize bulk campaign evidence, long-run node records, and CI ledgers
    through the existing content-addressed observer path.
17. Keep small canonical fixtures and compact manifests in Git.
18. Verify every externalized artifact by digest after retrieval.
19. Run backup and restore against a clean observer data directory.
20. Update Architecture status from repository heads, CI, contracts, migration
    state, and evidence-location manifests.
21. Update install, migration, rollback, and operator inspection documentation.
22. Produce a Month 4 compatibility and rollback closure packet.

=======================================================================
ADDITIONAL CODEX RECOMMENDATIONS - ADDED TO THE ORIGINAL MONTH 4 ROADMAP
=======================================================================

Tag every node below with:

  recommendation_source=codex_additional_month4

These nodes are additional recommendations. They must remain separately
counted in the workgraph and final report.

ADDITIONAL-01: RSI continuity gate across consolidation
- Build a compact baseline manifest over the accepted bounded recursive-
  improvement, authority-denial, rollback, policy, Sentinel, Promoter, and
  Command fixtures that Month 4 must preserve.
- Re-run the same baseline before and after each migrated package.
- Compare canonical digests and semantic verdict fields.
- Block the migration when a bounded capability disappears, an authority
  denial weakens, a completion count changes, or an unrestricted RSI claim is
  introduced.
- This gate preserves current bounded RSI work. It does not create a new proof
  class or authorize unrestricted RSI.

ADDITIONAL-02: Differential replay gate for every compatibility wrapper
- Feed the same producer fixtures to the pre-migration CLI and its wrapper.
- Compare exit code, canonical JSON, denial fields, artifact digests, and exact
  next action.
- Require byte equality where the public contract requires it and documented
  semantic equality where paths, timestamps, or wrapper metadata differ.
- Reject wrapper-authored success when the owning producer rejects the input.

ADDITIONAL-03: Deterministic run-provenance envelope
- Add one shared provenance shape for model-backed and provider-free run
  records without calling a provider in this wave.
- Record requested and resolved provider, model, model profile, reasoning
  effort, CLI version, prompt digest, context-pack digest, repository commit,
  toolchain identity, policy digest, approval digest, sandbox profile,
  concurrency level, retry number, start and finish times, and fallback state.
- Make silent fallback invalid.
- Require consumers to preserve the envelope rather than copying a partial
  subset into new contracts.

ADDITIONAL-04: Evidence-growth delta guard
- Establish accepted Atlas and Foundry evidence baselines by tracked byte count,
  file count, and largest generated files.
- Start with a report-only PR delta that classifies new evidence as compact
  fixture, manifest, operator readback, or bulk externalizable evidence.
- Do not impose an arbitrary blocking threshold on historical files.
- After enough PR observations exist, set an evidence-backed blocking budget
  for new bulk evidence.
- Fail immediately when a PR commits an externalized artifact without its
  digest manifest or commits a manifest whose observer artifact cannot be
  retrieved and verified.

ADDITIONAL-05: Close assurance CI asymmetry before consolidation
- Confirm Arena and Crucible hosted CI remains green.
- Add the missing Sentinel hosted CI workflow with read-only permissions.
- Run Sentinel tests, vet or lint, build, schema validation, and public-safety
  scans on pull requests and pushes to the default branch.
- Do not consolidate the assurance release boundary until all four packages
  have hosted evidence and their responsibility-specific checks remain visible.

ADDITIONAL-06: Per-slice rollback journal and migration stop rule
- For every migrated package, record source head, target head, wrapper version,
  contract digests, verification commands, CI URLs, rollback command or revert
  commit, restored location, and post-rollback parity result.
- Exercise rollback before deleting or deprecating the source implementation.
- Stop Month 4 migration after two consecutive failed parity or rollback
  attempts on the same slice. Preserve evidence and return a bounded repair
  request instead of widening scope.

ADDITIONAL-07: Thin-module extraction rule
- Use consolidation to extract existing responsibilities from oversized files;
  do not combine extraction with unrelated behavior changes.
- Each extracted module must have one owner, one interface, focused tests, and
  no duplicate policy or release authority.
- Prioritize the exact code touched by migration wrappers. Do not launch a
  broad monolith rewrite during Month 4.

=======================================================================
END ADDITIONAL CODEX RECOMMENDATIONS
=======================================================================

Required migration sequence for each control package:
1. Read the owning repository implementation and tests.
2. Write a migration ADR with current owner, target package, dependency edges,
   compatibility wrapper, rollback, and excluded authority.
3. Capture a pre-migration differential replay baseline.
4. Add or strengthen tests before moving code.
5. Extract one focused package behind the current CLI.
6. Run targeted tests and the full owning-repository gate.
7. Run producer-consumer conformance and RSI continuity checks.
8. Open one reviewable PR, wait for CI, merge, sync main, and delete the branch.
9. Exercise or rehearse rollback and record the journal entry.
10. Advance to the next package only after parity and rollback pass.

Required assurance consolidation sequence:
1. Verify Arena, Crucible, Sentinel, and Promoter repository heads and hosted
   CI status.
2. Close missing hosted CI before moving package code.
3. Define independent package interfaces and artifact types.
4. Preserve separate benchmark, adversarial, monitoring, and promotion
   decisions.
5. Run the same fixture corpus through old CLIs and consolidated wrappers.
6. Prove Promoter cannot activate, publish, deploy, or mutate through the
   assurance workspace.
7. Produce one release-boundary plan and rollback proof without publishing a
   release in this wave.

Required evidence externalization sequence:
1. Generate a tracked evidence inventory and immutable pre-migration baseline.
2. Classify every candidate artifact before moving or deleting it.
3. Ingest bulk evidence through the existing AO2 Control Plane content-
   addressed path.
4. Record canonical digest, artifact kind, source repository, source commit,
   retention class, and retrieval path in a compact Git manifest.
5. Fetch each artifact into a clean temporary directory and recompute its
   digest.
6. Run backup and restore against a clean control-plane data directory.
7. Remove tracked bulk evidence only after ingest, retrieval, restore, PR
   review, and rollback evidence pass.
8. Recompute Atlas and Foundry tracked evidence reductions.

Routing:
- AO Mission owns mission supervision, durable phase state, recovery, and the
  final return gate.
- AO Blueprint owns requirements and build authorization.
- AO Atlas owns workgraph and context compilation plus one-ready-node handoff.
- AO Foundry owns safe-next-work selection and serialized migration sequencing.
- AO Forge owns one GoalRun and bounded per-run orchestration.
- AO Covenant owns schema registry, policy, approval, trust, and revocation.
- AO2 owns execution, transactional mutation, provider adapters, and evaluator
  closure.
- AO2 Control Plane owns observer storage, content addressing, indexing,
  retention, backup, restore, and read APIs without approval authority.
- AO Command owns read-only operator presentation.
- AO Arena owns benchmarks.
- AO Crucible owns adversarial and failure-injection probes.
- AO Sentinel owns runtime, CI, policy, freshness, and safety monitoring.
- AO Promoter owns promotion readiness decisions without activation authority.
- AO Architecture owns current topology, migration ADRs, compatibility status,
  and generated source-of-truth documentation.

Execution rules:
1. Record all pre-existing dirty state and do not revert unrelated work.
2. Keep exactly one executable mutation node active.
3. Use isolated branches and worktrees for implementation.
4. Do not change behavior and move packages in the same slice unless the
   behavior change is required to preserve an existing public contract.
5. Run targeted tests before full repository verification.
6. Use existing repository auth only for normal branch, PR, CI, and merge
   operations without exposing credentials.
7. Wait for required CI before merge.
8. Sync the default branch and delete local and remote task branches after a
   successful merge.
9. Record node gate, candidate, rollback, tests, verification, Sentinel,
   Promoter, Command, run-link, checkpoint, and exact next action after each
   node.
10. Continue automatically while ready nodes or an exact next action remain.

Verification:
- AO Mission, Blueprint, Atlas, Foundry, Forge, Covenant, Command, Arena,
  Crucible, Sentinel, and Promoter: run their complete Go test, vet, build, and
  repository-specific readiness gates.
- AO2: run the complete Rust workspace tests plus its documented verification
  and public-safety gates for touched code.
- AO2 Control Plane: run the complete Rust workspace tests, storage tests,
  ingest/retrieval checks, retention checks, backup/restore drills, and
  repository-specific readiness gates.
- AO Architecture: run topology, lockfile, contract inventory, contract owner,
  compatibility, readiness-wording, and architecture verifiers.
- Run `git diff --check` in every touched repository.
- Run scoped secret, local-path, unsafe-authority, release, provider, and RSI
  wording scans over every changed surface.

Safety boundaries:
- no direct default-branch mutation
- no credential, token, secret, cookie, or environment inspection
- no provider calls in this wave
- no release, deploy, publish, upload, or tag
- no dependency update unless separately authorized
- no auth, policy, config, repository-scope, or approval widening
- no hidden instruction mutation
- no concurrent mutation
- no automatic promotion or activation
- no deletion before classification, externalization, restore, and rollback
- no claim that fixture evidence proves live execution
- no new broad RSI claim
- unrestricted RSI, policy-changing autonomy, and unrestricted self-
  modification remain denied

Final response is allowed only when:
- the Month 4 minimum node budget is satisfied;
- no ready nodes or exact next actions remain for the generated wave;
- `blocked_nodes=0` and `failed_nodes=0`;
- `final_response_allowed=true`;
- every migration slice has compatibility and rollback evidence;
- all required local verification and GitHub CI pass;
- assurance packages preserve independent verdicts;
- externalized evidence passes ingest, retrieval, digest, backup, and restore;
- the RSI continuity gate reports no lost bounded capability and no weakened
  denial;
- Promoter reports `no_promotion_requested` or the exact no-promotion
  equivalent;
- Command agrees with the terminal readback;
- public-safety scans pass;
- touched repositories are clean and synchronized with origin/default branch;
  and
- local and remote task branches are deleted after merge.

Final report must separate:
1. Baseline Month 4 roadmap nodes.
2. `codex_additional_month4` nodes.

For each group report:
- completed, ready, blocked, and failed counts;
- merged PRs and final repository heads;
- implemented capabilities versus fixture-only or design-only results;
- contract and wrapper parity;
- rollback results;
- evidence externalization totals and retained Git footprint;
- assurance CI and responsibility parity;
- RSI continuity result and unchanged denied boundaries;
- Promoter, Command, and public-safety readbacks;
- verification commands and results;
- remaining blockers and exact next action; and
- at least 10 ranked recommendations for Month 5 beta hardening.
```
