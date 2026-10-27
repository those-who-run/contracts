package contracts_test

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	eventsv1 "github.com/those-who-run/contracts/gen/go/twr/events/v1"
)

// The session fields and values paddock#276 adds: what ecu works out at a
// session's start and end, the end reasons for a restart, the next start and
// the end of the input, and the session types telemetry can tell apart.

// Each new field and value keeps the number it was released with: a consumer
// built against one release reads the next by these numbers.
func TestNewFieldAndValueNumbers(t *testing.T) {
	for _, tc := range []struct {
		msg   proto.Message
		field protoreflect.Name
		want  protoreflect.FieldNumber
	}{
		{&eventsv1.SessionStarted{}, "fuel_at_start", 10},
		{&eventsv1.SessionEnded{}, "last_seq", 4},
		{&eventsv1.SessionSummary{}, "fuel_used", 5},
		{&eventsv1.SessionSummary{}, "final_position", 6},
	} {
		desc := tc.msg.ProtoReflect().Descriptor()
		fd := desc.Fields().ByName(tc.field)
		if fd == nil {
			t.Errorf("%s has no field %s", desc.FullName(), tc.field)
			continue
		}
		if fd.Number() != tc.want {
			t.Errorf("%s.%s is field %d, want %d", desc.FullName(), tc.field, fd.Number(), tc.want)
		}
	}
	for _, tc := range []struct {
		value protoreflect.Enum
		want  protoreflect.EnumNumber
	}{
		{eventsv1.SessionType_SESSION_TYPE_LAP_RACE, 5},
		{eventsv1.SessionType_SESSION_TYPE_SOLO_LAPS, 6},
		{eventsv1.SessionType_SESSION_TYPE_TIMED_FIELD, 7},
		{eventsv1.SessionType_SESSION_TYPE_SOLO_TIMED, 8},
		{eventsv1.EndReason_END_REASON_RESTARTED, 6},
		{eventsv1.EndReason_END_REASON_NEXT_SESSION_STARTED, 7},
		{eventsv1.EndReason_END_REASON_INPUT_ENDED, 8},
	} {
		if got := tc.value.Number(); got != tc.want {
			t.Errorf("%v is %d, want %d", tc.value, got, tc.want)
		}
	}
}

// roundTrip encodes msg in protobuf's JSON form and its binary form, and decodes
// each back as every reader does; both must give the same values.
func roundTrip(t *testing.T, msg proto.Message) {
	t.Helper()
	data, err := protojson.Marshal(msg)
	if err != nil {
		t.Fatalf("encoding JSON: %v", err)
	}
	got := msg.ProtoReflect().New().Interface()
	if err := readers.Unmarshal(data, got); err != nil {
		t.Fatalf("decoding JSON %s: %v", data, err)
	}
	if !proto.Equal(got, msg) {
		t.Errorf("through JSON %s: got %v, want %v", data, got, msg)
	}
	bin, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("encoding binary: %v", err)
	}
	got = msg.ProtoReflect().New().Interface()
	if err := proto.Unmarshal(bin, got); err != nil {
		t.Fatalf("decoding binary: %v", err)
	}
	if !proto.Equal(got, msg) {
		t.Errorf("through binary: got %v, want %v", got, msg)
	}
}

// Each new field, set on its own, survives both forms with its value.
func TestNewFieldsRoundTrip(t *testing.T) {
	for _, tc := range []struct {
		name string
		msg  proto.Message
	}{
		{"SessionStarted.fuel_at_start", &eventsv1.SessionStarted{FuelAtStart: 41.75}},
		{"SessionEnded.last_seq", &eventsv1.SessionEnded{LastSeq: 61840}},
		{"SessionSummary.fuel_used", &eventsv1.SessionSummary{FuelUsed: 37.625}},
		{"SessionSummary.fuel_used negative", &eventsv1.SessionSummary{FuelUsed: -12.5}},
		{"SessionSummary.final_position", &eventsv1.SessionSummary{FinalPosition: proto.Int32(7)}},
		{"SessionSummary.final_position absent", &eventsv1.SessionSummary{}},
		{"SessionEnded with every new field", &eventsv1.SessionEnded{
			SessionId: "01JB8Z3K5Q9T2M7XW4R6N1PDHE",
			Reason:    eventsv1.EndReason_END_REASON_RESTARTED,
			Summary:   &eventsv1.SessionSummary{LapsCompleted: 2, FuelUsed: 6.25, FinalPosition: proto.Int32(3)},
			LastSeq:   4211,
		}},
	} {
		t.Run(tc.name, func(t *testing.T) { roundTrip(t, tc.msg) })
	}
}

// Each new value survives both forms, and the JSON form carries it by its name.
func TestNewValuesRoundTrip(t *testing.T) {
	for _, st := range []eventsv1.SessionType{
		eventsv1.SessionType_SESSION_TYPE_LAP_RACE,
		eventsv1.SessionType_SESSION_TYPE_SOLO_LAPS,
		eventsv1.SessionType_SESSION_TYPE_TIMED_FIELD,
		eventsv1.SessionType_SESSION_TYPE_SOLO_TIMED,
	} {
		t.Run(st.String(), func(t *testing.T) {
			msg := &eventsv1.SessionStarted{SessionType: st}
			roundTrip(t, msg)
			assertJSONNames(t, msg, st.String())
		})
	}
	for _, r := range []eventsv1.EndReason{
		eventsv1.EndReason_END_REASON_RESTARTED,
		eventsv1.EndReason_END_REASON_NEXT_SESSION_STARTED,
		eventsv1.EndReason_END_REASON_INPUT_ENDED,
	} {
		t.Run(r.String(), func(t *testing.T) {
			msg := &eventsv1.SessionEnded{Reason: r}
			roundTrip(t, msg)
			assertJSONNames(t, msg, r.String())
		})
	}
}

