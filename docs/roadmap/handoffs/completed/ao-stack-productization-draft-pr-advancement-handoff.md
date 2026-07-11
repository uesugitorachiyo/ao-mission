> COMPLETED HISTORICAL HANDOFF - DO NOT REEXECUTE.
> Preserved for audit history only; use the active handoff directory for current execution.

# AO Mission Continuation: Advancement Journal And Draft PR Lifecycle

Continue the existing AO Mission `mission-4d91b0a9e4ab273e`. Do not create a
new mission, replace its objective, or count this remote-lifecycle handoff as a
completed implementation node.

## Exact Operator Authorization

By sending this prompt, I explicitly authorize only the following actions:

1. Write and update the local advancement journal described below under the
   existing untracked `.ao-mission` store.
2. Read remote metadata and fetch existing `origin` refs without changing
   credentials or repository settings.
3. Push the three exact branches and exact branch-head commits listed in this
   prompt to each repository's already configured `origin` remote.
4. Create or update one draft pull request per authorized branch, targeting
   that repository's default branch.
5. Read CI, review, and pull-request status and record those readbacks in the
   advancement journal.

This authorization does not permit:

- force-push, history rewriting, rebase, amend, or replacement commits;
- pushing any branch or commit not listed here;
- pushing the separate AO Architecture design branch;
- marking a draft PR ready, approving or merging a PR, enabling auto-merge, or
  bypassing branch protection;
- closing PRs, deleting branches, tags, or releases;
- release, deployment, publication, upload, package publication, or credential
  creation or modification;
- direct default-branch mutation;
- AO2 provider execution, patch application, promotion, or RSI authority; or
- any action whose resolved remote owner or repository differs from the
  current configured `origin` for the corresponding local repository.

If an action falls outside the explicit allowlist, stop and record it as the
exact next authorization request. Do not infer broader permission from the
phrase "continue the mission."

Use `codex_sol_high`: provider `codex`, model `gpt-5.6-sol`, reasoning effort
`high`, for this gate-sensitive lifecycle work. Do not downgrade to Terra or
Luna. Record requested and resolved profile, model, reasoning effort, CLI
version, role, start and finish times, and fallback status. Silent fallback is
forbidden.

## Durable Advancement Journal

Create these local, public-safe, untracked files:

- `.ao-mission/handoffs/mission-4d91b0a9e4ab273e/advancement/advancement-log.jsonl`
- `.ao-mission/handoffs/mission-4d91b0a9e4ab273e/advancement/advancement-summary.md`
- `.ao-mission/handoffs/mission-4d91b0a9e4ab273e/advancement/advancement-manifest.json`

Never stage or commit `.ao-mission`. Never hand-edit the Mission record JSON.
The journal is a readback artifact and grants no authority.

Append one compact JSON object after every material event. Never delete or
rewrite an earlier entry. Each entry must contain:

- `schema: "ao.mission.advancement-entry.v0.1"`;
- monotonically increasing `sequence`;
- `mission_id` and UTC timestamp;
- program phase, slice ID, event type, and status;
- repository, branch, exact base and head commit, and clean-worktree result;
- exact verification command, exit code, and concise result;
- relevant artifact, patch, authorization, and workgraph digests;
- remote repository identity, PR URL, CI check URL, and remote commit when
  available;
- requested and resolved model profile evidence;
- authority flags, all false for execution, approval, promotion, release,
  credential use, direct-main mutation, and concurrent mutation;
- blocker or residual risk; and
- one exact next governed action.

After every append, regenerate `advancement-summary.md` from the JSONL rather
than maintaining a second independent narrative. Regenerate
`advancement-manifest.json` with:

- `schema: "ao.mission.advancement-manifest.v0.1"`;
- mission ID;
- entry count;
- SHA-256 of the exact JSONL bytes;
- SHA-256 of the generated Markdown summary;
- last sequence and last event;
- all authority flags false; and
- generation timestamp.

