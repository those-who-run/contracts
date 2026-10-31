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
`<run>:frame:<seq>`, without the source. A callout's is `<id>`, its own id, unique per
callout within the duplicate window; the callout's level, `p0` to `p2`, is in the subject.
A subject with no stream is core NATS only, never kept.

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
| `twr.v1.<src>.evt.<sid>.session.paused` | `twr.events.v1.SessionPaused` | `run.thosewho.session.paused.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:session.paused:<pause>` |
| `twr.v1.<src>.evt.<sid>.session.resumed` | `twr.events.v1.SessionResumed` | `run.thosewho.session.resumed.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:session.resumed:<pause>` |
| `twr.v1.<src>.evt.<sid>.lap.best` | `twr.events.v1.LapBestChanged` | `run.thosewho.lap.best.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:lap.best:<n>` |
| `twr.v1.<src>.evt.<sid>.race.final_lap` | `twr.events.v1.RaceFinalLapStarted` | `run.thosewho.race.final_lap.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:race.final_lap:1` |
| `twr.v1.<src>.evt.<sid>.race.position` | `twr.events.v1.RacePositionChanged` | `run.thosewho.race.position.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:race.position:<run>:<seq>` |
| `twr.v1.<src>.evt.<sid>.race.finished` | `twr.events.v1.RaceFinished` | `run.thosewho.race.finished.v1` | `EVENTS` | ecu | logbook, radio | `<src>:<sid>:race.finished:1` |
| `twr.v1.<src>.evt.<sid>.pit.entered` | `twr.events.v1.PitEntered` | `run.thosewho.pit.entered.v1` | `EVENTS` | ecu | logbook, radio | `<src>:<sid>:pit.entered:<entry>` |
| `twr.v1.<src>.evt.<sid>.pit.exited` | `twr.events.v1.PitExited` | `run.thosewho.pit.exited.v1` | `EVENTS` | ecu | logbook, radio | `<src>:<sid>:pit.exited:<stop>` |
| `twr.v1.<src>.evt.<sid>.pit.teleport` | `twr.events.v1.PitTeleport` | `run.thosewho.pit.teleport.v1` | `EVENTS` | ecu | logbook, radio | `<src>:<sid>:pit.teleport:<entry>` |
| `twr.v1.<src>.evt.<sid>.car.alarm` | `twr.events.v1.CarAlarmRaised` | `run.thosewho.car.alarm.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:car.alarm:<check>:<run>:<seq>` |
| `twr.v1.<src>.evt.<sid>.telemetry.lost` | `twr.events.v1.TelemetryLost` | `run.thosewho.telemetry.lost.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:telemetry.lost:<run>:<seq>` |
| `twr.v1.<src>.evt._.telemetry.lost` | `twr.events.v1.TelemetryLost` | `run.thosewho.telemetry.lost.v1` | `EVENTS` | ecu | logbook | `<src>:<run>:telemetry.lost:<seq>` |
| `twr.v1.<src>.evt.<sid>.telemetry.restored` | `twr.events.v1.TelemetryRestored` | `run.thosewho.telemetry.restored.v1` | `EVENTS` | ecu | logbook | `<src>:<sid>:telemetry.restored:<run>:<seq>` |
| `twr.v1.<src>.evt._.telemetry.restored` | `twr.events.v1.TelemetryRestored` | `run.thosewho.telemetry.restored.v1` | `EVENTS` | ecu | logbook | `<src>:<run>:telemetry.restored:<seq>` |
| `twr.v1.<src>.evt.<sid>.card.posted` | `twr.events.v1.CardPosted` | `run.thosewho.card.posted.v1` | `EVENTS` | podium | logbook | `<src>:<sid>:card.posted:1` |
| `twr.v1.<src>.callout.p0` | `twr.events.v1.Callout` | `run.thosewho.callout.raised.v1` | `CALLOUTS` | ecu | radio | `<id>` |
| `twr.v1.<src>.callout.p1` | `twr.events.v1.Callout` | `run.thosewho.callout.raised.v1` | `CALLOUTS` | ecu | radio | `<id>` |
| `twr.v1.<src>.callout.p2` | `twr.events.v1.Callout` | `run.thosewho.callout.raised.v1` | `CALLOUTS` | ecu | radio | `<id>` |
| `twr.v1.<src>.frame.gt7.lo` | `twr.telemetry.v1.Frame` | `run.thosewho.frame.gt7.v1` | none | ecu | dash | `<run>:frame:<seq>` |

## KV keys

Each value is the message in protobuf's JSON form, written with `Put`, so it has no
headers.

