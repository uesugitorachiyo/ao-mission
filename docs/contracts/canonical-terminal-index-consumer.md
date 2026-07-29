# Canonical Terminal Index Consumer

AO Mission consumes `ao.canonical-terminal-index.v1` as a read-only
reconciliation artifact. It independently verifies the index digest, schema
digest, every source-artifact digest, mission identity, ordered lineage,
regular-file and path containment rules, counts, lease classification, conflict
codes, safety boundaries, readiness, return gate, and exact next action.

The importer accepts at most 128 artifacts, 1 MiB per artifact, and 16 MiB in
total. It rejects malformed JSON, duplicate keys, symlinks, traversal,
out-of-root paths, identity drift, digest changes, non-monotonic lineage,
semantic contradictions, and any execution, approval, mutation, provider,
publication, release, deployment, or authority-advance flag.

Imports are durable and idempotent by index digest. A different second digest is
rejected before the state file changes. Inspect, checkpoint, event-index, and
Command-compatible views all read the same persisted reconciliation record.
The import never schedules work or changes an AO Mission record.

```sh
ao-mission terminal-index import \
  --root /path/to/evidence \
  --index /path/to/evidence/canonical-terminal-index.json \
  --state /path/to/read-only-import-state.json

ao-mission terminal-index inspect --state /path/to/read-only-import-state.json
ao-mission terminal-index checkpoint --state /path/to/read-only-import-state.json
ao-mission terminal-index event-index --state /path/to/read-only-import-state.json
ao-mission terminal-index command-readback --state /path/to/read-only-import-state.json
```

Historical Mission evidence can be indexed additively with `terminal-index
historical`. Missing terminal evidence remains `no_canonical_terminal`; it is
never converted into a live objective or automatic continuation request.
