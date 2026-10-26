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
go get github.com/those-who-run/contracts@vX.Y.Z
```

```go
import (
	eventsv1 "github.com/those-who-run/contracts/gen/go/twr/events/v1"
	telemetryv1 "github.com/those-who-run/contracts/gen/go/twr/telemetry/v1"
)
```

A component requires a released tag, never a branch, a pseudo-version or a
`replace`. A change here merges and is released before any component uses it.

The module also needs `google.golang.org/protobuf`, the runtime the generated
code is written for, at the version this module's `go.mod` requires or later.

A reader decodes with protobuf's JSON form and ignores unknown fields (Go's
`protojson.UnmarshalOptions{DiscardUnknown: true}`), so a field added in a
later release never breaks it (ADR 0002, point 3).

## The messages

| Message | What it is | Example |
| --- | --- | --- |
| `twr.telemetry.v1.Frame` | One decoded telemetry packet, about 60 a second: every value v1 stored per frame, its `seq` and the packet's raw `flags`. | [`frame.json`](examples/telemetry/v1/frame.json) |
| `twr.events.v1.SessionStarted` | A session began. | [`session_started.json`](examples/events/v1/session_started.json) |
| `twr.events.v1.LapCompleted` | A lap ended: its time, fuel, position and pit role. | [`lap_completed.json`](examples/events/v1/lap_completed.json) |
| `twr.events.v1.SessionEnded` | A session ended, why, and its summary. | [`session_ended.json`](examples/events/v1/session_ended.json) |

Which subject each message travels on, with its `ce-type`, stream,
producer, consumers and `Nats-Msg-Id` scheme, is in
[`subjects.yaml`](subjects.yaml), rendered as
[`docs/subjects.md`](docs/subjects.md).

Each field's comment in `proto/` gives its unit. The examples under
`examples/` are what `nats sub` shows, and the round-trip tests in
[`messages_test.go`](messages_test.go) decode each one.

Only Go is generated so far. TypeScript and Python come when their first
consumer needs them (Python is those-who-run/paddock#265).

## Layout

- `proto/twr/<area>/v1/*.proto`: the contracts, one protobuf package per area
  and version.
- `examples/`: one JSON example of each message, in protobuf's JSON form.
- [`subjects.yaml`](subjects.yaml): the catalogue of NATS subjects (ADR 0002,
  point 5), checked by [`internal/catalogue`](internal/catalogue/catalogue_test.go)
  and rendered to [`docs/subjects.md`](docs/subjects.md) by `make generate`.
- [`removed.txt`](removed.txt): the released messages and enums removed on
  purpose, each with the reason.
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
- no field is a 64-bit integer, which protobuf's JSON form writes as a
  string (ADR 0002, point 3);
- every message and enum of the same baseline is still here, or is listed in
  `removed.txt` with its reason: `buf breaking` in `WIRE_JSON` mode lets a
  whole message go, and a released one only goes on purpose;
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
and its example in `doc_test.go`. The v0.1 messages replaced them
(those-who-run/paddock#24).

## Releases

Merging the release PR tags `vX.Y.Z`; the module is used by that tag. It ships
no image. The first release is `v0.1.0`. Only the release App creates a `v*`
tag: the tag ruleset and the agents' guard keep everyone else from it.
