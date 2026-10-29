package contracts_test

import (
	"bytes"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/those-who-run/contracts/callout"
	"github.com/those-who-run/contracts/ce"
	eventsv1 "github.com/those-who-run/contracts/gen/go/twr/events/v1"
)

// The callout paddock#71 adds.

// testdata/v1/suzuka_no_stop_callouts.json is what v1.0.0's engine queued for
// real-nuc-suzuka-no-stop-2026-10-15.jsonl.gz with its default rules (the
// replay as v1's tests/test_real_recording.py runs it): v1's Callout, the
// rule's pit_lane, and the frame it was raised on. The examples' clock puts
// the recording's first frame at 2026-10-15T19:00:00Z, to the microsecond, and
// a frame's seq is its packet_id less the first one's, plus 1. The examples
// carry the expires_at a live ecu would send (the queue's 8 s, or the race
// summary's 30 s): the replay that produced them ran v1's queue without a
// limit, and a replay's callouts carry none.
var suzukaFirstFrame = time.Date(2026, 10, 15, 19, 0, 0, 0, time.UTC)

// v1QueueMaxAge is v1's queue.max_age_seconds, a callout's life when its rule
// sets none.
const v1QueueMaxAge = 8 * time.Second

var pitLanes = map[string]eventsv1.PitLane{
	"speak": eventsv1.PitLane_PIT_LANE_SPEAK,
	"hold":  eventsv1.PitLane_PIT_LANE_HOLD,
	"drop":  eventsv1.PitLane_PIT_LANE_DROP,
}

// v1Callouts is testdata/v1/suzuka_no_stop_callouts.json.
type v1Callouts struct {
	Recording  string `json:"recording"`
	FirstFrame struct {
		T        float64 `json:"t"`
		PacketID float64 `json:"packet_id"`
	} `json:"first_frame"`
	Callouts []struct {
		Callout struct {
			Text      string   `json:"text"`
			Priority  int32    `json:"priority"`
			Rule      string   `json:"rule"`
			CreatedAt float64  `json:"created_at"`
			MaxAge    *float64 `json:"max_age"`
			Hold      float64  `json:"hold"`
			PitLane   string   `json:"pit_lane"`
		} `json:"callout"`
		Frame struct {
			T          float64 `json:"t"`
			PacketID   float64 `json:"packet_id"`
			CurrentLap int32   `json:"current_lap"`
			TotalLaps  int32   `json:"total_laps"`
			LastLapMs  int32   `json:"last_lap_ms"`
			BestLapMs  int32   `json:"best_lap_ms"`
			Position   int32   `json:"position"`
			TotalCars  int32   `json:"total_cars"`
		} `json:"frame"`
		// The position rule's place before the call (its `announced`).
		PreviousPosition *int32 `json:"previous_position"`
	} `json:"callouts"`
}

// Four of v1's callouts, as the examples carry them through ce: the start
// brief (p0), a position call held 1 s and dropped in the pit lane (p1), a
// new best (p1) and the race summary that lives 30 s (p2). Each has v1's rule,
// priority, pit_lane, created_at, max_age and hold, and the kind and slots
// v1's rule filled its phrase from (R6); radio, not ecu, phrases it, so none
// carries v1's text.
func TestCalloutsFromV1(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "v1", "suzuka_no_stop_callouts.json"))
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var v1 v1Callouts
	if err := dec.Decode(&v1); err != nil {
		t.Fatal(err)
	}
	if len(v1.Callouts) != 5 {
		t.Fatalf("%d v1 callouts, want 5", len(v1.Callouts))
	}
	first := v1.FirstFrame
	utc := func(at, plus float64) *timestamppb.Timestamp {
		return timestamppb.New(suzukaFirstFrame.Add(time.Duration(math.Round((at-first.T)*1e6)+math.Round(plus*1e6)) * time.Microsecond))
	}
	brief, position, lap2, lap3, summary := v1.Callouts[0], v1.Callouts[1], v1.Callouts[2], v1.Callouts[3], v1.Callouts[4]

	for _, tc := range []struct {
		example, kind, text string
		i                   int
		slots               *eventsv1.CalloutSlots
		level               callout.Level
	}{
		{"callout_start_brief.json", "template", "Six laps. P ten of twenty.", 0, &eventsv1.CalloutSlots{
			Laps: proto.Int32(brief.Frame.TotalLaps), Position: proto.Int32(brief.Frame.Position),
			TotalCars: proto.Int32(brief.Frame.TotalCars),
		}, callout.P0},
		{"callout_position.json", "template_lost", "Down to twelfth.", 1, &eventsv1.CalloutSlots{
			Position: proto.Int32(position.Frame.Position), PreviousPosition: position.PreviousPosition,
		}, callout.P1},
		// A new best's delta is the lap less GT7's best after the lap before.
		{"callout_new_best.json", "template_new_best", "Lap three. one oh three point one. New best by one point five.", 3, &eventsv1.CalloutSlots{
			Lap: proto.Int32(lap3.Frame.CurrentLap - 1), LapTimeMs: proto.Int32(lap3.Frame.LastLapMs),
			DeltaMs: proto.Int32(lap3.Frame.LastLapMs - lap2.Frame.BestLapMs),
		}, callout.P1},
		{"callout_race_summary.json", "template", "Chequered flag. Finished second. Best lap one oh one point two.", 4, &eventsv1.CalloutSlots{
			Position: proto.Int32(summary.Frame.Position), BestLapMs: proto.Int32(summary.Frame.BestLapMs),
		}, callout.P2},
	} {
		t.Run(tc.example, func(t *testing.T) {
			v := v1.Callouts[tc.i]
			c := v.Callout
			if c.Text != tc.text {
				t.Errorf("v1 said %q, want %q", c.Text, tc.text)
			}
			if v.Frame.T != c.CreatedAt {
				t.Errorf("v1 raised it at %v, on the frame at %v", c.CreatedAt, v.Frame.T)
			}
			maxAge := v1QueueMaxAge.Seconds()
			if c.MaxAge != nil {
				maxAge = *c.MaxAge
			}
			var ex eventsv1.Callout
			if err := ce.Unmarshal(readExample(t, "events/v1/"+tc.example), &ex); err != nil {
				t.Fatal(err)
			}
			want := &eventsv1.Callout{
				SessionId: "01JB5N2E8R4T6W9Y1A3C5E7G9H",
				Id:        ex.GetId(),
				Rule:      c.Rule,
				Kind:      tc.kind,
				Slots:     tc.slots,
				Priority:  c.Priority,
				At:        uint32(v.Frame.PacketID - first.PacketID + 1),
				RunId:     "01JB5N1Z7Q3S5V8X0B2D4F6H8K",
				CreatedAt: utc(c.CreatedAt, 0),
				ExpiresAt: utc(c.CreatedAt, maxAge),
				PitLane:   pitLanes[c.PitLane],
			}
			if c.Hold > 0 {
				want.HoldUntil = utc(c.CreatedAt, c.Hold)
			}
			got := throughCE(t, &ex, ex.GetId())
			if !proto.Equal(got, want) {
				t.Errorf("not v1's %q:\ngot  %v\nwant %v", c.Text, got, want)
			}
			if l := callout.LevelOf(got.GetPriority()); l != tc.level {
				t.Errorf("%s, want %s", l, tc.level)
			}
		})
	}
	if d := lap3.Frame.LastLapMs - lap2.Frame.BestLapMs; d >= 0 {
		t.Errorf("lap 3's delta is %d, want a new best", d)
	}
	if position.Callout.Hold != 1 || summary.Callout.MaxAge == nil || *summary.Callout.MaxAge != 30 {
		t.Error("v1's position hold is not 1 s, or its race summary's max_age 30 s")
	}
}

