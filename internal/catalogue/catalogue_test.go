package catalogue_test

import (
	"bytes"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	// The generated messages, which register themselves in GlobalFiles.
	_ "github.com/those-who-run/contracts/gen/go/twr/events/v1"
	_ "github.com/those-who-run/contracts/gen/go/twr/state/v1"
	_ "github.com/those-who-run/contracts/gen/go/twr/telemetry/v1"
	"github.com/those-who-run/contracts/internal/catalogue"
)

var update = flag.Bool("update", false, "write docs/subjects.md from subjects.yaml")

// The repository's root and this package's fixtures, read through fs.FS.
var (
	root     = filepath.Join("..", "..")
	repo     = os.DirFS(root)
	testdata = os.DirFS("testdata")
)

const (
	subjects   = "subjects.yaml"
	subjectsMD = "docs/subjects.md"
)

func load(t *testing.T, fsys fs.FS, name string) *catalogue.Catalogue {
	t.Helper()
	f, err := fsys.Open(name)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	c, err := catalogue.Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

// The catalogue lists v0.1's four subjects (paddock#25, R1), the session and
// race events (paddock#69, R4), the 20 Hz frame (paddock#137) and the KV
// keys (paddock#72, R3), and passes every check.
func TestTheCatalogueChecksOut(t *testing.T) {
	c := load(t, repo, subjects)
	if err := c.Check(protoregistry.GlobalFiles); err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"twr.v1.<src>.frame.gt7":                 "twr.telemetry.v1.Frame",
		"twr.v1.<src>.evt.<sid>.session.started": "twr.events.v1.SessionStarted",
		"twr.v1.<src>.evt.<sid>.lap.completed":   "twr.events.v1.LapCompleted",
		"twr.v1.<src>.evt.<sid>.session.ended":   "twr.events.v1.SessionEnded",
		// paddock#69.
		"twr.v1.<src>.evt.<sid>.session.continued": "twr.events.v1.SessionContinued",
		"twr.v1.<src>.evt.<sid>.session.discarded": "twr.events.v1.SessionDiscarded",
		"twr.v1.<src>.evt.<sid>.session.kind":      "twr.events.v1.SessionKindChanged",
		"twr.v1.<src>.evt.<sid>.session.track":     "twr.events.v1.SessionTrackChanged",
		"twr.v1.<src>.evt.<sid>.session.paused":    "twr.events.v1.SessionPaused",
		"twr.v1.<src>.evt.<sid>.session.resumed":   "twr.events.v1.SessionResumed",
		"twr.v1.<src>.evt.<sid>.lap.best":          "twr.events.v1.LapBestChanged",
		"twr.v1.<src>.evt.<sid>.race.final_lap":    "twr.events.v1.RaceFinalLapStarted",
		"twr.v1.<src>.evt.<sid>.race.position":     "twr.events.v1.RacePositionChanged",
		"twr.v1.<src>.evt.<sid>.race.finished":     "twr.events.v1.RaceFinished",
		// paddock#70, with the telemetry events on evt._ as well.
		"twr.v1.<src>.evt.<sid>.pit.entered":        "twr.events.v1.PitEntered",
		"twr.v1.<src>.evt.<sid>.pit.exited":         "twr.events.v1.PitExited",
		"twr.v1.<src>.evt.<sid>.pit.teleport":       "twr.events.v1.PitTeleport",
		"twr.v1.<src>.evt.<sid>.car.alarm":          "twr.events.v1.CarAlarmRaised",
		"twr.v1.<src>.evt.<sid>.telemetry.lost":     "twr.events.v1.TelemetryLost",
		"twr.v1.<src>.evt._.telemetry.lost":         "twr.events.v1.TelemetryLost",
		"twr.v1.<src>.evt.<sid>.telemetry.restored": "twr.events.v1.TelemetryRestored",
		"twr.v1.<src>.evt._.telemetry.restored":     "twr.events.v1.TelemetryRestored",
		// paddock#71: one callout message, on three levels.
		"twr.v1.<src>.callout.p0": "twr.events.v1.Callout",
		"twr.v1.<src>.callout.p1": "twr.events.v1.Callout",
		"twr.v1.<src>.callout.p2": "twr.events.v1.Callout",
		// paddock#137: the 20 Hz frame, in no stream.
		"twr.v1.<src>.frame.gt7.lo": "twr.telemetry.v1.Frame",
	}
	// paddock#72: every key carries its source; the open session is in ecu's
	// own bucket, written and read by ecu alone. paddock#62: radio writes its
	// own status.
	wantKV := map[string]struct {
		message, producers, consumers string
	}{
		"live ecu.status.<src>":   {"twr.state.v1.EcuStatus", "ecu", "pitwall radio"},
		"live ecu.session.<src>":  {"twr.state.v1.LiveSession", "ecu", "pitwall"},
		"live ecu.pit.<src>":      {"twr.state.v1.PitState", "ecu", "pitwall"},
		"live ecu.fuel.<src>":     {"twr.state.v1.FuelState", "ecu", "pitwall"},
		"live radio.status.<src>": {"twr.state.v1.RadioStatus", "radio", "pitwall"},
		"ecu sessions.open.<src>": {"twr.state.v1.OpenSession", "ecu", "ecu"},
	}
	if len(c.Subjects) != len(want)+len(wantKV) {
		t.Errorf("%d subjects and keys, want %d", len(c.Subjects), len(want)+len(wantKV))
	}
	for _, s := range c.Subjects {
		if s.IsKV() {
			w := wantKV[s.KV+" "+s.Key]
			if w.message != s.Message || w.consumers != strings.Join(s.Consumers, " ") || w.producers != strings.Join(s.Producers, " ") {
				t.Errorf("kv %s %s holds %s from %v for %v, want %q from %q for %q", s.KV, s.Key, s.Message, s.Producers, s.Consumers, w.message, w.producers, w.consumers)
			}
			continue
		}
		if want[s.Pattern] != s.Message {
			t.Errorf("%s carries %s, want %q", s.Pattern, s.Message, want[s.Pattern])
		}
	}
}

