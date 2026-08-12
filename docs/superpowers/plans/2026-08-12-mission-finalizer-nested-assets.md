# Mission Finalizer Nested Assets Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the governed Mission finalizer publish the three already-validated nested candidate archives.

**Architecture:** Preserve the validator and artifact layout. Change only the publisher inventory search from top-level to recursive while retaining its exact-three fail-closed assertion.

**Tech Stack:** GitHub Actions YAML, Go workflow contract tests.

## Global Constraints

- Do not weaken source, manifest, plan, digest, environment-review, or exact-count checks.
- Do not create a tag or release during qualification.
- Requalify any changed source head before live publication.

---

### Task 1: Regression and Root Fix

**Files:**
- Modify: `internal/mission/release_rehearsal_workflow_test.go`
- Modify: `.github/workflows/release-finalize.yml`

**Interfaces:**
- Consumes: validated archives stored one directory below `validated/`.
- Produces: a sorted three-path Bash array passed unchanged to `gh release create`.

- [ ] **Step 1: Write the failing test**

Require `find validated -type f` and reject `find validated -maxdepth 1 -type f` in `TestReleaseFinalizationImportsExactRehearsalArtifacts`.

- [ ] **Step 2: Verify the test fails**

Run: `go test ./internal/mission -run TestReleaseFinalizationImportsExactRehearsalArtifacts -count=1`

Expected: failure reporting the missing recursive archive search.

- [ ] **Step 3: Implement the minimum fix**

Replace the publisher's `find validated -maxdepth 1 -type f` command with `find validated -type f`; retain sorting, archive suffix filters, and `[ "${#archives[@]}" -eq 3 ]`.

- [ ] **Step 4: Verify focused and full gates**

Run the focused test, `go test ./internal/mission -count=1`, prescribed Go gates, production readiness, instruction-layout verifier, and `git diff --check`.

- [ ] **Step 5: Commit, publish branch, and qualify hosted behavior**

Commit only the design, plan, test, and workflow fix; push the bounded branch, open a PR, and require green hosted checks plus a fresh exact-head rehearsal and dry-run finalizer before merge/live publication.