func assertJSONNames(t *testing.T, msg proto.Message, name string) {
	t.Helper()
	data, err := protojson.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"`+name+`"`) {
		t.Errorf("the JSON form %s does not name %s", data, name)
	}
}

// final_position is optional because a position is never 0 but its absence
// means something (outside a field): set and absent stay apart through JSON,
// which leaves zeros out. fuel_used is plain: ecu always knows it, and 0 is a
// session that burned none.
func TestFinalPositionPresence(t *testing.T) {
	for _, tc := range []struct {
		name string
		msg  *eventsv1.SessionSummary
	}{
		{"set", &eventsv1.SessionSummary{FinalPosition: proto.Int32(1)}},
		{"absent", &eventsv1.SessionSummary{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := protojson.Marshal(tc.msg)
			if err != nil {
				t.Fatal(err)
			}
			got := &eventsv1.SessionSummary{}
			if err := readers.Unmarshal(data, got); err != nil {
				t.Fatal(err)
			}
			if (got.FinalPosition == nil) != (tc.msg.FinalPosition == nil) {
				t.Errorf("final_position presence changed through %s", data)
			}
		})
	}
}

// A declaration's leading comment in a .proto file, its `//` lines joined with
// spaces: what buf lint's COMMENTS rule checks is there, and what the generated
// code carries.
func leadingComment(t *testing.T, file, name string) string {
	t.Helper()
	f, err := os.Open(filepath.Join("proto", filepath.FromSlash(file)))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	decl := regexp.MustCompile(`\b` + regexp.QuoteMeta(name) + `\s*=\s*\d+;`)
	var comment []string
	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := strings.TrimSpace(scan.Text())
		switch {
		case strings.HasPrefix(line, "//"):
			comment = append(comment, strings.TrimSpace(strings.TrimPrefix(line, "//")))
		case decl.MatchString(line):
			return strings.Join(comment, " ")
		default:
			comment = nil
		}
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	t.Fatalf("%s declares no %s", file, name)
	return ""
}

// The comments state what ecu sends (paddock#276, R5 to R8), so a consumer reads
// the contract, not ecu's code.
func TestCommentsSayWhatEcuSends(t *testing.T) {
	for _, tc := range []struct {
		rule, file, name string
		says             []string
	}{
		// R5: the old session types are kept, and say ecu does not send them; the
		// new ones are v1's kinds, one to one.
		{"R5", "twr/events/v1/session.proto", "SESSION_TYPE_RACE", []string{"ecu does not send it", "SESSION_TYPE_LAP_RACE", "SESSION_TYPE_TIMED_FIELD"}},
		{"R5", "twr/events/v1/session.proto", "SESSION_TYPE_TIME_TRIAL", []string{"ecu does not send it", "SESSION_TYPE_SOLO_TIMED"}},
		{"R5", "twr/events/v1/session.proto", "SESSION_TYPE_PRACTICE", []string{"ecu does not send it", "SESSION_TYPE_SOLO_TIMED"}},
		{"R5", "twr/events/v1/session.proto", "SESSION_TYPE_UNKNOWN", []string{"v1's `unknown`", "ecu sends it"}},
		{"R5", "twr/events/v1/session.proto", "SESSION_TYPE_LAP_RACE", []string{"v1's `lap_race`", "ecu sends it"}},
		{"R5", "twr/events/v1/session.proto", "SESSION_TYPE_SOLO_LAPS", []string{"v1's `solo_laps`", "ecu sends it"}},
		{"R5", "twr/events/v1/session.proto", "SESSION_TYPE_TIMED_FIELD", []string{"timed race or qualifying", "v1's `timed_field`", "ecu sends it"}},
		{"R5", "twr/events/v1/session.proto", "SESSION_TYPE_SOLO_TIMED", []string{"time trial or practice", "v1's `solo_timed`", "ecu sends it"}},
		// R6: a finishing position only for a lap race with a field, at the flag.
		{"R6", "twr/events/v1/session.proto", "finishing_position", []string{"Set only for a lap race with a field", "SESSION_TYPE_LAP_RACE", "END_REASON_FINISHED", "qualifying", "solo laps"}},
		// R7: seq counts from 1.
		{"R7", "twr/telemetry/v1/frame.proto", "seq", []string{"from 1, so 0 always means not set"}},
		// R8: the joined lap is complete, as v1 stored it; the best lap is absent
		// until ecu decides it.
		{"R8", "twr/events/v1/lap.proto", "complete", []string{"The lap ecu joined a session in counts as complete", "as v1 stored it"}},
		{"R8", "twr/events/v1/session.proto", "best_lap", []string{"Absent until ecu decides it"}},
		{"R8", "twr/events/v1/session.proto", "best_lap_time_ms", []string{"Absent until ecu decides it"}},
		// R4: ecu stopping is not the end of a replayed recording.
		{"R4", "twr/events/v1/session.proto", "END_REASON_ENGINEER_STOPPED", []string{"END_REASON_INPUT_ENDED"}},
		{"R4", "twr/events/v1/session.proto", "END_REASON_INPUT_ENDED", []string{"replayed recording"}},
	} {
		t.Run(tc.rule+"/"+tc.name, func(t *testing.T) {
			comment := leadingComment(t, tc.file, tc.name)
			for _, want := range tc.says {
				if !strings.Contains(comment, want) {
					t.Errorf("%s's comment does not say %q:\n%s", tc.name, want, comment)
				}
			}
		})
	}
}
