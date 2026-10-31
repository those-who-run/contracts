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
	statev1 "github.com/those-who-run/contracts/gen/go/twr/state/v1"
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

A KV value is written with `Put` and has no headers, so it goes through
`ce.Marshal` and `ce.Unmarshal` alone:

```go
data, err := ce.Marshal(status)               // kv.Put(ctx, "ecu.status."+src, data)
err = ce.Unmarshal(entry.Value(), status)     // unknown fields ignored, as for an event
```

A callout's subject carries its level, from its priority: package
[`callout`](callout/callout.go) builds it, so every producer splits the levels
the same way (p0 under 10, p1 10 to 19, p2 20 and over).

```go
subject, err := callout.Subject("live", c) // twr.v1.live.callout.p0 for the start brief (8)
h, data, err := ce.Encode(c, ce.Event{ID: c.GetId(), Source: "/twr/ecu/gt7", Time: c.GetCreatedAt().AsTime()})
```

A number becomes text in package [`format`](format/format.go), v1's
`format.ts` in Go, so the results card and the pit wall never show the same
lap differently (paddock#128). Its cases are
[`format/format-cases.json`](format/format-cases.json), v1 1.0.0's file byte
for byte, which its tests check in full. A formatter in another language
checks the same release's file: the pit wall's `format.ts` test reads
`format/format-cases.json` from the contracts module its `go.mod` pins (`go
list -m -f '{{.Dir}}' github.com/those-who-run/contracts`), and a Go consumer
has it from `format.Cases()`. An unknown value is `format.Missing` ("—").

```go
best := format.RaceTime(int(s.GetBestLapTimeMs())) // "1:36.595", or "—" when absent
p, of := format.Position(3, 16)                     // "P3", "/16"
```

### Python

`gen/python/` is a Python package, `twr-contracts`, which a component installs
from a release tag with no package registry (paddock#265). In its
`pyproject.toml`:

```toml
dependencies = [
  "twr-contracts @ git+https://github.com/those-who-run/contracts@vX.Y.Z#subdirectory=gen/python",
]
```

It holds the generated messages, one module per `.proto` file under `twr`
(`twr.events.v1.callout_pb2`), with their type stubs, and the Python copies of
the two helpers, which behave as the Go ones do: `twr_contracts.ce` (protobuf's
JSON form, unknown fields and enum values ignored on read, and the same
CloudEvents headers) and `twr_contracts.callout` (a callout's subject from its
priority). It needs Python 3.13 or later and `protobuf`, the runtime the
generated code is written for, at the version its `pyproject.toml` allows;
mypy needs `types-protobuf` beside it for protobuf's own types.

```python
from twr.events.v1 import callout_pb2
from twr_contracts import callout, ce

# A consumer, on nats-py: headers are one value per name, as ce expects.
c = callout_pb2.Callout()
event = ce.decode(msg.headers or {}, msg.data, c)  # ce.WrongTypeError on another ce-type

# A producer.
headers, data = ce.encode(c, ce.Event(id=c.id, source="/twr/radio", time=created))
await nc.publish(callout.subject("live", c), data, headers=headers)
```

A KV value has no headers, so Python reads it with `ce.unmarshal` alone:

```python
from twr.state.v1 import live_pb2

status = live_pb2.EcuStatus()
ce.unmarshal(entry.value, status)  # await kv.get("ecu.status.live"), on nats-py
```

Python's `datetime` holds microseconds, so a `ce-time` with nanoseconds
decodes to the microsecond. protoc's version and its digests in
`scripts/protoc.py` are bumped by hand, with the `protobuf` floor in
`pyproject.toml` and `tests/python-constraints.txt` in the same change;
`tests/test_python_package.py` fails while the three disagree. `make python` installs the package into a new
virtual environment as a consumer would, and runs the round trip between Go
and Python (`python_test.go`): every example encoded by Go decodes in Python to
the same values, and back, headers included.

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
| `twr.events.v1.CardPosted` | podium posted a session's results card: when, and the Discord message's id. | [`card_posted.json`](examples/events/v1/card_posted.json) |
| `twr.events.v1.Callout` | Something to say to the driver: the rule, the script key (`kind`) and the values that fill it (`slots`), which radio phrases, or a sentence (`text`); its priority, when it goes stale, how long it waits, and what happens to it in the pit lane. | [`callout_lap_time.json`](examples/events/v1/callout_lap_time.json), [`callout_telemetry.json`](examples/events/v1/callout_telemetry.json); from v1's callouts, [`callout_start_brief.json`](examples/events/v1/callout_start_brief.json), [`callout_position.json`](examples/events/v1/callout_position.json), [`callout_new_best.json`](examples/events/v1/callout_new_best.json) and [`callout_race_summary.json`](examples/events/v1/callout_race_summary.json) |

The values ecu keeps in NATS KV, radio's status and podium's cards, each under
a key that carries its source (`<src>`), so a replay writes its own keys and never the
live source's:

| Message | Bucket and key | What it is | Example |
| --- | --- | --- | --- |
| `twr.state.v1.EcuStatus` | `live` `ecu.status.<src>` | What ecu is doing: its activity, whether it is idle and for how long, the telemetry link. v1's `/api/v1/status` and `/healthz`, but for radio's voice. | [`ecu_status.json`](examples/state/v1/ecu_status.json) |
| `twr.state.v1.LiveSession` | `live` `ecu.session.<src>` | The open session as of its last frame: kind, track, car, lap, position. | [`ecu_session.json`](examples/state/v1/ecu_session.json) |
| `twr.state.v1.PitState` | `live` `ecu.pit.<src>` | Whether the car is in the pit lane, its stint and its stops. | [`pit_state.json`](examples/state/v1/pit_state.json) |
| `twr.state.v1.FuelState` | `live` `ecu.fuel.<src>` | The fuel, the tank, and the laps it lasts. | [`fuel_state.json`](examples/state/v1/fuel_state.json) |
| `twr.state.v1.RadioStatus` | `live` `radio.status.<src>` | What radio is doing: where it speaks, whether it is in the voice channel, how many are listening, when it last spoke. v1's `voice_connected`. | [`radio_status.json`](examples/state/v1/radio_status.json) |
| `twr.state.v1.OpenSession` | `ecu` `sessions.open.<src>` | The open session and its running summary, which ecu continues or ends from after a restart. In ecu's own bucket, so a reader of `live` never sees it. | [`open_session.json`](examples/state/v1/open_session.json) |
| `twr.state.v1.CardRecord` | `cards` `<src>.<sid>` | podium's record of a session's results card: being posted or posted, and its Discord message id, so a `session.ended` delivered again posts nothing. | [`card_record.json`](examples/state/v1/card_record.json) |

`SessionSummary`, `SessionKind`, `TrackVerdict`, `TrackCandidate`,
`SessionGap`, `PitStop`, `TyreTemp` and `CalloutSlots` are parts of these events, not
events of their own, and `LiveTrack` and `OpenLap` parts of the values.
[`summary_v1_test.go`](summary_v1_test.go) checks the two summaries from v1
against what v1.0.0 wrote for its recordings, kept under
[`testdata/v1/`](testdata/v1), field by field, and
[`callout_test.go`](callout_test.go) does the same for four of v1's callouts, and
[`status_v1_test.go`](status_v1_test.go) for v1's status route.

Which subject each message travels on, with its `ce-type`, stream,
producers, consumers and `Nats-Msg-Id` scheme, and which KV key holds which
value, is in
[`subjects.yaml`](subjects.yaml), rendered as
[`docs/subjects.md`](docs/subjects.md).

Each field's comment in `proto/` gives its unit. The examples under
`examples/` are what `nats sub` shows, and the round-trip tests in
[`messages_test.go`](messages_test.go) decode each one.

Go and Python are generated ([Python](#python) says how a component installs
it). TypeScript comes with its first consumer, the pit wall.

## Layout

- `proto/twr/<area>/v1/*.proto`: the contracts, one protobuf package per area
  and version.
- `examples/`: one JSON example of each message, in protobuf's JSON form.
- [`format/`](format/format.go): how numbers become text, and
  `format-cases.json`, the cases every formatter of the project checks.
- [`subjects.yaml`](subjects.yaml): the catalogue of NATS subjects (ADR 0002,
  point 5), checked by [`internal/catalogue`](internal/catalogue/catalogue_test.go)
  and rendered to [`docs/subjects.md`](docs/subjects.md) by `make generate`.
- [`removed.txt`](removed.txt): the released messages and enums removed on
  purpose, each with the reason.
- [`buf.yaml`](buf.yaml): buf's module, lint rules (`STANDARD` and `COMMENTS`)
  and breaking-change rules (`WIRE_JSON`).
- [`buf.gen.yaml`](buf.gen.yaml): `protoc-gen-go` as a local plugin, writing
  `gen/go/`, and protoc's built-in `python` and `pyi` generators, writing
  `gen/python/twr/`.
- `gen/go/`: the generated Go code, committed and never edited by hand.
- `gen/python/`: the Python package. `twr/` is generated, committed and never
  edited by hand; [`pyproject.toml`](gen/python/pyproject.toml), its
  `README.md` and the helpers in `twr_contracts/` are written by hand.
- [`tools/go.mod`](tools/go.mod): `buf` and `protoc-gen-go`, beside
  golangci-lint and govulncheck, run with `go tool`.
- [`scripts/protoc.py`](scripts/protoc.py): protoc, for the Python
  generators, which are built into it: the release it pins, fetched from
  GitHub once, checked by its digest and kept in the user's cache. A bump of
  protoc bumps the `protobuf` floor in `gen/python/pyproject.toml` and
  `tests/python-constraints.txt` with it.

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
- `gen/` is what `buf generate` writes from `proto/`, Go and Python, apart
  from the Python package's own files, written by hand.

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
