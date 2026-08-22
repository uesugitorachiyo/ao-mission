# AO Stack Public Release Stabilization Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish one coherent AO Stack release set referenced by AO Architecture and prove fresh public installation on Ubuntu, macOS, and Windows.

**Architecture:** Complete each repository-owned release without weakening its signer or manifest contract, then make AO Architecture the single public version-and-digest index. The terminal gate downloads only public release assets and runs the existing three-platform canary plus clean-host first-use commands.

**Tech Stack:** GitHub Actions, GitHub CLI, Git/GPG, Go, Python standard library, JSON, SHA-256, PowerShell, Bash.

**Spec:** `docs/ao-stack-public-release-stabilization-handoff-prompt.md`

## Global Constraints

- Preserve Windows rollup SHA-256 `96e3397f00155ded2b7e567ac668df889f5d5e2181c5e82c7ffad98166eb5629` and completed historical evidence unchanged.
- Mission `v0.1.5` is already public; do not recreate or rewrite it.
- Never substitute an unsigned Forge tag or an invented Command/Atlas manifest digest.
- Every public tag, archive, and checksum must bind to the exact merged source named in this plan.
- Public verification must download release assets from GitHub, not reuse workflow workspace bytes.
- Keep provider calls, deployments, promotion, compatibility activation, and RSI authority denied.
- Do not add a universal installer during this cycle; first prove the existing assembly guide.

---

### Task 1: Freeze the exact public-release ledger

**Files:**
- Read: `../ao-architecture/stack/current-release-manifest.json`
- Read: `../ao-architecture/docs/current-release.md`
- Read: `../ao-architecture/docs/stack-assembly.md`
- Read: `../ao-architecture/scripts/run_public_stack_canary.py`
- Create outside repositories: a temporary release ledger under a directory returned by `mktemp -d`

**Interfaces:**
- Consumes: GitHub `main` heads, public releases, workflow runs, and release assets.
- Produces: one reviewed ledger containing repository, source SHA, version, tag, rehearsal run, finalizer run, release URL, archive names, sizes, and SHA-256 values.

- [ ] **Step 1: Record immutable source heads and public tags**

Run:

```bash
for repo in ao2 ao2-control-plane ao-mission ao-command ao-atlas ao-forge ao-covenant; do
  gh api "repos/uesugitorachiyo/$repo/commits/main" --jq '[.sha,.commit.committer.date,.commit.message|split("\n")[0]]|@tsv'
  gh release list --repo "uesugitorachiyo/$repo" --limit 3
done
```

Expected: Mission reports public `v0.1.5`; Command remains `v0.1.2`, Atlas `v0.2.0`, and Forge `v0.1.4` until their tasks below complete.

- [ ] **Step 2: Verify the accepted merged sources**

Run:

```bash
test "$(gh api repos/uesugitorachiyo/ao-command/commits/main --jq .sha)" = "ffef6d76306e892c3e7a7f39734433d5a832006a"
test "$(gh api repos/uesugitorachiyo/ao-atlas/commits/main --jq .sha)" = "3603a2bb8af5adafcd9ff17b807ab89f32283d18"
test "$(gh api repos/uesugitorachiyo/ao-forge/commits/main --jq .sha)" = "d1723769949269dcd0589916d83769dcb7275f98"
test "$(gh api repos/uesugitorachiyo/ao-mission/commits/main --jq .sha)" = "5d4562578a4751d56910ef108b930fbb8dc91e7d"
```

Expected: every command exits zero. If a head moved, bind the new descendant head through its green `main` CI before continuing.

- [ ] **Step 3: Preserve the ledger as operator evidence**

Record command, timestamp, exit status, exact response SHA, and any missing release input. Keep the ledger outside `docs/evidence/` until the release set is terminal and independently reviewed.

---

### Task 2: Publish AO Forge v0.1.5 through its signer gate

**Files:**
- Read: `../ao-forge/VERSION`
- Read: `../ao-forge/docs/release/V0.1.5-RELEASE-NOTES.md`
- Read: `../ao-forge/docs/release/RELEASE-SIGNERS.json`
- Read: `../ao-forge/.github/workflows/release-publish.yml`
- Read: `../ao-forge/.github/workflows/release-finalize.yml`

**Interfaces:**
- Consumes: source `d1723769949269dcd0589916d83769dcb7275f98`, rehearsal run `32532526229`, plan digest `f614c81743bf9f954bdf823bcbdfc49f5c1cc7543f48f0e8c3b5dab79157217f`.
- Produces: public signed release `v0.1.5` and independently verified public assets.

