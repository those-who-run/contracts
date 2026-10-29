"""How every Python component encodes and decodes a contract message and its
CloudEvents headers (ADR 0002 in paddock, points 2 to 4), the copy of the Go
module's package `ce`, so they all do it alike: protobuf's JSON form, unknown
fields ignored on read, and the headers set by hand in the NATS binding's binary
mode.

A producer encodes:

    headers, data = ce.encode(lap, ce.Event(id=f"{src}:{sid}:lap.completed:3",
                                             source="/twr/ecu/gt7", time=received))
    await nc.publish(subject, data, headers=headers)

and a consumer decodes, checking the ce-type before it reads the payload:

    lap = lap_pb2.LapCompleted()
    event = ce.decode(msg.headers or {}, msg.data, lap)

The headers are a plain mapping of one value per name, as nats-py's are, so the
module needs no NATS client, and it keeps no state.
"""

from __future__ import annotations

import dataclasses
import datetime
import json
import re
from collections.abc import Mapping

from google.protobuf import json_format
from google.protobuf.message import Message
from twr.events.v1 import (
    callout_pb2,
    car_pb2,
    lap_pb2,
    pit_pb2,
    race_pb2,
    session_pb2,
    telemetry_pb2,
)
from twr.telemetry.v1 import frame_pb2

# The headers a message carries (ADR 0002, point 4): the exact keys Go's ce
# writes and reads.
SPEC_VERSION_HEADER = "ce-specversion"
ID_HEADER = "ce-id"
TYPE_HEADER = "ce-type"
SOURCE_HEADER = "ce-source"
TIME_HEADER = "ce-time"
CONTENT_TYPE_HEADER = "ce-datacontenttype"
MSG_ID_HEADER = "Nats-Msg-Id"

# The CloudEvents version the headers follow.
SPEC_VERSION = "1.0"
# The payload's encoding: protobuf's JSON form.
CONTENT_TYPE = "application/json"


@dataclasses.dataclass(frozen=True)
class Event:
    """What the headers say about a message besides its type, which comes from
    the message itself."""

    # The event's deterministic id, the ce-id and the Nats-Msg-Id, built as
    # subjects.yaml's msg_id scheme says, such as `<src>:<sid>:lap.completed:<n>`.
    id: str
    # The producer: `/twr/<component>`, with the game after it only where the
    # producer is game-specific, such as `/twr/ecu/gt7`.
    source: str
    # When it happened, with its time zone; ce-time is written in UTC.
    time: datetime.datetime


# Each message's ce-type, the same as Go's ce table: the contracts' Python
# round-trip test (python_test.go) checks the two agree.
_TYPES: dict[str, str] = {
    frame_pb2.Frame.DESCRIPTOR.full_name: "run.thosewho.frame.gt7.v1",
    session_pb2.SessionStarted.DESCRIPTOR.full_name: "run.thosewho.session.started.v1",
    lap_pb2.LapCompleted.DESCRIPTOR.full_name: "run.thosewho.lap.completed.v1",
    session_pb2.SessionEnded.DESCRIPTOR.full_name: "run.thosewho.session.ended.v1",
    session_pb2.SessionContinued.DESCRIPTOR.full_name: "run.thosewho.session.continued.v1",
    session_pb2.SessionDiscarded.DESCRIPTOR.full_name: "run.thosewho.session.discarded.v1",
    session_pb2.SessionKindChanged.DESCRIPTOR.full_name: "run.thosewho.session.kind.v1",
    session_pb2.SessionTrackChanged.DESCRIPTOR.full_name: "run.thosewho.session.track.v1",
    session_pb2.SessionPaused.DESCRIPTOR.full_name: "run.thosewho.session.paused.v1",
    session_pb2.SessionResumed.DESCRIPTOR.full_name: "run.thosewho.session.resumed.v1",
    lap_pb2.LapBestChanged.DESCRIPTOR.full_name: "run.thosewho.lap.best.v1",
    race_pb2.RaceFinalLapStarted.DESCRIPTOR.full_name: "run.thosewho.race.final_lap.v1",
    race_pb2.RacePositionChanged.DESCRIPTOR.full_name: "run.thosewho.race.position.v1",
    race_pb2.RaceFinished.DESCRIPTOR.full_name: "run.thosewho.race.finished.v1",
    pit_pb2.PitEntered.DESCRIPTOR.full_name: "run.thosewho.pit.entered.v1",
    pit_pb2.PitExited.DESCRIPTOR.full_name: "run.thosewho.pit.exited.v1",
    pit_pb2.PitTeleport.DESCRIPTOR.full_name: "run.thosewho.pit.teleport.v1",
    car_pb2.CarAlarmRaised.DESCRIPTOR.full_name: "run.thosewho.car.alarm.v1",
    telemetry_pb2.TelemetryLost.DESCRIPTOR.full_name: "run.thosewho.telemetry.lost.v1",
    telemetry_pb2.TelemetryRestored.DESCRIPTOR.full_name: "run.thosewho.telemetry.restored.v1",
    callout_pb2.Callout.DESCRIPTOR.full_name: "run.thosewho.callout.raised.v1",
}


