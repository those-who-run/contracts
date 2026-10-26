"""How scripts/proto_check.py picks what `buf breaking` compares with: the highest
`vX.Y.Z` tag by version; with no tag, `main` once it has contracts; and nothing,
said in the log, while it has none. Each case is a local bare repository, so it
needs git and nothing else. Standard library only."""

from __future__ import annotations

import importlib.util
import os
import subprocess
import tempfile
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SPEC = importlib.util.spec_from_file_location("proto_check", ROOT / "scripts" / "proto_check.py")
assert SPEC and SPEC.loader
proto_check = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(proto_check)

GIT_ENV = {
    **{k: v for k, v in os.environ.items() if not k.startswith("GIT_")},
    "GIT_AUTHOR_NAME": "test",
    "GIT_AUTHOR_EMAIL": "test@example.com",
    "GIT_COMMITTER_NAME": "test",
    "GIT_COMMITTER_EMAIL": "test@example.com",
    "GIT_CONFIG_GLOBAL": os.devnull,
    "GIT_CONFIG_NOSYSTEM": "1",
}


class Baseline(unittest.TestCase):
    def setUp(self):
        tmp = tempfile.TemporaryDirectory()
        self.addCleanup(tmp.cleanup)
        self.work = Path(tmp.name) / "work"
        self.remote = str(Path(tmp.name) / "remote.git")
        self.git("init", "-q", "--bare", self.remote, cwd=tmp.name)
        self.git("init", "-q", "-b", "main", str(self.work), cwd=tmp.name)

    def git(self, *args: str, cwd: str | Path | None = None) -> None:
        subprocess.run(["git", *args], cwd=cwd or self.work, env=GIT_ENV, check=True, capture_output=True)

    def commit(self, name: str) -> None:
        (self.work / name).write_text(name + "\n", encoding="utf-8")
        self.git("add", name)
        self.git("commit", "-q", "-m", f"add {name}")

    def push(self, *refs: str) -> None:
        self.git("push", "-q", self.remote, "main", *refs)

    def test_the_highest_release_tag_by_version(self):
        self.commit("buf.yaml")
        # By version, not by name: v0.10.0 is later than v0.9.0. A pre-release
        # or another name is not a release.
        for tag in ("v0.2.0", "v0.9.0", "v0.10.0", "v1.0.0-rc.1", "vnext"):
            self.git("tag", tag)
        self.push("--tags")
        against, line = proto_check.baseline(self.remote)
        self.assertEqual(against, f"{self.remote}#tag=v0.10.0")
        self.assertIn("the latest release, v0.10.0", line)

    def test_main_before_the_first_tag(self):
        self.commit("buf.yaml")
        self.push()
        against, line = proto_check.baseline(self.remote)
        self.assertEqual(against, f"{self.remote}#branch=main")
        self.assertIn("No release tag yet: breaking changes are checked against main", line)

    def test_nothing_while_main_has_no_contracts(self):
        self.commit("README.md")
        self.push()
        against, line = proto_check.baseline(self.remote)
        self.assertIsNone(against)
        self.assertIn("buf breaking has nothing to compare against, so it is skipped", line)


if __name__ == "__main__":
    unittest.main()
