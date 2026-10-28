// Package ce is how every Go component encodes and decodes a contract message
// and its CloudEvents headers (ADR 0002 in paddock, points 2 to 4), so they all
// do it alike: protobuf's JSON form, unknown fields ignored on read, and the
// headers set by hand in the NATS binding's binary mode.
//
// A producer encodes:
//
//	h, data, err := ce.Encode(lap, ce.Event{ID: src + ":" + sid + ":lap.completed:3", Source: "/twr/ecu/gt7", Time: received})
//	msg := &nats.Msg{Subject: subject, Header: nats.Header(h), Data: data}
//
// and a consumer decodes, checking the ce-type before it reads the payload:
//
//	var lap eventsv1.LapCompleted
//	event, err := ce.Decode(ce.Header(msg.Header), msg.Data, &lap)
//
// Header has the shape of nats.Header, so the package needs no NATS client, and
// it keeps no state.
package ce

import (
	"errors"
	"fmt"
	"maps"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	eventsv1 "github.com/those-who-run/contracts/gen/go/twr/events/v1"
	telemetryv1 "github.com/those-who-run/contracts/gen/go/twr/telemetry/v1"
)

// The headers a message carries (ADR 0002, point 4). The CloudEvents NATS
// binding calls header names case-insensitive, but nats.go's Header looks them
// up by exact match, so these are the exact keys, written and read as they are.
const (
	SpecVersionHeader = "ce-specversion"
	IDHeader          = "ce-id"
	TypeHeader        = "ce-type"
	SourceHeader      = "ce-source"
	TimeHeader        = "ce-time"
	ContentTypeHeader = "ce-datacontenttype"
	MsgIDHeader       = "Nats-Msg-Id"
)

const (
	// SpecVersion is the CloudEvents version the headers follow.
	SpecVersion = "1.0"
	// ContentType is the payload's encoding: protobuf's JSON form.
	ContentType = "application/json"
)

// Header is a NATS message's headers, the same type as nats.Header: convert
// with nats.Header(h) and ce.Header(msg.Header).
type Header map[string][]string

// Get is the header's first value, or "".
func (h Header) Get(key string) string {
	if v := h[key]; len(v) > 0 {
		return v[0]
	}
	return ""
}

// Event is what the headers say about a message besides its type, which comes
// from the message itself.
type Event struct {
	// ID is the event's deterministic id, the ce-id and the Nats-Msg-Id, built
	// as subjects.yaml's msg_id scheme says, such as
	// `<src>:<sid>:lap.completed:<n>`.
	ID string
	// Source is the producer: `/twr/<component>`, with the game after it only
	// where the producer is game-specific, such as `/twr/ecu/gt7`.
	Source string
	// Time is when it happened; ce-time is written in UTC.
	Time time.Time
}

// types is each message's ce-type: the one place it is written in Go. A test
// checks it against subjects.yaml.
var types = map[protoreflect.FullName]string{
	(&telemetryv1.Frame{}).ProtoReflect().Descriptor().FullName():            "run.thosewho.frame.gt7.v1",
	(&eventsv1.SessionStarted{}).ProtoReflect().Descriptor().FullName():      "run.thosewho.session.started.v1",
	(&eventsv1.LapCompleted{}).ProtoReflect().Descriptor().FullName():        "run.thosewho.lap.completed.v1",
	(&eventsv1.SessionEnded{}).ProtoReflect().Descriptor().FullName():        "run.thosewho.session.ended.v1",
	(&eventsv1.SessionContinued{}).ProtoReflect().Descriptor().FullName():    "run.thosewho.session.continued.v1",
	(&eventsv1.SessionDiscarded{}).ProtoReflect().Descriptor().FullName():    "run.thosewho.session.discarded.v1",
	(&eventsv1.SessionKindChanged{}).ProtoReflect().Descriptor().FullName():  "run.thosewho.session.kind.v1",
	(&eventsv1.SessionTrackChanged{}).ProtoReflect().Descriptor().FullName(): "run.thosewho.session.track.v1",
	(&eventsv1.SessionPaused{}).ProtoReflect().Descriptor().FullName():       "run.thosewho.session.paused.v1",
	(&eventsv1.SessionResumed{}).ProtoReflect().Descriptor().FullName():      "run.thosewho.session.resumed.v1",
	(&eventsv1.LapBestChanged{}).ProtoReflect().Descriptor().FullName():      "run.thosewho.lap.best.v1",
	(&eventsv1.RaceFinalLapStarted{}).ProtoReflect().Descriptor().FullName(): "run.thosewho.race.final_lap.v1",
	(&eventsv1.RacePositionChanged{}).ProtoReflect().Descriptor().FullName(): "run.thosewho.race.position.v1",
	(&eventsv1.RaceFinished{}).ProtoReflect().Descriptor().FullName():        "run.thosewho.race.finished.v1",
	(&eventsv1.PitEntered{}).ProtoReflect().Descriptor().FullName():          "run.thosewho.pit.entered.v1",
	(&eventsv1.PitExited{}).ProtoReflect().Descriptor().FullName():           "run.thosewho.pit.exited.v1",
	(&eventsv1.PitTeleport{}).ProtoReflect().Descriptor().FullName():         "run.thosewho.pit.teleport.v1",
	(&eventsv1.CarAlarmRaised{}).ProtoReflect().Descriptor().FullName():      "run.thosewho.car.alarm.v1",
	(&eventsv1.TelemetryLost{}).ProtoReflect().Descriptor().FullName():       "run.thosewho.telemetry.lost.v1",
	(&eventsv1.TelemetryRestored{}).ProtoReflect().Descriptor().FullName():   "run.thosewho.telemetry.restored.v1",
}

