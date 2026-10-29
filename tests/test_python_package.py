"""gen/python, the Python package (paddock#265), and what pins it: the protobuf
runtime moves with the generator, as the Go one does with protoc-gen-go, so
scripts/protoc.py's runtime, the floor gen/python/pyproject.toml allows and the
version `make python` tests with (tests/python-constraints.txt) are one; the
package's version is the release's, which release-please moves; and protoc.py
names the platforms it has digests for. Standard library only."""

from __future__ import annotations

import importlib.util
import json
import re
import tomllib
import unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
PYPROJECT = Path("gen/python/pyproject.toml")


def load_protoc():
    spec = importlib.util.spec_from_file_location("protoc", ROOT / "scripts" / "protoc.py")
    assert spec and spec.loader
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


class PythonPackage(unittest.TestCase):
    def setUp(self):
        self.protoc = load_protoc()
        with (ROOT / PYPROJECT).open("rb") as f:
            self.project = tomllib.load(f)["project"]

    def test_the_runtime_moves_with_the_generator(self):
        runtime = self.protoc.PYTHON_RUNTIME
        major = runtime.split(".")[0]
        self.assertTrue(
            runtime.endswith("." + self.protoc.VERSION),
            f"protoc {self.protoc.VERSION} writes code for protobuf <major>.{self.protoc.VERSION}, "
            f"not {runtime}",
        )
        self.assertEqual(self.project["dependencies"], [f"protobuf>={runtime},<{int(major) + 1}"])
        constraints = (ROOT / "tests" / "python-constraints.txt").read_text(encoding="utf-8")
        self.assertEqual(re.findall(r"^protobuf==(\S+)$", constraints, re.M), [runtime])

    def test_the_generated_code_names_the_runtime(self):
        code = (ROOT / "gen/python/twr/events/v1/callout_pb2.py").read_text(encoding="utf-8")
        self.assertIn(f"# Protobuf Python Version: {self.protoc.PYTHON_RUNTIME}\n", code)

    def test_the_version_is_the_release(self):
        manifest = json.loads((ROOT / ".release-please-manifest.json").read_text(encoding="utf-8"))
        self.assertEqual(self.project["version"], manifest["."])
        config = json.loads((ROOT / "release-please-config.json").read_text(encoding="utf-8"))
        self.assertIn(
            {"type": "toml", "path": PYPROJECT.as_posix(), "jsonpath": "$.project.version"},
            config["packages"]["."]["extra-files"],
        )

    def test_it_allows_radios_python(self):
        self.assertEqual(self.project["requires-python"], ">=3.13")

    def test_each_platform_has_a_digest(self):
        for system, machine, want in (
            ("Linux", "x86_64", "linux-x86_64"),
            ("Linux", "aarch64", "linux-aarch_64"),
            ("Darwin", "arm64", "osx-aarch_64"),
            ("Darwin", "x86_64", "osx-x86_64"),
        ):
            with self.subTest(system=system, machine=machine):
                name = self.protoc.release_platform(system, machine)
                self.assertEqual(name, want)
                self.assertRegex(self.protoc.DIGESTS[name], r"^[0-9a-f]{64}$")
        with self.assertRaises(SystemExit):
            self.protoc.release_platform("Windows", "AMD64")


if __name__ == "__main__":
    unittest.main()