- [ ] **Step 1: Create the signed annotated tag on the authorized signer host**

Run only where the approved secret signing key is available:

```bash
git fetch origin main --tags
git tag -s v0.1.5 d1723769949269dcd0589916d83769dcb7275f98 -m "AO Forge v0.1.5"
git verify-tag v0.1.5
git push origin refs/tags/v0.1.5
```

Expected: `git verify-tag` reports a fingerprint present and active in `docs/release/RELEASE-SIGNERS.json`; the remote tag resolves to the exact source commit.

- [ ] **Step 2: Dispatch the protected draft publisher**

Run:

```bash
gh workflow run .github/workflows/release-publish.yml \
  --repo uesugitorachiyo/ao-forge --ref main \
  -f tag=v0.1.5 \
  -f source_commit=d1723769949269dcd0589916d83769dcb7275f98 \
  -f release_rehearsal_run_id=32532526229 \
  -f expected_plan_digest=f614c81743bf9f954bdf823bcbdfc49f5c1cc7543f48f0e8c3b5dab79157217f \
  -f confirm_publish=true \
  -f exact_confirmation=publish-ao-forge:v0.1.5:d1723769949269dcd0589916d83769dcb7275f98:32532526229:f614c81743bf9f954bdf823bcbdfc49f5c1cc7543f48f0e8c3b5dab79157217f
```

Expected: the workflow creates one draft release from the rehearsed candidates and does not rebuild them.

- [ ] **Step 3: Dispatch the finalizer**

Run:

```bash
gh workflow run .github/workflows/release-finalize.yml \
  --repo uesugitorachiyo/ao-forge --ref main \
  -f tag=v0.1.5 \
  -f source_commit=d1723769949269dcd0589916d83769dcb7275f98 \
  -f workflow_source_commit=d1723769949269dcd0589916d83769dcb7275f98 \
  -f expected_plan_digest=f614c81743bf9f954bdf823bcbdfc49f5c1cc7543f48f0e8c3b5dab79157217f \
  -f exact_confirmation=finalize-ao-forge:v0.1.5:d1723769949269dcd0589916d83769dcb7275f98:d1723769949269dcd0589916d83769dcb7275f98:f614c81743bf9f954bdf823bcbdfc49f5c1cc7543f48f0e8c3b5dab79157217f
```

Expected: protected environment approval succeeds, the release becomes public and non-prerelease, and finalizer verification exits zero.

---

### Task 3: Produce and publish the AO Command v0.1.3 approved candidate set

**Files:**
- Read: `../ao-command/docs/release/V0.1.3-OPERATOR-CLOSEOUT.md`
- Read: `../ao-command/scripts/release-rehearsal-verify.py`
- Read: `../ao-command/.github/workflows/release-rehearsal.yml`
- Create outside repositories: `approved-release-manifest.json`

**Interfaces:**
- Consumes: source `ffef6d76306e892c3e7a7f39734433d5a832006a`, version `0.1.3`, tag `v0.1.3`, and three approved archive digests.
- Produces: one strict `ao.command.approved-release-manifest.v0.1`, a successful rehearsal plan, and public three-platform assets.

- [ ] **Step 1: Obtain the approved candidate archives from the authorized asset producer**

Require exactly:

```text
ao-command-0.1.3-linux-x86_64.tar.gz
ao-command-0.1.3-macos-aarch64.tar.gz
ao-command-0.1.3-windows-x86_64.zip
```

Reject any archive built from a source other than `ffef6d76306e892c3e7a7f39734433d5a832006a` or without successful native Windows qualification.

- [ ] **Step 2: Create and verify the approved manifest**

Create strict JSON from the computed bytes:

```bash
linux_sha=$(shasum -a 256 ao-command-0.1.3-linux-x86_64.tar.gz | awk '{print $1}')
macos_sha=$(shasum -a 256 ao-command-0.1.3-macos-aarch64.tar.gz | awk '{print $1}')
windows_sha=$(shasum -a 256 ao-command-0.1.3-windows-x86_64.zip | awk '{print $1}')
release_notes_digest=$(shasum -a 256 ../ao-command/docs/release/V0.1.3-OPERATOR-CLOSEOUT.md | awk '{print $1}')
jq -n \
  --arg linux_sha "$linux_sha" \
  --arg macos_sha "$macos_sha" \
  --arg windows_sha "$windows_sha" \
  --arg notes "$release_notes_digest" \
  '{candidates:[
      {archive:"ao-command-0.1.3-linux-x86_64.tar.gz",archive_sha256:$linux_sha,target:"linux-x86_64"},
      {archive:"ao-command-0.1.3-macos-aarch64.tar.gz",archive_sha256:$macos_sha,target:"macos-aarch64"},
      {archive:"ao-command-0.1.3-windows-x86_64.zip",archive_sha256:$windows_sha,target:"windows-x86_64"}
    ],immutable:true,release_notes_digest:$notes,repository:"ao-command",
    schema_version:"ao.command.approved-release-manifest.v0.1",
    source_commit:"ffef6d76306e892c3e7a7f39734433d5a832006a",
    tag:"v0.1.3",version:"0.1.3"}' > approved-release-manifest.json
```

