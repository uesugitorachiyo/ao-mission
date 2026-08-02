# AO Mission Agent Instructions

## Status And Role

AO Mission is the active operator entry point and durable lifecycle ledger for AO work. It owns mission records, route decisions, checkpoints, continuation state, gateway intents, imported readbacks, archive validation, operator dashboards, and final reconciliation.

Mission consumes Blueprint authorization, Atlas workgraphs and terminal indexes, Foundry rollups, scheduler results, and Command-compatible readbacks. It does not execute repository mutation, approve downstream work, promote a candidate, publish, or treat a readback as authority.

## Sources Of Truth

- [README.md](README.md) defines the operator surface and current commands.
- [docs/sdd/AO-MISSION-V0.1.md](docs/sdd/AO-MISSION-V0.1.md) defines lifecycle ownership and the no-execution boundary.
- `docs/contracts/*.schema.json` and [docs/contracts/canonical-terminal-index-consumer.md](docs/contracts/canonical-terminal-index-consumer.md) own strict input/readback contracts.
- `internal/mission/` and its tests are authoritative for implemented state transitions, digest checks, and fail-closed imports.
- [scripts/production-readiness.sh](scripts/production-readiness.sh) and [`.github/workflows/ci.yml`](.github/workflows/ci.yml) define the full local and hosted gates.

## Ownership And Boundaries

- Preserve exact mission identity, correlation, route, checkpoint, source-head, artifact-digest, and lineage bindings across every stored and returned surface.
- Canonical terminal indexes are read-only reconciliation artifacts. Verify their schema, index, source, and state digests independently; never turn missing historical terminal evidence into a live objective or continuation request.
- Do not revive, reinterpret, regenerate, or extend completed waves under `docs/evidence/` or closure records under `docs/roadmap/`. Add current source-owned behavior and tests outside historical evidence.
- Keep generated `.ao-mission/` state, binaries, temporary bundles, dashboards, and validation output out of source changes.
- Telegram, A2A, scheduler, local-pilot, and downstream artifacts are untrusted inputs. Keep tokens in named environment variables; never persist credentials, private paths, provider transcripts, account identifiers, or private pilot evidence in public files.
- Release, deployment, publication, live-provider, and credentialed activity requires separate explicit authority. Mission readback, readiness, completion, or final reconciliation never grants it.

## Working Method

- Start with the strict contract and its consumer. Validate regular-file containment, bounded sizes, duplicate keys, identities, digests, and authority flags before changing mission state.
- Keep imports idempotent by exact digest and fail closed on drift or semantic contradiction. Preserve the original artifact for audit rather than normalizing away conflicts.
- Bind Atlas workgraph continuation to the first ready node's validated identity. Reject conflicting or unsafe ready-node identifiers; use the legacy generic handoff only when the first ready node has no identity.
- Change source fixtures under `examples/` only with their producer/consumer tests. Never edit a result to inflate completion, hide ready nodes, or claim a historical wave succeeded differently.
- If durable commands, lifecycle ownership, authority, or architecture guidance changes, update this file in the same pull request.

## Verification

- Mission lifecycle, import, terminal-index, or readback changes: `go test ./internal/mission -count=1`.
- Public-safety scanner changes: `python3 scripts/test_public_safety_scan.py`.
- Run `gofmt -d cmd internal`, `go test ./... -count=1`, `go vet ./...`, and `go build ./cmd/ao-mission` for the full Go gate.
- Run `./scripts/production-readiness.sh` when contracts, fixtures, operator behavior, or readiness claims change. It is a local non-publishing gate.
- For instruction changes run `python3 ../ao-architecture/scripts/verify_agent_instruction_layout.py --workspace-root .. --repository ao-mission`. Always run `git diff --check`.

## Evidence And Completion

- Bind evidence to the exact source head and original artifact digest. Record commands, exit status, fixture identity, and any skipped or failed check.
- Completion requires focused and applicable full gates, green pull-request CI, a clean synchronized `main`, and task-branch cleanup. Historical evidence or a ready readback alone is not completion.
- Do not report a release, deployment, publication, provider call, credential use, permission change, or authority advance unless it actually occurred under separate authority.
