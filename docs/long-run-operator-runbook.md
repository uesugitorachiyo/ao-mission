# AO Mission Long-Run Operator Runbook

Use this runbook when the operator asks for a 2-3 hour AO run, a doubled
recommendation wave, a minimum node budget, or a correction that must not return
after a short batch.

## Request Shape

Start with an explicit lease and node budget:

```sh
ao-mission start "<objective>"
ao-mission continue --mission <mission-id> --until-done --max-iterations 60 --min-nodes 30 --min-minutes 120 --max-minutes 180 --return-only-when mission_done_or_true_hard_blocker_or_no_ready_work_and_no_exact_next_action --checkpoint-policy after_each_node_or_timed_interval
```

Use at least 30 bounded nodes for a doubled wave. Use 120-180 minutes for the
time lease. The mission may continue past the minimum node count when ready work
or exact next actions remain.

## Role Routing

AO Mission owns the long-run lease, checkpoint policy, route reconciliation,
return gate, final rollup, and feature-depth continuation pressure.

AO Atlas owns the workgraph, context-heavy sequencing, exact next node
selection, and final-state reconciliation packet. Atlas must refuse a final
response while ready nodes, missing evidence, or exact next actions remain.

AO Foundry owns exactly one bounded implementation node at a time. It produces
node gate, candidate, rollback, implementation, test, verification, Sentinel,
Promoter, and Command/readback evidence for that node.

AO Command owns compact status readback across Mission, Atlas, Foundry,
Promoter, checkpoint freshness, return gate, and early-return risk. It remains
read-only.

AO Blueprint is not a batching queue. Use Blueprint only for missing
requirements, missing authorization, or an unclear safety boundary. Ready Atlas
nodes route directly to Foundry, not back through Blueprint.

## Stop Gate

Do not stop after one PR, one CI pass, one Foundry import, one route decision,
one rollup, one evidence artifact, or one short batch. Continue until one of
these is true:

- all generated nodes are complete, no ready nodes remain, no exact next action
  remains, and `final_response_allowed=true`;
- the lease minimums are met, no ready work remains, all required readbacks
  agree, and Command status is complete;
- a true hard blocker remains after safe repair, repack, or support work has
  already been attempted.

Treat `final_response_allowed=false`, ready nodes, stale checkpoints, stale route
decisions, mismatched Foundry/Atlas/Command counts, or an `exact_next_action` as
continuation pressure, not completion.

## Per-Node Evidence

Each bounded node should leave these artifacts or readbacks:

- node gate;
- candidate record;
- rollback record;
- Foundry import or owning-repo equivalent;
- implementation evidence;
- Sentinel/public-safety wording evidence where applicable;
- Promoter no-promotion or promotion-readiness evidence where applicable;
- Command/readback evidence where applicable;
- focused tests;
- verification command output;
- PR, CI, merge, sync, and branch cleanup evidence when remote lifecycle is
  available.

Keep exactly one executable mutation node active at a time. Docs-only,
readback-only, and evidence-only tasks still need the same rollback and
verification trail.

## Continuation Commands

After each node or checkpoint interval, refresh the route and event evidence:

```sh
ao-mission mission events index --out tmp/mission-event-index.json
ao-mission mission events query-index --index tmp/mission-event-index.json --out tmp/mission-timeline-query-index.json
ao-mission mission events search --mission <mission-id> --query "route" --index tmp/mission-event-index.json --json
ao-mission mission events search --mission <mission-id> --query "checkpoint" --index tmp/mission-event-index.json --json
ao-mission mission events resume-prompt --mission <mission-id> --out tmp/<mission-id>-compaction-resume-prompt.json --json
ao-mission mission dashboard --mission <mission-id> --compact --out tmp/<mission-id>-dashboard.json
ao-mission command status --mission <mission-id> --json
ao-mission final rollup --mission <mission-id>
ao-mission final reconcile --mission <mission-id>
```

If final rollup emits Feature Depth Recommendations, convert them into the next
Atlas workgraph wave with at least 10 concrete tasks by default and at least 30
tasks for a doubled 2-3 hour request.

## Prompt Template

Use this shape when handing a doubled wave back to Atlas:

```text
You are AO Atlas, continuing the AO Mission doubled long-run wave.

Do not ask the operator for permission. Do not stop after one repo, one PR, one
CI pass, one Foundry import, one route decision, one evidence artifact, one
rollup, or one short batch.

Target 2-3 hours of useful work. Complete at least 30 bounded nodes unless all
generated nodes are complete or a true hard blocker remains after safe repair,
repack, and support work.

Load the latest AO Mission lease, checkpoint bundle, event index, Atlas
workgraph, Foundry rollup, Promoter readback, and Command compact status.

Route ownership:
- Mission owns lease, checkpoint, return gate, and final rollup.
- Atlas owns workgraph/context-heavy sequencing and final-response refusal.
- Foundry owns exactly one bounded implementation node at a time.
- Command owns compact readback.
- Blueprint is used only for missing requirements, missing authorization, or an
  unclear safety boundary.

For each node, produce node gate, candidate, rollback, implementation,
Sentinel/public-safety, Promoter/no-promotion or promotion-readiness,
Command/readback, tests, verification output, PR/CI/merge evidence when
available, and branch cleanup evidence.

Final response is denied while ready nodes, exact next actions, stale
checkpoints, stale route decisions, mismatched readbacks, or unmet lease
minimums remain.

Return only with completed node count, merged PRs, evidence roots, verification,
clean repo status, Command readback, Foundry rollup, Atlas workgraph status, and
at least 10 next Feature Depth Recommendations.
```
