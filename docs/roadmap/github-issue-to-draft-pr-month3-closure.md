# GitHub Issue To Draft PR Month 3 Closure

Month 3 defines the isolated repair, verification, rollback, replay, and resume
gate for the GitHub issue-to-draft-PR roadmap.

The required Month 3 state is:

- Only fixtures classified as authentic bugs after Month 2 can enter repair.
- Repair runs in a bounded disposable workgraph and isolated workspace.
- A failing pre-patch regression and a passing negative control are recorded
  before implementation.
- The selected repair is the smallest sufficient repair and preserves the
  regression test.
- False fixes that delete or weaken tests, disable checks, change unrelated
  files, widen authority, or add artifacts are rejected.
- Post-patch verification, rollback, replay, and interruption-resume evidence
  are recorded.
- Rollback restores the exact pre-change digest.
- Replay accepts only matching evidence digests.
- Resume does not duplicate edits.
- Control Plane observation and Command readback show the repair state.
- Sentinel catches false-fix, rollback, replay, provider-pilot, release, and
  RSI overclaims.

Feature-generated pull requests still do not exist in Month 3. When later
authorized, feature-generated pull requests remain draft-only by policy; they
are not merged, marked ready, or approved by this roadmap.

RSI remains denied. External beta is not launched. Promotion is not requested
or granted. No release, tag, upload, or deployment is selected by Month 3.

Next action after Month 3 closure: continue to Month 4 AO repository issue-to-draft-PR workflow.
Create any feature-generated pull request only as a draft, and do not approve,
mark ready, or merge it automatically.