def types() -> dict[str, str]:
    """Every message's ce-type, by the message's full name."""
    return dict(_TYPES)


def ce_type(m: Message) -> str:
    """The message's ce-type, such as run.thosewho.lap.completed.v1."""
    name = m.DESCRIPTOR.full_name
    try:
        return _TYPES[name]
    except KeyError:
        raise ValueError(f"{name} has no ce-type: it is not a message that travels on its own") from None


def marshal(m: Message) -> bytes:
    """The message in protobuf's JSON form, with protobuf's own field names in
    lowerCamelCase, as every language's encoder writes them."""
    return json.dumps(json_format.MessageToDict(m), ensure_ascii=False, separators=(",", ":")).encode()


def unmarshal(data: bytes | str, m: Message) -> None:
    """Decodes protobuf's JSON form into the message, ignoring the fields and
    enum values it doesn't know, so one a later release adds never breaks a
    reader (ADR 0002, point 3)."""
    json_format.Parse(data, m, ignore_unknown_fields=True)


def encode(m: Message, e: Event) -> tuple[dict[str, str], bytes]:
    """The headers and payload of a message: every CloudEvents header, with
    Nats-Msg-Id the same as ce-id."""
    t = ce_type(m)
    if not e.id:
        raise ValueError("the event has no id")
    if not e.source:
        raise ValueError("the event has no source")
    if e.time.utcoffset() is None:
        raise ValueError("the event's time has no time zone")
    for name, value in (("id", e.id), ("source", e.source)):
        if not _plain(value):
            raise ValueError(
                f"the event's {name} {value!r} has a character outside U+0021 to U+007E, "
                "or a %: the NATS binding would need it percent-encoded"
            )
    headers = {
        SPEC_VERSION_HEADER: SPEC_VERSION,
        ID_HEADER: e.id,
        TYPE_HEADER: t,
        SOURCE_HEADER: e.source,
        TIME_HEADER: _format_time(e.time),
        CONTENT_TYPE_HEADER: CONTENT_TYPE,
        MSG_ID_HEADER: e.id,
    }
    return headers, marshal(m)


def _plain(s: str) -> bool:
    """Whether s needs no percent-encoding in a header, by the CloudEvents NATS
    binding: every character printable ASCII, U+0021 to U+007E, and none of
    them %."""
    return all("\x21" <= c <= "\x7e" and c != "%" for c in s)


def _format_time(t: datetime.datetime) -> str:
    """t in UTC as Go's RFC 3339 with nanoseconds writes it: the fraction
    without its trailing zeros, none when it is zero, then Z."""
    t = t.astimezone(datetime.UTC)
    s = t.strftime("%Y-%m-%dT%H:%M:%S")
    if t.microsecond:
        s += "." + f"{t.microsecond:06d}".rstrip("0")
    return s + "Z"


# RFC 3339's date and time, the shape Go's time.Parse with RFC3339Nano takes:
# fromisoformat alone also takes ISO 8601's other forms, which Go refuses.
_RFC3339 = re.compile(r"\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?(Z|[+-]\d{2}:\d{2})", re.ASCII)


class WrongTypeError(ValueError):
    """decode's error for a message whose ce-type is not the one the caller
    expects, raised before the payload is read: Go's ce.ErrWrongType."""


def decode(headers: Mapping[str, str], data: bytes | str, m: Message) -> Event:
    """Checks the headers, the ce-type against m's, and decodes the payload into
    m, ignoring unknown fields. It returns what the headers say."""
    want = ce_type(m)
    got = headers.get(TYPE_HEADER, "")
    if got != want:
        raise WrongTypeError(f"wrong ce-type: {TYPE_HEADER} is {got!r}, want {want!r}")
    for key, value in ((SPEC_VERSION_HEADER, SPEC_VERSION), (CONTENT_TYPE_HEADER, CONTENT_TYPE)):
        if (got := headers.get(key, "")) != value:
            raise ValueError(f"{key} is {got!r}, want {value!r}")
    id_, source = headers.get(ID_HEADER, ""), headers.get(SOURCE_HEADER, "")
    if not id_ or not source:
        raise ValueError(f"{ID_HEADER} or {SOURCE_HEADER} is missing")
    raw = headers.get(TIME_HEADER, "")
    if not _RFC3339.fullmatch(raw):
        raise ValueError(f"{TIME_HEADER}: {raw!r} is not an RFC 3339 time")
    try:
        time = datetime.datetime.fromisoformat(raw)
    except ValueError as err:
        raise ValueError(f"{TIME_HEADER}: {err}") from None
    if time.utcoffset() is None:
        raise ValueError(f"{TIME_HEADER}: {raw!r} has no time zone")
    try:
        unmarshal(data, m)
    except json_format.ParseError as err:
        raise ValueError(f"decoding {want}: {err}") from err
    return Event(id=id_, source=source, time=time)
