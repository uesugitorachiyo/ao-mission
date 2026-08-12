# Mission Finalizer Release Notes Repair Design

## Problem

Mission `v0.1.4` was created with the exact source and three sealed archives,
but its public body is empty. The finalizer reads the expected release-notes
digest from the promotion plan without digest-validating, retaining, or passing
the corresponding `release-notes.md` to GitHub CLI.

## Design

During import validation, require exactly one `release-notes.md`, verify its
SHA-256 against `release_notes_sha256` in the immutable promotion plan, and
copy it into the validated import. New releases use that retained file through
`gh release create --notes-file`.

Add explicit boolean input `repair_empty_release_notes`, default false. A live
repair requires its own exact `repair-empty-ao-mission-release-notes-...`
confirmation and the existing protected `ao-mission-release` environment.
Before editing, fail closed unless the public release is non-draft and
non-prerelease, has the exact title, has an empty body, its tag resolves to the
exact source SHA, and its exact three public archive names and SHA-256 values
match the immutable plan. Then PATCH only the `body` field of the fixed numeric
release ID through the GitHub REST API and verify the exact post-state.

Never delete or recreate the release, move or rewrite its tag, upload or delete
assets, or change permissions, reviewers, credentials, or environment policy.

## Verification

Add workflow-contract regressions for notes digest validation, retained notes,
explicit new-release notes, distinct repair confirmation, strict public-state
checks, exact asset digest checks, and metadata-only edit. Observe focused red
before implementation and green afterward; run all local gates, hosted CI,
merged-head rehearsal and dry-run finalizer; then dispatch the exact protected
metadata repair and independently reverify the public release.
