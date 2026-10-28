package contracts_test

import (
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/those-who-run/contracts/ce"
	eventsv1 "github.com/those-who-run/contracts/gen/go/twr/events/v1"
)

// SessionSummary against v1's summary.json (paddock#69, AC3 and AC4): the
// summaries v1.0.0 writes for two of its recordings, kept as v1 wrote them
// under testdata/v1, and the session.ended examples made from them.
//
//   - suzuka_pit_stop: real-nuc-suzuka-pit-stop-2026-10-15.jsonl.gz, a lap race
//     with two stints and one stop, ended at the flag;
//   - suzuka_continued: real-nuc-suzuka-before-restart-2026-10-18.jsonl.gz then
//     real-nuc-suzuka-joined-mid-race-2026-10-18.jsonl.gz, a timed race
//     continued across an engineer restart on lap 7 (one gap), ended where the
//     frames stop.
//
// Both were replayed through v1's engine and store as its tests/test_restart.py
// does, with that test's clock, so started_at and ended_at are that clock's.
var v1Summaries = []struct {
	name   string
	reason eventsv1.EndReason
}{
	{"suzuka_pit_stop", eventsv1.EndReason_END_REASON_FINISHED},
	{"suzuka_continued", eventsv1.EndReason_END_REASON_INPUT_ENDED},
}

// The summary.json fields SessionSummary leaves out on purpose, each with why
// (AC4). The message's comment says the same.
var leftOut = map[string]string{
	"session_id":        "SessionEnded's session_id",
	"type":              "v1's race, time trial or practice guess; session_type replaces it (paddock#276)",
	"track_id":          "track's track_id",
	"engineer_shutdown": "END_REASON_ENGINEER_STOPPED, with no session.ended on a shutdown (paddock#76)",
	"laps.frames":       "v1's store's own: a lap's frames are bounded by first_seq and last_seq",
	"laps.file":         "v1's store's own: the lap's Parquet file",
	"laps.recorded_at":  "v1's store's own: when it wrote the lap",
}

// v1Record is a summary.json object being read: each field taken is removed,
// so what is left over is a field the message does not carry.
type v1Record struct {
	t      *testing.T
	prefix string
	m      map[string]any
}

func (r v1Record) take(key string) any {
	r.t.Helper()
	v, ok := r.m[key]
	if !ok {
		r.t.Fatalf("summary.json has no %s%s", r.prefix, key)
	}
	delete(r.m, key)
	return v
}

func (r v1Record) str(key string) string {
	s, _ := r.take(key).(string)
	return s
}

func (r v1Record) num(key string) float64 {
	n, _ := r.take(key).(float64)
	return n
}

func (r v1Record) i32(key string) int32 { return int32(r.num(key)) }

func (r v1Record) flag(key string) bool {
	b, _ := r.take(key).(bool)
	return b
}

// opt is an optional int32: absent for v1's null.
func (r v1Record) opt(key string) *int32 {
	if n, ok := r.take(key).(float64); ok {
		return proto.Int32(int32(n))
	}
	return nil
}

// nonZero is an optional int32 that v1 wrote as 0 for absent, as a timed
// session's lap total.
func (r v1Record) nonZero(key string) *int32 {
	if n := r.i32(key); n != 0 {
		return proto.Int32(n)
	}
	return nil
}

func (r v1Record) at(key string) *timestamppb.Timestamp {
	sec, frac := math.Modf(r.num(key))
	return timestamppb.New(time.Unix(int64(sec), int64(math.Round(frac*1e6))*1e3).UTC())
}

func (r v1Record) ms(key string) int32 { return int32(math.Round(r.num(key) * 1000)) }

func (r v1Record) doubles(key string) []float64 {
	var out []float64
	for _, v := range r.take(key).([]any) {
		out = append(out, v.(float64))
	}
	return out
}

func (r v1Record) objects(key string) []v1Record {
	var out []v1Record
	for _, v := range r.take(key).([]any) {
		out = append(out, v1Record{r.t, key + ".", v.(map[string]any)})
	}
	return out
}

func (r v1Record) object(key string) (v1Record, bool) {
	m, ok := r.take(key).(map[string]any)
	return v1Record{r.t, key + ".", m}, ok
}

// done fails on a field the conversion did not take and leftOut does not list.
func (r v1Record) done() {
	r.t.Helper()
	for key := range r.m {
		if _, ok := leftOut[r.prefix+key]; !ok {
			r.t.Errorf("summary.json's %s%s is neither in SessionSummary nor in leftOut", r.prefix, key)
		}
	}
}

func upper(prefix, v1 string) string { return prefix + strings.ToUpper(v1) }

