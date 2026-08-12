# Mission Finalizer Nested Assets Design

## Problem

The finalizer validates and retains each release candidate in a separate
`validated/ao-mission-release-candidate-*/` directory. Its publisher searches
only the `validated/` top level, finds zero archives, and exits before release
creation.

## Design

Keep the validated evidence layout unchanged. Search recursively beneath
`validated/` for regular `.tar.gz` and `.zip` files, sort the paths, and retain
the exact-three assertion before calling `gh release create`. The imported
validator remains responsible for candidate identity and digest validation.

Add a workflow regression assertion requiring recursive discovery and
forbidding the broken top-level-only search. No helper, dependency, artifact
flattening, release-contract change, or authority expansion is needed.

## Verification

Run the focused Mission workflow tests, full repository gates, a dry-run
finalizer against fresh exact-head rehearsal evidence, and hosted CI before a
new live finalizer attempt.
