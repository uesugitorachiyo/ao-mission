# Evidence-Bound Slice Checkpoints Design

## Purpose

Allow AO Mission to record one durable, authority-neutral checkpoint for each
validated campaign slice without fabricating a lifecycle transition. The
existing `checkpoint create` behavior remains unchanged for callers that do
not supply slice evidence.

This capability closes a mismatch discovered while executing the approved AO
cross-platform development-baseline campaign: neutral evidence imports retain
artifacts but intentionally do not change route, phase, or next action, while
ordinary checkpoint creation deduplicates an unchanged lifecycle snapshot.

## Considered Approaches

### Recommended: evidence-bound checkpoint mode

Extend `checkpoint create` with paired `--slice` and `--evidence-digest`
arguments. A validated pair appends an ordinary v0.3 Mission checkpoint whose
`result` encodes the passed slice and exact evidence digest. Existing schema,
record, bundle, archive, and readback shapes remain compatible.

This approach is minimal, deterministic, auditable, and preserves the
no-execution boundary.

### Rejected: fabricate a lifecycle transition

Changing route, phase, status, or next action merely to defeat checkpoint
deduplication would corrupt the durable Mission lifecycle and violate neutral
evidence semantics.

### Rejected: treat artifact import as the checkpoint

An artifact reference is durable evidence but does not satisfy the campaign's
explicit checkpoint-count and latest-checkpoint requirements. Conflating the
two would weaken the plan instead of implementing it.

## Command Contract

The existing command remains valid and idempotent:

```text
ao-mission checkpoint create --mission <id> [--json]
```

Evidence-bound mode is:

```text
ao-mission checkpoint create --mission <id> --slice S01 --evidence-digest sha256:<64-lowercase-hex> [--json]
```

Rules:

- `--slice` and `--evidence-digest` must be supplied together.
- Slice must be exactly one of `S01` through `S07`.
- Evidence digest must be `sha256:` followed by 64 lowercase hexadecimal
  characters.
- Mission must already retain an artifact with the exact digest. A path,
  mutable reference, or unretained digest is rejected.
- The retained content is rehashed before checkpoint creation. Drift or a
  missing retained object fails closed.
- Evidence JSON must identify the same Mission correlation, name the same
  slice, report `result: "pass"`, and preserve false execution, approval,
  repository-mutation, provider, credential, release, publication, deployment,
  promotion, compatibility, external-beta, and RSI authority.
- The checkpoint result is
  `slice_pass:<slice>:sha256:<64-lowercase-hex>`.
- Exact replay returns the existing checkpoint bundle without appending.
- A second digest for an already checkpointed slice is rejected as a conflict.
- Slice order is strict: S01 may follow only non-slice checkpoints; each later
  slice must immediately follow the prior numbered slice.

## State And Authority Semantics

Evidence-bound creation appends only one `MissionCheckpoint`. It preserves
Mission status, current route, current phase, blockers, steps, route history,
exact next action, artifact refs, correlated imports, approval state, and all
authority boundaries.

The checkpoint retains the current route, phase, iteration, exact next action,
and resume command. Its result carries the immutable evidence binding. It does
not claim that Mission executed, approved, scheduled, promoted, published, or
released anything.

## Validation And Failure Behavior

Validation occurs before mutation. Invalid flags, unknown slices, uppercase or
malformed digests, missing retained evidence, digest drift, identity mismatch,
non-pass results, widened authority, out-of-order slices, and conflicting
replays return errors and leave both the Mission record and checkpoint bundle
unchanged.

The implementation uses the existing recoverable record/checkpoint transaction
and content-addressed evidence store. No network, provider, credential, or
repository operation is introduced.

## Testing

Tests prove:

- legacy checkpoint creation remains idempotent;
- S01 appends exactly once from retained passing evidence;
- exact replay is idempotent;
- S02 is rejected before S01 and accepted after it;
- conflicting evidence for one slice is rejected;
- paired flags, slice names, digest syntax, retained-byte integrity,
  correlation, result, and all denied-authority fields fail closed;
- failures do not mutate Mission workflow state or checkpoint count; and
- CLI JSON and text readbacks preserve the existing false-authority boundary.

## Approval

The operator granted standing approval on 2026-08-22 to complete the remaining
campaign proactively without pausing for routine permission prompts. Technical
verification and denied-authority gates remain mandatory.