// AC3: a lap time callout with its slots (lap 3, 99000 ms, delta 1000 ms) and
// no text, and a telemetry callout with no session, both through ce, each with
// its id as the ce-id and Nats-Msg-Id (R4) and its level's subject.
func TestTheLapTimeAndTelemetryCallouts(t *testing.T) {
	for _, tc := range []struct {
		file, subject string
		check         func(*eventsv1.Callout) bool
	}{
		{"callout_lap_time.json", "twr.v1.live.callout.p2", func(c *eventsv1.Callout) bool {
			s := c.GetSlots()
			return c.GetSessionId() != "" && c.GetRule() == "lap_time" && c.GetKind() == "delta_slower" &&
				s.GetLap() == 3 && s.GetLapTimeMs() == 99000 && s.GetDeltaMs() == 1000
		}},
		{"callout_telemetry.json", "twr.v1.live.callout.p0", func(c *eventsv1.Callout) bool {
			return c.GetSessionId() == "" && c.GetRunId() != "" && c.GetRule() == "telemetry" &&
				c.GetKind() == "template_lost" && c.Slots == nil
		}},
	} {
		t.Run(tc.file, func(t *testing.T) {
			var c eventsv1.Callout
			if err := ce.Unmarshal(readExample(t, "events/v1/"+tc.file), &c); err != nil {
				t.Fatal(err)
			}
			h, payload, err := ce.Encode(&c, ce.Event{ID: c.GetId(), Source: "/twr/ecu/gt7", Time: c.GetCreatedAt().AsTime()})
			if err != nil {
				t.Fatal(err)
			}
			if h.Get(ce.IDHeader) != c.GetId() || h.Get(ce.MsgIDHeader) != c.GetId() || h.Get(ce.TypeHeader) != "run.thosewho.callout.raised.v1" {
				t.Errorf("headers %v", h)
			}
			var got eventsv1.Callout
			if _, err := ce.Decode(h, payload, &got); err != nil {
				t.Fatal(err)
			}
			if !proto.Equal(&got, &c) || !tc.check(&got) || got.GetText() != "" {
				t.Errorf("got %v", &got)
			}
			if subject, err := callout.Subject("live", &got); err != nil || subject != tc.subject {
				t.Errorf("subject %s (%v), want %s", subject, err, tc.subject)
			}
		})
	}
}

// The slots keep 0 apart from absent, as a delta equal to the best (0) is
// said and a first lap's (absent) is not; the times are absent when not set.
func TestCalloutFieldShapes(t *testing.T) {
	slots := (&eventsv1.CalloutSlots{}).ProtoReflect().Descriptor()
	for i := range slots.Fields().Len() {
		fd := slots.Fields().Get(i)
		if fd.Kind() == protoreflect.Int32Kind && !fd.HasPresence() {
			t.Errorf("CalloutSlots.%s is not optional", fd.Name())
		}
	}
	c := (&eventsv1.Callout{}).ProtoReflect().Descriptor()
	for _, name := range []protoreflect.Name{"created_at", "expires_at", "hold_until"} {
		if fd := c.Fields().ByName(name); fd == nil || fd.Message().FullName() != "google.protobuf.Timestamp" {
			t.Errorf("Callout.%s is not a Timestamp", name)
		}
	}
	roundTrip(t, &eventsv1.Callout{Slots: &eventsv1.CalloutSlots{DeltaMs: proto.Int32(0)}})
	var equal eventsv1.Callout
	if err := ce.Unmarshal([]byte(`{"slots":{"deltaMs":0}}`), &equal); err != nil || equal.GetSlots().DeltaMs == nil {
		t.Errorf("a delta of 0 read as absent: %v", err)
	}
}
