#!/usr/bin/env python3
"""Tests for the dependency-free production-readiness JSON checks."""

import importlib.util
import json
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
HELPER = ROOT / "scripts" / "production_readiness_json.py"
READINESS = ROOT / "scripts" / "production-readiness.sh"

EXPECTED_PROFILES = {
    "atlas_recommendation_import",
    "atlas_recommendation_inspect",
    "final_reconciliation_runtime",
    "timeline_query_index",
    "restart_recovery_proof",
    "event_search_runtime",
    "atlas_continuation_prompt",
    "atlas_wave_synthesis_runtime",
    "atlas_final_synthesis_import",
    "atlas_final_synthesis_inspect",
    "checkpoint_resume_bundle",
    "doctor_runtime",
    "final_reconciliation_fixture",
    "final_reconciliation_mismatch_fixture",
    "final_rollup_ready_node_denial",
    "sentinel_public_safety_scan",
    "production_readiness_branch_cleanup",
    "promoter_no_promotion_summary",
    "foundry_terminal_state_binding",
    "command_compact_timeline",
    "mission_status_timeline_vector",
    "command_status_lease_checkpoint",
    "doctor_command_compact_risk",
    "beta_incident_stop_rule",
    "pilot_feedback_capture",
    "final_reconciliation_event_search",
    "promoter_no_promotion_node",
    "sentinel_public_safety_node",
    "wave_boundary_readiness",
    "merged_pr_branch_cleanup",
    "atlas_wave_final_synthesis_fixture",
    "post_merge_final_closure",
    "wave_duration_ledger",
    "codex_session_duration",
    "atlas_final_synthesis_fixture",
    "event_search_production_smoke",
    "event_evidence_alias_readback",
    "event_evidence_alias_searches",
    "bounded_autonomy_month3",
    "bounded_autonomy_month4",
    "bounded_autonomy_month5",
    "bounded_autonomy_month6",
    "bounded_autonomy_repair",
    "sqlite_migration_dry_run",
}

STATIC_CASES = {
    "final_reconciliation_fixture": "examples/valid/final-reconciliation-packet.json",
    "final_reconciliation_mismatch_fixture": "examples/valid/final-reconciliation-mismatch-packet.json",
    "final_rollup_ready_node_denial": "examples/valid/final-rollup-ready-node-denial.json",
    "sentinel_public_safety_scan": "docs/evidence/ao-mission-atlas-wave-import-v01/sentinel-public-safety-scan.json",
    "production_readiness_branch_cleanup": "docs/evidence/ao-mission-atlas-wave-import-v01/production-readiness-branch-cleanup.json",
    "promoter_no_promotion_summary": "docs/evidence/ao-mission-atlas-wave-import-v01/promoter-no-promotion-summary.json",
    "foundry_terminal_state_binding": "examples/valid/foundry-terminal-state-binding.json",
    "command_compact_timeline": "examples/valid/command-compact-timeline-readback.json",
    "mission_status_timeline_vector": "examples/valid/mission-status-timeline-compatibility-vector.json",
    "command_status_lease_checkpoint": "examples/valid/command-status-lease-checkpoint-readback.json",
    "doctor_command_compact_risk": "examples/valid/doctor-command-compact-early-return-risk.json",
    "beta_incident_stop_rule": "examples/valid/beta-incident-stop-rule-readback.json",
    "pilot_feedback_capture": "examples/valid/pilot-feedback-capture-packet.json",
    "final_reconciliation_event_search": "examples/valid/final-reconciliation-event-search-readback.json",
    "wave_boundary_readiness": "docs/evidence/ao-mission-atlas-wave-import-v01/wave-boundary-readiness.json",
    "merged_pr_branch_cleanup": "docs/evidence/ao-mission-atlas-wave-import-v01/merged-pr-branch-cleanup.json",
    "atlas_wave_final_synthesis_fixture": "docs/evidence/ao-mission-atlas-wave-import-v01/final-synthesis.json",
    "post_merge_final_closure": "docs/evidence/ao-mission-atlas-wave-import-v01/post-merge-final-closure.json",
    "wave_duration_ledger": "docs/evidence/ao-mission-doubled-wave-v01/duration-ledger.json",
    "codex_session_duration": "docs/evidence/ao-mission-doubled-wave-v01/codex-session-duration-readback.json",
    "atlas_final_synthesis_fixture": "examples/valid/atlas-final-synthesis-readback.json",
    "event_search_production_smoke": "docs/evidence/ao-mission-atlas-wave-import-v01/event-search-production-smoke.json",
    "event_evidence_alias_readback": "docs/evidence/ao-mission-doubled-wave-v01/nodes/node-10-event-evidence-aliases/event-alias-search-readbacks.json",
    "event_evidence_alias_searches": "examples/valid/event-evidence-alias-search-readbacks.json",
    "bounded_autonomy_month3": "examples/valid/bounded-autonomy-month3-recovery-readback.json",
    "bounded_autonomy_month4": "examples/valid/bounded-autonomy-month4-controlled-improvement-readback.json",
    "bounded_autonomy_month5": "examples/valid/bounded-autonomy-month5-dogfood-readback.json",
    "bounded_autonomy_month6": "examples/valid/bounded-autonomy-month6-qualification-readback.json",
    "bounded_autonomy_repair": "examples/valid/bounded-autonomy-repair-from-month3-readback.json",
    "sqlite_migration_dry_run": "examples/valid/mission-sqlite-migration-dry-run.json",
}


def run_helper(*args):
    return subprocess.run(
        [sys.executable, str(HELPER), *map(str, args)],
        cwd=ROOT,
        capture_output=True,
        text=True,
        encoding="utf-8",
        check=False,
    )


class ProductionReadinessJSONTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        spec = importlib.util.spec_from_file_location("production_readiness_json", HELPER)
        if spec is None or spec.loader is None:
            raise AssertionError("could not load production readiness JSON helper")
        cls.helper = importlib.util.module_from_spec(spec)
        spec.loader.exec_module(cls.helper)

    def write_json(self, directory, name, value):
        path = Path(directory) / name
        path.write_text(json.dumps(value), encoding="utf-8")
        return path

    def test_named_profiles_cover_every_readiness_predicate(self):
        self.assertEqual(set(self.helper.CHECKS), EXPECTED_PROFILES)

        body = READINESS.read_text(encoding="utf-8")
        used = set()
        for line in body.splitlines():
            words = line.split()
            if words[:1] == ["json_check"]:
                used.add(words[1])
            elif "check-tree" in words:
                used.add(words[words.index("check-tree") + 1])
        self.assertEqual(used, EXPECTED_PROFILES)
        self.assertEqual(body.count("extract-mission-id"), 3)
        self.assertEqual(body.count("bind-mission-id"), 1)

    def test_every_static_profile_accepts_its_repository_fixture(self):
        self.assertEqual(len(STATIC_CASES), 30)
        for profile, path in STATIC_CASES.items():
            with self.subTest(profile=profile):
                result = run_helper("check", profile, ROOT / path)
                self.assertEqual(result.returncode, 0, result.stderr)

    def test_extract_mission_id_uses_strict_json(self):
        with tempfile.TemporaryDirectory() as directory:
            valid = self.write_json(directory, "valid.json", {"mission_id": "mission-123"})
            result = run_helper("extract-mission-id", valid)
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(result.stdout, "mission-123\n")

            duplicate = Path(directory) / "duplicate.json"
            duplicate.write_text('{"mission_id":"a","mission_id":"b"}', encoding="utf-8")
            result = run_helper("extract-mission-id", duplicate)
            self.assertEqual(result.returncode, 2)
            self.assertIn("duplicate JSON key: mission_id", result.stderr)

            malformed = Path(directory) / "malformed.json"
            malformed.write_text("{", encoding="utf-8")
            result = run_helper("extract-mission-id", malformed)
            self.assertEqual(result.returncode, 2)
            self.assertIn("invalid JSON", result.stderr)

            missing = self.write_json(directory, "missing.json", {})
            result = run_helper("extract-mission-id", missing)
            self.assertEqual(result.returncode, 2)
            self.assertIn("mission_id must be a non-empty string", result.stderr)

    def test_bind_mission_id_preserves_document_and_writes_utf8(self):
        with tempfile.TemporaryDirectory() as directory:
            source = self.write_json(directory, "source.json", {"mission_id": "old", "label": "caf\u00e9"})
            destination = Path(directory) / "bound.json"
            result = run_helper("bind-mission-id", source, destination, "mission-new")
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(
                json.loads(destination.read_text(encoding="utf-8")),
                {"mission_id": "mission-new", "label": "caf\u00e9"},
            )

    def test_checks_fail_closed_for_malformed_missing_and_wrong_types(self):
        with tempfile.TemporaryDirectory() as directory:
            malformed = Path(directory) / "malformed.json"
            malformed.write_text("{", encoding="utf-8")
            for path in (malformed, self.write_json(directory, "missing.json", {})):
                result = run_helper("check", "final_reconciliation_fixture", path)
                self.assertEqual(result.returncode, 2)
                self.assertIn("production-readiness JSON error:", result.stderr)

            wrong_type = self.write_json(
                directory,
                "wrong-type.json",
                {
                    "schema": "ao.mission.final-reconciliation-packet.v0.1",
                    "status": "ready",
                    "artifacts_agree": "true",
                    "promotion_claimed": False,
                    "rsi_remains_denied": True,
                    "claims_authority_advance": False,
                    "safe_to_execute": False,
                    "executes_work": False,
                    "approves_work": False,
                },
            )
            result = run_helper("check", "final_reconciliation_fixture", wrong_type)
            self.assertEqual(result.returncode, 2)
            self.assertIn("artifacts_agree", result.stderr)

    def test_batch_check_validates_every_matching_file(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            good = {
                "promotion_claimed": False,
            }
            self.write_json(root, "promoter-no-promotion.json", good)
            nested = root / "nested"
            nested.mkdir()
            self.write_json(nested, "promoter-no-promotion.json", good)
            result = run_helper(
                "check-tree", "promoter_no_promotion_node", root, "promoter-no-promotion.json"
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            self.write_json(nested, "promoter-no-promotion.json", {"promotion_claimed": True})
            result = run_helper(
                "check-tree", "promoter_no_promotion_node", root, "promoter-no-promotion.json"
            )
            self.assertEqual(result.returncode, 2)
            self.assertIn("nested", result.stderr)

    def test_readiness_script_is_jq_free_cleanup_safe_and_read_only_formatting(self):
        body = READINESS.read_text(encoding="utf-8")
        executable = "\n".join(
            line for line in body.splitlines() if line.strip() and not line.lstrip().startswith("#")
        )
        self.assertNotRegex(executable, r"(^|[|;&( ])jq([ )]|$)")
        self.assertNotIn("gofmt -w", executable)
        self.assertIn("PYTHONDONTWRITEBYTECODE=1", executable)
        self.assertRegex(executable, r"trap .*EXIT")
        self.assertNotRegex(executable, r"go build ./cmd/ao-mission")
        self.assertRegex(executable, r"go build -o \"\$tmp_home/")


if __name__ == "__main__":
    unittest.main()
