package contracts_test

import (
	"os"
	"path/filepath"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/those-who-run/contracts/ce"
	eventsv1 "github.com/those-who-run/contracts/gen/go/twr/events/v1"
	pitwallv1 "github.com/those-who-run/contracts/gen/go/twr/pitwall/v1"
	storev1 "github.com/those-who-run/contracts/gen/go/twr/store/v1"
)

// paddock#121's examples: logbook's answers about the Suzuka East pit stop
// race of session_ended_suzuka_pit_stop.json, and the pit wall's API's answers
// mapped from them, which messages_test.go decodes and round-trips with the
// rest.
func init() {
	examples = append(examples, serviceExamples()...)
}

const suzukaSID = "01JB9QZ4V3X8M2K7T5R1N6PDWA"

// suzukaPitStop is session_ended_suzuka_pit_stop.json, the event the store's
// row is kept from.
func suzukaPitStop() *eventsv1.SessionEnded {
	data, err := os.ReadFile(filepath.Join("examples", "events", "v1", "session_ended_suzuka_pit_stop.json"))
	if err != nil {
		panic(err)
	}
	var ended eventsv1.SessionEnded
	if err := ce.Unmarshal(data, &ended); err != nil {
		panic(err)
	}
	return &ended
}

// suzukaRow is the race as logbook keeps it: its summary, laps left out, and
// when podium posted its card.
func suzukaRow() (*storev1.SessionRow, []*eventsv1.LapCompleted) {
	ended := suzukaPitStop()
	summary := proto.CloneOf(ended.GetSummary())
	laps := summary.GetLaps()
	summary.Laps = nil
	return &storev1.SessionRow{
		SessionId:    suzukaSID,
		Source:       "live",
		Closed:       true,
		EndReason:    ended.GetReason(),
		Summary:      summary,
		CardPostedAt: timestamppb.New(time.Date(2026, 10, 18, 13, 28, 3, 250_000_000, time.UTC)),
	}, laps
}

// suzukaSession is the race as the pit wall's API gives it.
func suzukaSession() *pitwallv1.Session {
	return &pitwallv1.Session{
		SessionId: suzukaSID,
		Title:     "Suzuka Circuit East Course · 6-lap race",
		Result:    "Finished P1 of 2 in 6:29.934",
		StartedAt: timestamppb.New(time.Date(2026, 10, 18, 13, 21, 29, 15_985_000, time.UTC)),
		EndedAt:   timestamppb.New(time.Date(2026, 10, 18, 13, 28, 1, 176_172_000, time.UTC)),
		Closed:    true,
		Kind:      "lap_race",
		KindLabel: "6-lap race",
		Track: &pitwallv1.Track{
			TrackId: proto.Int32(442), Name: "Suzuka Circuit East Course", Sure: true,
			Evidence: proto.String("finish line crossed, a full lap fits the layout (1.00)"),
		},
		Car:           proto.String("Honda Civic Si Group A version '87"),
		CarCode:       proto.Int32(683),
		Finished:      true,
		FinalPosition: proto.Int32(1),
		TotalCars:     proto.Int32(2),
		PositionShown: true,
		TotalTimeMs:   proto.Int32(389934),
		BestLapMs:     proto.Int32(59541),
		BestLap:       proto.Int32(5),
		LapsCompleted: 6,
		TotalLaps:     proto.Int32(6),
		FuelUse:       proto.Bool(true),
		FuelBurned:    proto.Float64(130.393),
		FuelCapacity:  proto.Float64(100),
		FuelByStint:   []float64{73.825, 56.568},
		PitStops:      1,
		DrivingMs:     proto.Int32(390108),
		PausedMs:      proto.Int32(2052),
		CardPostedAt:  timestamppb.New(time.Date(2026, 10, 18, 13, 28, 3, 250_000_000, time.UTC)),
	}
}

