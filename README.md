# AO Mission

ao-mission is the governed AO mission entry point, router, continuation ledger, communication gateway, governance snapshot producer, and scheduler adapter. It records and routes mission work through existing AO gates without expanding authority.

## v0.1 Boundaries

AO Mission does not approve policy, execute provider calls, mutate repositories directly, publish releases, deploy, upload, tag releases, update dependencies, widen auth or config, or grant mutation authority from Telegram or A2A. AO Command and ao2-control-plane remain read-only. codex-cron is only a scheduler adapter.

Denied boundaries remain denied: unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, forbidden surface expansion, credential use, provider calls, release/deploy/publish/upload/tag authority, dependency update authority, direct-main mutation, concurrent mutation, broad public claims, unrestricted RSI, and broad_RSI.

## Commands

```sh
ao-mission init
ao-mission start "<objective>"
ao-mission mission list [--status <status>] [--route <route>] [--json]
ao-mission mission inspect --mission <id> [--json]
ao-mission continue --mission <id> [--until-done] [--max-iterations N]
ao-mission status --mission <id> [--json]
ao-mission next --mission <id> [--json]
ao-mission pause --mission <id>
ao-mission resume --mission <id>
ao-mission stop --mission <id>
ao-mission schedule --mission <id> --every <duration> --event-loop
ao-mission daemon install|status|uninstall
ao-mission telegram serve
ao-mission a2a serve
ao-mission governance snapshot --mission <id>
ao-mission artifacts --mission <id>
ao-mission artifacts manifest --mission <id>
ao-mission command status --mission <id> [--json]
ao-mission validate contract --path <json>
ao-mission import blueprint-authorization --mission <id> --path <json>
ao-mission import atlas-workgraph --mission <id> --path <json>
ao-mission import foundry-run-link --mission <id> --path <json>
ao-mission import foundry-final-rollup --mission <id> --path <json>
ao-mission import scheduler-readback --mission <id> --path <json>
ao-mission final rollup --mission <id>
```

By default state is stored under `.ao-mission/`. Use `AO_MISSION_HOME` to choose another state root.
Every command also accepts `--home <dir>` before the command name for explicit local state routing.

## Gateway References

The messaging surface follows the same split used by Hermes-style gateways: CLI and messaging platforms are separate entry points into one mission ledger, and messaging commands create intents/readbacks instead of direct mutation. The A2A local gateway exposes an Agent Card and task-style readbacks for local interoperability while preserving `mutation_authority=false`.

Telegram is disabled by default. A config file may name the environment variable that contains the real token and a chat allowlist, but ao-mission never prints or persists the token value.

See [Gateway Readback Runbook](docs/gateway-readback-runbook.md) for the fixture-backed command matrix, denied command examples, A2A JSON-RPC parameter checks, and intent-only authority boundary.

## Readback Surfaces

`ao-mission continue` persists `ao.mission.event-loop-decision.v0.1` after each continuation step so the zero-wait event loop has durable no-authority readback. `ao-mission import atlas-workgraph` records node counts from Atlas workgraphs, `ao-mission import scheduler-readback` records codex-cron wakeup evidence without granting execution authority and rejects any scheduler readback that claims `executes_work=true`, and `ao-mission import foundry-final-rollup` marks the mission done only when completed and total node counts agree. `ao-mission command status` emits a read-only AO Command compatible status packet. `ao-mission artifacts manifest` emits a digest-bound local manifest over mission artifacts without granting execution or approval authority.
