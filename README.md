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

A Go component encodes and decodes through package [`ce`](ce/ce.go), so every
component does it the same way: protobuf's JSON form, unknown fields ignored on
read (a field added in a later release never breaks a reader), and the
CloudEvents headers set and checked by hand (ADR 0002, points 2 to 4).

```go
// A producer: the payload and every header, Nats-Msg-Id the same as ce-id.
h, data, err := ce.Encode(lap, ce.Event{ID: src + ":" + sid + ":lap.completed:3", Source: "/twr/ecu/gt7", Time: received})
msg := &nats.Msg{Subject: subject, Header: nats.Header(h), Data: data}

// A consumer: the ce-type is checked before the payload is read.
var lap eventsv1.LapCompleted
event, err := ce.Decode(ce.Header(msg.Header), msg.Data, &lap) // errors.Is(err, ce.ErrWrongType)
```

`ce-source` is `/twr/<component>`, with the game after it only where the
producer is game-specific: `/twr/ecu/gt7`, `/twr/logbook`. An id or source is
printable ASCII without `%` (`Encode` refuses anything else), so no header ever
needs the NATS binding's percent-encoding. `ce.Header` has the
shape of `nats.Header`, so `ce` needs no NATS client. Encoding a frame with its
headers takes about 6 µs (`go test -bench EncodeFrame ./ce`).

## The messages

| Message | What it is | Example |
| --- | --- | --- |
| `twr.telemetry.v1.Frame` | One decoded telemetry packet, about 60 a second: every value v1 stored per frame, its `seq` and the packet's raw `flags`. | [`frame.json`](examples/telemetry/v1/frame.json) |
| `twr.events.v1.SessionStarted` | A session began. | [`session_started.json`](examples/events/v1/session_started.json) |
| `twr.events.v1.LapCompleted` | A lap ended: its time, fuel, position and pit role. | [`lap_completed.json`](examples/events/v1/lap_completed.json) |
| `twr.events.v1.SessionEnded` | A session ended, why, and its summary: v1's `summary.json`, laps and pit stops included. | [`session_ended.json`](examples/events/v1/session_ended.json); from v1's summaries, [`session_ended_suzuka_pit_stop.json`](examples/events/v1/session_ended_suzuka_pit_stop.json) and [`session_ended_suzuka_continued.json`](examples/events/v1/session_ended_suzuka_continued.json) |
| `twr.events.v1.SessionContinued` | A new ecu run continued the session the run before it left open, with the gap and the carried track. | [`session_continued.json`](examples/events/v1/session_continued.json) |
| `twr.events.v1.SessionDiscarded` | A held start was a GT7 replay or demo, not a drive. | [`session_discarded.json`](examples/events/v1/session_discarded.json) |
| `twr.events.v1.SessionKindChanged` | ecu's verdict on the session's kind changed. | [`session_kind.json`](examples/events/v1/session_kind.json) |
| `twr.events.v1.SessionTrackChanged` | ecu's verdict on the session's track changed. | [`session_track.json`](examples/events/v1/session_track.json) |
| `twr.events.v1.SessionPaused` | The session paused. | [`session_paused.json`](examples/events/v1/session_paused.json) |
| `twr.events.v1.SessionResumed` | The session's pause ended. | [`session_resumed.json`](examples/events/v1/session_resumed.json) |
| `twr.events.v1.LapBestChanged` | The session's best lap changed. | [`lap_best.json`](examples/events/v1/lap_best.json) |
| `twr.events.v1.RaceFinalLapStarted` | The final lap of a lap race began. | [`race_final_lap.json`](examples/events/v1/race_final_lap.json) |
| `twr.events.v1.RacePositionChanged` | The car's position changed and settled. | [`race_position.json`](examples/events/v1/race_position.json) |
| `twr.events.v1.RaceFinished` | The race finished, at the flag or a timed finish. | [`race_finished.json`](examples/events/v1/race_finished.json) |
| `twr.events.v1.PitEntered` | The car came into the pit lane. | [`pit_entered.json`](examples/events/v1/pit_entered.json) |
| `twr.events.v1.PitExited` | The car left the pit lane, with the stop it made and the stint that starts. | [`pit_exited.json`](examples/events/v1/pit_exited.json) |
| `twr.events.v1.PitTeleport` | A pit entry was not a stop, so it is withdrawn. | [`pit_teleport.json`](examples/events/v1/pit_teleport.json) |
| `twr.events.v1.CarAlarmRaised` | A tyre, water or oil temperature check raised its alarm. | [`car_alarm.json`](examples/events/v1/car_alarm.json) |
| `twr.events.v1.TelemetryLost` | The frames stopped arriving, in a session or with none open. | [`telemetry_lost.json`](examples/events/v1/telemetry_lost.json) |
| `twr.events.v1.TelemetryRestored` | The frames came back after a loss. | [`telemetry_restored.json`](examples/events/v1/telemetry_restored.json) |

`SessionSummary`, `SessionKind`, `TrackVerdict`, `TrackCandidate`,
`SessionGap`, `PitStop` and `TyreTemp` are parts of these events, not events of their own.
[`summary_v1_test.go`](summary_v1_test.go) checks the two summaries from v1
against what v1.0.0 wrote for its recordings, kept under
[`testdata/v1/`](testdata/v1), field by field.

Which subject each message travels on, with its `ce-type`, stream,
producers, consumers and `Nats-Msg-Id` scheme, and which KV key holds which
value, is in
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
