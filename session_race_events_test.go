package contracts_test

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/those-who-run/contracts/ce"
	eventsv1 "github.com/those-who-run/contracts/gen/go/twr/events/v1"
)

// The session and race events paddock#69 adds, and what it adds to v0.1's.

// at is each event's `at` field: the `seq` of the frame it happened on (ADR
// 0001, point 3). The three events released before paddock#69 already carried
// it under their own names; every later one calls it `at`.
var at = map[protoreflect.FullName]protoreflect.Name{
	"twr.events.v1.SessionStarted": "first_seq",
	"twr.events.v1.LapCompleted":   "last_seq",
	"twr.events.v1.SessionEnded":   "last_seq",
}

// Every event carries its session (`session_id`, the sid) and its `at`, a
// uint32 (R3; no 64-bit integers, which make proto checks too).
func TestEveryEventCarriesItsSessionAndAt(t *testing.T) {
	n := 0
	for name := range ce.Types() {
		if !strings.HasPrefix(string(name), "twr.events.") {
			continue
		}
		n++
		desc := descriptorOf(t, name)
		if fd := desc.Fields().ByName("session_id"); fd == nil || fd.Kind() != protoreflect.StringKind || fd.Number() != 1 {
			t.Errorf("%s has no string session_id as field 1", name)
		}
		field := protoreflect.Name("at")
		if f, ok := at[name]; ok {
			field = f
		}
		if fd := desc.Fields().ByName(field); fd == nil || fd.Kind() != protoreflect.Uint32Kind || fd.HasPresence() {
			t.Errorf("%s has no plain uint32 %s", name, field)
		}
	}
	if n != 13 {
		t.Errorf("%d events, want 13: v0.1's 3 and paddock#69's 10", n)
	}
}

func descriptorOf(t *testing.T, name protoreflect.FullName) protoreflect.MessageDescriptor {
	t.Helper()
	for _, m := range []protoreflect.ProtoMessage{
		&eventsv1.SessionStarted{}, &eventsv1.LapCompleted{}, &eventsv1.SessionEnded{},
		&eventsv1.SessionContinued{}, &eventsv1.SessionDiscarded{}, &eventsv1.SessionKindChanged{},
		&eventsv1.SessionTrackChanged{}, &eventsv1.SessionPaused{}, &eventsv1.SessionResumed{},
		&eventsv1.LapBestImproved{}, &eventsv1.RaceFinalLapStarted{}, &eventsv1.RacePositionChanged{},
		&eventsv1.RaceFinished{},
	} {
		if d := m.ProtoReflect().Descriptor(); d.FullName() == name {
			return d
		}
	}
	t.Fatalf("no message %s", name)
	return nil
}

