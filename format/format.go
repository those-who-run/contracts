// Package format is the one place a number becomes text, for every Go
// component that shows one: the results card in podium and the pit wall's
// API (paddock#128, ADR 0004 point 5 in those-who-run/those-who-run). It is
// v1's web/src/lib/format.ts, function for function, so the card and the page
// never show the same lap differently.
//
// Its cases are format-cases.json, v1 1.0.0's tests/fixtures/format-cases.json
// byte for byte. This package's tests check every case, and Cases hands the
// file to a consumer that checks its own formatter against the same version,
// such as the pit wall's format.ts.
//
// A value that is unknown is written as Missing ("—"), never 0, blank or
// "None". A function that has a no-value form of its own (LapTime and RaceTime
// at 0 or less, Position at 0) writes Missing for it, so a protobuf getter's
// zero for an absent field reads right:
//
//	best := format.RaceTime(int(s.GetBestLapTimeMs())) // "1:36.595", or "—" when absent
//
// Where zero is a real value (Gap, the fuel figures, SpeedKph), the caller
// writes Missing for an absent field itself.
package format

import (
	_ "embed" // for the cases
	"fmt"
	"math"
	"regexp"
	"strconv"
	"time"
)

//go:embed format-cases.json
var cases []byte

// Cases is format-cases.json, the cases every formatter of the project
// checks: an object with "about" and "cases", each case calling "fn" (v1's
// name, such as raceTime) with "args" and expecting "out". It is a copy, so a
// caller may change it.
func Cases() []byte {
	return append([]byte(nil), cases...)
}

// Missing is how an unknown value is written: an em dash.
const Missing = "—"

// minus is the true minus sign, U+2212: a hyphen is too short beside tabular
// figures.
const minus = "−"

// fuelOff is every fuel figure with fuel use off, when none means anything.
const fuelOff = "Fuel off"

// LapTime is a lap time from whole milliseconds, truncated: "1:27.764",
// "49.677". No time (0 or less) is Missing.
func LapTime(ms int) string {
	if ms <= 0 {
		return Missing
	}
	return clock(ms)
}

// RaceTime is a race time from whole milliseconds, truncated: "9:56.826",
// "1:02:03.456", and "49.677" under a minute. It is v1's results.race_time,
// which the card writes. No time (0 or less) is Missing.
func RaceTime(ms int) string {
	return LapTime(ms)
}

// Gap is a gap in milliseconds, always signed: "+0.866", "−0.312", "+0.000",
// "+1:01.234"; "Best" when best is set, for the best lap itself.
func Gap(ms int, best bool) string {
	if best {
		return "Best"
	}
	if ms < 0 {
		return minus + clock(-ms)
	}
	return "+" + clock(ms)
}

// Position is a position in two parts, so the field can be muted: "P3" and
// "/16". A position under 1 is Missing with an empty of, and a field size
// under 1 is an empty of. Whether a position is shown at all is the caller's.
func Position(p, cars int) (position, of string) {
	if p < 1 {
		return Missing, ""
	}
	if cars < 1 {
		return "P" + strconv.Itoa(p), ""
	}
	return "P" + strconv.Itoa(p), "/" + strconv.Itoa(cars)
}

// FuelPct is fuel left, in per cent of the tank, rounded half to even as v1's
// card rounds: "43%". Below 0 is Missing, and "Fuel off" when fuelUse is
// false.
func FuelPct(pct float64, fuelUse bool) string {
	switch {
	case !fuelUse:
		return fuelOff
	case !(pct >= 0) || math.IsInf(pct, 1):
		return Missing
	}
	return strconv.FormatFloat(roundHalfEven(pct, 0), 'f', -1, 64) + "%"
}

// FuelPerLap is fuel used in a lap, in per cent of the tank: "3.2%". A
// refuelled lap (below 0) is Missing, and "Fuel off" when fuelUse is false.
func FuelPerLap(pct float64, fuelUse bool) string {
	return tenths(pct, fuelUse, "%")
}

// FuelAmount is an amount of fuel in GT7's own units, for when the tank's
// size is unknown: "18.7 units". Below 0 (a refuel) is Missing, and "Fuel
// off" when fuelUse is false.
func FuelAmount(units float64, fuelUse bool) string {
	return tenths(units, fuelUse, " units")
}

func tenths(v float64, fuelUse bool, unit string) string {
	switch {
	case !fuelUse:
		return fuelOff
	case !(v >= 0) || math.IsInf(v, 1):
		return Missing
	}
	return strconv.FormatFloat(roundHalfEven(v, 1), 'f', 1, 64) + unit
}

// SpeedKph is a speed: "287 km/h". Below 0 is Missing.
func SpeedKph(kph float64) string {
	if !(kph >= 0) || math.IsInf(kph, 1) {
		return Missing
	}
	return strconv.FormatFloat(jsRound(kph), 'f', 0, 64) + " km/h"
}

// Whole is a whole number on a chart or its readout, grouped in thousands:
// "287", "6,450", "−12".
func Whole(v float64) string {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return Missing
	}
	n := jsRound(v)
	text := group(strconv.FormatFloat(math.Abs(n), 'f', 0, 64))
	if n < 0 {
		return minus + text
	}
	return text
}

// PedalPct is a pedal, 0 to 100: "44%". Below 0 is Missing.
func PedalPct(pct float64) string {
	if !(pct >= 0) || math.IsInf(pct, 1) {
		return Missing
	}
	return strconv.FormatFloat(jsRound(pct), 'f', 0, 64) + "%"
}

