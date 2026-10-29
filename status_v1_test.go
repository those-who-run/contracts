package contracts_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/those-who-run/contracts/ce"
	eventsv1 "github.com/those-who-run/contracts/gen/go/twr/events/v1"
	statev1 "github.com/those-who-run/contracts/gen/go/twr/state/v1"
)

// EcuStatus and LiveSession against v1's status route (paddock#72, AC2):
// testdata/v1/gilles_villeneuve_status.json is the `GET /api/v1/status` and
// `/healthz` bodies v1.0.0 answers mid-race, on lap 2 of
// real-nuc-gilles-villeneuve-race-2026-10-17.jsonl.gz, fed through the live
// app as v1's tests/test_status_api.py does. That test hands the app its
// frames with no listener of its own, so v1 says the telemetry is not
// connected and counts no packets; started_at is the clock of the run that
// made it.

// The v1 fields EcuStatus and LiveSession leave out on purpose, each with why.
var statusLeftOut = map[string]string{
	"status.voice_connected":     "radio's, in its own live key (paddock#62)",
	"healthz.voice_connected":    "radio's, in its own live key (paddock#62)",
	"healthz.ok":                 "/healthz's own, from ecu's tasks and listener (paddock#88, R4)",
	"healthz.tasks":              "/healthz's own (paddock#88, R4)",
	"healthz.telemetry_restarts": "/healthz's own: how often v1 rebound its listener",
	"healthz.queue":              "the callouts waiting, now radio's (paddock#62)",
	"healthz.session":            "the session's kind: SessionKindChanged's, and LiveSession's session_type",
	"healthz.track":              "the session's track: SessionTrackChanged's, and LiveSession's track",
	"status.session.id":          "v1's own id format; session_id is the ULID ecu mints",
}

// v2Only are the fields v1 had no counterpart for, which the test takes from
// the example. It also takes the telemetry link from the example, which shows
// a live drive: v1's test fed its app frames with no listener, so the golden
// says not connected, no packets.
func v2Only(status *statev1.EcuStatus, session *statev1.LiveSession, wantStatus *statev1.EcuStatus, wantSession *statev1.LiveSession) {
	status.Source, status.RunId, status.UpdatedAt = wantStatus.GetSource(), wantStatus.GetRunId(), wantStatus.GetUpdatedAt()
	status.TelemetryConnected, status.Packets = wantStatus.GetTelemetryConnected(), wantStatus.GetPackets()
	status.LastPacketAgeSecs = wantStatus.LastPacketAgeSecs
	session.SessionId = wantSession.GetSessionId()
}

// rest fails on a field of r the conversion did not take and statusLeftOut
// does not list.
func rest(t *testing.T, r v1Record) {
	t.Helper()
	for key := range r.m {
		if _, ok := statusLeftOut[r.prefix+key]; !ok {
			t.Errorf("v1's %s%s is neither in EcuStatus or LiveSession nor in statusLeftOut", r.prefix, key)
		}
	}
}

func activityFromV1(t *testing.T, v1 string) statev1.Activity {
	t.Helper()
	a, ok := statev1.Activity_value[upper("ACTIVITY_", v1)]
	if !ok {
		t.Fatalf("v1's activity %q has no Activity", v1)
	}
	return statev1.Activity(a)
}

