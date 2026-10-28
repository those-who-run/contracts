package contracts_test

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/those-who-run/contracts/ce"
	eventsv1 "github.com/those-who-run/contracts/gen/go/twr/events/v1"
)

// The pit, car and telemetry events paddock#70 adds.

// The stop v1.0.0's pit detector records for
// real-nuc-gilles-villeneuve-pit-stop-2026-10-17.jsonl.gz (AC2): in on lap 6
// at 8.7 litres, a full tank and new tyres, out on lap 7. testdata/v1 keeps
// what v1's PitDetector returned at the entry and at the exit, each with its
// frame's `t` and packet_id, as v1's tests/test_pitstop.py replays it.
//
// A frame's seq is its packet_id less the recording's first one, plus 1: the
// count ecu would have given every packet, where the recording keeps only
// some. The examples' clock puts the recording's first frame at
// 2026-10-17T18:00:00Z, to the microsecond.
var gillesFirstFrame = time.Date(2026, 10, 17, 18, 0, 0, 0, time.UTC)

// v1Frame is a frame's seq, its `t` and its UTC time on the examples' clock,
// from the frame and the recording's first frame's packet_id and t.
func v1Frame(r v1Record, firstPacket, firstT float64) (uint32, float64, *timestamppb.Timestamp) {
	packet, t := r.num("packet_id"), r.num("t")
	at := gillesFirstFrame.Add(time.Duration(math.Round((t-firstT)*1e6)) * time.Microsecond)
	return uint32(packet - firstPacket + 1), t, timestamppb.New(at)
}

// throughCE encodes msg with its ce-type and decodes it again, as a producer
// and a consumer do.
func throughCE[M proto.Message](t *testing.T, msg M, id string) M {
	t.Helper()
	h, payload, err := ce.Encode(msg, ce.Event{ID: id, Source: "/twr/ecu/gt7", Time: gillesFirstFrame})
	if err != nil {
		t.Fatal(err)
	}
	got := msg.ProtoReflect().New().Interface().(M)
	if _, err := ce.Decode(h, payload, got); err != nil {
		t.Fatal(err)
	}
	return got
}

// pit.entered and pit.exited are what v1 records for the Gilles Villeneuve
// stop, through ce, and pit.exited's stop is v1's PitStop.record() at the exit
// field for field (AC2).
func TestPitStopFromV1(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "v1", "gilles_villeneuve_pit_stop.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	r := v1Record{t, "", m}
	first, _ := r.object("first_frame")
	entry, _ := r.object("entry")
	exit, _ := r.object("exit")

	firstPacket, firstT := first.num("packet_id"), first.num("t")
	entryFrame, _ := entry.object("frame")
	entrySeq, entryT, enteredAt := v1Frame(entryFrame, firstPacket, firstT)
	atEntry, _ := entry.object("stop")
	if stint := entry.i32("stint"); stint != 1 {
		t.Errorf("v1's stint at the entry is %d, want 1", stint)
	}
	stopAtEntry := pitStopFromV1(atEntry)
	wantEntered := &eventsv1.PitEntered{
		SessionId:      "01JBBX7Q2M5R8T1V4W6Y9K3NZC",
		Entry:          1,
		Stop:           stopAtEntry.GetStop(),
		InLap:          stopAtEntry.GetInLap(),
		EnteredT:       entryT,
		EnteredAt:      enteredAt,
		FuelBefore:     stopAtEntry.GetFuelBefore(),
		PositionBefore: stopAtEntry.PositionBefore,
		Evidence:       stopAtEntry.GetEvidence(),
		At:             entrySeq,
	}
	var entered eventsv1.PitEntered
	if err := ce.Unmarshal(readExample(t, "events/v1/pit_entered.json"), &entered); err != nil {
		t.Fatal(err)
	}
	got := throughCE(t, &entered, "live:"+entered.GetSessionId()+":pit.entered:1")
	if !proto.Equal(got, wantEntered) {
		t.Errorf("pit.entered is not v1's entry:\ngot  %v\nwant %v", got, wantEntered)
	}

	exitFrame, _ := exit.object("frame")
	exitSeq, _, exitedAt := v1Frame(exitFrame, firstPacket, firstT)
	stopAtExit, _ := exit.object("stop")
	wantExited := &eventsv1.PitExited{
		SessionId: "01JBBX7Q2M5R8T1V4W6Y9K3NZC",
		Stop:      pitStopFromV1(stopAtExit),
		Stint:     exit.i32("stint"),
		Entry:     1,
		EnteredAt: enteredAt,
		ExitedAt:  exitedAt,
		At:        exitSeq,
	}
	var exited eventsv1.PitExited
	if err := ce.Unmarshal(readExample(t, "events/v1/pit_exited.json"), &exited); err != nil {
		t.Fatal(err)
	}
	gotExit := throughCE(t, &exited, "live:"+exited.GetSessionId()+":pit.exited:1")
	if !proto.Equal(gotExit, wantExited) {
		t.Errorf("pit.exited is not v1's stop:\ngot  %v\nwant %v", gotExit, wantExited)
	}
	// The issue's figures: a full tank and tyres, and the stint that starts.
	s := gotExit.GetStop()
	if s.GetFuelAfter() != 100 || !s.GetTyresChanged() || gotExit.GetStint() != 2 || s.GetOutLap() != 7 {
		t.Errorf("the Gilles Villeneuve stop is %v", gotExit)
	}
	// pit_lane_ms is v1's pit_lane_secs, which is the exit less the entry on
	// the frames' clock, to 10 ms.
	lane := exitedAt.AsTime().Sub(enteredAt.AsTime())
	if d := lane - time.Duration(s.GetPitLaneMs())*time.Millisecond; d < -5*time.Millisecond || d > 5*time.Millisecond {
		t.Errorf("pit_lane_ms %d is not exited_at less entered_at, %v", s.GetPitLaneMs(), lane)
	}
	r.done()
}

