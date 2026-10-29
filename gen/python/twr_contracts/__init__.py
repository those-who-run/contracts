"""The helpers every Python component uses with the contracts (paddock#265).

The generated messages are under `twr`, one module per .proto file, such as
`twr.events.v1.callout_pb2`. These helpers are the Python copies of the Go
module's packages, and behave as they do:

- `twr_contracts.ce` encodes and decodes a message and its CloudEvents headers,
  as package `ce` does (ADR 0002 in paddock, points 2 to 4);
- `twr_contracts.callout` builds the subject a callout travels on from its
  priority, as package `callout` does (paddock#71).
"""
