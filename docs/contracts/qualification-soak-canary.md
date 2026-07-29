# Qualification soak canary contract

`ao-mission qualification soak-canary` is a bounded operational consumer for
one fixed ten-node local qualification canary. It is separate from
`qualification soak-plan`: the planner remains read-only and reports planning
eligibility only. Execution also requires an independently digest-bound
`ao.mission.soak-canary-authority.v1` record.

The command consumes strict, bounded JSON for the plan input, authority,
command catalog, and activation manifest. It rejects unknown fields, duplicate
keys, trailing JSON, oversized files, symlinks, special files, changed
digests, and paths outside the authority-bound evidence root. Validation
rebuilds the planner readback and recomputes every authority, catalog, and
activation digest before the executor is reachable.

```sh
ao-mission qualification soak-canary \
  --plan /path/to/plan.json \
  --authority /path/to/authority.json \
  --catalog /path/to/catalog.json \
  --activation /path/to/activation.json \
  --checkpoint /path/to/evidence/checkpoints/checkpoint.json \
  --evidence-root /path/to/evidence \
  --repository-root /path/to/ao-mission \
  --validate-only \
  --json
```

Remove `--validate-only` only for an explicitly authorized canary. Execution
requires a clean repository whose exact `HEAD` matches the plan, catalog,
authority, and activation manifest.

## Fixed execution boundary

The command catalog contains exactly one approved scale test and nine approved
regular tests. Every command uses an absolute regular `go` executable whose
bytes match its SHA-256, repository-relative working directory `.`, package
`./internal/mission`, race mode, an exact anchored test or subtest expression,
`-json`, a millisecond timeout, and the approved effective repeat count.

Commands are passed to `exec.CommandContext` as argv. No shell is involved.
The manifest cannot select another executable, package, test, working
directory, environment variable, or network mode. The environment always
binds `GOTOOLCHAIN=local`, `GOPROXY=off`, `GOSUMDB=off`, and `GOVCS=*:off`.

The scale test runs with repeat one and cannot retry. Regular tests run with
repeat three. One authority-bound regular node records a single
`transient_infrastructure` attempt before process creation and may then launch
its unchanged command once. This yields eleven attempt records and ten child
process launches for ten completed nodes.

## Checkpoint and evidence

Every attempt binds the original phase start, source head, plan and policy
digests, execution profile, command catalog, authority, activation manifest,
argv, node, partition, test, repeat count, and safety boundaries. Atomic
checkpoints preserve the attempt digest chain, completed-node set, controlled
retry consumption, and scale-launch consumption. Restart recomputes completion
from signed attempts, skips completed nodes, and cannot repeat the scale node.

Stdout and stderr are bounded before persistence and recorded with relative
paths, byte counts, truncation state, and SHA-256. Reconciliation reopens each
artifact through the bounded regular-file reader and rejects digest or path
changes. Exact Go JSON pass events must equal one for the scale test and three
for each regular test. Actual duration must remain within the planned estimate,
attempt timeout, total-node timeout, node budget, aggregate allowance, lease,
and 45-minute authority wall.

## Terminal truth

The operational summary alone reports
`local_test_execution_performed=true` and the child-process launch count.
Planner and terminal-reader surfaces remain read-only and non-executing.
Mission's `inspect`, `checkpoint`, `event-index`, and `command-readback`
surfaces share one canonical payload and `index_digest`; each surface has its
own valid `state_digest`. Distinct surface digests are expected and do not
indicate disagreement.

The checked-in activation examples demonstrate a valid self-digest and a
digest-mismatch rejection. They are contract examples, not reusable execution
authority. A real activation must be regenerated for the exact repository
head, executable, evidence root, handoff digest, phase start, plan, and command
catalog.
