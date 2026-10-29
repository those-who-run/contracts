from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class FinishKind(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    FINISH_KIND_UNSPECIFIED: _ClassVar[FinishKind]
    FINISH_KIND_FLAG: _ClassVar[FinishKind]
    FINISH_KIND_TIMED: _ClassVar[FinishKind]
FINISH_KIND_UNSPECIFIED: FinishKind
FINISH_KIND_FLAG: FinishKind
FINISH_KIND_TIMED: FinishKind

class RaceFinalLapStarted(_message.Message):
    __slots__ = ("session_id", "lap", "at")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    LAP_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    lap: int
    at: int
    def __init__(self, session_id: _Optional[str] = ..., lap: _Optional[int] = ..., at: _Optional[int] = ...) -> None: ...

class RacePositionChanged(_message.Message):
    __slots__ = ("session_id", "previous_position", "position", "at", "run_id")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    PREVIOUS_POSITION_FIELD_NUMBER: _ClassVar[int]
    POSITION_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    previous_position: int
    position: int
    at: int
    run_id: str
    def __init__(self, session_id: _Optional[str] = ..., previous_position: _Optional[int] = ..., position: _Optional[int] = ..., at: _Optional[int] = ..., run_id: _Optional[str] = ...) -> None: ...

class RaceFinished(_message.Message):
    __slots__ = ("session_id", "kind", "lap", "position", "total_time_ms", "best_lap", "best_lap_time_ms", "at")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    LAP_FIELD_NUMBER: _ClassVar[int]
    POSITION_FIELD_NUMBER: _ClassVar[int]
    TOTAL_TIME_MS_FIELD_NUMBER: _ClassVar[int]
    BEST_LAP_FIELD_NUMBER: _ClassVar[int]
    BEST_LAP_TIME_MS_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    kind: FinishKind
    lap: int
    position: int
    total_time_ms: int
    best_lap: int
    best_lap_time_ms: int
    at: int
    def __init__(self, session_id: _Optional[str] = ..., kind: _Optional[_Union[FinishKind, str]] = ..., lap: _Optional[int] = ..., position: _Optional[int] = ..., total_time_ms: _Optional[int] = ..., best_lap: _Optional[int] = ..., best_lap_time_ms: _Optional[int] = ..., at: _Optional[int] = ...) -> None: ...
