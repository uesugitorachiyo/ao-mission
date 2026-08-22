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
credential, external-beta, or RSI authority.

Correlation ID:
ao-cross-platform-development-baseline-20260822

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