Run the repository verifier with `VERSION`, `TAG`, `SOURCE_COMMIT`, `RELEASE_NOTES_DIGEST`, and `APPROVED_MANIFEST_DIGEST` bound to the exact bytes. Expected: `python3 scripts/release-rehearsal-verify.py manifest` exits zero.

- [ ] **Step 3: Run dry rehearsal and verify the immutable plan**

Dispatch `.github/workflows/release-rehearsal.yml` with `dry_run=true`, the exact source/version/tag, base64 manifest bytes, and manifest digest. Download the plan artifact and run:

```bash
python3 scripts/release-rehearsal-verify.py verify
```

Expected: the three candidate archive digests equal the approved manifest and the plan reports no public mutation.

- [ ] **Step 4: Run the plan-bound live publisher**

Redispatch the same workflow with `dry_run=false`, the successful plan digest, and the exact confirmation string emitted by the workflow contract. Approve only the `protected-release` environment for that exact run.

Expected: `v0.1.3` is public, tag target equals the exact source, and independent public asset verification succeeds.

---

### Task 4: Publish AO Atlas v0.2.1 through its governed finalizer

**Files:**
- Read: `../ao-atlas/docs/release/v0.2.1.md`
- Read: `../ao-atlas/.github/workflows/release-rehearsal.yml`
- Read: `../ao-atlas/.github/workflows/release-finalize.yml`
- Read: `../ao-atlas/scripts/verify-release-rehearsal-candidates.sh`

**Interfaces:**
- Consumes: source `3603a2bb8af5adafcd9ff17b807ab89f32283d18`, version `v0.2.1`, tag `v0.2.1`, approved asset-manifest digest `sha256:6b35e072aa7a1c4d07fcf546fb1af1706870a625c13e61cd838ad905ffbe881b`, rehearsal run `32537634684`, and promotion-plan digest `03392a4ffdb197a4cfc69ac44bf0523deeac01dea449d51102b0abcb41f17bc3`.
- Produces: successful rehearsal and finalizer plans plus verified public Linux, macOS, and Windows assets.

- [ ] **Step 1: Bind the approved asset-manifest digest**

Hash the reviewed asset-manifest bytes:

```bash
approved_manifest_digest="sha256:$(shasum -a 256 approved-asset-manifest.json | awk '{print $1}')"
test "${#approved_manifest_digest}" -eq 71
```

Expected: the digest identifies the exact reviewed Atlas `v0.2.1` asset manifest; no digest is invented from release notes or source files.

- [ ] **Step 2: Dispatch and verify rehearsal**

Run `.github/workflows/release-rehearsal.yml` on exact source with `dry_run=true`, `version=v0.2.1`, `tag=v0.2.1`, and the approved manifest digest. Download the promotion plan and run `scripts/verify-release-rehearsal-candidates.sh` with the exact source, version, tag, and digest.

Expected: all three native candidates pass, the immutable plan is ready, and publication attempts remain zero.

- [ ] **Step 3: Dispatch the governed finalizer**

Run `.github/workflows/release-finalize.yml` with the producer run ID, exact source/version/tag, approved manifest digest, downloaded plan SHA-256, `dry_run=false`, and the exact `publish-imported-ao-atlas-...` confirmation string validated by the workflow.

Expected: the protected publisher and independent post-public verifier both pass; release `v0.2.1` is public and non-prerelease.

---

### Task 5: Replace stale AO Architecture public-release truth

**Files:**
- Modify: `../ao-architecture/stack/current-release-manifest.json`
- Modify: `../ao-architecture/docs/current-release.md`
- Modify: `../ao-architecture/docs/stack-assembly.md`
- Modify: `../ao-architecture/scripts/run_public_stack_canary.py`
- Modify: `../ao-architecture/scripts/test_run_public_stack_canary.py`
- Modify: `../ao-architecture/scripts/test_verify_current_release_manifest.py`

