# AO Mission

ao-mission is the governed AO mission entry point, router, continuation ledger, communication gateway, governance snapshot producer, and scheduler adapter. It records and routes mission work through existing AO gates without expanding authority.

## v0.1 Boundaries

AO Mission does not approve policy, execute provider calls, mutate repositories directly, publish releases, deploy, upload, tag releases, update dependencies, widen auth or config, or grant mutation authority from Telegram or A2A. AO Command and ao2-control-plane remain read-only. codex-cron is only a scheduler adapter.

Denied boundaries remain denied: unrestricted self-modification, hidden instruction mutation, policy-changing autonomy, forbidden surface expansion, credential use, provider calls, release/deploy/publish/upload/tag authority, dependency update authority, direct-main mutation, concurrent mutation, broad public claims, unrestricted RSI, and broad_RSI.

## Commands

```sh
ao-mission init
ao-mission start "<objective>"
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
```

By default state is stored under `.ao-mission/`. Use `AO_MISSION_HOME` to choose another state root.
