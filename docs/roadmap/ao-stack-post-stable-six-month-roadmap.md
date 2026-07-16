# AO Stack Post-Stable Six-Month Roadmap

Status: future reference  
Created: 2026-07-15  
Trigger: start after AO2 v0.5.0 stable publication is complete and verified.

## Theme

Move AO Stack from a verified stable release to a controlled self-improvement system. The roadmap favors reliability, evidence quality, operator clarity, and repeatable release discipline before any higher-autonomy work.

RSI remains denied. The roadmap may add controlled self-improvement workflows, but those workflows must require measurement, sandboxing, rollback plans, and human approval before merge or release.

## Month 1: Stable release closure and adoption readiness

Status: closed on 2026-07-15. See
[Month 1 stable release closure](month1-stable-release-closure.md).

Goal: make the stable release usable, documented, and supportable.

Work:

- Verify AO2 v0.5.0 and Control Plane compatibility after publication.
- Update install docs, README, quickstarts, and release notes.
- Create a “first 30 minutes with AO2” guide.
- Add troubleshooting docs for approvals, manifests, local pilots, rollback, and offline verification.
- Run fresh installs from public assets on macOS, Linux, and Windows.
- Create public issue templates and support triage docs.

Success criteria:

- Public stable release is verified.
- Install, rollback, and uninstall are tested.
- A new user can run a basic workflow without reading internal evidence docs.

Closure readback:

- AO2 `v0.5.0` stable release is public and verified.
- AO2 Control Plane `v0.1.15` remains the compatible stable companion.
- Public install, first-operator, troubleshooting, rollback, offline
  verification, and support issue paths are documented.
- The clean-machine support drill passed after AO2 docs fixes in PR #282.
- No provider pilot, external user contact, forbidden-tool work, RSI work, `/tt` work,
  modules work, or new release work was required.

## Month 2: Workflow reliability hardening

Status: closed. The workflow reliability waves and compatibility workgraph
handoff are complete. Month 3 closed the remaining compatibility matrix edges.

Goal: reduce operator confusion and agent overreach.

Work:

- Harden approval-gate UX.
- Improve “waiting for approval” recovery.
- Add clearer action digest explanations.
- Add reusable local-private-pilot workflow support.
- Improve final-status rules so agents do not overclaim.
- Add better failure summaries for CI, build, release, and runtime blockers.
- Expand replay and offline verification examples.

Compatibility evidence handoff:

- AO2 execution receipt to AO2 Control Plane evidence event is tested.
- AO2 Control Plane readback to AO Command operator status is tested.
- AO Mission run status/timeline to AO Command operator timeline is tested.
- All 16 live AO Architecture matrix edges now have canonical vectors and
  consumer tests after Month 3 closure.
- The Architecture compatibility gate remains false under the current
  proposed/gated matrix status.
- External beta has not launched, promotion is not requested, and RSI remains denied.

Success criteria:

- Failed or paused runs are easier to resume.
- Approval-bound workflows are understandable to a solo operator.
- Evidence is readable without deep AO Stack context.

## Month 3: Evidence and audit system upgrade

Status: closed on 2026-07-15. See
[Month 3 evidence and audit compatibility closure](month3-evidence-audit-compatibility-closure.md).

Goal: make AO evidence useful as a product feature, not just logs.

Work:

- Standardize evidence packs across AO2, Mission, Atlas, Command, Sentinel, and related repos.
- Add evidence summary generation.
- Add machine-readable and human-readable report pairs.
- Add artifact guard reports.
- Add API and ABI preservation reports.
- Add device/runtime smoke evidence templates.
- Add local-only/private-mode evidence examples.

Success criteria:

- Every serious AO task produces consistent evidence.
- Evidence can answer what changed, why it changed, how it was verified, and what remains blocked.

Closure readback:

- AO Architecture records 16 tested current-release compatibility edges.
- `canonical_vector_count=16` and `consumer_test_count=16`.
- Remaining proposed edges: 0.
- The compatibility gate remains false under the current proposed/gated matrix
  status.
- External beta has not launched, promotion is not requested or granted, and
  RSI remains denied.

## Month 4: Controlled self-improvement loop

Status: closed on 2026-07-16. See
[Month 4 controlled self-improvement dry-run closure](month4-controlled-self-improvement-dry-run-closure.md).

Goal: let AO improve workflows safely, without RSI.

Work:

- Add an “AO proposes improvement” workflow.
- Add before/after measurement requirements.
- Add sandbox-only improvement testing.
- Require human approval before applying improvements.
- Require a rollback plan for every self-improvement change.
- Create evaluation suites for prompt, workflow, and template changes.
- Track whether changes improve completion, safety, or clarity.

Success criteria:

- AO can suggest improvements to AO.
- AO cannot autonomously merge, release, or expand authority.
- Every improvement has measurable before/after evidence.

Closure readback:

- The controlled self-improvement loop is defined as fixture-only and dry-run
  only.
- Human approval, rollback proof, observation, operator readback, Sentinel
  wording checks, and Promoter no-RSI/no-promotion verdict evidence are merged.
- RSI remains denied.
- Live self-modification, provider execution, external beta, promotion, release,
  tag, upload, and deployment remain denied.

## Month 5: Multi-repo product coordination

Status: closed on 2026-07-16. See
[Month 5 operator workflow hardening closure](month5-operator-workflow-hardening-closure.md).

Goal: make AO Stack feel like one product instead of many repos.

Work:

- Maintain a cross-repo version compatibility matrix.
- Add contract checks between AO2, Control Plane, Mission, Atlas, Command, Sentinel, Promoter, and related repos.
- Add a shared release-state manifest.
- Add a shared docs map.
- Add one “what repo owns what” page.
- Improve local development bootstrap for all AO repos.

Success criteria:

- A maintainer can tell which repo owns each function.
- Stable releases are coordinated.
- Compatibility failures are caught before publication.

Closure readback:

- AO Architecture defines the operator workflow source of truth.
- AO Command presents current stack state, release pair, compatibility state,
  dry-run/no-RSI state, policy gate state, safe-next-work, and support evidence.
- AO Foundry and AO Forge provide safe-next-work and run-state fixtures.
- AO Covenant records policy approval requirements and denied authorities.
- AO Sentinel catches Month 5 operator workflow overclaims.
- AO Promoter records no-promotion and no-RSI readback.
- AO2 and AO2 Control Plane support/readback paths were reviewed and did not
  need Month 5 changes.
- RSI remains denied. External beta remains not launched, and promotion is not
  requested or granted.

## Month 6: Next stable release train

Goal: ship the next stable release with stronger process than AO2 v0.5.0.

Work:

- Select the next stable target, likely AO2 v0.6.0 or equivalent.
- Include improvements from Months 1 through 5.
- Run full release qualification.
- Run public asset verification.
- Run private/local pilot only if needed, not as endless canary work.
- Publish stable release only after green evidence.

Success criteria:

- Another stable release ships.
- Release process is smoother than AO2 v0.5.0.
- AO Stack has controlled self-improvement support, while RSI remains denied.

## Recommendation

Start Month 6 with next stable release train planning and readiness assessment.
Use the Month 3 compatibility matrix, Month 4 dry-run evidence, and Month 5
operator workflow readback as the evidence base. RSI remains denied.
