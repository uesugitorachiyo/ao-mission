# AO Stack Month 1 Closure — Stable Release and First-Operator Readiness

Status: closed
Closed: 2026-07-15
Next: Month 2 workflow reliability hardening

## Outcome

Month 1 is closed after AO2 `v0.5.0` stable release verification, public
install/support documentation updates, and the clean-machine support drill were
completed and verified.

The closure is evidence-based. It does not create a new release, change tags,
upload assets, deploy anything, run provider pilots, or expand autonomy. RSI
remains denied.

## Release State

- AO2 stable release: https://github.com/uesugitorachiyo/ao2/releases/tag/v0.5.0
- Release readback: `isDraft=false`, `isPrerelease=false`, `asset_count=23`
- AO2 tag target: `a1e82b0adb723dd5ae2be6d93355ffdc2caa549d`
- Compatible stable companion:
  [AO2 Control Plane v0.1.15](https://github.com/uesugitorachiyo/ao2-control-plane/releases/tag/v0.1.15)
- Clean-machine archive: `ao2-0.5.0-macos-aarch64.tar.gz`
- Clean-machine archive SHA-256:
  `a168235641779cb05186b3ed5b55bdd6676be275f0f64727eeab2e94ee6d2dfe`
- Clean-machine extracted binary SHA-256:
  `914ac7de8a5cbd50a52a96cb1a3a1aed2b43311219e72e3e97f5449da70d5532`

## PRs And Commits

- AO Mission PR #83: https://github.com/uesugitorachiyo/ao-mission/pull/83
  - Merge commit: `cb04f19c60519a6b5a4ec0a07781f244e696a034`
  - Result: added the post-stable six-month roadmap.
- AO2 PR #281: https://github.com/uesugitorachiyo/ao2/pull/281
  - Merge commit: `a1cfbf3997bef6b39e3c57a1cb1ac2d8e27dbe22`
  - Result: added public install/support docs, first-operator guide,
    troubleshooting, release docs, and issue templates for stable `v0.5.0`.
- AO2 PR #282: https://github.com/uesugitorachiyo/ao2/pull/282
  - Merge commit: `5644b36dc58471ef521e68007f988fabb794122d`
  - Result: closed clean-machine support drill gaps in docs and support
    template fields.

## Evidence

- Roadmap kickoff evidence:
  `canary-test/ao-stack-post-stable-roadmap-start-20260715T132103Z`
- Roadmap start report:
  `canary-test/ao-stack-post-stable-roadmap-start-20260715T132103Z/roadmap-start-report.md`
- Month 1 kickoff evidence report:
  `canary-test/ao-stack-post-stable-roadmap-start-20260715T132103Z/month1-evidence-report.md`
- Clean-machine support drill evidence:
  `canary-test/ao-stack-clean-machine-support-drill-20260715T141053Z`
- Clean-machine support drill final report:
  `canary-test/ao-stack-clean-machine-support-drill-20260715T141053Z/final-report.md`
- AO2 PR #282 CI readback: `95` checks in `SUCCESS`.

## What Changed For Users

- First 30 minutes guide:
  [`docs/FIRST-30-MINUTES.md`](https://github.com/uesugitorachiyo/ao2/blob/main/docs/FIRST-30-MINUTES.md)
- Install and update guide:
  [`docs/INSTALL.md`](https://github.com/uesugitorachiyo/ao2/blob/main/docs/INSTALL.md)
- Troubleshooting guide:
  [`docs/TROUBLESHOOTING.md`](https://github.com/uesugitorachiyo/ao2/blob/main/docs/TROUBLESHOOTING.md)
- Public release verification docs:
  [`docs/release/PUBLIC-RELEASE-VERIFICATION.md`](https://github.com/uesugitorachiyo/ao2/blob/main/docs/release/PUBLIC-RELEASE-VERIFICATION.md)
- Stable release notes:
  [`docs/release/v0.5.0-stable.md`](https://github.com/uesugitorachiyo/ao2/blob/main/docs/release/v0.5.0-stable.md)
- Public issue templates:
  [`bug_report.yml`](https://github.com/uesugitorachiyo/ao2/blob/main/.github/ISSUE_TEMPLATE/bug_report.yml)
  and
  [`support.yml`](https://github.com/uesugitorachiyo/ao2/blob/main/.github/ISSUE_TEMPLATE/support.yml)

The docs now cover public download, selected checksum verification, offline
archive verification, install, PATH setup, first command checks, a local
governed demo, rollback, uninstall, manifest mismatch handling, approval digest
context, and public-safe issue filing.

## Clean-Machine Support Drill Result

The clean-machine support drill started from the public AO2 `v0.5.0` release
and followed the public first-operator docs with a clean HOME and clean work
directory.

Initial public docs did not pass as written:

- `gh release download` required GitHub CLI auth in a clean HOME and had no
  direct public URL fallback.
- The post-install `ao2 version --json` command failed until the install
  directory was added to `PATH`.
- The governed demo accepted the run, but the guide did not explain that a
  nested provider score can be low for the provider-free scripted demo.
- The install guide needed clearer first-operator scope.

AO2 PR #282 fixed those blocker and major documentation gaps. Current AO2
`main` passes the verified macOS arm64 direct-download path:

1. Public `curl` download.
2. Selected archive checksum verification.
3. Archive extraction.
4. Offline release verification.
5. Install.
6. PATH setup.
7. `ao2 version --json`.
8. `ao2 doctor --json`.

## Closure Criteria

- Stable AO2 `v0.5.0` public release verified: closed.
- Control Plane companion documented as `v0.1.15`: closed.
- Public install path documented: closed.
- Public first-operator path documented: closed.
- Clean-machine support drill completed: closed.
- Blocker and major documentation gaps closed: closed.
- AO2 issue templates support stable user reports: closed.
- Rollback and offline verification guidance exists: closed.
- No provider pilot or external user contact was required: closed.
- RSI remains denied: closed.
- No Helix work: closed.
- Month 2 next action is clear: closed.

## Boundaries

- No `/tt` work.
- No modules work.
- No provider pilot.
- No external user contact.
- No Helix work.
- No RSI work; RSI remains denied.
- No release, tag, upload, deployment, or new binary publication from this
  closure task.

## Month 2 Handoff

Start Month 2 with workflow reliability hardening. This is not a new release
train by default.

Focus areas:

- Approval and replay reliability.
- Manifest mismatch diagnostics.
- Workflow template tests.
- Cross-platform packaging verification stability.
- Support issue reproduction fixtures.
