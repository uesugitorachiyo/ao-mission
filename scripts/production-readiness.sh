#!/usr/bin/env sh
set -eu
go test ./... -count=1 >/tmp/ao-mission-go-test.log
go vet ./... >/tmp/ao-mission-go-vet.log
go build ./cmd/ao-mission
gofmt -w cmd internal
git diff --check
local_path="/""Users/"
private_key="BEGIN (RSA |OPENSSH |PRIVATE )?PRIVATE ""KEY"
openai_key="sk-""[A-Za-z0-9]{20,}"
github_key="gh[pousr]_""[A-Za-z0-9]{20,}"
secret_assign="tok""en[[:space:]]*[:=][[:space:]]*[^[:space:]]+"
public_safety_pattern="(${local_path}|${private_key}|${openai_key}|${github_key}|${secret_assign})"
for f in README.md docs examples internal cmd scripts; do
  if grep -R -nE "$public_safety_pattern" "$f" >/tmp/ao-mission-public-safety.log 2>/dev/null; then
    echo "public safety scan failed"
    cat /tmp/ao-mission-public-safety.log
    exit 1
  fi
done
tmp_home="$(mktemp -d)"
mission_json="$(mktemp)"
import_json="$(mktemp)"
inspect_json="$(mktemp)"
reconcile_json="$(mktemp)"
./ao-mission --home "$tmp_home" start "import completed Atlas recommendation wave" >"$mission_json"
mission_id="$(jq -r '.mission_id' "$mission_json")"
./ao-mission --home "$tmp_home" import atlas-recommendation-readback --mission "$mission_id" --path examples/valid/atlas-recommendation-readback.json >"$import_json"
jq -e '.kind == "atlas-recommendation-readback" and .safe_to_execute == false and .executes_work == false and .approves_work == false' "$import_json" >/dev/null
./ao-mission --home "$tmp_home" mission inspect --mission "$mission_id" --json >"$inspect_json"
jq -e '.status == "done" and .current_route == "complete" and .current_phase == "complete" and .evidence.atlas_recommendation.completed_nodes == 40 and .return_gate.final_response_allowed == true' "$inspect_json" >/dev/null
./ao-mission --home "$tmp_home" final reconcile --mission "$mission_id" >"$reconcile_json"
jq -e '.schema == "ao.mission.final-reconciliation-packet.v0.1" and .artifacts_agree == true and .final_response_allowed == true and .claims_authority_advance == false and .rsi_remains_denied == true' "$reconcile_json" >/dev/null
rm -rf "$tmp_home" "$mission_json" "$import_json" "$inspect_json" "$reconcile_json" ao-mission
grep -q "25-node Atlas recommendation import wave" docs/operator-next-actions.md
grep -q "Do not stop before 25 completed nodes" docs/evidence/ao-mission-atlas-wave-import-v01/next-recommended-prompt.md
grep -q "final-reconciliation-packet.json" docs/operator-next-actions.md
grep -q "Command and final reconciliation closure check" docs/operator-next-actions.md
jq -e '.schema == "ao.mission.final-reconciliation-packet.v0.1" and .status == "ready" and .artifacts_agree == true and .promotion_claimed == false and .rsi_remains_denied == true and .claims_authority_advance == false and .safe_to_execute == false and .executes_work == false and .approves_work == false' examples/valid/final-reconciliation-packet.json >/dev/null
jq -e '.schema == "ao.sentinel.public-safety-wording-readback.v0.1" and .status == "passed" and .unsafe_public_wording_found == false and .claims_authority_advance == false and .rsi_remains_denied == true and (.scanned_artifacts | index("docs/evidence/ao-mission-atlas-wave-import-v01/next-recommended-prompt.md")) and (.scanned_artifacts | index("examples/valid/final-reconciliation-packet.json"))' docs/evidence/ao-mission-atlas-wave-import-v01/sentinel-public-safety-scan.json >/dev/null
jq -e '.schema == "ao.mission.production-readiness-branch-cleanup.v0.1" and .status == "passed" and .mission == "ao-mission-atlas-wave-import-v01" and .completed_nodes_at_recording == 15 and .local_verification_passed == true and .github_ci_passed_through_previous_node == true and .stale_local_codex_branches_remaining == 0 and .stale_remote_codex_branches_remaining == 0 and .current_node_branch_cleanup_pending_pr_merge == true and .direct_main_mutation == false and .promotion_claimed == false and .rsi_remains_denied == true' docs/evidence/ao-mission-atlas-wave-import-v01/production-readiness-branch-cleanup.json >/dev/null
jq -e '.schema == "ao.promoter.no-promotion-readback.v0.1" and .status == "no_promotion_requested" and .mission_id == "ao-mission-atlas-wave-import-v01" and .completed_nodes_at_recording == 16 and .safe_to_promote == false and .promotion_claimed == false and .claims_authority_advance == false and .broad_RSI == "denied" and .rsi_remains_denied == true and .executes_work == false and .approves_work == false' docs/evidence/ao-mission-atlas-wave-import-v01/promoter-no-promotion-summary.json >/dev/null
jq -e '.schema == "ao.foundry.terminal-state-binding.v0.1" and (.states | length == 4) and ([.states[].status] | index("completed") and index("promoted") and index("denied") and index("blocked")) and ([.states[] | select(.status == "completed" or .status == "promoted") | .expected_mission_status] | all(. == "done")) and ([.states[] | select(.status == "denied" or .status == "blocked") | .expected_mission_status] | all(. == "blocked")) and .safe_to_execute == false and .executes_work == false and .approves_work == false and .rsi_remains_denied == true' examples/valid/foundry-terminal-state-binding.json >/dev/null
jq -e '.schema == "ao.command.compact-timeline-readback.v0.1" and .status == "ready" and .compact == true and (.includes_event_kinds | index("atlas_recommendation")) and (.includes_event_kinds | index("final_reconciliation")) and ([.recent_events[].kind] | index("atlas_recommendation") and index("final_reconciliation")) and .safe_to_execute == false and .executes_work == false and .approves_work == false and .mutates_repositories == false and .rsi_remains_denied == true' examples/valid/command-compact-timeline-readback.json >/dev/null
echo "AO Mission production readiness: 100/100 status=ready"