**Interfaces:**
- Consumes: fresh public release readbacks and downloaded SHA-256 values for Mission `v0.1.5`, Command `v0.1.3`, Atlas `v0.2.1`, and Forge `v0.1.5`.
- Produces: one canonical manifest and canary table that reference the same exact public assets.

- [ ] **Step 1: Write failing stale-version assertions**

Add assertions that the manifest, docs, and `_COMPONENTS` canary table contain the four versions above and reject the replaced versions `v0.1.4` for Mission, `v0.1.2` for Command, `v0.2.0` for Atlas, and `v0.1.4` for Forge.

- [ ] **Step 2: Verify the regression fails**

Run:

```bash
python3 scripts/test_verify_current_release_manifest.py
python3 scripts/test_run_public_stack_canary.py
```

Expected: failures identify the stale public release bindings.

- [ ] **Step 3: Update the four release records and canary assets**

For each component, copy exact tag target, workflow run, release URL, asset name, size, and freshly downloaded SHA-256 into the manifest and `_COMPONENTS`. Update both operator documents to the identical version set.

- [ ] **Step 4: Run Architecture verification**

Run:

```bash
python3 scripts/verify_current_release_manifest.py
python3 scripts/test_verify_current_release_manifest.py
python3 scripts/test_run_public_stack_canary.py
python3 scripts/verify_architecture.py
git diff --check
```

Expected: every command exits zero and no version/digest disagreement remains.

- [ ] **Step 5: Commit and merge through green PR CI**

Stage only the six files listed in this task. Commit with `docs: publish coherent AO stack release set`, open a non-draft PR, wait for every required check, and merge without bypassing protection.

---

### Task 6: Prove public clean-host use on three platforms

**Files:**
- Read: `../ao-architecture/.github/workflows/public-stack-canary.yml`
- Read: `../ao-architecture/scripts/run_public_stack_canary.py`
- Read: `../ao-architecture/docs/stack-assembly.md`
- Generated workflow artifacts only: `public-stack-canary-*.json`

**Interfaces:**
- Consumes: Architecture `main` containing Task 5 and public GitHub release assets only.
- Produces: green Linux, macOS, and Windows canaries plus operator first-use evidence.

- [ ] **Step 1: Dispatch the public canary on merged Architecture main**

Run:

```bash
gh workflow run .github/workflows/public-stack-canary.yml \
  --repo uesugitorachiyo/ao-architecture --ref main
```

Expected: `public-stack-canary (linux-x86_64)`, `(macos-aarch64)`, and `(windows-x86_64)` all pass.

- [ ] **Step 2: Verify clean-host first use**

On each supported host, follow `docs/stack-assembly.md` exactly: download the documented archive, verify SHA-256, extract, run version/help, run the credential-free governed demo, reconcile Mission/Atlas/Command identity, then remove temporary state.

Expected: zero undocumented environment changes, zero provider credentials, identical terminal identity, and zero remaining processes, listeners, or temporary state.

- [ ] **Step 3: Repair only observed operator friction**

If a documented command fails, add one focused regression to the owning repository, fix the shared root cause, rerun its repository matrix, republish only the affected patch, update Architecture bindings, and rerun the complete public canary.

---

### Task 7: Reconcile and declare the supported public set

**Files:**
- Modify: `../ao-architecture/docs/current-release.md`
- Modify: `../ao-architecture/stack/current-release-manifest.json`
- Modify only if lifecycle truth changes: `../ao-architecture/AGENTS.md`
- Do not modify historical closure or Windows evidence.

**Interfaces:**
- Consumes: public releases, fresh-download digests, canary run IDs, clean-host results, and clean synchronized `main` branches.
- Produces: one final Architecture-owned stability statement and exact next action.

- [ ] **Step 1: Verify terminal release state**

Require all of the following: four new public releases, exact tag targets, expected public assets, fresh hashes, green public canary on three platforms, clean first-use results, green default-branch CI, clean synchronized repositories, and no stale release branches/worktrees.

- [ ] **Step 2: Record the final canary and release identities**

Update `generated_at_utc`, `source_of_truth`, public canary run, artifact digests, version pairing, and release workflow links from actual readbacks. Do not set `full_stack_compatibility_complete` or activate compatibility unless the separately governed compatibility gate also passes.

- [ ] **Step 3: Run final gates**

Run Architecture verification, every affected repository's production-readiness gate, agent-instruction layout verification, and `git diff --check`. Confirm all hosted PR and default-branch checks are green.

- [ ] **Step 4: Publish the stability summary**

State the exact supported version set, platform limitations, checksum source, canary run, installation entry point, and remaining denied authorities. The summary may say stable only if every requirement in Step 1 is true.
