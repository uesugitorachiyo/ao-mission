# Bounded Autonomy Repair Closure: Reopened From Month 3

Evidence directory:
`ao-stack-bounded-autonomy-repair-from-month3-20260716T130827Z`.

## Outcome

The prior bounded-autonomy Month 1-6 closure is classified as
`PARTIAL_INVALID_CLOSURE`.

The useful merged PRs and green CI remain useful, but the prior Month 3-6
closure evidence was not accepted as final because Month 3 ended blocked and
later months depended on insufficient evidence.

This repair reopens from Month 3 and closes the repaired sequence with AO2 held
at `v0.5.1` and AO2 Control Plane patched to `v0.1.16`.

## Month 3 Repair

The prior evidence recorded:

- final reconciliation `status=blocked`
- Mission and Command still `active`
- `completed_nodes=0`
- `final_response_allowed=false`
- final rollup still had an exact continuation action

AO Mission now has a regression for terminal final rollups:

- `TestFinalRollupClearsExactNextActionWhenFinalResponseAllowed`

`BuildFinalRollup` clears `exact_next_action` when the return gate allows final
response. Ready-node denial behavior remains tested separately.

The repaired terminal run records:

- final reconciliation `status=ready`
- artifacts agree
- Mission and Command are `done`
- `completed_nodes=26`
- `ready_nodes=0`
- `final_response_allowed=true`
- final rollup `exact_next_action=""`

## Month 4 Repair

Rollback verification was rerun in isolated clones.

The old AO Mission rollback command conflicts on current `main` because later
Month 5 and Month 6 edits touched the same files. The repair records that
conflict and also verifies an explicit path-level rollback and restore method.

AO Command rollback of PR #132 applied cleanly in an isolated clone and restore
matched the original tree.

## Month 5 Repair

Month 5 was rerun as a genuine five-task portfolio:

1. AO Mission final-rollup reliability fix.
2. Month 3 recovery execution.
3. Month 4 rollback execution.
4. Control Plane binary-impact review.
5. Denied-authority boundary verification.

No prior Month 3 or Month 4 task was reused as a substitute.

## Month 6 Repair

Month 6 qualification was repeated.

Decision: `control_plane_patch_release`.

AO2 remains `v0.5.1`.

AO2 Control Plane is `v0.1.16`.

The Control Plane `spin` lockfile change is explicitly classified as a compiled
dependency impact, not metadata. Exact Control Plane stable-patch qualification
passed, public assets were verified, and the current public pair is AO2
`v0.5.1` plus AO2 Control Plane `v0.1.16`.

## Boundaries

- RSI remains denied.
- Live self-modification remains denied.
- The compatibility gate remains `ready`, not active.
- External beta has not launched.
- Promotion was not requested or granted.
- No provider pilot ran.
- No AO2 release occurred. The only release/tag/upload/new binary publication
  in this repair is AO2 Control Plane `v0.1.16`; no deployment occurred.
- No credentials were inspected.
- No `/tt` or modules work occurred.

## Next Action

Do not start the next roadmap from the invalidated closure. Use this repaired
closure as the baseline before considering any new roadmap.
