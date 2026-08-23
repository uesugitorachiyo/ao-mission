# AO Cross-Platform Development Baseline Handoff

Use AO Mission as the durable coordinator for this campaign. Execute the plan
at `docs/superpowers/plans/2026-08-22-cross-platform-development-baseline-campaign.md`
sequentially and checkpoint after every slice.

```text
Start one correlation-bound AO Mission with this objective:

Create and independently qualify one exact AO full-source development baseline
that materializes from clean inputs and produces semantically equivalent
credential-free results on macOS and Windows, without beginning AO Office Pool
or widening release, provider, deployment, promotion, compatibility,
credential, external-beta, or RSI authority. Coordinate this as one bounded
implementation workgraph.

Correlation ID:
ao-cross-platform-development-baseline-20260822-r2

Authoritative design:
ao-architecture/docs/superpowers/specs/2026-08-22-cross-platform-development-baseline-design.md
Architecture design commit: f8c8c1d

Execution model:
- Default to gpt-5.6-terra with medium reasoning.
- Use gpt-5.6-luna low only for exact mechanical inventory, hashing, schema
  formatting, or status readback after inputs and acceptance criteria freeze.
- Escalate only the affected slice to gpt-5.6-sol high for unexplained semantic
  parity, cleanup-identity ambiguity, or contradictory final evidence.
- Do not use max, ultra, or subagents unless the operator explicitly changes
  this authority.

Fresh-start intake:
- Use the AO Mission source checkout containing this reviewed handoff; invoke
  it with `go run ./cmd/ao-mission`, not an older installed binary or image.
- Record the exact AO Mission source commit before intake.
- Use a new private, empty `AO_MISSION_HOME` outside every source checkout and
  outside every previous campaign root. Reject an existing or non-empty root.
- Create a new Mission identity with the `-r2` correlation ID above. Preserve
  earlier reduced-route readbacks as diagnostic evidence, but never resume or
  import them into this campaign.
- Require the intake readback to report `routing_class=complex`,
  `initial_route=ao-atlas`, AO Atlas `required`, and AO Blueprint `omitted`.
  Stop before S01 if any field differs.
- AO Blueprint remains mandatory in the independently exercised S04 fixture;
  its omission from accepted Mission intake does not remove it from the
  full-stack qualification path.

Known intake pitfall:
- The earlier objective text lacked a complex-routing marker and therefore
  deterministically classified as `reduced`. The same behavior can surface in
  Docker Windows testing on macOS, but it is not platform-specific. Clean state
  prevents stale-ledger reuse; the bounded-workgraph wording above fixes the
  objective classification without changing Mission's global routing rules.

Run sequential slices:
S01 baseline contract and manifest;
S02 safe materialization and native preflight;
S03 all repository-owned development gates;
S04 complete credential-free AO workflow;
S05 independent semantic parity comparison;
S06 clean hosted macOS/Windows qualification and evidence closure;
S07 final Mission reconciliation and bounded baseline declaration.

Hard gates:
- Review each slice design and implementation plan before code mutation.
- Start only from the fresh Mission intake defined above; never reuse a failed
  Mission identity, correlation ID, Mission home, or consumed run root.
- Use TDD for every new verifier, controller, runner, comparator, and cleanup
  behavior.
- Never continue past a failed slice or silently change the frozen baseline.
- Apply portability fixes only in the repository that owns the failure, rerun
  its native gates, and freeze a new baseline identity.
- Mission owns lifecycle and evidence reconciliation, not execution or approval.
- Preserve completed historical Windows evidence unchanged and create new
  self-contained parity evidence.
- Use new empty roots. Never overwrite, merge, reset, or clean an existing
  checkout.
- Keep provider calls, credentials, releases, deployments, publication,
  promotion, compatibility activation, external beta, and RSI denied.
- Do not begin AO Office Pool.

Terminal success requires the same reviewed 14-repository identity on both
platforms, all mandatory development gates passing, the complete AO fixture
reaching the same semantic terminal state, zero undeclared differences, zero
run-owned residue, independently rehashed evidence, green merged-main checks,
and a final Architecture/Mission baseline declaration.

Stop after S07. Return the frozen baseline digest as the only dependency input
for a later, separately authorized AO Office Pool objective.
```
