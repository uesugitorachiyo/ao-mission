# AO Mission Long-Run Supervisor v0.3 Evidence

Mission: `ao-mission-long-run-supervisor-v03`

Status: implementation, readback evidence, post-merge verification, CI, and
normal PR lifecycle complete.

Scope:
- AO Mission owns the long-run lease, checkpoint, return gate, route
  reconciliation, event-index, doctor, and final-rollup recommendations.
- AO Atlas owns final-response refusal while ready nodes remain and the
  final-state reconciliation packet.
- AO Foundry owns terminal rollup readback binding for completed, promoted,
  denied, and blocked statuses.
- AO Command owns compact long-run Mission status readback.
- Blueprint remains requirements/authorization only.

Denied boundaries remain denied: unrestricted self-modification, hidden
instruction mutation, policy-changing autonomy, provider calls, credential use,
release/deploy/publish/upload/tag authority, dependency updates, direct-main
mutation, concurrent mutation, unrestricted RSI, and broad_RSI.

Evidence files:
- `workgraph.json`
- `node-evidence.json`
- `feature-depth-recommendations.json`
- `atlas-workgraph-status.json`
- `foundry-rollup.json`
- `command-readback.json`
- `promoter-no-promotion.json`
- `sentinel-public-safety-scan.json`
