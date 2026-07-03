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

Every command returns `mutation_authority=false`. Admin commands are still requests for AO Mission continuation or readback, not direct execution. Unsupported slash commands and non-slash command strings are rejected by the invalid fixtures under `examples/invalid/`; non-admin `/continue` and `/approve` remain denied intents.

Use `ao-mission telegram replay-updates --fixture examples/valid/telegram-update-replay.json --config examples/valid/telegram-config.json` to replay Telegram-style update payloads. The update replay proves allowlisted chat IDs, denied admin intents, invalid non-slash text, and `mutation_authority=false` without contacting Telegram or storing a token value.

## A2A

The local A2A gateway exposes an Agent Card and JSON-RPC style task readbacks for local interoperability. It follows the A2A idea of agent-to-agent communication without exposing AO Mission internals as unrestricted tools. The Agent Card includes local fixture metadata, `streaming=false`, `push_notifications=false`, and `mutation_authority=false`. External agents may request `mission.start`, `mission.status`, `mission.next`, `mission.continue`, `mission.pause`, `mission.resume`, `mission.cancel`, `mission.artifacts`, and `mission.governance_snapshot`.

Parameter validation is intentionally strict:

- `mission.start` requires an objective.
- Status, next, continue, pause, resume, cancel, artifacts, and governance snapshot requests require a mission ID.
- Unknown methods produce an invalid readback.

The invalid JSON-RPC fixtures under `examples/invalid/` cover missing objective, missing mission ID, and unknown method requests.

Every A2A response remains intent/readback only and reports `mutation_authority=false`.

## Operator Rule

Gateway output is never completion evidence by itself. AO Mission must still route authorized work through Blueprint, Atlas, Foundry, Forge/AO2, Covenant, Sentinel, Promoter, Command, CI, rollback, eval/regression, and Architecture wording gates as applicable.
