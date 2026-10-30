import datetime

from google.protobuf import timestamp_pb2 as _timestamp_pb2
from twr.events.v1 import session_pb2 as _session_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
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

class RadioSink(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    RADIO_SINK_UNSPECIFIED: _ClassVar[RadioSink]
    RADIO_SINK_CONSOLE: _ClassVar[RadioSink]
    RADIO_SINK_WAV: _ClassVar[RadioSink]
    RADIO_SINK_DISCORD: _ClassVar[RadioSink]
ACTIVITY_UNSPECIFIED: Activity
ACTIVITY_IDLE: Activity
ACTIVITY_DRIVING: Activity
ACTIVITY_PAUSED: Activity
ACTIVITY_REPLAY: Activity
ACTIVITY_DEMO: Activity
RADIO_SINK_UNSPECIFIED: RadioSink
RADIO_SINK_CONSOLE: RadioSink
RADIO_SINK_WAV: RadioSink
RADIO_SINK_DISCORD: RadioSink

class EcuStatus(_message.Message):
    __slots__ = ("source", "run_id", "version", "started_at", "activity", "idle", "idle_secs", "telemetry_connected", "telemetry_listening", "last_packet_age_secs", "packets", "races_started", "updated_at")
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    VERSION_FIELD_NUMBER: _ClassVar[int]
    STARTED_AT_FIELD_NUMBER: _ClassVar[int]
    ACTIVITY_FIELD_NUMBER: _ClassVar[int]
    IDLE_FIELD_NUMBER: _ClassVar[int]
    IDLE_SECS_FIELD_NUMBER: _ClassVar[int]
    TELEMETRY_CONNECTED_FIELD_NUMBER: _ClassVar[int]
    TELEMETRY_LISTENING_FIELD_NUMBER: _ClassVar[int]
    LAST_PACKET_AGE_SECS_FIELD_NUMBER: _ClassVar[int]
    PACKETS_FIELD_NUMBER: _ClassVar[int]
    RACES_STARTED_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    source: str
    run_id: str
    version: str
    started_at: _timestamp_pb2.Timestamp
    activity: Activity
    idle: bool
    idle_secs: float
    telemetry_connected: bool
    telemetry_listening: bool
    last_packet_age_secs: float
    packets: int
    races_started: int
    updated_at: _timestamp_pb2.Timestamp
    def __init__(self, source: _Optional[str] = ..., run_id: _Optional[str] = ..., version: _Optional[str] = ..., started_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., activity: _Optional[_Union[Activity, str]] = ..., idle: _Optional[bool] = ..., idle_secs: _Optional[float] = ..., telemetry_connected: _Optional[bool] = ..., telemetry_listening: _Optional[bool] = ..., last_packet_age_secs: _Optional[float] = ..., packets: _Optional[int] = ..., races_started: _Optional[int] = ..., updated_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

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
    __slots__ = ("session_id", "session_type", "track", "car_name", "lap", "total_laps", "position", "total_cars")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    SESSION_TYPE_FIELD_NUMBER: _ClassVar[int]
    TRACK_FIELD_NUMBER: _ClassVar[int]
    CAR_NAME_FIELD_NUMBER: _ClassVar[int]
    LAP_FIELD_NUMBER: _ClassVar[int]
    TOTAL_LAPS_FIELD_NUMBER: _ClassVar[int]
    POSITION_FIELD_NUMBER: _ClassVar[int]
    TOTAL_CARS_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    session_type: _session_pb2.SessionType
    track: LiveTrack
    car_name: str
    lap: int
    total_laps: int
    position: int
    total_cars: int
    def __init__(self, session_id: _Optional[str] = ..., session_type: _Optional[_Union[_session_pb2.SessionType, str]] = ..., track: _Optional[_Union[LiveTrack, _Mapping]] = ..., car_name: _Optional[str] = ..., lap: _Optional[int] = ..., total_laps: _Optional[int] = ..., position: _Optional[int] = ..., total_cars: _Optional[int] = ...) -> None: ...

class PitState(_message.Message):
    __slots__ = ("session_id", "in_pit_lane", "stint", "stops")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    IN_PIT_LANE_FIELD_NUMBER: _ClassVar[int]
    STINT_FIELD_NUMBER: _ClassVar[int]
    STOPS_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    in_pit_lane: bool
    stint: int
    stops: int
    def __init__(self, session_id: _Optional[str] = ..., in_pit_lane: _Optional[bool] = ..., stint: _Optional[int] = ..., stops: _Optional[int] = ...) -> None: ...

class FuelState(_message.Message):
    __slots__ = ("session_id", "laps_remaining", "fuel_level", "fuel_capacity")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    LAPS_REMAINING_FIELD_NUMBER: _ClassVar[int]
    FUEL_LEVEL_FIELD_NUMBER: _ClassVar[int]
    FUEL_CAPACITY_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    laps_remaining: float
    fuel_level: float
    fuel_capacity: float
    def __init__(self, session_id: _Optional[str] = ..., laps_remaining: _Optional[float] = ..., fuel_level: _Optional[float] = ..., fuel_capacity: _Optional[float] = ...) -> None: ...

class RadioStatus(_message.Message):
    __slots__ = ("source", "sink", "voice_connected", "voice_listeners", "last_spoken_at", "updated_at")
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    SINK_FIELD_NUMBER: _ClassVar[int]
    VOICE_CONNECTED_FIELD_NUMBER: _ClassVar[int]
    VOICE_LISTENERS_FIELD_NUMBER: _ClassVar[int]
    LAST_SPOKEN_AT_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    source: str
    sink: RadioSink
    voice_connected: bool
    voice_listeners: int
    last_spoken_at: _timestamp_pb2.Timestamp
    updated_at: _timestamp_pb2.Timestamp
    def __init__(self, source: _Optional[str] = ..., sink: _Optional[_Union[RadioSink, str]] = ..., voice_connected: _Optional[bool] = ..., voice_listeners: _Optional[int] = ..., last_spoken_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., updated_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...
