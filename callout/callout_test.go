package callout_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/those-who-run/contracts/callout"
	eventsv1 "github.com/those-who-run/contracts/gen/go/twr/events/v1"
	"github.com/those-who-run/contracts/internal/catalogue"
)

// Each v1 rule's default priority (rules.example.yaml) has the level the
// triage settled: p0 under 10, p1 10 to 19, p2 20 and over, the start brief
// p0 (paddock#71, AC2).
func TestEachV1PriorityHasItsLevel(t *testing.T) {
	for _, tc := range []struct {
		what     string
		priority int32
		want     callout.Level
	}{
		{"telemetry, critical fuel and the pit call", 5, callout.P0},
		{"the start brief", 8, callout.P0},
		{"the final lap", 10, callout.P1},
		{"a new best", 11, callout.P1},
		{"position", 12, callout.P1},
		{"temps", 14, callout.P1},
		{"fuel", 15, callout.P1},
		{"a lap time", 20, callout.P2},
		{"the race summary", 25, callout.P2},
		// The edges, and what no v1 rule uses.
		{"the most urgent", 0, callout.P0},
		{"below zero", -1, callout.P0},
		{"just under p1", 9, callout.P0},
		{"just under p2", 19, callout.P1},
		{"v1's base default", 50, callout.P2},
	} {
		if got := callout.LevelOf(tc.priority); got != tc.want {
			t.Errorf("%s (%d) is %s, want %s", tc.what, tc.priority, got, tc.want)
		}
	}
}

// The subject is the catalogue's callout pattern for the level, with the
// source in place of <src>, and every level has its entry.
func TestSubjectsAreTheCatalogues(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "subjects.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	c, err := catalogue.Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	patterns := map[string]bool{}
	for _, s := range c.Subjects {
		if s.Message == "twr.events.v1.Callout" {
			patterns[s.Pattern] = true
		}
	}
	if len(patterns) != len(callout.Levels) {
		t.Errorf("the catalogue has %d callout subjects, the helper %d levels", len(patterns), len(callout.Levels))
	}
	for _, tc := range []struct {
		priority int32
		level    callout.Level
	}{{8, callout.P0}, {11, callout.P1}, {20, callout.P2}} {
		got, err := callout.Subject("replay-suzuka", &eventsv1.Callout{Priority: tc.priority})
		if err != nil {
			t.Fatal(err)
		}
		if want := "twr.v1.replay-suzuka.callout." + string(tc.level); got != want {
			t.Errorf("priority %d: %s, want %s", tc.priority, got, want)
		}
		if pattern := strings.Replace(got, "replay-suzuka", "<src>", 1); !patterns[pattern] {
			t.Errorf("%s is not in the catalogue", pattern)
		}
	}
}

// A source that is not one subject token would publish somewhere no one
// subscribes, or match more than it should.
func TestABadSourceFails(t *testing.T) {
	for _, src := range []string{"", "live.x", "*", ">", "my replay"} {
		if got, err := callout.Subject(src, &eventsv1.Callout{}); err == nil {
			t.Errorf("source %q gave %s", src, got)
		}
	}
}
