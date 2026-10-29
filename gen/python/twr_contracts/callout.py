"""The subject a callout travels on, from its priority (paddock#71, ADR 0001
point 2 in paddock), the copy of the Go module's package `callout`:
twr.v1.<src>.callout.p0, .p1 or .p2, so radio can take p0 before p1 and p2. The
levels keep v1's order: p0 is a priority under 10, p1 10 to 19 and p2 20 and
over, and the start brief (8) is p0.

A consumer subscribes to each level, most urgent first:

    for level in callout.LEVELS:
        await nc.subscribe(callout.level_subject("live", level), cb=handle)
"""

from __future__ import annotations

import enum

from twr.events.v1 import callout_pb2


class Level(enum.StrEnum):
    """A callout's level, the subject's last token."""

    P0 = "p0"
    P1 = "p1"
    P2 = "p2"


# Every level, most urgent first: a consumer subscribes to each.
LEVELS = (Level.P0, Level.P1, Level.P2)

# What a subject token can't hold.
_NOT_IN_A_TOKEN = frozenset(".*> \t\r\n")


def level_of(priority: int) -> Level:
    """The level of a priority, v1's number where lower is more urgent: p0 under
    10, p1 10 to 19, p2 20 and over."""
    if priority < 10:
        return Level.P0
    if priority < 20:
        return Level.P1
    return Level.P2


def level_subject(src: str, level: Level) -> str:
    """The subject of a level's callouts from src, such as
    twr.v1.live.callout.p0."""
    if not src or any(c in _NOT_IN_A_TOKEN for c in src):
        raise ValueError(f"source {src!r} is not one subject token")
    return f"twr.v1.{src}.callout.{level}"


def subject(src: str, c: callout_pb2.Callout) -> str:
    """The subject the callout travels on from src (`live`, `parity` or
    `replay-<name>`): its priority's level. A publisher sets the priority from
    1, and 0 is unset, so a priority under 1 is refused rather than sent as p0."""
    if c.priority < 1:
        raise ValueError(f"the callout's priority {c.priority} is not set: a publisher sets it from 1")
    return level_subject(src, level_of(c.priority))
