"""scripts/proto_check.py, buf's checks, on a copy of the repository broken one way
at a time: the contracts as committed pass, and a breaking change, a .proto edited
without regenerating, and a message without a comment each fail the matching
check (those-who-run/paddock#23's acceptance criteria). It runs buf and
protoc-gen-go through `go tool`, so it needs Go, as `make test` does. Standard
library only."""

from __future__ import annotations

import os
import shutil
import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SCRIPT = ROOT / "scripts" / "proto_check.py"
PROTO = Path("proto/twr/placeholder/v1/placeholder.proto")
COPIED = ["go.mod", "go.sum", "buf.yaml", "buf.gen.yaml", "proto", "gen", "tools"]
ENV = {k: v for k, v in os.environ.items() if not k.startswith("GIT_")}


class ProtoCheck(unittest.TestCase):
    def setUp(self):
        tmp = tempfile.TemporaryDirectory()
        self.addCleanup(tmp.cleanup)
        self.before = Path(tmp.name) / "before"
        self.repo = Path(tmp.name) / "repo"
        for dst in (self.before, self.repo):
            dst.mkdir()
            for name in COPIED:
                src = ROOT / name
                if src.is_dir():
                    shutil.copytree(src, dst / name)
                else:
                    shutil.copy(src, dst / name)

    def check(self, *args: str) -> tuple[int, str]:
        proc = subprocess.run(
            [sys.executable, str(SCRIPT), *args],
            cwd=self.repo,
            env=ENV,
            capture_output=True,
            text=True,
        )
        return proc.returncode, proc.stdout + proc.stderr

    def edit(self, old: str, new: str) -> None:
        path = self.repo / PROTO
        text = path.read_text(encoding="utf-8")
        self.assertIn(old, text)
        path.write_text(text.replace(old, new, 1), encoding="utf-8")

    def test_the_committed_contracts_pass(self):
        for check in ("format", "lint", "fresh"):
            with self.subTest(check=check):
                code, out = self.check(check)
                self.assertEqual(code, 0, out)
        code, out = self.check("breaking", "--against", str(self.before))
        self.assertEqual(code, 0, out)

    def test_renaming_a_field_fails_breaking(self):
        self.edit("string note = 1;", "string remark = 1;")
        code, out = self.check("breaking", "--against", str(self.before))
        self.assertEqual(code, 1, out)
        self.assertIn('Field "1" with name "remark" on message "Placeholder" changed', out)

    def test_editing_a_proto_without_regenerating_fails_fresh(self):
        self.edit(
            "  string note = 1;\n",
            "  string note = 1;\n  // The lap it is about.\n  int32 lap = 2;\n",
        )
        code, out = self.check("fresh")
        self.assertEqual(code, 1, out)
        self.assertIn("gen/ is not what buf generates from proto/", out)
        self.assertIn("twr/placeholder/v1/placeholder.pb.go", out)

    def test_a_message_without_a_comment_fails_lint(self):
        self.edit("// Placeholder stands in for the first real message.\n", "")
        code, out = self.check("lint")
        self.assertEqual(code, 1, out)
        self.assertIn('Message "Placeholder" should have a non-empty comment', out)


if __name__ == "__main__":
    unittest.main()
