# Content-Addressed Evidence Retention Design

## Goal

Prevent an AO Mission artifact manifest from becoming unverifiable when an
accepted source artifact is later replaced, moved, or deleted.

## Ownership

AO Mission owns the repair because it accepts artifact imports, stores their
references, and emits the final artifact manifest. AO Atlas continues to own
strict terminal-index construction and verification; its contract does not
need to change.

## Design

Before an import updates Mission state, Mission writes the already validated
artifact bytes to a content-addressed object beneath the configured Mission
home:

```text
<AO_MISSION_HOME>/artifacts/sha256/<64-lowercase-hex-digest>
```

Creation is atomic and fail closed. An existing object is accepted only when
it is a regular file containing the exact expected bytes. Symlinks,
directories, digest collisions, partial writes, and replacement attempts are
rejected.

Mission retains accepted bytes at import time before committing the artifact
reference. The record keeps `ref` as the caller's provenance locator and adds
`content_ref` for the immutable object. An orphaned object after a failed
record transaction is harmless and may be reused by an exact later import.

## Manifest Behavior

New manifest construction emits `ao.mission.artifact-manifest.v0.2`. Each
entry binds `ref`, the byte-exact `digest`, and a contained relative
`content_ref`. When writing a manifest, Mission copies the import-time object
into an `artifacts/sha256/` directory beside the manifest and atomically writes
the manifest only after every object is present and verified. Validation reads
only `content_ref` for v0.2. Legacy v0.1 remains readable with its historical
line-ending behavior.

The repair-manifest command may upgrade v0.1 only when the currently available
source bytes still match the declared digest. It adds retained content without
changing `ref` or `digest`. A missing or mismatched historical source fails
closed and the original manifest remains untouched.

## Security Boundaries

- Evidence remains under the operator-selected `AO_MISSION_HOME`.
- No network, provider, release, deployment, publication, or authority change
  is introduced.
- Imported content passes the existing public-safety, duplicate-key, JSON,
  identity, and authority checks before retention.
- The object path is derived only from a validated SHA-256 digest.
- Existing objects are opened and checked as regular files without following
  a replacement symlink.
- V0.2 hashes byte-exact content. V0.1 retains its legacy CRLF normalization.

## Verification

Tests cover exact retention, source replacement and deletion after import,
deduplication, failed-import non-publication, symlink rejection, existing
object mismatch, concurrent exact imports, legacy references, and manifest
tampering. Full Mission tests, race tests, vet, build, readiness, and hosted CI
remain required before merge.

## Recertification

After merge, a fresh bounded Mission objective will import current exact-head
evidence into a new campaign directory under
`/Users/torachiyouesugi/Documents/canary-test`. It will reuse only evidence
that independently passes identity, freshness, and digest checks, rerun stale
gates, perform one restart and one compaction/replay, generate canonical
terminal views, and independently verify its complete manifest. The previous
campaign remains immutable and its historical digest contradiction remains
visible.
