# Content-Addressed Evidence Retention Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Preserve every newly imported AO Mission artifact and emit a portable, immutable v0.2 manifest without rewriting historical digests.

**Architecture:** Capture validated import bytes beneath `AO_MISSION_HOME`, bind the retained locator without replacing source provenance, and materialize portable content-addressed objects beside new v0.2 manifests. Keep v0.1 readable and make repair fail closed when declared historical bytes are unavailable.

**Tech Stack:** Go standard library, AO Mission store and import contracts, Go tests.

## Global Constraints

- Historical evidence is immutable and remains under `/Users/torachiyouesugi/Documents/canary-test`.
- New retained evidence is rooted beneath the configured `AO_MISSION_HOME`.
- `AO_MISSION_HOME` is a trusted operator-owned root; hostile same-user concurrent replacement of that root is outside scope.
- No network, provider, credential, release, deployment, publication, or authority expansion.
- Existing `ao.mission.artifact-manifest.v0.1` and artifact references without `content_ref` remain valid.
- Use test-first implementation and fail closed on non-regular objects, symlinks or reparse points present at operation time, digest mismatch, or partial writes.

---

### Task 1: Content-Addressed Object Store

**Files:**
- Create: `internal/mission/artifact_store.go`
- Create: `internal/mission/artifact_store_test.go`

**Interfaces:**
- Produces: `Store.retainArtifact(body []byte) (string, string, error)` returning the immutable object path and byte-exact `sha256:` digest.
- Consumes: `Store.Root` and the existing `digestBytes` helper.

- [ ] Write failing tests for first capture, exact deduplication, mismatched existing object, symlink rejection, and concurrent exact capture.
- [ ] Run `go test ./internal/mission -run 'TestRetainArtifact' -count=1` and confirm the tests fail because `retainArtifact` is absent.
- [ ] Implement bounded atomic creation at `artifacts/sha256/<digest>` with regular-file and exact-byte verification.
- [ ] Run the focused tests and confirm they pass.

### Task 2: Import And Manifest Binding

**Files:**
- Modify: `internal/mission/types.go`
- Modify: `internal/mission/imports.go`
- Modify: `internal/mission/atlas_workgraph_next_action_test.go`
- Modify: `internal/mission/mission_test.go`
- Modify: `internal/mission/validate.go`
- Modify: `internal/mission/readback.go`
- Modify: `internal/mission/cli.go`
- Create: `docs/contracts/artifact-manifest-v0.2.schema.json`

**Interfaces:**
- Consumes: `Store.retainArtifact(body)` from Task 1.
- Produces: optional `ArtifactRef.ContentRef`; new imports preserve `Ref` and bind the import-time object. New `--out` manifests materialize relative `content_ref` objects and emit v0.2.

- [ ] Write failing import tests proving source replacement and deletion cannot invalidate a newly generated manifest.
- [ ] Write negative tests proving validation failures do not add Mission references and retained-object tampering is rejected.
- [ ] Run the focused import and manifest tests and confirm the new assertions fail.
- [ ] Extend `ArtifactRef`, contract property typing, and import construction with `content_ref` while preserving `ref`.
- [ ] Add byte-exact v0.2 manifest generation and validation while retaining v0.1 validation compatibility.
- [ ] Change repair to preserve declared digests and fail closed when the original bytes no longer match.
- [ ] Preserve exact reimport idempotency and correlated-import behavior.
- [ ] Run focused import, manifest, and contract tests and confirm they pass.

### Task 3: AO Atlas V0.2 Consumer Compatibility

**Files:**
- Modify: `../ao-atlas/internal/atlas/mission_import.go`
- Modify: `../ao-atlas/internal/atlas/atlas_test.go`
- Modify: `../ao-atlas/internal/atlas/canonical_terminal_index_test.go`

**Interfaces:**
- Consumes: AO Mission manifest v0.2 with contained relative `content_ref` values.
- Produces: strict v0.2 import verification while preserving v0.1 fallback.

- [ ] Write failing tests for valid v0.2, missing or altered retained content, traversal, symlink, oversized content, and v0.1 compatibility.
- [ ] Run focused tests and confirm the v0.2 case fails before implementation.
- [ ] Implement strict schema dispatch and contained byte-exact retained-content verification.
- [ ] Add one Atlas-to-Mission canonical terminal-index round trip using CAS paths.
- [ ] Run focused Atlas tests and confirm they pass.

### Task 4: Operator Contract And Full Verification

**Files:**
- Modify: `README.md`
- Modify: `AGENTS.md`

**Interfaces:**
- Documents: content-addressed retention location, original-locator provenance, legacy compatibility, and no-authority behavior.

- [ ] Document the durable retention behavior and operational boundary.
- [ ] Run `gofmt -w internal/mission/artifact_store.go internal/mission/artifact_store_test.go internal/mission/types.go internal/mission/imports.go internal/mission/atlas_workgraph_next_action_test.go internal/mission/mission_test.go internal/mission/validate.go`.
- [ ] Run `go test ./internal/mission -count=1`.
- [ ] Run `go test -race ./internal/mission -count=1`.
- [ ] Run `go test ./... -count=1`, `go vet ./...`, and `go build ./cmd/ao-mission`.
- [ ] Run `./scripts/production-readiness.sh` and `git diff --check`.
- [ ] Commit the bounded Mission change, open one PR, wait for hosted CI, merge only when green, synchronize `main`, and remove the task branch/worktree.

### Task 5: Bounded Exact-Head Recertification

**Files:**
- Create only beneath: `/Users/torachiyouesugi/Documents/canary-test/ao-stack-production-recertification-<UTC>/`

**Interfaces:**
- Consumes: merged AO Mission head and preserved prior campaign evidence.
- Produces: fresh Mission identity, 8-12 node Atlas workgraph, four canonical terminal views, verified artifact manifest, completion audit, and one non-publishing readiness decision.

- [ ] Inventory exact maintained repository heads and classify prior evidence by identity, freshness, and digest validity.
- [ ] Execute only stale or missing gates, including required native and physical platform checks.
- [ ] Exercise one checkpoint/restart and one compaction/replay.
- [ ] Build and independently verify canonical terminal reconciliation and the complete content-addressed manifest.
- [ ] Emit `AO_STACK_PRODUCTION_READY_FOR_SEPARATE_RELEASE_AUTHORIZATION` only if every gate passes; otherwise emit `AO_STACK_PRODUCTION_NOT_READY` with one exact next action.
