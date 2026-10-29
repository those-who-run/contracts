//go:build python

// The Python leg of the round trip (paddock#265), which needs a Python with
// gen/python installed: `make python` builds one and runs this file, which the
// build tag keeps out of a plain `go test ./...`.

package contracts_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/those-who-run/contracts/callout"
	"github.com/those-who-run/contracts/ce"
	eventsv1 "github.com/those-who-run/contracts/gen/go/twr/events/v1"
)

type pyExample struct {
	File    string `json:"file"`
	Message string `json:"message,omitempty"`
	Go      string `json:"go,omitempty"`
	Python  string `json:"python,omitempty"`
}

type pyEvent struct {
	Message string            `json:"message"`
	Headers map[string]string `json:"headers"`
	Data    string            `json:"data"`
	ID      string            `json:"id,omitempty"`
	Source  string            `json:"source,omitempty"`
	Time    string            `json:"time,omitempty"`
}

type pyLevel struct {
	Priority int32   `json:"priority"`
	Level    string  `json:"level"`
	Subject  *string `json:"subject"`
}

type pyRequest struct {
	Examples   []pyExample `json:"examples"`
	Events     []pyEvent   `json:"events"`
	Priorities []int32     `json:"priorities"`
}

type pyResponse struct {
	Types    map[string]string `json:"types"`
	Examples []pyExample       `json:"examples"`
	Events   []pyEvent         `json:"events"`
	Levels   []pyLevel         `json:"levels"`
}

// python runs scripts/python_roundtrip.py on the request with $TWR_PYTHON, the
// Python `make python` installed gen/python into.
func python(t *testing.T, req pyRequest) pyResponse {
	t.Helper()
	interpreter := os.Getenv("TWR_PYTHON")
	if interpreter == "" {
		t.Fatal("TWR_PYTHON is not set: run the Python leg with `make python`")
	}
	data, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "request.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(interpreter, "-I", filepath.Join("scripts", "python_roundtrip.py"), path) //nolint:gosec // the Python make set up, on a file the test wrote
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("the Python leg: %v", err)
	}
	var resp pyResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		t.Fatalf("the Python leg's output: %v\n%s", err, out)
	}
	return resp
}