// fromV1 is the SessionSummary each field of v1's summary.json maps to, by the
// v1 names the message's comments give. counts_for_best, which v1 did not
// store, and finishing_position, which v2 sets only for a lap race at the flag,
// are left to the caller.
func fromV1(t *testing.T, data []byte) *eventsv1.SessionSummary {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	r := v1Record{t, "", m}
	s := &eventsv1.SessionSummary{
		LapsCompleted:     r.i32("laps_completed"),
		BestLapTimeMs:     r.opt("best_lap_ms"),
		FuelUsed:          r.num("fuel_used"),
		FinalPosition:     r.opt("final_position"),
		StartedAt:         r.at("started_at"),
		EndedAt:           r.at("ended_at"),
		DrivingMs:         r.i32("driving_ms"),
		PausedMs:          r.i32("paused_ms"),
		EndedPaused:       r.flag("ended_paused"),
		CarCode:           r.i32("car_code"),
		CarName:           r.str("car_name"),
		TotalLaps:         r.nonZero("total_laps"),
		FastestLapMs:      r.opt("fastest_lap_ms"),
		GameBestLapMs:     r.opt("gt7_best_lap_ms"),
		Finished:          r.flag("finished"),
		JoinedOnLap:       r.opt("joined_on_lap"),
		TotalTimeMs:       r.opt("total_time_ms"),
		TotalCars:         r.opt("total_cars"),
		FuelBurned:        r.num("fuel_burned"),
		FuelBurnedByStint: r.doubles("fuel_burned_by_stint"),
		FuelStints:        r.doubles("fuel_stints"),
		FuelCapacity:      r.num("fuel_capacity"),
		FuelAtStart:       r.num("fuel_at_start"),
		FuelAtEnd:         r.num("fuel_at_end"),
		Stints:            r.i32("stints"),
	}
	if k, ok := r.object("session_kind"); ok {
		s.Kind = &eventsv1.SessionKind{
			SessionType: eventsv1.SessionType(eventsv1.SessionType_value[upper("SESSION_TYPE_", k.str("kind"))]),
			TotalLaps:   k.nonZero("total_laps"),
			Field:       k.flag("field"),
			CarsOnTrack: k.flag("cars_on_track"),
			Start:       eventsv1.StartType(eventsv1.StartType_value[upper("START_TYPE_", k.str("start"))]),
			Evidence:    k.str("evidence"),
		}
		switch k.take("fuel_use") {
		case true:
			s.Kind.FuelUse = eventsv1.FuelUse_FUEL_USE_ON
		case false:
			s.Kind.FuelUse = eventsv1.FuelUse_FUEL_USE_OFF
		default:
			s.Kind.FuelUse = eventsv1.FuelUse_FUEL_USE_UNKNOWN
		}
		s.SessionType = s.Kind.SessionType
		k.done()
	}
	if tr, ok := r.object("track"); ok {
		s.Track = &eventsv1.TrackVerdict{
			TrackId:    tr.opt("id"),
			Name:       tr.str("name"),
			Confidence: eventsv1.TrackConfidence(eventsv1.TrackConfidence_value[upper("TRACK_CONFIDENCE_", tr.str("confidence"))]),
			Family:     tr.str("family"),
			Evidence:   tr.str("evidence"),
		}
		for _, c := range tr.objects("candidates") {
			s.Track.Candidates = append(s.Track.Candidates, &eventsv1.TrackCandidate{
				TrackId: c.i32("id"), Name: c.str("name"), Fit: c.num("fit"), LineCrossed: c.flag("line"),
			})
			c.done()
		}
		tr.done()
	}
	for _, p := range r.objects("pit_stops") {
		stop := &eventsv1.PitStop{
			Stop:           p.i32("stop"),
			InLap:          p.i32("in_lap"),
			OutLap:         p.opt("out_lap"),
			StationaryMs:   p.ms("stationary_secs"),
			FuelBefore:     p.num("fuel_before"),
			FuelAfter:      p.num("fuel_after"),
			FuelAdded:      p.num("fuel_added"),
			TyresChanged:   p.flag("tyres_changed"),
			PositionBefore: p.opt("position_before"),
			PositionAfter:  p.opt("position_after"),
			Evidence:       p.str("evidence"),
		}
		if v, ok := p.m["pit_lane_secs"].(float64); ok {
			stop.PitLaneMs = proto.Int32(int32(math.Round(v * 1000)))
		}
		delete(p.m, "pit_lane_secs")
		if v, ok := p.take("pit_lane_kph").(float64); ok {
			stop.PitLaneKph = proto.Float64(v)
		}
		s.PitStops = append(s.PitStops, stop)
		p.done()
	}
	for _, g := range r.objects("gaps") {
		s.Gaps = append(s.Gaps, &eventsv1.SessionGap{
			StartedAt: g.at("from"), EndedAt: g.at("to"), Lap: g.i32("lap"),
			Reason: map[string]eventsv1.GapReason{"engineer restart": eventsv1.GapReason_GAP_REASON_ENGINEER_RESTART}[g.str("reason")],
		})
		g.done()
	}
	roles := map[any]eventsv1.PitRole{
		nil: eventsv1.PitRole_PIT_ROLE_NONE, "in": eventsv1.PitRole_PIT_ROLE_IN_LAP,
		"out": eventsv1.PitRole_PIT_ROLE_OUT_LAP, "in_out": eventsv1.PitRole_PIT_ROLE_IN_AND_OUT,
	}
	for _, l := range r.objects("laps") {
		lap := &eventsv1.LapCompleted{
			Lap:         l.i32("lap"),
			Complete:    l.flag("complete"),
			FuelAtStart: l.num("fuel_start"),
			FuelAtEnd:   l.num("fuel_end"),
			FuelUsed:    l.num("fuel_used"),
			Position:    l.opt("position"),
			PitRole:     roles[l.take("pit")],
		}
		if ms := l.opt("lap_time_ms"); ms != nil {
			lap.LapTimeMs = *ms
		}
		s.Laps = append(s.Laps, lap)
		l.done()
	}
	r.done()
	return s
}

