"""buf's checks on the contracts, the ones `make proto` and CI's `test` run.

    python3 scripts/proto_check.py <check> [--against INPUT]

Run from the repository's root. The checks:

    format    `buf format` would change no .proto file
    lint      `buf lint`, with buf.yaml's rules (STANDARD and COMMENTS)
    breaking  `buf breaking` against the latest release tag (WIRE_JSON); before
              the first tag, against `main`, and only when `main` has no contracts
              yet does it skip, saying so
    wide      no field is a 64-bit integer (int64, uint64, sint64, fixed64,
              sfixed64), which protobuf's JSON form writes as a string (ADR 0002)
    removed   every message and enum the same baseline as `breaking` has is still
              here, or is listed in removed.txt with the reason it went: buf
              breaking in WIRE_JSON mode allows a whole message to go, and a
              released one should only go on purpose
    fresh     what `buf generate` writes from proto/ is what gen/ holds, apart
              from the Python package's own files, written by hand
    all       every check above, reporting each, failing if any fails

buf and protoc-gen-go are the versions tools/go.mod pins, run with `go tool`,
and protoc the version scripts/protoc.py pins (ADR 0008 in paddock). `--against` names another input for `breaking` and
`removed`, such as a directory; tests/test_proto_check.py uses it. Standard
library only.
"""

from __future__ import annotations

import argparse
import contextlib
import filecmp
import json
import os
import re
import subprocess
import sys
import tempfile

REPOSITORY = "https://github.com/those-who-run/contracts.git"
BUF = ["go", "tool", "-modfile=tools/go.mod", "github.com/bufbuild/buf/cmd/buf"]
ENV = {**os.environ, "GOWORK": "off"}
RELEASE_TAG = re.compile(r"^v(\d+)\.(\d+)\.(\d+)$")
WIDE = {"TYPE_INT64", "TYPE_UINT64", "TYPE_SINT64", "TYPE_FIXED64", "TYPE_SFIXED64"}
# The well-known wrappers of a 64-bit integer, whose JSON form is a string too.
WIDE_WRAPPERS = {".google.protobuf.Int64Value", ".google.protobuf.UInt64Value"}
REMOVED = "removed.txt"
# The Python package around the generated code, under gen/, written by hand
# (gen/python/pyproject.toml says what each is); `fresh` leaves them be.
HAND_WRITTEN = {"python/pyproject.toml", "python/README.md", "python/twr_contracts"}
# The marker that says the generated Python has type stubs (PEP 561): protoc
# doesn't write it, so `make generate` does, and `fresh` expects it.
PY_TYPED = "python/twr/py.typed"


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


