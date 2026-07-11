> COMPLETED HISTORICAL HANDOFF - DO NOT REEXECUTE.
> Preserved for audit history only; use the active handoff directory for current execution.

# AO Mission Continuation: Foundry Blueprint Contract Bridge Plan

Continue the existing AO Mission `mission-4d91b0a9e4ab273e`. Do not create a
new mission or replace its objective.

## Operator Review Decision

I reviewed the AO Stack Six-Month Productization Design at:

- repository: `ao-architecture`
- branch: `codex/ao-stack-productization-spec`
- commit: `e009428`
- file:
  `docs/superpowers/specs/2026-07-09-ao-stack-six-month-productization-design.md`

The design is approved for implementation planning, subject to the constraints
in this continuation. This approval authorizes read-only investigation and
creation of the first implementation plan in an isolated Foundry worktree. It
does not authorize implementation, AO2 provider execution, patch application,
merge, push, release, deployment, publication, upload, tagging, credential use,
or mutation of any default branch.

Use `codex_sol_high`: provider `codex`, model `gpt-5.6-sol`, reasoning effort
`high`, for this contract-sensitive planning work. Do not downgrade to Terra or
Luna. Record requested and resolved profile, model, and reasoning effort in the
planning evidence. The downstream AO2 provider path remains disabled.

## Verified Starting State

Preserve and revalidate these facts before planning:

- Mission status is `active`, phase is `atlas_workgraph_ready`, and current
  route is `ao-foundry`.
- Mission execution and promotion authority remain false.
- Blueprint authorization ref is
  `.ao-mission/handoffs/mission-4d91b0a9e4ab273e/blueprint-pack/build-authorization.json`.
- Blueprint authorization digest is
  `sha256:b7e18a1967b31b2806184444ab9aeab5e984e050f66261431ec57ece4cc833ee`.
- Atlas workgraph digest recorded by Mission is
  `sha256:f5274e17dc0835be69cf1bc97967f83dca4b8e83eb2bc0e4c3c9c0d0d755fa8c`.
- The Mission record contains a duplicate reference to that same Atlas
  workgraph. Treat this as idempotent readback hygiene, not a second
  authorization or second ready node.
- The producer-owned Blueprint contract identifies itself with
  `schema: "ao.blueprint.build-authorization.v0.1"`.
- Foundry `loadPulseIntakeSource` currently reads only `schema_version` and
  `contract_version`, so the canonical producer artifact fails with
  `unexpected source artifact schema ""`.
- Running Foundry from inside its repository with sibling paths beginning in
  `..` fails the separate public-safe path guard. Do not weaken that guard.
  The integration test must execute a built Foundry binary from the shared
  workspace root with repository-relative paths.

Recompute every digest from source bytes. Stop if any ref, digest, branch, or
commit differs.

## Workgraph Reconciliation

Create an Atlas repair/repack node before the current AO2 P0-A node. Name it
`foundry-blueprint-authorization-contract-bridge` or an equivalent stable ID.
The existing AO2 P0-A node must depend on this repair node. Preserve the rest of
the P0 sequence and all denied authorities.

The repair node has one primary repository owner: `ao-foundry`. Blueprint owns
the authorization schema and bytes; Atlas owns the import binding; Foundry owns
the consumer repair. Do not change the Blueprint artifact to match Foundry. Do
not create a generic adapter artifact or accept arbitrary schema aliases.

Repack and re-import the workgraph through the normal Atlas and Mission
contracts. Do not edit recorded Mission or Atlas evidence in place. Reconcile
the duplicate Mission artifact reference idempotently if the current Mission
contract supports it; otherwise record it as residual readback debt without
changing node counts.

## Planning Scope

Use the `superpowers:writing-plans` workflow. Write one plan for this Foundry
repair only. Save it in an isolated `ao-foundry` worktree as:

`docs/superpowers/plans/2026-07-09-blueprint-authorization-pulse-compatibility.md`

The plan must map exact files before tasks. Prefer a focused Blueprint
authorization consumer and focused tests over adding more responsibility to
the oversized `internal/cli/cli.go` and `internal/cli/cli_test.go`. Keep the
CLI call-site change minimal and follow existing `internal/cli` package
patterns. Do not include unrelated module splitting or cleanup.

The plan must use failing-first tests and include complete code for every code
step. Its first red test must pass the canonical producer artifact with the
`schema` field and demonstrate the current failure. Add negative tests that
reject:

- missing schema identity;
- a wrong schema;
- a `schema_version`-only replacement presented as canonical Blueprint
  authorization;
- blocked status or `approved_by_user=false`;
- missing required producer fields;
- unsafe or parent-traversing paths;
- authorization bytes whose SHA-256 differs from the Atlas import binding;
- a mismatched project or Blueprint pack digest; and
- Atlas artifacts that claim scheduling, execution, approval, mutation,
  provider-call, or release authority.

The implementation design must parse the producer-owned contract explicitly,
preserve the input bytes, compute the source SHA-256 from those exact bytes,
and compare it with the Atlas `build_authorization.digest`. It must not rewrite
the authorization, silently normalize it into a Foundry schema, weaken public
path validation, or treat a 100 score alone as build authorization.

## Required Verification In The Plan

Include exact commands and expected results for:

1. the focused red and green Foundry tests;
2. `go test ./... -count=1` in `ao-foundry`;
3. `go vet ./...` in `ao-foundry`;
4. `go build ./cmd/foundry` in `ao-foundry`;
5. producer contract or fixture verification in `ao-blueprint`;
6. Atlas import validation using the existing mission artifacts; and
7. a workspace-root end-to-end Pulse preflight using a freshly built Foundry
   binary and the exact Blueprint and Atlas files above.

The end-to-end result must be `ready` while all authority flags remain false.
The authorization SHA-256 before and after the run must remain
`sha256:b7e18a1967b31b2806184444ab9aeab5e984e050f66261431ec57ece4cc833ee`.

## Return Gate

Do not implement the plan in this continuation. Commit only the plan on the
isolated Foundry planning branch. Then return:

- the plan path, branch, and commit;
- the design-review result and any corrections incorporated;
- the exact planned file and test scope;
- the reproduced red-test failure;
- the repacked Atlas repair-node ID and new workgraph digest, or the exact
  blocker if Atlas cannot repack without implementation;
- the requested and resolved model profile evidence;
- the exact authorization text required to begin implementation; and
- `final_response_allowed=false`.

Do not continue to AO2 P0-A, write implementation code, or clear Foundry
readiness until the operator reviews and explicitly authorizes the committed
implementation plan.
