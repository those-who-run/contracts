package contracts_test

import (
	"fmt"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
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
	if n != 19 {
		t.Errorf("%d events, want 19: v0.1's 3, paddock#69's 10 and paddock#70's 6", n)
	}
}

func descriptorOf(t *testing.T, name protoreflect.FullName) protoreflect.MessageDescriptor {
	t.Helper()
	for _, m := range []protoreflect.ProtoMessage{
		&eventsv1.SessionStarted{}, &eventsv1.LapCompleted{}, &eventsv1.SessionEnded{},
		&eventsv1.SessionContinued{}, &eventsv1.SessionDiscarded{}, &eventsv1.SessionKindChanged{},
		&eventsv1.SessionTrackChanged{}, &eventsv1.SessionPaused{}, &eventsv1.SessionResumed{},
		&eventsv1.LapBestChanged{}, &eventsv1.RaceFinalLapStarted{}, &eventsv1.RacePositionChanged{},
		&eventsv1.RaceFinished{},
		&eventsv1.PitEntered{}, &eventsv1.PitExited{}, &eventsv1.PitTeleport{},
		&eventsv1.CarAlarmRaised{}, &eventsv1.TelemetryLost{}, &eventsv1.TelemetryRestored{},
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
		{&eventsv1.SessionPaused{}, "pause", 7, protoreflect.Int32Kind, false},
		{&eventsv1.SessionResumed{}, "pause", 8, protoreflect.Int32Kind, false},
		{&eventsv1.LapBestChanged{}, "ended_lap", 7, protoreflect.Int32Kind, false},
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

// lap.best is the best lap changing, not only improving: GT7's own best can
// replace a quicker lap ecu had counted, so gap_ms can be negative, and it
// survives both forms with its sign.
func TestLapBestChangedGapCanBeNegative(t *testing.T) {
	msg := &eventsv1.LapBestChanged{
		SessionId: "01JB9R1C8E4Q6W2Y9T3K7M5PXB", Lap: 2, LapTimeMs: 67585,
		PreviousBestMs: proto.Int32(62036), GapMs: proto.Int32(-5549), At: 4410, EndedLap: 2,
	}
	roundTrip(t, msg)
	comment := leadingComment(t, "twr/events/v1/lap.proto", "gap_ms")
	if !strings.Contains(comment, "Negative when GT7's own best replaced") {
		t.Errorf("gap_ms's comment does not say it can be negative:\n%s", comment)
	}
}

// The comments say what a consumer must know to read these events right.
func TestCommentsSayHowToReadThem(t *testing.T) {
	for _, tc := range []struct {
		file, name string
		says       []string
	}{
		// counts_for_best is GT7's rule, which leaves lap 1 of a timed race from a
		// grid out as well as a lap race's, and ecu knows it when the lap ends.
		{"twr/events/v1/lap.proto", "counts_for_best", []string{
			"GT7's rule", "lap 1 of a session with a lap total (a lap race or solo laps) never counts",
			"lap 1 of a timed session counts only when GT7 set its own best", "ecu knows it when the lap ends",
		}},
		// seqs across a restart gap.
		{"twr/events/v1/lap.proto", "first_seq", []string{"continuing run's first frame of the lap", "only within the run"}},
		{"twr/events/v1/session.proto", "laps", []string{"only within the", "two runs"}},
		// QUIT's meaning narrowed in v0.2.0, and how to read rows from before it.
		{"twr/events/v1/session.proto", "END_REASON_QUIT", []string{"Before the contracts release that added END_REASON_MENU and END_REASON_TIMED_FINISH", "END_REASON_MENU", "END_REASON_TIMED_FINISH", "maps them"}},
		// best_lap's known difference from v1.
		{"twr/events/v1/session.proto", "best_lap", []string{"Absent until ecu decides it", "known difference from v1", "pit laps included", "paddock#83"}},
		// The pause number is the id's <pause>, carried across a restart, and a
		// pause open across one keeps it.
		{"twr/events/v1/session.proto", "pause = 7", []string{"<pause>", "across an ecu restart", "a pause still open when ecu stopped keeps its number"}},
		// lap.best's <n> is the lap that published it, which does not repeat.
		{"twr/events/v1/lap.proto", "ended_lap", []string{"the <n> in the event's id", "lap alone could repeat"}},
		// A restart through a loading screen reads as QUIT; RESTARTED is in place.
		{"twr/events/v1/session.proto", "END_REASON_RESTARTED", []string{"restarted in place", "A restart through a loading screen at lap 0 is END_REASON_QUIT"}},
		{"twr/events/v1/session.proto", "END_REASON_QUIT", []string{"A race restarted through a loading screen", "reads as QUIT too"}},
		// A continued session keeps the start the earlier run saw.
		{"twr/events/v1/session.proto", "START_TYPE_UNKNOWN", []string{"keeps the start the earlier run saw"}},
		// best_lap when no lap has GT7's best time.
		{"twr/events/v1/session.proto", "best_lap", []string{"no lap with its time", "whatever its counts_for_best"}},
		// The position event differs from v1's callout on purpose.
		{"twr/events/v1/race.proto", "RacePositionChanged", []string{"differs from v1's position callout on purpose", "start_grace_secs", "min_interval_secs", "paddock#84 AC4", "a superset of calls"}},
		// paddock#70. An entry's own count, since a teleport gives its stop
		// number back, carried across a restart.
		{"twr/events/v1/pit.proto", "entry = 2", []string{"the <entry> in the event's id", "the ones a PitTeleport withdraws included", "across an ecu restart", "stop alone could repeat"}},
		{"twr/events/v1/pit.proto", "PitStop stop", []string{"the <stop> in the event's id", "across an ecu restart"}},
		{"twr/events/v1/pit.proto", "entry_at", []string{"Absent when", "earlier ecu run"}},
		{"twr/events/v1/pit.proto", "PitTeleport", []string{"without ever coming to rest", "no fuel added and no tyres changed", "undoes it"}},
		// The alarm's id part, and v1's rule for when it fires.
		{"twr/events/v1/car.proto", "check", []string{"the <check> in the event's id", "tyre_hot, tyres_cold, water_hot or oil_hot"}},
		{"twr/events/v1/car.proto", "CarAlarmRaised", []string{"sustain_secs", "hysteresis_c", "cooldown_secs"}},
		// The _ session token, and the run that replaces the session there.
		{"twr/events/v1/telemetry.proto", "TelemetryLost", []string{"evt._.telemetry.lost", "while none", "is"}},
		{"twr/events/v1/telemetry.proto", "run_id = 2", []string{"the <run> in the", "event's id", "evt._"}},
		{"twr/events/v1/telemetry.proto", "session_id = 1", []string{"Empty when no session is", "evt._"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			comment := leadingComment(t, tc.file, tc.name)
			for _, want := range tc.says {
				if !strings.Contains(comment, want) {
					t.Errorf("%s's comment does not say %q:\n%s", tc.name, want, comment)
				}
			}
		})
	}
}
