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
| `twr.v1.<src>.evt.<sid>.session.continued` | `twr.events.v1.SessionContinued` | `run.thosewho.session.continued.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:session.continued:<run>` |
| `twr.v1.<src>.evt.<sid>.session.discarded` | `twr.events.v1.SessionDiscarded` | `run.thosewho.session.discarded.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:session.discarded:1` |
| `twr.v1.<src>.evt.<sid>.session.kind` | `twr.events.v1.SessionKindChanged` | `run.thosewho.session.kind.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:session.kind:<run>:<seq>` |
| `twr.v1.<src>.evt.<sid>.session.track` | `twr.events.v1.SessionTrackChanged` | `run.thosewho.session.track.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:session.track:<run>:<seq>` |
| `twr.v1.<src>.evt.<sid>.session.paused` | `twr.events.v1.SessionPaused` | `run.thosewho.session.paused.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:session.paused:<run>:<seq>` |
| `twr.v1.<src>.evt.<sid>.session.resumed` | `twr.events.v1.SessionResumed` | `run.thosewho.session.resumed.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:session.resumed:<run>:<seq>` |
| `twr.v1.<src>.evt.<sid>.lap.best` | `twr.events.v1.LapBestImproved` | `run.thosewho.lap.best.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:lap.best:<n>` |
| `twr.v1.<src>.evt.<sid>.race.final_lap` | `twr.events.v1.RaceFinalLapStarted` | `run.thosewho.race.final_lap.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:race.final_lap:1` |
| `twr.v1.<src>.evt.<sid>.race.position` | `twr.events.v1.RacePositionChanged` | `run.thosewho.race.position.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:race.position:<run>:<seq>` |
| `twr.v1.<src>.evt.<sid>.race.finished` | `twr.events.v1.RaceFinished` | `run.thosewho.race.finished.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:race.finished:1` |

## What each one is

- `twr.v1.<src>.frame.gt7`: Every decoded GT7 packet, about 60 a second.
- `twr.v1.<src>.evt.<sid>.session.started`: A session began.
- `twr.v1.<src>.evt.<sid>.lap.completed`: A lap ended, complete or partial.
- `twr.v1.<src>.evt.<sid>.session.ended`: A session ended, why, and its summary.
- `twr.v1.<src>.evt.<sid>.session.continued`: A new ecu run continued the session the run before it left open, after a gap.
- `twr.v1.<src>.evt.<sid>.session.discarded`: A held start was a GT7 replay or demo, not a drive; its held frames are dropped.
- `twr.v1.<src>.evt.<sid>.session.kind`: ecu's verdict on the session's kind changed.
- `twr.v1.<src>.evt.<sid>.session.track`: ecu's verdict on the session's track changed.
- `twr.v1.<src>.evt.<sid>.session.paused`: The session paused, flagged by GT7 or read from a frozen stream.
- `twr.v1.<src>.evt.<sid>.session.resumed`: The session's pause ended.
- `twr.v1.<src>.evt.<sid>.lap.best`: The session's best lap improved.
- `twr.v1.<src>.evt.<sid>.race.final_lap`: The final lap of a lap race began.
- `twr.v1.<src>.evt.<sid>.race.position`: The car's position changed and settled for 3 s.
- `twr.v1.<src>.evt.<sid>.race.finished`: The race finished, at the chequered flag or a timed race's finish.

Parts of an event, not events themselves, so they have no subject: `twr.events.v1.PitStop`, `twr.events.v1.SessionGap`, `twr.events.v1.SessionKind`, `twr.events.v1.SessionSummary`, `twr.events.v1.TrackCandidate`, `twr.events.v1.TrackVerdict`.

## Placeholders

| Placeholder | Meaning | In |
| --- | --- | --- |
| `<src>` | where the telemetry comes from: `live`, `parity`, or `replay-` followed by the replay's name | subjects, KV keys, `msg_id` |
| `<sid>` | the session id, a ULID ecu mints | subjects, KV keys, `msg_id` |
| `<run>` | the ecu run's id, a ULID ecu mints when it starts (not a replay's name) | `msg_id` |
| `<seq>` | a frame's sequence number within the ecu run: a frame's own, or the `at` of the frame an event happened on | `msg_id` |
| `<n>` | the lap's number | `msg_id` |