Write journal and manifest updates atomically. Do not put credentials, tokens,
environment dumps, private prompts, full terminal transcripts, or private Git
URLs in the log. Store exit codes and concise public-safe summaries.

The first journal entry must capture the local advancement snapshot below
before any network or remote mutation action.

## Correct Local Advancement Snapshot

Record these verified local facts without collapsing branch heads into a false
"three commits total" claim:

### AO Architecture Planning Evidence

- worktree: `/tmp/ao-architecture-productization-spec`
- branch: `codex/ao-stack-productization-spec`
- head: `e0094288ec4e1e1162061007ffaf036f0f60bbd5`
- commit: `e009428 docs: add AO stack productization design`
- status: approved local design evidence, clean, not authorized for push by
  this prompt, and not part of the three-branch runtime dependency chain.

### AO Atlas Runtime Dependency 1

- worktree: `/private/tmp/ao-atlas-persisted-artifact-digests`
- branch: `codex/atlas-persisted-artifact-digests`
- base: `0d74243f6d61485f9e86a99067b1ab22e92c8da3`
- head: `0ff6a696fb965671fdb7c3894b1b70f3ea16536c`
- commit: `0ff6a69 Bind Atlas downstream artifacts to persisted bytes`
- expected delta: three Atlas source/test files, 58 insertions and 2 deletions.

### AO Foundry Runtime Dependency 2

- worktree: `/private/tmp/ao-foundry-blueprint-bridge-plan`
- branch: `codex/foundry-blueprint-authorization-bridge`
- base: current Foundry `main` at branch creation,
  `a90a4f950bbc7535ed1e05ab9331fb7ae12989be`
- head: `cdd06372f4be12725183d7da02ee5a36effd79a5`
- branch contains two commits beyond that base, not one:
  `a0c7cd1 docs: plan Blueprint authorization Pulse bridge` and
  `cdd06372 Validate canonical Blueprint authorization in Pulse`.
- expected delta: the plan plus the canonical Blueprint consumer, tests,
  fixtures, and `scripts/verify-blueprint-atlas-pulse-contract.sh`.

### AO Mission Runtime Dependency 3

- worktree: `/private/tmp/ao-mission-handoff-accounting`
- branch: `codex/mission-handoff-accounting`
- base: `8055933d9eebdd12a01d9ccca6edcbbf9bc69c37`
- head: `a69d806d069aa12b90a8fe40a4499038c27e2936`
- commit: `a69d806 Do not count handoffs as completed nodes`
- expected delta: `internal/mission/supervisor.go` and its regression test.

### Integration Evidence

- `go test ./... -count=1`, `go vet ./...`, and the public command build passed
  in each owning implementation repository according to the latest local
  readback. Re-run the repository-specific gates below before pushing.
- The disposable cross-repository preflight reported
  `blueprint_atlas_pulse_contract=ready`.
- The exact canonical Blueprint authorization SHA-256 remained
  `b7e18a1967b31b2806184444ab9aeab5e984e050f66261431ec57ece4cc833ee`.
- The current canonical Mission record still reflects the pre-merge mainline
  state. Local branch completion must not be recorded as merged or as a
  completed Atlas node.

## Pre-Push Gates

Process Atlas, Foundry, and Mission serially in that order. Before any push for
each repository:

1. Verify the worktree path, repository identity, branch name, exact head, and
   clean status against this prompt.
2. Verify `git show --check` and inspect the complete `origin/main...HEAD` diff
   for unrelated changes, credentials, local paths, generated bulk evidence,
   or authority widening.
3. Fetch `origin` read-only. Verify the default branch and determine whether it
   advanced after the recorded base.
4. Do not rebase or rewrite. If the branch no longer applies cleanly or the
   verified diff changes, append a blocked journal entry and stop that branch.
5. Check whether the remote branch or a PR already exists. If it exists at the
   same head, treat the push/PR action as idempotently complete. If it exists at
   a different head, do not force-push; record the mismatch and stop.
