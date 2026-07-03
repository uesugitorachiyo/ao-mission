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
ao-mission artifacts manifest --mission <mission-id>
ao-mission final rollup --mission <mission-id>
```

## Gateway Fixture Checks

```sh
ao-mission telegram replay --matrix examples/valid/telegram-command-matrix.json --config examples/valid/telegram-config.json
ao-mission a2a replay --fixture examples/valid/a2a-http-integration.json
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