// docs/subjects.md is what Render makes of subjects.yaml. `make generate` runs
// this test with -update to write it.
func TestSubjectsDocIsFresh(t *testing.T) {
	got := load(t, repo, subjects).Render()
	if *update {
		if err := os.WriteFile(filepath.Join(root, subjectsMD), got, 0o644); err != nil { //nolint:gosec // a doc in the repository, read by everyone
			t.Fatal(err)
		}
		return
	}
	want, err := fs.ReadFile(repo, subjectsMD)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("docs/subjects.md is not what subjects.yaml renders to: run `make generate` and commit the result")
	}
}

func assertFails(t *testing.T, c *catalogue.Catalogue, messages catalogue.Messages, want string) {
	t.Helper()
	err := c.Check(messages)
	if err == nil {
		t.Fatalf("the check passed, want an error with %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Errorf("got %v\nwant an error with %q", err, want)
	}
}

func assertPasses(t *testing.T, c *catalogue.Catalogue, messages catalogue.Messages) {
	t.Helper()
	if err := c.Check(messages); err != nil {
		t.Errorf("the check failed: %v", err)
	}
}

// withMessage is the generated messages plus a new message in a package, as a
// later .proto would add.
func withMessage(t *testing.T, pkg protoreflect.FullName, name string) *protoregistry.Files {
	t.Helper()
	files := new(protoregistry.Files)
	protoregistry.GlobalFiles.RangeFiles(func(f protoreflect.FileDescriptor) bool {
		if err := files.RegisterFile(f); err != nil {
			t.Fatal(err)
		}
		return true
	})
	fd, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Name:        proto.String(strings.ReplaceAll(string(pkg), ".", "/") + "/extra.proto"),
		Package:     proto.String(string(pkg)),
		Syntax:      proto.String("proto3"),
		MessageType: []*descriptorpb.DescriptorProto{{Name: proto.String(name)}},
	}, files)
	if err != nil {
		t.Fatal(err)
	}
	if err := files.RegisterFile(fd); err != nil {
		t.Fatal(err)
	}
	return files
}

