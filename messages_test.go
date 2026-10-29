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

// The messages, each with its example under examples/ and the values the
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
		SessionType:   eventsv1.SessionType_SESSION_TYPE_LAP_RACE,
		CarCode:       3431,
		TotalLaps:     proto.Int32(10),
		JoinedPartWay: false,
		FirstSeq:      1200,
		StartedAt:     timestamppb.New(time.Date(2026, 10, 26, 19, 30, 5, 250_000_000, time.UTC)),
		FuelAtStart:   100,
		CarName:       "Honda Civic Si Group A version '87",
		FrameTrackId:  proto.Int32(442),
		FuelCapacity:  100,
	}},
	{"events/v1/lap_completed.json", &eventsv1.LapCompleted{
		SessionId:     "01JB8Z3K5Q9T2M7XW4R6N1PDHE",
		Lap:           3,
		LapTimeMs:     98765,
		Complete:      true,
		FuelAtStart:   45.5,
		FuelAtEnd:     41.75,
		FuelUsed:      3.75,
		Position:      proto.Int32(2),
		PitRole:       eventsv1.PitRole_PIT_ROLE_NONE,
		FirstSeq:      13311,
		LastSeq:       19236,
		CountsForBest: true,
	}},
	{"events/v1/session_ended.json", &eventsv1.SessionEnded{
		SessionId: "01JB8Z3K5Q9T2M7XW4R6N1PDHE",
		Reason:    eventsv1.EndReason_END_REASON_FINISHED,
		Summary: &eventsv1.SessionSummary{
			LapsCompleted:     10,
			BestLap:           proto.Int32(4),
			BestLapTimeMs:     proto.Int32(97412),
			FinishingPosition: proto.Int32(1),
			FuelUsed:          37.625,
			FinalPosition:     proto.Int32(1),
			SessionType:       eventsv1.SessionType_SESSION_TYPE_LAP_RACE,
		},
		LastSeq: 61840,
	}},
	// paddock#69's events.
	{"events/v1/session_continued.json", &eventsv1.SessionContinued{
		SessionId: "01JB9R1C8E4Q6W2Y9T3K7M5PXB",
		Gap: &eventsv1.SessionGap{
			StartedAt: timestamppb.New(time.Date(2026, 10, 18, 13, 45, 7, 776_810_000, time.UTC)),
			EndedAt:   timestamppb.New(time.Date(2026, 10, 18, 13, 45, 16, 437_865_000, time.UTC)),
			Lap:       7,
			Reason:    eventsv1.GapReason_GAP_REASON_ENGINEER_RESTART,
		},
		Track: &eventsv1.TrackVerdict{
			TrackId: proto.Int32(10), Name: "Suzuka Circuit", Confidence: eventsv1.TrackConfidence_TRACK_CONFIDENCE_CONFIRMED,
			Evidence: "carried over from before the engineer restart", Carried: true,
		},
		At:    1,
		RunId: "01JB9R1BX4N7Q2S5V8Z3D6F9JP",
	}},
	{"events/v1/session_discarded.json", &eventsv1.SessionDiscarded{
		SessionId: "01JB9S2D5F7R3X8Z1V4L6N9QYC",
		Reason:    eventsv1.DiscardReason_DISCARD_REASON_REPLAY,
		Evidence:  "fuel is draining (98.3 -> 97.6), which only a replay does",
		FirstSeq:  40210,
		At:        40291,
	}},
	{"events/v1/session_kind.json", &eventsv1.SessionKindChanged{
		SessionId: "01JB9QZ4V3X8M2K7T5R1N6PDWA",
		Kind: &eventsv1.SessionKind{
			SessionType: eventsv1.SessionType_SESSION_TYPE_LAP_RACE, TotalLaps: proto.Int32(6), Field: true,
			FuelUse: eventsv1.FuelUse_FUEL_USE_ON, CarsOnTrack: true, Start: eventsv1.StartType_START_TYPE_ROLLING,
			Evidence: "6 laps, 2 cars; fuel use on; rolling start at 128 km/h",
		},
		At:    1384,
		RunId: "01JB9QYR2T6W8Y1A3C5E7G9HKM",
	}},
	{"events/v1/session_track.json", &eventsv1.SessionTrackChanged{
		SessionId: "01JB9QZ4V3X8M2K7T5R1N6PDWA",
		Track: &eventsv1.TrackVerdict{
			TrackId: proto.Int32(442), Name: "Suzuka Circuit East Course", Confidence: eventsv1.TrackConfidence_TRACK_CONFIDENCE_CONFIRMED,
			Evidence:   "finish line crossed, a full lap fits the layout (1.00)",
			Candidates: []*eventsv1.TrackCandidate{{TrackId: 442, Name: "Suzuka Circuit East Course", Fit: 0.995, LineCrossed: true}},
		},
		At:    4790,
		RunId: "01JB9QYR2T6W8Y1A3C5E7G9HKM",
	}},
	{"events/v1/session_paused.json", &eventsv1.SessionPaused{
		SessionId: "01JB9QZ4V3X8M2K7T5R1N6PDWA",
		Kind:      eventsv1.PauseKind_PAUSE_KIND_FROZEN,
		SinceT:    2412.375,
		Since:     timestamppb.New(time.Date(2026, 10, 18, 13, 25, 41, 250_000_000, time.UTC)),
		PausedMs:  1200,
		At:        9921,
		Pause:     2,
	}},
	{"events/v1/session_resumed.json", &eventsv1.SessionResumed{
		SessionId: "01JB9QZ4V3X8M2K7T5R1N6PDWA",
		Kind:      eventsv1.PauseKind_PAUSE_KIND_FROZEN,
		SinceT:    2412.375,
		Since:     timestamppb.New(time.Date(2026, 10, 18, 13, 25, 41, 250_000_000, time.UTC)),
		PauseMs:   45000,
		PausedMs:  46200,
		At:        12601,
		Pause:     2,
	}},
	{"events/v1/lap_best.json", &eventsv1.LapBestChanged{
		SessionId: "01JB9QZ4V3X8M2K7T5R1N6PDWA", Lap: 5, LapTimeMs: 59541,
		PreviousBestMs: proto.Int32(60538), GapMs: proto.Int32(997), At: 19236, EndedLap: 5,
	}},
	{"events/v1/race_final_lap.json", &eventsv1.RaceFinalLapStarted{
		SessionId: "01JB9QZ4V3X8M2K7T5R1N6PDWA", Lap: 6, At: 19237,
	}},
	{"events/v1/race_position.json", &eventsv1.RacePositionChanged{
		SessionId: "01JB9R1C8E4Q6W2Y9T3K7M5PXB", PreviousPosition: proto.Int32(11), Position: 3, At: 31877,
		RunId: "01JB9R1BX4N7Q2S5V8Z3D6F9JP",
	}},
	{"events/v1/race_finished.json", &eventsv1.RaceFinished{
		SessionId: "01JB9QZ4V3X8M2K7T5R1N6PDWA", Kind: eventsv1.FinishKind_FINISH_KIND_FLAG, Lap: 6,
		Position: proto.Int32(1), TotalTimeMs: proto.Int32(389934), BestLap: proto.Int32(5), BestLapTimeMs: proto.Int32(59541),
		At: 23410,
	}},
	// paddock#70's events. The pit stop is v1's at Gilles Villeneuve
	// (pit_car_telemetry_events_test.go), the teleport v1's reset to track
	// (v1's tests/test_pitstop.py), and the alarm v1's temps rule on
	// real-nuc-tsukuba-brz-drift-2026-10-15's first "Rear tyres hot.".
	{"events/v1/pit_entered.json", &eventsv1.PitEntered{
		SessionId: "01JBBX7Q2M5R8T1V4W6Y9K3NZC", Entry: 1, Stop: 1, InLap: 6, EnteredT: 6853843.563197786,
		EnteredAt:  timestamppb.New(time.Date(2026, 10, 17, 18, 3, 7, 588_057_000, time.UTC)),
		FuelBefore: 8.685, PositionBefore: proto.Int32(5), Evidence: "entry loading cut at 60.0 km/h", At: 11245,
	}},
	{"events/v1/pit_exited.json", &eventsv1.PitExited{
		SessionId: "01JBBX7Q2M5R8T1V4W6Y9K3NZC",
		Stop: &eventsv1.PitStop{
			Stop: 1, InLap: 6, OutLap: proto.Int32(7), PitLaneMs: proto.Int32(30220), StationaryMs: 17130,
			FuelBefore: 8.685, FuelAfter: 100, FuelAdded: 91.315, TyresChanged: true, PitLaneKph: proto.Float64(60),
			PositionBefore: proto.Int32(5), PositionAfter: proto.Int32(10),
			Evidence: "entry loading cut at 60.0 km/h; exit loading cut",
		},
		Stint: 2, Entry: 1,
		EnteredAt: timestamppb.New(time.Date(2026, 10, 17, 18, 3, 7, 588_057_000, time.UTC)),
		ExitedAt:  timestamppb.New(time.Date(2026, 10, 17, 18, 3, 37, 804_885_000, time.UTC)),
		At:        13056,
		ExitedT:   6853873.780025139,
		EnteredT:  6853843.563197786,
	}},
	{"events/v1/pit_teleport.json", &eventsv1.PitTeleport{
		SessionId: "01JBBX7Q2M5R8T1V4W6Y9K3NZC", Entry: 1, EntryAt: proto.Uint32(3), Lap: 3,
		Evidence: "entry loading cut at 40.0 km/h; over 100 km/h; never at rest, no fuel, no tyres", At: 241,
		WithdrawnT: 4,
	}},
	{"events/v1/car_alarm.json", &eventsv1.CarAlarmRaised{
		SessionId: "01JBA2T8M4K6W9X3Q7R5N1PZDE", Check: eventsv1.CarCheck_CAR_CHECK_TYRE_HOT,
		Tyres: []*eventsv1.TyreTemp{
			{Corner: eventsv1.Corner_CORNER_REAR_LEFT, TempC: 176.40371704101562},
			{Corner: eventsv1.Corner_CORNER_REAR_RIGHT, TempC: 169.6006317138672},
		},
		ValueC: 176.40371704101562, ThresholdC: 120, At: 1214, RunId: "01JBA2SZ9C3F6H8K1M4P7R2T5W",
	}},
	{"events/v1/telemetry_lost.json", &eventsv1.TelemetryLost{
		SessionId: "01JBBX7Q2M5R8T1V4W6Y9K3NZC", RunId: "01JBBWZ5C8E2H4K7M9P1R3T6VX",
		Phase: telemetryv1.Phase_PHASE_PAUSED, SilenceMs: 3412, At: 52340,
	}},
	{"events/v1/telemetry_restored.json", &eventsv1.TelemetryRestored{
		RunId: "01JBBWZ5C8E2H4K7M9P1R3T6VX", LostAt: 52340, GapMs: 1843250, At: 52341,
		LostPhase: telemetryv1.Phase_PHASE_PAUSED,
	}},
	// paddock#71's callout: AC3's lap time and telemetry callouts, and four of
	// v1's from the Suzuka no-stop race (callout_test.go).
	{"events/v1/callout_lap_time.json", &eventsv1.Callout{
		SessionId: "01JB8Z3K5Q9T2M7XW4R6N1PDHE", Id: "01JB8Z9F2H4K6M8P0R2T4V6X8Z", Rule: "lap_time", Kind: "delta_slower",
		Slots:    &eventsv1.CalloutSlots{Lap: proto.Int32(3), LapTimeMs: proto.Int32(99000), DeltaMs: proto.Int32(1000)},
		Priority: 20, At: 19236, RunId: "01JB8Z2W6Y8A0C2E4G6J8M0P2R",
		CreatedAt: timestamppb.New(time.Date(2026, 10, 26, 19, 35, 21, 875_000_000, time.UTC)),
		ExpiresAt: timestamppb.New(time.Date(2026, 10, 26, 19, 35, 29, 875_000_000, time.UTC)),
		PitLane:   eventsv1.PitLane_PIT_LANE_HOLD,
	}},
	{"events/v1/callout_telemetry.json", &eventsv1.Callout{
		Id: "01JBBX9A3C5E7G9J1M3P5R7T9V", Rule: "telemetry", Kind: "template_lost", Priority: 5, At: 52340,
		RunId:     "01JBBWZ5C8E2H4K7M9P1R3T6VX",
		CreatedAt: timestamppb.New(time.Date(2026, 10, 17, 19, 12, 4, 250_000_000, time.UTC)),
		ExpiresAt: timestamppb.New(time.Date(2026, 10, 17, 19, 12, 12, 250_000_000, time.UTC)),
		PitLane:   eventsv1.PitLane_PIT_LANE_HOLD,
	}},
	{"events/v1/callout_start_brief.json", &eventsv1.Callout{
		SessionId: "01JB5N2E8R4T6W9Y1A3C5E7G9H", Id: "01JB5N3A1C2E4G6J8M0P2R4T6V", Rule: "start_brief", Kind: "template",
		Slots:    &eventsv1.CalloutSlots{Laps: proto.Int32(6), Position: proto.Int32(10), TotalCars: proto.Int32(20)},
		Priority: 8, At: 1122, RunId: "01JB5N1Z7Q3S5V8X0B2D4F6H8K",
		CreatedAt: timestamppb.New(time.Date(2026, 10, 15, 19, 0, 18, 793_056_000, time.UTC)),
		ExpiresAt: timestamppb.New(time.Date(2026, 10, 15, 19, 0, 26, 793_056_000, time.UTC)),
		PitLane:   eventsv1.PitLane_PIT_LANE_HOLD,
	}},
	{"events/v1/callout_position.json", &eventsv1.Callout{
		SessionId: "01JB5N2E8R4T6W9Y1A3C5E7G9H", Id: "01JB5N3B3D5F7H9K1N3Q5S7W9Y", Rule: "position", Kind: "template_lost",
		Slots:    &eventsv1.CalloutSlots{Position: proto.Int32(12), PreviousPosition: proto.Int32(10)},
		Priority: 12, At: 2100, RunId: "01JB5N1Z7Q3S5V8X0B2D4F6H8K",
		CreatedAt: timestamppb.New(time.Date(2026, 10, 15, 19, 0, 35, 88_536_000, time.UTC)),
		ExpiresAt: timestamppb.New(time.Date(2026, 10, 15, 19, 0, 43, 88_536_000, time.UTC)),
		HoldUntil: timestamppb.New(time.Date(2026, 10, 15, 19, 0, 36, 88_536_000, time.UTC)),
		PitLane:   eventsv1.PitLane_PIT_LANE_DROP,
	}},
	{"events/v1/callout_new_best.json", &eventsv1.Callout{
		SessionId: "01JB5N2E8R4T6W9Y1A3C5E7G9H", Id: "01JB5N3C5E7G9J1M3P5R7T9X1Z", Rule: "lap_time", Kind: "template_new_best",
		Slots:    &eventsv1.CalloutSlots{Lap: proto.Int32(3), LapTimeMs: proto.Int32(63139), DeltaMs: proto.Int32(-1472)},
		Priority: 11, At: 12273, RunId: "01JB5N1Z7Q3S5V8X0B2D4F6H8K",
		CreatedAt: timestamppb.New(time.Date(2026, 10, 15, 19, 3, 24, 808_570_000, time.UTC)),
		ExpiresAt: timestamppb.New(time.Date(2026, 10, 15, 19, 3, 32, 808_570_000, time.UTC)),
		PitLane:   eventsv1.PitLane_PIT_LANE_HOLD,
	}},
	{"events/v1/callout_race_summary.json", &eventsv1.Callout{
		SessionId: "01JB5N2E8R4T6W9Y1A3C5E7G9H", Id: "01JB5N3D7F9H1K3N5Q7S9V1X3A", Rule: "race_summary", Kind: "template_position_podium",
		Slots:    &eventsv1.CalloutSlots{Position: proto.Int32(2), BestLapMs: proto.Int32(61186)},
		Priority: 25, At: 23585, RunId: "01JB5N1Z7Q3S5V8X0B2D4F6H8K",
		CreatedAt: timestamppb.New(time.Date(2026, 10, 15, 19, 6, 33, 531_882_000, time.UTC)),
		ExpiresAt: timestamppb.New(time.Date(2026, 10, 15, 19, 7, 3, 531_882_000, time.UTC)),
		PitLane:   eventsv1.PitLane_PIT_LANE_SPEAK,
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

// Each message survives protobuf's binary form too, which a subject may move to
// later (ADR 0002, point 2).
func TestMessagesRoundTripThroughBinary(t *testing.T) {
	for _, ex := range examples {
		t.Run(ex.file, func(t *testing.T) {
			data, err := proto.Marshal(ex.want)
			if err != nil {
				t.Fatalf("encoding: %v", err)
			}
			got := ex.want.ProtoReflect().New().Interface()
			if err := proto.Unmarshal(data, got); err != nil {
				t.Fatalf("decoding: %v", err)
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
