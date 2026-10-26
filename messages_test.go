package contracts_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	eventsv1 "github.com/those-who-run/contracts/gen/go/twr/events/v1"
	telemetryv1 "github.com/those-who-run/contracts/gen/go/twr/telemetry/v1"
)

// The v0.1 messages, each with its example under examples/ and the values the
// example holds. ADR 0002's JSON rules: readers ignore unknown fields, and tests
// compare decoded values, never JSON bytes.
var examples = []struct {
	file string
	want proto.Message
}{
	{"telemetry/v1/frame.json", &telemetryv1.Frame{
		Seq: 18234, T: 3021.517, PacketId: 512877, Lap: 3, Phase: telemetryv1.Phase_PHASE_IN_RACE,
		PosX: -412.5, PosY: 12.25, PosZ: 1088.75, VelX: 31.5, VelY: -0.25, VelZ: 40.75,
		SpeedKph: 185.5, EngineRpm: 6850, Gear: 4, SuggestedGear: 15, Throttle: 255, Brake: 0,
		FuelLevel:  41.75,
		TireFlTemp: 82, TireFrTemp: 84.5, TireRlTemp: 79, TireRrTemp: 80.25,
		TireFlRadius: 0.33, TireFrRadius: 0.33, TireRlRadius: 0.345, TireRrRadius: 0.345,
		SusFl: 0.071, SusFr: 0.069, SusRl: 0.082, SusRr: 0.08,
		WheelFlRps: -156.25, WheelFrRps: -156.5, WheelRlRps: -150, WheelRrRps: -150.25,
		WaterTemp: 85, OilTemp: 104, OilPressure: 5.5, BoostPressure: 1.75, BodyHeight: 0.105,
		Position: proto.Int32(2), TimeOfDayMs: 52200000, CarCode: 3431, TrackId: proto.Int32(351),
		RotX: 0.0125, RotY: -0.625, RotZ: 0.005, Orientation: 0.75,
		AngVelX: 0.01, AngVelY: 0.35, AngVelZ: -0.02,
		Flags: 33,
	}},
	{"events/v1/session_started.json", &eventsv1.SessionStarted{
		SessionId:     "01JB8Z3K5Q9T2M7XW4R6N1PDHE",
		Source:        "live",
		Game:          eventsv1.Game_GAME_GT7,
		SessionType:   eventsv1.SessionType_SESSION_TYPE_RACE,
		CarCode:       3431,
		TotalLaps:     proto.Int32(10),
		JoinedPartWay: false,
		FirstSeq:      1200,
		StartedAt:     timestamppb.New(time.Date(2026, 10, 26, 19, 30, 5, 250_000_000, time.UTC)),
	}},
	{"events/v1/lap_completed.json", &eventsv1.LapCompleted{
		SessionId:   "01JB8Z3K5Q9T2M7XW4R6N1PDHE",
		Lap:         3,
		LapTimeMs:   98765,
		Complete:    true,
		FuelAtStart: 45.5,
		FuelAtEnd:   41.75,
		FuelUsed:    3.75,
		Position:    proto.Int32(2),
		PitRole:     eventsv1.PitRole_PIT_ROLE_NONE,
		FirstSeq:    13311,
		LastSeq:     19236,
	}},
	{"events/v1/session_ended.json", &eventsv1.SessionEnded{
		SessionId: "01JB8Z3K5Q9T2M7XW4R6N1PDHE",
		Reason:    eventsv1.EndReason_END_REASON_FINISHED,
		Summary: &eventsv1.SessionSummary{
			LapsCompleted:     10,
			BestLap:           proto.Int32(4),
			BestLapTimeMs:     proto.Int32(97412),
			FinishingPosition: proto.Int32(1),
		},
	}},
}

var readers = protojson.UnmarshalOptions{DiscardUnknown: true}

func readExample(t *testing.T, file string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("examples", filepath.FromSlash(file)))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// decode reads data into a new message of want's type, as every reader does.
func decode(t *testing.T, opts protojson.UnmarshalOptions, data []byte, want proto.Message) (proto.Message, error) {
	t.Helper()
	got := want.ProtoReflect().New().Interface()
	return got, opts.Unmarshal(data, got)
}

// Each example decodes to the values the test expects.
func TestExamplesDecode(t *testing.T) {
	for _, ex := range examples {
		t.Run(ex.file, func(t *testing.T) {
			got, err := decode(t, readers, readExample(t, ex.file), ex.want)
			if err != nil {
				t.Fatalf("decoding the example: %v", err)
			}
			if !proto.Equal(got, ex.want) {
				t.Errorf("got %v\nwant %v", got, ex.want)
			}
		})
	}
}

// A field the reader doesn't know, such as one a later release adds, is ignored,
// and the rest of the message still decodes. Without DiscardUnknown it would fail.
func TestUnknownFieldsAreIgnored(t *testing.T) {
	for _, ex := range examples {
		t.Run(ex.file, func(t *testing.T) {
			var fields map[string]any
			if err := json.Unmarshal(readExample(t, ex.file), &fields); err != nil {
				t.Fatal(err)
			}
			fields["addedInALaterRelease"] = map[string]any{"value": 1, "note": "unknown"}
			data, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}

			if _, err := decode(t, protojson.UnmarshalOptions{}, data, ex.want); err == nil {
				t.Fatal("a strict reader accepted the unknown field, so this case checks nothing")
			}
			got, err := decode(t, readers, data, ex.want)
			if err != nil {
				t.Fatalf("decoding with an unknown field: %v", err)
			}
			if !proto.Equal(got, ex.want) {
				t.Errorf("got %v\nwant %v", got, ex.want)
			}
		})
	}
}

// Each message survives protobuf's JSON form: what one component writes, another
// reads back as the same values.
func TestMessagesRoundTripThroughJSON(t *testing.T) {
	for _, ex := range examples {
		t.Run(ex.file, func(t *testing.T) {
			data, err := protojson.Marshal(ex.want)
			if err != nil {
				t.Fatalf("encoding: %v", err)
			}
			got, err := decode(t, readers, data, ex.want)
			if err != nil {
				t.Fatalf("decoding %s: %v", data, err)
			}
			if !proto.Equal(got, ex.want) {
				t.Errorf("got %v\nwant %v", got, ex.want)
			}
		})
	}
}

// Zero values that mean something survive the JSON form, which leaves zeros out:
// an optional field set to 0 is still set, and one left unset is still absent.
func TestOptionalZeroIsNotAbsent(t *testing.T) {
	for _, tc := range []struct {
		name string
		msg  *eventsv1.SessionSummary
	}{
		{"set to zero", &eventsv1.SessionSummary{BestLapTimeMs: proto.Int32(0)}},
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
			if got.BestLapTimeMs == nil != (tc.msg.BestLapTimeMs == nil) {
				t.Errorf("%s: best_lap_time_ms presence changed through %s", tc.name, data)
			}
		})
	}
}