| Bucket | Key | Message | Writers | Readers |
| --- | --- | --- | --- | --- |
| `live` | `ecu.status.<src>` | `twr.state.v1.EcuStatus` | ecu | pitwall, radio |
| `live` | `ecu.session.<src>` | `twr.state.v1.LiveSession` | ecu | pitwall |
| `live` | `ecu.pit.<src>` | `twr.state.v1.PitState` | ecu | pitwall |
| `live` | `ecu.fuel.<src>` | `twr.state.v1.FuelState` | ecu | pitwall |
| `live` | `radio.status.<src>` | `twr.state.v1.RadioStatus` | radio | pitwall |
| `ecu` | `sessions.open.<src>` | `twr.state.v1.OpenSession` | ecu | ecu |
| `cards` | `<src>.<sid>` | `twr.state.v1.CardRecord` | podium | podium |

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
- `twr.v1.<src>.evt.<sid>.lap.best`: The session's best lap changed.
- `twr.v1.<src>.evt.<sid>.race.final_lap`: The final lap of a lap race began.
- `twr.v1.<src>.evt.<sid>.race.position`: The car's position changed and settled for 3 s.
- `twr.v1.<src>.evt.<sid>.race.finished`: The race finished, at the chequered flag or a timed race's finish.
- `twr.v1.<src>.evt.<sid>.pit.entered`: The car came into the pit lane; a pit.teleport may withdraw it.
- `twr.v1.<src>.evt.<sid>.pit.exited`: The car left the pit lane, with the stop it made; a stint starts.
- `twr.v1.<src>.evt.<sid>.pit.teleport`: A pit entry was not a stop (never at rest, no fuel, no tyres), so it is withdrawn.
- `twr.v1.<src>.evt.<sid>.car.alarm`: A tyre, water or oil temperature check raised its alarm.
- `twr.v1.<src>.evt.<sid>.telemetry.lost`: The frames stopped arriving while a session was open.
- `twr.v1.<src>.evt._.telemetry.lost`: The frames stopped arriving while no session was open.
- `twr.v1.<src>.evt.<sid>.telemetry.restored`: The frames came back after a loss, with a session open.
- `twr.v1.<src>.evt._.telemetry.restored`: The frames came back after a loss, with no session open.
- `twr.v1.<src>.evt.<sid>.card.posted`: podium posted the session's results card to the results channel.
- `twr.v1.<src>.callout.p0`: A callout with priority under 10, spoken first, such as lost telemetry, critical fuel, the pit call or the start brief.
- `twr.v1.<src>.callout.p1`: A callout with priority 10 to 19, such as the final lap, a new best, position, temperatures or fuel.
- `twr.v1.<src>.callout.p2`: A callout with priority 20 or over, such as a lap time or the race summary.
- `twr.v1.<src>.frame.gt7.lo`: The same frames at 20 Hz, at most one per 50 ms of frame time (ecu's internal/bus enforces it), with the same headers; core NATS only, for the dash.
- `live` `ecu.status.<src>`: What ecu is doing (activity, idle, the telemetry link), written on change, every second while idle and at least every 5 s while busy.
- `live` `ecu.session.<src>`: The open session as of its last frame (kind, track, car, lap, position); no key while none is open.
- `live` `ecu.pit.<src>`: Whether the car is in the pit lane, its stint and its stops; no key while no session is open.
- `live` `ecu.fuel.<src>`: The fuel level, the tank and the laps it lasts; no key while no session is open.
- `live` `radio.status.<src>`: What radio is doing (the sink, the voice connection, who is listening, when it last spoke), written on change, at most once a second, and at least every 5 s while radio runs.
- `ecu` `sessions.open.<src>`: The open session and its running summary, which a restarted ecu continues or ends from (paddock#76); ecu's own bucket, which no reader of live sees.
- `cards` `<src>.<sid>`: Whether the session's results card is being posted or is posted, and its Discord message id; no key for a session that gets no card.

Parts of an event or a value, not events or values themselves, so they have no subject or key: `twr.events.v1.CalloutSlots`, `twr.events.v1.PitStop`, `twr.events.v1.SessionGap`, `twr.events.v1.SessionKind`, `twr.events.v1.SessionSummary`, `twr.events.v1.TrackCandidate`, `twr.events.v1.TrackVerdict`, `twr.events.v1.TyreTemp`, `twr.state.v1.LiveTrack`, `twr.state.v1.OpenLap`.

## Placeholders

| Placeholder | Meaning | In |
| --- | --- | --- |
| `<src>` | where the telemetry comes from: `live`, `parity`, or `replay-` followed by the replay's name | subjects, KV keys, `msg_id` |
| `<sid>` | the session id, a ULID ecu mints | subjects, KV keys, `msg_id` |
| `<run>` | the ecu run's id, a ULID ecu mints when it starts (not a replay's name) | `msg_id` |
| `<seq>` | a frame's sequence number within the ecu run: a frame's own, or the `at` of the frame an event happened on | `msg_id` |
| `<n>` | a lap's number: lap.completed's own, or for lap.best the lap whose end published it (LapBestChanged's ended_lap) | `msg_id` |
| `<pause>` | the pause's number in the session, from 1, which ecu carries across a restart: a pause open across one keeps its number | `msg_id` |
| `<entry>` | a pit entry's number in the session, from 1, counting the entries a pit.teleport withdraws (PitEntered's entry), which ecu carries across a restart | `msg_id` |
| `<stop>` | a pit stop's number in the session, from 1 (PitStop's stop), which ecu carries across a restart | `msg_id` |
| `<check>` | a car alarm's check, CarCheck's value in lower case without its prefix: `tyre_hot`, `tyres_cold`, `water_hot` or `oil_hot` | `msg_id` |
| `<id>` | a callout's own id (Callout's id): unique per callout across runs and sources within the duplicate window, and kept by a publish sent again; minted (a ULID) or derived from run_id, at, rule and an index within the frame | `msg_id` |
