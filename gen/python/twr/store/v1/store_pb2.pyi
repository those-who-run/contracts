import datetime

from google.protobuf import timestamp_pb2 as _timestamp_pb2
from twr.events.v1 import lap_pb2 as _lap_pb2
from twr.events.v1 import session_pb2 as _session_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SessionRow(_message.Message):
    __slots__ = ("session_id", "source", "closed", "end_reason", "summary", "card_posted_at", "last_lap", "from_v1_rows", "deleted_at")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    CLOSED_FIELD_NUMBER: _ClassVar[int]
    END_REASON_FIELD_NUMBER: _ClassVar[int]
    SUMMARY_FIELD_NUMBER: _ClassVar[int]
    CARD_POSTED_AT_FIELD_NUMBER: _ClassVar[int]
    LAST_LAP_FIELD_NUMBER: _ClassVar[int]
    FROM_V1_ROWS_FIELD_NUMBER: _ClassVar[int]
    DELETED_AT_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    source: str
    closed: bool
    end_reason: _session_pb2.EndReason
    summary: _session_pb2.SessionSummary
    card_posted_at: _timestamp_pb2.Timestamp
    last_lap: int
    from_v1_rows: bool
    deleted_at: _timestamp_pb2.Timestamp
    def __init__(self, session_id: _Optional[str] = ..., source: _Optional[str] = ..., closed: _Optional[bool] = ..., end_reason: _Optional[_Union[_session_pb2.EndReason, str]] = ..., summary: _Optional[_Union[_session_pb2.SessionSummary, _Mapping]] = ..., card_posted_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., last_lap: _Optional[int] = ..., from_v1_rows: _Optional[bool] = ..., deleted_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class LapRow(_message.Message):
    __slots__ = ("lap", "frames", "top_speed_kph")
    LAP_FIELD_NUMBER: _ClassVar[int]
    FRAMES_FIELD_NUMBER: _ClassVar[int]
    TOP_SPEED_KPH_FIELD_NUMBER: _ClassVar[int]
    lap: _lap_pb2.LapCompleted
    frames: int
    top_speed_kph: float
    def __init__(self, lap: _Optional[_Union[_lap_pb2.LapCompleted, _Mapping]] = ..., frames: _Optional[int] = ..., top_speed_kph: _Optional[float] = ...) -> None: ...

class LapTrace(_message.Message):
    __slots__ = ("session_id", "lap", "t", "speed_kph", "throttle_pct", "brake_pct", "gear", "rpm", "x", "z")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    LAP_FIELD_NUMBER: _ClassVar[int]
    T_FIELD_NUMBER: _ClassVar[int]
    SPEED_KPH_FIELD_NUMBER: _ClassVar[int]
    THROTTLE_PCT_FIELD_NUMBER: _ClassVar[int]
    BRAKE_PCT_FIELD_NUMBER: _ClassVar[int]
    GEAR_FIELD_NUMBER: _ClassVar[int]
    RPM_FIELD_NUMBER: _ClassVar[int]
    X_FIELD_NUMBER: _ClassVar[int]
    Z_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    lap: int
    t: _containers.RepeatedScalarFieldContainer[float]
    speed_kph: _containers.RepeatedScalarFieldContainer[float]
    throttle_pct: _containers.RepeatedScalarFieldContainer[float]
    brake_pct: _containers.RepeatedScalarFieldContainer[float]
    gear: _containers.RepeatedScalarFieldContainer[int]
    rpm: _containers.RepeatedScalarFieldContainer[float]
    x: _containers.RepeatedScalarFieldContainer[float]
    z: _containers.RepeatedScalarFieldContainer[float]
    def __init__(self, session_id: _Optional[str] = ..., lap: _Optional[int] = ..., t: _Optional[_Iterable[float]] = ..., speed_kph: _Optional[_Iterable[float]] = ..., throttle_pct: _Optional[_Iterable[float]] = ..., brake_pct: _Optional[_Iterable[float]] = ..., gear: _Optional[_Iterable[int]] = ..., rpm: _Optional[_Iterable[float]] = ..., x: _Optional[_Iterable[float]] = ..., z: _Optional[_Iterable[float]] = ...) -> None: ...

class ValueCount(_message.Message):
    __slots__ = ("value", "count")
    VALUE_FIELD_NUMBER: _ClassVar[int]
    COUNT_FIELD_NUMBER: _ClassVar[int]
    value: str
    count: int
    def __init__(self, value: _Optional[str] = ..., count: _Optional[int] = ...) -> None: ...

class KindCount(_message.Message):
    __slots__ = ("session_type", "count")
    SESSION_TYPE_FIELD_NUMBER: _ClassVar[int]
    COUNT_FIELD_NUMBER: _ClassVar[int]
    session_type: _session_pb2.SessionType
    count: int
    def __init__(self, session_type: _Optional[_Union[_session_pb2.SessionType, str]] = ..., count: _Optional[int] = ...) -> None: ...

class FilterCounts(_message.Message):
    __slots__ = ("track", "car", "kind")
    TRACK_FIELD_NUMBER: _ClassVar[int]
    CAR_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    track: _containers.RepeatedCompositeFieldContainer[ValueCount]
    car: _containers.RepeatedCompositeFieldContainer[ValueCount]
    kind: _containers.RepeatedCompositeFieldContainer[KindCount]
    def __init__(self, track: _Optional[_Iterable[_Union[ValueCount, _Mapping]]] = ..., car: _Optional[_Iterable[_Union[ValueCount, _Mapping]]] = ..., kind: _Optional[_Iterable[_Union[KindCount, _Mapping]]] = ...) -> None: ...

class ListSessionsRequest(_message.Message):
    __slots__ = ("sources", "track", "car", "session_type", "from_day", "to_day", "time_zone", "offset", "limit", "known_data_version")
    SOURCES_FIELD_NUMBER: _ClassVar[int]
    TRACK_FIELD_NUMBER: _ClassVar[int]
    CAR_FIELD_NUMBER: _ClassVar[int]
    SESSION_TYPE_FIELD_NUMBER: _ClassVar[int]
    FROM_DAY_FIELD_NUMBER: _ClassVar[int]
    TO_DAY_FIELD_NUMBER: _ClassVar[int]
    TIME_ZONE_FIELD_NUMBER: _ClassVar[int]
    OFFSET_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    KNOWN_DATA_VERSION_FIELD_NUMBER: _ClassVar[int]
    sources: _containers.RepeatedScalarFieldContainer[str]
    track: str
    car: str
    session_type: _session_pb2.SessionType
    from_day: str
    to_day: str
    time_zone: str
    offset: int
    limit: int
    known_data_version: str
    def __init__(self, sources: _Optional[_Iterable[str]] = ..., track: _Optional[str] = ..., car: _Optional[str] = ..., session_type: _Optional[_Union[_session_pb2.SessionType, str]] = ..., from_day: _Optional[str] = ..., to_day: _Optional[str] = ..., time_zone: _Optional[str] = ..., offset: _Optional[int] = ..., limit: _Optional[int] = ..., known_data_version: _Optional[str] = ...) -> None: ...

class ListSessionsResponse(_message.Message):
    __slots__ = ("sessions", "options", "total", "matched", "data_version", "unchanged")
    SESSIONS_FIELD_NUMBER: _ClassVar[int]
    OPTIONS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_FIELD_NUMBER: _ClassVar[int]
    MATCHED_FIELD_NUMBER: _ClassVar[int]
    DATA_VERSION_FIELD_NUMBER: _ClassVar[int]
    UNCHANGED_FIELD_NUMBER: _ClassVar[int]
    sessions: _containers.RepeatedCompositeFieldContainer[SessionRow]
    options: FilterCounts
    total: int
    matched: int
    data_version: str
    unchanged: bool
    def __init__(self, sessions: _Optional[_Iterable[_Union[SessionRow, _Mapping]]] = ..., options: _Optional[_Union[FilterCounts, _Mapping]] = ..., total: _Optional[int] = ..., matched: _Optional[int] = ..., data_version: _Optional[str] = ..., unchanged: _Optional[bool] = ...) -> None: ...

class GetSessionRequest(_message.Message):
    __slots__ = ("sources", "session_id", "known_data_version")
    SOURCES_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    KNOWN_DATA_VERSION_FIELD_NUMBER: _ClassVar[int]
    sources: _containers.RepeatedScalarFieldContainer[str]
    session_id: str
    known_data_version: str
    def __init__(self, sources: _Optional[_Iterable[str]] = ..., session_id: _Optional[str] = ..., known_data_version: _Optional[str] = ...) -> None: ...

class GetSessionResponse(_message.Message):
    __slots__ = ("session", "data_version", "unchanged")
    SESSION_FIELD_NUMBER: _ClassVar[int]
    DATA_VERSION_FIELD_NUMBER: _ClassVar[int]
    UNCHANGED_FIELD_NUMBER: _ClassVar[int]
    session: SessionRow
    data_version: str
    unchanged: bool
    def __init__(self, session: _Optional[_Union[SessionRow, _Mapping]] = ..., data_version: _Optional[str] = ..., unchanged: _Optional[bool] = ...) -> None: ...

class ListLapsRequest(_message.Message):
    __slots__ = ("sources", "session_id", "known_data_version")
    SOURCES_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    KNOWN_DATA_VERSION_FIELD_NUMBER: _ClassVar[int]
    sources: _containers.RepeatedScalarFieldContainer[str]
    session_id: str
    known_data_version: str
    def __init__(self, sources: _Optional[_Iterable[str]] = ..., session_id: _Optional[str] = ..., known_data_version: _Optional[str] = ...) -> None: ...

class ListLapsResponse(_message.Message):
    __slots__ = ("session", "laps", "data_version", "unchanged")
    SESSION_FIELD_NUMBER: _ClassVar[int]
    LAPS_FIELD_NUMBER: _ClassVar[int]
    DATA_VERSION_FIELD_NUMBER: _ClassVar[int]
    UNCHANGED_FIELD_NUMBER: _ClassVar[int]
    session: SessionRow
    laps: _containers.RepeatedCompositeFieldContainer[LapRow]
    data_version: str
    unchanged: bool
    def __init__(self, session: _Optional[_Union[SessionRow, _Mapping]] = ..., laps: _Optional[_Iterable[_Union[LapRow, _Mapping]]] = ..., data_version: _Optional[str] = ..., unchanged: _Optional[bool] = ...) -> None: ...

class GetTraceRequest(_message.Message):
    __slots__ = ("sources", "session_id", "lap", "points", "known_data_version")
    SOURCES_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    LAP_FIELD_NUMBER: _ClassVar[int]
    POINTS_FIELD_NUMBER: _ClassVar[int]
    KNOWN_DATA_VERSION_FIELD_NUMBER: _ClassVar[int]
    sources: _containers.RepeatedScalarFieldContainer[str]
    session_id: str
    lap: int
    points: int
    known_data_version: str
    def __init__(self, sources: _Optional[_Iterable[str]] = ..., session_id: _Optional[str] = ..., lap: _Optional[int] = ..., points: _Optional[int] = ..., known_data_version: _Optional[str] = ...) -> None: ...

class GetTraceResponse(_message.Message):
    __slots__ = ("trace", "data_version", "unchanged")
    TRACE_FIELD_NUMBER: _ClassVar[int]
    DATA_VERSION_FIELD_NUMBER: _ClassVar[int]
    UNCHANGED_FIELD_NUMBER: _ClassVar[int]
    trace: LapTrace
    data_version: str
    unchanged: bool
    def __init__(self, trace: _Optional[_Union[LapTrace, _Mapping]] = ..., data_version: _Optional[str] = ..., unchanged: _Optional[bool] = ...) -> None: ...

class ListSameTrackRequest(_message.Message):
    __slots__ = ("sources", "session_id", "offset", "limit", "known_data_version")
    SOURCES_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    OFFSET_FIELD_NUMBER: _ClassVar[int]
    LIMIT_FIELD_NUMBER: _ClassVar[int]
    KNOWN_DATA_VERSION_FIELD_NUMBER: _ClassVar[int]
    sources: _containers.RepeatedScalarFieldContainer[str]
    session_id: str
    offset: int
    limit: int
    known_data_version: str
    def __init__(self, sources: _Optional[_Iterable[str]] = ..., session_id: _Optional[str] = ..., offset: _Optional[int] = ..., limit: _Optional[int] = ..., known_data_version: _Optional[str] = ...) -> None: ...

class ListSameTrackResponse(_message.Message):
    __slots__ = ("sessions", "data_version", "matched", "unchanged")
    SESSIONS_FIELD_NUMBER: _ClassVar[int]
    DATA_VERSION_FIELD_NUMBER: _ClassVar[int]
    MATCHED_FIELD_NUMBER: _ClassVar[int]
    UNCHANGED_FIELD_NUMBER: _ClassVar[int]
    sessions: _containers.RepeatedCompositeFieldContainer[SessionRow]
    data_version: str
    matched: int
    unchanged: bool
    def __init__(self, sessions: _Optional[_Iterable[_Union[SessionRow, _Mapping]]] = ..., data_version: _Optional[str] = ..., matched: _Optional[int] = ..., unchanged: _Optional[bool] = ...) -> None: ...

class DeleteSessionRequest(_message.Message):
    __slots__ = ("sources", "session_id")
    SOURCES_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    sources: _containers.RepeatedScalarFieldContainer[str]
    session_id: str
    def __init__(self, sources: _Optional[_Iterable[str]] = ..., session_id: _Optional[str] = ...) -> None: ...

class DeleteSessionResponse(_message.Message):
    __slots__ = ("session", "data_version")
    SESSION_FIELD_NUMBER: _ClassVar[int]
    DATA_VERSION_FIELD_NUMBER: _ClassVar[int]
    session: SessionRow
    data_version: str
    def __init__(self, session: _Optional[_Union[SessionRow, _Mapping]] = ..., data_version: _Optional[str] = ...) -> None: ...

class RestoreSessionRequest(_message.Message):
    __slots__ = ("sources", "session_id")
    SOURCES_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    sources: _containers.RepeatedScalarFieldContainer[str]
    session_id: str
    def __init__(self, sources: _Optional[_Iterable[str]] = ..., session_id: _Optional[str] = ...) -> None: ...

class RestoreSessionResponse(_message.Message):
    __slots__ = ("session", "data_version")
    SESSION_FIELD_NUMBER: _ClassVar[int]
    DATA_VERSION_FIELD_NUMBER: _ClassVar[int]
    session: SessionRow
    data_version: str
    def __init__(self, session: _Optional[_Union[SessionRow, _Mapping]] = ..., data_version: _Optional[str] = ...) -> None: ...
