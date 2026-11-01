import datetime

from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Activity(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    ACTIVITY_UNSPECIFIED: _ClassVar[Activity]
    ACTIVITY_IDLE: _ClassVar[Activity]
    ACTIVITY_DRIVING: _ClassVar[Activity]
    ACTIVITY_PAUSED: _ClassVar[Activity]
    ACTIVITY_REPLAY: _ClassVar[Activity]
    ACTIVITY_DEMO: _ClassVar[Activity]
    ACTIVITY_OFFLINE: _ClassVar[Activity]
ACTIVITY_UNSPECIFIED: Activity
ACTIVITY_IDLE: Activity
ACTIVITY_DRIVING: Activity
ACTIVITY_PAUSED: Activity
ACTIVITY_REPLAY: Activity
ACTIVITY_DEMO: Activity
ACTIVITY_OFFLINE: Activity

class Track(_message.Message):
    __slots__ = ("track_id", "name", "sure", "evidence")
    TRACK_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    SURE_FIELD_NUMBER: _ClassVar[int]
    EVIDENCE_FIELD_NUMBER: _ClassVar[int]
    track_id: int
    name: str
    sure: bool
    evidence: str
    def __init__(self, track_id: _Optional[int] = ..., name: _Optional[str] = ..., sure: _Optional[bool] = ..., evidence: _Optional[str] = ...) -> None: ...

class Session(_message.Message):
    __slots__ = ("session_id", "title", "result", "started_at", "ended_at", "closed", "kind", "kind_label", "track", "car", "car_code", "finished", "final_position", "total_cars", "position_shown", "total_time_ms", "best_lap_ms", "best_lap", "laps_completed", "total_laps", "fuel_use", "fuel_burned", "fuel_capacity", "fuel_by_stint", "pit_stops", "driving_ms", "paused_ms", "joined_on_lap", "continued", "restart_laps", "card_posted_at", "from_v1_rows", "deleted_at")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    RESULT_FIELD_NUMBER: _ClassVar[int]
    STARTED_AT_FIELD_NUMBER: _ClassVar[int]
    ENDED_AT_FIELD_NUMBER: _ClassVar[int]
    CLOSED_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    KIND_LABEL_FIELD_NUMBER: _ClassVar[int]
    TRACK_FIELD_NUMBER: _ClassVar[int]
    CAR_FIELD_NUMBER: _ClassVar[int]
    CAR_CODE_FIELD_NUMBER: _ClassVar[int]
    FINISHED_FIELD_NUMBER: _ClassVar[int]
    FINAL_POSITION_FIELD_NUMBER: _ClassVar[int]
    TOTAL_CARS_FIELD_NUMBER: _ClassVar[int]
    POSITION_SHOWN_FIELD_NUMBER: _ClassVar[int]
    TOTAL_TIME_MS_FIELD_NUMBER: _ClassVar[int]
    BEST_LAP_MS_FIELD_NUMBER: _ClassVar[int]
    BEST_LAP_FIELD_NUMBER: _ClassVar[int]
    LAPS_COMPLETED_FIELD_NUMBER: _ClassVar[int]
    TOTAL_LAPS_FIELD_NUMBER: _ClassVar[int]
    FUEL_USE_FIELD_NUMBER: _ClassVar[int]
    FUEL_BURNED_FIELD_NUMBER: _ClassVar[int]
    FUEL_CAPACITY_FIELD_NUMBER: _ClassVar[int]
    FUEL_BY_STINT_FIELD_NUMBER: _ClassVar[int]
    PIT_STOPS_FIELD_NUMBER: _ClassVar[int]
    DRIVING_MS_FIELD_NUMBER: _ClassVar[int]
    PAUSED_MS_FIELD_NUMBER: _ClassVar[int]
    JOINED_ON_LAP_FIELD_NUMBER: _ClassVar[int]
    CONTINUED_FIELD_NUMBER: _ClassVar[int]
    RESTART_LAPS_FIELD_NUMBER: _ClassVar[int]
    CARD_POSTED_AT_FIELD_NUMBER: _ClassVar[int]
    FROM_V1_ROWS_FIELD_NUMBER: _ClassVar[int]
    DELETED_AT_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    title: str
    result: str
    started_at: _timestamp_pb2.Timestamp
    ended_at: _timestamp_pb2.Timestamp
    closed: bool
    kind: str
    kind_label: str
    track: Track
    car: str
    car_code: int
    finished: bool
    final_position: int
    total_cars: int
    position_shown: bool
    total_time_ms: int
    best_lap_ms: int
    best_lap: int
    laps_completed: int
    total_laps: int
    fuel_use: bool
    fuel_burned: float
    fuel_capacity: float
    fuel_by_stint: _containers.RepeatedScalarFieldContainer[float]
    pit_stops: int
    driving_ms: int
    paused_ms: int
    joined_on_lap: int
    continued: bool
    restart_laps: _containers.RepeatedScalarFieldContainer[int]
    card_posted_at: _timestamp_pb2.Timestamp
    from_v1_rows: bool
    deleted_at: _timestamp_pb2.Timestamp
    def __init__(self, session_id: _Optional[str] = ..., title: _Optional[str] = ..., result: _Optional[str] = ..., started_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., ended_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., closed: _Optional[bool] = ..., kind: _Optional[str] = ..., kind_label: _Optional[str] = ..., track: _Optional[_Union[Track, _Mapping]] = ..., car: _Optional[str] = ..., car_code: _Optional[int] = ..., finished: _Optional[bool] = ..., final_position: _Optional[int] = ..., total_cars: _Optional[int] = ..., position_shown: _Optional[bool] = ..., total_time_ms: _Optional[int] = ..., best_lap_ms: _Optional[int] = ..., best_lap: _Optional[int] = ..., laps_completed: _Optional[int] = ..., total_laps: _Optional[int] = ..., fuel_use: _Optional[bool] = ..., fuel_burned: _Optional[float] = ..., fuel_capacity: _Optional[float] = ..., fuel_by_stint: _Optional[_Iterable[float]] = ..., pit_stops: _Optional[int] = ..., driving_ms: _Optional[int] = ..., paused_ms: _Optional[int] = ..., joined_on_lap: _Optional[int] = ..., continued: _Optional[bool] = ..., restart_laps: _Optional[_Iterable[int]] = ..., card_posted_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., from_v1_rows: _Optional[bool] = ..., deleted_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class LapRow(_message.Message):
    __slots__ = ("lap", "time_ms", "complete", "fuel_start", "fuel_end", "fuel_used", "position", "frames", "top_speed_kph", "pit", "out_lap", "invalid", "best", "fuel_use", "restart", "clean")
    LAP_FIELD_NUMBER: _ClassVar[int]
    TIME_MS_FIELD_NUMBER: _ClassVar[int]
    COMPLETE_FIELD_NUMBER: _ClassVar[int]
    FUEL_START_FIELD_NUMBER: _ClassVar[int]
    FUEL_END_FIELD_NUMBER: _ClassVar[int]
    FUEL_USED_FIELD_NUMBER: _ClassVar[int]
    POSITION_FIELD_NUMBER: _ClassVar[int]
    FRAMES_FIELD_NUMBER: _ClassVar[int]
    TOP_SPEED_KPH_FIELD_NUMBER: _ClassVar[int]
    PIT_FIELD_NUMBER: _ClassVar[int]
    OUT_LAP_FIELD_NUMBER: _ClassVar[int]
    INVALID_FIELD_NUMBER: _ClassVar[int]
    BEST_FIELD_NUMBER: _ClassVar[int]
    FUEL_USE_FIELD_NUMBER: _ClassVar[int]
    RESTART_FIELD_NUMBER: _ClassVar[int]
    CLEAN_FIELD_NUMBER: _ClassVar[int]
    lap: int
    time_ms: int
    complete: bool
    fuel_start: float
    fuel_end: float
    fuel_used: float
    position: int
    frames: int
    top_speed_kph: float
    pit: bool
    out_lap: bool
    invalid: bool
    best: bool
    fuel_use: bool
    restart: bool
    clean: bool
    def __init__(self, lap: _Optional[int] = ..., time_ms: _Optional[int] = ..., complete: _Optional[bool] = ..., fuel_start: _Optional[float] = ..., fuel_end: _Optional[float] = ..., fuel_used: _Optional[float] = ..., position: _Optional[int] = ..., frames: _Optional[int] = ..., top_speed_kph: _Optional[float] = ..., pit: _Optional[bool] = ..., out_lap: _Optional[bool] = ..., invalid: _Optional[bool] = ..., best: _Optional[bool] = ..., fuel_use: _Optional[bool] = ..., restart: _Optional[bool] = ..., clean: _Optional[bool] = ...) -> None: ...

class Trace(_message.Message):
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

class FilterOption(_message.Message):
    __slots__ = ("value", "count")
    VALUE_FIELD_NUMBER: _ClassVar[int]
    COUNT_FIELD_NUMBER: _ClassVar[int]
    value: str
    count: int
    def __init__(self, value: _Optional[str] = ..., count: _Optional[int] = ...) -> None: ...

class KindOption(_message.Message):
    __slots__ = ("value", "label", "count")
    VALUE_FIELD_NUMBER: _ClassVar[int]
    LABEL_FIELD_NUMBER: _ClassVar[int]
    COUNT_FIELD_NUMBER: _ClassVar[int]
    value: str
    label: str
    count: int
    def __init__(self, value: _Optional[str] = ..., label: _Optional[str] = ..., count: _Optional[int] = ...) -> None: ...

class FilterOptions(_message.Message):
    __slots__ = ("track", "car", "kind")
    TRACK_FIELD_NUMBER: _ClassVar[int]
    CAR_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    track: _containers.RepeatedCompositeFieldContainer[FilterOption]
    car: _containers.RepeatedCompositeFieldContainer[FilterOption]
    kind: _containers.RepeatedCompositeFieldContainer[KindOption]
    def __init__(self, track: _Optional[_Iterable[_Union[FilterOption, _Mapping]]] = ..., car: _Optional[_Iterable[_Union[FilterOption, _Mapping]]] = ..., kind: _Optional[_Iterable[_Union[KindOption, _Mapping]]] = ...) -> None: ...

class ListSessionsRequest(_message.Message):
    __slots__ = ("track", "car", "kind", "from_day", "to_day", "time_zone")
    TRACK_FIELD_NUMBER: _ClassVar[int]
    CAR_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    FROM_DAY_FIELD_NUMBER: _ClassVar[int]
    TO_DAY_FIELD_NUMBER: _ClassVar[int]
    TIME_ZONE_FIELD_NUMBER: _ClassVar[int]
    track: str
    car: str
    kind: str
    from_day: str
    to_day: str
    time_zone: str
    def __init__(self, track: _Optional[str] = ..., car: _Optional[str] = ..., kind: _Optional[str] = ..., from_day: _Optional[str] = ..., to_day: _Optional[str] = ..., time_zone: _Optional[str] = ...) -> None: ...

class ListSessionsResponse(_message.Message):
    __slots__ = ("sessions", "options", "total", "matched")
    SESSIONS_FIELD_NUMBER: _ClassVar[int]
    OPTIONS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_FIELD_NUMBER: _ClassVar[int]
    MATCHED_FIELD_NUMBER: _ClassVar[int]
    sessions: _containers.RepeatedCompositeFieldContainer[Session]
    options: FilterOptions
    total: int
    matched: int
    def __init__(self, sessions: _Optional[_Iterable[_Union[Session, _Mapping]]] = ..., options: _Optional[_Union[FilterOptions, _Mapping]] = ..., total: _Optional[int] = ..., matched: _Optional[int] = ...) -> None: ...

class GetSessionRequest(_message.Message):
    __slots__ = ("session_id",)
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    def __init__(self, session_id: _Optional[str] = ...) -> None: ...

class GetSessionResponse(_message.Message):
    __slots__ = ("session",)
    SESSION_FIELD_NUMBER: _ClassVar[int]
    session: Session
    def __init__(self, session: _Optional[_Union[Session, _Mapping]] = ...) -> None: ...

class ListLapsRequest(_message.Message):
    __slots__ = ("session_id",)
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    def __init__(self, session_id: _Optional[str] = ...) -> None: ...

class ListLapsResponse(_message.Message):
    __slots__ = ("session_id", "laps")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    LAPS_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    laps: _containers.RepeatedCompositeFieldContainer[LapRow]
    def __init__(self, session_id: _Optional[str] = ..., laps: _Optional[_Iterable[_Union[LapRow, _Mapping]]] = ...) -> None: ...

class GetTraceRequest(_message.Message):
    __slots__ = ("session_id", "lap", "points")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    LAP_FIELD_NUMBER: _ClassVar[int]
    POINTS_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    lap: int
    points: int
    def __init__(self, session_id: _Optional[str] = ..., lap: _Optional[int] = ..., points: _Optional[int] = ...) -> None: ...

class GetTraceResponse(_message.Message):
    __slots__ = ("trace",)
    TRACE_FIELD_NUMBER: _ClassVar[int]
    trace: Trace
    def __init__(self, trace: _Optional[_Union[Trace, _Mapping]] = ...) -> None: ...

class ListSameTrackRequest(_message.Message):
    __slots__ = ("session_id",)
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    def __init__(self, session_id: _Optional[str] = ...) -> None: ...

class ListSameTrackResponse(_message.Message):
    __slots__ = ("sessions",)
    SESSIONS_FIELD_NUMBER: _ClassVar[int]
    sessions: _containers.RepeatedCompositeFieldContainer[Session]
    def __init__(self, sessions: _Optional[_Iterable[_Union[Session, _Mapping]]] = ...) -> None: ...

class TelemetryLink(_message.Message):
    __slots__ = ("connected", "last_packet_age_secs")
    CONNECTED_FIELD_NUMBER: _ClassVar[int]
    LAST_PACKET_AGE_SECS_FIELD_NUMBER: _ClassVar[int]
    connected: bool
    last_packet_age_secs: float
    def __init__(self, connected: _Optional[bool] = ..., last_packet_age_secs: _Optional[float] = ...) -> None: ...

class LiveTrack(_message.Message):
    __slots__ = ("track_id", "name", "sure")
    TRACK_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    SURE_FIELD_NUMBER: _ClassVar[int]
    track_id: int
    name: str
    sure: bool
    def __init__(self, track_id: _Optional[int] = ..., name: _Optional[str] = ..., sure: _Optional[bool] = ...) -> None: ...

class LiveSession(_message.Message):
    __slots__ = ("session_id", "kind", "track", "car", "lap", "total_laps", "position", "total_cars")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    TRACK_FIELD_NUMBER: _ClassVar[int]
    CAR_FIELD_NUMBER: _ClassVar[int]
    LAP_FIELD_NUMBER: _ClassVar[int]
    TOTAL_LAPS_FIELD_NUMBER: _ClassVar[int]
    POSITION_FIELD_NUMBER: _ClassVar[int]
    TOTAL_CARS_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    kind: str
    track: LiveTrack
    car: str
    lap: int
    total_laps: int
    position: int
    total_cars: int
    def __init__(self, session_id: _Optional[str] = ..., kind: _Optional[str] = ..., track: _Optional[_Union[LiveTrack, _Mapping]] = ..., car: _Optional[str] = ..., lap: _Optional[int] = ..., total_laps: _Optional[int] = ..., position: _Optional[int] = ..., total_cars: _Optional[int] = ...) -> None: ...

class EngineStatus(_message.Message):
    __slots__ = ("version", "started_at", "activity", "idle_secs", "telemetry", "voice_connected", "session")
    VERSION_FIELD_NUMBER: _ClassVar[int]
    STARTED_AT_FIELD_NUMBER: _ClassVar[int]
    ACTIVITY_FIELD_NUMBER: _ClassVar[int]
    IDLE_SECS_FIELD_NUMBER: _ClassVar[int]
    TELEMETRY_FIELD_NUMBER: _ClassVar[int]
    VOICE_CONNECTED_FIELD_NUMBER: _ClassVar[int]
    SESSION_FIELD_NUMBER: _ClassVar[int]
    version: str
    started_at: _timestamp_pb2.Timestamp
    activity: Activity
    idle_secs: float
    telemetry: TelemetryLink
    voice_connected: bool
    session: LiveSession
    def __init__(self, version: _Optional[str] = ..., started_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., activity: _Optional[_Union[Activity, str]] = ..., idle_secs: _Optional[float] = ..., telemetry: _Optional[_Union[TelemetryLink, _Mapping]] = ..., voice_connected: _Optional[bool] = ..., session: _Optional[_Union[LiveSession, _Mapping]] = ...) -> None: ...

class GetEngineStatusRequest(_message.Message):
    __slots__ = ()
    def __init__(self) -> None: ...

class GetEngineStatusResponse(_message.Message):
    __slots__ = ("status",)
    STATUS_FIELD_NUMBER: _ClassVar[int]
    status: EngineStatus
    def __init__(self, status: _Optional[_Union[EngineStatus, _Mapping]] = ...) -> None: ...

class DeleteSessionRequest(_message.Message):
    __slots__ = ("session_id",)
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    def __init__(self, session_id: _Optional[str] = ...) -> None: ...

class DeleteSessionResponse(_message.Message):
    __slots__ = ("session",)
    SESSION_FIELD_NUMBER: _ClassVar[int]
    session: Session
    def __init__(self, session: _Optional[_Union[Session, _Mapping]] = ...) -> None: ...

class RestoreSessionRequest(_message.Message):
    __slots__ = ("session_id",)
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    def __init__(self, session_id: _Optional[str] = ...) -> None: ...

class RestoreSessionResponse(_message.Message):
    __slots__ = ("session",)
    SESSION_FIELD_NUMBER: _ClassVar[int]
    session: Session
    def __init__(self, session: _Optional[_Union[Session, _Mapping]] = ...) -> None: ...
