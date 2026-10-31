package format

import (
	"fmt"
	"strconv"

	eventsv1 "github.com/those-who-run/contracts/gen/go/twr/events/v1"
)

// The results card's words for a session: its title, its result and its kind,
// from ecu's SessionSummary. podium's card and the pit wall's API
// (twr.pitwall.v1.Session's title, result and kind_label) both call them, so
// the card and the page never name the same session differently (ADR 0004
// point 5). They are v1 1.0.0's results.build_card and its helpers, line for
// line, with the pit wall's own words for a session still open or with no full
// lap (v1's pitwall/data.py), and are checked against v1's cards by the
// "cards" cases in format-cases.json.
//
// v1's `type`, its first-frame guess of race, time trial or practice, has no
// v2 field (the summary's kind is ecu's final verdict), so a session whose kind
// ecu could not tell is a "session", where v1 could say "race" or "time trial".

// GlyphFlag comes before a finished result: v1's brand.GLYPH_FLAG.
const GlyphFlag = "🚩"

// Title is the card's title, "<track> · <kind>": "Sportsland SUGO · 5-lap
// race", "Unknown track · time trial".
func Title(s *eventsv1.SessionSummary) string {
	return TrackLabel(s.GetTrack()) + " · " + KindLabel(sessionType(s), int(s.GetTotalLaps()))
}

// Result is the card's result line for a session that ended for reason, with
// GlyphFlag before a finished one: "🚩 Finished P3 of 20 in 8:14.302",
// "Retired on lap 3 of 5, running P3", "Paused on lap 3 of 5, P3",
// "Engineer stopped on lap 7, P1", "2 laps, P3 of 20".
//
// The pit wall's rows add their own words, as v1's did: "Running, 2 laps"
// while the session is open (closed false), where the card would say it
// retired, and "no full lap" (or "Running, no full lap yet") where no lap was
// completed and podium posts no card.
//
// lastLap is the highest lap seen, for a session ecu's stop ended; 0 works it
// out from the summary's laps, or the lap after the last completed one without
// them (store.v1.SessionRow's last_lap, where the summary has no laps).
func Result(s *eventsv1.SessionSummary, reason eventsv1.EndReason, closed bool, lastLap int) string {
	laps := int(s.GetLapsCompleted())
	position := shownPosition(s)
	switch {
	case laps < 1 && !closed:
		return "Running, no full lap yet"
	case laps < 1 && s.GetEndedPaused():
		return pausedResult(s, position) + ", no full lap"
	case laps < 1:
		return "no full lap"
	case !closed:
		return "Running, " + plural(laps, "lap")
	}

	totalLaps := int(s.GetTotalLaps())
	ofCars := ""
	if cars := s.GetTotalCars(); cars > 0 {
		ofCars = " of " + strconv.Itoa(int(cars))
	}
	switch {
	case s.GetFinished():
		result := "Finished"
		if position > 0 {
			result = fmt.Sprintf("Finished P%d%s", position, ofCars)
		}
		if total := int(s.GetTotalTimeMs()); total > 0 {
			result += " in " + RaceTime(total)
		}
		return GlyphFlag + " " + result
	case s.GetEndedPaused():
		return pausedResult(s, position)
	case reason == eventsv1.EndReason_END_REASON_ENGINEER_STOPPED:
		return stoppedResult(s, position, lastLap)
	case sessionType(s) == eventsv1.SessionType_SESSION_TYPE_LAP_RACE && totalLaps > 0:
		result := fmt.Sprintf("Retired on lap %d of %d", min(laps+1, totalLaps), totalLaps)
		if position > 0 {
			result += fmt.Sprintf(", running P%d", position)
		}
		return result
	}
	result := plural(laps, "lap")
	if position > 0 {
		result += fmt.Sprintf(", P%d%s", position, ofCars)
	}
	return result
}

