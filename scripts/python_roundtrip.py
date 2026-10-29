"""The Python leg of the round-trip test (paddock#265), run by python_test.go
with the Python of a virtual environment that has gen/python installed, as a
consumer installs it:

    <venv>/bin/python scripts/python_roundtrip.py <request.json>

The request is what Go wrote: each example under examples/ with the JSON Go's
ce encoded from the values the test expects, each as a whole event (headers and
payload, from ce.Encode), and callout priorities. For each, this script:

- decodes the example and Go's JSON through twr_contracts.ce, and fails unless
  the two hold the same values (never comparing JSON text), and unless the
  example with a field it doesn't know, which a strict reader refuses, decodes
  to them too;
- decodes Go's event with ce.decode, fails unless the headers say what Go
  encoded, and encodes it again with ce.encode;
- checks, on the first event, that ce.encode and ce.decode refuse what Go's
  ce refuses;
- gives each priority's level and subject from twr_contracts.callout.

It prints, as JSON, its ce-type table, what Python encoded and the callout
levels, for Go to decode and compare with its own. Its errors go to standard
error, and it exits 1.
"""

from __future__ import annotations

import dataclasses
import datetime
import json
import os
import sys
from typing import Any

from google.protobuf import descriptor_pool, json_format, message_factory
from google.protobuf.message import Message
from twr.events.v1 import callout_pb2
from twr_contracts import callout, ce

EXAMPLES = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "examples")


def new(full_name: str) -> Message:
    """A new, empty message of the type."""
    desc = descriptor_pool.Default().FindMessageTypeByName(full_name)
    return message_factory.GetMessageClass(desc)()


def example(ex: dict[str, Any], failures: list[str]) -> dict[str, str]:
    with open(os.path.join(EXAMPLES, ex["file"]), encoding="utf-8") as f:
        text = f.read()
    from_example, from_go = new(ex["message"]), new(ex["message"])
    ce.unmarshal(text, from_example)
    ce.unmarshal(ex["go"], from_go)
    if from_example != from_go:
        failures.append(f"{ex['file']}: Go's JSON decodes to\n{from_go}\nthe example to\n{from_example}")

    fields = json.loads(text)
    fields["addedInALaterRelease"] = {"value": 1, "note": "unknown"}
    unknown = json.dumps(fields)
    try:
        json_format.Parse(unknown, new(ex["message"]))
        failures.append(
            f"{ex['file']}: a strict reader accepted the unknown field, so this case checks nothing"
        )
    except json_format.ParseError:
        pass
    with_unknown = new(ex["message"])
    ce.unmarshal(unknown, with_unknown)
    if with_unknown != from_example:
        failures.append(f"{ex['file']}: with an unknown field it decodes to\n{with_unknown}")
    return {"file": ex["file"], "python": ce.marshal(from_go).decode()}


def event(ev: dict[str, Any], failures: list[str]) -> dict[str, Any]:
    m = new(ev["message"])
    got = ce.decode(ev["headers"], ev["data"].encode(), m)
    want = ce.Event(id=ev["id"], source=ev["source"], time=datetime.datetime.fromisoformat(ev["time"]))
    if got != want:
        failures.append(f"{ev['message']}: Go's headers decode to {got}, want {want}")
    headers, data = ce.encode(m, got)
    return {"message": ev["message"], "headers": headers, "data": data.decode()}


# ce-time values ISO 8601 or Python's fromisoformat take and RFC 3339, so Go's
# time.Parse, refuses; python_test.go checks Go refuses each of them too.
NOT_RFC3339 = (
    "2026-10-29 18:03:07Z",
    "20261029T180307Z",
    "2026-10-29t18:03:07Z",
    "2026-10-29T18:03:07+0100",
    "2026-10-29T18:03Z",
    "2026-W44-4T18:03:07Z",
)


def refusals(ev: dict[str, Any], failures: list[str]) -> None:
    """What Go's ce refuses, Python's refuses too: an event without an id or a
    source, with a character the NATS binding would percent-encode or a time
    without a zone; headers with another ce-type (WrongTypeError), another
    spec version or content type, no id or source, or a time that isn't one."""
    m = new(ev["message"])
    good = ce.Event(id=ev["id"], source=ev["source"], time=datetime.datetime.fromisoformat(ev["time"]))
    for what, e in (
        ("no id", dataclasses.replace(good, id="")),
        ("no source", dataclasses.replace(good, source="")),
        ("a space in the id", dataclasses.replace(good, id="a b")),
        ("a % in the source", dataclasses.replace(good, source="/twr/%41")),
        ("a non-ASCII id", dataclasses.replace(good, id="café")),
        ("a time without a zone", dataclasses.replace(good, time=good.time.replace(tzinfo=None))),
    ):
        try:
            ce.encode(m, e)
            failures.append(f"encode accepted an event with {what}")
        except ValueError:
            pass
    other = "run.thosewho.lap.completed.v1" if ev["message"] != "twr.events.v1.LapCompleted" else "x"
    for what, change, error in (
        ("another ce-type", {ce.TYPE_HEADER: other}, ce.WrongTypeError),
        ("another spec version", {ce.SPEC_VERSION_HEADER: "0.3"}, ValueError),
        ("another content type", {ce.CONTENT_TYPE_HEADER: "application/protobuf"}, ValueError),
        ("no id", {ce.ID_HEADER: ""}, ValueError),
        ("no source", {ce.SOURCE_HEADER: ""}, ValueError),
        ("a time that isn't one", {ce.TIME_HEADER: "yesterday"}, ValueError),
        ("a time without a zone", {ce.TIME_HEADER: "2026-10-29T18:03:07"}, ValueError),
        *((f"the time {t!r}, which isn't RFC 3339", {ce.TIME_HEADER: t}, ValueError) for t in NOT_RFC3339),
    ):
        try:
            ce.decode({**ev["headers"], **change}, ev["data"], new(ev["message"]))
            failures.append(f"decode accepted headers with {what}")
        except error:
            pass
    try:
        ce.ce_type(callout_pb2.CalloutSlots())
        failures.append("ce_type gave a part of a message a ce-type")
    except ValueError:
        pass


def level(priority: int) -> dict[str, Any]:
    try:
        subject: str | None = callout.subject("live", callout_pb2.Callout(priority=priority))
    except ValueError:
        subject = None
    return {"priority": priority, "level": str(callout.level_of(priority)), "subject": subject}


def main(argv: list[str]) -> int:
    with open(argv[0], encoding="utf-8") as f:
        request = json.load(f)
    failures: list[str] = []
    response = {
        "types": ce.types(),
        "notRfc3339": list(NOT_RFC3339),
        "examples": [example(ex, failures) for ex in request["examples"] or []],
        "events": [event(ev, failures) for ev in request["events"] or []],
        "levels": [level(p) for p in request["priorities"] or []],
    }
    for ev in (request["events"] or [])[:1]:
        refusals(ev, failures)
    if failures:
        print("\n".join(failures), file=sys.stderr)
        return 1
    json.dump(response, sys.stdout)
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
