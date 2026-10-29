import datetime

from google.protobuf import timestamp_pb2 as _timestamp_pb2
from twr.events.v1 import car_pb2 as _car_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class PitLane(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    PIT_LANE_UNSPECIFIED: _ClassVar[PitLane]
    PIT_LANE_SPEAK: _ClassVar[PitLane]
    PIT_LANE_HOLD: _ClassVar[PitLane]
    PIT_LANE_DROP: _ClassVar[PitLane]
PIT_LANE_UNSPECIFIED: PitLane
PIT_LANE_SPEAK: PitLane
PIT_LANE_HOLD: PitLane
PIT_LANE_DROP: PitLane

class Callout(_message.Message):
    __slots__ = ("session_id", "id", "rule", "kind", "slots", "text", "priority", "at", "run_id", "created_at", "expires_at", "hold_until", "pit_lane")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    ID_FIELD_NUMBER: _ClassVar[int]
    RULE_FIELD_NUMBER: _ClassVar[int]
    KIND_FIELD_NUMBER: _ClassVar[int]
    SLOTS_FIELD_NUMBER: _ClassVar[int]
    TEXT_FIELD_NUMBER: _ClassVar[int]
    PRIORITY_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    CREATED_AT_FIELD_NUMBER: _ClassVar[int]
    EXPIRES_AT_FIELD_NUMBER: _ClassVar[int]
    HOLD_UNTIL_FIELD_NUMBER: _ClassVar[int]
    PIT_LANE_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    id: str
    rule: str
    kind: str
    slots: CalloutSlots
    text: str
    priority: int
    at: int
    run_id: str
    created_at: _timestamp_pb2.Timestamp
    expires_at: _timestamp_pb2.Timestamp
    hold_until: _timestamp_pb2.Timestamp
    pit_lane: PitLane
    def __init__(self, session_id: _Optional[str] = ..., id: _Optional[str] = ..., rule: _Optional[str] = ..., kind: _Optional[str] = ..., slots: _Optional[_Union[CalloutSlots, _Mapping]] = ..., text: _Optional[str] = ..., priority: _Optional[int] = ..., at: _Optional[int] = ..., run_id: _Optional[str] = ..., created_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., expires_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., hold_until: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., pit_lane: _Optional[_Union[PitLane, str]] = ...) -> None: ...

class CalloutSlots(_message.Message):
    __slots__ = ("lap", "lap_time_ms", "delta_ms", "laps", "laps_left", "position", "previous_position", "total_cars", "track", "corners", "best_lap_ms", "timed")
    LAP_FIELD_NUMBER: _ClassVar[int]
    LAP_TIME_MS_FIELD_NUMBER: _ClassVar[int]
    DELTA_MS_FIELD_NUMBER: _ClassVar[int]
    LAPS_FIELD_NUMBER: _ClassVar[int]
    LAPS_LEFT_FIELD_NUMBER: _ClassVar[int]
    POSITION_FIELD_NUMBER: _ClassVar[int]
    PREVIOUS_POSITION_FIELD_NUMBER: _ClassVar[int]
    TOTAL_CARS_FIELD_NUMBER: _ClassVar[int]
    TRACK_FIELD_NUMBER: _ClassVar[int]
    CORNERS_FIELD_NUMBER: _ClassVar[int]
    BEST_LAP_MS_FIELD_NUMBER: _ClassVar[int]
    TIMED_FIELD_NUMBER: _ClassVar[int]
    lap: int
    lap_time_ms: int
    delta_ms: int
    laps: int
    laps_left: int
    position: int
    previous_position: int
    total_cars: int
    track: str
    corners: _containers.RepeatedScalarFieldContainer[_car_pb2.Corner]
    best_lap_ms: int
    timed: bool
    def __init__(self, lap: _Optional[int] = ..., lap_time_ms: _Optional[int] = ..., delta_ms: _Optional[int] = ..., laps: _Optional[int] = ..., laps_left: _Optional[int] = ..., position: _Optional[int] = ..., previous_position: _Optional[int] = ..., total_cars: _Optional[int] = ..., track: _Optional[str] = ..., corners: _Optional[_Iterable[_Union[_car_pb2.Corner, str]]] = ..., best_lap_ms: _Optional[int] = ..., timed: _Optional[bool] = ...) -> None: ...