// The fields paddock#69 adds to released messages keep the numbers, types and
// presence they are released with: buf breaking guards them only once a
// release holds them (R1, R2).
func TestAddedFieldNumbers(t *testing.T) {
	for _, tc := range []struct {
		msg      protoreflect.ProtoMessage
		field    protoreflect.Name
		number   protoreflect.FieldNumber
		kind     protoreflect.Kind
		optional bool
	}{
		{&eventsv1.SessionStarted{}, "car_name", 11, protoreflect.StringKind, false},
		{&eventsv1.SessionStarted{}, "frame_track_id", 12, protoreflect.Int32Kind, true},
		{&eventsv1.SessionStarted{}, "joined_on_lap", 13, protoreflect.Int32Kind, true},
		{&eventsv1.SessionStarted{}, "fuel_capacity", 14, protoreflect.DoubleKind, false},
		{&eventsv1.LapCompleted{}, "counts_for_best", 12, protoreflect.BoolKind, false},
		{&eventsv1.SessionSummary{}, "started_at", 8, protoreflect.MessageKind, true},
		{&eventsv1.SessionSummary{}, "driving_ms", 10, protoreflect.Int32Kind, false},
		{&eventsv1.SessionSummary{}, "track", 13, protoreflect.MessageKind, true},
		{&eventsv1.SessionSummary{}, "kind", 14, protoreflect.MessageKind, true},
		{&eventsv1.SessionSummary{}, "total_laps", 17, protoreflect.Int32Kind, true},
		{&eventsv1.SessionSummary{}, "game_best_lap_ms", 19, protoreflect.Int32Kind, true},
		{&eventsv1.SessionSummary{}, "total_cars", 23, protoreflect.Int32Kind, true},
		{&eventsv1.SessionSummary{}, "fuel_stints", 26, protoreflect.DoubleKind, false},
		{&eventsv1.SessionSummary{}, "pit_stops", 31, protoreflect.MessageKind, false},
		{&eventsv1.SessionSummary{}, "gaps", 32, protoreflect.MessageKind, false},
		{&eventsv1.SessionSummary{}, "laps", 33, protoreflect.MessageKind, false},
	} {
		desc := tc.msg.ProtoReflect().Descriptor()
		fd := desc.Fields().ByName(tc.field)
		switch {
		case fd == nil:
			t.Errorf("%s has no field %s", desc.FullName(), tc.field)
		case fd.Number() != tc.number || fd.Kind() != tc.kind || fd.HasPresence() != tc.optional:
			t.Errorf("%s.%s is %d %v optional %v, want %d %v optional %v", desc.FullName(), tc.field,
				fd.Number(), fd.Kind(), fd.HasPresence(), tc.number, tc.kind, tc.optional)
		}
	}
	// The summary's laps are lap.completed's own message, not a copy of it.
	if fd := (&eventsv1.SessionSummary{}).ProtoReflect().Descriptor().Fields().ByName("laps"); fd.Message().FullName() != "twr.events.v1.LapCompleted" {
		t.Errorf("SessionSummary.laps is %s, want twr.events.v1.LapCompleted", fd.Message().FullName())
	}
	for _, tc := range []struct {
		value eventsv1.EndReason
		want  protoreflect.EnumNumber
	}{
		{eventsv1.EndReason_END_REASON_TIMED_FINISH, 9},
		{eventsv1.EndReason_END_REASON_MENU, 10},
		{eventsv1.EndReason_END_REASON_PAUSE_LIMIT, 11},
	} {
		if got := tc.value.Number(); got != tc.want {
			t.Errorf("%v is %d, want %d", tc.value, got, tc.want)
		}
	}
}

// Each new end reason survives both forms, and the JSON form names it.
func TestAddedEndReasonsRoundTrip(t *testing.T) {
	for _, r := range []eventsv1.EndReason{
		eventsv1.EndReason_END_REASON_TIMED_FINISH,
		eventsv1.EndReason_END_REASON_MENU,
		eventsv1.EndReason_END_REASON_PAUSE_LIMIT,
	} {
		t.Run(r.String(), func(t *testing.T) {
			msg := &eventsv1.SessionEnded{Reason: r}
			roundTrip(t, msg)
			assertJSONNames(t, msg, r.String())
		})
	}
}

// Each field paddock#69 adds to the summary, and each of the first three,
// names the v1 field it carries, so the parity harness and logbook can map
// them (R6).
func TestSummaryCommentsNameTheV1Field(t *testing.T) {
	fields := (&eventsv1.SessionSummary{}).ProtoReflect().Descriptor().Fields()
	for i := range fields.Len() {
		fd := fields.Get(i)
		if fd.Number() > 3 && fd.Number() < 8 {
			continue // paddock#276's, which say where they differ from v1
		}
		comment := leadingComment(t, "twr/events/v1/session.proto", fmt.Sprintf("%s = %d", fd.Name(), fd.Number()))
		if !strings.Contains(comment, "v1's `") {
			t.Errorf("SessionSummary.%s's comment names no v1 field:\n%s", fd.Name(), comment)
		}
	}
}
