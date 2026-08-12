# Mission Finalizer Explicit Repository Design

## Problem

The finalizer intentionally publishes sealed imported artifacts without
checking out the repository. `gh release create` therefore cannot infer a
repository from Git metadata and exits before creating the tag or release.

## Design

Keep the publisher checkout-free and pass `--repo "$GITHUB_REPOSITORY"` to the
existing `gh release create` command. Retain the exact source, tag, title,
validated archive inventory, protected environment, and permissions unchanged.

Add one workflow regression assertion requiring the explicit repository flag.
Do not add a checkout step, implicit repository environment, helper, dependency,
or release-authority change.

## Verification

Watch the focused workflow test fail before the workflow edit, then pass after
the one-line fix. Run the applicable full local gates, hosted CI, fresh
merged-head rehearsal and finalizer dry run, then dispatch a newly bound live
finalizer through the existing protected environment.