// An event message without an entry fails the check: in a fixture catalogue
// that leaves one out, and when the generated code has an event the catalogue
// doesn't list, in any version of the events package.
func TestAnEventMessageWithoutAnEntryFails(t *testing.T) {
	assertFails(t, load(t, testdata, "missing-event.yaml"), protoregistry.GlobalFiles,
		"event message twr.events.v1.LapCompleted has no entry in the catalogue")
	assertFails(t, load(t, repo, subjects), withMessage(t, "twr.events.v1", "PenaltyServed"),
		"event message twr.events.v1.PenaltyServed has no entry in the catalogue")
	assertFails(t, load(t, repo, subjects), withMessage(t, "twr.events.v2", "LapCompleted"),
		"event message twr.events.v2.LapCompleted has no entry in the catalogue")
	// A message outside the events and state packages needs none.
	assertPasses(t, load(t, repo, subjects), withMessage(t, "twr.config.v1", "Settings"))
}

// A state message without a KV entry fails the check, in any version of the
// state package (paddock#72): every value's key is in the catalogue.
func TestAStateMessageWithoutAnEntryFails(t *testing.T) {
	assertFails(t, load(t, repo, subjects), withMessage(t, "twr.state.v1", "LogbookStatus"),
		"state message twr.state.v1.LogbookStatus has no KV entry in the catalogue, and is not a helper")
	assertFails(t, load(t, repo, subjects), withMessage(t, "twr.state.v2", "EcuStatus"),
		"state message twr.state.v2.EcuStatus has no KV entry in the catalogue, and is not a helper")
	c := load(t, repo, subjects)
	c.Subjects = slices.DeleteFunc(c.Subjects, func(s catalogue.Subject) bool { return s.Key == "sessions.open.<src>" })
	assertFails(t, c, protoregistry.GlobalFiles,
		"state message twr.state.v1.OpenSession has no KV entry in the catalogue")
	// A message only some subject carries is not a KV entry.
	c = load(t, repo, subjects)
	c.Subjects[0].Message = "twr.state.v1.EcuStatus"
	c.Subjects = slices.DeleteFunc(c.Subjects, func(s catalogue.Subject) bool { return s.Key == "ecu.status.<src>" })
	assertFails(t, c, protoregistry.GlobalFiles,
		"state message twr.state.v1.EcuStatus has no KV entry in the catalogue")
}

// helpers names the events-package messages that are not events: they need no
// entry, and only they may go without one.
func TestHelpers(t *testing.T) {
	c := load(t, repo, subjects)
	c.Helpers = nil
	assertFails(t, c, protoregistry.GlobalFiles,
		"event message twr.events.v1.SessionSummary has no entry in the catalogue, and is not a helper")
	assertFails(t, c, protoregistry.GlobalFiles,
		"state message twr.state.v1.OpenLap has no KV entry in the catalogue, and is not a helper")

	for _, tc := range []struct {
		helper string
		want   string
	}{
		{"twr.events.v1.LapCompleted", "helper twr.events.v1.LapCompleted has an entry: a helper is not an event"},
		{"twr.state.v1.EcuStatus", "helper twr.state.v1.EcuStatus has an entry: a helper is not an event or a value"},
		{"twr.telemetry.v1.Frame", "helper twr.telemetry.v1.Frame is not in an events package"},
		{"twr.events.v1.Summary", "helper twr.events.v1.Summary is not in the generated code"},
		{"twr.events.v1.PitRole", "helper twr.events.v1.PitRole is not a message"},
	} {
		t.Run(tc.helper, func(t *testing.T) {
			c := load(t, repo, subjects)
			c.Helpers = append(c.Helpers, tc.helper)
			assertFails(t, c, protoregistry.GlobalFiles, tc.want)
		})
	}
}

