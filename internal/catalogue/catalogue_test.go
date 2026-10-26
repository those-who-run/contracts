package catalogue_test

import (
	"bytes"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"

	// The generated messages, which register themselves in GlobalFiles.
	_ "github.com/those-who-run/contracts/gen/go/twr/events/v1"
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

// The catalogue lists v0.1's four subjects (paddock#25, R1) and passes every check.
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
	}
	if len(c.Subjects) != len(want) {
		t.Errorf("%d subjects, want %d", len(c.Subjects), len(want))
	}
	for _, s := range c.Subjects {
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

// An event message without an entry fails the check: in a fixture catalogue
// that leaves one out, and when the generated code has an event the catalogue
// doesn't list.
func TestAnEventMessageWithoutAnEntryFails(t *testing.T) {
	assertFails(t, load(t, testdata, "missing-event.yaml"), protoregistry.GlobalFiles,
		"event message twr.events.v1.LapCompleted has no entry in the catalogue")

	assertFails(t, load(t, repo, subjects), withEvent(t, "PitEntered"),
		"event message twr.events.v1.PitEntered has no entry in the catalogue")
}

// withEvent is the generated messages plus a new event message in the events
// package, as a later .proto under proto/twr/events/ would add.
func withEvent(t *testing.T, name string) *protoregistry.Files {
	t.Helper()
	files := new(protoregistry.Files)
	protoregistry.GlobalFiles.RangeFiles(func(f protoreflect.FileDescriptor) bool {
		if err := files.RegisterFile(f); err != nil {
			t.Fatal(err)
		}
		return true
	})
	fd, err := protodesc.NewFile(&descriptorpb.FileDescriptorProto{
		Name:        proto.String("twr/events/v1/extra.proto"),
		Package:     proto.String(catalogue.EventsPackage),
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

func TestTwoEntriesWithOneCETypeFail(t *testing.T) {
	assertFails(t, load(t, testdata, "duplicate-ce-type.yaml"), protoregistry.GlobalFiles,
		`ce_type "run.thosewho.session.started.v1" is subjects[1]'s too`)
}

// Each of the other checks, on the real catalogue with one entry broken.
func TestEachCheck(t *testing.T) {
	for _, tc := range []struct {
		name   string
		change func(*catalogue.Subject)
		want   string
	}{
		{"a pattern outside twr.v1.", func(s *catalogue.Subject) { s.Pattern = "twr.v2.<src>.frame.gt7" }, "does not start with twr.v1."},
		{"an unknown placeholder", func(s *catalogue.Subject) { s.Pattern = "twr.v1.<source>.frame.gt7" }, "unknown placeholder <source> in the pattern"},
		{"a pattern outside ADR 0001's scheme", func(s *catalogue.Subject) { s.Pattern = "twr.v1.<src>.frames.gt7" }, "does not follow ADR 0001's subject scheme"},
		{"a frame in the wrong stream", func(s *catalogue.Subject) { s.Stream = "EVENTS" }, "a frame subject is captured by the TELEMETRY stream"},
		{"neither a stream nor a kv", func(s *catalogue.Subject) { s.Stream = "" }, "it needs a stream or a kv"},
		{"a stream and a kv", func(s *catalogue.Subject) { s.KV = "live" }, "it needs a stream or a kv, and not both"},
		{"a ce_type without a version", func(s *catalogue.Subject) { s.CEType = "run.thosewho.frame.gt7" }, "is not run.thosewho.<area>.<what>.v<N>"},
		{"another content type", func(s *catalogue.Subject) { s.ContentType = "application/protobuf" }, `content_type "application/protobuf" is not application/json`},
		{"a message not generated", func(s *catalogue.Subject) { s.Message = "twr.telemetry.v1.Frames" }, "message twr.telemetry.v1.Frames is not in the generated code"},
		{"an enum named as the message", func(s *catalogue.Subject) { s.Message = "twr.telemetry.v1.Phase" }, "twr.telemetry.v1.Phase is not a message"},
		{"no msg_id", func(s *catalogue.Subject) { s.MsgID = "" }, "no msg_id"},
		{"an unknown placeholder in msg_id", func(s *catalogue.Subject) { s.MsgID = "<run>:frame:<n2>" }, "unknown placeholder <n2> in msg_id"},
		{"<sid> in msg_id without it in the pattern", func(s *catalogue.Subject) { s.MsgID = "<sid>:frame:<seq>" }, "msg_id uses <sid>, which the pattern does not carry"},
		{"no producer", func(s *catalogue.Subject) { s.Producer = "" }, "no producer"},
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

// A key subjects.yaml doesn't define, such as a misspelling, fails to parse
// rather than being dropped.
func TestAnUnknownKeyFails(t *testing.T) {
	_, err := catalogue.Parse(strings.NewReader("subjects:\n  - pattern: twr.v1.<src>.frame.gt7\n    stram: TELEMETRY\n"))
	if err == nil || !strings.Contains(err.Error(), "field stram not found") {
		t.Errorf("got %v, want an error about the unknown key", err)
	}
}

// The event messages are those that travel on their own: SessionSummary, held by
// SessionEnded, is not one.
func TestEventMessages(t *testing.T) {
	got := catalogue.EventMessages(protoregistry.GlobalFiles)
	want := []protoreflect.FullName{
		"twr.events.v1.LapCompleted", "twr.events.v1.SessionEnded", "twr.events.v1.SessionStarted",
	}
	if strings.Join(names(got), " ") != strings.Join(names(want), " ") {
		t.Errorf("got %v, want %v", got, want)
	}
}

func names(n []protoreflect.FullName) []string {
	out := make([]string, len(n))
	for i, name := range n {
		out[i] = string(name)
	}
	return out
}
