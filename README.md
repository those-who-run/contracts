# Contracts

The contracts every Those Who Run v2 component builds against: the
protobuf messages and services, and `subjects.yaml`, the catalogue of
NATS subjects.
Components use them only by released tag, through the code generated
from them ([ADR 0002](https://github.com/those-who-run/paddock/blob/main/docs/adr/0002-contracts.md) in paddock).

Issues are off in this repository. Every issue, the roadmap and the decisions
(ADRs) live in [those-who-run/paddock](https://github.com/those-who-run/paddock). A pull request here
names its paddock issue: `Closes those-who-run/paddock#N.` when it meets every
acceptance criterion, `Part of those-who-run/paddock#N` otherwise
([`AGENTS.md`](AGENTS.md)).

## Checks

`make check` runs what CI runs; [`AGENTS.md`](AGENTS.md) says what that is and
what a new clone runs first.

## Using the contracts

The Go module is `github.com/those-who-run/contracts`, at the repository root,
and its releases are plain `vX.Y.Z` tags ([ADR 0008](https://github.com/those-who-run/paddock/blob/main/docs/adr/0008-contracts-tooling-and-tags.md) in paddock). A
component pins a release like any Go module, and imports the generated
packages under `gen/go/`, one per protobuf package:

```sh
go get github.com/those-who-run/contracts@v0.1.0
```

```go
import placeholderv1 "github.com/those-who-run/contracts/gen/go/twr/placeholder/v1"
```

A component requires a released tag, never a branch, a pseudo-version or a
`replace`. A change here merges and is released before any component uses it.

The module also needs `google.golang.org/protobuf`, the runtime the generated
code is written for, at the version this module's `go.mod` requires or later.

Only Go is generated so far. TypeScript and Python come when their first
consumer needs them (Python is those-who-run/paddock#265).

## Layout

- `proto/twr/<area>/v1/*.proto`: the contracts, one protobuf package per area
  and version. `twr.placeholder.v1` stands in until the first messages
  (those-who-run/paddock#24).
- [`buf.yaml`](buf.yaml): buf's module, lint rules (`STANDARD` and `COMMENTS`)
  and breaking-change rules (`WIRE_JSON`).
- [`buf.gen.yaml`](buf.gen.yaml): `protoc-gen-go` as a local plugin, writing
  `gen/go/`.
- `gen/go/`: the generated Go code, committed and never edited by hand.
- [`tools/go.mod`](tools/go.mod): `buf` and `protoc-gen-go`, beside
  golangci-lint and govulncheck, run with `go tool`.

`make proto` runs buf's checks, which CI's `test` runs too:

- `buf format` changes nothing;
- `buf lint` passes;
- `buf breaking` finds no change a consumer would notice, against the latest
  release tag. Before the first tag it compares with `main`, and it is skipped,
  saying so in the log, only while `main` has no contracts;
- `gen/` is what `buf generate` writes from `proto/`.

## Conventions

The ones every Go component shares, from the
[`chassis`](https://github.com/those-who-run/chassis) template.
This repository is a library, with no command or service yet; they apply to any
it adds under `cmd/`:

- **Configuration from the environment**, read once at start-up into a typed
  struct; an unset variable takes its default and a value that can't be read
  stops the process.
- **Logging with `log/slog`, as JSON** on standard output, one object per line,
  with key-value pairs.
- **A health endpoint**, `GET /healthz`, which answers 200 while the process can
  serve.

## From the template

This repository was made from `chassis` with `copier copy`, and
`.copier-answers.yml` records the release it came from and the answers given.
Renovate proposes each new `chassis` release as a PR that runs
`copier update`; chassis's README has the rest.

The module started with the template's placeholder package, `Name` in `doc.go`
and its example in `doc_test.go`, so that every check had code to run on; the
library's own API and tests replace them.

## Releases

Merging the release PR tags `vX.Y.Z`; the module is used by that tag. It ships
no image. The first release is `v0.1.0`. Only the release App creates a `v*`
tag: the tag ruleset and the agents' guard keep everyone else from it.