// A ce-type and a message are one to one (ADR 0002, point 4), though a message
// may travel on several subjects with its one ce-type.
func TestCETypesAndMessagesAreOneToOne(t *testing.T) {
	// Two messages with one ce-type, as #25's acceptance criterion has it.
	assertFails(t, load(t, testdata, "duplicate-ce-type.yaml"), protoregistry.GlobalFiles,
		"ce_type run.thosewho.session.started.v1 is twr.events.v1.SessionStarted's too")

	// One message with two ce-types.
	c := load(t, repo, subjects)
	other := c.Subjects[2]
	other.Pattern, other.CEType = "twr.v1.<src>.evt.<sid>.lap.done", "run.thosewho.lap.done.v1"
	c.Subjects = append(c.Subjects, other)
	assertFails(t, c, protoregistry.GlobalFiles,
		"twr.events.v1.LapCompleted has the ce_type run.thosewho.lap.completed.v1 elsewhere")

	// One message and ce-type on several subjects passes: the same event with no
	// session open (the _ token, paddock#70), and one entry per callout level.
	c = load(t, repo, subjects)
	// SessionKindChanged, which carries the run its <run> needs.
	noSession := c.Subjects[slices.IndexFunc(c.Subjects, func(s catalogue.Subject) bool { return s.Message == "twr.events.v1.SessionKindChanged" })]
	noSession.Pattern, noSession.MsgID = "twr.v1.<src>.evt._.session.kind", "<src>:<run>:session.kind:<seq>"
	c.Subjects = append(c.Subjects, noSession)
	// The real catalogue's callout, one message and ce-type on p0, p1 and p2.
	if n := slices.IndexFunc(c.Subjects, func(s catalogue.Subject) bool { return s.Pattern == "twr.v1.<src>.callout.p2" }); n < 0 ||
		c.Subjects[n].Message != "twr.events.v1.Callout" {
		t.Fatal("the catalogue has no callout.p2 carrying the Callout")
	}
	assertPasses(t, c, protoregistry.GlobalFiles)

	// The same subject twice does not.
	c = load(t, repo, subjects)
	c.Subjects = append(c.Subjects, c.Subjects[1])
	assertFails(t, c, protoregistry.GlobalFiles, "subjects[1] is the same subject or key")
}

// An event's msg_id has the one style (paddock#25, 2026-10-26; paddock#279):
// `<src>:<sid>:<area>.<what>:<part>`, with its ce-type's <area>.<what>.
func TestEventMsgIDs(t *testing.T) {
	for _, tc := range []struct {
		name    string
		pattern string
		msgID   string
		want    string
	}{
		{"the old colon form", "", "<src>:<sid>:lap:<n>:completed", `msg_id "<src>:<sid>:lap:<n>:completed" is not <src>:<sid>:<area>.<what>`},
		{"no source in front, as v0.1.4 had", "", "<sid>:lap.completed:<n>", `msg_id "<sid>:lap.completed:<n>" is not <src>:<sid>:<area>.<what>`},
		{"the source after the session", "", "<sid>:<src>:lap.completed:<n>", "is not <src>:<sid>:<area>.<what>"},
		{"no part after the type", "", "<src>:<sid>:lap.completed", "is not <src>:<sid>:<area>.<what>"},
		{"another event's area and what", "", "<src>:<sid>:lap.started:<n>", "names lap.started, but the ce_type is lap.completed's"},
		{"<run> on a session's subject", "", "<src>:<run>:lap.completed:<n>", "has <run> after the source, want <sid>"},
		{"<sid> with no session open", "twr.v1.<src>.evt._.lap.completed", "<src>:<sid>:lap.completed:<n>", "msg_id uses <sid>, which the pattern does not carry"},
		{"<sid> in its place with no session open", "twr.v1.<src>.evt._.lap.completed", "<src>:<n>:lap.completed:<n>", "is not <src>:<sid>:<area>.<what>"},
		{"no source with no session open", "twr.v1.<src>.evt._.lap.completed", "<run>:lap.completed:<n>", "is not <src>:<sid>:<area>.<what>"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := load(t, repo, subjects)
			if tc.pattern != "" {
				c.Subjects[2].Pattern = tc.pattern
			}
			c.Subjects[2].MsgID = tc.msgID
			assertFails(t, c, protoregistry.GlobalFiles, tc.want)
		})
	}
	// An event whose id has <run> carries the run in its message (paddock#70),
	// so an id is built from its message: LapCompleted has no run_id.
	c0 := load(t, repo, subjects)
	c0.Subjects[2].MsgID = "<src>:<sid>:lap.completed:<run>:<seq>"
	assertFails(t, c0, protoregistry.GlobalFiles,
		`msg_id "<src>:<sid>:lap.completed:<run>:<seq>" has <run>, but twr.events.v1.LapCompleted has no string run_id`)
	// More than one part is fine, as #70's car alarms have.
	c := load(t, repo, subjects)
	c.Subjects[2].MsgID = "<src>:<sid>:lap.completed:<n>:1"
	assertPasses(t, c, protoregistry.GlobalFiles)
}

