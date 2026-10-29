import datetime

from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class PitStop(_message.Message):
    __slots__ = ("stop", "in_lap", "out_lap", "pit_lane_ms", "stationary_ms", "fuel_before", "fuel_after", "fuel_added", "tyres_changed", "pit_lane_kph", "position_before", "position_after", "evidence")
    STOP_FIELD_NUMBER: _ClassVar[int]
    IN_LAP_FIELD_NUMBER: _ClassVar[int]
    OUT_LAP_FIELD_NUMBER: _ClassVar[int]
    PIT_LANE_MS_FIELD_NUMBER: _ClassVar[int]
    STATIONARY_MS_FIELD_NUMBER: _ClassVar[int]
    FUEL_BEFORE_FIELD_NUMBER: _ClassVar[int]
    FUEL_AFTER_FIELD_NUMBER: _ClassVar[int]
    FUEL_ADDED_FIELD_NUMBER: _ClassVar[int]
    TYRES_CHANGED_FIELD_NUMBER: _ClassVar[int]
    PIT_LANE_KPH_FIELD_NUMBER: _ClassVar[int]
    POSITION_BEFORE_FIELD_NUMBER: _ClassVar[int]
    POSITION_AFTER_FIELD_NUMBER: _ClassVar[int]
    EVIDENCE_FIELD_NUMBER: _ClassVar[int]
    stop: int
    in_lap: int
    out_lap: int
    pit_lane_ms: int
    stationary_ms: int
    fuel_before: float
    fuel_after: float
    fuel_added: float
    tyres_changed: bool
    pit_lane_kph: float
    position_before: int
    position_after: int
    evidence: str
    def __init__(self, stop: _Optional[int] = ..., in_lap: _Optional[int] = ..., out_lap: _Optional[int] = ..., pit_lane_ms: _Optional[int] = ..., stationary_ms: _Optional[int] = ..., fuel_before: _Optional[float] = ..., fuel_after: _Optional[float] = ..., fuel_added: _Optional[float] = ..., tyres_changed: _Optional[bool] = ..., pit_lane_kph: _Optional[float] = ..., position_before: _Optional[int] = ..., position_after: _Optional[int] = ..., evidence: _Optional[str] = ...) -> None: ...

class PitEntered(_message.Message):
    __slots__ = ("session_id", "entry", "stop", "in_lap", "entered_t", "entered_at", "fuel_before", "position_before", "evidence", "at")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    ENTRY_FIELD_NUMBER: _ClassVar[int]
    STOP_FIELD_NUMBER: _ClassVar[int]
    IN_LAP_FIELD_NUMBER: _ClassVar[int]
    ENTERED_T_FIELD_NUMBER: _ClassVar[int]
    ENTERED_AT_FIELD_NUMBER: _ClassVar[int]
    FUEL_BEFORE_FIELD_NUMBER: _ClassVar[int]
    POSITION_BEFORE_FIELD_NUMBER: _ClassVar[int]
    EVIDENCE_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    entry: int
    stop: int
    in_lap: int
    entered_t: float
    entered_at: _timestamp_pb2.Timestamp
    fuel_before: float
    position_before: int
    evidence: str
    at: int
    def __init__(self, session_id: _Optional[str] = ..., entry: _Optional[int] = ..., stop: _Optional[int] = ..., in_lap: _Optional[int] = ..., entered_t: _Optional[float] = ..., entered_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., fuel_before: _Optional[float] = ..., position_before: _Optional[int] = ..., evidence: _Optional[str] = ..., at: _Optional[int] = ...) -> None: ...

class PitExited(_message.Message):
    __slots__ = ("session_id", "stop", "stint", "entry", "entered_at", "exited_at", "at", "exited_t", "entered_t")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    STOP_FIELD_NUMBER: _ClassVar[int]
    STINT_FIELD_NUMBER: _ClassVar[int]
    ENTRY_FIELD_NUMBER: _ClassVar[int]
    ENTERED_AT_FIELD_NUMBER: _ClassVar[int]
    EXITED_AT_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    EXITED_T_FIELD_NUMBER: _ClassVar[int]
    ENTERED_T_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    stop: PitStop
    stint: int
    entry: int
    entered_at: _timestamp_pb2.Timestamp
    exited_at: _timestamp_pb2.Timestamp
    at: int
    exited_t: float
    entered_t: float
    def __init__(self, session_id: _Optional[str] = ..., stop: _Optional[_Union[PitStop, _Mapping]] = ..., stint: _Optional[int] = ..., entry: _Optional[int] = ..., entered_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., exited_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., at: _Optional[int] = ..., exited_t: _Optional[float] = ..., entered_t: _Optional[float] = ...) -> None: ...

class PitTeleport(_message.Message):
    __slots__ = ("session_id", "entry", "entry_at", "lap", "evidence", "at", "withdrawn_t")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    ENTRY_FIELD_NUMBER: _ClassVar[int]
    ENTRY_AT_FIELD_NUMBER: _ClassVar[int]
    LAP_FIELD_NUMBER: _ClassVar[int]
    EVIDENCE_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    WITHDRAWN_T_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    entry: int
    entry_at: int
    lap: int
    evidence: str
    at: int
    withdrawn_t: float
    def __init__(self, session_id: _Optional[str] = ..., entry: _Optional[int] = ..., entry_at: _Optional[int] = ..., lap: _Optional[int] = ..., evidence: _Optional[str] = ..., at: _Optional[int] = ..., withdrawn_t: _Optional[float] = ...) -> None: ...
