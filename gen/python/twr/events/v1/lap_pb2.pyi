from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class PitRole(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    PIT_ROLE_UNSPECIFIED: _ClassVar[PitRole]
    PIT_ROLE_NONE: _ClassVar[PitRole]
    PIT_ROLE_IN_LAP: _ClassVar[PitRole]
    PIT_ROLE_OUT_LAP: _ClassVar[PitRole]
    PIT_ROLE_IN_AND_OUT: _ClassVar[PitRole]
PIT_ROLE_UNSPECIFIED: PitRole
PIT_ROLE_NONE: PitRole
PIT_ROLE_IN_LAP: PitRole
PIT_ROLE_OUT_LAP: PitRole
PIT_ROLE_IN_AND_OUT: PitRole

class LapCompleted(_message.Message):
    __slots__ = ("session_id", "lap", "lap_time_ms", "complete", "fuel_at_start", "fuel_at_end", "fuel_used", "position", "pit_role", "first_seq", "last_seq", "counts_for_best")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    LAP_FIELD_NUMBER: _ClassVar[int]
    LAP_TIME_MS_FIELD_NUMBER: _ClassVar[int]
    COMPLETE_FIELD_NUMBER: _ClassVar[int]
    FUEL_AT_START_FIELD_NUMBER: _ClassVar[int]
    FUEL_AT_END_FIELD_NUMBER: _ClassVar[int]
    FUEL_USED_FIELD_NUMBER: _ClassVar[int]
    POSITION_FIELD_NUMBER: _ClassVar[int]
    PIT_ROLE_FIELD_NUMBER: _ClassVar[int]
    FIRST_SEQ_FIELD_NUMBER: _ClassVar[int]
    LAST_SEQ_FIELD_NUMBER: _ClassVar[int]
    COUNTS_FOR_BEST_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    lap: int
    lap_time_ms: int
    complete: bool
    fuel_at_start: float
    fuel_at_end: float
    fuel_used: float
    position: int
    pit_role: PitRole
    first_seq: int
    last_seq: int
    counts_for_best: bool
    def __init__(self, session_id: _Optional[str] = ..., lap: _Optional[int] = ..., lap_time_ms: _Optional[int] = ..., complete: _Optional[bool] = ..., fuel_at_start: _Optional[float] = ..., fuel_at_end: _Optional[float] = ..., fuel_used: _Optional[float] = ..., position: _Optional[int] = ..., pit_role: _Optional[_Union[PitRole, str]] = ..., first_seq: _Optional[int] = ..., last_seq: _Optional[int] = ..., counts_for_best: _Optional[bool] = ...) -> None: ...

class LapBestChanged(_message.Message):
    __slots__ = ("session_id", "lap", "lap_time_ms", "previous_best_ms", "gap_ms", "at", "ended_lap")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    LAP_FIELD_NUMBER: _ClassVar[int]
    LAP_TIME_MS_FIELD_NUMBER: _ClassVar[int]
    PREVIOUS_BEST_MS_FIELD_NUMBER: _ClassVar[int]
    GAP_MS_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    ENDED_LAP_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    lap: int
    lap_time_ms: int
    previous_best_ms: int
    gap_ms: int
    at: int
    ended_lap: int
    def __init__(self, session_id: _Optional[str] = ..., lap: _Optional[int] = ..., lap_time_ms: _Optional[int] = ..., previous_best_ms: _Optional[int] = ..., gap_ms: _Optional[int] = ..., at: _Optional[int] = ..., ended_lap: _Optional[int] = ...) -> None: ...