def latest_release_tag(repository: str) -> str | None:
    """The repository's highest `vX.Y.Z` tag, by version, or None."""
    out = subprocess.run(
        ["git", "ls-remote", "--tags", "--refs", repository, "v*"],
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


def main_has_contracts(repository: str) -> bool:
    with tempfile.TemporaryDirectory() as tmp:
        git = ["git", "-C", tmp]
        subprocess.run([*git, "init", "-q"], check=True)
        subprocess.run([*git, "fetch", "-q", "--depth=1", repository, "main"], check=True)
        found = subprocess.run([*git, "cat-file", "-e", "FETCH_HEAD:buf.yaml"], capture_output=True)
        return found.returncode == 0


def baseline(repository: str = REPOSITORY) -> tuple[str | None, str]:
    """What `breaking` and `removed` compare with: the git ref (a tag, or main),
    or None when there is nothing yet; and the line that says so in the log."""
    if tag := latest_release_tag(repository):
        return tag, f"Compared with the latest release, {tag}."
    if main_has_contracts(repository):
        return "main", "No release tag yet: compared with main."
    return None, (
        "No release tag and no contracts on main yet: there is nothing to compare "
        "with, so the check is skipped."
    )


@contextlib.contextmanager
def checkout(ref: str, repository: str = REPOSITORY):
    """A shallow clone of the repository at the ref, made by git, as a directory
    buf can read: buf's own clone of a git input doesn't follow every git
    setting git itself does."""
    with tempfile.TemporaryDirectory() as tmp:
        subprocess.run(
            ["git", "-c", "advice.detachedHead=false", "clone", "-q", "--depth=1", "--branch", ref, repository, tmp],
            check=True,
        )
        yield tmp


@contextlib.contextmanager
def against_input(args: argparse.Namespace):
    """The input to compare with: --against when given, else the baseline checked
    out; None, with the log line printed, when there is none."""
    if args.against:
        yield args.against
        return
    ref, line = baseline()
    print(line)
    if ref is None:
        yield None
        return
    with checkout(ref) as tmp:
        yield tmp


def check_breaking(args: argparse.Namespace) -> bool:
    with against_input(args) as against:
        return against is None or buf("breaking", "--against", against)


def image(*input_: str) -> list[dict] | None:
    """The files of a buf input (the repository's own module when none is given),
    as buf's JSON image, without their imports; None when buf fails."""
    cmd = [*BUF, "build", *input_, "--exclude-imports", "-o", "-#format=json"]
    out = subprocess.run(cmd, env=ENV, capture_output=True, text=True)
    if out.returncode != 0:
        print(out.stderr, end="", file=sys.stderr)
        return None
    return json.loads(out.stdout).get("file", [])


def declarations(files: list[dict]):
    """Every message (nested ones too) and enum in the files, by full name, with
    the message's descriptor, or None for an enum."""
    def walk(prefix: str, messages: list[dict], enums: list[dict]):
        for enum in enums:
            yield f"{prefix}.{enum['name']}", None
        for message in messages:
            name = f"{prefix}.{message['name']}"
            yield name, message
            yield from walk(name, message.get("nestedType", []), message.get("enumType", []))

    for f in files:
        yield from walk(f.get("package", ""), f.get("messageType", []), f.get("enumType", []))


def wide_type(field: dict) -> str | None:
    """The field's type when it is a 64-bit integer or a wrapper of one."""
    if field.get("type") in WIDE:
        return field["type"][5:].lower()
    if field.get("typeName") in WIDE_WRAPPERS:
        return field["typeName"][1:]
    return None


def check_wide(_: argparse.Namespace) -> bool:
    files = image()
    if files is None:
        return False
    wide = [
        f"{name}.{field['name']} ({wide_type(field)})"
        for name, message in declarations(files)
        if message
        for field in message.get("field", [])
        if wide_type(field)
    ]
    if wide:
        print(
            "a 64-bit integer field, which protobuf's JSON form writes as a string "
            "(ADR 0002): use int32, uint32 or double. " + ", ".join(wide),
            file=sys.stderr,
        )
        return False
    print("No field is a 64-bit integer.")
    return True


def allowed_removals() -> dict[str, str]:
    """removed.txt: one released message or enum per line, by full name, then the
    reason it went. Blank lines and lines starting with # are skipped."""
    allowed: dict[str, str] = {}
    if not os.path.exists(REMOVED):
        return allowed
    with open(REMOVED, encoding="utf-8") as f:
        for line in f:
            line = line.strip()
            if line and not line.startswith("#"):
                name, _, reason = line.partition(" ")
                allowed[name] = reason.strip()
    return allowed


def check_removed(args: argparse.Namespace) -> bool:
    # Messages and enums only. Services and their methods come with the pit
    # wall's Connect API, and this check takes them in then.
    allowed = allowed_removals()
    now = image()
    if now is None:
        return False
    names_now = {name for name, _ in declarations(now)}
    # A line for a name that is still here would approve a removal in advance.
    early = sorted(name for name in allowed if name in names_now)
    if early:
        print(
            f"{REMOVED} lists a name that is still here, which would approve its "
            "removal before anyone sees it: " + ", ".join(early),
            file=sys.stderr,
        )
        return False
    with against_input(args) as against:
        if against is None:
            return True
        before = image(against)
    if before is None:
        return False
    gone = sorted({name for name, _ in declarations(before)} - names_now)
    unlisted = [name for name in gone if not allowed.get(name)]
    if unlisted:
        print(
            f"released and now gone, without a line in {REMOVED} saying why: "
            + ", ".join(unlisted),
            file=sys.stderr,
        )
        return False
    for name in gone:
        print(f"Removed on purpose: {name}: {allowed[name]}")
    return True


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
        open(os.path.join(want, PY_TYPED), "w").close()
        stale = [path for path in differences(want, "gen") if path not in HAND_WRITTEN]
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
    "wide": check_wide,
    "removed": check_removed,
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
