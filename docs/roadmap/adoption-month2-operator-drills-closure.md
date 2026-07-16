# Adoption/Evidence Month 2 Closure: Operator Adoption Drills

Status: closed on 2026-07-16
Decision: operator adoption drills passed with the current public pair and
refreshed evidence base
Evidence directory:
`canary-test/ao-stack-adoption-month2-operator-drills-20260716T051751Z`

## Objective

Month 2 tested whether a solo operator can read the current AO Stack state,
identify the public release pair, understand the compatibility gate, follow
safe-next-work and run-state readbacks, inspect policy and observation evidence,
collect support evidence, and avoid unsupported release, external beta,
promotion, provider, or RSI claims.

This was not a release train. No release, tag, upload, deployment, provider
pilot, external user contact, promotion, external beta launch, live
self-modification, or RSI work was authorized or performed.

## Baseline

Current public pair:

- AO2 `v0.5.1`
  - Release: https://github.com/uesugitorachiyo/ao2/releases/tag/v0.5.1
  - Tag target: `80ec5321f42d4bab17d5e64fdae6aa099ba59d4a`
- AO2 Control Plane `v0.1.15`
  - Release: https://github.com/uesugitorachiyo/ao2-control-plane/releases/tag/v0.1.15
  - Tag target: `f1702b387607566cac457458af9adb5871a5c412`

Adoption/Evidence Month 1 left AO Architecture with 16 total compatibility
edges, 16 tested edges, 16 canonical vectors, 16 consumer tests, 0 proposed
edges, and compatibility gate state `ready`, not active.

## Completed Nodes

| Repo | PR | Merge commit | Result |
| --- | --- | --- | --- |
| AO Architecture | https://github.com/uesugitorachiyo/ao-architecture/pull/127 | `807c7b7d3b76e780afaa249539ebbbaa6a150eec` | Adoption drill source of truth and verifier |
| AO Command | https://github.com/uesugitorachiyo/ao-command/pull/126 | `899d9a696389823b52994b3d72ad0511c447c66f` | Operator adoption readback fixture and tests |
| AO Foundry | Existing evidence | `9c7bc148` | Month 5 safe-next-work fixture already satisfies the drill |
| AO Forge | Existing evidence | `b737611` | Month 5 run-state fixture already satisfies the drill |
| AO Covenant | Existing evidence | `b3aacf4` | Month 5 policy readback fixture already satisfies the drill |
| AO2 | Existing evidence | `6f8ec3a` | Support reproduction docs and tests cover install, checksum, manifest, approval/replay, rollback, and operator evidence |
| AO2 Control Plane | Existing evidence | `867bf38` | Observation and public-pair readback tests satisfy the drill |
| AO Sentinel | https://github.com/uesugitorachiyo/ao-sentinel/pull/48 | `757e55728fc3279f738439e78aa5aa357ac7b348` | Adoption wording and overclaim checks |
| AO Promoter | https://github.com/uesugitorachiyo/ao-promoter/pull/58 | `e5f90ca1e662a0d52c298e8f452eeff9b3aff1fd` | No-promotion/no-RSI adoption verdict |

Blocked nodes: none.

## Operator Drill Result

The adoption drill proves an operator can read:

- current public pair;
- 16/16 compatibility matrix state;
- compatibility gate state `ready`, not active;
- safe-next-work summary;
- run-state handoff summary;
- policy approval and denied authority summary;
- observation and support evidence categories;
- denied states for RSI, external beta, promotion, provider pilot, release,
  tag, upload, deployment, and live self-modification;
- next safe operator action.

AO Command now has a tested adoption readback fixture:
`examples/operator/adoption-month2-operator-drill-readback.json`.

## Support Evidence Result

AO2 support evidence paths remain sufficient for:

- install;
- checksum;
- manifest mismatch;
- approval/replay;
- rollback;
- operator readback issue reporting.

The AO2 support checks ran against the existing public support fixtures and
docs. No AO2 change was required for this Month 2 drill.

## Policy And Observation Result

AO Covenant already records approval-required, denied authority,
revocation/rollback, no provider, no live self-modification, no promotion, and
no RSI authority in the Month 5 policy readback fixture.

AO2 Control Plane observation tests passed for public release-pair verification
and active stack release handoff readback. No Control Plane change was required.

## Sentinel And Promoter Result

AO Sentinel now catches adoption drill overclaims, including:

- compatibility gate ready mistaken for active;
- external beta launch claims;
- promotion claims;
- RSI claims;
- provider-pilot claims;
- release claims;
- missing denied-state boundaries.

AO Promoter now records the adoption drill verdict:

- `promotion_requested=false`;
- `promotion_granted=false`;
- `rsi_authorized=false`;
- `external_beta_launched=false`;
- `compatibility_gate_active=false`.

## Verification Summary

Cross-repo readback passed after implementation PRs merged:

```sh
python3 scripts/verify_adoption_operator_drill.py
python3 scripts/verify_evidence_freshness.py
python3 scripts/verify_current_release_manifest.py
python3 scripts/verify_compatibility_matrix.py
python3 -m unittest discover scripts
go test ./internal/cli -run 'TestAdoptionMonth2OperatorDrillReadback|TestAdoptionMonth1CompatibilityGateReadback|TestMonth5OperatorWorkflowReadback|TestMonth6NoReleaseOperatorWorkflowReadback' -count=1
go test ./internal/cli -run 'TestMonth5SafeNextWorkOperatorFixture|TestFoundrySafeNextWorkCompatibilityVectorProducesForgeGoalRun' -count=1
go test ./internal/cli -run 'TestMonth5RunStateOperatorFixture' -count=1
go test ./internal/cli -run 'TestMonth5OperatorPolicyReadbackFixture' -count=1
python3 -m pytest tests/test_public_stabilization.py tests/test_release_publish_approved_assets.py tests/test_month4_controlled_self_improvement.py -q
cargo test --workspace --quiet
python3 -m pytest tests/test_public_release_pair_verify.py tests/test_active_stack_release_handoff_readback.py -q
go test ./internal/cli -run 'TestAdoptionMonth2OperatorDrillWordingProfile|TestAdoptionMonth1GateReadinessWordingProfile|TestMonth5OperatorWorkflow' -count=1
go test ./internal/cli -run 'TestAdoptionMonth2OperatorDrillNoPromotionFixture|TestAdoptionMonth1GateReadinessNoPromotionFixture|TestMonth5OperatorWorkflowNoPromotionFixture' -count=1
```

Full local test suites also passed for the changed Go repos. Hosted CI passed
for all implementation PRs before merge.

## Boundaries

- RSI remains denied.
- Live self-modification remains denied.
- Compatibility gate is `ready`, not active.
- External beta has not launched.
- Promotion is not requested or granted.
- Provider pilot did not run.
- No release was published.
- No tag was created or moved.
- No upload occurred.
- No deployment occurred.
- No new binary publication occurred.
- No `/tt` or modules work occurred.
- Credentials were not inspected.

## Month 3 Recommendation

Start Month 3 with evidence maintenance automation using the refreshed evidence
base and operator drill results. Do not start a release, external beta,
promotion, provider pilot, live self-modification, or RSI work without separate
exact-scope authorization.
