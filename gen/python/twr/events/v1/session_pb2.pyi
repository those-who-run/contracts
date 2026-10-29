import datetime

from google.protobuf import timestamp_pb2 as _timestamp_pb2
from twr.events.v1 import lap_pb2 as _lap_pb2
from twr.events.v1 import pit_pb2 as _pit_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Game(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    GAME_UNSPECIFIED: _ClassVar[Game]
    GAME_GT7: _ClassVar[Game]

class SessionType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    SESSION_TYPE_UNSPECIFIED: _ClassVar[SessionType]
    SESSION_TYPE_RACE: _ClassVar[SessionType]
    SESSION_TYPE_TIME_TRIAL: _ClassVar[SessionType]
    SESSION_TYPE_PRACTICE: _ClassVar[SessionType]
    SESSION_TYPE_UNKNOWN: _ClassVar[SessionType]
    SESSION_TYPE_LAP_RACE: _ClassVar[SessionType]
    SESSION_TYPE_SOLO_LAPS: _ClassVar[SessionType]
    SESSION_TYPE_TIMED_FIELD: _ClassVar[SessionType]
    SESSION_TYPE_SOLO_TIMED: _ClassVar[SessionType]

class EndReason(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    END_REASON_UNSPECIFIED: _ClassVar[EndReason]
    END_REASON_FINISHED: _ClassVar[EndReason]
    END_REASON_QUIT: _ClassVar[EndReason]
    END_REASON_TELEMETRY_LOST: _ClassVar[EndReason]
    END_REASON_ENGINEER_STOPPED: _ClassVar[EndReason]
    END_REASON_UNKNOWN: _ClassVar[EndReason]
    END_REASON_RESTARTED: _ClassVar[EndReason]
    END_REASON_NEXT_SESSION_STARTED: _ClassVar[EndReason]
    END_REASON_INPUT_ENDED: _ClassVar[EndReason]
    END_REASON_TIMED_FINISH: _ClassVar[EndReason]
    END_REASON_MENU: _ClassVar[EndReason]
    END_REASON_PAUSE_LIMIT: _ClassVar[EndReason]

class FuelUse(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    FUEL_USE_UNSPECIFIED: _ClassVar[FuelUse]
    FUEL_USE_ON: _ClassVar[FuelUse]
    FUEL_USE_OFF: _ClassVar[FuelUse]
    FUEL_USE_UNKNOWN: _ClassVar[FuelUse]

class StartType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    START_TYPE_UNSPECIFIED: _ClassVar[StartType]
    START_TYPE_GRID: _ClassVar[StartType]
    START_TYPE_ROLLING: _ClassVar[StartType]
    START_TYPE_FLYING: _ClassVar[StartType]
    START_TYPE_UNKNOWN: _ClassVar[StartType]

class TrackConfidence(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    TRACK_CONFIDENCE_UNSPECIFIED: _ClassVar[TrackConfidence]
    TRACK_CONFIDENCE_NONE: _ClassVar[TrackConfidence]
    TRACK_CONFIDENCE_AMBIGUOUS: _ClassVar[TrackConfidence]
    TRACK_CONFIDENCE_PROBABLE: _ClassVar[TrackConfidence]
    TRACK_CONFIDENCE_CONFIRMED: _ClassVar[TrackConfidence]

class GapReason(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    GAP_REASON_UNSPECIFIED: _ClassVar[GapReason]
    GAP_REASON_ENGINEER_RESTART: _ClassVar[GapReason]

class DiscardReason(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    DISCARD_REASON_UNSPECIFIED: _ClassVar[DiscardReason]
    DISCARD_REASON_REPLAY: _ClassVar[DiscardReason]
    DISCARD_REASON_DEMO: _ClassVar[DiscardReason]

class PauseKind(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    PAUSE_KIND_UNSPECIFIED: _ClassVar[PauseKind]
    PAUSE_KIND_FLAGGED: _ClassVar[PauseKind]
    PAUSE_KIND_FROZEN: _ClassVar[PauseKind]
GAME_UNSPECIFIED: Game
GAME_GT7: Game
SESSION_TYPE_UNSPECIFIED: SessionType
SESSION_TYPE_RACE: SessionType
SESSION_TYPE_TIME_TRIAL: SessionType
SESSION_TYPE_PRACTICE: SessionType
SESSION_TYPE_UNKNOWN: SessionType
SESSION_TYPE_LAP_RACE: SessionType
SESSION_TYPE_SOLO_LAPS: SessionType
SESSION_TYPE_TIMED_FIELD: SessionType
SESSION_TYPE_SOLO_TIMED: SessionType
END_REASON_UNSPECIFIED: EndReason
END_REASON_FINISHED: EndReason
END_REASON_QUIT: EndReason
END_REASON_TELEMETRY_LOST: EndReason
END_REASON_ENGINEER_STOPPED: EndReason
END_REASON_UNKNOWN: EndReason
END_REASON_RESTARTED: EndReason
END_REASON_NEXT_SESSION_STARTED: EndReason
END_REASON_INPUT_ENDED: EndReason
END_REASON_TIMED_FINISH: EndReason
END_REASON_MENU: EndReason
END_REASON_PAUSE_LIMIT: EndReason
FUEL_USE_UNSPECIFIED: FuelUse
FUEL_USE_ON: FuelUse
FUEL_USE_OFF: FuelUse
FUEL_USE_UNKNOWN: FuelUse
START_TYPE_UNSPECIFIED: StartType
START_TYPE_GRID: StartType
START_TYPE_ROLLING: StartType
START_TYPE_FLYING: StartType
START_TYPE_UNKNOWN: StartType
TRACK_CONFIDENCE_UNSPECIFIED: TrackConfidence
TRACK_CONFIDENCE_NONE: TrackConfidence
TRACK_CONFIDENCE_AMBIGUOUS: TrackConfidence
TRACK_CONFIDENCE_PROBABLE: TrackConfidence
TRACK_CONFIDENCE_CONFIRMED: TrackConfidence
GAP_REASON_UNSPECIFIED: GapReason
GAP_REASON_ENGINEER_RESTART: GapReason
DISCARD_REASON_UNSPECIFIED: DiscardReason
DISCARD_REASON_REPLAY: DiscardReason
DISCARD_REASON_DEMO: DiscardReason
PAUSE_KIND_UNSPECIFIED: PauseKind
PAUSE_KIND_FLAGGED: PauseKind
PAUSE_KIND_FROZEN: PauseKind

class SessionKind(_message.Message):
    __slots__ = ("session_type", "total_laps", "field", "fuel_use", "cars_on_track", "start", "evidence")
    SESSION_TYPE_FIELD_NUMBER: _ClassVar[int]
    TOTAL_LAPS_FIELD_NUMBER: _ClassVar[int]
    FIELD_FIELD_NUMBER: _ClassVar[int]
    FUEL_USE_FIELD_NUMBER: _ClassVar[int]
    CARS_ON_TRACK_FIELD_NUMBER: _ClassVar[int]
    START_FIELD_NUMBER: _ClassVar[int]
    EVIDENCE_FIELD_NUMBER: _ClassVar[int]
    session_type: SessionType
    total_laps: int
    field: bool
    fuel_use: FuelUse
    cars_on_track: bool
    start: StartType
    evidence: str
    def __init__(self, session_type: _Optional[_Union[SessionType, str]] = ..., total_laps: _Optional[int] = ..., field: _Optional[bool] = ..., fuel_use: _Optional[_Union[FuelUse, str]] = ..., cars_on_track: _Optional[bool] = ..., start: _Optional[_Union[StartType, str]] = ..., evidence: _Optional[str] = ...) -> None: ...

class TrackCandidate(_message.Message):
    __slots__ = ("track_id", "name", "fit", "line_crossed")
    TRACK_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    FIT_FIELD_NUMBER: _ClassVar[int]
    LINE_CROSSED_FIELD_NUMBER: _ClassVar[int]
    track_id: int
    name: str
    fit: float
    line_crossed: bool
    def __init__(self, track_id: _Optional[int] = ..., name: _Optional[str] = ..., fit: _Optional[float] = ..., line_crossed: _Optional[bool] = ...) -> None: ...

class TrackVerdict(_message.Message):
    __slots__ = ("track_id", "name", "confidence", "family", "evidence", "candidates", "carried")
    TRACK_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    CONFIDENCE_FIELD_NUMBER: _ClassVar[int]
    FAMILY_FIELD_NUMBER: _ClassVar[int]
    EVIDENCE_FIELD_NUMBER: _ClassVar[int]
    CANDIDATES_FIELD_NUMBER: _ClassVar[int]
    CARRIED_FIELD_NUMBER: _ClassVar[int]
    track_id: int
    name: str
    confidence: TrackConfidence
    family: str
    evidence: str
    candidates: _containers.RepeatedCompositeFieldContainer[TrackCandidate]
    carried: bool
    def __init__(self, track_id: _Optional[int] = ..., name: _Optional[str] = ..., confidence: _Optional[_Union[TrackConfidence, str]] = ..., family: _Optional[str] = ..., evidence: _Optional[str] = ..., candidates: _Optional[_Iterable[_Union[TrackCandidate, _Mapping]]] = ..., carried: _Optional[bool] = ...) -> None: ...

class SessionGap(_message.Message):
    __slots__ = ("started_at", "ended_at", "lap", "reason")
    STARTED_AT_FIELD_NUMBER: _ClassVar[int]
    ENDED_AT_FIELD_NUMBER: _ClassVar[int]
    LAP_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    started_at: _timestamp_pb2.Timestamp
    ended_at: _timestamp_pb2.Timestamp
    lap: int
    reason: GapReason
    def __init__(self, started_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., ended_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., lap: _Optional[int] = ..., reason: _Optional[_Union[GapReason, str]] = ...) -> None: ...

class SessionStarted(_message.Message):
    __slots__ = ("session_id", "source", "game", "session_type", "car_code", "total_laps", "joined_part_way", "first_seq", "started_at", "fuel_at_start", "car_name", "frame_track_id", "joined_on_lap", "fuel_capacity")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    GAME_FIELD_NUMBER: _ClassVar[int]
    SESSION_TYPE_FIELD_NUMBER: _ClassVar[int]
    CAR_CODE_FIELD_NUMBER: _ClassVar[int]
    TOTAL_LAPS_FIELD_NUMBER: _ClassVar[int]
    JOINED_PART_WAY_FIELD_NUMBER: _ClassVar[int]
    FIRST_SEQ_FIELD_NUMBER: _ClassVar[int]
    STARTED_AT_FIELD_NUMBER: _ClassVar[int]
    FUEL_AT_START_FIELD_NUMBER: _ClassVar[int]
    CAR_NAME_FIELD_NUMBER: _ClassVar[int]
    FRAME_TRACK_ID_FIELD_NUMBER: _ClassVar[int]
    JOINED_ON_LAP_FIELD_NUMBER: _ClassVar[int]
    FUEL_CAPACITY_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    source: str
    game: Game
    session_type: SessionType
    car_code: int
    total_laps: int
    joined_part_way: bool
    first_seq: int
    started_at: _timestamp_pb2.Timestamp
    fuel_at_start: float
    car_name: str
    frame_track_id: int
    joined_on_lap: int
    fuel_capacity: float
    def __init__(self, session_id: _Optional[str] = ..., source: _Optional[str] = ..., game: _Optional[_Union[Game, str]] = ..., session_type: _Optional[_Union[SessionType, str]] = ..., car_code: _Optional[int] = ..., total_laps: _Optional[int] = ..., joined_part_way: _Optional[bool] = ..., first_seq: _Optional[int] = ..., started_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., fuel_at_start: _Optional[float] = ..., car_name: _Optional[str] = ..., frame_track_id: _Optional[int] = ..., joined_on_lap: _Optional[int] = ..., fuel_capacity: _Optional[float] = ...) -> None: ...

class SessionContinued(_message.Message):
    __slots__ = ("session_id", "gap", "track", "at", "run_id")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    GAP_FIELD_NUMBER: _ClassVar[int]
    TRACK_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    gap: SessionGap
    track: TrackVerdict
    at: int
    run_id: str
    def __init__(self, session_id: _Optional[str] = ..., gap: _Optional[_Union[SessionGap, _Mapping]] = ..., track: _Optional[_Union[TrackVerdict, _Mapping]] = ..., at: _Optional[int] = ..., run_id: _Optional[str] = ...) -> None: ...

class SessionDiscarded(_message.Message):
    __slots__ = ("session_id", "reason", "evidence", "first_seq", "at")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    EVIDENCE_FIELD_NUMBER: _ClassVar[int]
    FIRST_SEQ_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    reason: DiscardReason
    evidence: str
    first_seq: int
    at: int
    def __init__(self, session_id: _Optional[str] = ..., reason: _Optional[_Union[DiscardReason, str]] = ..., evidence: _Optional[str] = ..., first_seq: _Optional[int] = ..., at: _Optional[int] = ...) -> None: ...

class SessionKindChanged(_message.Message):
    __slots__ = ("session_id", "kind", "at", "run_id")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    kind: SessionKind
    at: int
    run_id: str
    def __init__(self, session_id: _Optional[str] = ..., kind: _Optional[_Union[SessionKind, _Mapping]] = ..., at: _Optional[int] = ..., run_id: _Optional[str] = ...) -> None: ...

class SessionTrackChanged(_message.Message):
    __slots__ = ("session_id", "track", "at", "run_id")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    TRACK_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    track: TrackVerdict
    at: int
    run_id: str
    def __init__(self, session_id: _Optional[str] = ..., track: _Optional[_Union[TrackVerdict, _Mapping]] = ..., at: _Optional[int] = ..., run_id: _Optional[str] = ...) -> None: ...

class SessionPaused(_message.Message):
    __slots__ = ("session_id", "kind", "since_t", "since", "paused_ms", "at", "pause")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    SINCE_T_FIELD_NUMBER: _ClassVar[int]
    SINCE_FIELD_NUMBER: _ClassVar[int]
    PAUSED_MS_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    PAUSE_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    kind: PauseKind
    since_t: float
    since: _timestamp_pb2.Timestamp
    paused_ms: int
    at: int
    pause: int
    def __init__(self, session_id: _Optional[str] = ..., kind: _Optional[_Union[PauseKind, str]] = ..., since_t: _Optional[float] = ..., since: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., paused_ms: _Optional[int] = ..., at: _Optional[int] = ..., pause: _Optional[int] = ...) -> None: ...

class SessionResumed(_message.Message):
    __slots__ = ("session_id", "kind", "since_t", "since", "pause_ms", "paused_ms", "at", "pause")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    SINCE_T_FIELD_NUMBER: _ClassVar[int]
    SINCE_FIELD_NUMBER: _ClassVar[int]
    PAUSE_MS_FIELD_NUMBER: _ClassVar[int]
    PAUSED_MS_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    PAUSE_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    kind: PauseKind
    since_t: float
    since: _timestamp_pb2.Timestamp
    pause_ms: int
    paused_ms: int
    at: int
    pause: int
    def __init__(self, session_id: _Optional[str] = ..., kind: _Optional[_Union[PauseKind, str]] = ..., since_t: _Optional[float] = ..., since: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., pause_ms: _Optional[int] = ..., paused_ms: _Optional[int] = ..., at: _Optional[int] = ..., pause: _Optional[int] = ...) -> None: ...

class SessionSummary(_message.Message):
    __slots__ = ("laps_completed", "best_lap", "best_lap_time_ms", "finishing_position", "fuel_used", "final_position", "session_type", "started_at", "ended_at", "driving_ms", "paused_ms", "ended_paused", "track", "kind", "car_code", "car_name", "total_laps", "fastest_lap_ms", "game_best_lap_ms", "finished", "joined_on_lap", "total_time_ms", "total_cars", "fuel_burned", "fuel_burned_by_stint", "fuel_stints", "fuel_capacity", "fuel_at_start", "fuel_at_end", "stints", "pit_stops", "gaps", "laps")
    LAPS_COMPLETED_FIELD_NUMBER: _ClassVar[int]
    BEST_LAP_FIELD_NUMBER: _ClassVar[int]
    BEST_LAP_TIME_MS_FIELD_NUMBER: _ClassVar[int]
    FINISHING_POSITION_FIELD_NUMBER: _ClassVar[int]
    FUEL_USED_FIELD_NUMBER: _ClassVar[int]
    FINAL_POSITION_FIELD_NUMBER: _ClassVar[int]
    SESSION_TYPE_FIELD_NUMBER: _ClassVar[int]
    STARTED_AT_FIELD_NUMBER: _ClassVar[int]
    ENDED_AT_FIELD_NUMBER: _ClassVar[int]
    DRIVING_MS_FIELD_NUMBER: _ClassVar[int]
    PAUSED_MS_FIELD_NUMBER: _ClassVar[int]
    ENDED_PAUSED_FIELD_NUMBER: _ClassVar[int]
    TRACK_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    CAR_CODE_FIELD_NUMBER: _ClassVar[int]
    CAR_NAME_FIELD_NUMBER: _ClassVar[int]
    TOTAL_LAPS_FIELD_NUMBER: _ClassVar[int]
    FASTEST_LAP_MS_FIELD_NUMBER: _ClassVar[int]
    GAME_BEST_LAP_MS_FIELD_NUMBER: _ClassVar[int]
    FINISHED_FIELD_NUMBER: _ClassVar[int]
    JOINED_ON_LAP_FIELD_NUMBER: _ClassVar[int]
    TOTAL_TIME_MS_FIELD_NUMBER: _ClassVar[int]
    TOTAL_CARS_FIELD_NUMBER: _ClassVar[int]
    FUEL_BURNED_FIELD_NUMBER: _ClassVar[int]
    FUEL_BURNED_BY_STINT_FIELD_NUMBER: _ClassVar[int]
    FUEL_STINTS_FIELD_NUMBER: _ClassVar[int]
    FUEL_CAPACITY_FIELD_NUMBER: _ClassVar[int]
    FUEL_AT_START_FIELD_NUMBER: _ClassVar[int]
    FUEL_AT_END_FIELD_NUMBER: _ClassVar[int]
    STINTS_FIELD_NUMBER: _ClassVar[int]
    PIT_STOPS_FIELD_NUMBER: _ClassVar[int]
    GAPS_FIELD_NUMBER: _ClassVar[int]
    LAPS_FIELD_NUMBER: _ClassVar[int]
    laps_completed: int
    best_lap: int
    best_lap_time_ms: int
    finishing_position: int
    fuel_used: float
    final_position: int
    session_type: SessionType
    started_at: _timestamp_pb2.Timestamp
    ended_at: _timestamp_pb2.Timestamp
    driving_ms: int
    paused_ms: int
    ended_paused: bool
    track: TrackVerdict
    kind: SessionKind
    car_code: int
    car_name: str
    total_laps: int
    fastest_lap_ms: int
    game_best_lap_ms: int
    finished: bool
    joined_on_lap: int
    total_time_ms: int
    total_cars: int
    fuel_burned: float
    fuel_burned_by_stint: _containers.RepeatedScalarFieldContainer[float]
    fuel_stints: _containers.RepeatedScalarFieldContainer[float]
    fuel_capacity: float
    fuel_at_start: float
    fuel_at_end: float
    stints: int
    pit_stops: _containers.RepeatedCompositeFieldContainer[_pit_pb2.PitStop]
    gaps: _containers.RepeatedCompositeFieldContainer[SessionGap]
    laps: _containers.RepeatedCompositeFieldContainer[_lap_pb2.LapCompleted]
    def __init__(self, laps_completed: _Optional[int] = ..., best_lap: _Optional[int] = ..., best_lap_time_ms: _Optional[int] = ..., finishing_position: _Optional[int] = ..., fuel_used: _Optional[float] = ..., final_position: _Optional[int] = ..., session_type: _Optional[_Union[SessionType, str]] = ..., started_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., ended_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., driving_ms: _Optional[int] = ..., paused_ms: _Optional[int] = ..., ended_paused: _Optional[bool] = ..., track: _Optional[_Union[TrackVerdict, _Mapping]] = ..., kind: _Optional[_Union[SessionKind, _Mapping]] = ..., car_code: _Optional[int] = ..., car_name: _Optional[str] = ..., total_laps: _Optional[int] = ..., fastest_lap_ms: _Optional[int] = ..., game_best_lap_ms: _Optional[int] = ..., finished: _Optional[bool] = ..., joined_on_lap: _Optional[int] = ..., total_time_ms: _Optional[int] = ..., total_cars: _Optional[int] = ..., fuel_burned: _Optional[float] = ..., fuel_burned_by_stint: _Optional[_Iterable[float]] = ..., fuel_stints: _Optional[_Iterable[float]] = ..., fuel_capacity: _Optional[float] = ..., fuel_at_start: _Optional[float] = ..., fuel_at_end: _Optional[float] = ..., stints: _Optional[int] = ..., pit_stops: _Optional[_Iterable[_Union[_pit_pb2.PitStop, _Mapping]]] = ..., gaps: _Optional[_Iterable[_Union[SessionGap, _Mapping]]] = ..., laps: _Optional[_Iterable[_Union[_lap_pb2.LapCompleted, _Mapping]]] = ...) -> None: ...

class SessionEnded(_message.Message):
    __slots__ = ("session_id", "reason", "summary", "last_seq")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    SUMMARY_FIELD_NUMBER: _ClassVar[int]
    LAST_SEQ_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    reason: EndReason
    summary: SessionSummary
    last_seq: int
    def __init__(self, session_id: _Optional[str] = ..., reason: _Optional[_Union[EndReason, str]] = ..., summary: _Optional[_Union[SessionSummary, _Mapping]] = ..., last_seq: _Optional[int] = ...) -> None: ...
