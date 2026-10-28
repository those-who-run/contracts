package ce_test

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/those-who-run/contracts/ce"
	eventsv1 "github.com/those-who-run/contracts/gen/go/twr/events/v1"
	telemetryv1 "github.com/those-who-run/contracts/gen/go/twr/telemetry/v1"
	"github.com/those-who-run/contracts/internal/catalogue"
)

var (
	repo     = os.DirFS("..")
	testdata = os.DirFS("testdata")
	at       = time.Date(2026, 10, 26, 19, 30, 5, 250_000_000, time.UTC)
)

// Each v0.1 message, from its example under examples/, with an event for it.
var messages = []struct {
	example string
	msg     proto.Message
	event   ce.Event
}{
	{"telemetry/v1/frame.json", &telemetryv1.Frame{}, ce.Event{ID: "01JB8Z4W7Q2M9T3XK5R6N8PDHF:frame:18234", Source: "/twr/ecu/gt7", Time: at}},
	{"events/v1/session_started.json", &eventsv1.SessionStarted{}, ce.Event{ID: "live:01JB8Z3K5Q9T2M7XW4R6N1PDHE:session.started:1", Source: "/twr/ecu/gt7", Time: at}},
	{"events/v1/lap_completed.json", &eventsv1.LapCompleted{}, ce.Event{ID: "live:01JB8Z3K5Q9T2M7XW4R6N1PDHE:lap.completed:3", Source: "/twr/ecu/gt7", Time: at}},
	{"events/v1/session_ended.json", &eventsv1.SessionEnded{}, ce.Event{ID: "live:01JB8Z3K5Q9T2M7XW4R6N1PDHE:session.ended:1", Source: "/twr/ecu/gt7", Time: at}},
}

// example is the message's example, decoded.
func example(t *testing.T, name string, empty proto.Message) proto.Message {
	t.Helper()
	data, err := fs.ReadFile(repo, "examples/"+name)
	if err != nil {
		t.Fatal(err)
	}
	m := empty.ProtoReflect().New().Interface()
	if err := ce.Unmarshal(data, m); err != nil {
		t.Fatal(err)
	}
	return m
}

// Encoding then decoding each v0.1 message gives back its values and its event.
func TestEachMessageRoundTrips(t *testing.T) {
	for _, tc := range messages {
		t.Run(tc.example, func(t *testing.T) {
			want := example(t, tc.example, tc.msg)
			h, data, err := ce.Encode(want, tc.event)
			if err != nil {
				t.Fatal(err)
			}
			got := tc.msg.ProtoReflect().New().Interface()
			event, err := ce.Decode(h, data, got)
			if err != nil {
				t.Fatal(err)
			}
			if !proto.Equal(got, want) {
				t.Errorf("got %v\nwant %v", got, want)
			}
			if event.ID != tc.event.ID || event.Source != tc.event.Source || !event.Time.Equal(tc.event.Time) {
				t.Errorf("event %+v, want %+v", event, tc.event)
			}
		})
	}
}

