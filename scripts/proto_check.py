"""buf's checks on the contracts, the ones `make proto` and CI's `test` run.

    python3 scripts/proto_check.py <check> [--against INPUT]

Run from the repository's root. The checks:

    format    `buf format` would change no .proto file
    lint      `buf lint`, with buf.yaml's rules (STANDARD and COMMENTS)
    breaking  `buf breaking` against the latest release tag (WIRE_JSON); before
              the first tag, against `main`, and only when `main` has no contracts
              yet does it skip, saying so
    fresh     what `buf generate` writes from proto/ is what gen/ holds
    all       every check above, reporting each, failing if any fails

buf and protoc-gen-go are the versions tools/go.mod pins, run with `go tool`
(ADR 0008 in paddock). `--against` names another input for `breaking`, such as
a directory; tests/test_proto_check.py uses it. Standard library only.
"""

from __future__ import annotations

import argparse
import filecmp
import os
import re
import subprocess
import sys
import tempfile

REPOSITORY = "https://github.com/those-who-run/contracts.git"
BUF = ["go", "tool", "-modfile=tools/go.mod", "github.com/bufbuild/buf/cmd/buf"]
ENV = {**os.environ, "GOWORK": "off"}
RELEASE_TAG = re.compile(r"^v(\d+)\.(\d+)\.(\d+)$")


def buf(*args: str) -> bool:
    cmd = [*BUF, *args]
    print("+ buf " + " ".join(args), flush=True)
    return subprocess.run(cmd, env=ENV).returncode == 0


def check_format(_: argparse.Namespace) -> bool:
    if buf("format", "--diff", "--exit-code"):
        return True
    print("a .proto file is not formatted: run `make proto-fmt`", file=sys.stderr)
    return False


def check_lint(_: argparse.Namespace) -> bool:
    return buf("lint")


def latest_release_tag() -> str | None:
    out = subprocess.run(
        ["git", "ls-remote", "--tags", "--refs", REPOSITORY, "v*"],
        capture_output=True,
        text=True,
        check=True,
    )
    versions = []
    for line in out.stdout.splitlines():
        tag = line.split("refs/tags/", 1)[-1]
        if match := RELEASE_TAG.match(tag):
            versions.append((tuple(int(n) for n in match.groups()), tag))
    return max(versions)[1] if versions else None


def main_has_contracts() -> bool:
    with tempfile.TemporaryDirectory() as tmp:
        git = ["git", "-C", tmp]
        subprocess.run([*git, "init", "-q"], check=True)
        subprocess.run([*git, "fetch", "-q", "--depth=1", REPOSITORY, "main"], check=True)
        found = subprocess.run([*git, "cat-file", "-e", "FETCH_HEAD:buf.yaml"], capture_output=True)
        return found.returncode == 0


def check_breaking(args: argparse.Namespace) -> bool:
    against = args.against
    if not against:
        if tag := latest_release_tag():
            print(f"Breaking changes are checked against the latest release, {tag}.")
            against = f"{REPOSITORY}#tag={tag}"
        elif main_has_contracts():
            print("No release tag yet: breaking changes are checked against main.")
            against = f"{REPOSITORY}#branch=main"
        else:
            print(
                "No release tag and no contracts on main yet: buf breaking has "
                "nothing to compare against, so it is skipped."
            )
            return True
    return buf("breaking", "--against", against)


def differences(want: str, have: str) -> list[str]:
    """Paths under either directory that are missing from the other or differ."""
    found = []
    cmp = filecmp.dircmp(want, have)
    stack = [("", cmp)]
    while stack:
        prefix, c = stack.pop()
        found += [prefix + name for name in c.left_only + c.right_only + c.diff_files + c.funny_files]
        stack += [(prefix + name + "/", sub) for name, sub in c.subdirs.items()]
    return sorted(found)


def check_fresh(_: argparse.Namespace) -> bool:
    with tempfile.TemporaryDirectory() as tmp:
        if not buf("generate", "--output", tmp):
            return False
        want = os.path.join(tmp, "gen")
        os.makedirs(want, exist_ok=True)
        os.makedirs("gen", exist_ok=True)
        stale = differences(want, "gen")
    if stale:
        print(
            "gen/ is not what buf generates from proto/: run `make generate` and "
            "commit the result. Differs: gen/" + ", gen/".join(stale),
            file=sys.stderr,
        )
        return False
    print("gen/ is what buf generates from proto/.")
    return True


CHECKS = {
    "format": check_format,
    "lint": check_lint,
    "breaking": check_breaking,
    "fresh": check_fresh,
}


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("check", choices=[*CHECKS, "all"])
    parser.add_argument("--against", help="the input `breaking` compares with")
    args = parser.parse_args(argv)
    names = list(CHECKS) if args.check == "all" else [args.check]
    failed = [name for name in names if not CHECKS[name](args)]
    if failed:
        print("failed: " + ", ".join(failed), file=sys.stderr)
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
