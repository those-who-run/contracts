import datetime

from google.protobuf import timestamp_pb2 as _timestamp_pb2
from twr.events.v1 import lap_pb2 as _lap_pb2
from twr.events.v1 import pit_pb2 as _pit_pb2
from twr.events.v1 import session_pb2 as _session_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class OpenLap(_message.Message):
    __slots__ = ("lap", "fuel_at_start", "pit_role")
    LAP_FIELD_NUMBER: _ClassVar[int]
    FUEL_AT_START_FIELD_NUMBER: _ClassVar[int]
    PIT_ROLE_FIELD_NUMBER: _ClassVar[int]
    lap: int
    fuel_at_start: float
    pit_role: _lap_pb2.PitRole
    def __init__(self, lap: _Optional[int] = ..., fuel_at_start: _Optional[float] = ..., pit_role: _Optional[_Union[_lap_pb2.PitRole, str]] = ...) -> None: ...

class OpenSession(_message.Message):
    __slots__ = ("session_id", "source", "run_id", "last_seen_at", "summary", "lap", "pauses", "entries", "pause", "pit_entry", "final_lap_started", "published_position")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    LAST_SEEN_AT_FIELD_NUMBER: _ClassVar[int]
    SUMMARY_FIELD_NUMBER: _ClassVar[int]
    LAP_FIELD_NUMBER: _ClassVar[int]
    PAUSES_FIELD_NUMBER: _ClassVar[int]
    ENTRIES_FIELD_NUMBER: _ClassVar[int]
    PAUSE_FIELD_NUMBER: _ClassVar[int]
    PIT_ENTRY_FIELD_NUMBER: _ClassVar[int]
    FINAL_LAP_STARTED_FIELD_NUMBER: _ClassVar[int]
    PUBLISHED_POSITION_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    source: str
    run_id: str
    last_seen_at: _timestamp_pb2.Timestamp
    summary: _session_pb2.SessionSummary
    lap: OpenLap
    pauses: int
    entries: int
    pause: _session_pb2.SessionPaused
    pit_entry: _pit_pb2.PitEntered
    final_lap_started: bool
    published_position: int
    def __init__(self, session_id: _Optional[str] = ..., source: _Optional[str] = ..., run_id: _Optional[str] = ..., last_seen_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., summary: _Optional[_Union[_session_pb2.SessionSummary, _Mapping]] = ..., lap: _Optional[_Union[OpenLap, _Mapping]] = ..., pauses: _Optional[int] = ..., entries: _Optional[int] = ..., pause: _Optional[_Union[_session_pb2.SessionPaused, _Mapping]] = ..., pit_entry: _Optional[_Union[_pit_pb2.PitEntered, _Mapping]] = ..., final_lap_started: _Optional[bool] = ..., published_position: _Optional[int] = ...) -> None: ...
