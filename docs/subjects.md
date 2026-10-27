# Subjects

Rendered from [`subjects.yaml`](../subjects.yaml) by `make generate`; edit that file,
not this one.

## Subjects

Every message is sent as protobuf's JSON form (`ce-datacontenttype: application/json`)
with the full CloudEvents header set, and its `Nats-Msg-Id` is the same as its
`ce-id`. An event subject has `_` in place of `<sid>` when it is published while
no session is open. An event's id is `<src>:<sid>:<area>.<what>:<part>`, with its
`ce-type`'s `<area>.<what>`, such as `<src>:<sid>:lap.completed:<n>`; one that happens
once per session ends in `:1`, and with no session open `<run>` takes `<sid>`'s place.
The source comes first because `EVENTS` deduplicates by `Nats-Msg-Id` across the
whole stream, and two sources can carry the same session id. A frame's id is
`<run>:frame:<seq>`, without the source.

| Subject | Message | `ce-type` | Stream | Producers | Consumers | `ce-id` and `Nats-Msg-Id` |
| --- | --- | --- | --- | --- | --- | --- |
| `twr.v1.<src>.frame.gt7` | `twr.telemetry.v1.Frame` | `run.thosewho.frame.gt7.v1` | `TELEMETRY` | ecu | logbook | `<run>:frame:<seq>` |
| `twr.v1.<src>.evt.<sid>.session.started` | `twr.events.v1.SessionStarted` | `run.thosewho.session.started.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:session.started:1` |
| `twr.v1.<src>.evt.<sid>.lap.completed` | `twr.events.v1.LapCompleted` | `run.thosewho.lap.completed.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:lap.completed:<n>` |
| `twr.v1.<src>.evt.<sid>.session.ended` | `twr.events.v1.SessionEnded` | `run.thosewho.session.ended.v1` | `EVENTS` | ecu | logbook, podium | `<src>:<sid>:session.ended:1` |

## What each one is

- `twr.v1.<src>.frame.gt7`: Every decoded GT7 packet, about 60 a second.
- `twr.v1.<src>.evt.<sid>.session.started`: A session began.
- `twr.v1.<src>.evt.<sid>.lap.completed`: A lap ended, complete or partial.
- `twr.v1.<src>.evt.<sid>.session.ended`: A session ended, why, and its summary.

Parts of an event, not events themselves, so they have no subject: `twr.events.v1.SessionSummary`.

## Placeholders

| Placeholder | Meaning | In |
| --- | --- | --- |
| `<src>` | where the telemetry comes from: `live`, `parity`, or `replay-` followed by the replay's name | subjects, KV keys, `msg_id` |
| `<sid>` | the session id, a ULID ecu mints | subjects, KV keys, `msg_id` |
| `<run>` | the ecu run's id, a ULID ecu mints when it starts (not a replay's name) | `msg_id` |
| `<seq>` | the frame's sequence number within the ecu run | `msg_id` |
| `<n>` | the lap's number | `msg_id` |
