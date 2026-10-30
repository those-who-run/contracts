package format_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"testing"
	"time"
	_ "time/tzdata" // Europe/London on a host without a zoneinfo database

	"github.com/those-who-run/contracts/format"
)

// v1Digest is the SHA-256 of tests/fixtures/format-cases.json in
// those-who-run/those-who-run at v1.0.0 (paddock#128 AC1).
const v1Digest = "079f29ba9bc01d7f1163393a7a0c0e4d368ebd1060d6add608bfc4aa43d75821"

type formatCase struct {
	Fn   string `json:"fn"`
	Args []any  `json:"args"`
	Out  any    `json:"out"`
}

func readCases(t *testing.T) []formatCase {
	t.Helper()
	var file struct {
		Cases []formatCase `json:"cases"`
	}
	if err := json.Unmarshal(format.Cases(), &file); err != nil {
		t.Fatalf("format-cases.json: %v", err)
	}
	if len(file.Cases) == 0 {
		t.Fatal("format-cases.json has no cases")
	}
	return file.Cases
}

// arg is args[i] as a number, and false when it is null or absent: the
// caller's unknown value, which it writes as format.Missing.
func arg(args []any, i int) (float64, bool) {
	if i >= len(args) || args[i] == nil {
		return 0, false
	}
	return args[i].(float64), true
}

// opt is the boolean option name of args[i], such as {"best": true}, and def
// when it is absent.
func opt(args []any, i int, name string, def bool) bool {
	if i >= len(args) || args[i] == nil {
		return def
	}
	if v, ok := args[i].(map[string]any)[name].(bool); ok {
		return v
	}
	return def
}

func zone(t *testing.T, args []any, i int) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(args[i].(string))
	if err != nil {
		t.Fatal(err)
	}
	return loc
}

func unix(args []any, i int) (time.Time, bool) {
	v, ok := arg(args, i)
	return time.Unix(int64(v), 0), ok
}

// call is what the Go formatter named for v1's fn writes for args. A null
// value is the caller's to write as Missing, as the package says.
func call(t *testing.T, c formatCase) any {
	t.Helper()
	a := c.Args
	switch c.Fn {
	case "missing":
		return format.Missing
	case "shortDay":
		year, _ := arg(a, 1)
		return format.ShortDay(a[0].(string), int(year))
	}
	v, known := arg(a, 0)
	if !known {
		if c.Fn == "position" {
			return map[string]any{"position": format.Missing, "of": ""}
		}
		return format.Missing
	}
	switch c.Fn {
	case "lapTime":
		return format.LapTime(int(v))
	case "raceTime":
		return format.RaceTime(int(v))
	case "gap":
		return format.Gap(int(v), opt(a, 1, "best", false))
	case "position":
		cars, _ := arg(a, 1)
		p, of := format.Position(int(v), int(cars))
		return map[string]any{"position": p, "of": of}
	case "fuelPct":
		return format.FuelPct(v, opt(a, 1, "fuelUse", true))
	case "fuelPerLap":
		return format.FuelPerLap(v, opt(a, 1, "fuelUse", true))
	case "fuelAmount":
		return format.FuelAmount(v, opt(a, 1, "fuelUse", true))
	case "speedKph":
		return format.SpeedKph(v)
	case "whole":
		return format.Whole(v)
	case "pedalPct":
		return format.PedalPct(v)
	case "gear":
		return format.Gear(int(v))
	case "traceTime":
		return format.TraceTime(v)
	case "duration":
		return format.Duration(int(v))
	case "sessionDate", "sessionDay":
		at, _ := unix(a, 0)
		now, _ := unix(a, 1)
		if c.Fn == "sessionDay" {
			return format.SessionDay(at, now, zone(t, a, 2))
		}
		return format.SessionDate(at, now, zone(t, a, 2))
	case "isoDay":
		at, _ := unix(a, 0)
		return format.ISODay(at, zone(t, a, 1))
	}
	t.Fatalf("no Go formatter for %q: add one, and its case here", c.Fn)
	return nil
}

func TestEveryCasePasses(t *testing.T) {
	for _, c := range readCases(t) {
		t.Run(fmt.Sprintf("%s%v", c.Fn, c.Args), func(t *testing.T) {
			if got := call(t, c); !reflect.DeepEqual(got, c.Out) {
				t.Errorf("%s(%v) = %q, want %q", c.Fn, c.Args, got, c.Out)
			}
		})
	}
}

func TestCasesAreV1sFileByteForByte(t *testing.T) {
	sum := sha256.Sum256(format.Cases())
	if got := hex.EncodeToString(sum[:]); got != v1Digest {
		t.Errorf("format-cases.json's SHA-256 is %s, want v1.0.0's %s", got, v1Digest)
	}
}

func TestCasesIsACopy(t *testing.T) {
	a := format.Cases()
	a[0] = 'x'
	if b := format.Cases(); bytes.Equal(a, b) {
		t.Error("a change to what Cases returned reached the next call")
	}
}

// The card's three shapes of a time, as v1's test_the_cases_cover_the_card_times.
func TestTheCasesCoverTheCardTimes(t *testing.T) {
	want := map[string]bool{"49.677": false, "1:36.595": false, "1:02:03.456": false}
	for _, c := range readCases(t) {
		if out, ok := c.Out.(string); ok && (c.Fn == "lapTime" || c.Fn == "raceTime") {
			if _, card := want[out]; card {
				want[out] = true
			}
		}
	}
	for out, seen := range want {
		if !seen {
			t.Errorf("no case writes %s", out)
		}
	}
}

// Values the cases don't reach: never "NaN" or "+Inf", and nothing out of
// range.
func TestOutOfRange(t *testing.T) {
	nan, inf := math.NaN(), math.Inf(1)
	for name, got := range map[string]string{
		"FuelPct(NaN)":        format.FuelPct(nan, true),
		"FuelPct(+Inf)":       format.FuelPct(inf, true),
		"FuelPerLap(NaN)":     format.FuelPerLap(nan, true),
		"FuelAmount(+Inf)":    format.FuelAmount(inf, true),
		"SpeedKph(-1)":        format.SpeedKph(-1),
		"SpeedKph(NaN)":       format.SpeedKph(nan),
		"Whole(NaN)":          format.Whole(nan),
		"Whole(-Inf)":         format.Whole(math.Inf(-1)),
		"PedalPct(-1)":        format.PedalPct(-1),
		"Gear(-1)":            format.Gear(-1),
		"TraceTime(-0.1)":     format.TraceTime(-0.1),
		"TraceTime(+Inf)":     format.TraceTime(inf),
		"Duration(-1)":        format.Duration(-1),
		"ISODay(zero)":        format.ISODay(time.Time{}, nil),
		"SessionDay(zero)":    format.SessionDay(time.Time{}, time.Now(), nil),
		"SessionDate(zero)":   format.SessionDate(time.Time{}, time.Now(), nil),
		"ShortDay(2026-13-1)": format.ShortDay("2026-13-01", 2026),
	} {
		if got != format.Missing {
			t.Errorf("%s = %q, want %q", name, got, format.Missing)
		}
	}
	if got := format.Whole(1234567.4); got != "1,234,567" {
		t.Errorf("Whole(1234567.4) = %q", got)
	}
	if got := format.ShortDay("2025-10-17", 0); got != "17 Oct" {
		t.Errorf("ShortDay without a year = %q", got)
	}
	at := time.Date(2026, 10, 17, 22, 0, 0, 0, time.UTC)
	if got := format.SessionDate(at, at, nil); got != "Today 22:00" {
		t.Errorf("SessionDate in UTC by default = %q", got)
	}
}
