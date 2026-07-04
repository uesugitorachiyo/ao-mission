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
ao-mission mission history --mission <id> [--json]
ao-mission mission compact --mission <id> [--keep-route-history N] [--keep-steps N]
ao-mission continue --mission <id> [--until-done] [--max-iterations N]
ao-mission status --mission <id> [--json]
ao-mission next --mission <id> [--json]
ao-mission pause --mission <id>
ao-mission resume --mission <id>
ao-mission stop --mission <id>
ao-mission schedule --mission <id> --every <duration> --event-loop
ao-mission schedule replay --fixture <scheduler-readback-replay.json>
ao-mission schedule alerts --fixture <scheduler-readback-replay.json>
ao-mission schedule recover --mission <id> --fixture <scheduler-readback-replay.json>
ao-mission daemon install|status|uninstall
ao-mission telegram serve
ao-mission telegram replay --matrix <matrix.json> --config <telegram-config.json>
ao-mission telegram replay-updates --fixture <telegram-update-replay.json> --config <telegram-config.json>
ao-mission a2a serve
ao-mission a2a replay --fixture <a2a-http-integration.json>
ao-mission a2a lifecycle --fixture <a2a-task-lifecycle.json>
ao-mission gateway ledger --mission <id> --telegram-updates <fixture> --telegram-config <config> --a2a-http <fixture> --out <ledger.json>
ao-mission governance snapshot --mission <id>
ao-mission artifacts --mission <id>
ao-mission artifacts manifest --mission <id> [--out <manifest.json>]
ao-mission artifacts validate-manifest --path <manifest.json>
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

The messaging surface follows the same split used by Hermes-style gateways: CLI and messaging platforms are separate entry points into one mission ledger, and messaging commands create intents/readbacks instead of direct mutation. The A2A local gateway exposes an Agent Card with local protocol metadata and task-style readbacks for local interoperability while preserving `mutation_authority=false`.

Telegram is disabled by default. A config file may name the environment variable that contains the real token and a chat allowlist, but ao-mission never prints or persists the token value.

See [Gateway Readback Runbook](docs/gateway-readback-runbook.md) for the fixture-backed command matrix, denied command examples, A2A JSON-RPC parameter checks, and intent-only authority boundary. See [Operator Next Actions](docs/operator-next-actions.md) for concrete next commands after Mission emits route readback.

## Readback Surfaces

`ao-mission continue` persists `ao.mission.event-loop-decision.v0.1` after each continuation step so the zero-wait event loop has durable no-authority readback. `ao-mission next` appends `ao.mission.route-decision.v0.1` entries to the mission route history, `ao-mission mission history` exports that history for AO Command or Atlas inspection, and `ao-mission mission compact` trims retained route/step history while recording `ao.mission.ledger-compaction-readback.v0.1` evidence without repository mutation. `ao-mission import atlas-workgraph` records node counts from Atlas workgraphs, `ao-mission import scheduler-readback` records codex-cron wakeup evidence without granting execution authority, rejects any scheduler readback that claims `executes_work=true`, and classifies evidence freshness. `ao-mission schedule replay` classifies fresh, stale, and unknown scheduler readback fixtures, `ao-mission schedule alerts` turns stale or unknown scheduler readbacks into an attention-required summary, and `ao-mission schedule recover` emits an immediate `ao-mission continue` recommendation when scheduler wakeups are stale or unknown. `ao-mission gateway ledger` persists Telegram and A2A replay intents into `ao.mission.gateway-intent-ledger.v0.1` with no mutation authority, and `ao-mission a2a lifecycle` validates cancellation lifecycle fixtures as readback only. `ao-mission import foundry-final-rollup` marks the mission done only when completed and total node counts agree. `ao-mission command status` emits a read-only AO Command compatible status packet. `ao-mission artifacts manifest` emits or writes a digest-bound local manifest over mission artifacts without granting execution or approval authority, and `ao-mission artifacts validate-manifest` recomputes the manifest and referenced artifact digests so tampering fails closed.