// KindLabel is the card's kind of session, from ecu's verdict on it (kind) and
// the laps the race runs: "6-lap race", "timed race", "time trial", "3-lap
// run", and "qualifying or practice" for a session ranked by lap time with a
// field (SESSION_TYPE_RANKED; telemetry can't tell the two apart). A kind ecu
// could not tell is "session".
func KindLabel(kind eventsv1.SessionType, totalLaps int) string {
	switch kind {
	case eventsv1.SessionType_SESSION_TYPE_LAP_RACE:
		return fmt.Sprintf("%d-lap race", totalLaps)
	case eventsv1.SessionType_SESSION_TYPE_TIMED_FIELD:
		return "timed race"
	case eventsv1.SessionType_SESSION_TYPE_RANKED:
		return "qualifying or practice"
	case eventsv1.SessionType_SESSION_TYPE_SOLO_TIMED:
		return "time trial"
	case eventsv1.SessionType_SESSION_TYPE_SOLO_LAPS:
		return fmt.Sprintf("%d-lap run", totalLaps)
	}
	return "session"
}

// TrackLabel is the card's track: the layout's name when ecu is sure of it
// (probable or confirmed), else the circuit's, else "Unknown track".
func TrackLabel(t *eventsv1.TrackVerdict) string {
	switch c := t.GetConfidence(); {
	case (c == eventsv1.TrackConfidence_TRACK_CONFIDENCE_PROBABLE ||
		c == eventsv1.TrackConfidence_TRACK_CONFIDENCE_CONFIRMED) && t.GetName() != "":
		return t.GetName()
	case t.GetFamily() != "":
		return t.GetFamily()
	}
	return "Unknown track"
}

// PositionShown says whether a position means anything, as the engineer says
// it: the kind has a field, and GT7 flags the car on track with data (the
// cars-on-track flag, which says nothing of other cars).
func PositionShown(kind *eventsv1.SessionKind) bool {
	return kind.GetField() && kind.GetCarsOnTrack()
}

// sessionType is ecu's final kind of the session, from its kind verdict as
// v1's card read it.
func sessionType(s *eventsv1.SessionSummary) eventsv1.SessionType {
	if kind := s.GetKind(); kind != nil {
		return kind.GetSessionType()
	}
	return eventsv1.SessionType_SESSION_TYPE_UNKNOWN
}

// shownPosition is the final position where PositionShown, else 0, none.
func shownPosition(s *eventsv1.SessionSummary) int {
	if PositionShown(s.GetKind()) {
		return int(s.GetFinalPosition())
	}
	return 0
}

// pausedResult is a session that ended in a pause: "Paused on lap 1, P2",
// "Paused on lap 3 of 5". It is not a retirement: the race may still be paused
// on the PS5. position 0 is none shown.
func pausedResult(s *eventsv1.SessionSummary, position int) string {
	lap, total := int(s.GetLapsCompleted())+1, int(s.GetTotalLaps())
	if total > 0 {
		return withPosition(fmt.Sprintf("Paused on lap %d of %d", min(lap, total), total), position)
	}
	return withPosition(fmt.Sprintf("Paused on lap %d", lap), position)
}

// stoppedResult is a session ecu's stop ended and no restart continued:
// "Engineer stopped on lap 7, P1". It is not a retirement: the race went on.
// The lap is the last one seen. position 0 is none shown.
func stoppedResult(s *eventsv1.SessionSummary, position, lastLap int) string {
	lap := lastLap
	if lap < 1 {
		lap = int(s.GetLapsCompleted()) + 1
		if len(s.GetLaps()) > 0 {
			lap = 0
			for _, l := range s.GetLaps() {
				lap = max(lap, int(l.GetLap()))
			}
		}
	}
	result := fmt.Sprintf("Engineer stopped on lap %d", lap)
	if total := int(s.GetTotalLaps()); total > 0 {
		result += fmt.Sprintf(" of %d", total)
	}
	return withPosition(result, position)
}

func withPosition(result string, position int) string {
	if position > 0 {
		return fmt.Sprintf("%s, P%d", result, position)
	}
	return result
}

func plural(n int, word string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, word)
	}
	return fmt.Sprintf("%d %ss", n, word)
}