// A callout's msg_id is its own id (paddock#71, R4), which the message
// carries, so the id is built from the message; anything else fails.
func TestCalloutMsgIDs(t *testing.T) {
	c := load(t, repo, subjects)
	callouts := 0
	for i, s := range c.Subjects {
		if !strings.Contains(s.Pattern, ".callout.") {
			continue
		}
		callouts++
		if s.MsgID != catalogue.CalloutID || s.Stream != "CALLOUTS" || s.CEType != "run.thosewho.callout.raised.v1" {
			t.Errorf("%s: msg_id %s, stream %s, ce_type %s", s.Pattern, s.MsgID, s.Stream, s.CEType)
		}
		for _, msgID := range []string{"<src>:<id>", "<run>:<seq>", "<src>:<sid>:callout.raised:<id>"} {
			t.Run(s.Pattern+" "+msgID, func(t *testing.T) {
				c := load(t, repo, subjects)
				c.Subjects[i].MsgID = msgID
				assertFails(t, c, protoregistry.GlobalFiles, "is not <id>: a callout's Nats-Msg-Id is its own id")
			})
		}
	}
	if callouts != 3 {
		t.Errorf("%d callout subjects, want p0, p1 and p2", callouts)
	}
	// <id> is built from the message's string id: a frame has none.
	c.Subjects[0].MsgID = "<id>"
	assertFails(t, c, protoregistry.GlobalFiles,
		`msg_id "<id>" has <id>, but twr.telemetry.v1.Frame has no string id to build it from`)
}

// A token of an event's <area> or <what>, and of its ce-type, is lower-case
// words joined by single underscores, as race.final_lap is (paddock#69): an
// underscore at either end, two together, or a capital still fail.
func TestTokensMayJoinWordsWithUnderscores(t *testing.T) {
	c := load(t, repo, subjects)
	assertPasses(t, c, protoregistry.GlobalFiles)
	finalLap := slices.IndexFunc(c.Subjects, func(s catalogue.Subject) bool {
		return s.Pattern == "twr.v1.<src>.evt.<sid>.race.final_lap"
	})
	if finalLap < 0 {
		t.Fatal("the catalogue has no race.final_lap")
	}
	for _, what := range []string{"_final", "final_", "final__lap", "Final_lap"} {
		t.Run(what, func(t *testing.T) {
			c := load(t, repo, subjects)
			s := &c.Subjects[finalLap]
			s.Pattern = "twr.v1.<src>.evt.<sid>.race." + what
			s.CEType = "run.thosewho.race." + what + ".v1"
			s.MsgID = "<src>:<sid>:race." + what + ":1"
			assertFails(t, c, protoregistry.GlobalFiles, "does not follow ADR 0001's subject scheme")
			assertFails(t, c, protoregistry.GlobalFiles, "is not run.thosewho.<area>.<what>.v<N>")
		})
	}
	// The msg_id's <area>.<what> follows the same rule.
	c = load(t, repo, subjects)
	c.Subjects[finalLap].MsgID = "<src>:<sid>:race.final__lap:1"
	assertFails(t, c, protoregistry.GlobalFiles, "is not <src>:<sid>:<area>.<what>")
}

// kvEntry is the index of the real catalogue's KV entry for ecu.status.<src>
// (paddock#72), checked beside every event the catalogue has.
func kvEntry(t *testing.T, c *catalogue.Catalogue) int {
	t.Helper()
	n := slices.IndexFunc(c.Subjects, func(s catalogue.Subject) bool { return s.KV == "live" && s.Key == "ecu.status.<src>" })
	if n < 0 {
		t.Fatal("the catalogue has no live ecu.status.<src>")
	}
	return n
}

