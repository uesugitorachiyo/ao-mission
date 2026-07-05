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
event_index_json="$(mktemp)"
event_search_json="$(mktemp)"
atlas_prompt_json="$(mktemp)"
synthesis_json="$(mktemp)"
doctor_json="$(mktemp)"
final_synthesis_import_json="$(mktemp)"
final_synthesis_inspect_json="$(mktemp)"
final_synthesis_checkpoint_json="$(mktemp)"
./ao-mission --home "$tmp_home" start "import completed Atlas recommendation wave" >"$mission_json"
mission_id="$(jq -r '.mission_id' "$mission_json")"
./ao-mission --home "$tmp_home" import atlas-recommendation-readback --mission "$mission_id" --path examples/valid/atlas-recommendation-readback.json >"$import_json"
jq -e '.kind == "atlas-recommendation-readback" and .safe_to_execute == false and .executes_work == false and .approves_work == false' "$import_json" >/dev/null
./ao-mission --home "$tmp_home" mission inspect --mission "$mission_id" --json >"$inspect_json"
jq -e '.status == "done" and .current_route == "complete" and .current_phase == "complete" and .evidence.atlas_recommendation.completed_nodes == 40 and .return_gate.final_response_allowed == true' "$inspect_json" >/dev/null
./ao-mission --home "$tmp_home" final reconcile --mission "$mission_id" >"$reconcile_json"
jq -e '.schema == "ao.mission.final-reconciliation-packet.v0.1" and .artifacts_agree == true and .final_response_allowed == true and .claims_authority_advance == false and .rsi_remains_denied == true' "$reconcile_json" >/dev/null
./ao-mission --home "$tmp_home" mission events index --out "$event_index_json" >/dev/null
./ao-mission --home "$tmp_home" mission events search --mission "$mission_id" --kind final_reconciliation --index "$event_index_json" --json >"$event_search_json"
jq -e '.schema == "ao.mission.event-search-readback.v0.1" and .status == "ready" and .total_matches >= 1 and .events[0].kind == "final_reconciliation" and .safe_to_execute == false and .executes_work == false and .approves_work == false' "$event_search_json" >/dev/null
./ao-mission --home "$tmp_home" final atlas-prompt --mission "$mission_id" --event-index "$event_index_json" --out "$atlas_prompt_json" >/dev/null
jq -e '.schema == "ao.mission.atlas-continuation-prompt-packet.v0.1" and .status == "ready" and .mission_id == "'"$mission_id"'" and (.event_index_digest | test("^sha256:[0-9a-f]{64}$")) and (.final_rollup_digest | test("^sha256:[0-9a-f]{64}$")) and (.prompt | contains("AO Atlas")) and (.prompt | contains("Do not produce a final response if ready_nodes > 0 or exact_next_action remains.")) and (.feature_depth_recommendations | length >= 10) and ([.feature_depth_recommendations[] | select((.gate | length > 0) and (.continuation_command | length > 0) and (.estimated_minutes >= 6) and (.evidence_required | length >= 3))] | length >= 10) and ([.feature_depth_recommendations[].estimated_minutes] | add >= 60) and .safe_to_execute == false and .executes_work == false and .approves_work == false and .mutates_repositories == false' "$atlas_prompt_json" >/dev/null
./ao-mission --home "$tmp_home" final synthesize --mission "$mission_id" --evidence-root docs/evidence/ao-mission-doubled-wave-v01 >"$synthesis_json"
jq -e '.schema == "ao.mission.atlas-wave-final-synthesis.v0.1" and .mission == "ao-mission-doubled-wave-v01" and .completed_nodes >= 10 and .ready_nodes >= 50 and .final_response_allowed == false and (.feature_depth_recommendations | length >= 20) and ([.feature_depth_recommendations[] | select((.gate | length > 0) and (.continuation_command | length > 0) and (.estimated_minutes >= 6) and (.evidence_required | length >= 3))] | length >= 20) and ([.feature_depth_recommendations[].estimated_minutes] | add >= 120) and .safe_to_execute == false and .executes_work == false and .approves_work == false and .rsi_remains_denied == true' "$synthesis_json" >/dev/null
./ao-mission --home "$tmp_home" start "import Atlas final synthesis readback" >"$mission_json"
final_synthesis_mission_id="$(jq -r '.mission_id' "$mission_json")"
./ao-mission --home "$tmp_home" import atlas-final-synthesis-readback --mission "$final_synthesis_mission_id" --path examples/valid/atlas-final-synthesis-readback.json >"$final_synthesis_import_json"
jq -e '.kind == "atlas-final-synthesis-readback" and .safe_to_execute == false and .executes_work == false and .approves_work == false' "$final_synthesis_import_json" >/dev/null
./ao-mission --home "$tmp_home" mission inspect --mission "$final_synthesis_mission_id" --json >"$final_synthesis_inspect_json"
jq -e '.status == "done" and .current_route == "complete" and .current_phase == "complete" and .evidence.atlas_final_synthesis.command_readback == "ready" and .evidence.atlas_final_synthesis.promoter_status == "no_promotion_requested" and .route_reconciliation.command_readback_bound == true and .route_reconciliation.promoter_readback_bound == true and .route_reconciliation.atlas_ready_nodes == 0 and .return_gate.final_response_allowed == true' "$final_synthesis_inspect_json" >/dev/null
cp "$tmp_home/missions/$final_synthesis_mission_id.checkpoint-resume-bundle.json" "$final_synthesis_checkpoint_json"
jq -e '.schema == "ao.mission.checkpoint-resume-bundle.v0.3" and .mission_id == "'"$final_synthesis_mission_id"'" and .status == "ready" and .return_gate.final_response_allowed == true and .safe_to_execute == false and .executes_work == false and .approves_work == false and .mutates_repositories == false' "$final_synthesis_checkpoint_json" >/dev/null
./ao-mission --home "$tmp_home" start "doctor active lease health smoke" >"$mission_json"
doctor_mission_id="$(jq -r '.mission_id' "$mission_json")"
./ao-mission --home "$tmp_home" continue --mission "$doctor_mission_id" --until-done --max-iterations 2 >/dev/null
./ao-mission --home "$tmp_home" doctor --json >"$doctor_json"
jq -e '.schema == "ao.mission.doctor-readback.v0.1" and .status == "ready" and .lease_health_status == "healthy" and .checkpoint_freshness_status == "fresh" and .early_return_risk_status == "risk_detected" and .stale_route_decision_status == "clear" and ([.risk_missions[].kind] | index("early_return")) and (.exact_next_action | length > 0) and .safe_to_execute == false and .executes_work == false and .approves_work == false and .mutates_repositories == false' "$doctor_json" >/dev/null
rm -rf "$tmp_home" "$mission_json" "$import_json" "$inspect_json" "$reconcile_json" "$event_index_json" "$event_search_json" "$atlas_prompt_json" "$synthesis_json" "$doctor_json" "$final_synthesis_import_json" "$final_synthesis_inspect_json" "$final_synthesis_checkpoint_json" ao-mission
grep -q "25-node Atlas recommendation import wave" docs/operator-next-actions.md
grep -q "Do not stop before 25 completed nodes" docs/evidence/ao-mission-atlas-wave-import-v01/next-recommended-prompt.md
grep -q "final-reconciliation-packet.json" docs/operator-next-actions.md
grep -q "Command and final reconciliation closure check" docs/operator-next-actions.md
jq -e '.schema == "ao.mission.final-reconciliation-packet.v0.1" and .status == "ready" and .artifacts_agree == true and .promotion_claimed == false and .rsi_remains_denied == true and .claims_authority_advance == false and .safe_to_execute == false and .executes_work == false and .approves_work == false' examples/valid/final-reconciliation-packet.json >/dev/null
jq -e '.schema == "ao.mission.final-reconciliation-packet.v0.1" and .status == "blocked" and .artifacts_agree == false and (.blocker | contains("Foundry completed_nodes=39")) and (.blocker | contains("Atlas completed_nodes=40")) and .promotion_claimed == false and .rsi_remains_denied == true and .claims_authority_advance == false and .safe_to_execute == false and .executes_work == false and .approves_work == false' examples/valid/final-reconciliation-mismatch-packet.json >/dev/null
jq -e '.schema == "ao.sentinel.public-safety-wording-readback.v0.1" and .status == "passed" and .unsafe_public_wording_found == false and .claims_authority_advance == false and .rsi_remains_denied == true and (.scanned_artifacts | index("docs/evidence/ao-mission-atlas-wave-import-v01/next-recommended-prompt.md")) and (.scanned_artifacts | index("examples/valid/final-reconciliation-packet.json"))' docs/evidence/ao-mission-atlas-wave-import-v01/sentinel-public-safety-scan.json >/dev/null
jq -e '.schema == "ao.mission.production-readiness-branch-cleanup.v0.1" and .status == "passed" and .mission == "ao-mission-atlas-wave-import-v01" and .completed_nodes_at_recording == 15 and .local_verification_passed == true and .github_ci_passed_through_previous_node == true and .stale_local_codex_branches_remaining == 0 and .stale_remote_codex_branches_remaining == 0 and .current_node_branch_cleanup_pending_pr_merge == true and .direct_main_mutation == false and .promotion_claimed == false and .rsi_remains_denied == true' docs/evidence/ao-mission-atlas-wave-import-v01/production-readiness-branch-cleanup.json >/dev/null
jq -e '.schema == "ao.promoter.no-promotion-readback.v0.1" and .status == "no_promotion_requested" and .mission_id == "ao-mission-atlas-wave-import-v01" and .completed_nodes_at_recording == 16 and .safe_to_promote == false and .promotion_claimed == false and .claims_authority_advance == false and .broad_RSI == "denied" and .rsi_remains_denied == true and .executes_work == false and .approves_work == false' docs/evidence/ao-mission-atlas-wave-import-v01/promoter-no-promotion-summary.json >/dev/null
jq -e '.schema == "ao.foundry.terminal-state-binding.v0.1" and (.states | length == 4) and ([.states[].status] | index("completed") and index("promoted") and index("denied") and index("blocked")) and ([.states[] | select(.status == "completed" or .status == "promoted") | .expected_mission_status] | all(. == "done")) and ([.states[] | select(.status == "denied" or .status == "blocked") | .expected_mission_status] | all(. == "blocked")) and .safe_to_execute == false and .executes_work == false and .approves_work == false and .rsi_remains_denied == true' examples/valid/foundry-terminal-state-binding.json >/dev/null
jq -e '.schema == "ao.command.compact-timeline-readback.v0.1" and .status == "ready" and .compact == true and (.includes_event_kinds | index("atlas_recommendation")) and (.includes_event_kinds | index("final_reconciliation")) and ([.recent_events[].kind] | index("atlas_recommendation") and index("final_reconciliation")) and .safe_to_execute == false and .executes_work == false and .approves_work == false and .mutates_repositories == false and .rsi_remains_denied == true' examples/valid/command-compact-timeline-readback.json >/dev/null
jq -e '.schema == "ao.mission.event-search-readback.v0.1" and .status == "ready" and .kind == "final_reconciliation" and .total_matches == 1 and .events[0].kind == "final_reconciliation" and (.events[0].summary | contains("artifacts_agree=true")) and (.events[0].summary | contains("rsi_remains_denied=true")) and .safe_to_execute == false and .executes_work == false and .approves_work == false and .mutates_repositories == false' examples/valid/final-reconciliation-event-search-readback.json >/dev/null
find docs/evidence/ao-mission-atlas-wave-import-v01/nodes -name promoter-no-promotion.json -print0 | xargs -0 -n1 jq -e '.promotion_claimed == false' >/dev/null
find docs/evidence/ao-mission-atlas-wave-import-v01/nodes -name sentinel-public-safety.json -print0 | xargs -0 -n1 jq -e '.claims_authority_advance == false and .rsi_remains_denied == true' >/dev/null
jq -e '.schema == "ao.mission.wave-boundary-readiness.v0.1" and .status == "passed" and .mission == "ao-mission-atlas-wave-import-v01" and .completed_nodes_at_recording == 23 and .promoter_no_promotion_records >= 23 and .sentinel_public_safety_records >= 23 and .promotion_claimed == false and .claims_authority_advance == false and .rsi_remains_denied == true' docs/evidence/ao-mission-atlas-wave-import-v01/wave-boundary-readiness.json >/dev/null
jq -e '.schema == "ao.mission.merged-pr-branch-cleanup.v0.1" and .status == "passed" and .mission == "ao-mission-atlas-wave-import-v01" and .completed_nodes_through_previous_node == 23 and (.merged_prs | length == 23) and (.merged_prs[0] == 21) and (.merged_prs[-1] == 43) and .stale_local_codex_branches_remaining == 0 and .stale_remote_codex_branches_remaining == 0 and .current_node_branch_cleanup_pending_pr_merge == true and .direct_main_mutation == false and .rsi_remains_denied == true' docs/evidence/ao-mission-atlas-wave-import-v01/merged-pr-branch-cleanup.json >/dev/null
jq -e '.schema == "ao.mission.atlas-wave-final-synthesis.v0.1" and .status == "completed" and .mission == "ao-mission-atlas-wave-import-v01" and .completed_nodes >= 25 and .ready_nodes == 0 and .blocked_nodes == 0 and .final_response_allowed == true and .current_node_pr_pending == false and .promotion_claimed == false and .claims_authority_advance == false and .rsi_remains_denied == true and (.feature_depth_recommendations | length >= 10)' docs/evidence/ao-mission-atlas-wave-import-v01/final-synthesis.json >/dev/null
jq -e '.schema == "ao.mission.post-merge-final-closure.v0.1" and .status == "completed" and .mission == "ao-mission-atlas-wave-import-v01" and .completed_nodes == 26 and .ready_nodes == 0 and .blocked_nodes == 0 and (.merged_prs | length == 25) and (.merged_prs[0] == 21) and (.merged_prs[-1] == 45) and .stale_local_codex_branches_remaining == 0 and .stale_remote_codex_branches_remaining == 0 and .final_response_allowed == true and .rsi_remains_denied == true' docs/evidence/ao-mission-atlas-wave-import-v01/post-merge-final-closure.json >/dev/null
grep -q "Do not stop before 30 completed nodes" docs/evidence/ao-mission-atlas-wave-import-v01/next-wave-recommended-prompt.md
jq -e '.schema == "ao.mission.wave-duration-ledger.v0.1" and .mission == "ao-mission-doubled-wave-v01" and .status == "active" and .minimum_minutes == 120 and .target_minutes == 120 and .max_minutes == 180 and .minimum_minutes_met == false and .final_response_allowed == false and .rsi_remains_denied == true and .safe_to_execute == false and .executes_work == false and .approves_work == false' docs/evidence/ao-mission-doubled-wave-v01/duration-ledger.json >/dev/null
jq -e '.schema == "ao.mission.codex-session-duration-readback.v0.1" and .mission == "ao-mission-doubled-wave-v01" and .status == "metadata_available" and .content_read == false and .secret_values_read == false and .session_log_files_found > 0 and .safe_to_execute == false and .executes_work == false and .approves_work == false and .rsi_remains_denied == true' docs/evidence/ao-mission-doubled-wave-v01/codex-session-duration-readback.json >/dev/null
jq -e '.contract_version == "ao.atlas.ao-mission-final-synthesis-readback.v0.1" and .completed_nodes == 26 and .ready_nodes == 0 and .blocked_nodes == 0 and .final_response_allowed == true and .command_readback == "ready" and .promoter_status == "no_promotion_requested" and .rsi_remains_denied == true and .safe_to_execute == false and .executes_work == false and .approves_work == false' examples/valid/atlas-final-synthesis-readback.json >/dev/null
jq -e '.schema == "ao.mission.event-search-production-smoke.v0.1" and .status == "passed" and .mission == "ao-mission-atlas-wave-import-v01" and .searched_kind == "final_reconciliation" and .total_matches_minimum == 1 and .safe_to_execute == false and .executes_work == false and .approves_work == false and .rsi_remains_denied == true' docs/evidence/ao-mission-atlas-wave-import-v01/event-search-production-smoke.json >/dev/null
jq -e '.schema == "ao.mission.event-evidence-alias-readback.v0.1" and .status == "passed" and .mission == "ao-mission-doubled-wave-v01" and (.event_kinds | length == 6) and (.event_kinds | index("route_evidence") and index("node_evidence") and index("pr_evidence") and index("ci_evidence") and index("rollup_evidence") and index("blocker_evidence")) and .safe_to_execute == false and .executes_work == false and .approves_work == false and .rsi_remains_denied == true' docs/evidence/ao-mission-doubled-wave-v01/nodes/node-10-event-evidence-aliases/event-alias-search-readbacks.json >/dev/null
echo "AO Mission production readiness: 100/100 status=ready"
