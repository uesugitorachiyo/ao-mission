# AO Mission Gateway Readback Runbook

AO Mission uses Hermes-style gateway separation: CLI, Telegram, and A2A are entry points into one durable mission ledger, but gateway messages create operator intents and readbacks only. They do not execute mutation, approve policy, call providers, publish releases, use credentials, or widen repository authority.

## Telegram

Telegram is disabled by default. A config file names the environment variable that holds the real bot credential and an allowlist of chat IDs. AO Mission never prints or persists the credential value. Public fixtures use fake chat IDs and redacted credential names only.

Supported Telegram commands are represented in `examples/valid/telegram-command-matrix.json`:

- `/status`
- `/next`
- `/continue`
- `/pause`
- `/resume`
- `/stop`
- `/approve`
- `/deny`
- `/where`
- `/help`

Every command returns `mutation_authority=false`. Admin commands are still requests for AO Mission continuation or readback, not direct execution.

## A2A

The local A2A gateway exposes an Agent Card and JSON-RPC style task readbacks for local interoperability. It follows the A2A idea of agent-to-agent communication without exposing AO Mission internals as unrestricted tools. External agents may request `mission.start`, `mission.status`, `mission.next`, `mission.continue`, `mission.pause`, `mission.resume`, `mission.cancel`, `mission.artifacts`, and `mission.governance_snapshot`.

Parameter validation is intentionally strict:

- `mission.start` requires an objective.
- Status, next, continue, pause, resume, cancel, artifacts, and governance snapshot requests require a mission ID.
- Unknown methods produce an invalid readback.

Every A2A response remains intent/readback only and reports `mutation_authority=false`.

## Operator Rule

Gateway output is never completion evidence by itself. AO Mission must still route authorized work through Blueprint, Atlas, Foundry, Forge/AO2, Covenant, Sentinel, Promoter, Command, CI, rollback, eval/regression, and Architecture wording gates as applicable.