// A KV entry has its own rules: a bucket, a key, a message, writers, readers and
// a description, and no subject, stream, ce-type or msg_id.
func TestKVEntries(t *testing.T) {
	messages := protoregistry.GlobalFiles
	c := load(t, repo, subjects)
	assertPasses(t, c, messages)
	doc := string(c.Render())
	for _, row := range []string{
		"| `live` | `ecu.status.<src>` | `twr.state.v1.EcuStatus` | ecu | pitwall, radio |",
		"| `ecu` | `sessions.open.<src>` | `twr.state.v1.OpenSession` | ecu | ecu |",
	} {
		if !strings.Contains(doc, row) {
			t.Errorf("the KV key is not in the rendered doc: %s\n%s", row, doc)
		}
	}

	for _, tc := range []struct {
		name   string
		change func(*catalogue.Subject)
		want   string
	}{
		{"a ce_type", func(s *catalogue.Subject) { s.CEType = "run.thosewho.ecu.status.v1" }, "a KV entry has no ce_type"},
		{"a msg_id", func(s *catalogue.Subject) { s.MsgID = "<src>" }, "a KV entry has no msg_id"},
		{"a stream", func(s *catalogue.Subject) { s.Stream = "EVENTS" }, "a KV entry has no stream"},
		{"a pattern", func(s *catalogue.Subject) { s.Pattern = "twr.v1.<src>.ecu.status" }, "a KV entry has no pattern"},
		{"a bucket with capitals", func(s *catalogue.Subject) { s.KV = "Live" }, `kv bucket "Live" is not lower-case`},
		{"a key with a space", func(s *catalogue.Subject) { s.Key = "ecu status" }, `key "ecu status" is not dot-separated`},
		{"a placeholder not allowed in a key", func(s *catalogue.Subject) { s.Key = "ecu.frame.<seq>" }, "unknown placeholder <seq> in the key"},
		{"a key without its source", func(s *catalogue.Subject) { s.Key = "ecu.status" }, `key "ecu.status" does not carry <src>: every key in the live bucket`},
		{"no writers", func(s *catalogue.Subject) { s.Producers = nil }, "no producers"},
		{"no readers", func(s *catalogue.Subject) { s.Consumers = nil }, "no consumers"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := load(t, repo, subjects)
			tc.change(&c.Subjects[kvEntry(t, c)])
			assertFails(t, c, messages, tc.want)
		})
	}

	// The open session's bucket is one per source too.
	c = load(t, repo, subjects)
	open := slices.IndexFunc(c.Subjects, func(s catalogue.Subject) bool { return s.KV == "ecu" })
	c.Subjects[open].Key = "sessions.open"
	assertFails(t, c, messages, `key "sessions.open" does not carry <src>: every key in the ecu bucket`)

	// A key in another bucket, such as the planned config bucket's, need not
	// carry <src>.
	c = load(t, repo, subjects)
	n := kvEntry(t, c)
	c.Subjects[n].KV, c.Subjects[n].Key = "config", "ecu.ps5.address"
	assertPasses(t, c, messages)

	c = load(t, repo, subjects)
	c.Subjects[0].Key = "frames"
	assertFails(t, c, protoregistry.GlobalFiles, "a key without a kv bucket")
}

