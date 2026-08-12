# Mission Finalizer Explicit Repository Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the checkout-free Mission finalizer publish to its explicit GitHub repository.

**Architecture:** Keep the existing imported-artifact validation and protected publisher unchanged. Add the repository selector directly to the existing GitHub CLI release command and lock that requirement with the existing workflow contract test.

**Tech Stack:** GitHub Actions YAML, GitHub CLI, Go tests.

## Global Constraints

- Preserve exact producer, source, version, tag, manifest, archive, environment, and permission bindings.
- Do not add a checkout step, dependency, helper, implicit repository environment, or authority change.
- Do not dispatch another live release until merged-head qualification passes.

---

### Task 1: Bind release creation to the explicit repository

**Files:**
- Modify: `internal/mission/release_rehearsal_workflow_test.go`
- Modify: `.github/workflows/release-finalize.yml`

**Interfaces:**
- Consumes: GitHub Actions' built-in `GITHUB_REPOSITORY` value in `owner/name` form.
- Produces: `gh release create` invocation containing `--repo "$GITHUB_REPOSITORY"`.

- [ ] **Step 1: Write the failing regression assertion**

Add this assertion to `TestReleaseFinalizationImportsExactRehearsalArtifacts`:

```go
wantPublisher := `gh release create "$TAG" --repo "$GITHUB_REPOSITORY" --target "$SOURCE_SHA" --title "AO Mission $VERSION" "${archives[@]}"`
if !strings.Contains(workflow, wantPublisher) {
	t.Fatalf("release finalization publisher is not bound to the explicit repository: want %q", wantPublisher)
}
```

- [ ] **Step 2: Verify the regression fails for the production defect**

Run: `go test ./internal/mission -run TestReleaseFinalizationImportsExactRehearsalArtifacts -count=1`

Expected: FAIL with `release finalization publisher is not bound to the explicit repository`.

- [ ] **Step 3: Implement the one-line workflow fix**

Change the publisher command to:

```bash
gh release create "$TAG" --repo "$GITHUB_REPOSITORY" --target "$SOURCE_SHA" --title "AO Mission $VERSION" "${archives[@]}"
```

- [ ] **Step 4: Verify focused and full local gates**

Run:

```bash
go test ./internal/mission -run TestReleaseFinalizationImportsExactRehearsalArtifacts -count=1
go test ./internal/mission -count=1
gofmt -d cmd internal
go test ./... -count=1
go vet ./...
go build ./cmd/ao-mission
python3 scripts/test_public_safety_scan.py
./scripts/production-readiness.sh
python3 ../ao-architecture/scripts/verify_agent_instruction_layout.py --workspace-root .. --repository ao-mission
git diff --check
```

Expected: every command exits zero, formatting produces no diff, production readiness reports `100/100`, and instruction layout reports zero conflicts.

- [ ] **Step 5: Commit the implementation**

```bash
git add .github/workflows/release-finalize.yml internal/mission/release_rehearsal_workflow_test.go docs/superpowers/plans/2026-08-12-mission-finalizer-explicit-repository.md
git commit -m "fix: bind Mission release publisher repository"
```
