# AO Stack Bounded Autonomy Month 1 Closure: Operational Baseline

Status: closed.

Evidence ID:
`ao-stack-bounded-autonomy-month1-6-plus-stable-release-20260716T071830Z`.

## Objective

Month 1 established the bounded-autonomy operational baseline for the current
public AO Stack pair and created a repeatable benchmark corpus for Months 2-6.

## Current Public Pair

- AO2 `v0.5.1`
  - Release: https://github.com/uesugitorachiyo/ao2/releases/tag/v0.5.1
  - Tag target: `80ec5321f42d4bab17d5e64fdae6aa099ba59d4a`
  - Public asset count: 23
  - Approved manifest digest:
    `bd8103e7a038f47e1b4fef1a2a19ae65cc221675ea11149d39cfb679ae2a08fc`
- AO2 Control Plane `v0.1.15`
  - Release:
    https://github.com/uesugitorachiyo/ao2-control-plane/releases/tag/v0.1.15
  - Tag target: `f1702b387607566cac457458af9adb5871a5c412`
  - Public asset count: 6

The AO Architecture current-release manifest, stack lock, compatibility matrix,
evidence freshness, and evidence maintenance verifiers passed. The
compatibility gate remains `ready`, not active.

## Baseline Benchmark

AO Architecture now records:

- bounded-autonomy benchmark corpus
- bounded-autonomy result schema
- Month 1 baseline results
- verifier and tests for required task classes, metrics, failure classes, and
  denied authority boundaries

The benchmark covers:

- documentation correction
- deterministic single-repository code fix
- cross-repository contract update
- approval-required mutation
- rollback-required mutation
- failed-CI diagnosis and repair fixture
- interrupted mission and resume fixture

Baseline metrics are machine-readable and include completion rate, first-pass
verification rate, recovery rate, human approvals, human interventions,
duplicate work, orphan branches, ready nodes, retries, rollback result,
evidence integrity, escaped defects, and unsupported claims.

## Operator Readback

AO Command now presents the bounded-autonomy Month 1 benchmark baseline with:

- benchmark version
- baseline status
- task class count
- completion, first-pass verification, and recovery rates
- rollback result
- unsupported claim count
- current public pair
- compatibility matrix and gate state
- denied release, provider, promotion, external beta, and RSI states

## PRs

- AO Architecture PR #131:
  https://github.com/uesugitorachiyo/ao-architecture/pull/131
  - Merge commit: `d96032d5faf6551a68006c88efb66da0f4b7c80f`
- AO Command PR #129:
  https://github.com/uesugitorachiyo/ao-command/pull/129
  - Merge commit: `4b96af831346afb7315ff1084af756dd1a932727`

## Verification

- AO Mission production readiness: 100/100 ready
- AO Mission `go test ./...`: passed
- AO Architecture unit tests and benchmark verifier: passed
- AO Architecture current-release, stack-lock, compatibility matrix, evidence
  freshness, and evidence maintenance verifiers: passed
- AO Command focused bounded-autonomy readback test: passed
- AO Command `go test ./...`: passed
- PR CI: green for AO Architecture and AO Command
- Artifact guards: passed
- Private-info scans: passed

## Boundaries

- RSI remains denied.
- Live self-modification remains denied.
- Compatibility gate was not activated.
- External beta was not launched.
- Promotion was not requested or granted.
- No provider pilot ran.
- No release, tag, upload, deployment, or new binary publication occurred.
- No `/tt` or modules work occurred.
- No credentials were inspected.

## Month 2 Handoff

Month 2 should execute end-to-end engineering workflows using this benchmark
baseline:

1. documentation or support correction
2. deterministic single-repository code change
3. cross-repository producer/consumer contract change

Month 2 is not a release train by default.
