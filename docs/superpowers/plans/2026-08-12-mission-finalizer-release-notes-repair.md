# Mission Finalizer Release Notes Repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish digest-bound release notes for new Mission releases and safely repair the empty body of exact release `v0.1.4`.

**Architecture:** Extend the existing imported-release validator and protected publisher only. The immutable promotion plan remains authoritative; a distinct input and confirmation constrain the one authorized existing-release metadata repair.

**Tech Stack:** GitHub Actions YAML, Bash, Python standard library, GitHub CLI, Go workflow-contract tests.

## Global Constraints

- Preserve source `cee287597024b5a1e990c6e272518236bc9e32fa`, release ID `369514351`, tag `v0.1.4`, and the exact three public assets.
- Only an empty release body may be repaired, using the sealed notes whose SHA-256 is `9a84817e6d75b197c72a3219f7f851cb31935da679688bb14e8560eea0bf1022`.
- Keep the protected environment and least permissions unchanged.

---

### Task 1: Lock the release-notes contract

**Files:**
- Modify: `internal/mission/release_rehearsal_workflow_test.go`

- [ ] Add assertions for exact notes digest validation and retention.
- [ ] Add assertions for `--notes-file` on new release creation.
- [ ] Add assertions for the explicit repair input, distinct confirmation, tag/source/title/body/release-state checks, exact asset digest checks, and metadata-only edit.
- [ ] Run `go test ./internal/mission -run TestReleaseFinalizationImportsExactRehearsalArtifacts -count=1` and observe failure.

### Task 2: Implement the minimal protected repair

**Files:**
- Modify: `.github/workflows/release-finalize.yml`

- [ ] Validate and retain the exact sealed `release-notes.md`.
- [ ] Publish new releases with `--notes-file`.
- [ ] Add `repair_empty_release_notes` and its distinct confirmation.
- [ ] In repair mode, independently verify the existing public release and exact assets, then edit only its notes.
- [ ] Run the focused test and all repository-required local gates.

### Task 3: Publish, review, merge, and qualify

- [ ] Commit, push, and open a draft PR.
- [ ] Obtain independent review and all hosted checks.
- [ ] Merge, synchronize `main`, and remove the task branch.
- [ ] Run fresh merged-head rehearsal and finalizer dry run.
- [ ] Dispatch the exact protected metadata repair, wait for human approval, and independently reverify Mission Gate 3.
