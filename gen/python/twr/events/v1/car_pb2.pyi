from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CarCheck(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    CAR_CHECK_UNSPECIFIED: _ClassVar[CarCheck]
    CAR_CHECK_TYRE_HOT: _ClassVar[CarCheck]
    CAR_CHECK_TYRES_COLD: _ClassVar[CarCheck]
    CAR_CHECK_WATER_HOT: _ClassVar[CarCheck]
    CAR_CHECK_OIL_HOT: _ClassVar[CarCheck]

class Corner(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    CORNER_UNSPECIFIED: _ClassVar[Corner]
    CORNER_FRONT_LEFT: _ClassVar[Corner]
    CORNER_FRONT_RIGHT: _ClassVar[Corner]
    CORNER_REAR_LEFT: _ClassVar[Corner]
    CORNER_REAR_RIGHT: _ClassVar[Corner]
CAR_CHECK_UNSPECIFIED: CarCheck
CAR_CHECK_TYRE_HOT: CarCheck
CAR_CHECK_TYRES_COLD: CarCheck
CAR_CHECK_WATER_HOT: CarCheck
CAR_CHECK_OIL_HOT: CarCheck
CORNER_UNSPECIFIED: Corner
CORNER_FRONT_LEFT: Corner
CORNER_FRONT_RIGHT: Corner
CORNER_REAR_LEFT: Corner
CORNER_REAR_RIGHT: Corner

class TyreTemp(_message.Message):
    __slots__ = ("corner", "temp_c")
    CORNER_FIELD_NUMBER: _ClassVar[int]
    TEMP_C_FIELD_NUMBER: _ClassVar[int]
    corner: Corner
    temp_c: float
    def __init__(self, corner: _Optional[_Union[Corner, str]] = ..., temp_c: _Optional[float] = ...) -> None: ...

class CarAlarmRaised(_message.Message):
    __slots__ = ("session_id", "check", "tyres", "value_c", "threshold_c", "at", "run_id")
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    CHECK_FIELD_NUMBER: _ClassVar[int]
    TYRES_FIELD_NUMBER: _ClassVar[int]
    VALUE_C_FIELD_NUMBER: _ClassVar[int]
    THRESHOLD_C_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    RUN_ID_FIELD_NUMBER: _ClassVar[int]
    session_id: str
    check: CarCheck
    tyres: _containers.RepeatedCompositeFieldContainer[TyreTemp]
    value_c: float
    threshold_c: float
    at: int
    run_id: str
    def __init__(self, session_id: _Optional[str] = ..., check: _Optional[_Union[CarCheck, str]] = ..., tyres: _Optional[_Iterable[_Union[TyreTemp, _Mapping]]] = ..., value_c: _Optional[float] = ..., threshold_c: _Optional[float] = ..., at: _Optional[int] = ..., run_id: _Optional[str] = ...) -> None: ...
