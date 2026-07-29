#!/usr/bin/env python3

import subprocess
import sys
import tempfile
import unittest
from pathlib import Path


REPOSITORY_ROOT = Path(__file__).resolve().parents[1]
SCANNER = REPOSITORY_ROOT / "scripts" / "public-safety-scan.py"


class PublicSafetyScanTests(unittest.TestCase):
    def run_scan(self, root: Path, *paths: str) -> subprocess.CompletedProcess[str]:
        selected = paths or (".",)
        return subprocess.run(
            [
                sys.executable,
                str(SCANNER),
                "--root",
                str(root),
                *selected,
            ],
            check=False,
            capture_output=True,
            text=True,
        )

    def write(self, root: Path, relative: str, body: str) -> Path:
        path = root / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(body, encoding="utf-8")
        return path

    def test_allows_code_assignments_regex_fragments_and_example_paths(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.write(
                root,
                "safe.go",
                "\n".join(
                    [
                        'token = strings.ReplaceAll(token, "~", "~0")',
                        "token = loadTokenFromEnvironment()",
                        "token = otherIdentifier",
                        r"gitHubToken := `gh[pousr]_` + `[A-Za-z0-9]{20,}`",
                        'fixture := "/Users/example/evidence.json"',
                    ]
                ),
            )

            result = self.run_scan(root)

            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertEqual(result.stdout, "")
            self.assertEqual(result.stderr, "")

    def test_rejects_private_non_example_user_path(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            private_path = "/Users/" + "operator/private/evidence.json"
            self.write(root, "unsafe.txt", private_path)

            result = self.run_scan(root)

            self.assertEqual(result.returncode, 1)
            self.assertIn("local_user_path unsafe.txt:1", result.stderr)
            self.assertNotIn(private_path, result.stderr)

    def test_rejects_literal_secret_assignments_without_echoing_values(self) -> None:
        for field in ("token", "secret", "api_key", "password"):
            with self.subTest(field=field), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                secret_value = "synthetic-" + "value-938475"
                assignment = field + ' = "' + secret_value + '"'
                self.write(root, "unsafe.env", assignment)

                result = self.run_scan(root)

                self.assertEqual(result.returncode, 1)
                self.assertIn(
                    "literal_secret_assignment unsafe.env:1",
                    result.stderr,
                )
                self.assertNotIn(secret_value, result.stderr)

    def test_rejects_private_key_openai_and_github_patterns(self) -> None:
        cases = [
            (
                "private_key",
                "-----BEGIN " + "PRIVATE KEY-----",
            ),
            (
                "openai_credential",
                "sk-" + "AbCdEf0123456789GhIjKlMn",
            ),
            (
                "github_credential",
                "ghp_" + "AbCdEf0123456789GhIjKlMn",
            ),
        ]
        for category, body in cases:
            with self.subTest(category=category), tempfile.TemporaryDirectory() as directory:
                root = Path(directory)
                self.write(root, "unsafe.txt", body)

                result = self.run_scan(root)

                self.assertEqual(result.returncode, 1)
                self.assertIn(f"{category} unsafe.txt:1", result.stderr)
                self.assertNotIn(body, result.stderr)

    def test_rejects_symlinks_and_out_of_root_inputs(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory) / "root"
            outside = Path(directory) / "outside.txt"
            root.mkdir()
            outside.write_text("public fixture", encoding="utf-8")
            (root / "escape").symlink_to(outside)

            symlink_result = self.run_scan(root)
            outside_result = self.run_scan(root, str(outside))

            self.assertEqual(symlink_result.returncode, 1)
            self.assertIn("symlink escape", symlink_result.stderr)
            self.assertEqual(outside_result.returncode, 2)
            self.assertIn("outside scan root", outside_result.stderr)

    def test_output_is_deterministic(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            private_path = "/Users/" + "operator/private/evidence.json"
            self.write(root, "b.txt", private_path)
            self.write(root, "a.txt", private_path)

            first = self.run_scan(root)
            second = self.run_scan(root)

            self.assertEqual(first.returncode, 1)
            self.assertEqual(first.stderr, second.stderr)
            self.assertLess(first.stderr.index("a.txt:1"), first.stderr.index("b.txt:1"))

    def test_rejects_oversized_input(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            self.write(root, "oversized.txt", "x" * (1024 * 1024 + 1))

            result = self.run_scan(root)

            self.assertEqual(result.returncode, 2)
            self.assertIn("exceeds per-file byte limit", result.stderr)

    def test_scanner_source_passes_its_own_rules(self) -> None:
        result = self.run_scan(
            REPOSITORY_ROOT,
            "scripts/public-safety-scan.py",
        )

        self.assertEqual(result.returncode, 0, result.stderr)


if __name__ == "__main__":
    unittest.main()