// Every example goes from Go to Python and back: Python decodes what Go's ce
// encoded to the example's values, an unknown field included, and Go decodes
// what Python's helper encodes to the values the test expects. The headers make
// the same trip through ce.Encode, Python's ce.decode and ce.encode, and
// ce.Decode. Both sides compare decoded values, never JSON text (paddock#265
// AC3 and AC4: the callouts, among them ones with no text and with a kind and
// slots, and v1's Gilles Villeneuve stop in pit_exited.json).
func TestPythonRoundTrip(t *testing.T) {
	// Python's datetime holds microseconds, so the events' times stop there.
	at := time.Date(2026, 10, 29, 18, 3, 7, 588_057_000, time.UTC)
	var req pyRequest
	want := map[string]proto.Message{}
	for i, ex := range examples {
		data, err := ce.Marshal(ex.want)
		if err != nil {
			t.Fatal(err)
		}
		name := string(ex.want.ProtoReflect().Descriptor().FullName())
		req.Examples = append(req.Examples, pyExample{File: ex.file, Message: name, Go: string(data)})
		want[ex.file] = ex.want

		e := ce.Event{ID: fmt.Sprintf("live:01JBBX7Q2M5R8T1V4W6Y9K3NZC:example:%d", i), Source: "/twr/ecu/gt7", Time: at.Add(time.Duration(i) * time.Millisecond)}
		h, data, err := ce.Encode(ex.want, e)
		if err != nil {
			t.Fatal(err)
		}
		headers := map[string]string{}
		for k, v := range h {
			headers[k] = v[0]
		}
		req.Events = append(req.Events, pyEvent{
			Message: name, Headers: headers, Data: string(data),
			ID: e.ID, Source: e.Source, Time: e.Time.Format(time.RFC3339Nano),
		})
	}
	resp := python(t, req)

	if len(resp.Examples) != len(examples) || len(resp.Events) != len(examples) {
		t.Fatalf("Python answered for %d examples and %d events, want %d", len(resp.Examples), len(resp.Events), len(examples))
	}
	for _, ex := range resp.Examples {
		got := want[ex.File].ProtoReflect().New().Interface()
		if err := ce.Unmarshal([]byte(ex.Python), got); err != nil {
			t.Errorf("%s: decoding Python's %s: %v", ex.File, ex.Python, err)
			continue
		}
		if !proto.Equal(got, want[ex.File]) {
			t.Errorf("%s: Python's JSON decodes to %v\nwant %v", ex.File, got, want[ex.File])
		}
	}
	for i, ev := range resp.Events {
		ex := examples[i]
		h := ce.Header{}
		for k, v := range ev.Headers {
			h[k] = []string{v}
		}
		got := ex.want.ProtoReflect().New().Interface()
		e, err := ce.Decode(h, []byte(ev.Data), got)
		if err != nil {
			t.Errorf("%s: decoding Python's event: %v", ex.file, err)
			continue
		}
		if !proto.Equal(got, ex.want) {
			t.Errorf("%s: Python's event decodes to %v\nwant %v", ex.file, got, ex.want)
		}
		sent := req.Events[i]
		if e.ID != sent.ID || e.Source != sent.Source || e.Time.Format(time.RFC3339Nano) != sent.Time {
			t.Errorf("%s: Python's headers say %+v, Go sent %s, %s, %s", ex.file, e, sent.ID, sent.Source, sent.Time)
		}
		if ev.Headers[ce.MsgIDHeader] != ev.Headers[ce.IDHeader] {
			t.Errorf("%s: Python's Nats-Msg-Id %q is not its ce-id %q", ex.file, ev.Headers[ce.MsgIDHeader], ev.Headers[ce.IDHeader])
		}
	}
}

// Python's ce-type table is Go's, which a test checks against subjects.yaml.
func TestPythonTypesAreGos(t *testing.T) {
	got := python(t, pyRequest{}).Types
	want := ce.Types()
	if len(got) != len(want) {
		t.Errorf("Python has %d ce-types, Go %d", len(got), len(want))
	}
	for name, typ := range want {
		if got[string(name)] != typ {
			t.Errorf("%s: Python's ce-type is %q, Go's %q", name, got[string(name)], typ)
		}
	}
}

// Python's callout helper gives every priority the level and subject Go's
// does, the edges of callout_test.go's table and the int32 extremes among them,
// and refuses the same unset priorities (paddock#265 R6).
func TestPythonCalloutLevelsAreGos(t *testing.T) {
	priorities := []int32{-2147483648, 2147483647}
	for p := int32(-2); p <= 60; p++ {
		priorities = append(priorities, p)
	}
	levels := python(t, pyRequest{Priorities: priorities}).Levels
	if len(levels) != len(priorities) {
		t.Fatalf("Python answered for %d priorities, want %d", len(levels), len(priorities))
	}
	for _, l := range levels {
		if want := callout.LevelOf(l.Priority); l.Level != string(want) {
			t.Errorf("priority %d: Python's level is %s, Go's %s", l.Priority, l.Level, want)
		}
		subject, err := callout.Subject("live", &eventsv1.Callout{Priority: l.Priority})
		switch {
		case err != nil && l.Subject != nil:
			t.Errorf("priority %d: Python's subject is %s, Go refuses it: %v", l.Priority, *l.Subject, err)
		case err == nil && l.Subject == nil:
			t.Errorf("priority %d: Python refuses it, Go's subject is %s", l.Priority, subject)
		case err == nil && *l.Subject != subject:
			t.Errorf("priority %d: Python's subject is %s, Go's %s", l.Priority, *l.Subject, subject)
		}
	}
}
