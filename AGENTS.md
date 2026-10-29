# Agents and contributors

`contracts` is part of Those Who Run v2; [`README.md`](README.md) says what it is.

The rules for every v2 repository live once, in paddock's
[`AGENTS.md`](https://github.com/those-who-run/paddock/blob/main/AGENTS.md), and apply here in full, for
coding agents and for people. This file adds only what is specific to this
repository, and never repeats or weakens a rule from there.

## Issues

Issues are off here: every issue lives in
[those-who-run/paddock](https://github.com/those-who-run/paddock/issues). Read one with
`gh issue view N -R those-who-run/paddock --comments`. A PR here says
`Closes those-who-run/paddock#N.` only when every acceptance criterion is met,
and `Part of those-who-run/paddock#N` otherwise, as paddock's `AGENTS.md` says.

## Definition of done

`make check` passes. It runs what CI runs, in three parts:

- the commit hooks on every file (CI's `checks / checks`);
- the Go checks, in the order go.yml in `those-who-run/.github` runs them (CI's
  `go / check`): gofmt, `go mod tidy` for the module and `tools/go.mod`, go vet,
  golangci-lint with [`.golangci.yml`](.golangci.yml), the tests with the race
  detector (a skipped test fails, and so does total coverage below the floor,
  `COVERAGE_FLOOR` in the [`Makefile`](Makefile)), go build and govulncheck;
- the repository's own checks (with the Go tests, CI's `test`): every internal
  Markdown link and anchor resolves, the guard and settings in `.claude/` are
  `those-who-run/.github`'s `agent/` at the commit CI pins, buf's checks
  (`make proto`: format, lint, breaking changes, no 64-bit integers, no
  released message removed without its reason, and `gen/` is fresh), the
  tests in `tests/`, and the Python leg (`make python`).

It needs Go (the version in `go.mod`'s `toolchain` line, which the `go`
command fetches when `GOTOOLCHAIN` allows it), `python3`, `pre-commit` and the
network. CI's `image / build`, where the repository ships an image, also builds
the Dockerfile and smoke-tests the image; `make check` doesn't.

A change also carries a test for each acceptance criterion of its issue
([Go](#go) says where), docs changed in the same PR as what they describe,
and a recording that pins a behaviour where one applies.

## Go

- **Packages.** Anything not meant to be imported by another module lives
  under `internal/`, which the `go` command keeps other modules from
  importing. A package outside `internal/` is an API other repositories may
  build against.
- **Tests.** Each acceptance criterion has a test in a `_test.go` file named
  for what it checks, beside the code it tests, such as
  [`internal/dependencies`](internal/dependencies/dependencies_test.go) for the dependency table.
  Fixtures go under `testdata/`, which the `go` command skips.
- **Tool versions.** The Go version is `go.mod`'s `go` and `toolchain`
  lines. golangci-lint, govulncheck, `buf` and `protoc-gen-go` are pinned in
  [`tools/go.mod`](tools/go.mod), checksummed in `tools/go.sum`, and run with
  `go tool -modfile=tools/go.mod`, as the Makefile does; never a copy
  installed some other way. protoc, for the Python generators, is pinned with
  its digests in [`scripts/protoc.py`](scripts/protoc.py), which `buf
  generate` runs, and bumped there by hand. Renovate bumps the Go tools; a new
  one is added there
  the way `those-who-run/.github`'s
  [workflows README](https://github.com/those-who-run/.github/blob/main/.github/workflows/README.md#toolsgomod)
  says.
- **Formatting and linting.** gofmt (`make fmt`) and golangci-lint with
  [`.golangci.yml`](.golangci.yml) (`make lint`). The commit hooks, `make
  check` and CI's `go / check` all refuse unformatted code and lint findings.
  A finding is fixed, or silenced at its line with
  `//nolint:<linter> // <why>`.
- **Generated files.** `gen/go/` and `gen/python/twr/` are written by `make
  generate` from `proto/`; the rest of `gen/python/`, the package around the
  generated Python, is written by hand.
  A generated file is written by a `make` target or a `//go:generate` line,
  starts with Go's `// Code generated … DO NOT EDIT.` line (protoc's `DO NOT
  EDIT!` line in Python), is committed, and
  is never edited by hand: change its source and generate it again. That line also leaves the
  file out of the coverage floor, in `make check` and in CI.
- **Dependencies.** Declared in `go.mod` and locked in `go.sum`, both
  committed and tidy (`go mod tidy`; the commit hooks and CI check it). No
  `vendor/` directory. Each direct `require` has a row in
  [`docs/dependencies.md`](docs/dependencies.md) saying why it is there, added
  in the same PR; `internal/dependencies` fails `make check` without it.
- **Contracts.** The contracts module, and any other module of the project,
  is required at a released tag (`vX.Y.Z`), never at a pseudo-version from a
  branch or an untagged commit, and never through a `replace` to a local path
  or a fork.

## Contracts

What [ADR 0002](https://github.com/those-who-run/paddock/blob/main/docs/adr/0002-contracts.md)
and [ADR 0008](https://github.com/those-who-run/paddock/blob/main/docs/adr/0008-contracts-tooling-and-tags.md)
in paddock mean for a change here:

- **Every message and field has a comment**, and so does every enum, value,
  service and method. `buf lint` fails without one.
- **No 64-bit integers** (`int64`, `uint64`, `sint64`, `fixed64`,
  `sfixed64`, or the `Int64Value` and `UInt64Value` wrappers): protobuf's JSON
  form writes them as strings. `int32`, `uint32`
  and `double` are enough, and `make proto` fails on one.
- **`optional` wherever zero is a real value** distinct from "absent", since the
  JSON form leaves zeros out; every enum starts with `<ENUM>_UNSPECIFIED = 0`.
- **Units go in the field's name or its comment**, such as `speed_kph` or "in
  litres".
- **A field number or name is never reused.** A field that is removed is
  `reserved`, by number and by name, in the same change.
- **A released message or enum is removed only on purpose:** its full name and
  the reason go in [`removed.txt`](removed.txt) in the same change, and the
  review sees both. `make proto` fails a removal without its line, and a line
  for a name that is still here, which would approve a removal in advance.
- **Messages are encoded and decoded through [`ce`](ce/ce.go)**, never with
  protojson or the headers by hand. A new message that travels on its own gets
  its `ce-type` in `ce`'s table and in `subjects.yaml` in the same change; a
  test fails while the two disagree.
- **Every subject and KV key is in [`subjects.yaml`](subjects.yaml)**, with its
  message, producers and consumers. A subject also has its `ce-type`
  (`run.thosewho.<area>.<what>.v<N>`, one to one with its message), stream and
  `Nats-Msg-Id` scheme, which for an event is `<src>:<sid>:<area>.<what>:<part>`
  (`<src>:<sid>:lap.completed:<n>`); a KV key its bucket, and a key in the
  `live` and `ecu` buckets carries its `<src>` (`catalogue.SourceBuckets`). A new message in an events package
  (`twr.events.*`, any version) needs its entry in the same change, and one in
  a state package (`twr.state.*`) its KV entry, or a line
  under `helpers` if it is only a part of an event or a value: `internal/catalogue`'s test
  fails without one, and on a pattern outside ADR 0001's subject scheme. A new
  placeholder goes in `catalogue.Placeholders`, which the check and the doc
  both read. `make generate` renders `docs/subjects.md` from the file.
- **Each message has a JSON example** under `examples/`, which
  [`messages_test.go`](messages_test.go) decodes and round-trips, an unknown
  field included.
- **A breaking change is a new package version beside the old one**, such as
  `twr.laps.v2` next to `twr.laps.v1`, never an edit of the old one. `buf
  breaking` fails a change a consumer would notice on the wire or in protobuf's
  JSON form.
- **A change merges and is released before any component uses it.** A component
  requires a released tag, never a branch.
- **`gen/` is regenerated in the same change** as the `.proto` it comes from:
  `make proto-fmt`, then `make generate`, and commit both. CI fails when `gen/`
  is not what buf writes.
- **The protobuf runtime and the generator move together.**
  `google.golang.org/protobuf` in `go.mod` is a released version compatible with
  `protoc-gen-go` in `tools/go.mod`, never behind it, and the two change in the
  same PR. So do protoc in `scripts/protoc.py` and the `protobuf` floor in
  `gen/python/pyproject.toml` and `tests/python-constraints.txt`
  (`tests/test_python_package.py` checks them).
- **Go and Python are generated** (Python for `radio`, paddock#265); another
  language waits until a consumer needs it.
- **Python reads and writes through
  [`twr_contracts`](gen/python/twr_contracts/ce.py)**, the copy of `ce` and
  `callout`, which behaves as they do. A new message that travels on its own
  gets its `ce-type` in its table too; `make python`'s round trip
  (`python_test.go`) fails while the two disagree, or while Go and Python
  decode an example to different values.

## Make targets

- `make check`: everything a PR must pass, as CI runs it.
- `make test`: the Go tests with the coverage floor, and the repository's own
  checks: CI's `test` job, while iterating.
- `make fmt`: formats the Go code; `make lint`: golangci-lint.
- `make proto`: buf's checks; `make proto-fmt`: formats the `.proto` files;
  `make generate`: writes `gen/` from `proto/`.
- `make python`: installs `gen/python` into a new virtual environment and runs
  the round trip between Go and Python (part of `make test`).
- `make hooks`: installs the commit hooks. A new clone or worktree runs it
  once, so the hooks run on every commit and commit message.

`make help` lists the rest, among them each Go check on its own.

## The guard

`.claude/settings.json` and `.claude/hooks/guard.py` are copies: a change to
them is made in `those-who-run/.github`, then copied here unchanged with the
bump of the `checks.yml` pin in `.github/workflows/ci.yml`.
`.claude/guard.json` holds this repository's own values.