// A teleport gives its stop number back, so the next entry has the same stop
// as the one withdrawn; the entry count keeps their ids apart, where <stop>
// would repeat inside EVENTS' 2 min duplicate window and drop the real
// entry.
func TestAWithdrawnEntryAndTheNextHaveTheirOwnIDs(t *testing.T) {
	withdrawn := &eventsv1.PitEntered{SessionId: "01JBBX7Q2M5R8T1V4W6Y9K3NZC", Entry: 1, Stop: 1, InLap: 3, At: 3}
	teleport := &eventsv1.PitTeleport{SessionId: withdrawn.GetSessionId(), Entry: 1, EntryAt: proto.Uint32(3), Lap: 3, At: 241}
	next := &eventsv1.PitEntered{SessionId: withdrawn.GetSessionId(), Entry: 2, Stop: 1, InLap: 4, At: 6020}
	id := func(e *eventsv1.PitEntered) string {
		return strings.Join([]string{"live", e.GetSessionId(), "pit.entered", strconv.Itoa(int(e.GetEntry()))}, ":")
	}
	if withdrawn.GetStop() != next.GetStop() || id(withdrawn) == id(next) {
		t.Errorf("ids %s and %s", id(withdrawn), id(next))
	}
	if teleport.GetEntry() != withdrawn.GetEntry() {
		t.Errorf("the teleport withdraws entry %d, want %d", teleport.GetEntry(), withdrawn.GetEntry())
	}
	for _, msg := range []proto.Message{withdrawn, teleport, next} {
		roundTrip(t, msg)
	}
	// entry_at 0 would be a real seq's absence, so it is optional.
	if fd := teleport.ProtoReflect().Descriptor().Fields().ByName("entry_at"); !fd.HasPresence() {
		t.Error("PitTeleport.entry_at is not optional")
	}
}

// A telemetry event published with no session open has no session id, and
// says which ecu run its seqs count in; restored names the loss it ends.
func TestTelemetryEventsWithNoSession(t *testing.T) {
	var restored eventsv1.TelemetryRestored
	if err := ce.Unmarshal(readExample(t, "events/v1/telemetry_restored.json"), &restored); err != nil {
		t.Fatal(err)
	}
	if restored.GetSessionId() != "" || restored.GetRunId() == "" {
		t.Errorf("the evt._ example has session %q and run %q", restored.GetSessionId(), restored.GetRunId())
	}
	var lost eventsv1.TelemetryLost
	if err := ce.Unmarshal(readExample(t, "events/v1/telemetry_lost.json"), &lost); err != nil {
		t.Fatal(err)
	}
	if restored.GetLostAt() != lost.GetAt() || restored.GetRunId() != lost.GetRunId() || restored.GetAt() <= lost.GetAt() {
		t.Errorf("restored %v does not end lost %v", &restored, &lost)
	}
	got := throughCE(t, &restored, "live:"+restored.GetRunId()+":telemetry.restored:"+strconv.Itoa(int(restored.GetAt())))
	if !proto.Equal(got, &restored) {
		t.Errorf("through ce: got %v, want %v", got, &restored)
	}
}
