"""scripts/proto_check.py, buf's checks, on a copy of the repository broken one way
at a time: the contracts as committed pass, and a breaking change, a .proto edited
without regenerating, a message without a comment (those-who-run/paddock#23's
acceptance criteria), a 64-bit integer field, and a released message removed
without a line in removed.txt each fail the matching check. It runs buf and
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
PROTO = Path("proto/twr/events/v1/lap.proto")
COPIED = ["go.mod", "go.sum", "buf.yaml", "buf.gen.yaml", "removed.txt", "proto", "gen", "tools"]
# A message only the baseline has, for the removed check.
EXTRA = """syntax = "proto3";

// A package the baseline released.
package twr.extra.v1;

// Extra is a released message.
message Extra {
  // A field.
  int32 n = 1;
}
"""
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
        for check in ("format", "lint", "fresh", "wide"):
            with self.subTest(check=check):
                code, out = self.check(check)
                self.assertEqual(code, 0, out)
        for check in ("breaking", "removed"):
            with self.subTest(check=check):
                code, out = self.check(check, "--against", str(self.before))
                self.assertEqual(code, 0, out)

    def test_renaming_a_field_fails_breaking(self):
        self.edit("int32 lap_time_ms = 3;", "int32 lap_ms = 3;")
        code, out = self.check("breaking", "--against", str(self.before))
        self.assertEqual(code, 1, out)
        self.assertIn('Field "3" with name "lap_ms" on message "LapCompleted" changed', out)

    def test_editing_a_proto_without_regenerating_fails_fresh(self):
        self.edit(
            "  uint32 last_seq = 11;\n",
            "  uint32 last_seq = 11;\n  // The stint the lap was in.\n  int32 stint = 12;\n",
        )
        code, out = self.check("fresh")
        self.assertEqual(code, 1, out)
        self.assertIn("gen/ is not what buf generates from proto/", out)
        self.assertIn("twr/events/v1/lap.pb.go", out)

    def test_a_message_without_a_comment_fails_lint(self):
        self.edit(
            "// LapCompleted says a lap ended: the lap counter moved on, or the session ended\n"
            "// part way through it.\n",
            "",
        )
        code, out = self.check("lint")
        self.assertEqual(code, 1, out)
        self.assertIn('Message "LapCompleted" should have a non-empty comment', out)

    def test_a_64_bit_integer_fails_wide(self):
        for wide in ("int64", "uint64", "sint64", "fixed64", "sfixed64"):
            with self.subTest(type=wide):
                field = f"  // A wide field.\n  {wide} big = 12;\n"
                self.edit("  uint32 last_seq = 11;\n", "  uint32 last_seq = 11;\n" + field)
                code, out = self.check("wide")
                self.assertEqual(code, 1, out)
                self.assertIn(f"twr.events.v1.LapCompleted.big ({wide})", out)
                self.edit(field, "")

    def test_a_wrapper_of_a_64_bit_integer_fails_wide(self):
        self.edit('package twr.events.v1;\n', 'package twr.events.v1;\n\nimport "google/protobuf/wrappers.proto";\n')
        for wrapper in ("Int64Value", "UInt64Value"):
            with self.subTest(type=wrapper):
                field = f"  // A wrapped wide field.\n  google.protobuf.{wrapper} big = 12;\n"
                self.edit("  uint32 last_seq = 11;\n", "  uint32 last_seq = 11;\n" + field)
                code, out = self.check("wide")
                self.assertEqual(code, 1, out)
                self.assertIn(f"twr.events.v1.LapCompleted.big (google.protobuf.{wrapper})", out)
                self.edit(field, "")

    def test_a_removal_cannot_be_listed_in_advance(self):
        with (self.repo / "removed.txt").open("a", encoding="utf-8") as f:
            f.write("twr.events.v1.LapCompleted Going in the next change.\n")
        code, out = self.check("removed", "--against", str(self.before))
        self.assertEqual(code, 1, out)
        self.assertIn("lists a name that is still here", out)
        self.assertIn("twr.events.v1.LapCompleted", out)

    def test_removing_a_released_message_needs_a_reason(self):
        extra = self.before / "proto" / "twr" / "extra" / "v1"
        extra.mkdir(parents=True)
        (extra / "extra.proto").write_text(EXTRA, encoding="utf-8")
        code, out = self.check("removed", "--against", str(self.before))
        self.assertEqual(code, 1, out)
        self.assertIn("without a line in removed.txt saying why: twr.extra.v1.Extra", out)
        # A name with no reason is not enough.
        with (self.repo / "removed.txt").open("a", encoding="utf-8") as f:
            f.write("twr.extra.v1.Extra\n")
        self.assertEqual(self.check("removed", "--against", str(self.before))[0], 1)
        with (self.repo / "removed.txt").open("a", encoding="utf-8") as f:
            f.write("twr.extra.v1.Extra Replaced by twr.laps.v1.Extra.\n")
        code, out = self.check("removed", "--against", str(self.before))
        self.assertEqual(code, 0, out)
        self.assertIn("Removed on purpose: twr.extra.v1.Extra: Replaced by twr.laps.v1.Extra.", out)


if __name__ == "__main__":
    unittest.main()
