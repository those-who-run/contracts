package contracts_test

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/those-who-run/contracts/ce"
	eventsv1 "github.com/those-who-run/contracts/gen/go/twr/events/v1"
	statev1 "github.com/those-who-run/contracts/gen/go/twr/state/v1"
)

// OpenSession holds what ecu needs to continue a session after a restart, or to
// end it from the KV alone (paddock#72, R2; paddock#76). open_session.json is
// the Suzuka timed race v1 continued across an engineer restart on lap 7
// (real-nuc-suzuka-before-restart-2026-10-18.jsonl.gz, then
// joined-mid-race), as ecu would have kept it when it stopped: its laps, stop,
// track and kind are session_ended_suzuka_continued.json's, which
// summary_v1_test.go checks against v1's summary.json.
func TestOpenSessionIsTheSessionBeforeTheRestart(t *testing.T) {
	open := example(t, "state/v1/open_session.json", &statev1.OpenSession{}).(*statev1.OpenSession)
	ended := example(t, "events/v1/session_ended_suzuka_continued.json", &eventsv1.SessionEnded{}).(*eventsv1.SessionEnded)
	s, final := open.GetSummary(), ended.GetSummary()
	gap := final.GetGaps()[0]

	if open.GetSessionId() != ended.GetSessionId() {
		t.Errorf("session %s, want %s", open.GetSessionId(), ended.GetSessionId())
	}
	// v1's continue match (continue_mismatch): the car, the type, the lap
	// total, the field, the last lap and the time since it was last seen.
	if s.GetCarCode() != final.GetCarCode() || s.GetSessionType() != final.GetSessionType() ||
		s.GetTotalLaps() != final.GetTotalLaps() || s.GetTotalCars() != final.GetTotalCars() {
		t.Error("the car, the type, the lap total or the field is not the session's")
	}
	if open.GetLap().GetLap() != gap.GetLap() {
		t.Errorf("lap %d, want the gap's lap %d", open.GetLap().GetLap(), gap.GetLap())
	}
	if !proto.Equal(open.GetLastSeenAt(), gap.GetStartedAt()) {
		t.Errorf("last seen %v, want the gap's start %v", open.GetLastSeenAt().AsTime(), gap.GetStartedAt().AsTime())
	}
	// The track and the kind are carried over.
	if !proto.Equal(s.GetTrack(), final.GetTrack()) {
		t.Errorf("track %v, want %v", s.GetTrack(), final.GetTrack())
	}
	kind := proto.Clone(s.GetKind()).(*eventsv1.SessionKind)
	kind.Evidence = final.GetKind().GetEvidence()
	if !proto.Equal(kind, final.GetKind()) {
		t.Errorf("kind %v, want %v", s.GetKind(), final.GetKind())
	}
	// The laps a LapCompleted carried before the restart, and the stop.
	for i, lap := range s.GetLaps() {
		if !proto.Equal(lap, final.GetLaps()[i]) {
			t.Errorf("lap %d is %v, want %v", i+1, lap, final.GetLaps()[i])
		}
		if lap.GetLap() >= open.GetLap().GetLap() {
			t.Errorf("lap %d is in laps, but the car is on it", lap.GetLap())
		}
	}
	if len(s.GetPitStops()) != 1 || !proto.Equal(s.GetPitStops()[0], final.GetPitStops()[0]) {
		t.Errorf("pit stops %v, want the stop on lap 4", s.GetPitStops())
	}
	if stops := len(s.GetPitStops()); int(s.GetStints()) != stops+1 || int(open.GetEntries()) != stops {
		t.Errorf("%d stints and %d entries for %d stops", s.GetStints(), open.GetEntries(), len(s.GetPitStops()))
	}
	// The pause on lap 3 is the only one, and the session paused before the
	// restart only.
	if open.GetPauses() != 1 || s.GetPausedMs() != final.GetPausedMs() || open.GetPause() != nil {
		t.Errorf("%d pauses, %d ms paused, open pause %v", open.GetPauses(), s.GetPausedMs(), open.GetPause())
	}
	for _, v := range []struct {
		name string
		set  bool
	}{
		{"started_at", proto.Equal(s.GetStartedAt(), final.GetStartedAt())},
		{"best_lap", s.GetBestLap() == final.GetBestLap() && s.GetBestLapTimeMs() == final.GetBestLapTimeMs()},
		{"game_best_lap_ms", s.GetGameBestLapMs() == final.GetGameBestLapMs()},
		{"no ended_at while open", s.GetEndedAt() == nil && !s.GetEndedPaused()},
		{"source and run", open.GetSource() == "live" && open.GetRunId() != ""},
		// A timed session never has a final lap (paddock#84, R1).
		{"no final lap in a timed session", !open.GetFinalLapStarted()},
		// The lead on lap 7, settled before the restart (paddock#84, R2).
		{"published_position", open.PublishedPosition != nil && open.GetPublishedPosition() == final.GetFinalPosition()},
	} {
		if !v.set {
			t.Errorf("%s is not what the session had before the restart", v.name)
		}
	}
}

// The session can be ended from the value alone (paddock#76, R5): the lap the
// car was on becomes the last, partial lap, and the session.ended goes
// through ce like any other.
func TestOpenSessionEndsFromTheKVAlone(t *testing.T) {
	open := throughJSON(t, example(t, "state/v1/open_session.json", &statev1.OpenSession{}).(*statev1.OpenSession))
	summary := proto.Clone(open.GetSummary()).(*eventsv1.SessionSummary)
	summary.EndedAt = open.GetLastSeenAt()
	summary.Laps = append(summary.Laps, &eventsv1.LapCompleted{
		SessionId: open.GetSessionId(), Lap: open.GetLap().GetLap(), FuelAtStart: open.GetLap().GetFuelAtStart(),
		FuelAtEnd: summary.GetFuelAtEnd(), FuelUsed: open.GetLap().GetFuelAtStart() - summary.GetFuelAtEnd(),
		Position: summary.FinalPosition, PitRole: open.GetLap().GetPitRole(),
	})
	end := &eventsv1.SessionEnded{
		SessionId: open.GetSessionId(), Reason: eventsv1.EndReason_END_REASON_ENGINEER_STOPPED, Summary: summary,
	}
	got := throughCE(t, end, open.GetSource()+":"+open.GetSessionId()+":session.ended:1")
	if !proto.Equal(got, end) {
		t.Errorf("got %v\nwant %v", got, end)
	}
	if n := len(got.GetSummary().GetLaps()); n != 7 || got.GetSummary().GetLaps()[6].GetComplete() {
		t.Errorf("%d laps, want 6 complete and lap 7 partial", n)
	}
	if got.GetSummary().GetLapsCompleted() != 6 {
		t.Errorf("%d laps completed, want 6", got.GetSummary().GetLapsCompleted())
	}
	if _, err := ce.Type(got); err != nil {
		t.Error(err)
	}
}