// Every field of v1's EngineStatus and LiveSession is in EcuStatus or
// LiveSession, or listed in statusLeftOut, and what v1 answered converts to
// ecu_status.json and ecu_session.json, through the JSON form ce writes.
func TestStatusFromV1(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "v1", "gilles_villeneuve_status.json"))
	if err != nil {
		t.Fatal(err)
	}
	var bodies map[string]map[string]any
	if err := json.Unmarshal(data, &bodies); err != nil {
		t.Fatal(err)
	}
	// v1's kind, whose field and cars_on_track decide whether a position shows
	// (results.position_shown), read before the records take their fields.
	kindRec, _ := bodies["healthz"]["session"].(map[string]any)
	shown := kindRec["field"] == true && kindRec["cars_on_track"] == true
	status := v1Record{t, "status.", bodies["status"]}
	health := v1Record{t, "healthz.", bodies["healthz"]}

	telemetry, _ := status.object("telemetry")
	telemetry.prefix = "status.telemetry."
	got := &statev1.EcuStatus{
		Version:            status.str("version"),
		StartedAt:          status.at("started_at"),
		Activity:           activityFromV1(t, status.str("activity")),
		IdleSecs:           status.num("idle_secs"),
		TelemetryConnected: telemetry.flag("connected"),
		Idle:               health.flag("idle"),
		TelemetryListening: health.flag("telemetry_listening"),
		Packets:            uint32(health.num("packets")),
		RacesStarted:       health.i32("races_started"),
	}
	if n, ok := telemetry.take("last_packet_age_s").(float64); ok {
		got.LastPacketAgeSecs = proto.Float64(n)
	}
	// /healthz says the same as the status route, which EcuStatus says once.
	if health.num("idle_secs") != got.GetIdleSecs() || health.flag("telemetry_connected") != got.GetTelemetryConnected() {
		t.Error("v1's /healthz and status route disagree on idle_secs or the telemetry")
	}
	if got.GetIdle() != (got.GetActivity() == statev1.Activity_ACTIVITY_IDLE) {
		t.Error("idle and the activity disagree")
	}

	s, ok := status.object("session")
	if !ok {
		t.Fatal("v1's status has no session")
	}
	s.prefix = "status.session."
	kind, ok := eventsv1.SessionType_value[upper("SESSION_TYPE_", s.str("kind"))]
	if !ok {
		t.Fatal("v1's session kind has no SessionType")
	}
	session := &statev1.LiveSession{
		SessionType: eventsv1.SessionType(kind),
		CarName:     s.strOrNull("car"),
		Lap:         s.opt("lap"),
		TotalLaps:   s.opt("laps_total"),
		Position:    s.opt("position"),
		TotalCars:   s.opt("cars"),
	}
	if track, ok := s.object("track"); ok {
		track.prefix = "status.session.track."
		session.Track = &statev1.LiveTrack{TrackId: track.opt("id"), Name: track.strOrNull("name"), Sure: track.flag("sure")}
		rest(t, track)
	}

	if (session.Position != nil) != shown || (session.TotalCars != nil) != shown {
		t.Errorf("position %v and cars %v, but v1's position_shown is %v", session.Position, session.TotalCars, shown)
	}
	if !shown {
		t.Error("the golden's kind hides the position, so this checks nothing")
	}
	for _, r := range []v1Record{status, health, telemetry, s} {
		rest(t, r)
	}

	wantStatus := example(t, "state/v1/ecu_status.json", &statev1.EcuStatus{}).(*statev1.EcuStatus)
	wantSession := example(t, "state/v1/ecu_session.json", &statev1.LiveSession{}).(*statev1.LiveSession)
	v2Only(got, session, wantStatus, wantSession)
	if got.GetSource() == "" || got.GetRunId() == "" || got.GetUpdatedAt() == nil || session.GetSessionId() == "" ||
		!got.GetTelemetryConnected() || got.GetPackets() == 0 || got.LastPacketAgeSecs == nil {
		t.Error("the example leaves a field v1 had no counterpart for unset")
	}
	for _, tc := range []struct {
		name      string
		got, want proto.Message
	}{
		{"ecu.status", throughJSON(t, got), wantStatus},
		{"ecu.session", throughJSON(t, session), wantSession},
	} {
		if !proto.Equal(tc.got, tc.want) {
			t.Errorf("%s: v1's body gives\n%v\nthe example is\n%v", tc.name, tc.got, tc.want)
		}
	}
}

// example is an example under examples/, decoded as a reader decodes a KV
// value.
func example(t *testing.T, file string, empty proto.Message) proto.Message {
	t.Helper()
	m := empty.ProtoReflect().New().Interface()
	if err := ce.Unmarshal(readExample(t, file), m); err != nil {
		t.Fatal(err)
	}
	return m
}

// throughJSON encodes a KV value with ce.Marshal and decodes it with
// ce.Unmarshal, as ecu writes it with Put and a reader gets it.
func throughJSON[M proto.Message](t *testing.T, msg M) M {
	t.Helper()
	data, err := ce.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	got := msg.ProtoReflect().New().Interface().(M)
	if err := ce.Unmarshal(data, got); err != nil {
		t.Fatal(err)
	}
	return got
}