// Each of the other checks, on the real catalogue with one entry broken.
func TestEachCheck(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*catalogue.Subject)
		want   string
	}{
		{"no pattern", func(s *catalogue.Subject) { s.Pattern = "" }, "no pattern"},
		{"a pattern outside twr.v1.", func(s *catalogue.Subject) { s.Pattern = "twr.v2.<src>.frame.gt7" }, "does not start with twr.v1."},
		{"an unknown placeholder", func(s *catalogue.Subject) { s.Pattern = "twr.v1.<source>.frame.gt7" }, "unknown placeholder <source> in the pattern"},
		{"a msg_id placeholder in the pattern", func(s *catalogue.Subject) { s.Pattern = "twr.v1.<src>.frame.<run>" }, "unknown placeholder <run> in the pattern"},
		{"a pattern outside ADR 0001's scheme", func(s *catalogue.Subject) { s.Pattern = "twr.v1.<src>.frames.gt7" }, "does not follow ADR 0001's subject scheme"},
		{"a callout level outside p0 to p2", func(s *catalogue.Subject) { s.Pattern, s.Stream = "twr.v1.<src>.callout.p3", "CALLOUTS" }, "does not follow ADR 0001's subject scheme"},
		{"a frame in the wrong stream", func(s *catalogue.Subject) { s.Stream = "EVENTS" }, "a frame subject is captured by the TELEMETRY stream"},
		{"a frame in no stream", func(s *catalogue.Subject) { s.Stream = "" }, `a frame subject is captured by the TELEMETRY stream, not ""`},
		{"a 20 Hz frame in a stream", func(s *catalogue.Subject) { s.Pattern = "twr.v1.<src>.frame.gt7.lo" }, `a 20 Hz frame subject is captured by no stream, not "TELEMETRY"`},
		{"a ce_type without a version", func(s *catalogue.Subject) { s.CEType = "run.thosewho.frame.gt7" }, "is not run.thosewho.<area>.<what>.v<N>"},
		{"another content type", func(s *catalogue.Subject) { s.ContentType = "application/protobuf" }, `content_type "application/protobuf" is not application/json`},
		{"no message", func(s *catalogue.Subject) { s.Message = "" }, "no message"},
		{"a message not generated", func(s *catalogue.Subject) { s.Message = "twr.telemetry.v1.Frames" }, "message twr.telemetry.v1.Frames is not in the generated code"},
		{"an enum named as the message", func(s *catalogue.Subject) { s.Message = "twr.telemetry.v1.Phase" }, "twr.telemetry.v1.Phase is not a message"},
		{"no msg_id", func(s *catalogue.Subject) { s.MsgID = "" }, "no msg_id"},
		{"an unknown placeholder in msg_id", func(s *catalogue.Subject) { s.MsgID = "<run>:frame:<n2>" }, "unknown placeholder <n2> in msg_id"},
		{"<sid> in msg_id without it in the pattern", func(s *catalogue.Subject) { s.MsgID = "<sid>:frame:<seq>" }, "msg_id uses <sid>, which the pattern does not carry"},
		{"no producers", func(s *catalogue.Subject) { s.Producers = nil }, "no producers"},
		{"no consumers", func(s *catalogue.Subject) { s.Consumers = nil }, "no consumers"},
		{"no description", func(s *catalogue.Subject) { s.Description = "" }, "no description"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := load(t, repo, subjects)
			tc.change(&c.Subjects[0])
			assertFails(t, c, protoregistry.GlobalFiles, tc.want)
		})
	}
}

// The 20 Hz frame (paddock#137) is the frame on its own subject: the same
// message, ce-type, producer and msg_id, in no stream (ADR 0001, point 1).
func TestTheLoFrameIsTheFrame(t *testing.T) {
	c := load(t, repo, subjects)
	by := map[string]catalogue.Subject{}
	for _, s := range c.Subjects {
		by[s.Pattern] = s
	}
	frame, lo := by["twr.v1.<src>.frame.gt7"], by["twr.v1.<src>.frame.gt7.lo"]
	if lo.Message != frame.Message || lo.CEType != frame.CEType || lo.MsgID != frame.MsgID ||
		!slices.Equal(lo.Producers, frame.Producers) {
		t.Errorf(".lo is %+v, want the frame's message, ce_type, msg_id and producers: %+v", lo, frame)
	}
	if lo.Stream != "" {
		t.Errorf(".lo is captured by %s, want no stream", lo.Stream)
	}
	if !slices.Contains(lo.Consumers, "dash") {
		t.Errorf(".lo's consumers are %v, want the dash", lo.Consumers)
	}
}

// A key subjects.yaml doesn't define, such as a misspelling or the old
// singular producer, fails to parse rather than being dropped.
func TestAnUnknownKeyFails(t *testing.T) {
	for _, key := range []string{"stram", "producer"} {
		_, err := catalogue.Parse(strings.NewReader("subjects:\n  - pattern: twr.v1.<src>.frame.gt7\n    " + key + ": x\n"))
		if err == nil || !strings.Contains(err.Error(), "field "+key+" not found") {
			t.Errorf("%s: got %v, want an error about the unknown key", key, err)
		}
	}
}

