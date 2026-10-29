from twr.telemetry.v1 import frame_pb2 as _frame_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class TelemetryLost(_message.Message):
    __slots__ = ("session_id", "run_id", "phase", "silence_ms", "at")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    PHASE_FIELD_NUMBER: _ClassVar[int]
    SILENCE_MS_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    run_id: str
    phase: _frame_pb2.Phase
    silence_ms: int
    at: int
    def __init__(self, session_id: _Optional[str] = ..., run_id: _Optional[str] = ..., phase: _Optional[_Union[_frame_pb2.Phase, str]] = ..., silence_ms: _Optional[int] = ..., at: _Optional[int] = ...) -> None: ...

class TelemetryRestored(_message.Message):
    __slots__ = ("session_id", "run_id", "lost_at", "gap_ms", "at", "lost_phase")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    LOST_AT_FIELD_NUMBER: _ClassVar[int]
    GAP_MS_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    LOST_PHASE_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    run_id: str
    lost_at: int
    gap_ms: int
    at: int
    lost_phase: _frame_pb2.Phase
    def __init__(self, session_id: _Optional[str] = ..., run_id: _Optional[str] = ..., lost_at: _Optional[int] = ..., gap_ms: _Optional[int] = ..., at: _Optional[int] = ..., lost_phase: _Optional[_Union[_frame_pb2.Phase, str]] = ...) -> None: ...