var rfc3339UTC = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d+)?Z$`)

// Every header is there and well formed, and Nats-Msg-Id is ce-id.
func TestTheHeaders(t *testing.T) {
	local := time.FixedZone("UTC+2", 2*60*60)
	lap := &eventsv1.LapCompleted{SessionId: "01JB8Z3K5Q9T2M7XW4R6N1PDHE", Lap: 3}
	h, _, err := ce.Encode(lap, ce.Event{ID: "live:01JB8Z3K5Q9T2M7XW4R6N1PDHE:lap.completed:3", Source: "/twr/ecu/gt7", Time: at.In(local)})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"ce-specversion":     "1.0",
		"ce-id":              "live:01JB8Z3K5Q9T2M7XW4R6N1PDHE:lap.completed:3",
		"ce-type":            "run.thosewho.lap.completed.v1",
		"ce-source":          "/twr/ecu/gt7",
		"ce-time":            "2026-10-26T19:30:05.25Z",
		"ce-datacontenttype": "application/json",
		"Nats-Msg-Id":        "live:01JB8Z3K5Q9T2M7XW4R6N1PDHE:lap.completed:3",
	}
	if len(h) != len(want) {
		t.Errorf("%d headers, want %d: %v", len(h), len(want), h)
	}
	for key, value := range want {
		if got := h[key]; len(got) != 1 || got[0] != value {
			t.Errorf("%s is %q, want %q", key, got, value)
		}
	}
	if !rfc3339UTC.MatchString(h.Get(ce.TimeHeader)) {
		t.Errorf("ce-time %q is not RFC 3339 in UTC", h.Get(ce.TimeHeader))
	}
}

// A payload with a field the reader doesn't know still decodes.
func TestAnUnknownFieldIsIgnored(t *testing.T) {
	want := &eventsv1.LapCompleted{SessionId: "01JB8Z3K5Q9T2M7XW4R6N1PDHE", Lap: 3, LapTimeMs: 98765, Complete: true}
	h, data, err := ce.Encode(want, messages[2].event)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	fields["stint"] = 2
	if data, err = json.Marshal(fields); err != nil {
		t.Fatal(err)
	}
	got := &eventsv1.LapCompleted{}
	if _, err := ce.Decode(h, data, got); err != nil {
		t.Fatalf("decoding with an unknown field: %v", err)
	}
	if !proto.Equal(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
}

// The payload's field names are protobuf's lowerCamelCase JSON names, the
// decision #26 made: a switch to UseProtoNames (packet_id) fails here.
func TestFieldNamesAreLowerCamelCase(t *testing.T) {
	data, err := ce.Marshal(&telemetryv1.Frame{PacketId: 512877})
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatal(err)
	}
	if _, ok := fields["packetId"]; !ok {
		t.Errorf("no packetId in %s", data)
	}
}

// A message whose ce-type is not the expected one is refused before its payload
// is read.
func TestAMismatchedTypeIsRefused(t *testing.T) {
	h, data, err := ce.Encode(&eventsv1.SessionStarted{SessionId: "01JB8Z3K5Q9T2M7XW4R6N1PDHE"}, messages[1].event)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ce.Decode(h, data, &eventsv1.SessionEnded{})
	if !errors.Is(err, ce.ErrWrongType) {
		t.Fatalf("got %v, want ErrWrongType", err)
	}
	if !strings.Contains(err.Error(), `"run.thosewho.session.started.v1", want "run.thosewho.session.ended.v1"`) {
		t.Errorf("the error doesn't name both types: %v", err)
	}
	// A body that isn't JSON at all still gets ErrWrongType: the type is
	// checked before the payload is parsed.
	if _, err := ce.Decode(h, []byte("\x00not json"), &eventsv1.SessionEnded{}); !errors.Is(err, ce.ErrWrongType) {
		t.Errorf("with a body that isn't JSON: got %v, want ErrWrongType", err)
	}
}

// Headers that are missing or not what the reader expects are refused.
func TestBadHeadersAreRefused(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(ce.Header)
		want   string
	}{
		{"no ce-type", func(h ce.Header) { delete(h, ce.TypeHeader) }, "wrong ce-type"},
		{"another spec version", func(h ce.Header) { h[ce.SpecVersionHeader] = []string{"0.3"} }, `ce-specversion is "0.3"`},
		{"binary protobuf", func(h ce.Header) { h[ce.ContentTypeHeader] = []string{"application/protobuf"} }, `ce-datacontenttype is "application/protobuf"`},
		{"no ce-id", func(h ce.Header) { delete(h, ce.IDHeader) }, "ce-id or ce-source is missing"},
		{"no ce-source", func(h ce.Header) { delete(h, ce.SourceHeader) }, "ce-id or ce-source is missing"},
		{"a time that isn't RFC 3339", func(h ce.Header) { h[ce.TimeHeader] = []string{"26/10/2026 19:30"} }, "ce-time"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h, data, err := ce.Encode(&eventsv1.LapCompleted{Lap: 3}, messages[2].event)
			if err != nil {
				t.Fatal(err)
			}
			tc.change(h)
			if _, err := ce.Decode(h, data, &eventsv1.LapCompleted{}); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("got %v, want an error with %q", err, tc.want)
			}
		})
	}
	h, _, err := ce.Encode(&eventsv1.LapCompleted{Lap: 3}, messages[2].event)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ce.Decode(h, []byte(`{"lap": "three"}`), &eventsv1.LapCompleted{}); err == nil {
		t.Error("a payload that isn't the message decoded")
	}
}

// Encode refuses an event it can't write every header for, and a message with
// no ce-type.
func TestEncodeRefuses(t *testing.T) {
	lap := &eventsv1.LapCompleted{Lap: 3}
	for _, tc := range []struct {
		name  string
		msg   proto.Message
		event ce.Event
		want  string
	}{
		{"no id", lap, ce.Event{Source: "/twr/ecu/gt7", Time: at}, "no id"},
		{"no source", lap, ce.Event{ID: "x", Time: at}, "no source"},
		{"no time", lap, ce.Event{ID: "x", Source: "/twr/ecu/gt7"}, "no time"},
		{"a space in the id", lap, ce.Event{ID: "lap 3", Source: "/twr/ecu/gt7", Time: at}, `id "lap 3" has a character outside`},
		{"a % in the id", lap, ce.Event{ID: "lap%203", Source: "/twr/ecu/gt7", Time: at}, `id "lap%203"`},
		{"a non-ASCII source", lap, ce.Event{ID: "x", Source: "/twr/écu", Time: at}, `source "/twr/écu" has a character outside`},
		{"a control character in the source", lap, ce.Event{ID: "x", Source: "/twr/ecu\n", Time: at}, "source"},
		{"a part of a message", &eventsv1.SessionSummary{}, messages[3].event, "twr.events.v1.SessionSummary has no ce-type"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, _, err := ce.Encode(tc.msg, tc.event); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("got %v, want an error with %q", err, tc.want)
			}
		})
	}
	if _, err := ce.Decode(ce.Header{}, nil, &eventsv1.SessionSummary{}); err == nil {
		t.Error("decoding into a message with no ce-type passed")
	}
}

func load(t *testing.T, fsys fs.FS, name string) *catalogue.Catalogue {
	t.Helper()
	f, err := fsys.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	c, err := catalogue.Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// The package's ce-types are subjects.yaml's (paddock#25), and a catalogue that
// disagrees fails.
func TestTypesAgreeWithTheCatalogue(t *testing.T) {
	if err := load(t, repo, "subjects.yaml").Agrees(ce.Types()); err != nil {
		t.Fatal(err)
	}
	err := load(t, testdata, "other-ce-type.yaml").Agrees(ce.Types())
	if err == nil || !strings.Contains(err.Error(),
		"twr.events.v1.LapCompleted: the catalogue's ce-type is run.thosewho.lap.completed.v2, the helper's run.thosewho.lap.completed.v1") {
		t.Errorf("got %v, want the disagreement on LapCompleted", err)
	}
	types := ce.Types()
	delete(types, "twr.events.v1.SessionEnded")
	types["twr.events.v1.PitEntered"] = "run.thosewho.pit.entered.v1"
	err = load(t, repo, "subjects.yaml").Agrees(types)
	for _, want := range []string{
		"twr.events.v1.SessionEnded (run.thosewho.session.ended.v1) has no ce-type in the helper",
		"twr.events.v1.PitEntered (run.thosewho.pit.entered.v1) is in the helper but not the catalogue",
	} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("got %v, want an error with %q", err, want)
		}
	}
}

// BenchmarkEncodeFrame is the cost of one frame's payload and headers, which ecu
// pays about 60 times a second (paddock#26, R5).
func BenchmarkEncodeFrame(b *testing.B) {
	data, err := fs.ReadFile(repo, "examples/telemetry/v1/frame.json")
	if err != nil {
		b.Fatal(err)
	}
	frame := &telemetryv1.Frame{}
	if err := ce.Unmarshal(data, frame); err != nil {
		b.Fatal(err)
	}
	event := messages[0].event
	b.ReportAllocs()
	for b.Loop() {
		if _, _, err := ce.Encode(frame, event); err != nil {
			b.Fatal(err)
		}
	}
}
