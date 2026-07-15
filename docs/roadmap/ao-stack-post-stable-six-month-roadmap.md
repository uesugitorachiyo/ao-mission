# AO Stack Post-Stable Six-Month Roadmap

Status: future reference  
Created: 2026-07-15  
Trigger: start after AO2 v0.5.0 stable publication is complete and verified.

## Theme

Move AO Stack from a verified stable release to a controlled self-improvement system. The roadmap favors reliability, evidence quality, operator clarity, and repeatable release discipline before any higher-autonomy work.

RSI remains denied. The roadmap may add controlled self-improvement workflows, but those workflows must require measurement, sandboxing, rollback plans, and human approval before merge or release.

## Month 1: Stable release closure and adoption readiness

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

## Month 2: Workflow reliability hardening

Goal: reduce operator confusion and agent overreach.

Work:

- Harden approval-gate UX.
- Improve “waiting for approval” recovery.
- Add clearer action digest explanations.
- Add reusable local-private-pilot workflow support.
- Improve final-status rules so agents do not overclaim.
- Add better failure summaries for CI, build, release, and runtime blockers.
- Expand replay and offline verification examples.

Success criteria:

- Failed or paused runs are easier to resume.
- Approval-bound workflows are understandable to a solo operator.
- Evidence is readable without deep AO Stack context.

## Month 3: Evidence and audit system upgrade

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

## Month 4: Controlled self-improvement loop

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

## Month 5: Multi-repo product coordination

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

Start with Month 1 after AO2 v0.5.0 stable publication is verified. Do not start RSI work. Build the controlled self-improvement loop in Month 4 after release reliability and evidence quality are stronger.