// Types is every message's ce-type, by the message's full name.
func Types() map[protoreflect.FullName]string { return maps.Clone(types) }

// Type is the message's ce-type, such as run.thosewho.lap.completed.v1.
func Type(m proto.Message) (string, error) {
	name := m.ProtoReflect().Descriptor().FullName()
	t, ok := types[name]
	if !ok {
		return "", fmt.Errorf("%s has no ce-type: it is not a message that travels on its own", name)
	}
	return t, nil
}

// Marshal encodes the message as protobuf's JSON form, with protobuf's own
// field names in lowerCamelCase, as every language's encoder writes them.
func Marshal(m proto.Message) ([]byte, error) {
	return protojson.Marshal(m)
}

// readers ignore unknown fields, so a field a later release adds never breaks
// a consumer (ADR 0002, point 3).
var readers = protojson.UnmarshalOptions{DiscardUnknown: true}

// Unmarshal decodes protobuf's JSON form into the message, ignoring fields it
// doesn't know.
func Unmarshal(data []byte, m proto.Message) error {
	return readers.Unmarshal(data, m)
}

// Encode is the headers and payload of a message: every CloudEvents header,
// with Nats-Msg-Id the same as ce-id.
func Encode(m proto.Message, e Event) (Header, []byte, error) {
	t, err := Type(m)
	if err != nil {
		return nil, nil, err
	}
	switch {
	case e.ID == "":
		return nil, nil, errors.New("the event has no id")
	case e.Source == "":
		return nil, nil, errors.New("the event has no source")
	case e.Time.IsZero():
		return nil, nil, errors.New("the event has no time")
	}
	for _, f := range [...]struct{ name, value string }{{"id", e.ID}, {"source", e.Source}} {
		if !plain(f.value) {
			return nil, nil, fmt.Errorf("the event's %s %q has a character outside U+0021 to U+007E, or a %%: "+
				"the NATS binding would need it percent-encoded", f.name, f.value)
		}
	}
	data, err := Marshal(m)
	if err != nil {
		return nil, nil, fmt.Errorf("encoding %s: %w", t, err)
	}
	return Header{
		SpecVersionHeader: {SpecVersion},
		IDHeader:          {e.ID},
		TypeHeader:        {t},
		SourceHeader:      {e.Source},
		TimeHeader:        {e.Time.UTC().Format(time.RFC3339Nano)},
		ContentTypeHeader: {ContentType},
		MsgIDHeader:       {e.ID},
	}, data, nil
}

// plain is whether s needs no percent-encoding in a header, by the CloudEvents
// NATS binding: every character printable ASCII, U+0021 to U+007E, and none of
// them %. Refusing the rest keeps every producer out of the encoding.
func plain(s string) bool {
	for _, r := range s {
		if r < 0x21 || r > 0x7e || r == '%' {
			return false
		}
	}
	return true
}

// ErrWrongType is Decode's error for a message whose ce-type is not the one
// the caller expects, checked before the payload is read.
var ErrWrongType = errors.New("wrong ce-type")

// Decode checks the headers, the ce-type against m's, and decodes the payload
// into m, ignoring unknown fields. It returns what the headers say.
func Decode(h Header, data []byte, m proto.Message) (Event, error) {
	want, err := Type(m)
	if err != nil {
		return Event{}, err
	}
	if got := h.Get(TypeHeader); got != want {
		return Event{}, fmt.Errorf("%w: %s is %q, want %q", ErrWrongType, TypeHeader, got, want)
	}
	if got := h.Get(SpecVersionHeader); got != SpecVersion {
		return Event{}, fmt.Errorf("%s is %q, want %q", SpecVersionHeader, got, SpecVersion)
	}
	if got := h.Get(ContentTypeHeader); got != ContentType {
		return Event{}, fmt.Errorf("%s is %q, want %q", ContentTypeHeader, got, ContentType)
	}
	e := Event{ID: h.Get(IDHeader), Source: h.Get(SourceHeader)}
	if e.ID == "" || e.Source == "" {
		return Event{}, fmt.Errorf("%s or %s is missing", IDHeader, SourceHeader)
	}
	if e.Time, err = time.Parse(time.RFC3339Nano, h.Get(TimeHeader)); err != nil {
		return Event{}, fmt.Errorf("%s: %w", TimeHeader, err)
	}
	if err := Unmarshal(data, m); err != nil {
		return Event{}, fmt.Errorf("decoding %s: %w", want, err)
	}
	return e, nil
}