// Each session.ended example is v1's summary.json for its recording, field by
// field (AC3), every v1 field is in the message or in leftOut (AC4), and the
// example goes through ce, the way ecu publishes and a consumer reads it.
func TestSummariesFromV1(t *testing.T) {
	for _, tc := range v1Summaries {
		t.Run(tc.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", "v1", tc.name+"_summary.json"))
			if err != nil {
				t.Fatal(err)
			}
			want := fromV1(t, data)

			var example eventsv1.SessionEnded
			if err := ce.Unmarshal(readExample(t, "events/v1/session_ended_"+tc.name+".json"), &example); err != nil {
				t.Fatal(err)
			}
			h, payload, err := ce.Encode(&example, ce.Event{
				ID: "live:" + example.GetSessionId() + ":session.ended:1", Source: "/twr/ecu/gt7", Time: example.GetSummary().GetEndedAt().AsTime(),
			})
			if err != nil {
				t.Fatal(err)
			}
			var got eventsv1.SessionEnded
			if _, err := ce.Decode(h, payload, &got); err != nil {
				t.Fatal(err)
			}
			if !proto.Equal(&got, &example) {
				t.Fatalf("through ce: got %v, want %v", &got, &example)
			}
			if got.GetReason() != tc.reason {
				t.Errorf("reason %v, want %v", got.GetReason(), tc.reason)
			}

			summary := proto.Clone(got.GetSummary()).(*eventsv1.SessionSummary)
			checkBest(t, summary)
			// What v1 did not store: which laps count for best, the best lap's
			// number, the laps' session id, and v2's finishing position.
			want.BestLap, want.FinishingPosition = summary.BestLap, summary.FinishingPosition
			for _, lap := range summary.GetLaps() {
				if lap.GetSessionId() != got.GetSessionId() {
					t.Errorf("lap %d's session_id is %q, not the event's", lap.GetLap(), lap.GetSessionId())
				}
				lap.SessionId, lap.CountsForBest = "", false
			}
			if !proto.Equal(summary, want) {
				t.Errorf("the example is not v1's summary:\ngot  %v\nwant %v", summary, want)
			}
		})
	}
}

// best_lap names a lap that counts for best, with best_lap_time_ms, and no lap
// that counts is quicker: counts_for_best, which v1 did not store, agrees with
// v1's best_lap_ms.
func checkBest(t *testing.T, s *eventsv1.SessionSummary) {
	t.Helper()
	i := slices.IndexFunc(s.GetLaps(), func(l *eventsv1.LapCompleted) bool { return l.GetLap() == s.GetBestLap() })
	if i < 0 || !s.GetLaps()[i].GetCountsForBest() || s.GetLaps()[i].GetLapTimeMs() != s.GetBestLapTimeMs() {
		t.Errorf("best_lap %d is not a lap that counts for best with best_lap_time_ms %d", s.GetBestLap(), s.GetBestLapTimeMs())
	}
	for _, l := range s.GetLaps() {
		if l.GetCountsForBest() && (!l.GetComplete() || l.GetPitRole() != eventsv1.PitRole_PIT_ROLE_NONE || l.GetLapTimeMs() < s.GetBestLapTimeMs()) {
			t.Errorf("lap %d counts for best, but is partial, a pit lap or quicker than the best", l.GetLap())
		}
	}
}
