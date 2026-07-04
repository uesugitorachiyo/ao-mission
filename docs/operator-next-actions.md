# AO Mission Operator Next Actions

AO Mission should leave the operator with a concrete command, not only a
generated artifact path. Use this short sequence when checking or continuing a
mission.

## Start And Inspect

```sh
ao-mission start "<objective>"
ao-mission mission list --json
ao-mission next --mission <mission-id>
ao-mission status --mission <mission-id>
```

## Continue Locally

```sh
ao-mission continue --mission <mission-id> --until-done --max-iterations 10
ao-mission mission history --mission <mission-id>
ao-mission mission compact --mission <mission-id> --keep-route-history 25 --keep-steps 25
ao-mission mission compact --mission <mission-id> --keep-route-history 25 --keep-steps 25 --timeline
ao-mission mission archive --mission <mission-id> --out tmp/<mission-id>-archive.json
ao-mission mission validate-archive --path tmp/<mission-id>-archive.json --out tmp/<mission-id>-archive-validation.json
ao-mission artifacts manifest --mission <mission-id> --out tmp/<mission-id>-artifact-manifest.json
ao-mission final rollup --mission <mission-id>
```

## Gateway Fixture Checks

```sh
ao-mission telegram replay --matrix examples/valid/telegram-command-matrix.json --config examples/valid/telegram-config.json
ao-mission telegram replay-updates --fixture examples/valid/telegram-update-replay.json --config examples/valid/telegram-config.json
ao-mission telegram webhook-replay --fixture examples/valid/telegram-webhook-replay.json --config examples/valid/telegram-config.json
ao-mission telegram role-matrix --config examples/valid/telegram-config.json --out tmp/telegram-role-matrix.json
ao-mission a2a serve --http --once
ao-mission a2a replay --fixture examples/valid/a2a-http-integration.json
ao-mission a2a lifecycle --fixture examples/valid/a2a-task-lifecycle-edges.json
ao-mission a2a compatibility --agent-card examples/valid/a2a-agent-card.json --http examples/valid/a2a-http-integration.json --lifecycle examples/valid/a2a-task-lifecycle-artifacts.json --out tmp/a2a-compatibility.json
ao-mission a2a streaming-denial --agent-card examples/invalid/a2a-agent-card-streaming.json --out tmp/a2a-streaming-denial.json
ao-mission gateway replay-suite --telegram-config examples/valid/telegram-config.json --telegram-webhook examples/valid/telegram-webhook-replay.json --telegram-updates examples/valid/telegram-update-replay.json --a2a-http examples/valid/a2a-http-integration.json --a2a-lifecycle examples/valid/a2a-task-lifecycle-artifacts.json --out tmp/gateway-replay-suite.json
ao-mission gateway readiness-rollup --mission <mission-id> --suite tmp/gateway-replay-suite.json --a2a-compatibility tmp/a2a-compatibility.json --archive-validation tmp/<mission-id>-archive-validation.json --snapshot-diff <snapshot-diff.json> --out tmp/gateway-readiness-rollup.json
ao-mission schedule replay --fixture examples/valid/scheduler-readback-replay.json
ao-mission schedule alerts --fixture examples/valid/scheduler-readback-replay.json
ao-mission schedule recover --mission <mission-id> --fixture examples/valid/scheduler-readback-replay.json
ao-mission import scheduler-recovery-readback --mission <mission-id> --path examples/valid/scheduler-recovery-readback.json
ao-mission import ledger-compaction-readback --mission <mission-id> --path examples/valid/ledger-compaction-readback.json
ao-mission validate contract --path examples/valid/timeline-compaction-readback.json
ao-mission mission compact --mission <mission-id> --keep-route-history 25 --keep-steps 25 --dry-run
ao-mission artifacts repair-manifest --path <artifact-manifest.json> --out <artifact-manifest.repaired.json>
ao-mission governance diff --before <snapshot-before.json> --after <snapshot-after.json>
ao-mission mission archive --mission <mission-id> --out tmp/<mission-id>-archive.json
```

Telegram and A2A fixture checks record intent/readback only. They do not grant
execution authority, approval authority, repository mutation, provider calls,
credential use, release/deploy/publish/upload/tag authority, dependency update
authority, direct-main mutation, concurrent mutation, hidden instruction
mutation, unrestricted self-modification, unrestricted RSI, or broad_RSI.

## Route Handoff

If `ao-mission next --mission <mission-id>` returns `ao-blueprint`, Move to AO
Blueprint and create or inspect the Blueprint pack.

If it returns `ao-atlas`, Move to AO Atlas and compile/import the authorized
Mission context into an Atlas workgraph.

If it returns `ao-foundry`, Move to AO Foundry and consume only the first safe
Atlas node through Foundry gates.

If it returns `complete`, read the final rollup and recommended next tasks. Do
not treat a generated handoff file alone as completion.
