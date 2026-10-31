import datetime

from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class CardState(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    CARD_STATE_UNSPECIFIED: _ClassVar[CardState]
    CARD_STATE_POSTING: _ClassVar[CardState]
    CARD_STATE_POSTED: _ClassVar[CardState]
CARD_STATE_UNSPECIFIED: CardState
CARD_STATE_POSTING: CardState
CARD_STATE_POSTED: CardState

class CardRecord(_message.Message):
    __slots__ = ("source", "session_id", "state", "at", "message_id")
    SOURCE_FIELD_NUMBER: _ClassVar[int]
    SESSION_ID_FIELD_NUMBER: _ClassVar[int]
    STATE_FIELD_NUMBER: _ClassVar[int]
    AT_FIELD_NUMBER: _ClassVar[int]
    MESSAGE_ID_FIELD_NUMBER: _ClassVar[int]
    source: str
    session_id: str
    state: CardState
    at: _timestamp_pb2.Timestamp
    message_id: str
    def __init__(self, source: _Optional[str] = ..., session_id: _Optional[str] = ..., state: _Optional[_Union[CardState, str]] = ..., at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., message_id: _Optional[str] = ...) -> None: ...