// Gear is a gear: "3", and "N" for GT7's 0 (neutral, or reverse). Below 0 is
// Missing.
func Gear(g int) string {
	switch {
	case g < 0:
		return Missing
	case g == 0:
		return "N"
	}
	return strconv.Itoa(g)
}

// TraceTime is a moment in a lap, in seconds from its start, to the tenth of
// a second: "48.3", "1:02.0". Below 0 is Missing.
func TraceTime(s float64) string {
	if !(s >= 0) || math.IsInf(s, 1) {
		return Missing
	}
	t := int(jsRound(s * 10))
	minutes, rest := t/600, t%600
	secs := fmt.Sprintf("%d.%d", rest/10, rest%10)
	if minutes == 0 {
		return secs
	}
	return fmt.Sprintf("%d:%04s", minutes, secs)
}

// Duration is a duration in milliseconds, to the second under an hour and to
// the minute from one: "42 s", "9 min 57 s", "1 h 04 min". Below 0 is
// Missing.
func Duration(ms int) string {
	if ms < 0 {
		return Missing
	}
	s := (ms + 500) / 1000
	switch {
	case s < 60:
		return fmt.Sprintf("%d s", s)
	case s < 3600:
		return fmt.Sprintf("%d min %02d s", s/60, s%60)
	}
	minutes := (s + 30) / 60
	return fmt.Sprintf("%d h %02d min", minutes/60, minutes%60)
}

// SessionDate is when a session started, seen from now in loc (UTC when
// nil): "Today 14:04", "Yesterday 21:10", "Thu 15 Oct 19:00", and the year
// when it isn't now's: "Fri 17 Oct 2025 14:04". A time at or before the Unix
// epoch, the zero time included, is Missing.
func SessionDate(at, now time.Time, loc *time.Location) string {
	if at.Unix() <= 0 {
		return Missing
	}
	at = at.In(location(loc))
	return dayName(at, now.In(location(loc))) + at.Format(" 15:04")
}

// SessionDay is the day a session started, as the session list heads it,
// seen from now in loc (UTC when nil): "Today", "Yesterday", "Thu 15 Oct",
// "Fri 17 Oct 2025". A time at or before the Unix epoch is Missing.
func SessionDay(at, now time.Time, loc *time.Location) string {
	if at.Unix() <= 0 {
		return Missing
	}
	return dayName(at.In(location(loc)), now.In(location(loc)))
}

// ISODay is the calendar date of a time in loc (UTC when nil), as the pit
// wall's API takes a day: "2026-10-17". A time at or before the Unix epoch is
// Missing.
func ISODay(at time.Time, loc *time.Location) string {
	if at.Unix() <= 0 {
		return Missing
	}
	return at.In(location(loc)).Format(time.DateOnly)
}

var isoDay = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})$`)

// ShortDay is a YYYY-MM-DD day as a filter chip says it: "17 Oct", or
// "17 Oct 2025" when it isn't in thisYear (0 never shows the year). Anything
// else is Missing.
func ShortDay(iso string, thisYear int) string {
	m := isoDay.FindStringSubmatch(iso)
	if m == nil {
		return Missing
	}
	y, _ := strconv.Atoi(m[1])
	mo, _ := strconv.Atoi(m[2])
	d, _ := strconv.Atoi(m[3])
	if mo < 1 || mo > 12 {
		return Missing
	}
	text := fmt.Sprintf("%d %s", d, time.Month(mo).String()[:3])
	if thisYear == 0 || thisYear == y {
		return text
	}
	return fmt.Sprintf("%s %d", text, y)
}

// clock is v1's results.race_time, line for line, for ms of 0 or more.
func clock(ms int) string {
	millis := ms % 1000
	seconds := ms / 1000
	minutes := seconds / 60
	seconds %= 60
	hours := minutes / 60
	minutes %= 60
	switch {
	case hours > 0:
		return fmt.Sprintf("%d:%02d:%02d.%03d", hours, minutes, seconds, millis)
	case minutes == 0:
		return fmt.Sprintf("%d.%03d", seconds, millis)
	}
	return fmt.Sprintf("%d:%02d.%03d", minutes, seconds, millis)
}

// roundHalfEven is Python's round(), halves to even, as format.ts writes it,
// so a figure the card also writes rounds the same.
func roundHalfEven(v float64, digits int) float64 {
	scale := math.Pow(10, float64(digits))
	return math.RoundToEven(v*scale) / scale
}

// jsRound is JavaScript's Math.round: halves up, towards +∞.
func jsRound(v float64) float64 {
	return math.Floor(v + 0.5)
}

// group puts a comma between each three digits of a whole number, as en-GB
// does: "6450" -> "6,450".
func group(digits string) string {
	out := make([]byte, 0, len(digits)+len(digits)/3)
	for i := range len(digits) {
		if i > 0 && (len(digits)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, digits[i])
	}
	return string(out)
}

func location(loc *time.Location) *time.Location {
	if loc == nil {
		return time.UTC
	}
	return loc
}

// dayName is the day at is, seen from today: "Today", "Yesterday",
// "Thu 15 Oct", "Fri 17 Oct 2025".
func dayName(at, today time.Time) string {
	switch dayNumber(today) - dayNumber(at) {
	case 0:
		return "Today"
	case 1:
		return "Yesterday"
	}
	if at.Year() == today.Year() {
		return at.Format("Mon 2 Jan")
	}
	return at.Format("Mon 2 Jan 2006")
}

// dayNumber is the days since 1970-01-01 of t's calendar date, so two dates
// are whole days apart, DST or not.
func dayNumber(t time.Time) int64 {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC).Unix() / 86400
}