// The event messages are every events-package message but the helpers.
func TestEventMessages(t *testing.T) {
	got := load(t, repo, subjects).EventMessages(protoregistry.GlobalFiles)
	want := "twr.events.v1.Callout twr.events.v1.CarAlarmRaised twr.events.v1.LapBestChanged twr.events.v1.LapCompleted " +
		"twr.events.v1.PitEntered twr.events.v1.PitExited twr.events.v1.PitTeleport " +
		"twr.events.v1.RaceFinalLapStarted " +
		"twr.events.v1.RaceFinished twr.events.v1.RacePositionChanged twr.events.v1.SessionContinued " +
		"twr.events.v1.SessionDiscarded twr.events.v1.SessionEnded twr.events.v1.SessionKindChanged " +
		"twr.events.v1.SessionPaused twr.events.v1.SessionResumed twr.events.v1.SessionStarted " +
		"twr.events.v1.SessionTrackChanged twr.events.v1.TelemetryLost twr.events.v1.TelemetryRestored"
	var names []string
	for _, n := range got {
		names = append(names, string(n))
	}
	if strings.Join(names, " ") != want {
		t.Errorf("got %v, want %s", got, want)
	}
}

// The doc explains every placeholder the check knows, from the same list.
func TestEveryPlaceholderIsExplained(t *testing.T) {
	doc := string(load(t, repo, subjects).Render())
	for _, p := range catalogue.Placeholders {
		if !strings.Contains(doc, "| `"+p.Token+"` | "+p.Meaning+" |") {
			t.Errorf("%s is not explained in the doc", p.Token)
		}
	}
}

// Agrees reports a ce-type that differs, and a message only one side has.
func TestAgrees(t *testing.T) {
	c := load(t, repo, subjects)
	types := ceTypes(c)
	if err := c.Agrees(types); err != nil {
		t.Fatalf("the catalogue disagrees with its own ce-types: %v", err)
	}
	types["twr.events.v1.LapCompleted"] = "run.thosewho.lap.completed.v2"
	delete(types, "twr.events.v1.SessionEnded")
	types["twr.events.v1.PenaltyServed"] = "run.thosewho.penalty.served.v1"
	err := c.Agrees(types)
	for _, want := range []string{
		"twr.events.v1.LapCompleted: the catalogue's ce-type is run.thosewho.lap.completed.v1, the helper's run.thosewho.lap.completed.v2",
		"twr.events.v1.SessionEnded (run.thosewho.session.ended.v1) has no ce-type in the helper",
		"twr.events.v1.PenaltyServed (run.thosewho.penalty.served.v1) is in the helper but not the catalogue",
	} {
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("got %v, want an error with %q", err, want)
		}
	}
}

// ceTypes is the catalogue's own ce-type table, as a helper would hold it: one
// per message, and none for a KV entry, which has no ce-type.
func ceTypes(c *catalogue.Catalogue) map[protoreflect.FullName]string {
	types := map[protoreflect.FullName]string{}
	for _, s := range c.Subjects {
		if !s.IsKV() {
			types[protoreflect.FullName(s.Message)] = s.CEType
		}
	}
	return types
}

// Agrees passes over KV entries, which have no ce-type, and takes a message on
// several subjects once.
func TestAgreesSkipsKVAndRepeats(t *testing.T) {
	c := load(t, repo, subjects)
	types := ceTypes(c)
	if _, ok := types["twr.state.v1.EcuStatus"]; ok {
		t.Fatal("the KV entry's message is in the ce-type table")
	}
	if err := c.Agrees(types); err != nil {
		t.Errorf("a catalogue with a KV entry disagrees with its own ce-types: %v", err)
	}

	// LapCompleted also on evt._, with the same ce-type: still one message.
	again := c.Subjects[2]
	again.Pattern = "twr.v1.<src>.evt._.lap.completed"
	c.Subjects = append(c.Subjects, again)
	if err := c.Agrees(types); err != nil {
		t.Errorf("a message on two subjects disagrees: %v", err)
	}
	delete(types, "twr.events.v1.LapCompleted")
	err := c.Agrees(types)
	if err == nil || strings.Count(err.Error(), "twr.events.v1.LapCompleted") != 1 {
		t.Errorf("got %v, want LapCompleted reported once", err)
	}
}
