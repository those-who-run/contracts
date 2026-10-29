"""protoc, the version pinned here, for buf's built-in Python generator.

    python3 scripts/protoc.py <protoc's arguments>

buf.gen.yaml names this script as the `protoc_path` of its `python` and `pyi`
plugins. Those generators are compiled into protoc, so there is no Go module to
pin them in tools/go.mod as protoc-gen-go is: this script fetches protoc's
release from GitHub once, checks it against the SHA-256 digest below, keeps it
in the user's cache ($XDG_CACHE_HOME, or ~/.cache, then
twr-contracts/protoc/<version>/), and runs it with the arguments it was given.
A laptop and CI so run the same generator (ADR 0008 in paddock), without
buf.build's remote plugins.

The Python runtime and the generator move together: protoc <M>.<m> writes code
that needs protobuf <major>.<M>.<m> or later, so a bump here bumps
gen/python/pyproject.toml's protobuf floor and tests/python-constraints.txt in
the same change (tests/test_python_package.py checks the three agree). Standard library
only.
"""

from __future__ import annotations

import hashlib
import os
import platform
import shutil
import subprocess
import sys
import tempfile
import urllib.request
import zipfile

VERSION = "36.2"
# The protobuf Python package this protoc's generated code needs, at least.
PYTHON_RUNTIME = "7.36.2"
# protoc's release archives, by platform, with the digests GitHub lists for them.
DIGESTS = {
    "linux-x86_64": "121f6c7afe1d4d0e3ea6aab9432038599250134cbf4474cb1167d2c7decd4278",
    "linux-aarch_64": "8b8f18bd2b30346efbc698dd5a73dd7c805f3ef8380f6dfc95c768f3f1852f6a",
    "osx-x86_64": "228cc7add4616cc14ca5e80dee83209d44449a7aee95a914ae748fa374efb078",
    "osx-aarch_64": "9cd98a532c5c5e0c4161314de0225de27e4c8a323917b6ea7b1b714d3ae23466",
}
URL = (
    "https://github.com/protocolbuffers/protobuf/releases/download/v{version}/protoc-{version}-{platform}.zip"
)
CACHE = os.path.join(
    os.environ.get("XDG_CACHE_HOME") or os.path.join(os.path.expanduser("~"), ".cache"),
    "twr-contracts",
    "protoc",
)


def release_platform(system: str | None = None, machine: str | None = None) -> str:
    """protoc's name for the platform, as its release archives spell it."""
    system = (system or platform.system()).lower()
    machine = (machine or platform.machine()).lower()
    os_name = {"linux": "linux", "darwin": "osx"}.get(system)
    arch = {"x86_64": "x86_64", "amd64": "x86_64", "aarch64": "aarch_64", "arm64": "aarch_64"}.get(machine)
    name = f"{os_name}-{arch}"
    if name not in DIGESTS:
        raise SystemExit(f"protoc.py: no pinned protoc for {system} on {machine}")
    return name


def install(dest: str, name: str) -> None:
    """Fetches the archive, checks its digest and unpacks it into dest. Another
    buf plugin may be doing the same at the same time, so it is unpacked beside
    dest and moved into place in one rename."""
    parent = os.path.dirname(dest)
    os.makedirs(parent, exist_ok=True)
    with tempfile.TemporaryDirectory(dir=parent) as tmp:
        archive = os.path.join(tmp, "protoc.zip")
        with (
            urllib.request.urlopen(URL.format(version=VERSION, platform=name)) as r,
            open(archive, "wb") as f,
        ):
            shutil.copyfileobj(r, f)
        with open(archive, "rb") as f:
            digest = hashlib.sha256(f.read()).hexdigest()
        if digest != DIGESTS[name]:
            raise SystemExit(
                f"protoc.py: protoc {VERSION} for {name} has digest {digest}, want {DIGESTS[name]}"
            )
        unpacked = os.path.join(tmp, "protoc")
        with zipfile.ZipFile(archive) as z:
            z.extractall(unpacked)
        os.chmod(os.path.join(unpacked, "bin", "protoc"), 0o755)
        try:
            os.rename(unpacked, dest)
        except OSError:
            if not os.path.isdir(dest):
                raise


def main(argv: list[str]) -> int:
    name = release_platform()
    dest = os.path.join(CACHE, VERSION, name)
    if not os.path.isdir(dest):
        install(dest, name)
    return subprocess.run([os.path.join(dest, "bin", "protoc"), *argv]).returncode


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
