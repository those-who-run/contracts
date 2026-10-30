package contracts_test

import (
	"regexp"
	"strings"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	statev1 "github.com/those-who-run/contracts/gen/go/twr/state/v1"
)

// The KV values paddock#72 adds, and radio's (paddock#62).
var kvValues = []protoreflect.ProtoMessage{
	&statev1.EcuStatus{}, &statev1.LiveSession{}, &statev1.PitState{}, &statev1.FuelState{}, &statev1.OpenSession{},
	&statev1.RadioStatus{},
}

// secretish is a field name that would hold a token, a secret, an address or a
// Discord id.
var secretish = regexp.MustCompile(`token|secret|password|key|address|addr|ip|host|discord|channel|guild|user`)

// No value has a field for a token, a secret, a PS5 address or a Discord id
// (R4), at any depth: the same rule as v1's status route. The values of other
// packages a value holds, such as OpenSession's summary, are checked too.
func TestNoValueHasASecretField(t *testing.T) {
	seen := map[protoreflect.FullName]bool{}
	var walk func(m protoreflect.MessageDescriptor)
	walk = func(m protoreflect.MessageDescriptor) {
		if seen[m.FullName()] || strings.HasPrefix(string(m.FullName()), "google.protobuf.") {
			return
		}
		seen[m.FullName()] = true
		for i := range m.Fields().Len() {
			f := m.Fields().Get(i)
			for _, word := range strings.Split(string(f.Name()), "_") {
				if secretish.MatchString(word) && secretish.FindString(word) == word {
					t.Errorf("%s.%s may hold a secret, an address or a Discord id", m.FullName(), f.Name())
				}
			}
			if f.Message() != nil {
				walk(f.Message())
			}
		}
	}
	for _, v := range kvValues {
		walk(v.ProtoReflect().Descriptor())
	}
	if len(seen) < 10 {
		t.Errorf("walked %d messages, want the values and every part of them", len(seen))
	}
}

// Fields where 0 or false is a real value distinct from "absent" have
// presence, and the counts a restart carries on are plain (paddock#72).
func TestKVFieldShapes(t *testing.T) {
	for _, tc := range []struct {
		msg      protoreflect.ProtoMessage
		field    protoreflect.Name
		presence bool
	}{
		// 0 s since the last packet is a value; before the first, absent.
		{&statev1.EcuStatus{}, "last_packet_age_secs", true},
		{&statev1.EcuStatus{}, "idle_secs", false},
		{&statev1.LiveSession{}, "lap", true},
		{&statev1.LiveSession{}, "total_laps", true},
		{&statev1.LiveSession{}, "position", true},
		{&statev1.LiveSession{}, "total_cars", true},
		{&statev1.LiveTrack{}, "track_id", true},
		{&statev1.FuelState{}, "laps_remaining", true},
		{&statev1.OpenSession{}, "pauses", false},
		{&statev1.OpenSession{}, "entries", false},
		{&statev1.OpenSession{}, "pause", true},
		{&statev1.OpenSession{}, "pit_entry", true},
		{&statev1.OpenSession{}, "final_lap_started", false},
		// P1 is a place; before the first race.position, absent.
		{&statev1.OpenSession{}, "published_position", true},
		// false is voice down; without the discord sink, absent.
		{&statev1.RadioStatus{}, "voice_connected", true},
		// 0 listeners is the empty channel radio warns of; before the first
		// voice state, absent (paddock#61, R2).
		{&statev1.RadioStatus{}, "voice_listeners", true},
	} {
		f := tc.msg.ProtoReflect().Descriptor().Fields().ByName(tc.field)
		if f == nil {
			t.Errorf("%s has no %s", tc.msg.ProtoReflect().Descriptor().FullName(), tc.field)
			continue
		}
		if f.HasPresence() != tc.presence {
			t.Errorf("%s.%s: presence %v, want %v", tc.msg.ProtoReflect().Descriptor().FullName(), tc.field, f.HasPresence(), tc.presence)
		}
	}
}

// The comments say what a reader and ecu must know to use the values right.
func TestKVCommentsSayHowToReadThem(t *testing.T) {
	for _, tc := range []struct {
		file, name string
		says       []string
	}{
		{"twr/state/v1/live.proto", "EcuStatus", []string{"`ecu.status.<src>`", "never the live source's", "every second while idle", "at least every 5 s while busy", "updated_at", "`voice_connected` is radio's"}},
		{"twr/state/v1/live.proto", "idle_secs", []string{"as of updated_at", "0 while idle is false", "While idle is true, a reader adds the time since updated_at", "idle_secs is 0 whatever the time"}},
		{"twr/state/v1/live.proto", "ACTIVITY_IDLE", []string{"exactly when EcuStatus's idle is true", "left open by a feed that stopped is idle"}},
		{"twr/state/v1/live.proto", "ACTIVITY_REPLAY", []string{"GT7 is playing back a race", "Not a recording replayed through ecu"}},
		{"twr/state/v1/live.proto", "LiveSession", []string{"`ecu.session.<src>`", "deletes the key when the session ends"}},
		{"twr/state/v1/live.proto", "position", []string{"Absent where the results card hides it", "unless the session's SessionKind has field and cars_on_track both true (v1's `position_shown`)", "when the game reported none"}},
		{"twr/state/v1/live.proto", "total_cars", []string{"unless the session's SessionKind has field and cars_on_track both true (v1's `position_shown`)", "when the game reported none"}},
		{"twr/state/v1/live.proto", "in_pit_lane", []string{"PitTeleport"}},
		{"twr/state/v1/live.proto", "RadioStatus", []string{"`radio.status.<src>`", "never the live source's", "at most once a second", "at least every 5 s while it runs", "off the speaking path", "updated_at is far older than 5 s as radio gone"}},
		{"twr/state/v1/live.proto", "voice_listeners", []string{"not bots and not deafened", "0 is a channel with no one listening", "Absent before radio has seen the channel's members"}},
		{"twr/state/v1/open_session.proto", "OpenSession", []string{
			"`sessions.open.<src>`", "a replay never reads or clears the live source's", "a reader of `live` never sees the running summary",
			"revision check", "<pause> count is pauses", "<entry> count is entries", "<stop> count is",
		}},
		{"twr/state/v1/open_session.proto", "SessionSummary summary", []string{"the same session.ended", "laps holds only the laps a LapCompleted has carried", "ended_at and ended_paused"}},
		{"twr/state/v1/open_session.proto", "run_id", []string{"PitTeleport's entry_at"}},
		{"twr/state/v1/open_session.proto", "final_lap_started", []string{"once per session", "paddock#84, R1", "2 min duplicate window"}},
		{"twr/state/v1/open_session.proto", "published_position", []string{"last RacePositionChanged published", "which summary's final_position, the last frame's place, is not"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			comment := leadingComment(t, tc.file, tc.name)
			for _, want := range tc.says {
				if !strings.Contains(comment, want) {
					t.Errorf("%s's comment does not say %q:\n%s", tc.name, want, comment)
				}
			}
		})
	}
}