6. Run fresh verification:
   - Atlas: `go test ./... -count=1`, `go vet ./...`, and build its public CLI.
   - Foundry: `go test ./... -count=1`, `go vet ./...`, and
     `go build ./cmd/foundry`.
   - Mission: `go test ./... -count=1`, `go vet ./...`, and
     `go build ./cmd/ao-mission`, then remove only the generated local binary.
7. Re-run the disposable Atlas-to-Foundry contract verifier if its workspace
   still exists. If it does not, record the prior verified result and the
   missing disposable workspace honestly; do not fabricate a fresh pass.
8. Recompute the Blueprint authorization SHA-256 and require the exact value
   above.
9. Append a pre-push journal entry and regenerate the summary and manifest.

If any gate fails, do not repair or create replacement commits under this
authorization. Log the failure and return an exact bounded repair request,
because a new commit would fall outside the authorized head SHA.

## Authorized Push And Draft PR Sequence

Use the installed GitHub workflow or `gh` only after the corresponding
pre-push gates pass. Never expose credentials in output.

### 1. AO Atlas

Push only `codex/atlas-persisted-artifact-digests` at
`0ff6a696fb965671fdb7c3894b1b70f3ea16536c`. Create one draft PR to the
verified default branch. The PR body must include scope, changed files,
verification, rollback by reverting the commit, denied authorities, and the
fact that Foundry and Mission PRs depend on it.

Append separate journal entries for push and draft-PR creation, including the
remote head and PR URL.

### 2. AO Foundry

Push only `codex/foundry-blueprint-authorization-bridge` at
`cdd06372f4be12725183d7da02ee5a36effd79a5`. The draft PR must disclose both
commits on the branch, link the Atlas draft PR as a dependency, identify the
canonical Blueprint contract and authorization SHA-256, list positive and
negative tests, and state that public-safe path and digest checks remain
fail-closed.

Append separate journal entries for push and draft-PR creation.

### 3. AO Mission

Push only `codex/mission-handoff-accounting` at
`a69d806d069aa12b90a8fe40a4499038c27e2936`. The draft PR must link the Atlas
and Foundry draft PRs, explain that handoff steps no longer inflate completed
node counts, include full Mission verification, and preserve all no-authority
boundaries.

Append separate journal entries for push and draft-PR creation.

## CI And Review Readback

After creating all permitted draft PRs:

1. Read each PR back from GitHub and verify repository, base, head branch, head
   SHA, draft status, and dependency links.
2. Observe required CI checks until they reach a terminal state or a bounded
   waiting limit is reached.
3. Record check name, conclusion, URL, and tested head SHA. Do not treat a check
   on an older SHA as current evidence.
4. Do not rerun privileged workflows, approve deployments, dismiss reviews, or
   change branch protection.
5. If CI fails, record the failure and exact first failing check. Do not create
   a repair commit under this authorization.
6. Regenerate the advancement summary and manifest after every CI state change
   that changes the exact next action.

## Merge Gate And Return Contract

Do not merge any PR in this continuation. Even if all CI passes, stop at the
merge approval gate. The required merge order is Atlas, Foundry, then Mission.
AO Architecture remains a separate planning-evidence integration decision.

Return one concise readback containing:

- advancement journal, summary, and manifest paths and SHA-256 values;
- entry count and last sequence;
- every local and remote branch head;
- all draft PR URLs and dependency order;
- CI status and tested SHA for every required check;
- the unchanged Blueprint authorization SHA-256;
- blockers and residual risks;
- confirmation that no merge, force-push, release, credential modification,
  direct-main mutation, AO2 execution, or Architecture push occurred;
- the exact next merge-authorization prompt, scoped to PR numbers and immutable
  head SHAs; and
- `final_response_allowed=false`.

Do not respond only with "remote mutation is prohibited." This prompt provides
exact authorization for the listed pushes and draft PRs. If credentials,
remote identity, branch protection, head SHA, or CI prevents those actions,
write the advancement entry first and then return the precise blocker without
retry loops or broader credential requests.