func serviceExamples() []struct {
	file string
	want proto.Message
} {
	row, laps := suzukaRow()
	return []struct {
		file string
		want proto.Message
	}{
		{"store/v1/list_sessions_request.json", &storev1.ListSessionsRequest{
			Sources:     []string{"live", "v1-import"},
			Track:       proto.String("Suzuka Circuit East Course"),
			SessionType: eventsv1.SessionType_SESSION_TYPE_LAP_RACE,
			FromDay:     "2026-10-01",
			TimeZone:    "Europe/Dublin",
			Limit:       200,
		}},
		{"store/v1/get_session_response.json", &storev1.GetSessionResponse{
			Session: row, DataVersion: "5:1829",
		}},
		// The in-lap and the out-lap of its stop.
		{"store/v1/list_laps_response.json", &storev1.ListLapsResponse{
			Session: row,
			Laps: []*storev1.LapRow{
				{Lap: laps[2], Frames: 4953, TopSpeedKph: proto.Float64(171.4)},
				{Lap: laps[3], Frames: 4441, TopSpeedKph: proto.Float64(168.9)},
			},
			DataVersion: "5:1829",
		}},
		{"store/v1/get_trace_response.json", &storev1.GetTraceResponse{
			Trace: &storev1.LapTrace{
				SessionId: suzukaSID, Lap: 5,
				T:           []float64{0, 0.017, 21.384, 59.541},
				SpeedKph:    []float64{142.7, 142.9, 63.2, 151.8},
				ThrottlePct: []float64{100, 100, 0, 100},
				BrakePct:    []float64{0, 0, 87.1, 0},
				Gear:        []int32{4, 4, 2, 4},
				Rpm:         []float64{7210, 7235, 5480, 7590},
				X:           []float64{-412.5, -410.1, 118.3, -414.2},
				Z:           []float64{1088.8, 1086.3, 512.6, 1092.4},
			},
			DataVersion: "5:1829:lap-5",
		}},
		{"pitwall/v1/get_session_response.json", &pitwallv1.GetSessionResponse{Session: suzukaSession()}},
		{"pitwall/v1/list_sessions_response.json", &pitwallv1.ListSessionsResponse{
			Sessions: []*pitwallv1.Session{suzukaSession()},
			Options: &pitwallv1.FilterOptions{
				Track: []*pitwallv1.FilterOption{{Value: "Suzuka Circuit East Course", Count: 1}, {Value: "Unknown track", Count: 0}},
				Car:   []*pitwallv1.FilterOption{{Value: "Honda Civic Si Group A version '87", Count: 1}},
				Kind:  []*pitwallv1.KindOption{{Value: "lap_race", Label: "lap race", Count: 1}, {Value: "solo_timed", Label: "time trial or practice"}},
			},
			Total:   14,
			Matched: 1,
		}},
		{"pitwall/v1/list_laps_response.json", &pitwallv1.ListLapsResponse{
			SessionId: suzukaSID,
			Laps: []*pitwallv1.LapRow{
				{
					Lap: 3, TimeMs: proto.Int32(82544), Complete: true,
					FuelStart: proto.Float64(37.908443450927734), FuelEnd: proto.Float64(22.89894676208496), FuelUsed: proto.Float64(15.009),
					Position: proto.Int32(1), Frames: 4953, TopSpeedKph: proto.Float64(171.4),
					Pit: true, FuelUse: true,
				},
				{
					Lap: 4, TimeMs: proto.Int32(69892), Complete: true,
					FuelStart: proto.Float64(22.89894676208496), FuelEnd: proto.Float64(81.47602844238281), FuelUsed: proto.Float64(-58.577),
					Position: proto.Int32(1), Frames: 4441, TopSpeedKph: proto.Float64(168.9),
					OutLap: true, FuelUse: true,
				},
			},
		}},
		{"pitwall/v1/get_trace_response.json", &pitwallv1.GetTraceResponse{
			Trace: &pitwallv1.Trace{
				SessionId: suzukaSID, Lap: 5,
				T:           []float64{0, 0.017, 21.384, 59.541},
				SpeedKph:    []float64{142.7, 142.9, 63.2, 151.8},
				ThrottlePct: []float64{100, 100, 0, 100},
				BrakePct:    []float64{0, 0, 87.1, 0},
				Gear:        []int32{4, 4, 2, 4},
				Rpm:         []float64{7210, 7235, 5480, 7590},
				X:           []float64{-412.5, -410.1, 118.3, -414.2},
				Z:           []float64{1088.8, 1086.3, 512.6, 1092.4},
			},
		}},
		// Mid-race, lap 3, the driver listening on Discord.
		{"pitwall/v1/get_engine_status_response.json", &pitwallv1.GetEngineStatusResponse{
			Status: &pitwallv1.EngineStatus{
				Version:        "0.1.14",
				StartedAt:      timestamppb.New(time.Date(2026, 10, 18, 9, 2, 41, 0, time.UTC)),
				Activity:       pitwallv1.Activity_ACTIVITY_DRIVING,
				Telemetry:      &pitwallv1.TelemetryLink{Connected: true, LastPacketAgeSecs: proto.Float64(0)},
				VoiceConnected: proto.Bool(true),
				Session: &pitwallv1.LiveSession{
					SessionId: suzukaSID, Kind: "lap_race",
					Track: &pitwallv1.LiveTrack{TrackId: proto.Int32(442), Name: proto.String("Suzuka Circuit East Course"), Sure: true},
					Car:   proto.String("Honda Civic Si Group A version '87"),
					Lap:   proto.Int32(3), TotalLaps: proto.Int32(6), Position: proto.Int32(1), TotalCars: proto.Int32(2),
				},
			},
		}},
		// ecu gone: no status, or one past the stale limit.
		{"pitwall/v1/get_engine_status_offline.json", &pitwallv1.GetEngineStatusResponse{
			Status: &pitwallv1.EngineStatus{Activity: pitwallv1.Activity_ACTIVITY_OFFLINE},
		}},
	}
}
