# Subjects

Rendered from [`subjects.yaml`](../subjects.yaml) by `make generate`; edit that file,
not this one. Every message is sent as protobuf's JSON form
(`ce-datacontenttype: application/json`) with the full CloudEvents header set,
and its `Nats-Msg-Id` is the same as its `ce-id`.

| Subject | Message | `ce-type` | Captured by | Producer | Consumers | `ce-id` and `Nats-Msg-Id` |
| --- | --- | --- | --- | --- | --- | --- |
| `twr.v1.<src>.frame.gt7` | `twr.telemetry.v1.Frame` | `run.thosewho.frame.gt7.v1` | stream `TELEMETRY` | ecu | logbook | `<run>:frame:<seq>` |
| `twr.v1.<src>.evt.<sid>.session.started` | `twr.events.v1.SessionStarted` | `run.thosewho.session.started.v1` | stream `EVENTS` | ecu | logbook | `<sid>:session:started` |
| `twr.v1.<src>.evt.<sid>.lap.completed` | `twr.events.v1.LapCompleted` | `run.thosewho.lap.completed.v1` | stream `EVENTS` | ecu | logbook | `<sid>:lap:<n>:completed` |
| `twr.v1.<src>.evt.<sid>.session.ended` | `twr.events.v1.SessionEnded` | `run.thosewho.session.ended.v1` | stream `EVENTS` | ecu | logbook, podium | `<sid>:session:ended` |

## What each one is

- `twr.v1.<src>.frame.gt7`: Every decoded GT7 packet, about 60 a second.
- `twr.v1.<src>.evt.<sid>.session.started`: A session began.
- `twr.v1.<src>.evt.<sid>.lap.completed`: A lap ended, complete or partial.
- `twr.v1.<src>.evt.<sid>.session.ended`: A session ended, why, and its summary.

## Placeholders

- `<src>`: where the telemetry comes from: `live`, `replay-<run>` or `parity`.
- `<sid>`: the session id, a ULID ecu mints.
- `<run>`: the ecu run's id, a ULID ecu mints when it starts.
- `<seq>`: the frame's sequence number within the run.
- `<n>`: the lap's number.
